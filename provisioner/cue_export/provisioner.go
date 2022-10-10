//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package cue_export

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const pluginType = "packer.provisioner.cue"

// Config models our CUE export provisioner config.
//
// Users are expected to provide a combination of package, module root,
// expression, etc. that selects a Value. That Value becomes a file on disk
// during the Packer build.
type Config struct {

	// CUE module params
	ModuleRoot string `mapstructure:"module_root"`
	Module     string `mapstructure:"module"`

	Package    string   `mapstructure:"package"`
	Expression string   `mapstructure:"expression"`
	Tags       []string `mapstructure:"tags"`
	TagVars    []string `mapstructure:"tag_vars"` // remove?
	Dir        string   `mapstructure:"dir"`

	// Destination file; TODO(cm) rhyme w/ Packer's file prov.
	DestFile     string `mapstructure:"dest"`
	DestFileMode int    `mapstructure:"file_mode"`

	// Must be one of yaml, toml, json
	Serialize string `mapstructure:"serialize"`

	// Packer internals
	ctx interpolate.Context
}

// Provisioner implements the Packer plugin's interface.
type Provisioner struct {
	config Config

	// Serialization methods set this buffer, which is then passed to Upload
	buf *bytes.Buffer

	// cue instanances and value created in Prepare and used in Provision
	instances []*build.Instance
	value     cue.Value
}

// ConfigSpec is called by Packer to get the HCL version of our config.
func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

// Prepare loads and validates our config(s). Configs should be merged in some sensible way.
func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         pluginType,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Ensure serialize has a valid value.
	if ser := p.config.Serialize; ser != "" {
		validSerializers := []string{"json", "yaml", "toml"}
		var isValid bool
		for _, s := range validSerializers {
			if ser == s {
				isValid = true
			}
		}
		if !isValid {
			return fmt.Errorf("'%s' is not valid; serialize field, if present, must be one of: toml, json, yaml", p.config.Serialize)
		}
	}

	ctx := cuecontext.New()

	// load the cue package
	instances := load.Instances([]string{}, &load.Config{
		Context:    nil,
		ModuleRoot: p.config.ModuleRoot, // what do we put here?
		Module:     p.config.Module,
		Package:    p.config.Package,
		Dir:        p.config.Dir, // usually same as ModuleRoot?
		// What to do here?
		Tags: p.config.Tags,
		//TagVars:     p.config.TagVars,
		//
		// Do we need configs for these other values?
		AllCUEFiles: false,
		Tests:       false,
		Tools:       false,
		DataFiles:   false,
		StdRoot:     "",
		ParseFile:   nil,
		Overlay:     nil,
		Stdin:       nil,
	})
	p.instances = instances
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
	p.value = val
	return nil
}

func (p *Provisioner) Provision(_ context.Context, ui packer.Ui, comm packer.Communicator, _generatedData map[string]interface{}) error {
	ui.Say(fmt.Sprintf("cue-export provisioning file %s", p.config.DestFile))

	ui.Message(fmt.Sprintf("cue module: %s; package: %s; expression: %s",
		p.config.ModuleRoot, p.config.Package, p.config.Expression))

	val := p.value

	log.Printf("cue kind: %v\n", val.Kind())
	switch val.Kind() {

	// BytesKind is rendered as-is to a file
	case cue.BytesKind:
		bytesValue, err := val.Bytes()
		if err != nil {
			return err
		}
		if err := p.serializeBytes(bytesValue); err != nil {
			return err
		}

	// StringKind is rendered as-is to a file, also
	case cue.StringKind:
		stringValue, err := val.String()
		if err != nil {
			return err
		}
		if err := p.serializeString(stringValue); err != nil {
			return err
		}

	// StructKind uses a serializer to render json, toml, or yaml
	case cue.StructKind:
		var msi map[string]interface{}
		if err := val.Decode(&msi); err != nil {
			ui.Error("could not decode to map[string]interface{}")
			return err
		}

		if err := p.serializeStruct(msi); err != nil {
			return err
		}

	// ListKind isn't implemented, but should render like StructKind?
	case cue.ListKind:

		panic("not implemented")

	// BottomKind, or anything else, is unsupported. Perhaps the expression
	// does not render a concrete value.
	default:
		return fmt.Errorf("unsuppored CUE kind: %v", val.Kind())
	}

	// Our buf has been filled; Upload reads it into a remote file.
	// TODO(cm) set mode?
	if err := comm.Upload(p.config.DestFile, p.buf, nil); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) serializeString(s string) error {
	buf := bytes.NewBufferString(s)
	p.buf = buf
	return nil
}

func (p *Provisioner) serializeBytes(b []byte) error {
	buf := bytes.NewBuffer(b)
	p.buf = buf
	return nil
}

func (p *Provisioner) serializeStruct(msi map[string]interface{}) error {
	var buf = &bytes.Buffer{}
	switch p.config.Serialize {
	case "yaml":
		bytesVal, err := yaml.Marshal(msi)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(bytesVal)

	case "toml":
		if err := toml.NewEncoder(buf).Encode(msi); err != nil {
			return err
		}

	case "json":
		bytesVal, err := json.MarshalIndent(msi, "", "    ")
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(bytesVal)

	default:
		panic("invalid serialize type") // should never happen
	}

	p.buf = buf

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
