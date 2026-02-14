# ABOUTME: This file specifies the required Terraform version and cloud provider configuration.
# ABOUTME: It ensures a consistent working environment for infrastructure as code.
terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
