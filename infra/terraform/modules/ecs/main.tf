# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.name_prefix}-alb"
  internal           = var.alb_internal
  load_balancer_type = "application"
  security_groups    = [var.alb_sg_id]
  subnets            = var.alb_internal ? var.private_subnet_ids : var.public_subnet_ids

  enable_deletion_protection = false

  tags = {
    Name = "${var.name_prefix}-alb"
  }
}

# Target Group
resource "aws_lb_target_group" "main" {
  name        = "${var.name_prefix}-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200-399"
  }

  deregistration_delay = 30

  tags = {
    Name = "${var.name_prefix}-tg"
  }
}

# HTTPS Listener (only created when certificate ARN is provided)
resource "aws_lb_listener" "https" {
  count             = var.alb_certificate_arn != "" ? 1 : 0
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-2017-01"
  certificate_arn   = var.alb_certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }
}

# HTTP Listener - Forward to target group
# CloudFront がオリジンに HTTP で接続するため forward を使用
# 独自ドメイン取得後は redirect に変更する
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }
}

# P2: CloudFront VPC Origin — internal ALB を CloudFront バックボーン経由で接続
# NOTE: CloudFront distribution 自体は Terraform 管理外（手動作成済み）。
#       この resource apply 後に手動で distribution のオリジンを VPC Origin に切り替える。
# NOTE: apply 後、AWS が CloudFront-VPCOrigins-Service-SG を自動生成する。
#       Phase 2 SG 絞り込みは docs/infra/STG_AWS_CHANGE_READINESS.md §3.2 を参照。
resource "aws_cloudfront_vpc_origin" "alb" {
  count = var.alb_internal ? 1 : 0

  vpc_origin_endpoint_config {
    name                   = "${var.name_prefix}-vpc-origin"
    arn                    = aws_lb.main.arn
    http_port              = 80
    https_port             = 443
    origin_protocol_policy = "http-only"
    origin_ssl_protocols {
      items    = ["TLSv1.2"]
      quantity = 1
    }
  }

  tags = {
    Name = "${var.name_prefix}-vpc-origin"
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${var.name_prefix}-cluster"

  setting {
    name = "containerInsights"
    # コスト最適化: STG では Container Insights のカスタムメトリクス（CW:MetricMonitorUsage 約 $8/月）が不要。
    # CloudWatch Logs（無料枠内）は task definition 側で維持する。
    value = "disabled"
  }

  tags = {
    Name = "${var.name_prefix}-cluster"
  }
}

# Capacity Providers — STG コスト最適化: Fargate Spot を既定にし同一サイズで単価約 -70%。
# Spot は 2 分前通知で中断・自動再配置するが STG では許容。
resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name       = aws_ecs_cluster.main.name
  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}

# IAM Role for ECS Task Execution
resource "aws_iam_role" "task_execution" {
  name = "${var.name_prefix}-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.name_prefix}-task-execution-role"
  }
}

# Attach AmazonECSTaskExecutionRolePolicy
resource "aws_iam_role_policy_attachment" "task_execution" {
  role       = aws_iam_role.task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# H-5: containerDefinitions[*].secrets の valueFrom（SSM Parameter Store）を ECS Agent が
# タスク起動時に解決するための権限。AmazonECSTaskExecutionRolePolicy には ssm:GetParameters が
# 含まれないため個別付与する。パラメータ名は infra/CLAUDE.md の既存命名（/animalekarte/<env>/...）
# に合わせ、name_prefix（例: animalekarte-stg）から環境名を導出してスコープを絞る。
# SecureString は既定で AWS 管理キー（alias/aws/ssm）を使うため kms:Decrypt も合わせて付与する。
# 適用（terraform apply）は SSM パラメータ実登録（H-5 ユーザー実施）とセットで行うこと。
locals {
  ssm_parameter_env  = trimprefix(var.name_prefix, "animalekarte-")
  ssm_parameter_path = "/animalekarte/${local.ssm_parameter_env}"
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role_policy" "task_execution_ssm_secrets" {
  name = "${var.name_prefix}-ecs-task-execution-ssm-secrets"
  role = aws_iam_role.task_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ReadAppSecretParameters"
        Effect = "Allow"
        Action = [
          "ssm:GetParameters",
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter${local.ssm_parameter_path}/*"
      },
      {
        Sid    = "DecryptSecureStringWithDefaultSSMKey"
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "kms:ViaService" = "ssm.${var.aws_region}.amazonaws.com"
          }
        }
      }
    ]
  })
}

# IAM Role for ECS Task
resource "aws_iam_role" "task" {
  name = "${var.name_prefix}-ecs-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.name_prefix}-task-role"
  }
}

# IAM Policy for ECS Exec
resource "aws_iam_role_policy" "ecs_exec" {
  name = "${var.name_prefix}-ecs-exec-policy"
  role = aws_iam_role.task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssmmessages:CreateControlChannel",
          "ssmmessages:CreateDataChannel",
          "ssmmessages:OpenControlChannel",
          "ssmmessages:OpenDataChannel"
        ]
        Resource = "*"
      }
    ]
  })
}

# ECS Task Definition
resource "aws_ecs_task_definition" "main" {
  family                   = "${var.name_prefix}-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([
    {
      name      = "api"
      image     = "${var.ecr_repository_url}:latest"
      essential = true

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      environment = []
      secrets     = []

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = var.cloudwatch_log_group_name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "api"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    }
  ])

  tags = {
    Name = "${var.name_prefix}-task-def"
  }
}

# ECS Service
resource "aws_ecs_service" "main" {
  name            = "${var.name_prefix}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.desired_count

  # コスト最適化: Fargate Spot を強く優先（weight 4）しつつ、Spot 在庫枯渇時は
  # on-demand FARGATE（weight 1）へ fallback。夜間スケジュールの朝 8:00 自動起動が
  # Spot 在庫切れで失敗しないようにするための保険。大半のタスクは Spot で起動する。
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 4
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_sg_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = "api"
    container_port   = 8080
  }

  enable_execute_command = true

  # capacity provider はサービス作成前にクラスタへ関連付けが必要
  depends_on = [aws_lb_listener.http, aws_ecs_cluster_capacity_providers.main]

  lifecycle {
    # desired_count: 夜間スケジューラが 0/1 を制御するため terraform は管理しない。
    # task_definition: deploy パイプライン(.env.staging から env 注入)が管理する。
    #   terraform の task def は env 空のスケルトンのため、terraform に管理させると
    #   service が env 無しタスクで起動して STG が落ちる（2026-06-01 に実際に発生）。
    # capacity_provider_strategy: 変更が service 置換を強制し、置換時に空スケルトン
    #   task def で再作成されて STG が落ちるため、runtime(CLI)管理にして ignore する。
    #   現状 live は FARGATE_SPOT weight4 + FARGATE weight1（Spot 優先 + on-demand fallback）。
    ignore_changes = [desired_count, task_definition, capacity_provider_strategy]
  }

  tags = {
    Name = "${var.name_prefix}-service"
  }
}

