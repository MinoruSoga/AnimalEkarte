#!/usr/bin/env sh
# infra-terraform-plan-preflight.sh
#
# STG P2 full plan の preflight。秘密値を一切読まずに前提条件を検証し、
# terraform.tfvars が存在する場合のみ validate -> full plan(-out=tfplan) を実行する。
#
# 設計方針:
#   - terraform.tfvars は「存在」だけを stat で確認し、内容は読まない/表示しない。
#   - tfvars 不在時は full plan を実行せず BLOCKED で即終了する（秘密値の取得も行わない）。
#   - AWS_PROFILE=AnimalEkarte を明示する（infra/CLAUDE.md Terraform 安全ルール #1）。
#   - validate は plan より前に必ず実行する。
#   - 対象を絞らない full plan のみを生成する。scoped target plan は使わない。
#   - dummy 秘密値や TF_VAR 注入で plan を強制しない。
#   - 適用ステップは承認後に承認者が手動実行する。このスクリプトは plan までで止まる。
#
# 終了コード:
#   0  plan 生成成功（承認者の差分レビュー段階へ）
#   1  validate または plan が失敗
#   2  BLOCKED: terraform ディレクトリが見つからない
#   3  BLOCKED: terraform.tfvars 不在（plan 未実行・秘密値未読込）
#   4  BLOCKED: terraform 未初期化（.terraform 不在 -> init が必要）
#   5  BLOCKED: terraform コマンドが PATH に無い

set -eu

PROFILE="AnimalEkarte"
PLAN_OUT="tfplan"

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
DEFAULT_TF_DIR="$SCRIPT_DIR/../infra/terraform"
TF_DIR="${TF_DIR:-$DEFAULT_TF_DIR}"

log()  { printf '%s\n' "[preflight] $1"; }
fail() { printf '%s\n' "[preflight] $1" >&2; }

# --- AWS profile を明示（CLI 認証 + 別アカウント誤操作防止）---
export AWS_PROFILE="$PROFILE"
log "AWS_PROFILE=$AWS_PROFILE"

# --- terraform ディレクトリ確認 ---
if [ ! -d "$TF_DIR" ]; then
  fail "BLOCKED: terraform ディレクトリが見つからない: $TF_DIR"
  exit 2
fi
TF_DIR=$(cd "$TF_DIR" && pwd)

# --- tfvars の存在のみ確認（内容は読まない・最初に評価する純ローカルチェック）---
TFVARS="$TF_DIR/terraform.tfvars"
if [ ! -f "$TFVARS" ]; then
  fail "BLOCKED: terraform.tfvars missing ($TFVARS)"
  fail "full plan は実行しない。db_password と alb_internal=true を含む実 tfvars をローカルに用意せよ。"
  fail "手順は docs/ops/p2-terraform-plan-runbook.md を参照。"
  exit 3
fi
log "terraform.tfvars detected (内容は読み取らない)"

# --- terraform バイナリ確認 ---
if ! command -v terraform >/dev/null 2>&1; then
  fail "BLOCKED: terraform コマンドが PATH に無い。terraform をインストールせよ。"
  exit 5
fi

# --- init 済み確認（init は backend/network を触るため自動実行しない）---
if [ ! -d "$TF_DIR/.terraform" ]; then
  fail "BLOCKED: terraform 未初期化（.terraform 不在）。次を手動実行せよ:"
  fail "  (cd \"$TF_DIR\" && AWS_PROFILE=$PROFILE terraform init)"
  exit 4
fi

# --- validate を plan より前に必ず実行 ---
log "terraform validate を実行..."
if ! terraform -chdir="$TF_DIR" validate; then
  fail "BLOCKED: terraform validate 失敗。plan は実行しない。"
  exit 1
fi
log "validate OK"

# --- 対象を絞らない full plan を out 付きで生成 ---
log "full plan を実行: terraform plan -out=$PLAN_OUT"
if ! terraform -chdir="$TF_DIR" plan -out="$PLAN_OUT"; then
  fail "BLOCKED: terraform plan 失敗。"
  exit 1
fi

log "plan 生成完了: $TF_DIR/$PLAN_OUT"
log "次のアクション: 承認者が差分をレビュー（runbook の expected/unexpected 表）-> 承認後に手動で適用する。"
log "承認判断は full plan の結果のみで行う。"
exit 0
