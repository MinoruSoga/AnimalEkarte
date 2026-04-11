# NAMING-TABLENAME-DOUBLE-PREFIX: TableName() のプレフィックス二重化バグ

- **作成日**: 2026-04-12
- **ステータス**: OPEN
- **重大度**: **CRITICAL** — DB テーブルが見つからず全 CRUD が失敗する

---

## 概要

sed 一括置換で `reservation_setting` → `line_reservation_setting` を実行した際、
既に `line_reservation_settings` だった TableName() 戻り値にさらに置換が適用され、
`line_line_reservation_settings` という存在しないテーブル名になった。
`medical_record_images` も同様に `medical_medical_record_images` に二重化。

---

## 検出結果

### CRITICAL（2件）

| ファイル | 行 | 現在の値 | 正しい値 |
|---------|-----|---------|---------|
| `backend/internal/model/line_reservation_setting.go:40` | `"line_line_reservation_settings"` | `"line_reservation_settings"` |
| `backend/internal/model/medical_record_image.go:45` | `"medical_medical_record_images"` | `"medical_record_images"` |

### HIGH（ドキュメント更新）

| ファイル | 件数 | 内容 |
|---------|------|------|
| `docs/ERD.md` | 88件 | 旧テーブル名が残存 |
| `backend/docs/api.yaml` | 14件 | 旧テーブル名/エンドポイント名が残存 |

### INFO（変更不要）

| ファイル | 理由 |
|---------|------|
| Go エラーメッセージ `"medical_record_image"` 等 | 新名リソース識別子 |
| models.ts source コメント | codegen 生成 |
| .claude/rules/ 禁止例 | 旧名を禁止例として引用 |

---

## 修正方針

1. `line_reservation_setting.go:40` — `"line_line_reservation_settings"` → `"line_reservation_settings"`
2. `medical_record_image.go:45` — `"medical_medical_record_images"` → `"medical_record_images"`
3. `docs/ERD.md` — 旧テーブル名を新名に更新
4. `backend/docs/api.yaml` — 旧名を新名に更新
