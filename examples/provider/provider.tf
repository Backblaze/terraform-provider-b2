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
  bucket_name = "my-b2-bucket"
  bucket_type = "allPublic"
}
