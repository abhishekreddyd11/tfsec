# Define an S3 bucket with secure configurations
resource "aws_s3_bucket" "example_bucket" {
  # Updated bucket name
  bucket = "example-secure-bucket"
  # Set ACL to private to prevent public access
  acl    = "private"

  # Tags for identifying the bucket
  tags = {
    Name        = "example-secure-bucket"
    Environment = "Test"
  }
}

# Add a public access block to the bucket
resource "aws_s3_bucket_public_access_block" "example_bucket_block" {
  # Reference the bucket
  bucket = aws_s3_bucket.example_bucket.id
  # Block public ACLs
  block_public_acls = true
  # Block public bucket policies
  block_public_policy = true
  # Ignore public ACLs applied to objects
  ignore_public_acls = true
  # Restrict public buckets entirely
  restrict_public_buckets = true
}

# Enable server-side encryption for the bucket using AWS KMS
resource "aws_s3_bucket_server_side_encryption_configuration" "example_bucket_encryption" {
  # Reference the bucket
  bucket = aws_s3_bucket.example_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      # Use KMS for encryption
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.example_key.arn
    }
  }
}

# Define a KMS (Key Management Service) key for encryption
resource "aws_kms_key" "example_key" {
  # Description of the KMS key
  description = "KMS key for S3 bucket encryption"
  # Key deletion window in days
  deletion_window_in_days = 10
  key_rotation_enabled = true
}

# Enable versioning for the S3 bucket to keep multiple versions of objects
resource "aws_s3_bucket_versioning" "example_bucket_versioning" {
  # Reference the bucket
  bucket = aws_s3_bucket.example_bucket.id

  versioning_configuration {
    # Enable versioning
    status = "Enabled"
  }
}

# Enable logging for the S3 bucket
resource "aws_s3_bucket_logging" "example_bucket_logging" {
  # Reference the bucket
  bucket        = aws_s3_bucket.example_bucket.id
  # Target bucket to store logs (you must replace this with an existing bucket)
  target_bucket = "logging-bucket-name"
  # Prefix for log files
  target_prefix = "example-secure-bucket-logs/"
}
