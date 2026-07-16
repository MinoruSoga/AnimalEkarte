# delivery/ — 納品ドキュメント

> **目的**: クライアント納品（2026-07-18 Go-live）に向けた納品物ドキュメントの索引を提供する。
> **読者**: 納品担当・先方管理者・現場スタッフ。
> **タイミング**: 納品準備・Go-live 当日・納品後の運用引き継ぎ時。

開発者向けの技術・運用文書は [../architecture/](../architecture/README.md)・[../spec/](../spec/README.md)・[../ops/](../ops/README.md) を参照。本フォルダは**先方に渡す文書とその作成過程**だけを置く。

## 索引

| ドキュメント | 対象 Issue | 読者 | 内容 |
|:---|:---|:---|:---|
| [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) | #258 | 先方管理者 | システム構成・管理者設定・運用手順（納品後に先方で自走するための文書） |
| [GOLIVE_RUNBOOK.md](GOLIVE_RUNBOOK.md) | #257 | 切替実施者 | 本番切替（7/18）の前提チェック・当日タイムライン・切り戻し基準 |
| [OPERATION_MANUAL.md](OPERATION_MANUAL.md) | #256 | 現場スタッフ | 画面操作への最短ナビゲーション（詳細はシステム内マニュアルが正本） |

## 運用ルール

- 本番切替の技術的な構築手順は [../ops/deploy/PRODUCTION_CF_SETUP.md](../ops/deploy/PRODUCTION_CF_SETUP.md)（開発側文書）が正本。本フォルダの GOLIVE_RUNBOOK は当日のオーケストレーションを担う。
- 納品完了後、時点性の強い文書（GOLIVE_RUNBOOK）は役目を終えたら削除して git 履歴に残す（凍結スナップショットを残さない — PRODUCT_PHILOSOPHY ②）。
