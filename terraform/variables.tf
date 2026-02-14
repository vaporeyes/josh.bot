# ABOUTME: This file defines the input variables for the Terraform configuration.
# ABOUTME: It allows for parameterizing the infrastructure deployment.

variable "bucket_name" {
  description = "The name of the S3 bucket for Terraform state."
  type        = string
}
