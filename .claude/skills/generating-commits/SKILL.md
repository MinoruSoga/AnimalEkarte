---
name: generating-commits
description: Gitコミットメッセージを生成。git commit、コミット作成、変更をコミット時に使用。
---

# Generating Commits

ステージングされた変更からConventional Commits形式のコミットメッセージを生成。

## Workflow

- [ ] `git diff --staged` で変更内容を取得
- [ ] 変更の種類を特定（feat/fix/refactor/docs/test/chore/perf/ci）
- [ ] 50文字以内の要約を作成
- [ ] 必要に応じて詳細説明を追加

## Format

```
<type>: <summary>

<description>
```

### Types
| Type | 用途 |
|------|------|
| feat | 新機能 |
| fix | バグ修正 |
| refactor | リファクタリング |
| docs | ドキュメント |
| test | テスト追加・修正 |
| chore | ビルド・設定変更 |
| perf | パフォーマンス改善 |
| ci | CI設定 |

## このプロジェクト固有の規約

- **Co-Authored-By 行は付けない**（2026-05-04 確定。グローバル規約より優先）。過去コミットに残る署名は無視してよい
（出典: memory feedback_co_authored_by_drift）

## Example

```
feat: 在庫一覧にフィルター機能を追加

- カテゴリ別フィルタリング
- 価格帯でのソート機能
- フィルター状態のURL永続化
```
