# packer-plugin-cue

A packer plugin that writes files to the image your provisioning, based
on values extracted from a CUE module.

A Packer data provider is TODO.

## Provided Plugins

Provisioners:

* `cue-export` - Given a module, package, and an optional CUE expression, the
  provisioner writes a file with the scalar value or struct yielded by `cue export`.

## Example Usage

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
field and render it, with the value of `config_file` interpolated.

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


