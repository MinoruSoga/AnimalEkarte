# FE-035: NotionFilter UI要素サイズ2倍化

**親タスク**: [TASK-009](../../docs/tasks/open/TASK-009-notion-filter-migration-and-size-fix.md)
**ステータス**: Open
**作成日**: 2026-03-18

---

## 概要

NotionFilterコンポーネント全体のアイコン・テキスト・ボタンサイズが小さすぎるため、現在の約2倍に拡大する。

## 現状と変更

### アイコンサイズ

| 箇所 | 現在 | 変更後 |
|------|------|--------|
| ツールバーアイコン（Filter/Sort/Search） | `h-3.5 w-3.5`（14px） | `h-5 w-5`（20px） |
| インラインアイコン（Chevron/X/Arrow） | `h-3 w-3`（12px） | `h-4 w-4`（16px） |
| プロパティアイコン（ドロップダウン内） | `h-3.5 w-3.5`（14px） | `h-5 w-5`（20px） |

### テキストサイズ

| 箇所 | 現在 | 変更後 |
|------|------|--------|
| ツールバーラベル（「フィルタを追加」等） | `text-xs`（12px） | `text-sm`（14px） |
| フィルタルール行テキスト | `text-xs`（12px） | `text-sm`（14px） |
| ソートピルテキスト | `text-xs`（12px） | `text-sm`（14px） |
| 件数表示 | `text-xs`（12px） | `text-sm`（14px） |

### ボタン・コンテナサイズ

| 箇所 | 現在 | 変更後 |
|------|------|--------|
| ツールバーボタンパディング | `px-2 py-1` | `px-3 py-1.5` |
| ソートピル高さ | `h-6`（24px） | `h-8`（32px） |
| フィルタピル高さ | `h-6`（24px） | `h-8`（32px） |
| ポップオーバー内アイテム間 | `gap-2` | `gap-3` |

## 対象ファイル

1. `frontend/src/components/shared/NotionFilter/NotionFilter.tsx`
2. `frontend/src/components/shared/NotionFilter/FilterAddPopover.tsx`
3. `frontend/src/components/shared/NotionFilter/FilterRuleRow.tsx`
4. `frontend/src/components/shared/NotionFilter/SortPopover.tsx`
5. `frontend/src/components/shared/NotionFilter/SortPill.tsx`

## 実装方針

- STYLE定数オブジェクトにサイズトークンを集約し、各コンポーネントで参照する
- Tailwind classの直接変更でOK（デザイントークン化は不要）
- 変更後、全一覧ページ（飼主/検査/会計/予防接種/カルテ/入院/トリミング/在庫）で目視確認

## 受入条件

- [ ] ツールバーのアイコン・テキストが現在の約2倍のサイズ
- [ ] フィルタポップオーバー内のプロパティリストが見やすいサイズ
- [ ] ソートピル・フィルタピルが十分な大きさ
- [ ] 既存のフィルタ機能（条件演算子、AND/OR、検索トグル）に影響なし
- [ ] `docker compose exec frontend pnpm lint` エラーなし
- [ ] `docker compose exec frontend pnpm build` 成功
