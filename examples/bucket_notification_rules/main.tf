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
  bucket_name = "test-b2-tfp-0000000000000000000"
  bucket_type = "allPublic"
}

resource "b2_bucket_notification_rules" "example" {
  bucket_id = b2_bucket.example.id
  notification_rules {
    name = "${b2_bucket.example.bucket_name}-rule-1"
    event_types = [
      "b2:ObjectCreated:*",
      "b2:ObjectDeleted:*",
    ]
    target_configuration {
      target_type = "webhook"
      url         = "https://example.com/webhook"
      custom_headers {
        name  = "myCustomHeader1"
        value = "myCustomHeaderVal1"
      }
    }
  }
}

data "b2_bucket" "example" {
  bucket_name = b2_bucket.example.bucket_name
  depends_on = [
    b2_bucket.example,
  ]
}

data "b2_bucket_notification_rules" "example" {
  bucket_id = b2_bucket.example.bucket_id
  depends_on = [
    b2_bucket_notification_rules.example,
  ]
}

output "bucket_example" {
  value = data.b2_bucket.example
}

output "bucket_notification_rules_example" {
  value = data.b2_bucket_notification_rules.example
}
