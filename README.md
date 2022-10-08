# packer-plugin-cue

A packer plugin that writes files to the image your provisioning, based
on values extracted from a CUE module.

A Packer data provider is TODO.

## Provided Plugins

Provisioners:

* `cue-export` - Given a module, package, and an optional CUE expression, the
  provisioner writes a file with the scalar value or struct yielded by `cue export`.

## Example Usage

Add this plugin to `required_plugins`

```hcl
packer {
  required_plugins {
    cue = {
      version = ">=0.1.0"
      source = "github.com/dontlaugh/cue"
    }
  }
}

```

Write the data yielded by the expression `world` in package `world` to **/tmp/some-file**

```hcl
  provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "hello"
    expression = "world"
    dest       = "/tmp/some-file"
  }
```

## Explanation

Let's say you have a CUE module with a file **my_service.cue** that looks like this

```cue
package systemd

config_file: "/etc/my-service/config.yaml"

unit_file: '''
[Unit]
Description=My example service

[Service]
Environment=MY_CONFIG=\(config_file)
ExecStart=/usr/bin/my-service

[Install]
WantedBy=multi-user.target
'''
```

Invoking `cue export -p systemd ./my_service.cue` gives you the entire package as JSON

```json
{
    "config_file": "/etc/my-service/config.yaml",
    "unit_file": "W1VuaXRdCkRlc2N...=="
}
```
_Note: We've elided most of the unit file's base64 string here._

Providing `--out binary` and `-e unit_file` let's us select the `unit_file`
field and render it. The value of `config_file` is interpolated.

```
$ cue export -p systemd -e unit_file --out binary ./my_service.cue
[Unit]
Description=My example service

[Service]
Environment=MY_CONFIG=/etc/my-service/config.yaml
ExecStart=/usr/bin/my-service

[Install]
WantedBy=multi-user.target
```

This manner of templating files and interpolating values is useful for machine
image provisioning, and this plugin provides a declarative API for rendering CUE
to files during a Packer build.

The previous examples use the `cue` CLI, which you should definitely download
and experiment with if you are unfamiliar with CUE. This plugin, however, uses
CUE as a Go library.

### But wait, there's more!

The `cue-export` provisioner doesn't need a string or binary template to write
files. If the package (with optional expression) evaluates  to a string, number,
or struct type, this plugin does the following

* String and number types are written as a single-line file to `dest`
* Struct types must be given a `serialize` config, one of "json", "yaml",
  or "toml". The entire struct will be serialized in tha format.

A quick example of a struct rendered to yaml

```hcl
provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "prometheus"
    expression = "config_yaml"
    serialize  = "yaml"
    dest       = "/tmp/prometheus.yaml"
}
```

If your expression evaluates to a struct and no serializer is set, it's an error.

