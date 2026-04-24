---
status: closed
closed_at: 2026-03-16
---

# [master] API queryKey が統一されていない（`["masters", "xxx"]` に統一せよ）

## 優先度
高

## 種別
API設計 / 保守性

## 対象ファイル
- `frontend/src/features/master/api/cages.ts`
- `frontend/src/features/master/api/medicines.ts`
- `frontend/src/features/master/api/consultations.ts`
- `frontend/src/features/master/api/procedures.ts`
- `frontend/src/features/master/api/checkup-types.ts`
- `frontend/src/features/master/api/hospitalization-plans.ts`
- `frontend/src/features/master/api/vaccines-master.ts`

## 問題

queryKey の命名に以下 4 種類が混在しており、`invalidateQueries({ queryKey: ["masters"] })` での一括無効化が機能しない。

| パターン | 使用ファイル |
|---------|------------|
| `["masters", "xxx"]` ← **正** | diagnosis, service-types, trimming, staffs |
| `["masterItems", category]` | get-master-items（汎用） |
| `["xxxMaster"]` | vaccines-master（`["vaccinesMaster"]`）, examinationTypesMaster |
| `["xxx"]` | **cages, medicines, consultations, procedures, checkup-types, hospitalizationPlans** |

## 修正方針

以下のマッピングで queryKey を統一する：

| ファイル | 現状 | 修正後 |
|---------|------|--------|
| cages.ts | `["cages"]` | `["masters", "cages"]` |
| medicines.ts | `["medicines"]` | `["masters", "medicines"]` |
| consultations.ts | `["consultations"]` | `["masters", "consultations"]` |
| procedures.ts | `["procedures"]` | `["masters", "procedures"]` |
| checkup-types.ts | `["checkupTypes"]` | `["masters", "checkup-types"]` |
| hospitalization-plans.ts | `["hospitalizationPlans"]` | `["masters", "hospitalization-plans"]` |
| vaccines-master.ts | `["vaccinesMaster"]` | `["masters", "vaccines"]` |

各ファイルの mutation `onSuccess` 内の `invalidateQueries` のキーも合わせて変更すること。

## 確認事項

`get-master-items.ts` の `["masterItems", category]` パターンは汎用エンドポイント（現在は stub）用であり、
005 で統合された後は廃止対象とする（別イシューで管理）。
