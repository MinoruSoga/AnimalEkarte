# STG Backend secrets + redeploy（2026-08-15）

## 実施

| 項目 | 結果 |
|------|------|
| CF Worker secret list（実施前） | SCHEDULER_* **0**（他 14 は既存） |
| put SCHEDULER_* ×7 | **done**（値はログ・repo に未記録） |
| 生成 | `SCHEDULER_OPS_SECRET` · `SCHEDULER_INTERNAL_TOKEN` · `SCHEDULER_ALERT_WEBHOOK_SECRET`（openssl rand） |
| プレースホルダ（要 USER 差し替え） | Access team/audience · Alert host/URL（pending 用） |
| Backend Deploy 再実行 | workflow_dispatch on staging |

## プレースホルダの意味

- **Access**: human JWT ops は未設定扱い。automation は OPS_SECRET 側。
- **Alert**: 実 webhook 未接続。失敗通知は届かない／失敗し得る。API deploy 遮断を優先して仮置き。
- **差し替え**: `cd backend && npx wrangler secret put <NAME>`（値は承認済み store から）

## 禁止

- secret 値を Linear / git / チャットに書かない
- STG reset しない

## Frontend

別件: Vercel `dist` Output Directory。Project Root が `frontend` か要確認。
