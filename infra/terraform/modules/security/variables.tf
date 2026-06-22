variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "vpc_cidr" {
  description = "VPC CIDR block. alb_internal=true のとき ALB ingress (80/443) を VPC 内からのみ許可するために使用（VPC Origin 経由の CloudFront トラフィックは VPC Origins ENI 経由で VPC CIDR 内から到達する）。"
  type        = string
}

variable "alb_internal" {
  description = "P2: true で ALB SG inbound を VPC CIDR に絞る（internal ALB + CloudFront VPC Origin）。false は 0.0.0.0/0（internet-facing・現行デフォルト）。ecs module の alb_internal と必ず同値にする。"
  type        = bool
  default     = false
}

