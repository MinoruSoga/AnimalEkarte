variable "name_prefix" {
  description = "Prefix for all resource names"
  type        = string
  default     = "animalekarte-test"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24"]
}

# Database Configuration
variable "db_name" {
  description = "Database name"
  type        = string
  default     = "ekarte_db"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "ekarte_admin"
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro"
}

variable "rds_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 20
}

variable "rds_backup_retention_period" {
  description = "RDS backup retention period in days"
  type        = number
  default     = 1
}

# AWS Region
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

# ECR Configuration
variable "ecr_repository_name" {
  description = "ECR repository name"
  type        = string
  default     = "animalekarte-api"
}

# ECS Configuration
variable "ecs_task_cpu" {
  description = "ECS task CPU units"
  type        = string
  default     = "256"
}

variable "ecs_task_memory" {
  description = "ECS task memory in MB"
  type        = string
  default     = "512"
}

variable "ecs_desired_count" {
  description = "Desired number of ECS tasks"
  type        = number
  default     = 1
}

# RDS Configuration
variable "use_public_rds" {
  description = "Whether to deploy RDS in public subnet with public access (for external tools like TablePlus)"
  type        = bool
  default     = false
}

# コスト最適化: RDS の public IP 付与を制御（subnet group 配置とは独立）。
# false で public IP を剥がし $3.65/月削減 + インターネット露出遮断。DB アクセスは SSM port-forward 経由。
variable "rds_publicly_accessible" {
  description = "Whether the RDS instance gets a public IP"
  type        = bool
  default     = false
}

# コスト最適化: NAT Gateway($32.67/月) を fck-nat EC2(t4g.nano, ~$3/月) に置換。
# STG では 2026-06-01 に切替済み・外向き検証済みのため default true（live 状態と一致）。
# 本番で managed NAT Gateway(HA) を使うなら tfvars で false に上書きする。
variable "use_nat_instance" {
  description = "Use a fck-nat EC2 instance instead of a managed NAT Gateway"
  type        = bool
  default     = true
}

# コスト最適化: 毎日 22:00–8:00 JST に ECS=0 + RDS stop（EventBridge Scheduler）。
# STG では 2026-06-01 に有効化済みのため default true（live 状態と一致）。
# 24/7 稼働が必要な環境では tfvars で false に上書きする。
variable "enable_off_hours_schedule" {
  description = "Enable off-hours stop/start schedules for ECS + RDS"
  type        = bool
  default     = true
}

# ALB HTTPS Configuration
variable "alb_certificate_arn" {
  description = "ACM certificate ARN for ALB HTTPS listener"
  type        = string
  default     = ""
}

# P2: ALB internal + CloudFront VPC Origin
# STG で有効化: terraform.tfvars または TF_VAR_alb_internal=true で上書きする。
# default=false は internet-facing（安全なデフォルト）。
variable "alb_internal" {
  description = "P2: true にすると ALB を internal scheme に変更し CloudFront VPC Origin 経由で接続。false は internet-facing（デフォルト）"
  type        = bool
  default     = false
}

# GitHub Configuration
variable "github_repo" {
  description = "GitHub repository in format owner/repo"
  type        = string
  default     = "MinoruSoga/AnimalEkarte"
}
