terraform {
  required_version = ">= 0.12"
  required_providers {
    b2 = {
      source  = "localhost/backblaze/b2"
      version = "~> 0.1"
    }
  }
}

provider "b2" {
}

resource "b2_bucket" "example" {
  bucket_name = "Example-TestBucket"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "example" {
  bucket_id = b2_bucket.example.id
  file_name = "example.txt"
  source = "example.txt"
}

data "b2_bucket" "example" {
  bucket_name = b2_bucket.example.bucket_name
}

output "bucket" {
  value = data.b2_bucket.example
}

data "b2_bucket_files" "example" {
  bucket_id = b2_bucket_file_version.example.bucket_id
}

output "bucket_file_info" {
  value = data.b2_bucket_files.example
}
