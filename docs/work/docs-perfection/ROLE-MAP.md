# Living Docs role map

対象は既存の `docs/`。本書は正本への参照表であり、製品仕様・規約・実行状態の新しい正本ではない。判定の優先順位は実行可能な検査 → 現在のコード・runtime → 採択ADRと機械化規則 → 現行の参照文書 → コメント・時点メモ。コードとの差分は「観測した挙動」と「採択済み要件」を分け、未承認の要件変更を文書へ押し込まない。

## Constitution

| Fact / 責務 | Canonical path | 読み取り・更新ルール |
|---|---|---|
| エージェント開発規約 | [.claude/CLAUDE.md](../../../.claude/CLAUDE.md) via [AGENTS.md](../../../AGENTS.md) | 規約の詳細を本書へコピーしない。最寄りのCLAUDE.mdを作業種類で選ぶ |
| 何を作るか・作らないか | [product-philosophy.md](../../product-philosophy.md) | 5ステップの全文owner。臨床安全は下行を優先 |
| 臨床安全・製品全体の採択要件 | [spec/specification.md](../../spec/specification.md) §2.1 | 画面実装との差分は仕様判断として扱う。安全要件を実装に合わせて弱めない |
| 個別業務・画面の仕様 | [spec/README.md](../../spec/README.md) → 各業務/画面文書 | 全体仕様へ個別の操作手順を重複させない |
| UIの形状・寸法と製品色 | [DESIGN.md](../../../DESIGN.md) ／ [spec/design-system.md](../../spec/design-system.md) | DESIGNは形状・寸法等、design-systemは製品色。別々のfact owner。準拠結果はStatus |
| API wire contract | [backend/docs/api.yaml](../../../backend/docs/api.yaml) | docsは用途の説明と参照のみ。別OpenAPIを作らない |
| DB実装・package依存の機械的制約 | [001_init.sql](../../../backend/migrations/001_init.sql) ／ [lintscan](../../../backend/internal/lintscan) | migration/検査が一次証拠。文書表を理由にコードやDDLを変更しない |

## Map

| Fact / 責務 | Canonical path | 重複を避ける境界 |
|---|---|---|
| docs全体のカテゴリ入口 | [docs/README.md](../../README.md) | ファイル一覧はカテゴリREADMEがowner |
| 技術説明・設計資料の入口 | [architecture/README.md](../../architecture/README.md) | 本書は一覧を複製しない |
| 現行backend構造の案内 | [architecture/overview.md](../../architecture/overview.md) | 採択理由はADR、機械allowlistはlintscan、GORM型別ownerは次行 |
| 型ごとのwrite-owner参照表 | [model-write-owner-catalog.md](../../architecture/model-write-owner-catalog.md) | ADR-006は所有原則、BE9 mapは履歴。3つを独立した現行owner表にしない |
| cross-domain write契約の検索 | [cross-domain-orchestration-catalog.md](../../architecture/cross-domain-orchestration-catalog.md) | 実装経路の検証は表が指すコードと検査へ。契約を他の索引へ転記しない |
| DB表inventory・RBAC説明 | [erd.md](../../architecture/erd.md) ／ [auth.md](../../architecture/auth.md) | DDLと[permission.go](../../../backend/internal/model/permission.go)が実装の一次証拠。ERD/authは説明のowner |
| 運用手順の入口 | [ops/README.md](../../ops/README.md) → [deploy/README.md](../../ops/deploy/README.md) | 構築/運用はops、納品当日はdelivery。未確認のprovider状態をMapへ混ぜない |
| ローカル受信機配布の前提 | [LAB_DEVICE_AGENT_MACOS.md](../../ops/deploy/LAB_DEVICE_AGENT_MACOS.md) | 実装契約は[LAB_DEVICE_CONNECTIVITY.md](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md)。院内操作の外部手順は同文書の参照先 |
| 納品パッケージの入口 | [delivery/README.md](../../delivery/README.md) | U1–U12はDELIVERY_PACKAGE、U13はOPERATION_MANUALがowner |
| 現場の詳細操作説明 | [OPERATION_MANUAL.md](../../delivery/OPERATION_MANUAL.md) → [frontend manual content](../../../frontend/src/features/manual/content) | deliveryはナビゲーション。参照先の矛盾はRQ-002へ記録し、docs外変更を勝手に許可しない |

## Status

| Fact / 責務 | Canonical path / system | 使い方・証拠境界 |
|---|---|---|
| 製品タスクの状態・担当・Done | [work/README.md](../README.md) が指す Linear | 今回は外部未照会。現在の状態をrepoから推定しない |
| repoに結び付く未完了作業・確認済みUAT製品FAIL | [todo.md](../../../todo.md) ／ [todo.md#product-bugs](../../../todo.md#product-bugs) | work/READMEが認める例外の担当範囲のみ。Linearと競合すればLinearを優先 |
| UAT結果のrepo集計 | [UAT-DOMAIN-STATUS.md](../../ops/testing/UAT-DOMAIN-STATUS.md) | 過去実行snapshot、ソース照合、原証跡の有無を区別。原証跡は非配布で現在未確認のものがある |
| UI準拠の静的対応表と過去runtime結果 | [ui-design-compliance.md](../../spec/ui-design-compliance.md) | runtime未実行ルートをPASSにしない。表の正本性と検証の鮮度は別 |
| docs保守の親状態と子ユニット状態 | [GOAL.yaml](GOAL.yaml) ／ [LEDGER.md](LEDGER.md) | 本依頼で明示された保守専用の例外。製品タスクの状態を複製しない。子の証拠は主coordinatorが直列統合 |
| docsの修復待ち | [REPAIR-QUEUE.md](REPAIR-QUEUE.md) | RQの状態は文書修復だけ。子の完了状態はLEDGERへリンク |
| 意図的に削除した文書のdelete-zone | [work/README.md](../README.md)「削除済み docs」 | 新しい削除候補はRQで根拠・後継・再作成条件を提案してから処理。旧STATUS等を再作成しない |

## History

| Fact / 責務 | Canonical path | 訂正・保存ルール |
|---|---|---|
| 採択された設計判断と理由 | [architecture/adr/README.md](../../architecture/adr/README.md) → 各ADR | 原則不変。明確な事実訂正には日付・何を・なぜ・根拠のcorrection note。新判断は新ADRでsupersedes。理由の静かな美化は禁止 |
| BE9分類と当時の測定 | [be9-2a-boundary-map.md](../../architecture/be9-2a-boundary-map.md) | historical input。現行allowlistはlintscanへリンク。古い日付/SHAだけではobsoleteにしない |
| ARCH-A4当時の着手判定 | [arch-a4-trigger-ledger.md](../../architecture/arch-a4-trigger-ledger.md) | 新しい着手には再測定。過去測定を現在のgo/no-goへ流用しない |
| 採択方針への短い参照 | [work/decisions/README.md](../decisions/README.md) | 詳細の会社側記録は既存CorpVaultポインタ。外部本文/状態は未検証 |
| この保守ユニットの判断・検証履歴 | [PROGRESS.md](PROGRESS.md) | 製品の履歴・runtimeログを再収集しない。日常の変更履歴はGit |

## 1 fact = 1 canonical owner の更新手順

1. 内容の変更は上表のownerで一度だけ行う。索引・別カテゴリ・古いメモはownerへリンクする。
2. docs更新と実装が矛盾したら、検査・コード・採択理由を比較してRQへ根拠を記録する。意図が不明なら子ユニットの該当部分をBLOCKEDにし、判断を発明しない。
3. 削除は後継のowner、削除理由、再作成条件をdelete-zoneへ用意してから行う。履歴への退避を「現行の正本」増設として扱わない。
4. 状態の更新は実行日時と直接のreceiptを伴う。静的ゲートの緑・Issue Closed・code completeだけでUAT/稼働を認定しない。

既存harnessは AGENTS.md → .claude/CLAUDE.md を入口とする。本ユニットではharnessファイルを変更していないため、このrole mapの自動ロードは主張しない。保守セッションは [work/README](../README.md) の入口から本書 → GOAL/LEDGER → 該当カテゴリの順に明示的に読む。
