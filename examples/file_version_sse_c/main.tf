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

provider "b2" {}

variable "encryption_key" {
  # value will be taken from env variable 'TF_VAR_encryption_key'
  description = "Encryption key"
  type        = string
  sensitive   = true
}

variable "source_file" {
  description = "Path to the local file to upload, relative to the working directory"
  type        = string
}

resource "b2_bucket" "test" {
  bucket_name = var.bucket_name
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id    = b2_bucket.test.id
  file_name    = "temp.bin"
  source       = var.source_file
  content_type = "octet/stream"
  file_info = {
    description = "the file"
  }
  server_side_encryption {
    mode      = "SSE-C"
    algorithm = "AES256"
    key {
      secret_b64 = var.encryption_key
      key_id     = "Identifier that will let client tools know which key to provide"
    }
  }
}
