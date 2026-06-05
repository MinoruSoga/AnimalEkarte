# コスト最適化（STG）: EventBridge Scheduler のネイティブ API ターゲット（Lambda 不要）で
# ECS service の desired_count と RDS instance を毎日スケジュール起動/停止する。

# EventBridge Scheduler が AWS API を呼ぶための IAM ロール
resource "aws_iam_role" "scheduler" {
  count = var.enabled ? 1 : 0
  name  = "${var.name_prefix}-scheduler-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "scheduler.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = {
    Name = "${var.name_prefix}-scheduler-role"
  }
}

resource "aws_iam_role_policy" "scheduler" {
  count = var.enabled ? 1 : 0
  name  = "${var.name_prefix}-scheduler-policy"
  role  = aws_iam_role.scheduler[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ecs:UpdateService"]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = ["rds:StopDBInstance", "rds:StartDBInstance", "rds:DescribeDBInstances"]
        Resource = "*"
      }
    ]
  })
}

# --- ECS: 起動（desired_count=N）---
resource "aws_scheduler_schedule" "ecs_start" {
  count = var.enabled ? 1 : 0
  name  = "${var.name_prefix}-ecs-start"

  flexible_time_window { mode = "OFF" }
  schedule_expression          = var.ecs_start_cron
  schedule_expression_timezone = var.timezone

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:ecs:updateService"
    role_arn = aws_iam_role.scheduler[0].arn
    input = jsonencode({
      Cluster      = var.ecs_cluster_name
      Service      = var.ecs_service_name
      DesiredCount = var.desired_count
    })
  }
}

# --- ECS: 停止（desired_count=0）---
resource "aws_scheduler_schedule" "ecs_stop" {
  count = var.enabled ? 1 : 0
  name  = "${var.name_prefix}-ecs-stop"

  flexible_time_window { mode = "OFF" }
  schedule_expression          = var.ecs_stop_cron
  schedule_expression_timezone = var.timezone

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:ecs:updateService"
    role_arn = aws_iam_role.scheduler[0].arn
    input = jsonencode({
      Cluster      = var.ecs_cluster_name
      Service      = var.ecs_service_name
      DesiredCount = 0
    })
  }
}

# --- RDS: 起動 ---
resource "aws_scheduler_schedule" "rds_start" {
  count = var.enabled ? 1 : 0
  name  = "${var.name_prefix}-rds-start"

  flexible_time_window { mode = "OFF" }
  schedule_expression          = var.rds_start_cron
  schedule_expression_timezone = var.timezone

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:rds:startDBInstance"
    role_arn = aws_iam_role.scheduler[0].arn
    input    = jsonencode({ DbInstanceIdentifier = var.db_instance_identifier })
  }
}

# --- RDS: 停止 ---
resource "aws_scheduler_schedule" "rds_stop" {
  count = var.enabled ? 1 : 0
  name  = "${var.name_prefix}-rds-stop"

  flexible_time_window { mode = "OFF" }
  schedule_expression          = var.rds_stop_cron
  schedule_expression_timezone = var.timezone

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:rds:stopDBInstance"
    role_arn = aws_iam_role.scheduler[0].arn
    input    = jsonencode({ DbInstanceIdentifier = var.db_instance_identifier })
  }
}
