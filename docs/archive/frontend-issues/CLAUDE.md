# Frontend Issues

## 概要

このディレクトリのイシューは `/task-create` コマンド（`.claude/skills/task-create/`）により自動生成される。

## 生成フロー

```
タスク依頼（ユーザー）
    ↓ /task-create
docs/tasks/TASK-XXX-*.md    ← タスク詳細ドキュメント（全体像・依存関係）
    ↓ 領域別に分割
backend/issues/open/BE-XXX-*.md
frontend/issues/open/FE-XXX-*.md  ← ここ
```

## ディレクトリ構造

```
frontend/issues/
├── CLAUDE.md        ← このファイル
├── open/            ← 未対応・対応中のイシュー
└── closed/          ← 対応完了したイシュー
```

## 採番ルール

- 形式: `FE-XXX-kebab-case-title.md`（XXX は3桁ゼロ埋め）
- 番号: `open/` と `closed/` の両方から最大番号 + 1 で採番
- 手動作成禁止: 必ず `/task-create` 経由で生成すること

## イシューのライフサイクル

1. `/task-create` で `open/` に生成
2. 実装完了後 `closed/` に移動
3. 元タスクドキュメント（`docs/tasks/TASK-XXX-*.md`）のステータスも更新

## 関連ドキュメント

- タスク詳細: `docs/tasks/TASK-XXX-*.md`
- スキル定義: `.claude/skills/task-create/SKILL.md`
