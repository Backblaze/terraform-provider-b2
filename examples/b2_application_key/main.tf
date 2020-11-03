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

data "b2_application_key" "test" {
  key_name = "TestKey"
}

output "application_key_id" {
  value = data.b2_application_key.test
}
