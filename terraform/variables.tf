# ABOUTME: This file defines the input variables for the Terraform configuration.
# ABOUTME: It allows for parameterizing the infrastructure deployment.

variable "bucket_name" {
  description = "The name of the S3 bucket for Terraform state."
  type        = string
}

variable "github_token" {
  description = "GitHub PAT with contents:write scope for diary publishing."
  type        = string
  sensitive   = true
  default     = ""
}

variable "diary_repo_owner" {
  description = "GitHub owner for the Obsidian diary repo."
  type        = string
  default     = "vaporeyes"
}

variable "diary_repo_name" {
  description = "GitHub repo name for the Obsidian diary."
  type        = string
  default     = "obsidian-diary"
}
