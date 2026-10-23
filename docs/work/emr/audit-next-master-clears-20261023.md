# 入院プラン・診断マスタのclear継続静的監査（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-master-clears-20261023 |
| 種別 | 528/529限定静的監査の受領記録 |
| 決定状態 | 既存clear契約・既知後続を継承。UI修正案は提案のみ |
| 実装状態 | 以下の実装済みsubsetを再利用。528の既知未接続UI1件を維持 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

基準mainは `171b56575e7ee028e54358dad22ae28574853125`（受領報告のGitHub branch読取2026-10-23 13:26 UTC）。[索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下のsource・fixture存在は静的確認で、今回の実行結果ではない。

<a id="master-clear-audit-conclusion"></a>
## 結論と状態の確認範囲

新規の実装欠陥は本限定範囲では確認しなかった。528の備考clear・PATCH三値・明示0価格と、529のType/Name両Update備考clearはmain実装済み。旧指摘を再計上しない。

既知・承認済みの未接続後続は528の「体格・料金単位を未設定へ戻すUI」1件。API/builderは表現可能だがside panelに解除操作がない。回答 `6dcdcb7a-2b82-4151-bc6c-c38a5a4eb2da` の延期は取消ではない。529は今回契約に追加製品修正不要。両票のruntime/DB/browser受入はNOT_RUNで、実装済みを実機PASSにしない。

受領監査では両票assignees=[]、state_group=started、過去コメントはReviewを示す。状態・担当は変更しない。UI後続は既存528 / MM-HOSPITAL-OPTIONAL / D枠。明示的な新claimは未確認であり、新票・別PR作成を要求しない。

<a id="master-clear-audit-adoption"></a>
## 証拠と採用履歴

528本文と全11コメント、529本文と全6コメントを読取した報告を受領（両next_page_results=false）。最新コメントは528が2026-10-23 00:34:46 UTC、529が2026-10-22 06:23:30 UTC。古い「実装未開始」より後日のmergeと固定mainを優先する。

| PR | 採用SHA | 限定範囲 |
|---|---|---|
| [806](https://github.com/MinoruSoga/AnimalEkarte/pull/806) | `ec8210c71cb594931eb39478b78460f91913aeaf` | API契約 |
| [809](https://github.com/MinoruSoga/AnimalEkarte/pull/809) | `b104aa62a0aa237b73ae6f70a6c97ec772fd5567` | 入院プラン実装 |
| [811](https://github.com/MinoruSoga/AnimalEkarte/pull/811) | `db07e72d87d9c2bd8f40fb815be9ffb5532e8285` | 448別マスタclear。528専有差分と混同しない |
| [812](https://github.com/MinoruSoga/AnimalEkarte/pull/812) | `745244c719965735149c90265ed56f985befbd25` | 診断Type/Name clear |
| [825](https://github.com/MinoruSoga/AnimalEkarte/pull/825) | `03932bbbdabcbc37306373cad0997638b88c9800` | 533/536 PATCH専用schema。528 null enum追補済み |

#803/#804の採択計画・固定画面仕様と最終回答を基準とする。旧共有schemaのnullable不整合を528の未修正として再計上しない。

<a id="master-clear-audit-hospitalization-evidence"></a>
## 528のAC照合

根拠は固定mainの[FE builder](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/routes/hospitalization-settings-model.ts)、[request](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/hospitalization_plan_request.go)、[service](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/hospitalization_plan_service.go)、[repository](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/hospitalization_plan_repository.go)。表の行番号はこのSHAでの限定接点。fixtureは存在又は期待値で、全実行NOT_RUN。

| AC/選定契約 | 静的判定・具体根拠 | fixture／残る検証 |
|---|---|---|
| 備考 非空→空 | 実装済み。builder:23–36はdescriptionをそのまま送信。request:72–99の*string、service:36–38で空文字をmap保存 | hospitalization-settings-model.test.ts:47、hospitalization_plan_repository_test.go:308 |
| body_size/billing_unit三値 | 実装済み。request:11–41でpresence保持、82–95で伝搬。service:39–50で省略key無し/nullはnil/enum値、113–133/238付近で不正enumを書込前拒否 | request_test.go:142/189、service_test.go:400/434、repository_test.go:279/293/308/325 |
| PATCH空文字はnull代用不可 | 実装済み。[api.yaml:6253–6269](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/docs/api.yaml#L6253-L6269)の専用PATCH enumはnullを含み空文字を含まない。26425でrequestへ結線 | backend/docs/fixtures/emr533/patch-body-size-null.request.json / patch-body-size-empty.request.json。manifest NOT_RUN |
| POST明示0価格 | 実装済み。builder:13のprice:data.price、request:46の*int64→service:196のPrice。nilと区別 | FE model.test.ts:24/37、service_test.go:352、repository_test.go:269 |
| TaxRate/SortOrder型境界 | 実装済み。service:69–76のExactNumeric adapter、186–201でexact税率/sortOrder pointer。旧float64直接代入は解消 | service_test.go:352、repository_test.go:269。全体build成功とはしない |
| read nullable | 実装済み。responseはomitempty維持、api/hospitalization-plans.ts:106–107の??null、lib/hospitalization-side-panel-model.ts:24–25で空フォーム写像 | DB再GET/実機NOT_RUN。read省略をPATCH省略と混同しない |
| POST空/省略と既定 | 既存確定挙動維持。request:50–51 omitempty、service:203–210は空のままCreate、model/hospitalization_plan.go:31 default per_day。body_size既定なし、PATCH nullと別 | POST空/省略→per_day・body_size未設定のHTTP→DB独立fixture実行証拠未確認。sourceによる静的推定でruntime PASSでない |
| 同院/他院・参照中削除拒否 | 実装済み経路維持。repository:45–46/64–72 scoped read/update、75–100 NOT EXISTS/soft-delete | repository_test.go:133/156/390 |
| 過去care/会計不変 | master単表Update維持、repository:71–72。固定画面仕様は退院課金に本master単価を使用しない | care不変fixture:347/371。会計row不変の実DB実証は未確認。405/481を依存gateにしない |
| 値ありUI→未設定 | **既知未接続後続**。HospitalizationSidePanel.tsx:153–181はenum選択のみ、api/hospitalization-plans.ts:73–88に解除値なし。builder:32–33は空フォーム→null対応済みだが画面で空へ戻せない | 操作→null PATCH→再読込のcomponent/integration fixture未準備。次節の限定案へ |

<a id="master-clear-audit-optional-ui-plan"></a>
## 528の限定UI後続案（未実装）

既存528内で体格/料金単位Selectへ明示的「未設定」操作を追加する提案。空文字SelectItemを直接使わずUI専用sentinel→フォーム空文字→既存Update builderのnull送信を維持する。Createでは空フォームを省略へ写像しper_day既定を維持。対象はside panel、必要なら専有model、操作fixtureだけ。共通Select全体や他masterのnullable規律は変えない。

受入期待値はsmall→未設定でbody_size:null、per_night→未設定でbilling_unit:null、保存失敗draft保持、再読込未設定、片方解除で他方保持、新規未選択はキー省略、0価格保持。API三値/POST既定/備考clearを再実装せず、BE/DDL/codegen/backfill/日数課金を追加しない。

既存回答で追跡中。#809はAPI/builder成果のためreopenしない。受領報告のopen PR検索では専用解除UI PRなし。ただし検索語に載らないsession不存在までは断定せず、実装前のclaim確認は必要。

<a id="master-clear-audit-diagnosis-evidence"></a>
## 529のAC照合

根拠は[diagnosis builder](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/routes/diagnosis-settings-model.ts)、[request](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/diagnosis_request.go)、[service](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/diagnosis_service.go)。

| AC | 判定・根拠 | fixture／境界 |
|---|---|---|
| Type/Name備考clear/設定/保持 | 実装済み。Update:35–43/57–65でCreate builder後にdescription:data.descriptionを上書き。Create:25–32/46–54は空→undefined維持 | diagnosis-settings-model.test.ts:28/36/44/52/61/70、NOT_RUN |
| 有効状態・所属の保持 | 実装済み。is_active:data.isActiveとNameのdiagnosis_type_id維持。service:65–100はpointerのあるkeyだけ更新 | Name fixture:61でカテゴリ12/false、Type:36でfalse。カテゴリ変更含む実画面往復未確認 |
| BE空文字保存・省略保持 | 実装済み。request:19–32/53–68、service:73–75/94–96 | 実DB往復NOT_RUN。nullは*string nilで保持、SQL NULL clear契約は追加しない |
| 所属カテゴリ必須・他院/不存在/0拒否 | 既存契約維持。request:37 required、service:280–283 create/311–315 updateで同院typeRepo.FindByID。update省略/null保持 | runtime NOT_RUN、分類nullable再設計は対象外 |
| 子カテゴリ/カルテ参照中削除拒否 | 既存guard維持。service:185–197/331–343、diagnosis_type_repository.go:96以降・diagnosis_name_repository.go:124以降のNOT EXISTS | 静的確認のみ。495複数診断snapshotは別ownerで再保存対象外 |

<a id="master-clear-audit-shared-save"></a>
## shared caller・証拠の限界

HospitalizationSettings.tsx:80、DiagnosisSettings.tsx:110/124から専有Update builderへ接続。[use-master-save.ts:105–145](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/hooks/use-master-save.ts#L105-L145)は対象ID/最新permissionを検査しreqをそのままupdateAsyncへ渡す。空文字/nullをundefinedへ落とさない。144–164は失敗時closeなし/成功後close、194–197のBoolean wrapperを維持。api/hospitalization-plans.ts:139、api/diagnosis.ts:82/112もreqを直接axios.patchへ渡す。shared save変更でclear修正が失われた証拠なし。

builder fixtureと通し試験を区別する。DiagnosisSettings.test.tsxはレイアウト中心でside panelをmockしており、clearの通し試験ではない。#809の当時のbase比較build/限定型検査と、既存型エラーによるmedicalrecord test/vet BLOCKED、FE vitest未実行を区別する。今も同じ失敗と再現してはいない。#812はNOT_RUN、#825の限定main比較結果も全体runtime受入ではない。

価格readの既存null→0表示、任意値一般のnull→0、POST全体null規律、税率全体精度再設計、495/405/481の業務拡張は追加しない。新規不具合数0、既知未接続後続1、実行証拠不足は別分類。本docs作業から製品/試験/DB/Plane変更は行わない。
