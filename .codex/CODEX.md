# Codex Settings

> OpenAI Codex エージェント専用の作業入口。
> 詳細ルールの一次情報は `.claude/CLAUDE.md`を参照する。

## 目的

仕様書と issue を読み、既存コードを踏まえて要件を分解し、BE/FE を分けて実装し、PR とセルフレビューまで閉じる。

## 標準手順

1. 仕様を読む
   - `docs/` の仕様書
   - `docs/work/phase2-deferred.md`（保留事項）
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
- 現在のタスク定義と、必要なら `docs/work/phase2-deferred.md`

担当範囲は backend/ のみです。
...
```

### FE

```text
あなたは FE 担当のコードエージェントです。

まず以下を読んで、仕様とタスクを理解してください。
- docs/<対象機能の仕様書>.md
- backend/docs/api.yaml（API contract 正本）
- 現在のタスク定義と、必要なら `docs/work/phase2-deferred.md`

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
/harness FEAT-123        # タスクIDを指定してハーネスで実装（FEAT-/PERF-/BUG-/SEED- 等。旧 BE-XXX/FE-XXX は docs/archive/ 移設済みで使用しない）
/harness                 # 未コミット変更を規約チェックのみ実行
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

### 規約チェック時の診断コマンド（スコープ限定のみ自律実行可・全体実行は自律実行禁止）

`go vet`・変更ファイル限定の lint/type-check はスコープ限定であれば自律実行してよい。以下の全体実行系コマンドは `.claude/CLAUDE.md` の Auto-Execution Prohibited Commands に該当し、自律実行してはならない（必要な場合はユーザーに実行を依頼する）:

```bash
docker compose exec backend golangci-lint run ./...   # NG: 全体lint。スコープ限定 golangci-lint run ./internal/xxx/... のみ可
docker compose exec backend go test ./...              # NG: 全体テスト
docker compose exec frontend pnpm type-check            # NG: 全体type-check
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
- `.codex/commands/` — 利用可能なコマンド一覧
