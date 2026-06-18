# P2: ALB SG inbound の許可元を alb_internal に連動させる。
#   - alb_internal=true（internal ALB + CloudFront VPC Origin）:
#       VPC Origin 経由の CloudFront トラフィックは VPC Origins サービスが VPC subnet に配置する ENI 経由で
#       到達するため source は VPC CIDR 内。許可元を var.vpc_cidr に絞る。
#       public CloudFront managed prefix list は internal ALB では使わない
#       （managed prefix list は SG ルール上限に対し重み ~45/参照 を消費し、80/443 の 2 参照で既定 quota 60 を
#        超過する。VPC CIDR ベースは重み 2 で quota 内に収まる）。
#   - alb_internal=false（internet-facing ALB / 現行デフォルト）:
#       CloudFront は public 経由でオリジン接続するため source は VPC 外。0.0.0.0/0 を維持し現行の
#       internet-facing 経路を壊さない（VPC CIDR に絞ると CloudFront → ALB が全断する）。
# Phase 2: internal ALB をさらに絞り込む場合は、aws_cloudfront_vpc_origin apply 後に AWS が自動生成する
# VPC Origins サービス管理 SG を参照する ingress rule に差し替える（名称は VPC Origin 作成後に確認）:
#   aws --profile AnimalEkarte --region us-east-1 ec2 describe-security-groups \
#     --filters "Name=description,Values=*VPCOrigins*" --query 'SecurityGroups[*].{Name:GroupName,ID:GroupId}'
locals {
  # alb_internal=true で VPC CIDR に絞り、false で internet-facing の 0.0.0.0/0 を維持する。
  alb_ingress_cidrs = var.alb_internal ? [var.vpc_cidr] : ["0.0.0.0/0"]
  alb_ingress_desc  = var.alb_internal ? "from VPC CIDR (internal ALB via CloudFront VPC Origin ENI)" : "from anywhere (internet-facing ALB via CloudFront public origin)"
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
