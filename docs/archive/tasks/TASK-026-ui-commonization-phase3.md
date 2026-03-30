# TASK-026: UI共通化リファクタリング第3弾

**作成日**: 2026-03-25
**ステータス**: Closed
**依頼元**: 「ボタンやフォームなどってできる限りすべてのページで共通化されてますか？」

---

## 概要

TASK-024（Phase 1）・TASK-025（Phase 2）に続く第3弾。
削除アイコンボタンの共有コンポーネント化・destructive系ボタンの variant 統一・残存インラインステータスカラーマップの status-helpers.ts への統合を行う。

## 依頼内容（原文）

> UI共通化リファクタリング第3弾: DeleteButton共有コンポーネント作成・destructiveボタン整理・残存ステータスカラーマップ統合

## 仕様確認ログ

確認事項なし（純粋リファクタリング）。

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|-----------|------|---------|------|------|
| 1 | DeleteIconButton 共有コンポーネント作成・10ファイル適用 | FE | FE-112 | - | [x] |
| 2 | ghost-danger variant 追加・4ファイルの outline 赤ボタン整理 | FE | FE-113 | - | [x] |
| 3 | BillingReviewSection の STATUS_BADGE_CLASS を status-helpers.ts に移行 | FE | FE-114 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: `DeleteIconButton` コンポーネントが `components/shared/` に存在し、10ファイルで使用されている
- [ ] AC-2: ゴーストアイコン削除ボタンが全ファイルで同一スタイル（`text-[#37352F]/40 hover:text-red-600 hover:bg-red-50`）になっている
- [ ] AC-3: `button.tsx` に `ghost-danger` variant が追加され、4ファイルの red テキストボタンで使用されている
- [ ] AC-4: `getBillingReviewStatusColor()` が `status-helpers.ts` に存在し、`BillingReviewSection.tsx` がインライン `STATUS_BADGE_CLASS` を持たない
- [ ] AC-5: `npm run lint` + `npm run build` がすべてエラーなし

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| DeleteButton のベース | `Button variant="ghost" size="icon"` + Trash2 | shadcn Button で統一 | native `<button>` + Tailwind |
| outline 赤ボタンの variant | `ghost-danger` + `className="border border-red-200"` | ghost-danger で色を、className で枠線を制御 | `destructive-outline` variant を追加（shadcn 標準からの乖離） |
| BillingReview カラーの移動先 | `status-helpers.ts` | 既存のステータスカラー関数群と統一 | `status-colors.ts`（予約/Dashboard 向けカラーマップとは用途が異なる） |

## 影響範囲

### Frontend
- `frontend/src/components/ui/button.tsx` — `ghost-danger` variant 追加
- `frontend/src/components/shared/DeleteIconButton/` — 新規作成
- `frontend/src/utils/status-helpers.ts` — `getBillingReviewStatusColor` 追加
- 10ファイル — DeleteIconButton 適用
- 4ファイル — ghost-danger variant 適用
- `BillingReviewSection.tsx` — STATUS_BADGE_CLASS 削除・関数利用へ

## 参照実装

- `components/shared/FormDialog/FormDialog.tsx` — 共有コンポーネントの作成パターン
- `utils/status-helpers.ts` — 既存のステータスカラー関数群

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| DeleteIconButton のサイズ差異（h-7〜h-10）| 低 | className prop で上書き可能にする |
| native `<button>` 要素（CheckupsTab/VitalsTab 等）の変換 | 中 | shadcn Button に変換する際の既存 className を確認 |

## 未解決事項

なし

## 実装順序

1. `button.tsx` に `ghost-danger` variant 追加（FE-113 の前提、小変更）
2. `DeleteIconButton` コンポーネント作成（FE-112）
3. 10ファイルに DeleteIconButton 適用（FE-112）
4. 4ファイルの outline 赤ボタンを ghost-danger に変更（FE-113）
5. `getBillingReviewStatusColor` を status-helpers.ts に追加・BillingReviewSection 更新（FE-114）

## 関連イシュー

- FE-112: [DeleteIconButton 共有コンポーネント作成](../../frontend/issues/open/FE-112-delete-icon-button-component.md)
- FE-113: [ghost-danger variant 追加・outline 赤ボタン整理](../../frontend/issues/open/FE-113-ghost-danger-button-variant.md)
- FE-114: [BillingReviewSection STATUS_BADGE_CLASS 統合](../../frontend/issues/open/FE-114-billing-review-status-color-consolidation.md)
