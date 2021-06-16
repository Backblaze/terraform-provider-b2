
variable "encryption_key" {
  # value will be taken from env variable 'TF_VAR_encryption_key'
  description = "Encryption key"
  type        = string
  sensitive   = true
}

resource "b2_bucket" "test" {
  bucket_name = "!bucketname!"
  bucket_type = "allPublic"
}

resource "b2_bucket_file_version" "test" {
  bucket_id    = b2_bucket.test.id
  file_name    = "temp.bin"
  source       = "!filename!"
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
