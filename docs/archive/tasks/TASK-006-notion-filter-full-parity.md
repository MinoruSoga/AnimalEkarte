# TASK-006: NotionFilter を Notion 本家のフィルタUIに完全準拠させる

**作成日**: 2026-03-17
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

現在の NotionFilter コンポーネントを、Notion データベースビューのフィルタ/ソート/検索UIに完全準拠させる。参考ページ: https://iced-tibia-111.notion.site/AI-218fa9a287af80a6b619c9db0c3a7b1b のテーブル右側のフィルタ。

## 依頼内容（原文）

> 一覧ページの検索フィルタの機能をNotionと同じにする
>
> 参考Notionページ
> https://iced-tibia-111.notion.site/AI-218fa9a287af80a6b619c9db0c3a7b1b
> このページのAIプロンプトマネージャーテーブルの右側のフィルタと同じにして。

## 現在の実装と Notion の差分

| 機能 | Notion 本家 | 現在の実装 | 対応 |
|------|-----------|-----------|------|
| フィルタ条件（is/is not/contains） | ✅ | ❌ | FE-023 |
| ソート統合（並べ替えボタン） | ✅ | ❌ 各ページ個別 | FE-024 |
| 検索トグル（クリックで展開） | ✅ | ❌ 常時表示 | FE-024 |
| フィルタルール行表示 | ✅ | △ ピル表示 | FE-023 |
| AND/OR 切替 | ✅ | ❌ AND固定 | FE-023 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | NotionFilter にフィルタ条件・ルール行表示・AND/OR 切替を追加 | FE | FE-023 | - |
| 2 | NotionFilter にソート統合・検索トグルを追加 | FE | FE-024 | - |
| 3 | 全一覧ページの NotionFilter 呼び出しを新 API に合わせて更新 | FE | FE-025 | #1, #2 |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend
- `frontend/src/components/shared/NotionFilter/` — コンポーネント拡張
- 全一覧ページの routes/ — NotionFilter props 更新

## 関連イシュー

- [FE-023: フィルタ条件・ルール行表示・AND/OR 切替](../frontend/issues/open/FE-023-notion-filter-conditions-and-or.md)
- [FE-024: ソート統合・検索トグル](../frontend/issues/open/FE-024-notion-filter-sort-search-toggle.md)
- [FE-025: 全一覧ページの新 API 対応](../frontend/issues/open/FE-025-notion-filter-pages-update.md)
