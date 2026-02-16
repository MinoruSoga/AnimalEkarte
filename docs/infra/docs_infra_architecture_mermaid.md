# AnimalEkarte インフラ構成図（Mermaid）

``` mermaid
flowchart TD

    subgraph Internet
        User[User / Browser]
    end

    subgraph Vercel
        Frontend[Vercel Frontend]
    end

    subgraph AWS us-east-1
        subgraph VPC 10.0.0.0/16
            
            subgraph Public Subnets
                ALB[Application Load Balancer]
                NAT[NAT Gateway]
            end

            subgraph Private Subnets
                ECS[ECS Fargate Service]
                RDS[(RDS PostgreSQL 16)]
            end

        end

        ECR[(ECR Repository)]
        SSM[(SSM Parameter Store)]
        CWL[(CloudWatch Logs)]
    end

    User -->|HTTPS| Frontend
    Frontend -->|HTTP /api| ALB
    ALB -->|HTTP:8080| ECS
    ECS -->|SSL require| RDS
    ECS -->|Pull Image| ECR
    ECS -->|Read Secrets| SSM
    ECS -->|Logs| CWL
    ECS -->|Outbound| NAT
```

## 説明

-   Vercel: フロントエンド配信
-   ALB: Public Subnetに配置されたインターネット向けロードバランサ
-   ECS Fargate: Private Subnet上で稼働するAPIコンテナ
-   RDS PostgreSQL: Private Subnet配置、SSL必須
-   ECR: コンテナイメージ保存
-   SSM: DB接続情報などの安全な管理
-   CloudWatch Logs: アプリログ集約
-   NAT Gateway: Private Subnetからのアウトバウンド通信
