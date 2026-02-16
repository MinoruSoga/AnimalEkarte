# ClaudeCode 用 アーキテクチャ理解ドキュメント

## System Summary

Cloud Native Medical Record Platform

## Key Constraints

-   Medical Compliance
-   Auditability
-   Data Durability

## Deployment Flow

Git Push → CI → Terraform → AWS

## Scaling Strategy

Test: Minimum Cost Prod: HA + Multi AZ
