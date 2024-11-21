resource "aws_s3_bucket" "example_bucket" {
  bucket = "example-unsecure-bucket"
  acl    = "private" # Updated ACL to private

  tags = {
    Name        = "example-bucket"
    Environment = "Test"
  }
}

resource "aws_s3_bucket_public_access_block" "example" {
  bucket                  = aws_s3_bucket.example_bucket.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "example" {
  bucket = aws_s3_bucket.example_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.example.arn
    }
  }
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example_bucket.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_security_group" "example" {
  ingress {
    cidr_blocks = ["192.168.1.0/24"] # Restrict ingress to specific subnet
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
  }

  egress {
    cidr_blocks = ["192.168.1.0/24"] # Restrict egress
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
  }
}

resource "aws_kms_key" "example" {
  description             = "Example KMS key"
  deletion_window_in_days = 10
}
