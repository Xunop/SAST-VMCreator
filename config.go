package main

import "os"

var (
	// AppID is the app id
	AppID         string
	AppSecret     string
	ExampleConfig string
	HelpMsg       string
)

func init() {
	AppID = os.Getenv("APP_ID")
	AppSecret = os.Getenv("APP_SECRET")
    readConfig()
	HelpMsg = `/createvm - 创建虚拟机
/help - 显示帮助信息
配置文件解释：
esxi_hostname  = "ip"                                     # ESXI 主机地址
esxi_hostport  = 22
esxi_hostssl   = 443
esxi_username  = "root"
esxi_password  = "password"

ssh_username   = "ubuntu"                                 # SSH 用户名
ssh_public_key = ""
hostname       = "vm"
vm_name        = "vm"                                     # 生成的 VM 名称，要确保唯一
numvcpus       = 2                                        # CPU 核数
memory         = 2048                                     # 内存大小，单位 MB
disk_size      = 10                                       # 硬盘大小，单位 GB
disk_type      = "thin"                                   # 硬盘类型，thin 或 thick
ovf_source     = "source_url"                             # 从 NAS 下载虚拟机配置
clone_from_vm  = ""                                       # 从已有 VM 克隆，为空则不克隆
datastore      = "datastore"                              # 存储名称，目前 sast esxi 上只有这个
network_name   = "VM Network"                             # 网络名称，默认
`
}

// readConfig reads the terraform.tfvars file to ExampleConfig variable
func readConfig() {
    f, err := os.Open("terraform/terraform.tfvars")
    if err != nil {
        return
    }
    defer f.Close()

    buf := make([]byte, 1024)
    n, _ := f.Read(buf)
    ExampleConfig = string(buf[:n])
}
