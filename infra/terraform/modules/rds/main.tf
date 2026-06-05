# DB Subnet Group - Private (default)
resource "aws_db_subnet_group" "private" {
  count      = var.use_public_access ? 0 : 1
  name       = "${var.name_prefix}-db-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.name_prefix}-db-subnet-group"
  }
}

# DB Subnet Group - Public (for external tools like TablePlus)
resource "aws_db_subnet_group" "public" {
  count      = var.use_public_access ? 1 : 0
  name       = "${var.name_prefix}-db-public-subnet-group"
  subnet_ids = var.public_subnet_ids

  tags = {
    Name = "${var.name_prefix}-db-public-subnet-group"
  }
}

# RDS PostgreSQL Instance
resource "aws_db_instance" "main" {
  identifier     = "${var.name_prefix}-db"
  engine         = "postgres"
  engine_version = "16.4"

  instance_class    = var.instance_class
  allocated_storage = var.allocated_storage
  storage_type      = "gp3"
  storage_encrypted = true
  # コスト最適化 + セキュリティ: public IP を剥がす（subnet group は use_public_access のまま据え置き）。
  publicly_accessible     = var.publicly_accessible
  multi_az                = false
  deletion_protection     = false
  skip_final_snapshot     = true
  backup_retention_period = var.backup_retention_period

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password
  port     = 5432

  vpc_security_group_ids = [var.rds_sg_id]
  db_subnet_group_name   = var.use_public_access ? aws_db_subnet_group.public[0].name : aws_db_subnet_group.private[0].name

  enabled_cloudwatch_logs_exports = ["postgresql"]

  tags = {
    Name = "${var.name_prefix}-db"
  }

  lifecycle {
    # db_subnet_group_name cannot be changed within the same VPC
    ignore_changes = [db_subnet_group_name, identifier]
  }
}
