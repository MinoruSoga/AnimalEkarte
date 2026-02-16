terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "AnimalEkarte"
}

# ---- S3 bucket for Terraform state ----
resource "aws_s3_bucket" "tfstate" {
  bucket = "animalekarte-tfstate-698109622668"

  tags = {
    Name        = "AnimalEkarte Terraform State"
    Environment = "shared"
    ManagedBy   = "terraform"
  }
}

resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "tfstate" {
  bucket                  = aws_s3_bucket.tfstate.id
  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = true
  restrict_public_buckets = true
}

# ---- DynamoDB table for state lock ----
resource "aws_dynamodb_table" "tflock" {
  name         = "animalekarte-terraform-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    Name        = "AnimalEkarte Terraform Lock"
    Environment = "shared"
    ManagedBy   = "terraform"
  }
}

output "tfstate_bucket" {
  value       = aws_s3_bucket.tfstate.bucket
  description = "S3 bucket name for Terraform state"
}

output "tfstate_bucket_arn" {
  value       = aws_s3_bucket.tfstate.arn
  description = "S3 bucket ARN for IAM policies"
}

output "tflock_table" {
  value       = aws_dynamodb_table.tflock.name
  description = "DynamoDB table name for state locking"
}

output "tflock_table_arn" {
  value       = aws_dynamodb_table.tflock.arn
  description = "DynamoDB table ARN for IAM policies"
}
