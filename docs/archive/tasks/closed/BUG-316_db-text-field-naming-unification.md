# BUG-316: テキストカラム命名の統一 — memo/notes/remarks/comment 混在

## ステータス: CLOSED — 現状許容（命名規則ドキュメントで明文化済み）

## 概要

フリーテキスト記述に4種類のカラム名（`memo`/`notes`/`remarks`/`comment`）が混在している。
SQL クエリ・Go struct・フロントエンド transform で一貫した命名がなく、
新テーブル追加時にどのカラム名を使うべきか判断できない状態。

---

## 現状の混在状況

| カラム名 | テーブル数 | 主な使用テーブル |
|---------|-----------|----------------|
| `notes`   | 9 | appointments, inquiries, estimates, vital_records, care_plan_items, care_logs, shift_entries, billing_items, care_plan_items |
| `memo`    | 7 | hospitalizations, treatments, treatment_plans, billing_confirmations, billings, 他 |
| `remarks` | 5 | owners, pets, trimming_records, vaccinations, 他 |
| `comment` | 1 | estimates（`notes` と同一テーブルに共存） |

---

## 修正方針

### 統一先: `notes`（最多使用・最も汎用的）

```sql
-- 統一後
memo    → notes   （7テーブル）
remarks → notes   （5テーブル）
comment → notes   （1テーブル）  ← estimates はすでに notes も持つため要精査
```

### 注意事項
- `estimates.comment` と `estimates.notes` が同一テーブルに共存 → 意味が異なる場合は `notes` + `description` 等への分割を検討
- `hospitalizations` の `staff_notes` テーブルは **変更不要**（専用テーブルのため）

---

## 影響範囲

| レイヤー | 件数 | 対象 |
|---------|------|------|
| DB スキーマ (`001_init.sql`) | ~22 カラム | CREATE TABLE 定義 |
| Go Model (`model/*.go`) | 13 ファイル | struct フィールド・json タグ |
| Go Repository | 対応 repo 全件 | GORM マッピング・buildUpdateFields |
| Go Handler | 対応 handler 全件 | request/response struct |
| Frontend transform | ~15 ファイル | `transforms.ts` のマッピング |
| Frontend types | ~15 ファイル | `types/index.ts` インターフェース |
| Frontend components | 多数 | `.remarks` / `.memo` 参照箇所 |

**合計工数**: L（3-4 フェーズに分割推奨）

---

## 実施フェーズ

### Phase 1: DB スキーマ (S)
- `001_init.sql` の `memo`/`remarks`/`comment` を `notes` にリネーム
- `estimates` テーブルの `comment`/`notes` 重複を精査・解消

### Phase 2: Go Model (S)
- `Memo`/`Remarks`/`Comment` struct フィールド → `Notes`
- json タグ更新 → `make codegen` 実行

### Phase 3: Go Repository / Handler (M)
- `buildXxxUpdateFields` の `"memo"/"remarks"/"comment"` キー → `"notes"`
- request/response struct 更新

### Phase 4: Frontend (M)
- `transforms.ts`: `data.memo → data.notes` 等
- `types/index.ts`: インターフェースフィールド更新
- コンポーネント: `.memo`/`.remarks` 参照の更新

---

## 準拠すべきプロジェクト規約

### `.claude/rules/naming-conventions.md` — テキスト列の用途別命名

> | 用途 | カラム名 | 使用場面 |
> | 運用時のフリーテキスト | `memo` | hospitalizations.memo, billings.memo |
> | 補足・備考 | `notes` | appointments.notes, vital_records.notes |
> | 飼主・ペットの備考 | `remarks` | owners.remarks, pets.remarks |

**現状**: 規約にも `memo`/`notes`/`remarks` の使い分けが定義されており、
厳密には完全統一ではなく **ドメイン別の使い分け確認** が先決。
統一先を `notes` に決定する前に、各カラムの意味的な区別を確認すること。

---

## 優先度

**Low** — 機能影響なし。影響範囲が広く段階的対応が必要。
Phase 1 着手前に「`memo`/`notes`/`remarks` を本当に1つに統一するか、
意味的に区別するか」の設計方針を決定すること。

## 関連チケット
- TABLE-NAMING-AUDIT: Phase 4 INFO 項目として検出
- BUG-313/315: 完了済みデザイントークン修正

## 関連ファイル
- `backend/migrations/001_init.sql` — スキーマ定義
- `backend/internal/model/*.go` — Go struct
- `.claude/rules/naming-conventions.md` — カラム命名規約
