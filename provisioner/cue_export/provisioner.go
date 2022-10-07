//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package cue_export

import (
	"bytes"
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"fmt"
	"os"

	_ "cuelang.org/go/cue/load"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Config models our CUE export provisioner config.
//
// Users are expected to provide a combination of package, module root,
// expression, etc. that yields a string that becomes a file on disk.
type Config struct {
	// CUE module params
	ModuleRoot string   `mapstructure:"module"`
	Package    string   `mapstructure:"package"`
	Expression string   `mapstructure:"expression"`
	Tags       []string `mapstructure:"tags"`
	TagVars    []string `mapstructure:"tag_vars"`
	Dir        string   `mapstructure:"dir"`

	// Destination file; TODO(cm) rhyme w/ Packer's file prov.
	DestFile     string `mapstructure:"dest"`
	DestFileMode uint

	// CUE output type?
	// todo

	// Packer internals
	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

// Prepare loads and validates our config(s). Configs should be merged in some sensible way.
func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "packer.provisioner.cue",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provisioner) Provision(_ context.Context, ui packer.Ui, comm packer.Communicator, _generatedData map[string]interface{}) error {
	ui.Say("cue provisioner")
	ctx := cuecontext.New()

	// load the cue package
	instances := load.Instances([]string{}, &load.Config{
		Context:    nil,
		ModuleRoot: p.config.ModuleRoot,
		Module:     "",
		Package:    p.config.Package,
		Dir:        p.config.Dir, // usually same as ModuleRoot?
		// What to do here?
		//Tags:        p.config.Tags,
		//TagVars:     p.config.TagVars,
		AllCUEFiles: false,
		Tests:       false,
		Tools:       false,
		DataFiles:   false,
		StdRoot:     "",
		ParseFile:   nil,
		Overlay:     nil,
		Stdin:       nil,
	})
	if err := instances[0].Err; err != nil {
		return fmt.Errorf("loading instances: %w", err)
	}
	val := ctx.BuildInstance(instances[0])
	if val.Err() != nil {
		return fmt.Errorf("building instances: %w", val.Err())
	}

	if p.config.Expression != "" {
		expr := cue.ParsePath(p.config.Expression)
		val = val.LookupPath(expr)
	}
	switch val.Kind() {
	case cue.BytesKind:
		// we render bytes directly to a file
		panic("not implemented ")
	case cue.StringKind:
		// we render string to a file, also
		stringValue, err := val.String()
		if err != nil {
			return err
		}
		buf := bytes.NewBufferString(stringValue)
		if err := comm.Upload(p.config.DestFile, buf, nil); err != nil {
			return err
		}
	case cue.StructKind, cue.ListKind:
		// struct or list can be serialized directly
		// to json, yaml, toml, etc.
		panic("not implemented ")
	default:
		// BottomKind, not a concrete value?
		return fmt.Errorf("unsuppored CUE kind: %v", val.Kind())

	}

	return nil
}

/* copied from an old provisioner i wrote */

// createDir creates a directory on the remote server
func (p *Provisioner) createDir(ctx context.Context, ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := packer.RemoteCmd{Command: fmt.Sprintf("mkdir -p '%s'", dir)}

	if err := execRemoteCommand(ctx, comm, &cmd, ui, "create dir"); err != nil {
		return err
	}
	return nil
}

// uploadFile uploads a file.
func (p *Provisioner) uploadFile(ctx context.Context, ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening: %s", err)
	}
	if err = comm.Upload(dst, f, nil); err != nil {
		return fmt.Errorf("error uploading %s: %s", src, err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

// uploadDir uploads a directory
func (p *Provisioner) uploadDir(ctx context.Context, ui packer.Ui, comm packer.Communicator, dst, src string) error {
	var ignore []string
	if err := p.createDir(ctx, ui, comm, dst); err != nil {
		return err
	}
	// TODO: support Windows '\'
	if src[len(src)-1] != '/' {
		src = src + "/"
	}
	return comm.UploadDir(dst, src, ignore)
}

// execRemoteCommand executes a packer.RemoteCommand, blocks, and checks for exit code 0.
func execRemoteCommand(ctx context.Context, comm packer.Communicator, cmd *packer.RemoteCmd, ui packer.Ui, msg string) error {
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return fmt.Errorf("error %s: %v", msg, err)
	}
	if code := cmd.ExitStatus(); code != 0 {
		return fmt.Errorf("%s non-zero exit status: %v", msg, code)
	}
	return nil
}