# EMR-518 診療項目パッケージ：契約固定と実装分割

**2026-10-20 15:55 UTC追補**: [採用storage契約・最新実装順](emr-518-storage-contract-20261020.md)が7表の列/制約・権限・lock・移行の現在設計。下記schema/API候補・PR順はPR #755時点の履歴で、具体storageと順序は新契約を優先する。財務unknownは単品で許される予定選択ではnull/provenance/warningを保持し、一律blockingにしない。owner/classification未解決は別のselection blocking。機械API/fixture公開は親の単一owner工程で、本PRではapi.yamlを編集しない。API候補のpackage専用operation lookupは採用案で494共通operations routeへ統一するが、本索引だけで機械契約公開済みとは扱わない。

出典: [EMR-518](https://app.plane.so/baritechllc/browse/EMR-518/)。2026-10-20の本人回答により、治療に限らずプランで追加できる種類を含めること、証明書料金は発行前に手動修正できること、package構成は医院によることを採用。[業務正本4FR/12AC](../../spec/domains/medical-records/treatment-packages.md) → [採用技術方向](../../design/domains/medical-records/treatment-packages.md)を読む。

文書作成基準main: `05da02f30cc5320c929095c6c9d6f96a4ac3f065`。原稿の調査基準 `86c9caa34d8498133eb46d4f019c0b08549db554` から更新し、対象入口を静的再確認した。論理要件・技術方向の採用であり、全DTO/FK・公開API固定完了、実装着手の依存解消、runtime完了の宣言ではない。Planeの更新・実装担当への配布・mergeは親担当。

## 今回の編集境界・検証状態

変更はdocs/spec/domains、関連docs/design、docs/workと既存概念正本の参照整合のみ。API YAML、runtime、DDL、model、生成型は変更しない。既存実装セッション/Fable branchも変更しない。

試験/fixture実行・Docker・隔離環境・DB適用・鍵作成は **NOT_RUN**。リンク・根拠パス・差分の静的確認だけを行う。条件表は後続試験の入力でありPASS証跡ではない。後段検証は本人の指示時にmainローカルで行い、今回PRの条件に隔離環境/Docker/DB試験を追加しない。

## 現状対照：省略しない対象

| 対象 | 基準mainの入口 | 目標への差分 |
|:---|:---|:---|
| 診察 | consultations→検索「診察」 | typed参照を共通予定へ保持 |
| 処置 | procedures→「処置」 | 同上。料金用途を名称推定しない |
| 薬剤 | medicines→「薬剤」 | dose gate・数量/在庫責務維持 |
| 検査 | 当該dialogにexam_types供給なし | D03確定仕様のD07型を初版から含む |
| 予防接種 | vaccines→「予防」 | otherへ落とさずD08型を保持 |
| 定期健診 | checkup_types→現在「検査」表示 | D07と混同せずD09型・子構造を保持 |
| 物販 | merchandise_items masterあり、当該dialog供給なし | D03正本の単価初期反映対象として含む |
| 手入力その他 | handleAddRowのother追加 | 正規手入力の理由/表示を保持 |
| 入院見出し | CATEGORY_ORDERにあるがitem供給なし | 選択可能品目の証拠ではない。専用記録/入院開始を自動生成しない |

これは限定調査であり、全repositoryに汎用セット機能が存在しないという断定ではない。既存健診packageはmaster投入manifest/専用Checkup系であり、廃止済みcheckup-package-import APIを復活させない。

<a id="schema-freeze"></a>
## PR #755時点のschema固定チェック（採用storage契約へ接続、実装未完）

以下は実装者への設計丸投げではなく、共通設計ownerがAPI契約公開前に埋める具体作業。DDL/model/ERDを同時更新する実装PRへ渡す。今回DDL実行・migration適用は行わない。

1. 7master FKの正確表名・id/clinic複合unique存在を照合し、exactly-one CHECKとmanual_otherの全FK NULL例外を確定する。
2. plan/application/record/pet複合整合、RLS、論理削除/RESTRICT、必要索引を固定する。client指定clinic/actorを正本にしない。
3. package component key→application→plan identityのunique、操作binding unique、revision unique、再送/取消履歴を確定する。
4. D08等の単一実績対応、D07パネル/複数結果、D09親子の多重度を個別表で固定する。一律1:1や実績共有の無条件許可を課さない。
5. ExactNumeric/NULL/未指定/0/false、master版のない型のserver fingerprint、単位なしと取得不能の区別を型ごとに固定する。単品で予定入力が許される未確定価格/税/保険はnullで出所/理由を保持し、previewからwarningを表示して各ownerの必要段階へ引き継ぐ。財務unknownを追加時に一律blockingせず、owner/classification未解決や単品と同じ追加時不適合のselection blockingと区別する。この段階別gateは[採用契約](emr-518-storage-contract-20261020.md#financial-unknown)で固定した。
6. 旧行はlegacy markerで現状表示する。根拠ある移行のみのsource identity一意化と、未知を移さない条件を固定する。
7. 全列DDL・FK/index・model・ERDの対応を同じ実装PRで確定する。共通schema編集ownerはEMR-516へ集約する。

<a id="api-freeze"></a>
## 初期候補の履歴と公開済み契約

server prefixは既存 `/api/v1`。下記はPR #755時点の初期候補の履歴であり、実装済みrouteを示さない。現在の機械契約はmain公開済みのPR #761の[API契約・fixture](emr-518-package-api-contract-20261020.md)を参照する。公開範囲は13操作 / 9 paths / 55 fixturesであり、runtime完成・fixture実行済みを意味しない。共通planの一覧・個別GET 2本を含むが、plan write/state-change/owner-linkは含まず、引き続き未固定とする。

| method | prefix以降の旧候補path | 用途 |
|:---|:---|:---|
| GET / POST | `/masters/treatment-packages` | 一覧/作成 |
| GET / PATCH / DELETE | `/masters/treatment-packages/{id}` | 個別読取/更新/論理削除 |
| PATCH | `/masters/treatment-packages/reorder` | 順序 |
| POST | `/medical-records/{record_id}/treatment-package-previews` | read-only preview |
| POST / GET | `/medical-records/{record_id}/treatment-package-applications` | apply/一覧 |
| GET | `/medical-records/{record_id}/treatment-package-applications/{id}` | 保存済み適用結果 |
| GET | `/medical-records/{record_id}/operations/{operation_key}?intent=medical_record.package_apply` | 494共通routeによる元操作結果照会（旧package専用lookup候補は廃止。公開済み契約を参照） |
| 未固定 | `/medical-records/{record_id}/plan-items` と個別adapter | 共通予定の更新/状態変更/履歴/owner-link |

初期候補で固定対象としたrequest/responseは、version必須範囲、canonical ID/decimal、preview依存fingerprint、全行適否（単品と同じ段階別gateのblocking/warning。財務unknownを一律blockingにしない）、apply component→保存plan ID、C06 status/結果参照、CAS/依存変更/失権/結果不明応答、scope切替時の無効化。state-changeはcoordinator専用intentとしgeneric state PATCHにしない。

新grantを追加せず既存master/record/各owner操作へ対応させる。現 `routes_masters.go` の診療masterはResourceMasterMedical、checkup-typesはResourceCheckups、merchandise-itemsはResourceMasterMerchandise。package管理が他種類の候補閲覧・実施権限を迂回しないよう、操作ごとの既存grant対応表を先に固定する。当時未固定だったgrant対応は[採用storage契約§7](emr-518-storage-contract-20261020.md)で既存actionへ固定した。「全員使用可能」と補完しない。

FEは公開済みroute/DTOとfixtureを参照し、独自draft契約を増やさない。実API結合/有効化は必要backend subsetの後とし、共有router/OpenAPI/modelの編集担当と順序は既定のまま維持する。

## 最小PR単位と依存subset

| 単位 | 範囲・開始条件 |
|:---|:---|
| 0 共通設計 | 本文書群、EMR-516共通identity/状態/owner-link/schema/API固定。共通部分を二重作成しない |
| 1 Backend package master | master/items DDL＋model＋CRUD/CAS/監査/typed参照＋api.yaml/fixture |
| 2 共通予定/Backend apply | plan/revision/link必要subset＋application、preview/apply/read/receipt。516と一つの実編集ownerへ集約 |
| 3 Frontend master | PR1契約公開後。固定APIの構成編集/順序/有効性/競合/行別エラー。mock準拠PRで独立可能 |
| 4 Frontend chart | PR2契約公開後。固定preview/apply、個別展開、結果不明照会、record scope、各タブ共通予定projection。master画面と編集pathを分離 |
| 5 狂犬病証明書接続 | 医院別セットの既存会計候補への明示追加、発行前料金編集。486/487の版保護/renderer再利用。一般packageに証明書ルールを埋め込まない |

依存は516共通subset→517接種/500健診、[481会計binding](https://app.plane.so/baritechllc/browse/EMR-481/)、[484実施ready](https://app.plane.so/baritechllc/browse/EMR-484/)、[489 receipt](https://app.plane.so/baritechllc/browse/EMR-489/)、[494編集権](https://app.plane.so/baritechllc/browse/EMR-494/)。全票Done待ちにせず必要subsetを固定する。既存5実装セッション488/485/467/513/514へ本追加設計を無断配布しない。

## 証明書・概念との整合

証明書authorの公開済み[D08 edition設計](../../design/domains/vaccinations/technical-spec.md#certificate-edition-proposal)と[今回の料金projection](../../design/domains/vaccinations/technical-spec.md#certificate-fee-projection-20261020)へ接続する。4欄optional/null、発行前手動編集、確認値のedition snapshot、会計への逆反映/自動同期/合計再計算なし。医院別構成・医院職員の既存会計候補への明示追加を維持し、一律手続完了/済票交付イベントや新rule engineは設けない。

既存[カルテ概念正本](../../spec/ontology/medical-records.md#treatment-package-concepts)にpackage定義/構成品/適用/予定identityの関係を追補した。既存の4状態・結果owner・保存版・会計snapshotの意味は変更しない。関係表の多重度は論理上の構成/適用と、未完のowner別実績対応を区別する。

<a id="static-evidence"></a>
## 静的根拠（基準main固定リンク）

以下は現実装/既存文書の根拠であり、新技術方向が実装済みという意味ではない。

- [D03 3タブ/実施と結果の分離](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/docs/spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-plan-domain-tabs)
- [現治療検索](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx)
- [現category→other変換](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/frontend/src/features/medical-records/lib/treatments-tab-model.ts)
- [createの在庫副作用](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/internal/medicalrecord/treatment_service_tx.go#L76-L81)
- [入院兼用treatment_plans](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/migrations/001_init.sql#L1629-L1651)
- [会計/見積numeric精度](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/migrations/001_init.sql#L9381-L9395)
- [単一接種sourceのunique](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/migrations/001_init.sql#L9491)
- [ExactNumeric](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/internal/model/exact_numeric.go)
- [C06共通操作](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/internal/operation/contract.go)
- [masterルート/権限](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/internal/medicalrecord/routes_masters.go)
- [物販ルート/権限](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/internal/inventory/routes.go)
- [健診正系統](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/docs/architecture/adr/004-checkup-canonical-system.md)
- [健診packageと廃止import](https://github.com/MinoruSoga/AnimalEkarte/blob/05da02f30cc5320c929095c6c9d6f96a4ac3f065/backend/checkup-packages/README.md)
- 共通設計窓口: [EMR-516](https://app.plane.so/baritechllc/browse/EMR-516/) / [EMR-517](https://app.plane.so/baritechllc/browse/EMR-517/) / [EMR-500](https://app.plane.so/baritechllc/browse/EMR-500/)

Plane本文/コメントの最新追跡と状態管理は親担当が行う。本作業は提供された採用原稿を基準に文書化し、Plane全件再取得済みとは記録しない。

## 文書静的確認記録

追加/追補8ファイルを対象に、追加行と新規本文のrepositoryリンク52件（相対パス・固定SHAパス・見出しアンカー）を確認し、欠落0件。追加内容に内部メッセージID・ローカル作業pathがないことを確認した。`git diff --check` は問題なし。

`python3 -B scripts/checks/verify-agent-task.py --base 05da02f30cc5320c929095c6c9d6f96a4ac3f065 --plan` はdocs8件、checks=[]、unmapped=[]、executed_count=0。これは静的な検証計画の確認で、試験実行ではない。試験/fixture・Docker・DB・実鍵作成は引き続きNOT_RUN。
