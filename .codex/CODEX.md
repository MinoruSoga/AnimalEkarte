# Codex Settings

> OpenAI Codex エージェント専用の作業入口。
> 詳細ルールの一次情報は `.claude/CLAUDE.md`、標準ワークフローは `docs/AI_DEVELOPMENT_WORKFLOW.md` を参照する。

## 目的

仕様書と issue を読み、既存コードを踏まえて要件を分解し、BE/FE を分けて実装し、PR とセルフレビューまで閉じる。

## 標準手順

1. 仕様を読む
   - `docs/` の仕様書
   - `docs/tasks/open/**/00-OVERVIEW.md`
   - `docs/tasks/open/**/ISSUE-*.md`
2. 既存コードを読む
   - 近い feature
   - handler / service / repository
   - 関連テスト
3. タスクに分解する
   - BE と FE を分ける
   - 依存関係を先に明示する
   - 1 issue = 1 検証責務に絞る
4. 実装用プロンプトを作る
   - 読む文書
   - 担当範囲
   - 触るファイル
   - 触らない範囲
   - 完了時の報告形式
5. エージェントと反復する
   - 返答を仕様と既存コードで照合する
   - ズレがあれば 1 点ずつ修正する
6. PR とセルフレビューで閉じる
   - 変更範囲
   - テスト結果
   - 未解決事項
   - 仕様との整合

## 実装時の指示

- 仕様にない機能は足さない。
- 既存パターンを優先する。
- LTV / LINE / タグ管理の残骸を混ぜない。
- 変更ファイルは最小限にする。
- テストが必要な場合は、何を確認するかを issue 単位で切る。

## プロンプトの基本形

### BE

```text
あなたは BE 担当のコードエージェントです。

まず以下を読んで、仕様とタスクを理解してください。
- docs/<対象機能の仕様書>.md
- backend/docs/api.yaml（API contract 正本）
- docs/tasks/open/<対象タスク>/00-OVERVIEW.md
- docs/tasks/open/<対象タスク>/BE.md
- docs/tasks/open/<対象タスク>/ISSUE-XXX-be-*.md（担当 issue のみ）

担当範囲は backend/ のみです。
...
```

### FE

```text
あなたは FE 担当のコードエージェントです。

まず以下を読んで、仕様とタスクを理解してください。
- docs/<対象機能の仕様書>.md
- backend/docs/api.yaml（API contract 正本）
- docs/tasks/open/<対象タスク>/00-OVERVIEW.md
- docs/tasks/open/<対象タスク>/FE.md
- docs/tasks/open/<対象タスク>/ISSUE-XXX-fe-*.md（担当 issue のみ）

担当範囲は frontend/ のみです。
...
```

## PR / セルフレビュー

- PR 前に issue の受入条件を全部満たす。
- 変更ファイルがスコープ内に収まっているか確認する。
- 差分・テスト・未解決事項を PR に書く。
- 自己レビューでは「仕様との一致」「命名」「境界値」「空値」「エラー」「ページング」「マルチテナント分離」を見る。

## 実装ハーネス（Implement → Verify → Approve Loop）

規約準拠（P1-P18 / React 19パターン）を自動ループで保証する実装フロー。
最大3イテレーションで承認を目指す。

### 使い方

```
/harness BE-042          # バックエンドイシューをハーネスで実装
/harness FE-038          # フロントエンドイシューをハーネスで実装
/harness                 # 未コミット変更を規約チェックのみ実行
/harness-status          # 現在のハーネス進行状態を確認
/harness-status reset    # 状態ファイルをリセット
```

### ループの流れ

```
[Phase 1] タスク分析・影響レイヤー特定
      ↓
[Phase 2] 実装（implementer エージェント）
      ↓
[Phase 3] 規約チェック（go-reviewer / typescript-reviewer / database-reviewer）
      ↓ PASS → 承認レポート出力
      ↓ FAIL → フィードバック付きで Phase 2 に戻る（最大3回）
```

### Codex での診断コマンド（規約チェック時に自律実行）

```bash
docker compose exec backend go vet ./...
docker compose exec backend golangci-lint run ./...
docker compose exec frontend pnpm type-check
```

### いつ使うか

| 状況 | 推奨 |
|------|------|
| イシューが明確でリスクが低い | `/implement` で一気通貫 |
| P1-P18 / React 19 準拠を自動保証したい | `/harness` でループ付き実装 |
| 設計が複雑・複数レイヤーにまたがる | `/plan` → `/harness` |

---

## 参照

- `.claude/CLAUDE.md`
- `docs/AI_DEVELOPMENT_WORKFLOW.md`
- `.codex/commands/` — 利用可能なコマンド一覧
