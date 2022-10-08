packer {
  required_plugins {

    lxd = {
      version = ">=1.0.0"
      source  = "github.com/hashicorp/lxd"
    }

  }
}


source "lxd" "ubuntu_2204" {
  image = "images:ubuntu/22.04/cloud"
}

locals {
  ts = formatdate("YYYYMMDDhhmmss", timestamp())
}

build {
  source "lxd.ubuntu_2204" {
    name           = "cue-test"
    output_image         = "cue-test"
    container_name = "cue-test-${local.ts}"
  }

  provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "hello"
    expression = "world"
    dest       = "/root/foo"
  }

  provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "hello"
    expression = "expect_bytes"
    dest       = "/root/expect_bytes"
  }


  provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "hello"
    expression = "expect_json"
    serialize = "json"
    dest       = "/root/expect_json"
  }

  provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "hello"
    expression = "expect_yaml"
    serialize = "yaml"
    dest       = "/root/expect_yaml"
  }

  provisioner "cue-export" {
    module     = "."
    dir        = "."
    package    = "hello"
    expression = "expect_toml"
    serialize = "toml"
    dest       = "/root/expect_toml"
  }
}
