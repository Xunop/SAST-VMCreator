variable "esxi_hostname" {
  type        = string
  description = "ESXi 主机 IP 地址"
  default     = "192.168.114.99"
}

variable "esxi_hostport" {
  type        = number
  description = "ESXi SSH 端口"
  default     = 22
}

variable "esxi_hostssl" {
  type        = number
  description = "ESXi HTTPS 端口"
  default     = 443
}

variable "esxi_username" {
  type        = string
  description = "ESXi 登录用户名"
  default     = "root"
}

variable "esxi_password" {
  type        = string
  description = "ESXi 登录密码"
  sensitive   = true
}

variable "ssh_username" {
  type        = string
  description = "虚拟机 SSH 用户名"
  default     = "ubuntu"
}

variable "hostname" {
  type        = string
  description = "虚拟机主机名"
  default     = "sast-vm"
}

variable "vm_name" {
  type        = string
  description = "虚拟机名称"
  default     = "vm"
}

variable "numvcpus" {
  type        = number
  description = "虚拟机 CPU 核数"
  default     = 2
}

variable "memory" {
  type        = number
  description = "虚拟机内存大小 (MB)"
  default     = 2048
}

variable "disk_size" {
  type        = number
  description = "虚拟机数据盘大小 (GB)"
  default     = 10
}

variable "disk_type" {
  type        = string
  description = "虚拟机数据盘类型 (thin 或 thick)"
  default     = "thin"
}

variable "ovf_source" {
  type        = string
  description = "OVF 模板路径"
  default     = "jammy-server-cloudimg-amd64.ova"
}

variable "clone_from_vm" {
  type        = string
  description = "克隆源虚拟机模板路径 (可选)"
  default     = ""
}

variable "datastore" {
  type        = string
  description = "虚拟机数据存储路径"
  default     = "99-datastore0"
}

variable "network_name" {
  type        = string
  description = "虚拟机网络名称"
  default     = "VM Network"
}

variable "ssh_public_key" {
  description = "SSH public key for the VM user"
  type        = string
}
