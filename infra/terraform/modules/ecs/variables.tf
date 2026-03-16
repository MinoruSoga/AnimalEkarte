variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for ALB"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "alb_sg_id" {
  description = "Security group ID for ALB"
  type        = string
}

variable "ecs_sg_id" {
  description = "Security group ID for ECS tasks"
  type        = string
}

variable "ecr_repository_url" {
  description = "ECR repository URL"
  type        = string
}

variable "cloudwatch_log_group_name" {
  description = "CloudWatch Logs group name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "db_address" {
  description = "RDS instance address"
  type        = string
}

variable "db_port" {
  description = "RDS instance port"
  type        = number
}

variable "db_name_parameter_arn" {
  description = "SSM Parameter ARN for DB name"
  type        = string
}

variable "db_user_parameter_arn" {
  description = "SSM Parameter ARN for DB user"
  type        = string
}

variable "db_password_parameter_arn" {
  description = "SSM Parameter ARN for DB password"
  type        = string
}

variable "ssm_parameter_arns" {
  description = "List of SSM Parameter ARNs for task execution role"
  type        = list(string)
}

variable "task_cpu" {
  description = "Task CPU units"
  type        = string
  default     = "256"
}

variable "task_memory" {
  description = "Task memory in MB"
  type        = string
  default     = "512"
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 1
}

variable "alb_certificate_arn" {
  description = "ACM certificate ARN for HTTPS listener"
  type        = string
  default     = ""
}

variable "cors_allowed_origin" {
  description = "Comma-separated list of allowed CORS origins"
  type        = string
  default     = "http://localhost:3000"
}

variable "cookie_cross_domain" {
  description = "Enable cross-domain cookie settings (SameSite=None + Secure)"
  type        = bool
  default     = false
}
