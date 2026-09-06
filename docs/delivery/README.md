# delivery/ — 納品ドキュメント

> **目的**: クライアント納品に向けた納品物ドキュメントの索引を提供する。
> **読者**: 納品担当・先方管理者・現場スタッフ。
> **タイミング**: 納品準備・本番切替準備・納品後の運用引き継ぎ時。
> **前提（repo 内最終記録: 2026-08-20）**: **Production（本番）は未構築**。STG の稼働記録があるが、現在の provider 状態は実行時 receipt なしに断定しない。本番構築は #253 / [../ops/infra/production/setup.md](../ops/infra/production/setup.md)。repo 由来 3 領域の同期正本は [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md)（#258）。

開発者向けの技術・運用文書は [../architecture/](../architecture/README.md)・[../spec/](../spec/README.md)・[../ops/](../ops/README.md) を参照。本フォルダは**先方に渡す文書とその作成過程**だけを置く。

## ファイル一覧（inventory）

| ファイル | 対象 Issue | 読者 | 内容 | 必要な入力・受入 |
|:---|:---|:---|:---|:---|
| [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) | #258 | 先方管理者 | システム構成概要・管理者向け初期設定・運用手順 | U1–U12（契約・本番実測・窓口等） |
| [OPERATION_MANUAL.md](OPERATION_MANUAL.md) | #256 | 現場スタッフ | 画面操作へのナビゲーション（詳細はシステム内マニュアルが正本） | U13 操作説明会・#254 FAQ/スクショ |
| [GOLIVE_RUNBOOK.md](GOLIVE_RUNBOOK.md) | #257 | 切替実施者 | 本番切替の前提チェック・当日タイムライン・切り戻し基準 | 切替前提・当日の承認と証跡 |

## 関連する受入条件

#252 の全院締め設定は [GOLIVE_RUNBOOK.md](GOLIVE_RUNBOOK.md) の切替前提。#254 は納品前の開発側デモ確認と納品後の現場 UAT を区別し、実 LINE・監査・別 sign-off を含む [close checklist](../ops/testing/scenarios/UAT-254-CLOSE-CHECKLIST.md) で確認する。#259 の Lステップ write 再開は納品後対応であり、設定完了だけでは再開しない。

Issue の OPEN/CLOSED は受入・本番稼働の証明ではない。未記入の契約・承認欄は各文書で管理し、実行状態は Linear と実行時の証跡で確認する。

## USER 入力待ち（U*）要約

正本は [DELIVERY_PACKAGE.md の USER 入力待ち表](DELIVERY_PACKAGE.md#user-input-waiting)。値・秘密・本番証跡は **発明しない**。

| ID | 要約 | 主ドキュメント |
|:---|:---|:---|
| U1–U4 | 契約名義・プラン・GitHub 運用 | DELIVERY_PACKAGE §1.2 |
| U5–U6 | 本番 LINE / Lステップ秘密（本書に書かない） | DELIVERY_PACKAGE §2 Step 6 |
| U7–U8 | 障害窓口・監視通知メール | DELIVERY_PACKAGE §3.2 / GOLIVE §5 |
| U9–U11 | バックアップ実測・R2 方針・監査保持 | DELIVERY_PACKAGE §3.1 / §3.3 |
| U12 | Production 構築完了証跡 | DELIVERY_PACKAGE §1.3 |
| U13 | 操作説明会。**2026-08-20 棚卸し: 未完**（日程・receipt・署名は未記入） | OPERATION_MANUAL §10 |

## 外部証跡が必要な残件

- `docs/ops/infra/production/runbook.md` §4: backup acquisition contract の必須欄を埋め、approved method と receipt を外部証跡で確定する。未確定の間は GOLIVE を HOLD とする。

## 運用ルール

- 本番切替の技術的な構築手順は [../ops/infra/production/setup.md](../ops/infra/production/setup.md)（開発側文書）が正本。本フォルダの GOLIVE_RUNBOOK は当日の実施手順を担う。
- 現行インフラ構成の正本は [../ops/infra/architecture.md](../ops/infra/architecture.md)。環境 URL・デプロイは [../ops/deploy/README.md](../ops/deploy/README.md)。
- DELIVERY_PACKAGE の管理者設定 path（`/settings/clinic`・`/settings/staff`・`/settings/permission-groups`・`/settings/closing-time` 等）は `frontend/src/config/paths.ts` と画面仕様書に一致させる。
- 契約名義・本番バックアップ実測・障害窓口・LINE/Lステップ秘密・通知先メールは repo 外入力（**USER 入力待ち** 表）。秘密値は納品ドキュメントに書かない。
- ログイン保護の正本は **IP レート制限（5 回 / 1 分）**。「失敗 5 回でアカウントロック」は誤記載（Q&A No.25 / #256）。
- 納品完了後、時点性の強い文書（GOLIVE_RUNBOOK）は役目を終えたら削除して git 履歴に残す（凍結スナップショットを残さない — PRODUCT_PHILOSOPHY ②）。
