terraform {
  required_version = ">= 0.13"
  required_providers {
    b2 = {
      source  = "Backblaze/b2"
      version = "~> 0.3"
    }
  }
}

provider "b2" {
}

resource "b2_bucket" "example" {
  bucket_name = "test-b2-tfp-0000000000000000000"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "example1" {
  bucket_id = b2_bucket.example.id
  file_name = "example.txt"
  source    = "example.txt"
}

resource "b2_bucket_file_version" "example2" {
  bucket_id = b2_bucket_file_version.example1.bucket_id
  file_name = b2_bucket_file_version.example1.file_name
  source    = b2_bucket_file_version.example1.source
  file_info = {
    description = "second version"
  }
}

resource "b2_bucket_file_version" "example3" {
  bucket_id = b2_bucket_file_version.example2.bucket_id
  file_name = "dir/example.txt"
  source    = b2_bucket_file_version.example2.source
  server_side_encryption {
    mode = "SSE-B2"
    algorithm = "AES256"
  }
}

data "b2_bucket" "example" {
  bucket_name = b2_bucket.example.bucket_name
}

data "b2_bucket_file" "example" {
  bucket_id     = b2_bucket_file_version.example2.bucket_id
  file_name     = b2_bucket_file_version.example2.file_name
  show_versions = true
}

data "b2_bucket_files" "example" {
  bucket_id = b2_bucket_file_version.example3.bucket_id
}

output "bucket_example" {
  value = data.b2_bucket.example
}

output "bucket_file_example" {
  value = data.b2_bucket_file.example
}

output "bucket_files_example" {
  value = data.b2_bucket_files.example
}
