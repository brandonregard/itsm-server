app = "itsm-server"
environment = "dev"
internal = "false"
container_port = "3000"
replicas = "1"
health_check = "/health"
region = "us-east-1"
aws_profile = "default"
vpc = "vpc-0b3c769c8c579098f"
private_subnets = "subnet-09fe7047d62dda00a,subnet-00575e5b8c6093fa9"
public_subnets = "subnet-0dde81d9e5468eb5d,subnet-0946d3f99953ac98a"
tags = {
  application = "itsm-server"
  environment = "dev"
}