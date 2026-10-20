# EMR-518 package機械API・fixtureの契約先行公開

出典: [EMR-518](https://app.plane.so/baritechllc/browse/EMR-518/)。2026-10-20採用設計に基づく公開。[業務4FR/12AC](../../spec/domains/medical-records/treatment-packages.md)、[共通plan設計](../../design/domains/medical-records/treatment-packages.md)、[storage正本](emr-518-storage-contract-20261020.md)、[EMR-494](emr-494-edit-draft-contract-20261020.md)を継承する。4FR/12ACの追加・医院別税/保険値の決定は行わない。

基準mainは`6cb8184caacffeb63874a64fb93d47c15628136d`（PR760証明書契約を通常fast-forwardで取込済み）。証明書9操作/schema/drift entryを保持する。api.yamlと限定drift mapの編集ownerはこの公開作業へ引渡し済み。runtime/DDL/model/DB/FE実装・実鍵・試験/Docker・CI無効化・mergeは対象外。

## 公開物と状態

[OpenAPI](../../../backend/docs/api.yaml)に **新規13操作・9path** をselectedとして追加する。基準mainにはmedical-record共通operations routeもその機械schemaも存在せず、今回は新規13であり12＋拡張1とは数えない。既存CertificateResultResponseはcertificate consumerの固有mappingを持つため変更せず、package consumerはGoの既存operation.ResultResponse wireを同形で記述する。新しいreceipt方式やwrapperは作らない。

[fixture README](../../../backend/docs/fixtures/emr518/README.md) / [manifest](../../../backend/docs/fixtures/emr518/manifest.json)に55件のダミーbody/投影/境界例を配置する。manifestの期待valid/invalidは実行結果ではない。全fixture/schema instance検証・runtime・試験・hash検算は **NOT_RUN**。この公開版をFE契約mockの参照にできる。SQL/runtime完成をmock開始条件にせず、実API結合/有効化は必要backend subsetの後とする。

## 操作一覧（base /api/v1）

| method | path | operationId | 成功 |
|:---|:---|:---|:---|
| GET | /masters/treatment-packages | listTreatmentPackages | 200 TreatmentPackageList |
| POST | /masters/treatment-packages | createTreatmentPackage | 201 TreatmentPackage、Location |
| GET | /masters/treatment-packages/{id} | getTreatmentPackage | 200 TreatmentPackage |
| PATCH | /masters/treatment-packages/{id} | updateTreatmentPackage | 200 TreatmentPackage |
| DELETE | /masters/treatment-packages/{id} | deleteTreatmentPackage | 204、version query必須 |
| PATCH | /masters/treatment-packages/reorder | reorderTreatmentPackages | 204 |
| POST | /medical-records/{record_id}/treatment-package-previews | previewTreatmentPackage | 200 TreatmentPackagePreview |
| POST | /medical-records/{record_id}/treatment-package-applications | applyTreatmentPackage | 既存C06形 |
| GET | /medical-records/{record_id}/treatment-package-applications | listTreatmentPackageApplications | 200 TreatmentPackageApplicationList |
| GET | /medical-records/{record_id}/treatment-package-applications/{application_id} | getTreatmentPackageApplication | 200 TreatmentPackageApplication |
| GET | /medical-records/{record_id}/operations/{operation_key} | getMedicalRecordOperationResult | 既存C06形、intent=medical_record.package_apply |
| GET | /medical-records/{record_id}/plan-items | listMedicalRecordPlanItems | 200 MedicalRecordPlanList |
| GET | /medical-records/{record_id}/plan-items/{plan_item_id} | getMedicalRecordPlanItem | 200 MedicalRecordPlanItem |

全操作にx-contract-status:selectedとEMR518/516/494責務・runtime未実装/FE mock先行を記載。package専用operation lookup、共通plan POST/PATCH/state-change/owner-linkは追加しない。既存7単品master route/DTOは変更しない。

## 論理名と公開schema対応

| 採用原稿の論理名 | OpenAPI component |
|:---|:---|
| Id / Version / Decimal / Yen / operation_key | TreatmentPackageId / TreatmentPackageVersion / TreatmentPackageDecimal / TreatmentPackageYen / TreatmentPackageCanonicalUuid |
| ItemRef / manual | TreatmentPackageItemRef（discriminator付き8分岐）/ TreatmentPackageManualDefinition |
| ComponentInput / PackageCreate / PackagePatch | TreatmentPackageComponentInput / TreatmentPackageCreate / TreatmentPackagePatch |
| Package / PackageList / PackageReorder | TreatmentPackage / TreatmentPackageList / TreatmentPackageReorder |
| PreviewRequest / SourceStamp / PriceSnapshot | TreatmentPackagePreviewRequest / TreatmentPackageSourceStamp / TreatmentPackagePriceSnapshot |
| ComponentIssue / PreviewItem / PackagePreview | TreatmentPackageComponentIssue / TreatmentPackagePreviewItem / TreatmentPackagePreview |
| PackageApply | TreatmentPackageApply |
| 既存C06 ResultResponse | MedicalRecordPlanOperationResultResponse（成功/非成功の同じwire） |
| ApplicationSummary / Application / AppliedComponent | TreatmentPackageApplicationSummary / TreatmentPackageApplication / TreatmentPackageAppliedComponent |
| PreviewItemSnapshot | TreatmentPackagePreviewItemSnapshot（issuesなし、price.uncertainty保持） |
| PlanItem / OwnerLink / collection read | MedicalRecordPlanItem / MedicalRecordPlanOwnerLink / MedicalRecordPlanList |
| SourceProjection / ComponentProjection / PreviewProjection | TreatmentPackageSourceProjection / TreatmentPackageComponentProjection / TreatmentPackagePreviewProjection |

ID/versionはcanonical string＋server signed64 range検査、JPYは非負整数string。regex/長さだけで上限検査完了とはしない。quantityはExactNumericの正数string、共通2桁制約なし。PATCH省略=維持、nullableでないfieldの明示null拒否と財務snapshotの許容nullを分ける。oneOfの他型FK混入・unknown write fieldを拒否する。

## 認可・CAS・副作用

package CRUDは既存ResourceMasterMedical各action＋各型master view。混在した健診/物販をmedical grantだけで読める迂回にせず、権限不足の構成を含むreadは403で全体非開示。chart read/preview/application/lookupは現在ResourceMedicalRecords.view、applyは既存単品POST同じcreate＋現在編集session/fence。edit grant/master管理createは追加要求しない。実施/結果操作は各owner権限。

trusted selected clinic/actorはserver認証contextのみ。record IDで他clinicへ切替せず404、body/query clinic/actor/petを受けない。source record pet NULLはpreview段階のcaptured identity不成立として422、apply新attemptは既定record lock後に非NULL/整合を再検査する。既成功replayは元target現在認可＋元receiptを先に返す。

apply bodyはoperation_key/edit_session_ref/edit_epoch/owner_expected_version/package_id/package_version/preview_tokenの平坦形。clinic/price/items overrideなし。Contentと一時Preconditionsを分け、同内容CAS-only retryは同key、内容変化は旧操作確定後に新明示操作。unknown/timeout/JSON parse失敗は元key照会。Gate→binding NOWAIT→receipt/unknown→savepoint→Apply内lifecycle→edit-right→record→plan/CAS→package/master順はstorage正本を維持する。

plan_items namespaceを既存dependency_versionsで共有し、read/previewの欠落はnull/uninitializedでINSERTしない。494 acquire/takeover成功TXだけがbaselineを準備し、apply遅延bootstrapなし。N構成でもaggregate版増分は1回。PlanExecutionCoordinator以外から状態writeせず、DB pending→wire not_performed、他3状態同値。commercialはselectedで会計fulfillmentと別。新planにunresolvedはなく、旧unknown legacyは別projection。実施/結果/出庫/請求/証明書は選択だけで生成しない。

## 財務・digestと医院別料金

単品で予定入力可能な価格/税/保険unknownはnull/provenance/uncertainty＋previewからwarningで保存。owner分類未解決のblocking/selectionと分け、全型adapterの不足を財務unknownの拒否理由にしない。税込110の保存値、0とnull、明示保険とlegacy_unset/物販not_configuredを区別する。合計は既存owner算出可能時だけで、ダミー例はunresolved/nullとする。親追加請求・固定総額/自動値引き/新税配賦なし。

digest投影の完全keyは[storage §3.4](emr-518-storage-contract-20261020.md#preview-digest-projection)を優先し、api.yamlの3Projection schemaへ対応。JCS UTF16 key順、array順、Unicode非正規化、duplicate key/lone surrogate/invalid encoding拒否、quote/control、null/false/"0"区別をfixture README/vectorsに記載した。application snapshot全体のhashではなく、生成plan IDs等を除く。prefixはtransportだけ、保存bytea32/source_digestのraw hexには付けない。一般tokenは未検算dummy、期待SHAは捏造しない。

医院別package構成を保ち、料金名称でclinical/commercialや一律発生イベントを推定しない。証明書4料金欄は別契約の発行前手動確認額であり、会計逆反映/同期/合計再計算なし。本package APIはcertificate fee role/金額override/発行routeを追加せず、PR760契約も変更しない。

## HTTP境界

pre-operation/master/previewはTreatmentPackageError（error/code＋任意safe issues）。400 invalid-input、401 unauthorized、403 not-authorized、404 scope-not-found、409 version-conflict/dependency-changed/edit-right-lost、422 validation、500 internal-error、503 temporarily-unavailable。SQL/他医院名称/現在価格snapshotを露出しない。

C06到達後はResultResponse形：succeeded/replayed200（元refs/versions/time）、not-succeeded200（成功ではない）、operation-in-progress202、result-unavailable503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost409、validation400、scope-not-found404、not-authorized403。非成功にrefs/versions/time/issuesを出さない。前段error envelopeとはoneOfで区別する。row説明は安全なread-only previewから取得する。共有ライブラリ自体がHTTPを決めるとせず、494 consumer mappingを適用する。

## 7型resolver引渡し

| 型 | 実在reader / model field | 採用adapter境界 |
|:---|:---|:---|
| consultation | ConsultationRepository.FindByID / Price *int64、TaxType/TaxRate ExactNumeric、IsNonInsurance *bool、Version | 保存値とnullable保険、native版を保持 |
| procedure | ProcedureRepository.FindByID / Price *int64、TaxType/TaxRate ExactNumeric、IsNonInsurance *bool、Version | 診察同様。IsSurgeryで税/保険を補完しない |
| medicine | MedicineRepository.FindByID / Price *int64、TaxType/TaxRate ExactNumeric、IsNonInsurance bool、DefaultQuantity *ExactNumeric | modelの正確値をstringへ。response float経由なし。dose/inventoryは別 |
| examination_type | ExamTypeRepository.FindByID（examTypeRepository実装）/ Price *int64、IsNonInsurance bool、Items | 税保存列なし。Itemsはowner definition、税接続前は許容段階でnull。native版を捏造しない |
| vaccine | VaccineRepository.FindByID / Price *int64、IsNonInsurance *bool、Version | 税保存列なし。species/interval/inventoryで税/保険を推定しない |
| checkup_type | CheckupTypeRepository.FindByID / Price *int64、IsNonInsurance *bool、Version、parent/fields | 税保存列なし。健診親子definitionを保持、会計自動生成なし |
| merchandise_item | MerchandiseItemRepository.FindByID / UnitPrice int64、TaxType/TaxRate ExactNumeric | 保険master列なし。カテゴリから税/保険を推定せず、会計候補初期値と区別 |

trusted clinic＋型IDでmodule readerを呼び、内部HTTP呼出をしない。既存billing候補作成をapplyから呼ばない。insuranceApplicableFromMasterSettingはnil保持/明示値反転という既存意味を維持する。旧billingのDefaultTaxRate/未リンクfallbackを新typed packageへ輸入しない。3master nullable税writer/read/model/API subsetはstorage後続実装で値backfillなし、一般予定入力の一律先行gateにはしない。

## 限定drift保守と後続owner

[knownPhantomInSpec](../../../backend/internal/apicontract/openapi_route_drift_test.go)に実runtime未登録のexact METHOD＋full /api/v1 path13件だけを追加。queryはidentityに含めず、既存PR760の9entryと従来entryを保持する。parser/registry/gate/stale検知・テスト条件は変更しない。

package10操作はEMR518、共通plan read2件はEMR516、共通lookup1件はEMR494がruntime登録と同じPRで各entryを削除する。lookup登録後のintent追加だけならphantomを新設しない。route検査試験の実行はNOT_RUNで、この例外をruntime実装免除にしない。

後続実装は494 anchor/fence/CAS、516共通plan store/coordinator閉包、package master/catalog、7型resolver/preview/apply、各owner接続。全票Done待ちでなく必要subsetを固定し、FEは公開schema/fixtureへ接続する。Plane/mergeは親担当。

## 静的確認記録

55 fixture/projection/vector JSONの構文、manifest operationId/component参照、OpenAPI YAML構文と13操作/9path/58新schemaの所在を静的に照合した。instanceのschema適合性検証はNOT_RUN。基準mainの全既存path/componentが同じ内容であり、PR760証明書契約を保持したことを静的比較で確認した。

knownPhantomInSpecは基準17entry（証明書9entryを含む）をすべて保持してexact13entryだけ追加。map外のparser/registry/gate/stale検知コードは基準と同一。runtime登録側のpackage/plan/common operationsに対応する新登録がないことを既存route文字列の限定読取で照合し、runtime不存在の全域証明やroute gate試験PASSとは扱わない。`git diff --check`は問題なし。Go/FE/DB/route drift試験・OpenAPI instance検証・hash検算・DockerはNOT_RUN。
