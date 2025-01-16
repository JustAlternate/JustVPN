terraform {
  required_providers {
    linode = {
      source = "linode/linode"
      version = "2.32.0"
    }
  }
}

provider "linode" {
  token = var.linode_token
}

resource "linode_instance" "wireguard-one-click" {
  label = "wireguardreg-one-click"
  image = "linode/ubuntu22.04"
  region = var.region
  type = "g6-nanode-1"
  authorized_keys = [ "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKSO4cOiA8s9hVyPtdhUXdshxDXXPU15qM8xE0Ixfc21 justalternate@archlinux" ]
  root_pass = var.root_pass
  private_ip = false
  stackscript_id = 401706
  stackscript_data = {
    port             = "51820"
    privateip        = "10.0.1.1/24"
    peerpubkey       = var.interface_public_key
    privateip_client = "10.0.1.2/24"
    endpoint         = var.endpoint
  }
}

output "instance_ip" {
  value = linode_instance.wireguard-one-click.ip_address
}
