# TASK-022: カルテ編集・ペット編集ページで飼主変更機能を追加

**作成日**: 2026-03-19
**ステータス**: Closed
**依頼元**: ユーザー

---

## 概要

カルテ編集ページ（MedicalRecordForm）のヘッダーとペット編集モーダル（PetEditModal）に、飼主を変更できるモーダルを追加する。

## 依頼内容（原文）

> カルテ編集ページ、ペット編集ページにて、飼主の変更ができるようにして欲しいです。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | カルテ編集ページの対象 | MedicalRecordForm ヘッダーに飼主変更モーダル追加 |
| 2 | ペット編集ページの対象 | PetEditModal に飼主変更機能追加 |
| 3 | UI操作フロー | (A) 「飼主変更」ボタン → モーダル → 飼主検索・選択 → 確定 |
| 4 | カルテ飼主変更時のペット連動 | (A) owner_id だけ変更（ペットはそのまま） |
| 5 | 確認ダイアログ | デフォルト: 出す |
| 6 | BE バリデーション追加 | デフォルト: 追加する |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | medical_record_service.Update に owner_id/pet_id バリデーション追加 | BE | BE-047 | - | [x] |
| 2 | カルテ編集ヘッダーに飼主変更モーダル追加 | FE | FE-077 | BE-047 | [x] |
| 3 | PetEditModal に飼主変更機能追加 | FE | FE-078 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: カルテ編集ページ（`/medical-records/:id`）で PatientInfoCard の飼主名をクリック → 飼主検索モーダルが開く → 飼主を検索・選択 → 確認ダイアログ「飼主を〇〇に変更します。よろしいですか？」→ 確定 → カルテの owner_id が PATCH API で更新される → ヘッダーの飼主名が即時反映される
- [ ] AC-2: ペット編集モーダル（PetEditModal）で「飼主変更」ボタンをクリック → 飼主検索モーダルが開く → 飼主を選択 → 確認ダイアログ「このペットは飼主Aの管理下から外れます。よろしいですか？」→ 確定 → ペットの owner_id が PATCH API で更新される
- [ ] AC-3: 飼主検索モーダルで飼主名・飼主No・電話番号で検索できる
- [ ] AC-4: カルテの飼主変更時、pet_id は変更されない（owner_id のみ変更）
- [ ] AC-5: BE: medical_record の owner_id 変更時、指定された owner が同一 clinic に所属していない場合は 400 エラーが返る

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 飼主検索モーダル | 共有コンポーネント `OwnerSearchModal` を新規作成 | カルテ編集・ペット編集の両方で使い回すため | 各画面に個別実装 |
| モーダル配置 | `components/shared/OwnerSearchModal/` | feature 間で共有（owners feature に置くと cross-feature import 違反） | features/owners/components/ |
| カルテの飼主更新タイミング | 選択確定時に即時 PATCH API 実行 | カルテの保存ボタンとは独立した操作として扱う | カルテ保存時にまとめて更新 |

## 影響範囲

### DB
- 変更なし（既存スキーマで対応可能）

### Backend
- `backend/internal/service/medical_record_service.go` — Update に owner_id/pet_id バリデーション追加
- `backend/internal/repository/owner_repository.go` — FindByID メソッド（既存・変更不要）

### Frontend
- `frontend/src/components/shared/OwnerSearchModal/` — 新規: 飼主検索・選択モーダル
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` — PatientInfoCard に onOwnerClick + 飼主変更モーダル追加
- `frontend/src/features/medical-records/hooks/use-medical-record-form.ts` — 飼主変更ハンドラ追加
- `frontend/src/features/owners/components/PetEditModal.tsx` — 飼主変更ボタン + モーダル追加

## 参照実装

- `components/shared/PetSelection/` — 検索フォーム + 結果テーブルのモーダル UI パターン
- `features/owners/api/get-owners.ts` — 飼主一覧取得 API
- `features/medical-records/components/StaffSelectionModal.tsx` — モーダル選択 UI パターン
- `PatientInfoCard.tsx:68-76` — 既に `onOwnerClick` prop が定義済み

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| カルテの飼主とペットの飼主が不一致になる可能性 | 中 | ビジネスルールとして許容（譲渡ケース）。UIで注意表示を検討 |
| 飼主変更がカルテの会計・割引率に影響 | 中 | owner_id 変更後に ownerDiscountRate を再取得する |

## 未解決事項

なし

## 実装順序

1. BE-047: medical_record_service に owner_id バリデーション追加
2. FE-077: カルテ編集ヘッダーに飼主変更モーダル（共有 OwnerSearchModal 作成含む）
3. FE-078: PetEditModal に飼主変更機能追加（OwnerSearchModal 再利用）

## 関連イシュー

- BE-047: [medical_record_service に owner_id 変更バリデーション追加](../../backend/issues/open/BE-047-medical-record-owner-validation.md)
- FE-077: [カルテ編集ヘッダーに飼主変更モーダル追加](../../frontend/issues/open/FE-077-karte-change-owner-modal.md)
- FE-078: [PetEditModal に飼主変更機能追加](../../frontend/issues/open/FE-078-pet-edit-change-owner.md)
