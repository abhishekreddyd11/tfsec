# Main Terraform file with vulnerabilities for testing
resource "aws_security_group" "example" {
  name        = "example-security-group"
  description = "Allow SSH and HTTP traffic"
  vpc_id      = "vpc-12345678"

  # Insecure ingress rule: allows access from the entire Internet
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow SSH from anywhere" # Vulnerable: unrestricted SSH
  }

  # Insecure ingress rule: allows HTTP access to the whole internet
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow HTTP traffic from anywhere"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic" # Vulnerable: unrestricted outbound traffic
  }

  tags = {
    Name = "example-security-group"
  }
}

resource "aws_s3_bucket" "example_bucket" {
  bucket = "example-unsecure-bucket"
  acl    = "public-read" # Vulnerable: publicly accessible bucket

  tags = {
    Name        = "example-bucket"
    Environment = "Test"
  }
}

resource "aws_iam_role" "example_role" {
  name               = "example-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "example_policy" {
  name = "example-policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "*", # Vulnerable: overly permissive action
      "Resource": "*"
    }
  ]
}
EOF
}
