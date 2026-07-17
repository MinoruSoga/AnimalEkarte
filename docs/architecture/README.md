# architecture/ — 説明系（技術・データ設計）

> **目的**: システムが「どう作られているか」の説明系ドキュメントの索引を提供する。
> **読者**: 全開発者・AI エージェント。
> **タイミング**: 設計調査・アーキテクチャ変更・DB/認可まわりの実装前。

仕様（何をするか）は [../spec/](../spec/README.md)、運用（どう動かすか）は [../ops/](../ops/README.md) を参照。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [overview.md](overview.md) | 軽量レイヤードアーキテクチャ（handler → service → repository）の定義 | 層構造・責務分担の確認時 |
| [erd.md](erd.md) | データベース設計（全テーブル・リレーション。テーブル数の正本） | スキーマ調査・migration 作成前 |
| [auth.md](auth.md) | RBAC 権限モデル・マルチテナント（clinic_id）隔離 | 認可・権限・テナント境界の実装前 |
| [data-flow.md](data-flow.md) | リクエスト追跡（Request ID）と非同期同期の仕組み | ログ調査・非同期処理の実装前 |
| [delete-soft-delete-patterns.md](delete-soft-delete-patterns.md) | Hard Delete / Soft Delete の使い分けと FK 制約 | 削除系機能・FK 設計の実装前 |
| [adr/](adr/README.md) | アーキテクチャ意思決定記録（ADR） | 「なぜこの設計か」の経緯確認時 |

## AI エージェント向け注記

- API contract の正本は本フォルダではなく [`backend/docs/api.yaml`](../../backend/docs/api.yaml)。
- erd.md・auth.md・overview.md が宣言する数値（テーブル数・リソース数等）は CI の docs-symbol-drift ゲート（`scripts/check-docs-symbol-drift.sh`）で実装と機械突合される。数値を変更する場合は実装側の実測値と一致させること。
