---
status: closed
closed_at: 2026-03-13
commit: ef82ffa
---

# [型定義] 自動生成型 models.ts への完全移行

## 背景

バックエンド（Go）のモデル struct から TypeScript 型を自動生成するパイプラインを導入した。

```
backend/internal/model/*.go
    ↓ make codegen（tygo）
frontend/src/types/generated/models.ts
```

`make codegen` を実行するだけで、Go モデルの変更が TypeScript 型に自動反映される。
手動で型定義を書く必要はなくなった。

## 対応内容

以下 7 feature の手書き `BackendXxx` interface を `models.ts` 型エイリアスに置換:

| Feature | 変更 |
|---|---|
| `accounting` | `BackendAccounting = Billing`, `BackendAccountingItem = BillingItem` |
| `hospitalization` | `BackendHospitalization = Hospitalization` + enum マッピング追加 |
| `trimming` | `BackendTrimming = TrimmingRecord` + ステータス英語化対応 |
| `reservations` | `BackendReservation = ReservationAppointment` |
| `dashboard` | `BackendDashboardReservation = ReservationAppointment` |
| `medical-records` | `BackendMedicalRecord = MedicalRecord` + inquiry/clinical_plan 参照に変更 |
| `master` | `BackendMasterItem` を削除（旧 STI・死に体コード） |

各 `transforms.ts` も合わせて更新:
- `owner?.name` → `owner?.owner_name`
- `pet?.species` → `pet?.animal_species?.name`
- ID フィールド: `number` → `String(id ?? 0)` 変換
- 英語 enum 値のマッピング追加（"admitted"→"入院中" 等）

## 完了条件

- [x] 全 feature の `api/types.ts` から手書き `BackendXxx` interface が削除されている
- [x] 全て `models.ts` からの import に置き換わっている
- [x] `docker compose exec frontend pnpm build` がエラーゼロで通る
- [x] `make codegen` 実行後もビルドが通る（型の自動更新が機能している）
