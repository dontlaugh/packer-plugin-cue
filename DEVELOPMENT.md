# Developer Guide

Packer runs 3rd party plugins as subprocesses, and communicates with them over
RPC via stdin/stdout. The packer libraries handle all of the low level details,
so we just need to implement interfaces to do CUE evaluation on the user's
workstation.

## Testing Your Binary

Running `make` drops a binary in the root of this repo.

There are two steps to make Packer use your development version of this plugin.

1. Copy the built binary to a suitable place on disk
2. Comment out any official released version from `required_plugins` in your HCL

This plugin is a "multi-component" plugin. Consult the [manual installation](https://www.packer.io/docs/plugins/install-plugins)
procedure for more details. You will need to move the binary to a location Packer
can find it. The easiest thing to do is to **move the binary to the directory
where you invoke packer**.

For example, if you want to run a build in the "examples" folder, copy `packer-plugin-cue`
there.

For step 2, simply make sure there is no reference to "github.com/dontlaugh/cue"
in your `required_plugins` config.

```diff
packer {
  required_plugins {
     lxd = {
       version = ">=1.0.0"
       source  = "github.com/hashicorp/lxd"
     }
-    cue = {
-      version = ">=0.3.1"
-      source = "github.com/dontlaugh/cue"
-    }
+    // cue = {
+    //  version = ">=0.3.1"
+    //  source = "github.com/dontlaugh/cue"
+    // }
   }
 }
```

The custom `cue-export` provisioner will still work.

## Debug Logging

Set `PACKER_LOG=1` in your environment for extra logging. This log output will
also indicate whether Packer has located your custom built plugin.

## HCL Config Codegen

If you make a change to a config struct, you need to regenerate some files.

```
make generate
```

## Regarding zlconf/go-cty

We _cannot_ upgrade this dependency until this issue is resolved

* https://github.com/hashicorp/packer-plugin-sdk/issues/128

