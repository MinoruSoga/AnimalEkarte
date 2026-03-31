---
status: open
---

# [api.yaml] 全スキーマの ID フィールド型が uuid だが実装は uint64（整数）

## 背景

api.yaml を最初に設計した時点では ID を UUID 想定で定義した。
その後バックエンド実装では GORM の `uint64` 連番 ID を採用したが、
api.yaml のスキーマ定義が更新されていない。

## 問題

```yaml
# api.yaml: 全スキーマ共通パターン
id:
  type: string
  format: uuid    # ← 間違い
  readOnly: true
owner_id:
  type: string
  format: uuid    # ← 間違い
```

```go
// 実装: uint64 の連番 ID
type Owner struct {
    ID       uint64 `json:"id"`
    ClinicID uint64 `json:"clinic_id"`
    OwnerID  uint64 `json:"owner_id"`
}
```

**影響範囲**: 全スキーマの `id`, `clinic_id`, `owner_id`, `pet_id`, `doctor_id` 等、
すべての ID フィールドが対象。

**フロントエンドの現状対応**: `frontend/src/types/generated/models.ts` は tygo で
Go モデルから自動生成されるため `number` 型として正しく生成されている。
各 transforms.ts で `String(p.id ?? 0)` 変換しているが、api.yaml を参照する
開発者が混乱する状態になっている。

## 修正方針

全スキーマの ID フィールドの型定義を修正:

```yaml
# 修正後
id:
  type: integer
  format: int64
  readOnly: true
owner_id:
  type: integer
  format: int64
```

対象フィールド（全スキーマ横断）:
- `id`, `clinic_id`, `owner_id`, `pet_id`, `doctor_id`, `staff_id`
- `reservation_appointment_id`, `insurance_id`, `medical_record_id`
- `service_type_id`, `animal_species_id`, `cage_id`
- その他 `_id` サフィックスを持つ全フィールド

## 完了条件

- [ ] 全スキーマの `id` と `*_id` フィールドが `type: integer, format: int64` になっている
- [ ] `format: uuid` の記載が ID フィールドから全て削除されている
- [ ] Swagger UI で確認してスキーマが正しく表示される
