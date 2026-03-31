# TASK-024: UI共通化リファクタリング

**作成日**: 2026-03-24
**ステータス**: Closed
**依頼元**: UI共通化リファクタリング: FormDialog共有化・ステータスカラーマップ統一・EditableTable統合

---

## 概要

フロントエンドの各 feature に散在する重複・類似コンポーネントを共通化し、保守性とコードの一貫性を向上させる。純粋なフロントエンドリファクタリングタスク（DB/Backend変更なし）。

## 依頼内容（原文）

> UI共通化リファクタリング: FormDialog共有化・ステータスカラーマップ統一・EditableTable統合

## 仕様確認ログ

確認事項なし（コードベース調査で実装方針を確定）。

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | useSortableData カスタムフック共有化（全リストページ 17+箇所） | FE | FE-101 | - | [x] |
| 2 | FormDialog 共有ラッパー作成（hospitalization Dialog 標準化） | FE | FE-098 | - | [x] |
| 3 | ステータスカラー定数を共有ファイルに集約 | FE | FE-099 | - | [x] |
| 4 | useModalState カスタムフック共有化（削除ダイアログ開閉） | FE | FE-102 | - | [x] |
| 5 | LoadingFallback / ErrorFallback / EmptyStateFallback 共通化 | FE | FE-103 | - | [x] |
| 6 | FilteringIndicator コンポーネント共有化（opacity トランジション） | FE | FE-104 | - | [x] |
| 7 | queryKey ファクトリーパターン導入（キャッシュ無効化の確実性向上） | FE | FE-106 | - | [x] |
| 8 | formatCurrency インライン実装を共有 util へ統一 | FE | FE-105 | - | [x] |
| 9 | 明細行アイテム（LineItem）型・金額計算ロジック共有化 | FE | FE-100 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: hospitalization の LogDialog/VitalDialog/TaskCompleteDialog/CarePlanDialog が共有 `FormDialog` ラッパーを使用しており、フッターのキャンセル・保存ボタンが統一されている
- [ ] AC-2: 予約ステータス色（confirmed, pending, checked_in 等 7種）が `frontend/src/utils/constants/status-colors.ts` に集約されており、`ReservationDetailModal.tsx` のインライン `STATUS_OPTIONS` が削除されている
- [ ] AC-3: visitType（初診/再診）のカラーロジックが `AppointmentCard.tsx` のインラインから共有定数に移行している
- [ ] AC-4: `TreatmentTable.tsx` と `EstimateLineItems.tsx` の金額計算ロジック（小計・割引・合計）が共有ユーティリティ関数として切り出されている
- [ ] AC-5: `npm run lint` と `npm run build` がエラーなしで通過する
- [ ] AC-6: 既存機能（hospitalization 入院管理、reservations 予約、medical-records カルテ、estimates 見積）の動作に変化がない

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| FormDialog の粒度 | ダイアログ共通ラッパー（タイトル + コンテンツ slot + フッター） | hospitalization の各 Dialog はフォームフィールドが異なるため、フッター/ヘッダーパターンのみ共通化 | フォームフィールドまで完全共通化（過度に汎用化） |
| ステータスカラー配置場所 | `frontend/src/utils/constants/status-colors.ts` | 既存の `utils/` 配下の構成に合わせる | `lib/` や `components/shared/` |
| EditableTable の統合方針 | 型定義 + 計算ロジック共有（UI は統合しない） | TreatmentTable（400行・グリッドレイアウト・inline editing）と EstimateLineItems（118行・shadcn Table・read-only）は構造が大きく異なる | UI ごと統合（リスクが高い） |

## 影響範囲

### Backend
なし

### Frontend
- `frontend/src/components/shared/FormDialog/` — **新規作成**
- `frontend/src/utils/constants/status-colors.ts` — **新規作成**
- `frontend/src/features/hospitalization/components/DailyRecord/LogDialog.tsx` — FormDialog 使用に変更
- `frontend/src/features/hospitalization/components/DailyRecord/VitalDialog.tsx` — FormDialog 使用に変更
- `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx` — FormDialog 使用に変更
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx` — FormDialog 使用に変更
- `frontend/src/features/reservations/components/ReservationDetailModal.tsx` — STATUS_OPTIONS を共有定数に移行
- `frontend/src/features/dashboard/components/AppointmentCard.tsx` — visitType カラーを共有定数に移行
- `frontend/src/features/medical-records/components/TreatmentTable.tsx` — 金額計算を共有ユーティリティに移行
- `frontend/src/features/estimates/components/EstimateLineItems/EstimateLineItems.tsx` — 金額計算を共有ユーティリティに移行

## 参照実装

- `frontend/src/components/shared/ConfirmDialog/ConfirmDialog.tsx` — FormDialog の設計参考
- `frontend/src/lib/design-tokens.ts` — カラー定数管理パターン
- `frontend/src/features/hospitalization/styles.ts` — H_STYLES パターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| hospitalization Dialog のリファクタリングで入院管理画面が壊れる | 高 | 変更後に hospitalization の全タブを手動確認 |
| ステータスカラー移行で予約カレンダーの色が変わる | 中 | 変更前後のスクリーンショット比較 |
| TreatmentTable の計算ロジック抽出でカルテ金額計算が狂う | 高 | 単体テスト追加後に移行 |

## 未解決事項

- なし

## 実装順序

1. **FE-101**（useSortableData）— 影響が広く価値が最大、ロジック移植のみで安全
2. **FE-102**（useModalState）— 独立、安全
3. **FE-099**（ステータスカラー定数）— 影響範囲が小さく安全
4. **FE-103**（DataStates）— 独立、安全
5. **FE-104**（FilteringIndicator）— 独立、安全
6. **FE-098**（FormDialog）— hospitalization に影響、手動確認必須
7. **FE-100**（LineItem 計算）— TreatmentTable が複雑なため最後

## 関連イシュー

- FE-098: [FormDialog 共有ラッパー作成](../../frontend/issues/open/FE-098-form-dialog-shared-wrapper.md)
- FE-099: [ステータスカラー定数集約](../../frontend/issues/open/FE-099-status-color-constants.md)
- FE-100: [LineItem 型・金額計算ロジック共有化](../../frontend/issues/open/FE-100-line-item-shared-logic.md)
- FE-101: [useSortableData カスタムフック共有化](../../frontend/issues/open/FE-101-use-sortable-data-hook.md)
- FE-102: [useModalState カスタムフック共有化](../../frontend/issues/open/FE-102-use-modal-state-hook.md)
- FE-103: [LoadingFallback / ErrorFallback / EmptyStateFallback 共通化](../../frontend/issues/open/FE-103-loading-error-empty-states.md)
- FE-104: [FilteringIndicator コンポーネント共有化](../../frontend/issues/open/FE-104-filtering-indicator.md)
