# 445/446/466/518 基盤・集約hub静的照合

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-master-foundation-hubs-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既決仕様を継承。新しい業務・技術契約を採択しない |
| 実装状態 | 下記の限定接点の分類。全FR/AC・全source監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告を整形したもので、文書編集時に全sourceを独立再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-additional-scope)を参照。

基準: main 77c28fbfbafa257dcca2421af99ff72da6889717（PR846採用SHA）。2026-10-23 14:35–14:40 UTC読取。製品修正、生成、DB、build/test、Mac操作、Plane投稿、新PRなし。

<a id="master-foundation-hubs-conclusions"></a>
## 結論
①保存基盤・②masterの採用済み部分はあるが、4親票の全AC完了ではない。新規障害を実行観測した報告ではなく、既存残件を現在の統合baselineと具体受領票へ再対応した。実装/fixture不足、契約未選定、依存供給待ち、NOT_RUNは別項目。①②③優先を維持し、518の④runtimeを新規着手しない。

最も重要な更新は、466最新コメント77df5b10の「#839 open」を現状として繰り返さないこと。#839はmerge済み24cd9d9184712f97f9af638726eb4c371fb8336eで、現001に003〜010のDDLが統合済み。旧ファイル不在を実装欠落とせず、009適用前既存行の増分fixtureと統合baseline初期状態/統合前後等価性を分離する。

<a id="master-foundation-hubs-coverage-limits"></a>
## coverageと限界
- Plane445/446/466/518の最新本文と全コメント62/38/31/7件取得。全てnext_page_results=false。追加516/521/522/523/524/535/553/554は最新本文を取得し、子の過去コメントは既存PR846監査を再利用。
- #744/#775/#820/#815/#839/#846のPR metadata/body/diffを取得。全採用PRの全discussionを新たに独立検証した意味ではない。
- 固定SHAの001、manifest、schema_drift、storage_finance、generated models、3 realdb fixture、DBOrTx inventory、MasterPageShell、TreatmentPackageSettings、package apply hookを直接取得。既存audit-next資料は補助索引であり、未読全repository/全consumerを網羅したとしない。
- manifest記載は253表/2807列/769FK。取得001バイトのSHA-256 f8f440437aaf36c49211c68b831a7bd811c41e592cb06ec2c712d8614e2a1ef6はmanifestの唯一のmigration entryと一致した。これはファイル対応の検算だけ。SQL解析、全253表の意味論照合、ERD生成/--check、DB適用は実行していない。
- schema_driftの登録文字列数を物理表数と単純比較して不足判定しない。全model/全DDLの双方向登録実査は未実施。#744当時241登録一致や#821〜824当時253表報告を現版の全AC PASSにしない。

<a id="master-foundation-hubs-adopted-scope"></a>
## 採用成果と完了境界
|成果|採用SHA|保持する範囲|
|---|---|---|
|#744|57f80744ea25650a2deaff50cfd8a696b78c1a5b|schema/model登録・testdb入口/fixture。全consumer完成ではない|
|#775|9a1d50f9dcc7d275456f1214507ef0a55441ac98|521/523 master CRUD/FE、524契約client/mock。522実API有効化なし|
|#820|ff369785ffa186539b2cd5ef776860c7c856f637|532のS06/S07/S08/S10。修正済み4件を再計上しない|
|#815|64c067a151964b59b5f42a9873655e2368d8e09b|仕様矛盾文書整理。runtime/受入ではない|
|#839|24cd9d9184712f97f9af638726eb4c371fb8336e|003〜010の001統合・参照移行。旧ファイル復活なし|
|#846|77c28fbfbafa257dcca2421af99ff72da6889717|限定監査19票＋448 subsetのdocs索引。全票全AC完成ではない|

PR839本文は統合前後schema diff、migrate、限定testdb、ERD等の履歴PASSと、model/sharedkernel等のBLOCKEDを記録する。本監査はその履歴を受領しただけで再実行なし。履歴の「STG未適用」「手動reset」も現環境を調査した結果/実行承認ではない。現DBの状態・旧checksum・データ取扱い・明示許可を確認するまで無条件make migrate/reset案内をしない。

<a id="master-foundation-hubs-handoff"></a>
## MECE残差・受領先
<a id="master-foundation-hubs-h01-integration-ledger"></a>
### H01 — 445 / integration_ledger
①②③と④を分離。PR846は限定20票のdocs監査索引で全件完了ではない。
次: 既存API残差/name必須・route inventory・401 header選定は各ownerと最新成果を突合して集約。親のReviewを全AC完成としない。

<a id="master-foundation-hubs-h02-adopted-storage"></a>
### H02 — 446/466 / adopted_storage
PR839で003〜010を001へ統合済み。独立旧migrationの削除は欠落でない。
次: 001/manifest同一SHAを受領。旧009の増分fixture要求・旧phase1/phase2手順を無条件再利用しない。

<a id="master-foundation-hubs-h03-missing-fixture"></a>
### H03 — 466 / missing_fixture
今回取得したclinic_operating_hours_realdb_test.go内のRLS fixtureはSELECT visibilityのみで、WITH CHECK書込拒否fixtureは当該ファイル内に存在しない。他pathの同等fixtureの有無は未確認。
次: 既存466で同院正例/他院INSERT・UPDATE拒否/zero-writeの既存fixtureを再利用できるか確認し、不足分を補完。取得したline_staff_start_policy_realdb_test.go内にRLS probeはなく、他pathの同等fixtureは未確認。anchor等のfixture整備待ちとH05のRLS契約未選定を別管理する。

<a id="master-foundation-hubs-h04-generation-supply-wait"></a>
### H04 — 446/466 / generation_supply_wait
schema_drift登録にClinicOperatingHours/LineStaffStartPolicyあり、generated/models.tsに2型なし。
次: 本人正規生成の固定入力SHA・11出力・単一受領ownerを記録。527旧採用を自動再開せず生成物手編集なし。

<a id="master-foundation-hubs-h05-unselected-contract"></a>
### H05 — 446/492 / unselected_contract
anchor ENABLE/FORCE、通常writer grant、重複uniqueの文書/DDL差が既存保留。
次: 方針選定を先行。全院FORCE化やgrantを本監査で勝手に実装しない。

<a id="master-foundation-hubs-h06-storage-contract-pending"></a>
### H06 — 446/466/483/489 / storage_contract_pending
ack request_hash必須とprice-only intent欠落の既存A/B受付は現001にも残存。
次: 446が単一再送正本・event intent最小案、466が同版保存統合、483がcaller、489がreceipt。ダミーreason/hash禁止。

<a id="master-foundation-hubs-h07-missing-inventory-fixture"></a>
### H07 — 521/455/554 / missing_inventory_fixture
package15＋clinic catalog2＋vaccine5のinventory/原子性証拠。
次: 既存Ready554を単一受領先。521へ17件、455へ5件。#845最終採用hunk保持は統合条件で全機能Done待ちではない。

<a id="master-foundation-hubs-h08-missing-connection"></a>
### H08 — 523/450/553 / missing_connection
MasterPageShell header戻るが直接navigate、package dirty guardを通らない。
次: 既存Ready553へ。optional onBack・package caller・実header接続fixture。全master変更なし。

<a id="master-foundation-hubs-h09-missing-connection"></a>
### H09 — 528/535 / missing_connection
入院plan体格/料金単位解除UIの既知残件。
次: 既存Ready535へ。#809/#825 API/builderを再実装しない。前回報告の528内追跡は535具体受領先に更新。

<a id="master-foundation-hubs-h10-missing-connection"></a>
### H10 — 518/524 / missing_connection
成功後query invalidationはあるがapplication/common-plan subscriber再表示の既存未接続。
次: 524既存mock ACの残差として保持。522実API未供給だけを理由にmock未完成を隠さない。④新規着手なし。

<a id="master-foundation-hubs-h11-dependency-supply-wait"></a>
### H11 — 518/522/516/494/489 / dependency_supply_wait
preview/applyと共通plan writer・S0/S1/S1a/F0必要subsetの供給/接続はmaster成果と別。
次: 522 application writer、516 coordinator/read/Append、494 edit fence、489 receiptを部分受領。全owner Done待ちにしない。

<a id="master-foundation-hubs-h12-NOT-RUN"></a>
### H12 — 466/all / NOT_RUN
現SHAのfresh/seed/RLS/deferred/rollback/競合/HTTP/UI受入未実施。
次: fixture作成不足と実行待ちを別列管理。PR839履歴の限定PASSを現main全受入へ昇格しない。

<a id="master-foundation-hubs-source-evidence"></a>
## 具体source照合
- 001:8241–8256のbilling_price_acknowledgementsはrequest_hash bytea NOT NULL/32byteとscope uniqueを保持。9456–9461のbilling_change_events intentはfinalize/correct/cancel/collect/refund/reallocate。price-only理由不要eventとのA/B接続は契約受付だけで完成していない。storage_finance modelが存在しても483 writer/receipt接続証拠ではない。
- schema_drift_test.go:306,310にClinicOperatingHours/LineStaffStartPolicy登録あり。generated/models.tsには両型の名称がない。本人生成未受領と、保存model未実装を混同しない。
- clinic_operating_hours_realdb_test.go:156–218はpolicy存在/他院不可視とGRANT SELECTのみ。WITH CHECK書込拒否の具体fixtureは今回取得した当該ファイル内には存在しない。取得したline_staff_start_policy_realdb_test.goはCHECK等を持つが、当該ファイル内にRLS probe fixtureはない。他pathの同等fixtureの有無は未確認であり、repository全体の不存在は主張しない。既存466で再利用可否を確認して不足分を限定補完し、544に黙って追加しない。fixture整備待ちはanchorのRLS方針未選定（H05）とは別分類。
- line_acceptance_version_realdb_test.go:3–8,19–35は統合001入口へ変更済みで初期version1/reviewed NULLのINSERT fixtureがある。旧009ファイルを別途適用する要件は失効。過去既存行移行を証明したとのassertメッセージは実操作と区別し、統合前後等価性は#839履歴へ参照する。
- lintscan/dbortx_inventory_lint_test.go:259はpackage LockAllPackagesForReorderのみ、:489–490はvaccine FindByID/Deleteのみ。554の22既存メソッドの登録/fixture対応が未完成という既知残差と一致する。非原子的製品保存を観測したとはしない。
- MasterPageShell.tsx:63はsettings直接navigate。TreatmentPackageSettings.tsx:235のdirty guardを:555のshellへ渡していない。553は正しい既存受領先。
- use-treatment-package-apply.ts:176–196,234–248は成功IDs保持とquery invalidationを備える。ただしこれだけで読取subscriber/再表示は成立しない。PR846の524残差を保持し、後続実API有効化とは別ACにする。

<a id="master-foundation-hubs-hub-completion"></a>
## hub別の完了条件
445: 領域ごとに「選定→公開→保存実装→生成受領→caller接続→fixture作成→実行」の証拠と責任票を索引化する。全512/非終端229という旧棚卸し件数は現在の技術完了件数でない。API name必須/route inventory/401選定の既知事項は最新owner成果で精査するまで新規欠陥数に数えない。
446: 001/model/ERDの同版差分と483 A/B・anchor RLS方針の技術選定を限定subsetで公開する。旧PR727への追記要求や全schema完成gateを戻さない。GoDoc/NOT VALID/indexなど既存低優先残差は採否記録を維持し、未検証だけで本番障害とはしない。
466: 統合001基準のfixture不足と生成受領を閉じる。model test compile残差は旧報告のまま現在FAILと断言せず、既存修復ownerと現head証拠を受領する。#839の独立migration前提を捨てるが、fresh/seed/RLS/deferred/rollback等の後日受入条件を削らない。
518: 521/523 master部分と522/516 runtime、524 mock部分を別段階で受領。553/554の計画Readyを実装済みとしない。package8種・原子的追加・snapshot不変・無副作用・二重課金防止の親ACは子の採用だけで全PASSにしない。

<a id="master-foundation-hubs-references"></a>
## 参照
- https://github.com/MinoruSoga/AnimalEkarte/pull/839
- https://github.com/MinoruSoga/AnimalEkarte/pull/846
- https://app.plane.so/baritechllc/browse/EMR-445/
- https://app.plane.so/baritechllc/browse/EMR-446/
- https://app.plane.so/baritechllc/browse/EMR-466/
- https://app.plane.so/baritechllc/browse/EMR-518/
- https://app.plane.so/baritechllc/browse/EMR-553/
- https://app.plane.so/baritechllc/browse/EMR-554/
- https://app.plane.so/baritechllc/browse/EMR-535/
固定sourceリンクは https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/ に上記pathを接続する。

主要な固定sourceへの入口：

- [001保存定義](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/migrations/001_init.sql)、[storage finance model](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/storage_finance.go)、[schema/model登録](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/schema_drift_test.go)。
- 取得範囲を限定した[医院時間fixture](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/clinic_operating_hours_realdb_test.go)、[LINE開始policy fixture](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/line_staff_start_policy_realdb_test.go)、[受付版fixture](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/line_acceptance_version_realdb_test.go)。
