# Tasks

## 概要

このディレクトリのタスクドキュメントは `/task-create` コマンド（`.claude/skills/task-create/`）により自動生成される。

## 生成フロー

```
タスク依頼（ユーザー）
    ↓ /task-create
docs/tasks/open/TASK-XXX-*.md     ← ここ（全体像・依存関係・影響範囲）
    ↓ 領域別に分割
backend/issues/open/BE-XXX-*.md   ← バックエンドイシュー
frontend/issues/open/FE-XXX-*.md  ← フロントエンドイシュー
```

## ディレクトリ構造

```
docs/tasks/
├── CLAUDE.md    ← このファイル
├── open/        ← 未完了タスク
└── closed/      ← 完了タスク
```

## 採番ルール

- 形式: `TASK-XXX-kebab-case-title.md`（XXX は3桁ゼロ埋め）
- 番号: `open/` と `closed/` の両方から最大番号 + 1 で採番

## タスクのライフサイクル

1. `/task-create` で `open/` に生成
2. 配下の全 BE/FE イシューが closed になったら `closed/` に移動

## 役割

タスクドキュメントは **親チケット** に相当する。1つのタスクから複数の BE/FE イシューが派生する。

- タスク = 「何をやるか」の全体像
- イシュー = 「どう実装するか」のAI実行可能な粒度

## 関連ドキュメント

- バックエンドイシュー: `backend/issues/`
- フロントエンドイシュー: `frontend/issues/`
- スキル定義: `.claude/skills/task-create/SKILL.md`
