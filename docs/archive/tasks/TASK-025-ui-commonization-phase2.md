# TASK-025: UI共通化リファクタリング第2弾

**作成日**: 2026-03-25
**ステータス**: Closed
**依頼元**: ボタンやフォームなどをできる限りすべてのページで共通化する

---

## 概要

TASK-024（FormDialog共有化・ステータスカラーマップ統一・EditableTable統合）の継続として、調査で発見した4つの残存重複パターンを解消する。機能変更なし・リファクタリングのみ。

## 依頼内容（原文）

> ボタンやフォームなどってできる限りすべてのページで共通化されてますか？→ yes

## 仕様確認ログ

確認事項なし（リファクタリングのみ・機能変更なし）

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|-----------|------|---------|------|------|
| 1 | `NumberInput` を全 feature に適用（14ファイル） | FE | FE-108 | - | [x] |
| 2 | シンプルダイアログ3件を `FormDialog` に移行 | FE | FE-109 | - | [x] |
| 3 | ボタンカラー直書きを Button variant 拡充で統一 | FE | FE-110 | - | [x] |
| 4 | `DashboardDetailModal` の `STATUS_COLOR` を `status-colors.ts` に統合 | FE | FE-111 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: `grep -rn 'type="number"' frontend/src/features` で残件が 0
- [ ] AC-2: `grep -rln 'from "@/components/ui/dialog"' frontend/src/features` の対象ファイルが FE-109 の3件分だけ減少
- [ ] AC-3: `FormDialog` の保存ボタンに `bg-[#2EAADC]` の直書きが消え、Button variant 経由になっている
- [ ] AC-4: `DashboardDetailModal` に `STATUS_COLOR` のインライン定義が存在しない
- [ ] AC-5: `npm run build` パス・`npm run lint` パス（エラー 0）

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| PetEditModal / ReservationDetailModal / DashboardDetailModal | FormDialog 移行しない | ダイアログ内に複雑なレイアウト・独自 footer ロジックがあるため、FormDialog の単純な save/cancel パターンと合わない | 一律 FormDialog 化 |
| ボタン variant 追加 | `primary` / `danger` variant を Button に追加 | shadcn Button の `className` 直書きより variant 管理の方が将来の色変更が容易 | 全て className で管理 |

## 影響範囲

### Backend
変更なし

### Frontend
- `frontend/src/components/ui/button.tsx` — variant 追加（primary, danger）
- `frontend/src/components/shared/FormDialog/FormDialog.tsx` — 保存ボタンを variant="primary" に変更
- `frontend/src/features/*/` — NumberInput 適用（14ファイル）、FormDialog 移行（3ファイル）、ボタンカラー修正
- `frontend/src/utils/constants/status-colors.ts` — STATUS_COLOR の定義追加

## 参照実装

- `features/hospitalization/components/DailyRecord/VitalDialog.tsx` — NumberInput + FormDialog の正しい使い方
- `features/hospitalization/components/DailyRecord/LogDialog.tsx` — FormDialog の正しい使い方
- `frontend/src/utils/constants/status-colors.ts` — ステータスカラー一元管理の現行パターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| Button variant 追加でスタイル崩れ | 低 | 新 variant は既存 className 指定と競合しないよう設計 |
| NumberInput の onChange シグネチャ差異 | 低 | `e.target.value` (string) → NumberInput の `onChange(value: string)` で同等 |

## 未解決事項

なし

## 実装順序

1. FE-110（Button variant 追加）→ FE-108・FE-109・FE-111 は並列実装可

## 関連イシュー

- FE-108: [NumberInput 全 feature 適用](../../frontend/issues/open/FE-108-number-input-all-features.md)
- FE-109: [シンプルダイアログ FormDialog 移行](../../frontend/issues/open/FE-109-simple-dialogs-form-dialog.md)
- FE-110: [ボタン variant 拡充・カラー直書き統一](../../frontend/issues/open/FE-110-button-variant-unification.md)
- FE-111: [DashboardDetailModal STATUS_COLOR を status-colors.ts に統合](../../frontend/issues/open/FE-111-dashboard-status-color-consolidation.md)
