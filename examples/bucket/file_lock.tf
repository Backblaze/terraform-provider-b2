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

resource "b2_bucket" "example" {
  bucket_name = "test-b2-lock"
  bucket_type = "allPublic"
  file_lock_configuration {
    is_file_lock_enabled = true
    default_retention {
      mode = "governance"
      period {
        duration = 7
        unit     = "days"
      }
    }
  }
}

