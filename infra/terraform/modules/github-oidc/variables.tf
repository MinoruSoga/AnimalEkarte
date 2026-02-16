variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "github_repo" {
  description = "GitHub repository in format owner/repo"
  type        = string
}

variable "task_execution_role_arn" {
  description = "ECS task execution role ARN"
  type        = string
}

variable "task_role_arn" {
  description = "ECS task role ARN"
  type        = string
}

variable "cloudwatch_log_group_arn" {
  description = "CloudWatch Logs group ARN"
  type        = string
}
