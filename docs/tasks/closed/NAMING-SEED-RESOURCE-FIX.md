# NAMING-SEED-RESOURCE-FIX: Seed データの resource 値が旧名

- **作成日**: 2026-04-12
- **ステータス**: CLOSED
- **重大度**: **CRITICAL** — 権限チェック不一致でアクセス制御が壊れる

---

## 概要

`003_seed_demo.sql` の `permission_group_rules` テーブルの `resource` カラム値が旧名のまま。
Go model の `ResourceMasterReservationType` 定数は `"master-reservation-type"` に変更済みだが、
seed データは `"master-reservation-category"` のまま。DB リセット後に権限チェックが不一致になる。

---

## 検出結果

### 003_seed_demo.sql — 6件

| 行番号 | 旧値 | 新値 |
|--------|------|------|
| 261 | `'master-reservation-category'` | `'master-reservation-type'` |
| 285 | `'master-reservation-category'` | `'master-reservation-type'` |
| 309 | `'master-reservation-category'` | `'master-reservation-type'` |
| 333 | `'master-reservation-category'` | `'master-reservation-type'` |
| 357 | `'master-reservation-category'` | `'master-reservation-type'` |
| 381 | `'master-reservation-category'` | `'master-reservation-type'` |

### models.ts source コメント — 2件（INFO、実害なし）

- `models.ts:1211` — `// source: medical_record_image.go`（旧ファイル名だが codegen が生成したもの。実害なし）
- `models.ts:1362` — `// source: reservation_customer.go`（同上）

---

## 修正方針

1. `003_seed_demo.sql`: `'master-reservation-category'` → `'master-reservation-type'`（6箇所）
2. models.ts: codegen 再実行で自動更新（手動修正不要）
