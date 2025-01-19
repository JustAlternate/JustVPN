# JustVPN

Deployable infrastructure to provision a selfhosted VPN on a linode nanode instance in one click using wireguard.

# Installation

1) Fill the secrets.tfvars.example

Rename secrets.tfvars.example to secrets.tfvars
```
mv secrets.tfvars.example secrets.tfvars
```

Init terraform

```
terraform init
```

Get the dependencies
```
go get
```

Start the API
```
go run src/main.go
```

# Usage

Request a wireguard server
```
curl --data "IP=<your_public_ip>&timeWantedBeforeDeletion=100" localhost:8081/start
```


