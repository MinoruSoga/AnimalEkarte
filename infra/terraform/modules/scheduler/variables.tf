variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "enabled" {
  description = "Enable off-hours start/stop schedules for ECS + RDS"
  type        = bool
  default     = false
}

variable "ecs_cluster_name" {
  description = "ECS cluster name to update"
  type        = string
}

variable "ecs_service_name" {
  description = "ECS service name to scale"
  type        = string
}

variable "db_instance_identifier" {
  description = "RDS DB instance identifier to stop/start"
  type        = string
}

variable "desired_count" {
  description = "Desired ECS task count during business hours"
  type        = number
  default     = 1
}

variable "timezone" {
  description = "IANA timezone for the cron schedules"
  type        = string
  default     = "Asia/Tokyo"
}

# 毎日 8:00–22:00 JST 起動。RDS は ECS より先に start / 後に stop する。
variable "rds_start_cron" {
  type    = string
  default = "cron(45 7 * * ? *)" # 07:45 — RDS 起動は数分かかるため先行
}
variable "ecs_start_cron" {
  type    = string
  default = "cron(0 8 * * ? *)" # 08:00
}
variable "ecs_stop_cron" {
  type    = string
  default = "cron(0 22 * * ? *)" # 22:00
}
variable "rds_stop_cron" {
  type    = string
  default = "cron(5 22 * * ? *)" # 22:05 — ECS 停止後に RDS を停止
}
