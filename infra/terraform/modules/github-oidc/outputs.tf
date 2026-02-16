output "oidc_provider_arn" {
  description = "GitHub OIDC Provider ARN"
  value       = aws_iam_openid_connect_provider.github.arn
}

output "terraform_role_arn" {
  description = "GitHub Actions Terraform role ARN"
  value       = aws_iam_role.github_terraform.arn
}

output "ecs_deploy_role_arn" {
  description = "GitHub Actions ECS deploy role ARN"
  value       = aws_iam_role.github_ecs_deploy.arn
}
