# AnimalEkarte インフラ構築ドキュメント v2（Test → Production Ready）

## 概要

電子カルテシステム向けクラウドネイティブインフラ設計・構築・運用指針。

------------------------------------------------------------------------

## 技術スタック

### Frontend

-   Next.js（Vercel）

### Backend

-   Golang API
-   ECS Fargate

### Database

-   PostgreSQL (RDS)

### IaC

-   Terraform

### CI/CD

-   GitHub Actions + OIDC

### Cloud

-   AWS

------------------------------------------------------------------------

## 環境構成

### Test

-   Single AZ
-   最小コスト構成

### Production

-   Multi AZ
-   高可用性構成

------------------------------------------------------------------------

## Terraform State

-   S3 Backend
-   DynamoDB Lock

------------------------------------------------------------------------

## セキュリティ基準

-   CloudTrail 必須
-   RDS Encryption
-   IAM Least Privilege
-   OIDC Federation
