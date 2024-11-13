terraform {
  required_version = ">= 0.13"
  required_providers {
    esxi = {
      source  = "registry.terraform.io/josenk/esxi"
      version = "1.10.3"
    }
  }
}

provider "esxi" {
  esxi_hostname = var.esxi_hostname
  esxi_hostport = var.esxi_hostport
  esxi_hostssl  = var.esxi_hostssl
  esxi_username = var.esxi_username
  esxi_password = var.esxi_password
}

resource "esxi_virtual_disk" "data_disk" {
  virtual_disk_disk_store = var.datastore
  virtual_disk_dir        = "vm-data-disk"
  virtual_disk_name       = "${var.vm_name}_data_disk.vmdk"
  virtual_disk_size       = var.disk_size
  virtual_disk_type       = var.disk_type
}

resource "esxi_guest" "vm" {
  guest_name = var.vm_name
  disk_store = var.datastore
  numvcpus   = var.numvcpus
  memsize    = var.memory

  clone_from_vm = var.clone_from_vm != "" ? var.clone_from_vm : null
  ovf_source    = var.clone_from_vm == "" ? var.ovf_source : null

  network_interfaces {
    virtual_network = var.network_name
  }

  guestinfo = {
    "userdata" = base64encode(templatefile("${path.module}/userdata.yaml", {
      ssh_username   = var.ssh_username
      hostname       = var.hostname
      ssh_public_key = var.ssh_public_key
    }))
    "userdata.encoding" = "base64"
  }

  virtual_disks {
    virtual_disk_id = esxi_virtual_disk.data_disk.id
    slot            = "0:1"
  }
}

output "ip" {
  value = [esxi_guest.vm.ip_address]
}

# VM name
# output "vm_name" {
#   value = esxi_guest.vm.guest_name
# }

# output "ssh_command" {
#   value = "ssh ${var.ssh_username}@${esxi_guest.vm.ip_address}"
# }
