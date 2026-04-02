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

