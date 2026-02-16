# Terraform Module 設計書

## Module構成

modules/ vpc/ ecs/ rds/ security/ monitoring/

------------------------------------------------------------------------

## vpc module

責務: - VPC - Subnet - Route Table - NAT Gateway

------------------------------------------------------------------------

## ecs module

責務: - Cluster - Service - Task Definition - Auto Scaling

------------------------------------------------------------------------

## rds module

責務: - PostgreSQL Instance - Parameter Group - Subnet Group - Backup
Policy
