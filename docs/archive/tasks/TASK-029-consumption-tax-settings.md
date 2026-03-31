# TASK-029: 消費税区分設定の実装

**作成日**: 2026-03-25
\*\*ステータス\*\*: Closed
**依頼元**: 消費税区分の設定実装

---

## 概要

消費税区分（内税・外税・非課税）と税率（10%/8%）をシステム全体で管理できるようにする。
医院全体の税率マスタ定義・商品マスタ別の課税区分設定・会計精算画面での確認・編集機能を実装する。

## 依頼内容（原文）

> 課税に関して、下記の仕様を満たすように、実装したいです。
>
> - 消費税区分の設定
>     - 医院全体
>     - 商品毎
>     - 会計毎（会計精算ページ）
>
> 全体設定
> - 税率設定
>     - 通常課税　10%
>     - 軽減税率　8%
>
> 商品設定
> - 課税区分選択
>     - 内税・外税・非課税
> - 税率選択
>     - 10% or 8%
>
> 会計画面
> - 課税区分　確認編集
> - 税率　確認編集
> - 税額　確認編集

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 全体設定の用途は？ | C: 税率マスタ（「通常課税=10%」「軽減税率=8%」の名前付け定義）。強制適用ではない |
| 2 | 商品設定の対象範囲は？ | 全マスタ商品（consultations, procedures, medicines, merchandise_items 等） |
| 3 | 内税・外税の計算式は？ | 外税=単価×数量×税率、内税=単価×数量×税率÷(1+税率)、非課税=0 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | DB migration + Goモデル変更 + codegen | DB/BE | BE-058 | - | [x] |
| 2 | Clinic 税率設定 API（GET/PATCH） | BE | BE-059 | #1 | [x] |
| 3 | 全マスタ商品 tax_type/tax_rate API 対応 | BE | BE-060 | #1 | [x] |
| 4 | BillingItem/EstimateItem 課税区分 API + 税額計算 | BE | BE-061 | #1 | [x] |
| 5 | 医院設定画面 - 税率マスタ設定 UI | FE | FE-121 | #2 | [x] |
| 6 | マスタ設定画面 - 課税区分・税率 UI | FE | FE-122 | #3 | [x] |
| 7 | 会計精算画面 - 課税区分・税率・税額 確認編集 UI | FE | FE-123 | #4 | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 医院設定 > 税率設定で「通常課税 10%」「軽減税率 8%」が表示・編集できる
- [ ] AC-2: マスタ設定の診察・処置・薬剤・物販 各項目に「課税区分（内税/外税/非課税）」「税率（10%/8%）」が設定・保存できる
- [ ] AC-3: 会計精算画面の明細行ごとに課税区分・税率・税額が表示され、編集可能
- [ ] AC-4: 外税の税額 = 単価 × 数量 × 税率、内税の税額 = 単価 × 数量 × 税率 ÷ (1 + 税率) で正しく計算される
- [ ] AC-5: 非課税の税額 = 0 で表示される
- [ ] AC-6: 既存の billing_items（tax_rate=0.10）のデフォルト値が壊れない（DB migration は後方互換）

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 全体設定の保存先 | `clinics` テーブルに `standard_tax_rate`, `reduced_tax_rate` を追加 | 既存の Clinic 型に追加するのが最小変更。新テーブル不要 | 新テーブル `clinic_tax_settings` |
| TaxType ENUM | DB ENUM `tax_type ('included', 'excluded', 'exempt')` | 型安全、Go/TSに連動 | VARCHAR + CHECK |
| 税額計算の場所 | Backend service 層（`billing_service.go`）で計算 | ビジネスロジックは service に集約 | Frontend で計算 |
| consultations.price カラム名 | price のまま（unit_price に変えない） | スコープ外、既存 API 破壊回避 | unit_price にリネーム |

## 影響範囲

### DB（001_init.sql）
- 新ENUM `tax_type`: `('included', 'excluded', 'exempt')`
- テーブル追加カラム:
  - `clinics`: `standard_tax_rate numeric DEFAULT 0.10`, `reduced_tax_rate numeric DEFAULT 0.08`
  - `consultations`: `tax_type tax_type DEFAULT 'excluded'`, `tax_rate numeric DEFAULT 0.10`
  - `procedures`: `tax_type tax_type DEFAULT 'excluded'`, `tax_rate numeric DEFAULT 0.10`
  - `medicines`: `tax_type tax_type DEFAULT 'excluded'`, `tax_rate numeric DEFAULT 0.10`
  - `hospitalization_plans`: `tax_type tax_type DEFAULT 'excluded'`, `tax_rate numeric DEFAULT 0.10`
  - `merchandise_items`: `tax_type tax_type DEFAULT 'excluded'`（tax_rate は既存）
  - `billing_items`: `tax_type tax_type DEFAULT 'excluded'`（tax_rate は既存）
  - `estimate_items`: `tax_type tax_type DEFAULT 'excluded'`（tax_rate は既存）

### Backend
- `backend/internal/model/clinic.go` — StandardTaxRate, ReducedTaxRate 追加
- `backend/internal/model/consultation.go` — TaxType, TaxRate 追加
- `backend/internal/model/procedure.go` — TaxType, TaxRate 追加
- `backend/internal/model/medicine.go` — TaxType, TaxRate 追加
- `backend/internal/model/hospitalization_plan.go` — TaxType, TaxRate 追加
- `backend/internal/model/merchandise_item.go` — TaxType 追加
- `backend/internal/model/accounting.go` — BillingItem.TaxType, EstimateItem.TaxType 追加
- `backend/internal/handler/clinic_handler.go` — 税率設定 GET/PATCH
- `backend/internal/service/billing_service.go` — 税額計算ロジック追加

### Frontend
- `frontend/src/features/hospital-settings/` — 税率設定 UI
- `frontend/src/features/accounting/` — 会計明細の課税区分 UI
- `frontend/src/features/master-settings/` — 診察・処置・薬剤・物販マスタへの課税区分追加
- `frontend/src/types/generated/models.ts` — `make codegen` で自動更新

## 参照実装

- `features/owners/` — useTransition フォーム、memo 最適化の参照実装
- `features/hospital-settings/routes/ClinicSettings.tsx` — 医院設定フォームの参照実装
- `features/accounting/` — 既存の会計 UI の参照実装

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 既存 billing_items.tax_rate デフォルト値破壊 | 高 | DB 変更は DEFAULT 追加のみ、既存値保持 |
| make codegen 後の型エラー | 中 | codegen 直後に FE 全体の型エラーを確認 |
| hospitalization_plans/trimming_courses 等への影響 | 低 | 今回は consultations/procedures/medicines/merchandise_items のみ対象 |

## 未解決事項

- [ ] trimming_courses（トリミングコースマスタ）も対象か（今回はスコープ外として保留）
- [ ] cages（ケージマスタ）も対象か（今回はスコープ外として保留）
- [ ] 税率が将来変更された場合の既存 billing_items への影響方針

## 実装順序

1. BE-058: DB + モデル変更 + `make codegen`
2. BE-059: Clinic 税率 API
3. BE-060: 全マスタ商品 tax_type API
4. BE-061: BillingItem 課税区分 API + 税額計算
5. FE-121: 医院設定 税率マスタ UI（BE-059 完了後）
6. FE-122: マスタ設定 課税区分 UI（BE-060 完了後）
7. FE-123: 会計精算 課税区分 UI（BE-061 完了後）

## 関連イシュー

- [BE-058](../../backend/issues/open/BE-058-tax-db-migration-and-models.md)
- [BE-059](../../backend/issues/open/BE-059-clinic-tax-rate-settings-api.md)
- [BE-060](../../backend/issues/open/BE-060-master-items-tax-type-api.md)
- [BE-061](../../backend/issues/open/BE-061-billing-item-tax-type-and-calculation.md)
- [FE-121](../../frontend/issues/open/FE-121-hospital-settings-tax-rate-ui.md)
- [FE-122](../../frontend/issues/open/FE-122-master-settings-tax-type-ui.md)
- [FE-123](../../frontend/issues/open/FE-123-accounting-tax-confirm-edit-ui.md)
