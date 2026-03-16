variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for DB subnet group"
  type        = list(string)
}

variable "rds_sg_id" {
  description = "Security group ID for RDS"
  type        = string
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro"
}

variable "allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 20
}

variable "backup_retention_period" {
  description = "Backup retention period in days"
  type        = number
  default     = 1
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_username" {
  description = "Database master username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for RDS (when using public access)"
  type        = list(string)
  default     = []
}

variable "use_public_access" {
  description = "Whether to deploy RDS in public subnet with public access"
  type        = bool
  default     = false
}
