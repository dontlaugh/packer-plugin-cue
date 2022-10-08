// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package cue_export

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	ModuleRoot   *string  `mapstructure:"module" cty:"module" hcl:"module"`
	Package      *string  `mapstructure:"package" cty:"package" hcl:"package"`
	Expression   *string  `mapstructure:"expression" cty:"expression" hcl:"expression"`
	Tags         []string `mapstructure:"tags" cty:"tags" hcl:"tags"`
	TagVars      []string `mapstructure:"tag_vars" cty:"tag_vars" hcl:"tag_vars"`
	Dir          *string  `mapstructure:"dir" cty:"dir" hcl:"dir"`
	DestFile     *string  `mapstructure:"dest" cty:"dest" hcl:"dest"`
	DestFileMode *int     `mapstructure:"file_mode" cty:"file_mode" hcl:"file_mode"`
	Serialize    *string  `mapstructure:"serialize" cty:"serialize" hcl:"serialize"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"module":     &hcldec.AttrSpec{Name: "module", Type: cty.String, Required: false},
		"package":    &hcldec.AttrSpec{Name: "package", Type: cty.String, Required: false},
		"expression": &hcldec.AttrSpec{Name: "expression", Type: cty.String, Required: false},
		"tags":       &hcldec.AttrSpec{Name: "tags", Type: cty.List(cty.String), Required: false},
		"tag_vars":   &hcldec.AttrSpec{Name: "tag_vars", Type: cty.List(cty.String), Required: false},
		"dir":        &hcldec.AttrSpec{Name: "dir", Type: cty.String, Required: false},
		"dest":       &hcldec.AttrSpec{Name: "dest", Type: cty.String, Required: false},
		"file_mode":  &hcldec.AttrSpec{Name: "file_mode", Type: cty.Number, Required: false},
		"serialize":  &hcldec.AttrSpec{Name: "serialize", Type: cty.String, Required: false},
	}
	return s
}
