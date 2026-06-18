module "vpc" {
  source = "./modules/vpc"

  name_prefix          = var.name_prefix
  vpc_cidr             = var.vpc_cidr
  availability_zones   = var.availability_zones
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  use_nat_instance     = var.use_nat_instance
}

module "security" {
  source = "./modules/security"

  name_prefix  = var.name_prefix
  vpc_id       = module.vpc.vpc_id
  vpc_cidr     = module.vpc.vpc_cidr
  alb_internal = var.alb_internal
}

module "rds" {
  source = "./modules/rds"

  name_prefix             = var.name_prefix
  private_subnet_ids      = module.vpc.private_subnet_ids
  public_subnet_ids       = module.vpc.public_subnet_ids
  rds_sg_id               = module.security.rds_sg_id
  instance_class          = var.rds_instance_class
  allocated_storage       = var.rds_allocated_storage
  backup_retention_period = var.rds_backup_retention_period
  db_name                 = var.db_name
  db_username             = var.db_username
  db_password             = var.db_password
  use_public_access       = var.use_public_rds
  publicly_accessible     = var.rds_publicly_accessible
}

module "ecr" {
  source = "./modules/ecr"

  name_prefix     = var.name_prefix
  repository_name = var.ecr_repository_name
}

module "ecs" {
  source = "./modules/ecs"

  name_prefix               = var.name_prefix
  vpc_id                    = module.vpc.vpc_id
  public_subnet_ids         = module.vpc.public_subnet_ids
  private_subnet_ids        = module.vpc.private_subnet_ids
  alb_sg_id                 = module.security.alb_sg_id
  ecs_sg_id                 = module.security.ecs_sg_id
  ecr_repository_url        = module.ecr.repository_url
  cloudwatch_log_group_name = module.security.cloudwatch_log_group_name
  aws_region                = var.aws_region
  db_address                = module.rds.db_address
  db_port                   = module.rds.db_port
  task_cpu                  = var.ecs_task_cpu
  task_memory               = var.ecs_task_memory
  desired_count             = var.ecs_desired_count
  alb_certificate_arn       = var.alb_certificate_arn
  alb_internal              = var.alb_internal
}

module "github_oidc" {
  source = "./modules/github-oidc"

  name_prefix              = var.name_prefix
  github_repo              = var.github_repo
  task_execution_role_arn  = module.ecs.task_execution_role_arn
  task_role_arn            = module.ecs.task_role_arn
  cloudwatch_log_group_arn = module.security.cloudwatch_log_group_arn
}

# コスト最適化: 夜間/早朝（22:00–8:00 JST）に ECS=0 + RDS stop
module "scheduler" {
  source = "./modules/scheduler"

  name_prefix            = var.name_prefix
  enabled                = var.enable_off_hours_schedule
  ecs_cluster_name       = "${var.name_prefix}-cluster"
  ecs_service_name       = "${var.name_prefix}-service"
  db_instance_identifier = "${var.name_prefix}-db"
  desired_count          = var.ecs_desired_count
}
