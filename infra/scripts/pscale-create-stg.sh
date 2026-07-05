#!/usr/bin/env bash
# P3-1: PlanetScale Postgres(STG) DB作成手順のスクリプト化。
# 2026-07-05 に本コマンドで animalekarte-stg(main ブランチ, region ap-northeast) を作成済み。
# 再実行は冪等ではない(既に存在する場合はエラーになる)ため、再作成が必要な場合のみ使用する。

set -euo pipefail

ORG="noah-animalekarte"
DB_NAME="animalekarte-stg"
REGION="ap-northeast"
CLUSTER_SIZE="PS-10"

echo "==> pscale auth check"
pscale auth check

echo "==> pscale database create ${DB_NAME}"
pscale database create "${DB_NAME}" \
  --org "${ORG}" \
  --region "${REGION}" \
  --cluster-size "${CLUSTER_SIZE}"

echo "==> pscale database show ${DB_NAME}"
pscale database show "${DB_NAME}" --org "${ORG}"

echo "==> pscale branch list ${DB_NAME}"
pscale branch list "${DB_NAME}" --org "${ORG}"

cat <<'EOS'

次のステップ:
  1. `pscale role reset-default <db> main --org <org> --force` で接続文字列を取得
     (postgres ユーザーの資格情報。値は出力・コミット・ログ保存しない)
  2. backend/migrations/*.sql を DATABASE_URL に対して順に適用
     (backend/cmd/migrate/main.go 相当。advisory lock を使うため直結接続で実行すること。
      Hyperdrive 経由では advisory lock が非対応)
  3. infra/scripts/validate-schema.sql で ENUM / text[] / jsonb / trigger / migration適用状況を検証
EOS
