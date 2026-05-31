# VPC
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.name_prefix}-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.name_prefix}-igw"
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.name_prefix}-public-${var.availability_zones[count.index]}"
    Type = "public"
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "${var.name_prefix}-private-${var.availability_zones[count.index]}"
    Type = "private"
  }
}

# Elastic IP for NAT Gateway (use_nat_instance=false 時のみ)
resource "aws_eip" "nat" {
  count  = var.use_nat_instance ? 0 : 1
  domain = "vpc"

  tags = {
    Name = "${var.name_prefix}-nat-eip"
  }

  depends_on = [aws_internet_gateway.main]
}

# NAT Gateway (use_nat_instance=false 時のみ。true では fck-nat インスタンスに置換)
resource "aws_nat_gateway" "main" {
  count         = var.use_nat_instance ? 0 : 1
  allocation_id = aws_eip.nat[0].id
  subnet_id     = aws_subnet.public[0].id

  tags = {
    Name = "${var.name_prefix}-nat"
  }

  depends_on = [aws_internet_gateway.main]
}

# ---- fck-nat: NAT Gateway の代替（コスト最適化, use_nat_instance=true 時のみ）----

# fck-nat 公式 AMI（arm64, owner=fck-nat）
data "aws_ami" "fck_nat" {
  count       = var.use_nat_instance ? 1 : 0
  most_recent = true
  owners      = ["568608671756"]

  filter {
    name   = "name"
    # fck-nat 公式 AMI は Amazon Linux 2023 ベース（fck-nat-al2023-hvm-<ver>-<date>-arm64-ebs）。
    # nat64 variant（fck-nat-nat64-*）はこのパターンに含まれない。most_recent で最新版を採用。
    values = ["fck-nat-al2023-hvm-*-arm64-ebs"]
  }
}

# fck-nat 用 SG: VPC 内（private subnet）からの全通信を許可、外向き全許可
resource "aws_security_group" "fck_nat" {
  count       = var.use_nat_instance ? 1 : 0
  name        = "${var.name_prefix}-fck-nat-sg"
  description = "Allow VPC traffic to be NAT-forwarded"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "All traffic from within VPC"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name_prefix}-fck-nat-sg"
  }
}

# fck-nat インスタンス（t4g.nano, public subnet, source/dest check 無効）
resource "aws_instance" "fck_nat" {
  count                       = var.use_nat_instance ? 1 : 0
  ami                         = data.aws_ami.fck_nat[0].id
  instance_type               = "t4g.nano"
  subnet_id                   = aws_subnet.public[0].id
  vpc_security_group_ids      = [aws_security_group.fck_nat[0].id]
  associate_public_ip_address = true
  # NAT 転送のため必須
  source_dest_check = false

  tags = {
    Name = "${var.name_prefix}-fck-nat"
  }
}

# fck-nat 用 EIP（外向き IP を安定化。LINE 等の IP allowlist 対策）
resource "aws_eip" "fck_nat" {
  count    = var.use_nat_instance ? 1 : 0
  domain   = "vpc"
  instance = aws_instance.fck_nat[0].id

  tags = {
    Name = "${var.name_prefix}-fck-nat-eip"
  }

  depends_on = [aws_internet_gateway.main]
}

# Public Route Table
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.name_prefix}-public-rt"
  }
}

# Public Route Table Association
resource "aws_route_table_association" "public" {
  count = length(aws_subnet.public)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Private Route Table
# 0.0.0.0/0 は inline route で定義（既存 state と整合。standalone aws_route への移行は
# 既存 inline route との RouteAlreadyExists 競合を起こすため避ける）。
# toggle 互換: one() で NAT Gateway / fck-nat ENI の存在する方を指す（片方は null = 省略扱い）。
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block           = "0.0.0.0/0"
    nat_gateway_id       = one(aws_nat_gateway.main[*].id)
    network_interface_id = one(aws_instance.fck_nat[*].primary_network_interface_id)
  }

  tags = {
    Name = "${var.name_prefix}-private-rt"
  }
}

# Private Route Table Association
resource "aws_route_table_association" "private" {
  count = length(aws_subnet.private)

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
