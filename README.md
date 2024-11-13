# SAST-VMCreator

SAST-VMCreator 用于 SAST 内部 ESXI 虚拟机创建。使用 `terraform` 和 `cloud-init` 通过向飞书机器人发送指令创建虚拟机。

## Requirements

- Terraform
- [ovftool](https://developer.broadcom.com/tools/open-virtualization-format-ovf-tool/latest)
- 飞书应用的 app_id 和 app_secret

## Usage

准备 `terraform.tfvars` 配置并放置到 `terraform` 目录，内容示例如下：

```hcl
esxi_hostname  = "ip"
esxi_hostport  = 22
esxi_hostssl   = 443
esxi_username  = "root"
esxi_password  = "passowrd"

ssh_username   = "your-ssh-username"
ssh_public_key = "your-public-key"
hostname       = "vm-hostname"
vm_name        = "vm-name"
numvcpus       = 2
memory         = 2048
disk_size      = 10
disk_type      = "thin"
ovf_source     = "you-vm-source-file"
clone_from_vm  = ""
datastore      = "datastore"
network_name   = "VM Network"
```
> memory 单位 MB, disk_size 单位 GB

```bash
export APP_ID="your-app-id" APP_SECRET="your-app-secret"
go run .
```

### Docker

这里的 Docker 镜像遵循能跑就行原则。

目前 docker 打包镜像用的是 Arch Linux，因为 `ovftool` 是一个二进制，在其他系统上暂时还没找到运行的解决方案。

如果需要使用 Docker 运行，需要提供 ovftool 下载 URL 和 icu60 的 Arch Linux 安装包。替换 `install-ovftool.sh` 脚本中的：

```bash
# Since we can't download from the original URL, we'll use a local mirror
DOWNLOAD_URL="http://192.168.114.4:54331/images/VMware-ovftool-4.6.3-24031167-lin.x86_64.zip"
# Donwload icu60 from local mirror
ICU60URL="http://192.168.114.4:54331/images/icu60-60.3-1-x86_64.pkg.tar.zst"
```

参考 AUR 中的 PKGBULD:

- icu60: https://aur.archlinux.org/packages/icu60
- ovftool: https://aur.archlinux.org/packages/vmware-ovftool

之后可以考虑添加 base 镜像直接将依赖安装在镜像中，而不用每次 build 镜像都需要重新下载依赖。

将 app id 和 app secret 填入 `.env` 文件中，准备 `terraform.tfvars` 配置并放置到 `terraform` 目录。

```bash
docker compose up -d
```

