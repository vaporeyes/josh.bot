# ABOUTME: This file configures the Terraform backend for remote state storage.
# ABOUTME: It uses an S3 bucket and DynamoDB table for state locking.
terraform {
  backend "s3" {
    use_lockfile = true
  }
}
