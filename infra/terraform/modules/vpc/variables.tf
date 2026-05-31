variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
}

# コスト最適化(STG): managed NAT Gateway($32.67/月)の代わりに fck-nat EC2(t4g.nano, ~$3/月)を使う。
# true で NAT Gateway を撤去し fck-nat インスタンスへ切替。外向きが壊れたら false に戻すだけでロールバック。
variable "use_nat_instance" {
  description = "Use a fck-nat EC2 instance instead of a managed NAT Gateway"
  type        = bool
  default     = false
}
