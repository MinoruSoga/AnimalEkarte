# delivery/ — 納品ドキュメント

> **目的**: クライアント納品に向けた納品物ドキュメントの索引を提供する。
> **読者**: 納品担当・先方管理者・現場スタッフ。
> **タイミング**: 納品準備・本番切替準備・納品後の運用引き継ぎ時。
> **前提（repo 内最終記録: 2026-08-20）**: **Production（本番）は未構築**。STG の稼働記録があるが、現在の provider 状態は実行時 receipt なしに断定しない。本番構築は #253 / [../ops/infra/production/setup.md](../ops/infra/production/setup.md)。repo 由来 3 領域の同期正本は [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md)（#258）。

開発者向けの技術・運用文書は [../architecture/](../architecture/README.md)・[../spec/](../spec/README.md)・[../ops/](../ops/README.md) を参照。本フォルダは**先方に渡す文書とその作成過程**だけを置く。

## ファイル一覧（inventory）

| ファイル | 対象 Issue | 所有 lane | 読者 | 内容 | repo 由来 | USER 残差 |
|:---|:---|:---|:---|:---|:---|:---|
| [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) | #258 | LANE-5（本 lane） | 先方管理者 | システム構成概要・管理者向け初期設定・運用手順 | 2026-08-31 docs 再監査で deployment 順序・論理テナント分離・権限・時点表示を訂正。U 表は 2026-08-20 に repo 確定 / **未記入** を分離 | **U1–U12 契約記入は未記入**（名義・秘密・本番実測・窓口） |
| [OPERATION_MANUAL.md](OPERATION_MANUAL.md) | #256 | LANE-5（本 lane） | 現場スタッフ | 画面操作への最短ナビゲーション（詳細はシステム内マニュアルが正本） | 2026-08-31 docs 再監査で password・会計確認→確定・trimming・飼主削除を訂正。埋め込みマニュアルは source follow-up (§12) | **U13 = 未完**（日程・receipt・署名は未記入）・#254 FAQ/スクショ |
| [GOLIVE_RUNBOOK.md](GOLIVE_RUNBOOK.md) | #257 | **LANE-2**（通常所有。本 docs 再監査では限定訂正） | 切替実施者 | 本番切替の前提チェック・当日タイムライン・切り戻し基準 | 2026-08-31 docs 再監査で NS 変更禁止・backup gate・実行時 receipt を訂正 | 切替当日の確定待ち多数 |
| [README.md](README.md) | — | LANE-5 | 納品担当 | 本索引 | — | — |

## USER 入力待ち（U*）要約

正本は [DELIVERY_PACKAGE.md の USER 入力待ち表](DELIVERY_PACKAGE.md#user-入力待ち委任外repo-では確定不能)。値・秘密・本番証跡は **発明しない**。

| ID | 要約 | 主ドキュメント |
|:---|:---|:---|
| U1–U4 | 契約名義・プラン・GitHub 運用 | DELIVERY_PACKAGE §1.2 |
| U5–U6 | 本番 LINE / Lステップ秘密（本書に書かない） | DELIVERY_PACKAGE §2 Step 6 |
| U7–U8 | 障害窓口・監視通知メール | DELIVERY_PACKAGE §3.2 / GOLIVE §5 |
| U9–U11 | バックアップ実測・R2 方針・監査保持 | DELIVERY_PACKAGE §3.1 / §3.3 |
| U12 | Production 構築完了証跡 | DELIVERY_PACKAGE §1.3 |
| U13 | 操作説明会。**2026-08-20 棚卸し: 未完**（日程・receipt・署名は未記入） | OPERATION_MANUAL §10 |

## 要追従（本 docs refresh の編集範囲外）

- `docs/ops/infra/staging/runbook.md`: STG の順序表記を workflow と同じ `deploy → migrate → post-migrate /health → optional smoke` に直す。
- `docs/ops/infra/production/runbook.md` §5.1: backup の取得方式・owner・size/checksum・retention を定義し、現行の restore rehearsal と一つの gate にする。定義完了までは GOLIVE を HOLD とする。
- frontend 埋め込みマニュアルと不可逆カルテ操作、backend の未登録 `VoidReopen` 経路、`.claude/CLAUDE.md` 圧縮要約: [OPERATION_MANUAL.md §12](OPERATION_MANUAL.md#12-ソース文書の要追従本-docs-更新の対象外) を参照する。

## 運用ルール

- 本番切替の技術的な構築手順は [../ops/infra/production/setup.md](../ops/infra/production/setup.md)（開発側文書）が正本。本フォルダの GOLIVE_RUNBOOK は当日のオーケストレーションを担う（**通常の編集 ownership は LANE-2**）。
- 現行インフラ構成の正本は [../ops/infra/architecture.md](../ops/infra/architecture.md)。環境 URL・デプロイは [../ops/deploy/README.md](../ops/deploy/README.md)。
- DELIVERY_PACKAGE の管理者設定 path（`/settings/clinic`・`/settings/staff`・`/settings/permission-groups`・`/settings/closing-time` 等）は `frontend/src/config/paths.ts` と画面仕様書に一致させる。
- 契約名義・本番バックアップ実測・障害窓口・LINE/Lステップ秘密・通知先メールは repo 外入力（**USER 入力待ち** 表）。秘密値は納品ドキュメントに書かない。
- ログイン保護の正本は **IP レート制限（5 回 / 1 分）**。「失敗 5 回でアカウントロック」は誤記載（Q&A No.25 / #256）。
- 納品完了後、時点性の強い文書（GOLIVE_RUNBOOK）は役目を終えたら削除して git 履歴に残す（凍結スナップショットを残さない — PRODUCT_PHILOSOPHY ②）。
