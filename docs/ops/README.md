# ops/ — 運用系（デプロイ・CI・テスト・インフラ）

> **目的**: システムを「どう動かすか」の運用系ドキュメントの索引を提供する。
> **読者**: 全開発者・DevOps・AI エージェント。
> **タイミング**: デプロイ・CI 変更・テスト実施・インフラ作業の前。

編集時のルールは [CLAUDE.md](CLAUDE.md) を参照。技術設計は [../architecture/](../architecture/README.md)、仕様は [../spec/](../spec/README.md)。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [deploy/](deploy/README.md) | デプロイハブ（環境 URL・Cloudflare デプロイ・障害時判断） | デプロイ・リリース作業前 |
| [deploy/runbooks/](deploy/runbooks/README.md) | 個別実行手順書（リリース前検証・scheduler運用・資格情報ローテーション・シークレット棚卸し） | 該当オペレーション実施時 |
| [deploy/LAB_DEVICE_CONNECTIVITY.md](deploy/LAB_DEVICE_CONNECTIVITY.md) | 検査機器の有線疎通。現行カルテは Drワン、新カルテは使わない | 院内で機器を新カルテへ取り込むとき |
| [testing/](testing/README.md) | テストアーキテクチャ・受入 scenarios（項目単位）・E2E・環境セットアップ | テスト実施・品質検証時 |
| [ci-policy.md](ci-policy.md) | CI ワークフローの決定事項記録（Actions バージョンピン方針等） | .github/workflows/ 変更前 |
| [coverage-policy.md](coverage-policy.md) | テストカバレッジ ratchet 方式の運用ポリシー | カバレッジゲート調整時 |
| [backlog-spreadsheet.md](backlog-spreadsheet.md) | Q&A バックログスプレッドシートの運用ルール | クライアント Q&A シート操作前 |
| [infra/architecture.md](infra/architecture.md) | インフラ構成図・ネットワーク・セキュリティ設計（`../architecture/overview.md` のレイヤード構造とは別物） | インフラ構成の調査・変更前 |
| ~~infra/_archive/aws-legacy/~~（git 履歴のみ） | 2026-07-20 廃止の AWS 基盤記録。2026-08-20 削除（`git show e0260d32f^:docs/ops/infra/_archive/aws-legacy/` で復元可・**実行禁止**） | 過去の判断・実施証跡を調査する時のみ |

## AI エージェント向け注記

- タスク台帳の実行 SoT は **Linear**（hub BRT-4。2026-08-14 移行）。root [`todo.md`](../../todo.md) · [`todo-po.md`](../../todo-po.md) は入口ポインタ。運用で発見した課題は Linear に起票する。旧 `q&a.html` / `STATUS.md` / `PO-todo.md` は削除済。
- **#261 / BRT-51 承認の所在（代行しない・`q&a.html` は捏造しない）:** 判定の書き先は GitHub [#261](https://github.com/MinoruSoga/AnimalEkarte/issues/261) 本文（A/B/C）。`q&a.html` は repo に無い。B 薬量は [BRT-39](https://linear.app/baritechllc/issue/BRT-39)。C 健診 seed は [BRT-40](https://linear.app/baritechllc/issue/BRT-40)。検査 / #249 はスコープ外。
- migration/seed に触れる作業は `migration-seed-safety` スキル、リリース前チェックは `stg-release-readiness` スキルを先に読むこと。
