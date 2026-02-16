output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "vpc_cidr" {
  description = "VPC CIDR block"
  value       = module.vpc.vpc_cidr
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

output "nat_gateway_id" {
  description = "NAT Gateway ID"
  value       = module.vpc.nat_gateway_id
}

# Security Outputs
output "alb_sg_id" {
  description = "ALB Security Group ID"
  value       = module.security.alb_sg_id
}

output "ecs_sg_id" {
  description = "ECS Security Group ID"
  value       = module.security.ecs_sg_id
}

output "rds_sg_id" {
  description = "RDS Security Group ID"
  value       = module.security.rds_sg_id
}

output "cloudwatch_log_group_name" {
  description = "CloudWatch Logs group name"
  value       = module.security.cloudwatch_log_group_name
}

# RDS Outputs
output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = module.rds.db_endpoint
}

output "rds_address" {
  description = "RDS instance address"
  value       = module.rds.db_address
}

output "rds_port" {
  description = "RDS instance port"
  value       = module.rds.db_port
}

output "rds_db_name" {
  description = "Database name"
  value       = module.rds.db_name
}

output "rds_username" {
  description = "Database username"
  value       = module.rds.db_username
  sensitive   = true
}

# ECR Outputs
output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = module.ecr.repository_url
}

output "ecr_repository_name" {
  description = "ECR repository name"
  value       = module.ecr.repository_name
}

# ECS/ALB Outputs
output "alb_dns_name" {
  description = "ALB DNS name"
  value       = module.ecs.alb_dns_name
}

output "alb_url" {
  description = "ALB URL"
  value       = module.ecs.alb_url
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.ecs_cluster_name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = module.ecs.ecs_service_name
}

output "task_execution_role_arn" {
  description = "Task execution role ARN"
  value       = module.ecs.task_execution_role_arn
}

output "task_role_arn" {
  description = "Task role ARN"
  value       = module.ecs.task_role_arn
}

# GitHub OIDC Outputs
output "github_oidc_provider_arn" {
  description = "GitHub OIDC Provider ARN"
  value       = module.github_oidc.oidc_provider_arn
}

output "github_terraform_role_arn" {
  description = "GitHub Actions Terraform role ARN"
  value       = module.github_oidc.terraform_role_arn
}

output "github_ecs_deploy_role_arn" {
  description = "GitHub Actions ECS deploy role ARN"
  value       = module.github_oidc.ecs_deploy_role_arn
}
