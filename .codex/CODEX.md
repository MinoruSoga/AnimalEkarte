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
- docs/CUSTOMER_AGGREGATION_SPEC.md
- docs/API_SPEC.md
- docs/tasks/closed/aggregation/00-OVERVIEW.md
- docs/tasks/closed/aggregation/BE.md
- docs/tasks/closed/aggregation/ISSUE-001-be-annual-sales-ranking.md
- docs/tasks/closed/aggregation/ISSUE-002-be-visit-count-aggregation.md
- docs/tasks/closed/aggregation/ISSUE-003-be-last-visit-classification.md
- docs/tasks/closed/aggregation/ISSUE-004-be-search-sort-pagination-filters.md
- docs/tasks/closed/aggregation/ISSUE-005-be-response-tests.md

担当範囲は backend/ のみです。
...
```

### FE

```text
あなたは FE 担当のコードエージェントです。

まず以下を読んで、仕様とタスクを理解してください。
- docs/CUSTOMER_AGGREGATION_SPEC.md
- docs/API_SPEC.md
- docs/tasks/closed/aggregation/00-OVERVIEW.md
- docs/tasks/closed/aggregation/FE.md
- docs/tasks/closed/aggregation/ISSUE-006-fe-dashboard-tabs.md
- docs/tasks/closed/aggregation/ISSUE-007-fe-list-display-switch.md
- docs/tasks/closed/aggregation/ISSUE-008-fe-filters-sort-csv.md
- docs/tasks/closed/aggregation/ISSUE-009-fe-empty-error-loading.md

担当範囲は frontend/ のみです。
...
```

## PR / セルフレビュー

- PR 前に issue の受入条件を全部満たす。
- 変更ファイルがスコープ内に収まっているか確認する。
- 差分・テスト・未解決事項を PR に書く。
- 自己レビューでは「仕様との一致」「命名」「境界値」「空値」「エラー」「ページング」「マルチテナント分離」を見る。

## 参照

- `.claude/CLAUDE.md`
- `docs/AI_DEVELOPMENT_WORKFLOW.md`
