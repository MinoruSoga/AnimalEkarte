---
name: generating-commits
description: Gitコミットメッセージを生成。git commit、コミット作成、変更をコミット時に使用。
---

# Generating Commits

ステージングされた変更からConventional Commits形式のコミットメッセージを生成。

## Workflow

- [ ] `git diff --staged` で変更内容を取得
- [ ] 変更の種類を特定（feat/fix/refactor/docs/style/test/chore）
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
| style | フォーマット変更 |
| test | テスト追加・修正 |
| chore | ビルド・設定変更 |

## Example

```
feat: 在庫一覧にフィルター機能を追加

- カテゴリ別フィルタリング
- 価格帯でのソート機能
- フィルター状態のURL永続化
```
