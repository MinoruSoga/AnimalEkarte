# delivery/ — 納品ドキュメント

> **目的**: クライアント納品に向けた納品物ドキュメントの索引を提供する。
> **読者**: 納品担当・先方管理者・現場スタッフ。
> **タイミング**: 納品準備・本番切替準備・納品後の運用引き継ぎ時。
> **前提**: **Production（本番）は未構築**（STG のみ稼働）。本番構築は #253 / [../ops/infra/production/setup.md](../ops/infra/production/setup.md)。repo 由来 3 領域の同期正本は [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md)（#258）。

開発者向けの技術・運用文書は [../architecture/](../architecture/README.md)・[../spec/](../spec/README.md)・[../ops/](../ops/README.md) を参照。本フォルダは**先方に渡す文書とその作成過程**だけを置く。

## 索引

| ドキュメント | 対象 Issue | 読者 | 内容 |
|:---|:---|:---|:---|
| [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) | #258 | 先方管理者 | システム構成概要・管理者向け初期設定・運用手順（STG SSOT 同期済み。Production 未構築と USER 入力待ちを明示） |
| [GOLIVE_RUNBOOK.md](GOLIVE_RUNBOOK.md) | #257 | 切替実施者 | 本番切替の前提チェック・当日タイムライン・切り戻し基準 |
| [OPERATION_MANUAL.md](OPERATION_MANUAL.md) | #256 | 現場スタッフ | 画面操作への最短ナビゲーション（詳細はシステム内マニュアルが正本） |

## 運用ルール

- 本番切替の技術的な構築手順は [../ops/infra/production/setup.md](../ops/infra/production/setup.md)（開発側文書）が正本。本フォルダの GOLIVE_RUNBOOK は当日のオーケストレーションを担う。
- 現行インフラ構成の正本は [../ops/infra/architecture.md](../ops/infra/architecture.md)。環境 URL・デプロイは [../ops/deploy/README.md](../ops/deploy/README.md)。
- DELIVERY_PACKAGE の管理者設定 path（`/settings/clinic`・`/settings/staff`・`/settings/permission-groups`・`/settings/closing-time` 等）は `frontend/src/config/paths.ts` と画面仕様書に一致させる。
- 契約名義・本番バックアップ実測・障害窓口・LINE/Lステップ秘密・通知先メールは repo 外入力（DELIVERY_PACKAGE の **USER 入力待ち** 表）。秘密値は納品ドキュメントに書かない。
- 納品完了後、時点性の強い文書（GOLIVE_RUNBOOK）は役目を終えたら削除して git 履歴に残す（凍結スナップショットを残さない — PRODUCT_PHILOSOPHY ②）。
