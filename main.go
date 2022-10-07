package main

import (
	"fmt"
	"os"
	// "packer-plugin-scaffolding/builder/scaffolding"
	// packercueData "packer-plugin-scaffolding/datasource/scaffolding"
	// packercuePP "packer-plugin-scaffolding/post-processor/scaffolding"
	packercueProv "packer-plugin-cue/provisioner/packercue"
	packercueVersion "packer-plugin-cue/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	// pps.RegisterBuilder("my-builder", new(packercue.Builder))
	pps.RegisterProvisioner("cue", new(packercueProv.Provisioner))
	// pps.RegisterPostProcessor("my-post-processor", new(packercuePP.PostProcessor))
	// pps.RegisterDatasource("my-datasource", new(packercueData.Datasource))
	pps.SetVersion(packercueVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
