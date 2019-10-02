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
packer buildpacker build   \
  -var "zone=$(gcloud config get-value compute/zone)" \
  -var "project_id=$(gcloud config get-value core/project)" \
  template.json

```

## Use The image

```
gcloud compute \
  --project=container-solutions-workshops instances create k3os \
  --zone=europe-west4-a \
  --machine-type=n1-standard-1 \
  --network=default \
  --address=34.90.235.121 \
  --network-tier=PREMIUM \
  --metadata=userdata=ssh_authorized_keys:$'\n'-\ github:lalyos$'\n'k3os:$'\n'\ \ token:\ k3s3cr3t --no-restart-on-failure \
  --maintenance-policy=TERMINATE \
  --preemptible \
  --service-account=692972058411-compute@developer.gserviceaccount.com \
  --scopes=https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append \
  --image=k3os-v020 \
  --image-project=container-solutions-workshops \
  --boot-disk-size=10GB \
  --boot-disk-type=pd-standard \
  --boot-disk-device-name=k3os
```