# Bucket names must be unique across all B2 accounts.
variable "bucket_name" {
  description = "A globally unique B2 bucket name"
  type        = string
}

terraform {
  required_version = ">= 1.0.0"
  required_providers {
    b2 = {
      source = "Backblaze/b2"
    }
  }
}

provider "b2" {
}

resource "b2_application_key" "example_key" {
  key_name     = "my-key"
  capabilities = ["readFiles"]
}

resource "b2_bucket" "example_bucket" {
  bucket_name = var.bucket_name
  bucket_type = "allPublic"
}
