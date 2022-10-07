packer {
  required_plugins {
    cue = {
      version = ">=v0.1.0"
      source  = "github.com/greymatter-io/cue"
    }

    lxd = {
      version = ">=1.0.0"
      source  = "github.com/hashicorp/lxd"
    }

  }
}


source "lxd" "ubuntu_2204" {
  image = "images:ubuntu/22.04/cloud"
}

build {
  sources = [
    "source.lxd.ubuntu_2204",
  ]

  provisioner "cue-export" {
    module_root = "."
    dir = "."
    package = "hello"
    expression = "world"
    dest_dir = "/root/foo"
  }

}
