# FE-025: 全一覧ページの NotionFilter 呼び出しを新 API に対応

**Status**: Open
**Priority**: Medium
**Affects**: 全一覧ページ（8ページ + 見積 + マスタ13画面）
**Date Created**: 2026-03-17
**Related**: TASK-006, FE-023, FE-024

## Summary

FE-023（フィルタ条件・AND/OR）と FE-024（ソート・検索トグル）で拡張された NotionFilter の新 API に合わせて、全一覧ページの呼び出しを更新する。`ActiveFilter` に `condition` フィールド追加、`sortProperties` / `activeSorts` 追加、`useTableSort` の NotionFilter 統合。

## 必要な変更

### 各ページの更新内容

1. **ActiveFilter に condition を追加**: フィルタ適用ロジックで condition を考慮する
2. **sortProperties を定義**: 各ページでソート可能なプロパティを定義
3. **useTableSort を NotionFilter に統合**: 個別の useTableSort hook を削除し、NotionFilter の onSortChange に統合
4. **filterLogic (AND/OR) の state 追加**: デフォルトは "and"

### ページ別作業量

| ページ | フィルタ条件対応 | ソート統合 | 作業量 |
|--------|----------------|----------|--------|
| 飼主一覧 | ✅ | useTableSort → NotionFilter | 中 |
| カルテ一覧 | ✅ | useTableSort → NotionFilter | 中 |
| 会計一覧 | ✅（ステータス） | なし → 追加 | 小 |
| 検査管理 | ✅（日付+ステータス） | なし → 追加 | 小 |
| 予防接種 | ✅（日付） | なし → 追加 | 小 |
| 入院 | ✅（ステータス） | なし → 追加 | 小 |
| トリミング | ✅（日付） | なし → 追加 | 小 |
| 在庫 | ✅（カテゴリ+ステータス） | なし → 追加 | 小 |
| 予約管理 | ✅（担当医） | カレンダーUI特有 | 小 |

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useDeferredValue` でテキスト検索遅延（既存パターン維持）

## 依存関係

- **FE-023** と **FE-024** が先に完了している必要がある

## 完了条件

- [ ] 全ページで新 API（condition + filterLogic + sortProperties）が使用されている
- [ ] useTableSort の個別実装が NotionFilter に統合されている
- [ ] フィルタ条件が各ページで正しく動作する
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
