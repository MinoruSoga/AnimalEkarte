# TASK-023: FE-075〜078 実装不備の修正

**作成日**: 2026-03-19
**ステータス**: Closed
**依頼元**: ブラウザテストで機能不全を確認

---

## 概要

直近で実装した FE-075〜078 をブラウザテストした結果、以下の問題を検出した。API レスポンス形式の不一致と、props 未配線が原因。

## 依頼内容（原文）

> 直近で作成してた仕様変更のタスクですが、実装されていないようです。ブラウザテストして、機能していないものを再度タスク化してください。

## 仕様確認ログ

確認事項なし（コードベース調査とAPIテストで問題を特定済み）

## 検出された問題

### 問題1: merchandise-items API レスポンス形式不一致（FE-075, FE-076 影響）

**根本原因**: バックエンド `merchandise_item_handler.go:34` が `newPaginatedResponse()` を使用し `{ data: [...], total, page, limit }` 形式を返すが、フロントエンドの API hook は `axios.get<MerchandiseItem[]>()` で直接配列を期待。`data.map(transform)` が `TypeError: data.map is not a function` でクラッシュする。

**影響**:
- FE-075: 物販マスタ管理画面 → 一覧が表示されない
- FE-076: 会計物販モーダル → 品目一覧が表示されない

**修正方針**: バックエンドのレスポンス形式を他のマスタ API（cages 等）と統一し、直接配列を返すように変更する。

### 問題2: PetEditModal に onChangeOwner が未配線（FE-078 影響）

**根本原因**: `PetEditModal` に `onChangeOwner` prop を追加したが、呼び出し元（`OwnerForm.tsx`, `OwnersList.tsx`）から prop が渡されていない。`isEdit && onChangeOwner` が常に false となり、飼主変更ボタンが表示されない。

**影響**:
- FE-078: ペット編集モーダルの飼主変更ボタンが表示されない

**修正方針**: `app/pages/OwnerFormPage.tsx` で `onChangeOwner` コールバックを実装し、PetEditModal に注入する（cross-feature import 禁止のため props 注入パターン）。

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | merchandise_item_handler のレスポンスを配列形式に統一 | BE | BE-048 | - | [x] |
| 2 | FE merchandise-items API hookのレスポンスパース修正（BE修正が間に合わない場合のFE側フォールバック） | FE | FE-084 | - | [x] |
| 3 | PetEditModal に onChangeOwner を親から注入 | FE | FE-085 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: `/settings/merchandise-items` にアクセスすると、物販マスタ一覧が正しく表示される（品目名・カテゴリ・単価・税率・ステータス）
- [ ] AC-2: 会計詳細画面で「物販・その他追加」モーダルを開くと、マスタ品目一覧が表示され、クリックで明細に追加される
- [ ] AC-3: ペット編集モーダル（既存ペット）に「飼主変更」ボタンが表示され、クリック → 飼主検索 → 選択 → 確認 → 変更が動作する
- [ ] AC-4: 新規ペット登録時は「飼主変更」ボタンが表示されない

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| merchandise-items レスポンス修正 | BE 側を配列形式に変更 | 他のマスタ API（cages, staffs 等）と統一 | FE 側で PaginatedResponse を unwrap |

## 影響範囲

### Backend
- `backend/internal/handler/merchandise_item_handler.go:34` — `newPaginatedResponse` → 直接配列

### Frontend
- `frontend/src/features/master/api/merchandise-items.ts` — レスポンスパース修正
- `frontend/src/features/accounting/api/get-merchandise-items.ts` — レスポンスパース修正
- `frontend/src/app/pages/OwnerFormPage.tsx` — onChangeOwner 実装・注入
- `frontend/src/features/owners/routes/OwnerForm.tsx` — onChangeOwner prop 受け渡し

## 参照実装

- `backend/internal/handler/cage_handler.go:27` — `c.JSON(http.StatusOK, cages)` 直接配列パターン
- `app/pages/OwnerFormPage.tsx` — cross-feature props 注入パターン

## リスク・懸念事項

特になし

## 未解決事項

なし

## 実装順序

1. BE-048: merchandise_item_handler レスポンス修正
2. FE-084: merchandise-items API hook 修正（BE/FE 両対応）
3. FE-085: PetEditModal onChangeOwner 配線

## 関連イシュー

- BE-048: [merchandise_item_handler レスポンス形式を配列に統一](../../backend/issues/open/BE-048-merchandise-item-response-format.md)
- FE-084: [物販マスタ API hook レスポンスパース修正](../../frontend/issues/open/FE-084-fix-merchandise-items-api-response.md)
- FE-085: [PetEditModal onChangeOwner 配線](../../frontend/issues/open/FE-085-pet-edit-modal-change-owner-wiring.md)
