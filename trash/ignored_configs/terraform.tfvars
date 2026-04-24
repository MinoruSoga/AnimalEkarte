# Terraform variables for staging environment

name_prefix = "animalekarte-stg"

# Database Configuration
db_name     = "ekarte_db"
db_username = "ekarte_admin"
db_password = "TempPass123!ChangeMe" # TODO: Change to secure password

# RDS Configuration
rds_instance_class          = "db.t4g.micro"
rds_allocated_storage       = 20
rds_backup_retention_period = 1
use_public_rds              = true  # テスト環境: TablePlus等からの直接接続を許可

# GitHub Configuration
github_repo = "MinoruSoga/AnimalEkarte"
