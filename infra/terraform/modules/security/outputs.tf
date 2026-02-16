output "alb_sg_id" {
  description = "ALB Security Group ID"
  value       = aws_security_group.alb.id
}

output "ecs_sg_id" {
  description = "ECS Security Group ID"
  value       = aws_security_group.ecs.id
}

output "rds_sg_id" {
  description = "RDS Security Group ID"
  value       = aws_security_group.rds.id
}

output "cloudwatch_log_group_name" {
  description = "CloudWatch Logs group name for ECS"
  value       = aws_cloudwatch_log_group.ecs.name
}

output "cloudwatch_log_group_arn" {
  description = "CloudWatch Logs group ARN for ECS"
  value       = aws_cloudwatch_log_group.ecs.arn
}

output "db_user_parameter_arn" {
  description = "SSM Parameter ARN for DB user"
  value       = aws_ssm_parameter.db_user.arn
}

output "db_password_parameter_arn" {
  description = "SSM Parameter ARN for DB password"
  value       = aws_ssm_parameter.db_password.arn
}

output "db_name_parameter_arn" {
  description = "SSM Parameter ARN for DB name"
  value       = aws_ssm_parameter.db_name.arn
}
