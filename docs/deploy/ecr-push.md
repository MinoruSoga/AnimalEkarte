# ECR Push & ECS Deploy手順

## 前提条件

- AWS CLI設定済み（profile: AnimalEkarte, region: us-east-1）
- Terraform apply完了（ECR, ECS作成済み）
- Docker Desktop起動済み

---

## 環境変数設定

```bash
export AWS_PROFILE=AnimalEkarte
export AWS_REGION=us-east-1
export AWS_ACCOUNT_ID=698109622668
export ECR_REPOSITORY=animalekarte-api
export ECR_REPOSITORY_URL=698109622668.dkr.ecr.us-east-1.amazonaws.com/animalekarte-api
export ECS_CLUSTER=animalekarte-test-cluster
export ECS_SERVICE=animalekarte-test-service
export ALB_URL=http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com
```

---

## 1. ECR認証

```bash
aws ecr get-login-password --region ${AWS_REGION} --profile ${AWS_PROFILE} | \
  docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
```

**成功条件:** `Login Succeeded` が表示される

---

## 2. Docker Image Build

```bash
cd backend
docker build -f Dockerfile.production -t ${ECR_REPOSITORY}:latest .
```

**成功条件:** `Successfully built` + `Successfully tagged` が表示される

**確認:**
```bash
docker images | grep ${ECR_REPOSITORY}
```

---

## 3. Docker Image Tag

```bash
docker tag ${ECR_REPOSITORY}:latest ${ECR_REPOSITORY_URL}:latest
```

**確認:**
```bash
docker images | grep ${ECR_REPOSITORY_URL}
```

---

## 4. ECR Push

```bash
docker push ${ECR_REPOSITORY_URL}:latest
```

**成功条件:**
```
latest: digest: sha256:... size: ...
```

**確認:**
```bash
aws ecr describe-images \
  --repository-name ${ECR_REPOSITORY} \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION}
```

---

## 5. ECS Service 強制再デプロイ

```bash
aws ecs update-service \
  --cluster ${ECS_CLUSTER} \
  --service ${ECS_SERVICE} \
  --force-new-deployment \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION}
```

**成功条件:**
```json
{
  "service": {
    "serviceName": "animalekarte-test-service",
    "desiredCount": 1,
    ...
  }
}
```

---

## 6. デプロイ進捗確認

```bash
# Service状態確認（約2-3分待機）
aws ecs describe-services \
  --cluster ${ECS_CLUSTER} \
  --services ${ECS_SERVICE} \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq '.services[0] | {runningCount, desiredCount, deployments}'
```

**成功条件:**
- `runningCount: 1`
- `desiredCount: 1`
- `deployments` に `PRIMARY` デプロイが存在

---

## 7. 疎通確認

```bash
# Health Check（200 OKを期待）
curl -i ${ALB_URL}/health

# 期待レスポンス
# HTTP/1.1 200 OK
# {
#   "status": "ok",
#   "timestamp": "...",
#   "version": "1.0.0",
#   "message": "Animal Ekarte API is running",
#   "database": "connected"
# }
```

**成功条件:**
- HTTP Status: `200 OK`
- `"status": "ok"`
- `"database": "connected"` または `"disconnected"`（疎通優先のためOK）

---

## デバッグ手順（5xx / Timeout発生時）

### Step 1: ECS Service状態確認

```bash
aws ecs describe-services \
  --cluster ${ECS_CLUSTER} \
  --services ${ECS_SERVICE} \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq '.services[0].events[0:5]'
```

**確認ポイント:**
- `service ... has reached a steady state.` → 正常
- エラーメッセージがある場合は原因調査

---

### Step 2: ECS Task状態確認

```bash
# Running Task一覧取得
TASK_ARN=$(aws ecs list-tasks \
  --cluster ${ECS_CLUSTER} \
  --service-name ${ECS_SERVICE} \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq -r '.taskArns[0]')

echo "Task ARN: ${TASK_ARN}"

# Task詳細確認
aws ecs describe-tasks \
  --cluster ${ECS_CLUSTER} \
  --tasks ${TASK_ARN} \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq '.tasks[0] | {lastStatus, healthStatus, containers: .containers[] | {name, lastStatus, healthStatus}}'
```

**確認ポイント:**
- `lastStatus: "RUNNING"`
- `healthStatus: "HEALTHY"`

---

### Step 3: CloudWatch Logs確認

```bash
# 最新ログ取得
aws logs tail /ecs/animalekarte-test \
  --follow \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION}
```

**確認ポイント:**
- エラーメッセージ
- `server starting` ログの有無
- DB接続エラー

---

### Step 4: ALB Target Health確認

```bash
# Target Group ARN取得
TG_ARN=$(aws elbv2 describe-target-groups \
  --names animalekarte-test-tg \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq -r '.TargetGroups[0].TargetGroupArn')

# Target Health確認
aws elbv2 describe-target-health \
  --target-group-arn ${TG_ARN} \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq '.TargetHealthDescriptions'
```

**成功条件:**
- `State: "healthy"`

**Unhealthy原因:**
- `initial`: まだヘルスチェック中（2-3分待機）
- `unhealthy`: コンテナが/healthに応答していない
  - CloudWatch Logsでアプリケーションエラー確認
  - Security Groupでポート8080が許可されているか確認

---

### Step 5: Security Group確認

```bash
# ECS Security Group確認
aws ec2 describe-security-groups \
  --filters "Name=tag:Name,Values=animalekarte-test-ecs-sg" \
  --profile ${AWS_PROFILE} \
  --region ${AWS_REGION} | \
  jq '.SecurityGroups[0].IpPermissions'
```

**確認ポイント:**
- Port 8080がALB Security Groupから許可されているか

---

## よくあるエラーと対処法

| エラー | 原因 | 対処法 |
|--------|------|--------|
| `No space left on device` | Dockerディスク容量不足 | `docker system prune -a` |
| `CannotPullContainerError` | ECRにイメージがない | Step 2-4を再実行 |
| `Target.FailedHealthChecks` | /healthが応答しない | CloudWatch Logsでアプリケーションエラー確認 |
| `ResourceInitializationError` | Task Definitionの設定誤り | メモリ/CPU不足、環境変数確認 |
| `503 Service Unavailable` | ALB → ECS疎通不可 | Security Group確認 |

---

## クイックリファレンス

```bash
# 全手順を1コマンドで実行
cd backend && \
  docker build -f Dockerfile.production -t ${ECR_REPOSITORY}:latest . && \
  docker tag ${ECR_REPOSITORY}:latest ${ECR_REPOSITORY_URL}:latest && \
  docker push ${ECR_REPOSITORY_URL}:latest && \
  aws ecs update-service \
    --cluster ${ECS_CLUSTER} \
    --service ${ECS_SERVICE} \
    --force-new-deployment \
    --profile ${AWS_PROFILE} \
    --region ${AWS_REGION} && \
  echo "Waiting for deployment (60s)..." && \
  sleep 60 && \
  curl -i ${ALB_URL}/health
```

---

## 完了条件

✅ ECRにイメージがpushされている
✅ ECS Serviceが `RUNNING` 状態
✅ Target Healthが `healthy`
✅ `curl http://ALB_URL/health` が200 OKを返す
