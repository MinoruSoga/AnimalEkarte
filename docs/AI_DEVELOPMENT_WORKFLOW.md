# AI 開発ワークフロー

この文書は、仕様書と issue から要件を読み取り、AI エージェントと協働して実装を進めるための標準手順をまとめたもの。

## 1. 仕様と issue から要件を読み取る

1. `docs/` の仕様書を読む。
2. `docs/tasks/open/**/00-OVERVIEW.md` を読む。
3. `docs/tasks/open/**/ISSUE-*.md` を読む。
4. 受入条件、対象外、命名、検証観点を抽出する。
5. 仕様にないものは実装しない。

### 読み取り時の観点

- 何が「正規名」か。
- 期間、優先順位、境界値、既定値は何か。
- どの画面 / API / DB 層まで影響するか。
- 既存実装を再利用するか、新規に切るか。
- どのテストで担保するか。

## 2. 既存コードを読む

1. 同じドメインの既存 feature を確認する。
2. handler / service / repository の既存パターンを確認する。
3. 関連テストを確認する。
4. 近い命名や型名を流用する。

### 読み方の原則

- いきなり全体を書き換えない。
- 新規ファイル作成より既存パターンの踏襲を優先する。
- 仕様とのズレが出そうな箇所を先に見つける。

## 3. タスクに分解する

### 分解の基準

- BE と FE を分ける。
- 画面、API、DB、テストを混ぜない。
- 1 issue は 1 つの検証責務に絞る。
- 依存があるものは先に前提を固定する。

### 典型分割

- BE: API 契約、集計ロジック、SQL、テスト
- FE: 画面構成、表示列、フィルタ、CSV、状態管理、テスト
- 共通: 命名規則、エラー文言、受入条件

## 4. 実装用プロンプトの作り方

### 基本構成

1. 読む文書を列挙する。
2. 担当範囲を明記する。
3. 変更対象ファイルを明記する。
4. 変更してよい範囲と触らない範囲を明記する。
5. 完了時の報告形式を指定する。

### BE 用テンプレート

```text
あなたは BE 担当のコードエージェントです。

まず以下を読んで、仕様とタスクを理解してください。
- docs/CUSTOMER_AGGREGATION_SPEC.md
- docs/API_SPEC.md
- docs/tasks/open/aggregation/00-OVERVIEW.md
- docs/tasks/open/aggregation/BE.md
- docs/tasks/open/aggregation/ISSUE-001-be-annual-sales-ranking.md
- docs/tasks/open/aggregation/ISSUE-002-be-visit-count-aggregation.md
- docs/tasks/open/aggregation/ISSUE-003-be-last-visit-classification.md
- docs/tasks/open/aggregation/ISSUE-004-be-search-sort-pagination-filters.md
- docs/tasks/open/aggregation/ISSUE-005-be-response-tests.md

担当範囲は backend/ のみです。
編集してよい主な対象:
- backend/internal/handler/aggregation_handler.go
- backend/internal/service/aggregation_service.go
- backend/internal/repository/ltv_repository.go
- その周辺の型、ルーティング、テスト

やること:
- 仕様の 3 つの集計要件を満たす
- 正規項目名を維持する
- 期間優先順位と境界値を仕様どおりにする
- 必要なテストを追加する

制約:
- 他人の変更は戻さない
- 担当範囲外のリファクタはしない
- LTV / LINE / タグ管理の残骸を混ぜない
- 変更ファイルは最小限にする

完了時に必ず出力してほしいもの:
- 変更したファイル一覧
- 実装内容の要約
- 実行したテスト
- 仕様上の未解決事項があれば明記
```

### FE 用テンプレート

```text
あなたは FE 担当のコードエージェントです。

まず以下を読んで、仕様とタスクを理解してください。
- docs/CUSTOMER_AGGREGATION_SPEC.md
- docs/API_SPEC.md
- docs/tasks/open/aggregation/00-OVERVIEW.md
- docs/tasks/open/aggregation/FE.md
- docs/tasks/open/aggregation/ISSUE-006-fe-dashboard-tabs.md
- docs/tasks/open/aggregation/ISSUE-007-fe-list-display-switch.md
- docs/tasks/open/aggregation/ISSUE-008-fe-filters-sort-csv.md
- docs/tasks/open/aggregation/ISSUE-009-fe-empty-error-loading.md

担当範囲は frontend/ のみです。
編集してよい主な対象:
- frontend/src/features/aggregation/*
- frontend/src/app/router.tsx
- frontend/src/config/paths.ts
- その周辺の型、表示、テスト

やること:
- 顧客集計ダッシュボードを実装する
- 3 軸の表示切り替えを行う
- 画面と CSV の列を仕様に合わせる
- フィルタ、ソート、ページング、空表示、エラー表示を整える
- 必要なテストを追加する

制約:
- 他人の変更は戻さない
- 担当範囲外の UI 改修はしない
- 画面名や文言は仕様書に合わせる
- LTV / LINE / タグ管理の残骸を混ぜない
- 変更ファイルは最小限にする

完了時に必ず出力してほしいもの:
- 変更したファイル一覧
- 実装内容の要約
- 実行したテスト
- 仕様上の未解決事項があれば明記
```

## 5. AI エージェントとのやりとり

### 報告を受けたときの確認軸

- 仕様と一致しているか。
- スコープ外の変更を含んでいないか。
- 命名がブレていないか。
- 既存コードの責務境界を壊していないか。
- テストが仕様の肝を押さえているか。

### 返答の基本方針

- まず事実を確認する。
- ズレがあれば、直すべき点を 1 つずつ示す。
- 無関係な要素は切り分ける。
- スコープを拡大しない。

### 短い修正指示の例

```text
その実装は今回の仕様に含まれていません。
削除してください。

次に確認したい点:
- 仕様に書かれている正規項目か
- 受入条件を満たしているか
- 余計な既存機能を巻き込んでいないか
```

## 6. PR 作成

### PR 前に確認すること

- 仕様書と issue の受入条件を満たしている。
- 変更ファイルがスコープ内に収まっている。
- backend / frontend の責務が混ざっていない。
- テスト結果が取れている。
- セルフレビュー済みである。

### PR に含めるもの

- Summary
- Test Plan
- 変更の範囲
- 未解決事項

### PR 本文テンプレート

```markdown
## Summary
- 何を実装したかを 1〜3 行で記載

## Test Plan
- [ ] 実行したテスト
- [ ] 手動確認した画面 / API
- [ ] 未解決事項があれば記載

## Notes
- 仕様上の補足
- レビューで見てほしい点
```

## 7. セルフレビュー

### 必須チェック

- 仕様の 3 つの要件が満たされているか。
- issue の受入条件に漏れがないか。
- 名前が仕様と一致しているか。
- 不要な旧実装が残っていないか。
- テストが境界値を含んでいるか。
- 変更が他機能に波及していないか。

### レビュー観点

- 期間優先順位
- 境界値
- 空データ
- エラー時の挙動
- ページング
- 多拠点分離

## 8. 変更後に残すもの

- 仕様書
- issue
- 実装コード
- テスト
- PR

会話のログではなく、**後から誰が見ても分かる文書**として残す。
