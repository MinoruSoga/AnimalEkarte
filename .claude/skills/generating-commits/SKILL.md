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

## 並行セッション環境でのコミット安全則

このリポジトリは複数 Claude セッションが同一 working tree を共有することがある。

- **commit 直前に `git log -1` / `git status` で、自分の add 以降に HEAD が動いていないか確認する**。動いていれば別セッションに割り込まれている（ステージ済みファイルが他セッションのコミットに混入した実例あり）
- **stage は必ず自分の変更パスに限定する**（`git add <明示パス>`）。working tree に無関係な変更が混在している前提で動く
- **doc の「一部だけコミット」指示では `git diff --cached --stat` の規模を想定と比較する**。想定より大きければ無関係な未コミット書き直しが道連れになっている——独断でコミットせずユーザーに確認する
- 別セッションが push 済みの混在コミットは reset/rebase で分離しない（公開履歴の破壊）

（出典: memory feedback_concurrent_session_commit_collision / feedback_doc_commit_structural_bleed_20260709 / be_refactor_execution_20260702）

## Example

```
feat: 在庫一覧にフィルター機能を追加

- カテゴリ別フィルタリング
- 価格帯でのソート機能
- フィルター状態のURL永続化
```
