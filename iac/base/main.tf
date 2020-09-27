terraform {
  required_version = ">= 0.12"
}

/**
 * main.tf
 * The main entry point for Terraform run
 * See variables.tf for common variables
 * See ecr.tf for creation of Elastic Container Registry for all environments
 */

# Using the AWS Provider
# https://www.terraform.io/docs/providers/
provider "aws" {
  version = ">= 2.23.0"
  region  = var.region
  profile = var.aws_profile
}

/*
 * Outputs
 * Results from a successful Terraform run (terraform apply)
 * To see results after a successful run, use `terraform output [name]`
 */

# Returns the name of the ECR registry, this will be used later in various scripts
output "docker_registry" {
  value = aws_ecr_repository.app.repository_url
}

