# S3 bucket with public ACL
resource "aws_s3_bucket" "example_bucket" {
  bucket = "example-vulnerable-bucket" # Bucket name
  acl    = "public-read"              # Publicly accessible bucket

  tags = {
    Name        = "example-vulnerable-bucket"
    Environment = "Test"
  }
}
