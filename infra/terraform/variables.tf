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

# ALB HTTPS Configuration
variable "alb_certificate_arn" {
  description = "ACM certificate ARN for ALB HTTPS listener"
  type        = string
  default     = ""
}

# GitHub Configuration
variable "github_repo" {
  description = "GitHub repository in format owner/repo"
  type        = string
  default     = "MinoruSoga/AnimalEkarte"
}
