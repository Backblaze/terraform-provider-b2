terraform {
  required_version = ">= 0.13"
  required_providers {
    b2 = {
      source = "Backblaze/b2"
    }
  }
}

provider "b2" {

}

resource "b2_bucket" "example" {
  bucket_name = "test-b2-lock-0000000004310000020"
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
