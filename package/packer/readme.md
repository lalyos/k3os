Run this commands in Google Shell:
[CloudShell](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/lalyos/k3os&tutorial=package/packer/readme.md)


## Install packer

```
curl -LO https://releases.hashicorp.com/packer/1.4.2/packer_1.4.2_linux_amd64.zip
unzip packer_1.4.2_linux_amd64.zip
sudo mv packer /usr/local/bin/
rm -rf packer_1.4.2_linux_amd64.zip
```

test installation
```
packer --version
```

## Build the image

```
cd package/packer/gcp/
packer build template.json
```