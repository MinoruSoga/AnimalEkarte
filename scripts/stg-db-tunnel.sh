#!/bin/bash
# ステージング RDS への SSM ポートフォワードトンネル
# 使い方: ./scripts/stg-db-tunnel.sh
# TablePlus: 127.0.0.1:15432 / ekarte_admin / TempPass123!ChangeMe / ekarte_db

set -e

CLUSTER="animalekarte-stg-cluster"
REGION="us-east-1"
PROFILE="AnimalEkarte"
DB_HOST="animalekarte-stg-db.cqbe28s44fta.us-east-1.rds.amazonaws.com"
LOCAL_PORT="15432"

echo "ECS タスクを取得中..."

TASK_ARN=$(aws ecs list-tasks \
  --cluster "$CLUSTER" \
  --region "$REGION" \
  --profile "$PROFILE" \
  --output text \
  --query 'taskArns[0]')

if [ -z "$TASK_ARN" ] || [ "$TASK_ARN" = "None" ]; then
  echo "ERROR: 実行中のタスクが見つかりません"
  exit 1
fi

TASK_ID=$(echo "$TASK_ARN" | awk -F'/' '{print $NF}')

RUNTIME_ID=$(aws ecs describe-tasks \
  --cluster "$CLUSTER" \
  --tasks "$TASK_ID" \
  --region "$REGION" \
  --profile "$PROFILE" \
  --query 'tasks[0].containers[0].runtimeId' \
  --output text)

TARGET="ecs:${CLUSTER}_${TASK_ID}_${RUNTIME_ID}"

echo "ターゲット: $TARGET"
echo "トンネル開始: localhost:${LOCAL_PORT} → ${DB_HOST}:5432"
echo "TablePlus 設定: Host=127.0.0.1  Port=${LOCAL_PORT}  User=ekarte_admin  DB=ekarte_db"
echo "終了: Ctrl+C"
echo ""

aws ssm start-session \
  --target "$TARGET" \
  --document-name AWS-StartPortForwardingSessionToRemoteHost \
  --parameters "{\"host\":[\"${DB_HOST}\"],\"portNumber\":[\"5432\"],\"localPortNumber\":[\"${LOCAL_PORT}\"]}" \
  --region "$REGION" \
  --profile "$PROFILE"
