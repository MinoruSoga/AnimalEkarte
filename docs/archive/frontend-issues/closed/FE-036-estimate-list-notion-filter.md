# FE-036: 見積一覧の NotionFilter 移行

**親タスク**: [TASK-009](../../docs/tasks/open/TASK-009-notion-filter-migration-and-size-fix.md)
**ステータス**: Open
**作成日**: 2026-03-18

---

## 概要

見積書管理ページ（/estimates）をNotionFilterコンポーネントに移行する。現在はタブフィルタ（すべて/下書き/送付済み/承認済み/却下）+ 検索アイコンのみの独自UIになっている。

## 現状

```
見積書管理
├── タブフィルタ: すべて / 下書き / 送付済み / 承認済み / 却下
├── 検索アイコン（右上）
└── テーブル: 見積番号 / タイトル / 飼主名 / 有効期限 / 合計金額 / ステータス / 操作
```

## 変更後

```
見積書管理
├── NotionFilter ツールバー
│   ├── N件
│   ├── + フィルタを追加（ステータス / 有効期限 プロパティ）
│   ├── 並べ替え
│   └── 検索トグル
└── テーブル: 見積番号 / タイトル / 飼主名 / 有効期限 / 合計金額 / ステータス / 操作
```

## フィルタプロパティ定義

| プロパティ名 | タイプ | 値 |
|-------------|--------|-----|
| ステータス | `select` | 下書き / 送付済み / 承認済み / 却下 |
| 有効期限 | `date-range` | 日付範囲 |

## 対象ファイル

- `frontend/src/features/estimates/routes/EstimateList.tsx`

## 実装方針

- 他の一覧ページ（飼主一覧、会計一覧等）の NotionFilter 実装パターンを踏襲
- タブフィルタを廃止し、NotionFilterの「ステータス」フィルタプロパティに置換
- `useDeferredValue` で検索フィルタを遅延
- Vercel React Best Practices 準拠

## 受入条件

- [ ] NotionFilterツールバー表示（フィルタを追加 + 並べ替え + 検索トグル）
- [ ] ステータスフィルタで絞り込み可能
- [ ] 有効期限の日付範囲フィルタで絞り込み可能
- [ ] 並べ替え機能が動作
- [ ] 検索トグルで検索入力が展開/折りたたみ
- [ ] `docker compose exec frontend pnpm lint` エラーなし
- [ ] `docker compose exec frontend pnpm build` 成功
