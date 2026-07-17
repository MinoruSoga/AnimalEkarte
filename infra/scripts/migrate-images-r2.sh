#!/usr/bin/env bash
# P2-4: S3(臨床画像バケット) → R2(animalekarte-stg-images) データ移行スクリプト雛形。
#
# 【スコープ外】本タスクでは実行しない(AWS 認証情報と実疎通確認済みの R2 S3互換
# 認証情報の両方が必要。現時点では Cloudflare API Token 不在のため R2 側の
# 認証情報も未発行 — infra/cloudflare/r2.tf のコメント参照)。
# 雛形のみ用意し、両方の認証情報が揃った時点で人間がレビューの上実行すること。
#
# 前提:
#   - rclone がインストール済み( brew install rclone )
#   - 環境変数で S3/R2 双方の認証情報を供給する(このファイルに値を書かない):
#       AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_REGION      (S3 側)
#       R2_ACCESS_KEY_ID  / R2_SECRET_ACCESS_KEY  / R2_ACCOUNT_ID   (R2 側)
#   - S3_BUCKET_NAME (移行元。現行 STG の画像バケット名)
#   - R2_BUCKET_NAME (移行先。既定 animalekarte-stg-images。infra/cloudflare/variables.tf と一致させる)

set -euo pipefail

: "${S3_BUCKET_NAME:?S3_BUCKET_NAME を設定してください(移行元 S3 バケット名)}"
: "${R2_BUCKET_NAME:?R2_BUCKET_NAME を設定してください(既定: animalekarte-stg-images)}"
: "${R2_ACCOUNT_ID:?R2_ACCOUNT_ID を設定してください}"
: "${R2_ACCESS_KEY_ID:?R2_ACCESS_KEY_ID を設定してください(R2 S3互換トークン)}"
: "${R2_SECRET_ACCESS_KEY:?R2_SECRET_ACCESS_KEY を設定してください}"

RCLONE_CONFIG_S3_TYPE=s3
RCLONE_CONFIG_S3_PROVIDER=AWS
RCLONE_CONFIG_S3_ENV_AUTH=true
export RCLONE_CONFIG_S3_TYPE RCLONE_CONFIG_S3_PROVIDER RCLONE_CONFIG_S3_ENV_AUTH

RCLONE_CONFIG_R2_TYPE=s3
RCLONE_CONFIG_R2_PROVIDER=Cloudflare
RCLONE_CONFIG_R2_ACCESS_KEY_ID="${R2_ACCESS_KEY_ID}"
RCLONE_CONFIG_R2_SECRET_ACCESS_KEY="${R2_SECRET_ACCESS_KEY}"
RCLONE_CONFIG_R2_ENDPOINT="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"
RCLONE_CONFIG_R2_ACL=private
export RCLONE_CONFIG_R2_TYPE RCLONE_CONFIG_R2_PROVIDER RCLONE_CONFIG_R2_ACCESS_KEY_ID \
  RCLONE_CONFIG_R2_SECRET_ACCESS_KEY RCLONE_CONFIG_R2_ENDPOINT RCLONE_CONFIG_R2_ACL

echo "==> Dry-run: S3:${S3_BUCKET_NAME} -> R2:${R2_BUCKET_NAME}"
rclone sync "S3:${S3_BUCKET_NAME}" "R2:${R2_BUCKET_NAME}" --checksum --dry-run --progress

read -r -p "上記 dry-run 結果を確認しました。実行しますか? (yes/no): " CONFIRM
if [[ "${CONFIRM}" != "yes" ]]; then
  echo "中断しました。"
  exit 1
fi

echo "==> 実行: S3:${S3_BUCKET_NAME} -> R2:${R2_BUCKET_NAME}"
rclone sync "S3:${S3_BUCKET_NAME}" "R2:${R2_BUCKET_NAME}" --checksum --progress

echo "==> P2-5 突合: rclone check"
rclone check "S3:${S3_BUCKET_NAME}" "R2:${R2_BUCKET_NAME}" --checksum
