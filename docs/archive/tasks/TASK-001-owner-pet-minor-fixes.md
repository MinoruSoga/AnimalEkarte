# TASK-001: 飼主・ペット 軽微修正

**作成日**: 2026-03-17
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

飼主・ペット機能に対する3件の軽微修正。割引の飼主単位管理と会計画面での表示制御、郵便番号→住所自動入力、動物種類のマスタ設定画面でのCRUD対応。

## 依頼内容（原文）

> 飼主・ペット　軽微修正
> - 割引は飼主単位で設定
>     - カルテの会計では編集不可、表示のみ
> - 飼主登録編集にて、郵便番号から住所を自動入力
> - CFBE(動物種類)のマスタ化

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | カルテ会計画面で飼主割引率を表示のみ（編集不可）に変更 | FE | FE-011 | - |
| 2 | 飼主フォームに郵便番号→住所自動入力を追加 | FE | FE-012 | - |
| 3 | animal_species マスタ CRUD API 追加 | BE | BE-040 | - |
| 4 | マスタ設定画面に動物種類の管理UI追加 | FE | FE-013 | #3 |

## 影響範囲

### DB
- 変更なし（`owners.discount_rate`, `animal_species` テーブルは既存）

### Backend
- `backend/internal/handler/animal_species_handler.go` — Create/Update/Delete/Reorder ハンドラ追加
- `backend/internal/handler/animal_species_request.go` — 新規作成（リクエスト型定義）
- `backend/internal/service/animal_species_service.go` — CRUD メソッド追加
- `backend/internal/repository/animal_species_repository.go` — CRUD メソッド追加

### Frontend
- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx` — 飼主割引率の読み取り専用表示
- `frontend/src/features/medical-records/components/TreatmentDetailedSummary.tsx` — 割引率表示UIの変更
- `frontend/src/features/owners/routes/OwnerForm.tsx` — 郵便番号→住所自動入力
- `frontend/src/features/owners/components/` — 郵便番号検索ボタン or 自動入力コンポーネント
- `frontend/src/features/master/` — 動物種類マスタ管理画面追加

## 実装順序

1. **BE-040**: animal_species CRUD API（Backend）
2. **FE-011**: カルテ会計 割引率表示（Frontend, 独立）
3. **FE-012**: 郵便番号→住所自動入力（Frontend, 独立）
4. **FE-013**: マスタ設定 動物種類管理UI（Frontend, BE-040 依存）

## 関連イシュー

- [BE-040: animal_species マスタ CRUD API](../backend/issues/open/BE-040-animal-species-crud-api.md)
- [FE-011: カルテ会計 飼主割引率 表示のみ](../frontend/issues/open/FE-011-medical-record-billing-owner-discount-readonly.md)
- [FE-012: 飼主フォーム 郵便番号→住所自動入力](../frontend/issues/open/FE-012-owner-form-postal-code-auto-address.md)
- [FE-013: マスタ設定 動物種類管理UI](../frontend/issues/open/FE-013-master-animal-species-crud-ui.md)
