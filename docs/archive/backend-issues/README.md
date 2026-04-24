# Backend Issues

バックエンドの技術的負債・改善タスクを管理する。

## ディレクトリ構成

```
issues/
├── open/     # 未対応・対応中
└── closed/   # 対応完了
```

## ワークフロー

1. 新規イシューは `open/` に作成
2. 対応完了後、`closed/` に移動
3. ファイル名は `NNN_タイトル.md`（連番）

## テンプレート

```markdown
---
status: open  # または closed
closed_at: YYYY-MM-DD  # closed 時に追記
commit: xxxxxxx  # 対応コミットハッシュ
---

# タイトル

## 背景
...

## 問題
...

## 修正方針
...

## 完了条件
- [ ] ...
```
