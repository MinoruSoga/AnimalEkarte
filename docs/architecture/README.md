# architecture/ — 説明系（技術・データ設計）

> **目的**: システムが「どう作られているか」の説明系ドキュメントの索引を提供する。
> **読者**: 全開発者・AI エージェント。
> **タイミング**: 設計調査・アーキテクチャ変更・DB/認可まわりの実装前。

仕様（何をするか）は [../spec/](../spec/README.md)、運用（どう動かすか）は [../ops/](../ops/README.md) を参照。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [overview.md](overview.md) | Go/Gin公式ベースラインに基づく backend 設計・request lifecycle | backend の package境界・責務・運用設計の確認時 |
| [cross-domain-orchestration-catalog.md](cross-domain-orchestration-catalog.md) | cross-domain write の same-tx / best-effort 契約表と new-path ルール（ARCH-A3） | 他 domain を跨ぐ write・intent・tx 境界の設計・レビュー時 |
| [model-write-owner-catalog.md](model-write-owner-catalog.md) | `internal/model` の主要 GORM 型 → write owner package 対応（ARCH-A2） | model 追加・所有曖昧な fact の PR / レビュー時 |
| [composition-root-conventions.md](composition-root-conventions.md) | `cmd/api` composition 規律・Application 横展開評価・route smoke（ARCH-A5） | composition / DI 配線・route 追加の設計・レビュー時 |
| [exception-package-discipline.md](exception-package-discipline.md) | 例外 package 規律（csvimport cmd 限定・identitylink 隔離・bucket 禁止）（ARCH-A8） | 横断 package 追加・cutover ツール・identity 境界のレビュー時 |
| [fe-feature-be-domain-map.md](fe-feature-be-domain-map.md) | FE `features/*` ↔ BE domain / RBAC resource 対応と共有昇格ルール（ARCH-A7） | 新規画面配置・FE 共有化・LIFF 境界の設計・レビュー時 |
| [be9-2a-boundary-map.md](be9-2a-boundary-map.md) | BE9-2A historical domain boundary map・分類 inventory（ADR-006 の入力。現行 allowlist は `internal/lintscan` が正本） | domain 分類・package 境界の根拠確認時 |
| [arch-a4-trigger-ledger.md](arch-a4-trigger-ledger.md) | ARCH-A4 着手トリガー実測と landed スライス記録 | domain 内 file 分割を始める前の go/no-go |
| [erd.md](erd.md) | データベース設計（全テーブル inventory・主要リレーション。テーブル数の正本） | スキーマ調査・migration 作成前 |
| [auth.md](auth.md) | RBAC 権限モデル・マルチテナント（clinic_id）隔離 | 認可・権限・テナント境界の実装前 |
| [data-flow.md](data-flow.md) | リクエスト追跡（Request ID）と非同期同期の仕組み | ログ調査・非同期処理の実装前 |
| [delete-soft-delete-patterns.md](delete-soft-delete-patterns.md) | Hard Delete / Soft Delete の使い分けと FK 制約 | 削除系機能・FK 設計の実装前 |
| [adr/](adr/README.md) | アーキテクチャ意思決定記録（ADR）。ADR-008 は ADR-007 のローカル受信 agent 判断を部分的に supersede | 「なぜこの設計か」の経緯確認時 |

## AI エージェント向け注記

- API contract の正本は本フォルダではなく [`backend/docs/api.yaml`](../../backend/docs/api.yaml)。
- `scripts/check-docs-symbol-drift.sh` の gate matrix は、`erd.md` のテーブル総数・テーブル inventory、`auth.md` の resource 数、screen/design-system docs の限定された CamelCase/file symbol である。`auth.md` / `overview.md` の本文や全固定 symbol を機械検証するものではない。数値を変更する場合は実装側の実測値と一致させること。
- docs-symbol-drift は ADR 本文の narrative 整合、status 文言、歴史的 snapshot の全面 alignment を検証しない。narrative の正本は各 ADR 本体と boundary map を読むこと。
