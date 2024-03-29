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

resource "b2_application_key" "example" {
  key_name     = "test-b2-tfp-0000000000000000000"
  capabilities = ["readFiles"]
}

data "b2_application_key" "example" {
  key_name = b2_application_key.example.key_name
}

output "application_key" {
  value = data.b2_application_key.example
}
