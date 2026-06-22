# P2: ALB SG inbound を alb_internal に連動させる。
#   - alb_internal=true（internal ALB + CloudFront VPC Origin）:
#       CloudFront → internal ALB の到達には、AWS が VPC Origin 作成時に自動生成する
#       VPC Origins サービス管理 SG（CloudFront-VPCOrigins-Service-SG）を source 参照する
#       port 80 ingress が【必須】。
#       ※ 当初 VPC CIDR(var.vpc_cidr) 許可で足りると想定したが、live 検証で否定された:
#         VPC CIDR のみだと CloudFront → ALB が全く到達せず（ALB RequestCount=0 / 504）、
#         service-managed SG を source 参照して初めて疎通する。VPC CIDR ingress は VPC 内部
#         呼び出し用に残すが、CloudFront 疎通には service SG ルールが必要。
#       public CloudFront managed prefix list は internal ALB では使わない
#       （managed prefix list は SG ルール上限に対し重み ~45/参照 を消費し、80/443 の 2 参照で
#        既定 quota 60 を超過する）。
#   - alb_internal=false（internet-facing ALB / 現行デフォルト）:
#       CloudFront は public 経由でオリジン接続するため source は VPC 外。0.0.0.0/0 を維持。
#       service SG ルールは作らない（VPC Origin が存在しないため）。
# 依存順: service SG は aws_cloudfront_vpc_origin(ecs module) 作成後に AWS が生成するため、
# 既存環境では下記 data source で解決できる。ゼロからの新規構築時のみ VPC Origin 作成後の
# 2 段階 apply が必要（VPC Origins 採用に伴う既知の chicken-and-egg 制約）。
locals {
  # alb_internal=true で VPC CIDR に絞り、false で internet-facing の 0.0.0.0/0 を維持する。
  alb_ingress_cidrs = var.alb_internal ? [var.vpc_cidr] : ["0.0.0.0/0"]
  alb_ingress_desc  = var.alb_internal ? "from VPC CIDR (internal ALB intra-VPC callers)" : "from anywhere (internet-facing ALB via CloudFront public origin)"
}

# VPC Origins サービス管理 SG（alb_internal=true のときのみ参照）。
# AWS が VPC Origin 作成時に "CloudFront-VPCOrigins-Service-SG" として自動生成する。
# data source は VPC Origin リソースへの依存 edge を持たないため循環依存にならない。
data "aws_security_group" "vpc_origins_service" {
  count  = var.alb_internal ? 1 : 0
  name   = "CloudFront-VPCOrigins-Service-SG"
  vpc_id = var.vpc_id
}

# Security Group for ALB
resource "aws_security_group" "alb" {
  name_prefix = "${var.name_prefix}-alb-sg-"
  description = "Security group for Application Load Balancer"
  vpc_id      = var.vpc_id

  ingress {
    description = "HTTP ${local.alb_ingress_desc}"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = local.alb_ingress_cidrs
  }

  ingress {
    description = "HTTPS ${local.alb_ingress_desc}"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = local.alb_ingress_cidrs
  }

  # CloudFront VPC Origin 経由の到達に必須（alb_internal=true のときのみ）。
  dynamic "ingress" {
    for_each = var.alb_internal ? [1] : []
    content {
      description     = "HTTP from CloudFront VPC Origins service SG"
      from_port       = 80
      to_port         = 80
      protocol        = "tcp"
      security_groups = [data.aws_security_group.vpc_origins_service[0].id]
    }
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name_prefix}-alb-sg"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Security Group for ECS
resource "aws_security_group" "ecs" {
  name_prefix = "${var.name_prefix}-ecs-sg-"
  description = "Security group for ECS tasks"
  vpc_id      = var.vpc_id

  ingress {
    description     = "HTTP from ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name_prefix}-ecs-sg"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Security Group for RDS
resource "aws_security_group" "rds" {
  name_prefix = "${var.name_prefix}-rds-sg-"
  description = "Security group for RDS PostgreSQL"
  vpc_id      = var.vpc_id

  ingress {
    description     = "PostgreSQL from ECS"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name_prefix}-rds-sg"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# CloudWatch Logs Group
resource "aws_cloudwatch_log_group" "ecs" {
  name = "/ecs/${var.name_prefix}"
  # コスト最適化（STG）: 保持を最小化。ログ課金自体は無料枠内で $0 だが PO 方針でログを最小化。
  # ただし完全除去はしない — Phase 4（毎朝の起動時 migrate）失敗を当日中にデバッグするため 1 日は保持する。
  retention_in_days = 1

  tags = {
    Name = "${var.name_prefix}-ecs-logs"
  }
}
