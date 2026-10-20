# EMR-518 採用storage契約と財務unknownの段階別gate

2026-10-20 15:55 UTC採用契約。要件責任者MinoruSogaのpackage対象拡張とGitHub PR作成承認に基づく。出典: [EMR-518](https://app.plane.so/baritechllc/browse/EMR-518/)。[業務正本4FR/12AC](../../spec/domains/medical-records/treatment-packages.md) → [共通plan設計](../../design/domains/medical-records/treatment-packages.md#storage-contract-selected) → 本書を読む。[前回索引](emr-518-treatment-packages-20261020.md)の未固定記載・旧PR順は本書の採用契約へ接続する。

本書は具体storageの設計固定であり、SQL/RLS・runtime/model/ERDの実装、API/fixture公開、試験PASSではない。今回backend/docs/api.yamlは証明書taskが単独ownerのため編集しない。package機械APIは別途親の単一ownerが統合する。内部の検討原稿を公開正本として参照せず、以下のrepo根拠と本書を引継ぎ先にする。

## 1. 既存物理構造の根拠と前提

文書作成基準mainは `a28f81bfe99010c440ff42a51b8f31811ae1904f`。調査元の[001_init.sql](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/migrations/001_init.sql)を静的に照合した。

| 既存物 | 採用する事実・追加が必要なもの |
|:---|:---|
| medical_record_dependency_versions | PK `(clinic_id,medical_record_id,owner_kind)`、正のbigint revision、record/clinic複合FK。`plan_items` namespaceを再利用し別aggregate版表を作らない |
| medical_records / treatments | 前者は `(id,clinic_id,pet_id)` unique、後者は `(id,clinic_id,medical_record_id)` unique。treatmentsにはpet_id列がない |
| 7master | consultations/procedures/medicines/exam_types/checkup_typesには `(id,clinic_id)` uniqueがある。vaccines/merchandise_itemsには不足し、2targetを追加する実装作業が必要 |
| 実績・結果 | 物理表はexams/vaccinations/checkups。examinationsという物理表はない。billing_itemsにはclinic_idと後段追加の `(id,clinic_id)` / `(id,clinic_id,billing_id)` uniqueがある |
| RLS | app_private.apply_rls_policyはENABLE＋FOR ALL USING/WITH CHECK、既存clinic表はhas_clinic_access(clinic_id)を使用。FORCEではない |
| 494 anchor | medical_record_operation_targetsは[494採用設計](emr-494-edit-draft-contract-20261020.md)の新設前提subsetで、読取DDLに既存とは扱わない。approved subsetを再利用し重複anchorを作らない |
| catalog版 | 既存のpackage catalog版表はない。本書で6domain表に加え小さな補助表を明示新設する |

既存treatment_plans/treatments/clinical_plans/入院care_plan_itemsは流用しない。旧pending/selected booleanから臨床実施を推定しない。

<a id="financial-unknown"></a>
## 2. 共通列・snapshotと財務unknown

新entity IDはbigserial PK、wireは正のsigned-64十進文字列。各domain行のclinic_idはbigint NOT NULL、clinics FKはRESTRICT。複合参照用にUNIQUE(id,clinic_id)を持つ。staff帰属はbigint→staffs RESTRICTで認証済み記録者を表し、医師を推定しない。作成時刻はtimestamptz NOT NULL DEFAULT clock_timestamp()、更新時刻もtimestamptz NOT NULL。clientが帰属・記録時刻を指定しない。全FKのdelete/updateはRESTRICT/NO ACTIONとし、証跡をSET NULL/CASCADEしない。

version/revisionはbigint NOT NULL、正数CHECK。真に新規のentityのみ初期default 1、checked CASで増分しoverflowは拒否する。既存master-native版はnullable snapshotとしversion1を捏造しない。

quantityはtypmodなしnumeric NOT NULL＋ExactNumeric、有限・正数CHECK。共通scaleや臨床用量上限を追加せず、単品と同じ段階でownerの既存精度/単位制約を検査する。owner保存に収める丸めはしない。

| 財務列 | 採用型・条件 |
|:---|:---|
| unit_price | bigint NULL、NULL又は0以上。既存の整数円を維持 |
| tax_type | 既存tax_type enum NULL（included/excluded/exempt） |
| tax_rate | numeric NULL、値がある時は有限。新しい税率/区分対応を作らない |
| is_non_insurance | boolean NULL。falseへの黙示反転/補完なし |

財務列に0/false/10%/excludedのdefaultを置かない。nullは未解決又は適用外で、明示provenanceを伴う。単品で予定入力が許される財務unknownはnullable＋provenance/uncertainty＋warningで保存し、請求ownerの必要段階で確認する。取得失敗を保存済み未設定へ置換しない。型別resolver全7種を完成させ、owner/classificationそのものの不明だけを別のselection blockingにする。保険なし物販はnot_configured_on_master、既存nullable保険はlegacy_unsetとして保持する。

JSONBはserver typed struct生成でcaller任意blobを受けない。object（指定箇所のみarray）、schema_version='1'、正確schemaをwrite前に検査する。snapshot内のID/version/decimal/moneyは文字列、nullableは明示JSON null。unknownはfield/reason/required_atとprovenanceを保持する。digestはschema固有RFC8785 JCS projectionのSHA256をbytea length32で保存し、順序を保持しないjsonb::textのhashを使わない。具体wire/fixtureは別ownerの機械契約公開に合わせる。

<a id="table-contracts"></a>
## 3. Package domain 3表

### 3.1 treatment_packages

共通id/clinicに加え、以下を採用する。

| 列 | 型・制約 |
|:---|:---|
| name / description | text NOT NULL、nameはbtrim非空。description DEFAULT '' |
| is_active | boolean NOT NULL。createで明示しDBが医院方針を推定しない |
| sort_order | integer NOT NULL、0以上。catalog lock下で生成しcaller任意順にしない |
| version | bigint NOT NULL DEFAULT 1、正数 |
| created_by / updated_by | bigint NOT NULL、staff FK |
| created_at / updated_at | 共通の時刻規約 |
| deleted_at / deleted_by | timestamptz NULL / bigint NULL、両方NULL又は両方非NULL |

名前uniqueは付けない。同名はidentity/再送一致を証明しない。list索引は `(clinic_id,deleted_at,is_active,sort_order,id)`。削除は論理のみでversion CAS増分、子/適用は保持する。無効/削除後も現在scope/grant下で過去applicationを読める。

### 3.2 treatment_package_items

| 列 | 型・制約 |
|:---|:---|
| id / clinic_id / package_id | 共通id/clinic、package_id bigint NOT NULL。FK `(package_id,clinic_id)`→treatment_packages |
| component_key | uuid NOT NULL、nil UUID拒否。同じ構成品の編集/並替では安定し、削除後別意味へ再利用しない |
| position / removed_at | integer NULL（NULL又は0以上）/ timestamptz NULL。`(removed_at IS NULL)=(position IS NOT NULL)` CHECK |
| item_kind | text NOT NULL、下記8選択肢 |
| 7typed FK | bigint NULL、下記対応の各FKはclinic_idを含む |
| quantity | 共通の有限・正数numeric NOT NULL |
| manual_definition | jsonb NULL、object。manual_otherだけ必須、それ以外は禁止 |
| created_at / updated_at | 共通時刻規約。独立item版は置かず親versionをcollection CASとする |

| item_kind | 対応FK列→物理master表 |
|:---|:---|
| consultation | consultation_id→consultations |
| procedure | procedure_id→procedures |
| medicine | medicine_id→medicines |
| examination_type | exam_type_id→exam_types |
| vaccine | vaccine_id→vaccines |
| checkup_type | checkup_type_id→checkup_types |
| merchandise_item | merchandise_item_id→merchandise_items |
| manual_other | 全7FK NULL。正規manual schema（name/unit/price/tax/insurance/reason） |

各master型は対応FK非NULLかつ7列num_nonnulls=1、manual_otherは全列NULLとなるCHECKを持つ。裸のmaster_id polymorphic FKは禁止。manualは欠落/無効masterのfallbackにしない。

UK `(clinic_id,package_id,component_key)`。positionはUNIQUE `(clinic_id,package_id,position)` DEFERRABLE INITIALLY DEFERREDで原子的並替を許す。同masterの意図した重複を許すため `(package,master)` uniqueは作らない。親CAS下でfull replacementし、既存key更新・新規key追加・欠落keyのremoved_at/position=NULLを保存する。removed keyもuniqueに残し、別意味で再導入しない。同じcomponent編集はkey維持、削除後の新itemは新UUID。物理削除なし。

application/revisionはmutable定義行でなく不変snapshot＋component UUIDを参照する。read/previewはremoved_at IS NULLのみ。親lock下のwriterがlive component最低1件を検査する。

### 3.3 treatment_package_applications

| 列 | 型・制約 |
|:---|:---|
| id / clinic_id | 共通entity規約 |
| medical_record_id / pet_id | bigint NOT NULL、FK `(medical_record_id,clinic_id,pet_id)`→medical_records |
| package_id / package_version | bigint NOT NULL、package/clinic FK。package_version正数は歴史値で現在mutable versionへのFKにしない |
| operation_id | bigint NOT NULL、FK `(operation_id,clinic_id)`→operation_bindings。UK `(clinic_id,operation_id)` |
| version | bigint NOT NULL DEFAULT 1 CHECK(version=1)、applicationは不変 |
| snapshot_schema_version / snapshot | integer NOT NULL CHECK(=1) / jsonb NOT NULL object |
| preview_digest | bytea NOT NULL、length32 |
| component_count | integer NOT NULL、正数 |
| created_by / created_at | bigint NOT NULL staff FK / 共通作成時刻 |

<a id="source-record-patient-gate"></a>
package previewはsource recordのpet_idがNULLならselection段階のblockingとして追加不可を返す。財務unknownのwarningとは別のcaptured identity不成立であり、患者をclient入力や名称から補完しない。新規apply attemptは§6の既存lock順に従いmedical_records rowを取得した後、application/planの書込み前に同じ非NULL・医院/record/pet整合を再検査する。既存のrecord lock helperがpet NULLも拒否すると仮定しない。認可済み過去成功receiptのreplayは既定順序を維持する。本追補はpackageの既定NOT NULL契約をpreview/applyで表す設計であり、既存単品全体のgrant変更やruntime helper変更を本PRで実施しない。

UK `(id,clinic_id,medical_record_id,pet_id)`をplan由来FKのtargetにする。snapshotはpackage ID/版/名称、順序付きcomponent keys、typed refs/source stamps/quantity/財務値/provenance/uncertainty、生成plan IDsを含む。財務nullも保持し、結果本文や親請求をコピーしない。

application＋全plan＋revision1＋監査＋成功receiptは同TX。削除/訂正はplanへ作用し歴史applicationを書き換えない。保持は現行保持規約に従い、新しい年限を導入しない。

<a id="preview-digest-projection"></a>
### 3.4 preview_digestの投影とreplay比較

preview_digestはapplication snapshot全体のhashではない。version付きPreviewProjectionのRFC 8785 JCS bytesをSHA-256した32 bytesを保存する。wireのpreview_tokenは`pkg-preview-v1:`＋lowercase 64 hex、source tokenは`pkg-source-v1:`＋lowercase 64 hexとする。prefixはtransportだけで、保存byteaには含めない。いずれも認可能力tokenではない。

PreviewProjectionの全keyはschema（`pkg-preview-v1`）、record_id、clinic_id、pet_id、package_id、package_version、price_policy_digest、components。ComponentProjectionの全keyはcomponent_key、item、quantity、source_digest、classification。componentsはpackage順で重複品目を保持する。price_policy_digestは照合済みowner価格/税/保険規則の版又はdefinition digest、未確定ならnull。生成application/plan IDs、actor、取得時刻、編集session/epoch、plan collection CAS、表示整形、合計表示の可否は投影に含めない。

source_digestはSourceProjectionのJCS bytesのSHA-256 lowercase hex。全keyはschema（`pkg-source-v1`）、clinic_id、item、active、name、unit_label、native_version、price、owner_definition_digest、eligibility_digest。itemは§3.2の型付き参照を新APIのkind/型別ID又は正規manual fieldsへ写像する。priceの全keyはunit_price、tax_type、tax_rate、is_non_insurance、insurance_setting_state、price_source、tax_source、insurance_source、amount_basis、uncertainty。各sourceはmaster/owner-policy/manual/unresolved、amount_basisはincluded/excluded/exempt/unresolved、uncertaintyはfield（unit_price/tax/insurance）、reason（legacy_unset/not_configured_on_master/pending_owner_confirmation）、required_at（selection/execution/billing）の配列である。存在しない版・単位・規則はschema上の明示nullとし、取得不能を保存済み未設定に置換しない。

全key必須・未知key禁止。ID/version/円はcanonical整数文字列、quantity/tax_rateは丸めないcanonical十進文字列。object key順はJCS、array順は保持し、Unicode正規化・float変換・JSONB文字列hashは行わない。applyはlock下のserver実値から同じ投影を再構築し、確認したpreview tokenと照合する。生成plan IDsはその後application snapshotへ保存するが、このdigestへ追加しない。

C06 Contentはrecord_ref/package_id/package_version/preview_tokenで、operation_keyはbindingを選ぶ。Preconditionsはowner_expected_version（plan collection版）/edit_session_ref/edit_epoch。同keyのContent差異はC06の既存同key異内容拒否とし、別packageを保存しない。一時Preconditionsだけを更新する同内容再試行は既存C06条件に従う。既成功なら元targetの現在read認可を満たした保存receiptを返し、現在masterや現在preview digestを再計算して過去成功を否定しない。未成功attemptでsource/previewが変わった場合はdependency-changedとして再preview・明示確認し、内容変更は旧操作の結果確定後に新keyで行う。unknownのまま新keyに移らない。

実API/fixtureは別ownerが公開する。本節は採用投影の文書化であり、公開済み機械schemaを意味しない。

## 4. 新設補助表 treatment_package_catalog_versions

clinic_id bigint PRIMARY KEY→clinics RESTRICT、revision bigint NOT NULL CHECK>0、updated_at timestamptz NOT NULL。既存generic master表と偽らない。

schema/data migrationで既存各医院にrevision1を初期化する。これは新設の空catalogのCAS開始で、旧master履歴再構築ではない。以後の医院作成は同clinic-create TXで行を作る。catalog GETは永続rowをreadし既存API形のorder_version:Versionを返す。read INSERT・架空1・initializedフィールドは作らず、row欠落はsetup/invariant error。実行許可ではなく明示migration要件である。

catalog mutationは既存認可/lifecycle後にclinic row→catalog rowをlockする。createはそのlockで末尾追加（append時client order版不要）、reorderはexpected catalog revision必須。create/delete/reorderはmembership/orderを変えcatalog revisionを1回増分。text/components/active編集はpackage.versionのみ増分しorder版を増やさない。inactiveもreorder対象に残す。deleteはcatalogをpackageより先にlockする。

reorderは非削除全package（inactive含む）を対象に、動いたpackageのsort_order/versionを各1回変更する。no-opは版不変。全catalog writerは同順序へ参加し、row欠落をfallbackで隠さない。

## 5. 共通plan domain 3表

### 5.1 medical_record_plan_items

| 列 | 型・制約 |
|:---|:---|
| id / clinic_id / medical_record_id / pet_id | 共通id/clinic、record/petはNOT NULL bigint。applicationと同じrecord/patient複合FK |
| version / sort_order | bigint NOT NULL DEFAULT1 CHECK>0 / integer NOT NULL CHECK>=0。順序同値はidで解決、append阻害uniqueなし |
| origin_kind | text NOT NULL、single/package/verified_legacy |
| application_id / component_key | bigint NULL / uuid NULL、pair-null。origin=package iff両方非NULL |
| item_kind / 7FK | §3.2のexact matrix。mutable masterのactiveをDB CHECKへ入れず、locked adapterで現在選択認可/有効性を検査 |
| quantity / name_snapshot / unit_snapshot | 有限正数numeric NOT NULL / text NOT NULL btrim非空 / text NULL。単位なしと取得不能はsnapshotで区別 |
| 財務4列 | §2のnullable unit_price/tax_type/tax_rate/is_non_insurance |
| insurance_setting_state | text NOT NULL、explicit/legacy_unset/not_configured_on_master |
| source_snapshot / selection_snapshot | jsonb NOT NULL object |
| source_digest / native_master_version | bytea NULL（値があればlength32）/ bigint NULL（値があれば正数） |
| execution_domain | text NOT NULL、clinical/commercialのみ |
| clinical_state | text NULL、pending/performed/deferred/cancelled。clinical iff非NULL |
| created_by / updated_by / timestamps | 共通staff/時刻規約 |
| removed_at / removed_by | timestamptz NULL / bigint NULL、pair-null。論理削除のみ |

UK `(id,clinic_id,medical_record_id,pet_id)` と `(id,clinic_id)`。由来FKは `(application_id,clinic_id,medical_record_id,pet_id)`→applications。UK `(clinic_id,application_id,component_key)`で一component一plan、singleのNULL行は衝突しない。

classificationは単品owner/正規の明示用途で確定した値だけ保存する。名前推定なし。未解決はpreviewだけclassification=unresolved、owner-contract-unresolved/severity=blocking/required_at=selection、applicable=false。applyは全構成zero-write。財務unknownのnullable保存とは別であり、新plan wire execution.kind=unresolvedも作らない。旧unknownはlegacy read projectionに残し推測移行しない。

| DB execution_domain / clinical_state | wire execution.kind / state |
|:---|:---|
| clinical / pending | clinical / not_performed |
| clinical / performed | clinical / performed |
| clinical / deferred | clinical / deferred |
| clinical / cancelled | clinical / cancelled |
| commercial / NULL | commercial / selected（予定projection、owner fulfillmentとは別） |

逆変換はcoordinatorの認可済みintent内だけ。自由state PATCHは不可。後日の明示価格/数量変更はCASでcurrent snapshotを更新しrevision追加、元application snapshotは不変。unknownはcurrent財務列がNULLでも保存する。DBはname/master有効性や保険の意味を判定せずtyped serviceが単品と同段階で検査する。

健診set/検査panelはowner単位の予定1itemを保ち、後から結果子をlinkできる。generic再帰展開や親＋子二重請求はしない。

### 5.2 medical_record_plan_item_revisions

| 列 | 型・制約 |
|:---|:---|
| id / clinic_id / plan_item_id | 共通entity規約、plan_item_id bigint NOT NULL、plan/clinic FK |
| revision / previous_revision_id | bigint NOT NULL CHECK>0 / bigint NULL |
| change_kind | text NOT NULL、create/update/state_change/owner_link/owner_unlink/remove/restore |
| snapshot / reason | jsonb NOT NULL object / text NULL |
| operation_id / recorded_by / recorded_at | bigint NOT NULL operation composite FK / bigint NOT NULL staff FK / timestamptz NOT NULL |
| legacy_source_table / legacy_source_id | text NULL（値があれば非空）/ bigint NULL（値があれば正数）。pair-null、非NULLはrevision=1のみ |

UK `(id,clinic_id,plan_item_id)`、`(clinic_id,plan_item_id,revision)`。self FK `(previous_revision_id,clinic_id,plan_item_id)`→同表。revision1はprevious NULL、後続は非NULL。plan lock下のserviceでprevious=n−1と正確version増分を守り、gapなし。repositoryからrevision payloadのupdate/deleteを許さない。

legacy source一意化は **この表の列のみ** の部分unique `(clinic_id,legacy_source_table,legacy_source_id) WHERE revision=1 AND legacy_source_table IS NOT NULL`。predicateでplan.origin_kindや別表を読まない。初期allowlistは外来予定と証明できたtreatment_plansだけで、未対応sourceは拒否する。verified_legacy createはこの列とtyped snapshotの同じ根拠が必須、通常createではNULL。後続revisionはsnapshot内provenanceを保持しsource-key列を複写しない。

snapshotはafter-state＋typed event payload・previous identity・owner依存を含み、execution eventと結果到着を分ける。理由は既存理由必須intent（cancel/実施取消/誤link訂正/正規manual）で必須、普通の編集を一律必須化しない。enumの存在は新endpointの許可ではない。

PlanExecutionCoordinatorだけが状態writer。状態event/clinical_state投影・参加ownerの明示実施write・監査・receiptは同最外TX/CAS。結果到着は別に可能でperformedを自動変更しない。結果待ち実施に架空exam resultを作らない。機器受信/importは既存認可でresult/link/dependencyのみ更新し、人の編集session迂回にしない。D09段階receiptは段階のまま。

### 5.3 medical_record_plan_owner_links

| 列 | 型・制約 |
|:---|:---|
| id / clinic_id / medical_record_id / pet_id / plan_item_id | bigint NOT NULL（idはbigserial PK）、FK `(plan_item_id,clinic_id,medical_record_id,pet_id)`→plan captured identity |
| plan_revision_id | bigint NOT NULL、FK `(plan_revision_id,clinic_id,plan_item_id)`→revisions |
| owner_kind | text NOT NULL、treatment/vaccination/examination/checkup/billing_item |
| treatment_id / vaccination_id / examination_id / checkup_id / billing_item_id | bigint NULL、kindに対応するexactly-one。各 `(owner_id,clinic_id)`→実表treatments/vaccinations/exams/checkups/billing_items |
| owner_version / owner_revision_ref | bigint NULL CHECK>0（値あり時）/ jsonb NULL typed immutable reference |
| link_purpose | text NOT NULL、execution_source/result/charge_source |
| source_component_path | text NULL、owner定義のpanel/健診子identity。名称一致ではない |
| linked_by / linked_at / operation_id | NOT NULL、staff bigint / timestamptz / operation bigint composite FK |
| revoked_by / revoked_at / revoke_reason / revocation_operation_id | 一緒にNULL又は非NULL、取消理由非空。staff/時刻/operationの型とFKは共通規約 |

owner_versionのNULLは真の版を持たないlegacy ownerだけで明示provenance必須。1を捏造しない。official結果版は元link事実を変更せず後続link revisionで表現できる。

active partial uniqueは以下を固定する。

1. 5owner列それぞれに `(clinic_id,plan_item_id,owner_id,link_purpose) WHERE revoked_at IS NULL AND owner_id IS NOT NULL`。
2. 5owner列それぞれに `(clinic_id,owner_id,link_purpose)`、同じactive/nonNULL述語。同じ役割のownerを別planへ共有しない。同planのresultとcharge役割に同ownerを使っても別請求を生成しない。
3. `(clinic_id,plan_item_id) WHERE revoked_at IS NULL AND link_purpose='execution_source' AND owner_kind IN('treatment','vaccination','checkup','billing_item')`。primary concrete ownerを一つにし、examination結果の複数linkは許す。健診子examはresult、親はexecution_source。
4. `(clinic_id,plan_item_id) WHERE revoked_at IS NULL AND link_purpose='charge_source'`。型経路をまたいでも一componentのcharge ownerは一つ。実請求claimは既存billing_source_bindingsが保持し、結果linkだけで請求しない。

<a id="active-link-integrity"></a>
### Active linkの患者/record整合と訂正

全5型の無条件 `(owner_id,clinic_id)` FKは維持する。一方、owner-record/petの無条件歴史FKは作らない。revoked linkまで拘束して正規の後続患者訂正を妨げるためである。代わりにINITIALLY IMMEDIATEのactive-link integrity constraint trigger（SECURITY DEFINER/新GRANTなし）をlink insert/changeと関連owner parent-key変更へ設ける採用設計。

- link triggerはrevoked_at IS NULLだけを対象に現在clinic/owner record/petを照合する。treatmentはmedical_record_id＋親recordのpet、exams/vaccinations/checkupsは保存record/petと参照recordを確認。不明/NULL owner identityを入ってきたlinkから推定しない。
- owner triggerはexams/vaccinations/checkupsのmedical_record_id/pet_id、treatmentsのmedical_record_id、依存plan/linked ownerのあるmedical_records.pet_id変更で、active linkがcaptured record/petと不一致になる場合に拒否する。正規owner訂正TXで影響linkを明示revokeしてからreparentし、旧linkは不変履歴として残す。
- 全service writerは§6のshared gate/plan/owner participant順を取得する。trigger readは最終DB検査であり、trigger独自の逆lock順を作らない。reparent経路も閉包してからcutoverする。inactive履歴もid/clinic FKは保持し、owner clinicの付替えで迂回しない。
- planのrecord/petは非NULLなので、不明record/petのlegacy ownerは根拠ある既存訂正なしにlinkできない。この制限は予定選択を止めたり旧結果を消したりしない。

plan/application自身のcaptured parent identityは不変とする。根拠は新しいpackage制約ではなく[既存D03の患者取り違え訂正](../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-wrong-patient-linked-transfer)と[受入3AC](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/docs/spec/domains/medical-records/lifecycle.md#L449-L451)、[元/先recordを分ける既存設計](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/docs/design/domains/medical-records/technical-spec.md#L127)。記入済み患者IDを直接付替えず、医師確認の必要内容だけを正しい患者の適切な記録へ転記する。転記先は既存記録でもよく、常に新規作成を要求しない。

選択planは訂正provenance付きの転記先新identityとし、元application/planと元保存版/日付を保持する。package全体の自動複写・実施/請求生成をしない。link revokeだけでplan/applicationのpet_id rewriteやCASCADEを許さない。部分FKを装わずactive guardを使い、owner側の新quadruple uniqueは不要。既存treatment triple uniqueでrevoked履歴を拘束しない。

billing_itemsに存在しないmedical_record_id/pet_id契約を捏造しない。billing_item/clinic FKを保持し、billing親をlockして既存adapterでowner/pet/source bindingを確認する。保護対象のbilling reparent/source訂正も同adapterでactive linkをrevoke/revalidateし、履歴を黙って更新しない。この機能のためにbilling_items.pet_idを追加しない。

revokeは単調で旧row保持、訂正は同CAS TX内でrevoke後に新linkを追加する。unlink/cancelで結果や確定会計を削除しない。健診fieldは別owner/chargeにせず、過去結果linkからperformedを推定しない。

<a id="lock-order"></a>
## 6. Lock順・CAS・receipt境界

package独自gateでなく既存C06/494順序を使う。

1. 認証済みselected clinicを一度解決する。nonlocking早期認可は可能だがrecord IDでclinic contextを変更しない。
2. 短いC06 claim TXはbinding＋494の承認済みimmutable target anchorだけを書く。anchorへrecord/master business FKを足さず、shared gateより先にbusiness rowをlockしない。claim TXを終了する。
3. replay/unknownは元target・現在read認可・保存receiptを先に確認し、認可済み過去成功を返すため現在masterを再構築しない。
4. 新business TXのRunner順は `p.Gate(ctx)`→`LockBindingNoWait`（SELECT FOR UPDATE NOWAIT）→保存result/unknown照会（現在認可によるdisclosure）→apply savepoint→`p.Apply`。busy bindingはC06 in-progressで二人目のwriterを作らない。Apply内は既存lifecycle保護→record_edit_rights row→medical_records row→既存plan rows ID昇順→dependency_versionsのplan_items row。baseline欠落は494 acquire/take-over成功TXだけが既存plan lock後に準備し、applyで遅いbootstrapをしない。全plan writerが同順序に参加し、content変更ごとにdependencyを1回増分する。
5. package row FOR SHARE→source masterをconsultations/procedures/medicines/exam_types/vaccines/checkup_types/merchandise_itemsの表順・ID昇順→owner定義子ID昇順でlockする。master writerも既存parent→childと互換表順を保ち、別TXを始めるserviceを呼ばない。
6. package版/active・typed refs・scope/grants・数量/source digest・preview hashを再検査する。単品で許容された財務nullは保持し、application/plan/revision1/監査/成功receipt＋dependency1回増分を同TXでcommitする。applyはcatalog lock不要。選択内容はpackage rowで保護しorderは価格/入力内容ではない。
7. execution/link TXはplan/dependency後に既存owner順で具体owner rowもlockする。owner起点writerは自分のrow取得前に共通coordinatorへ参加する。全閉包実装後だけpublish/cutoverする。billing親→items順、D09の段階TXを保持し、外部機器/mail/print呼出をDB TX内に保持しない。

根拠は[494 §5/5.1](emr-494-edit-draft-contract-20261020.md)、[Runner Gate/NOWAIT/receipt/unknown/savepoint](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/operation/runner.go#L294-L329)、[store UPDATE NOWAIT](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/operation/store.go#L114-L122)。494のbinding→lifecycle→edit-right→owner/依存子の短縮表記でも、外側Runnerのbinding NOWAITを省かない。edit-right-before-recordを維持し、binding保持後のshared gate後取りをしない。lookup認可と短claimは別phase。

catalog writeは既存auth/lifecycle gate→clinic→catalog version→package IDs昇順→component。component更新はpackage FOR UPDATE、参照masterを検査lockする場合は上記typed順。reorderはrecordを取得しない。applyがrecordを保持しcatalog待ち、catalogがrecord待ちとなる逆順を作らない。

no-opは架空revisionなし。business rejectionはplan/application/成功receiptをzero-write（許容C06 attempt metadataは部分packageではない）。unknownは元operation lookupへ戻し新keyで盲目的再送しない。

## 7. RLSと既存action権限

6domain＋catalog版の全7表に既存tenant_clinic_isolation helperとhas_clinic_access(clinic_id)を適用する採用設計。新GRANT/BYPASS role/鍵/接続設定なし。既存helperはENABLEでFORCEではないため強い保証へ読み替えない。repoはscoped queryとtrusted selected clinicを維持し、RLSの許可医院配列をrecord IDによる医院切替許可にしない。

<a id="non-force-rls-boundary"></a>
非FORCE RLSはtable owner・superuser・BYPASSRLS roleには適用されない。本書は本番接続roleがそれらに該当しないことを確認済みとはせず、ENABLEだけをtenant隔離の証明にしない。既存runtime接続roleと7表のownership/BYPASS属性を既存導入工程で照合し、その境界を確認できる場合だけRLSを追加の防御層として扱う。未確認の場合もscoped query・trusted selected clinic・各action認可は必須であり、RLSで代替しない。本PRはrole作成/変更、GRANT/FORCE、接続設定変更、DB適用、試験実行を行わない。実環境の境界確認や変更が必要なら既存担当工程へ分離し、未検証を検証済みと記録しない。

package CRUDはResourceMasterMedicalの各action。source lookupは個別resource（ResourceCheckups/ResourceMasterMerchandiseを含む）を追加確認する。予定追加applyは既存単品POST同様ResourceMedicalRecords.create＋現在edit-session/fence＋単品選択条件を使用し、ResourceMedicalRecords.editやmaster管理createを追加要求しない。edit-sessionは並行性制御でaction grantと別。既存item変更は各既存edit/delete対応、実施/結果は各owner既存権限。歴史applicationも現在record認可が必要。

根拠は[現行routes_records.go](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/medicalrecord/routes_records.go#L44-L48)と[D03権限表](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/docs/spec/domains/medical-records/lifecycle.md#L392-L395)。POST treatments=create、PATCH/PUT=edit、DELETE=deleteを踏襲する。RLSはaction認可の代替ではない。

## 8. 3master nullable税保存subset

exam_types/vaccines/checkup_typesだけにtax_type tax_type NULL、tax_rate numeric NULL（値あり時有限CHECK）、defaultなしを追加する実装契約。UPDATE backfillなしで旧行はNULL。既存単品master writerの組合せ検証を守り、package独自の税分類/免税判断や医院ごとの設定値を決めない。部分税編集を提供するwriterはPATCH省略と明示nullを区別する。

modelはnullable税enum pointer＋*ExactNumeric、DTOは正確decimal文字列/nullと省略追跡を使いfloat64経由なし。repository SELECT/INSERT/PATCH、response、OpenAPI、監査before/after、native版CAS/source digestを同subsetで整合する。OpenAPIの実編集は単一ownerへ引継ぐ。vaccines/checkup_typesは真のversion、版のないexam_typesは既存source digestを用いnative versionを捏造しない。広いmaster再設計や金額decimal移行をせず、価格はint64/bigint整数円のまま。

単品で許される税unknownはwarning/provenance（required_atはbilling）で選択可能。請求ownerだけが必要transitionをblockする。物販に保険propertyを新設せずnot_configured_on_master、legacy nullableはlegacy_unset、新masterに明示保存済み設定は実値保持。インフラread失敗を財務nullにしない。

<a id="delivery"></a>
## 9. 公開順・実装チェック（全件未実装/未実行）

設計判断を実装担当へ戻す未決事項はない。以下の未完は実装・契約公開・統合の作業を意味する。

| 順序 | 担当・必要subset |
|:---|:---|
| 0 機械契約公開 | 親の単一OpenAPI ownerが選定API/DTO/既存権限/C06/JCS/7型resolverと全型fixtureを同版で先行公開。backend完了待ちにしない。本PRでapi.yamlは編集しない |
| 1 共通plan基盤 | EMR-516窓口に一本化。494 anchor/依存版/編集権subsetを前提に、6domain表＋新catalog表、2master複合unique、active整合guard、catalog初期化/clinic-create参加、既存helper RLS、model/ERD/field-null整合を実装 |
| 2 master/catalog | package CRUD/catalog CASと3master nullable税writer/read/model/API subsetを整備。値backfill/defaultなし、既存unknownを理由に一般予定選択を止めない |
| 3 writer閉包/apply | 共通plan repo/coordinatorへsingle create/edit/reorder/remove/copy/state/link全writer参加。package CRUD/preview/apply/read/lookupを同storeへ接続し、監査/receipt snapshot、全7typed/正規manual fixtureを整備。閉包前dual-writer切替なし |
| 4 FE契約mock | 0公開後、1〜3と並列でmaster/カルテ契約mock PRを作れる。公開API版を使用し独自DTO/架空generic-master APIを作らない。実API結合/本番有効化だけを対応backend必要subset後に行う |

legacy変換はopt-in・根拠付き・冪等。revision1に物理source table/id、元type/ref/値、証拠、mapping operationを保存し§5.2のsource-key uniqueで同source再移行を防ぐ。名称でmaster、状態類似でcompleted、価格0/税/保険false/clinic/pet/実施時刻を推定しない。未照合行は明示legacy compatibility projectionで可読のまま残し、別legacy/new identityを検証変換まで区別する。同名/同日dedupeなし。

既存入院treatment_plans/care_plan_itemsは現在契約のまま。移行は実施/stock/invoice/print/通知/機器副作用や履歴状態一括書換えを行わない。本書はDB実行・grant拡張・実鍵作成・試験実行を許可するものではない。

## 10. 既存ACへのstorage確認観点（新ACを追加しない）

[既存12AC](../../spec/domains/medical-records/treatment-packages.md#package-acceptance)のfixture条件を次へ具体化する。全件NOT_RUN。

- 8参照選択肢、clinic FK、manualのexactly-zero FK、master/package論理削除後の歴史snapshotを保持する。
- exact小数数量/整数円と財務nullをpreview/apply/readで保持し、selectionだけに厳しい税/保険gateを作らない。
- application＋N plan＋N revision1＋dependency1回＋receiptを原子的保存し、replayで増やさない。
- 複数結果と単一charge ownerを両立し、owner訂正で患者/recordを黙って変更せず、performedはcoordinatorのみ。
- digestはJCSでありJSONB serializationではない。初期/再送snapshotは後日のsource変更に依存しない。
- catalog CASとplan CASを別々に永続化し、readはrow生成せず、既存catalog表を捏造しない。

## 11. 静的根拠と検証境界

[ExactNumeric実装](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/model/exact_numeric.go)、[D10正本](../../spec/domains/accounting/lifecycle.md)、[D10設計](../../design/domains/accounting/technical-spec.md)、[既存master権限](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/medicalrecord/routes_masters.go)、[物販権限](https://github.com/MinoruSoga/AnimalEkarte/blob/a28f81bfe99010c440ff42a51b8f31811ae1904f/backend/internal/inventory/routes.go)を参照する。

前回業務境界は[PR #755](https://github.com/MinoruSoga/AnimalEkarte/pull/755)でmain統合済み。本PRは財務unknown gateと上記storage設計の公開だけを行う。試験/fixture実行・runtime・Docker・DB適用・実鍵はNOT_RUN。静的リンク/差分確認だけを行い、SQL/RLS実装済みやFE機械契約公開完了とは記録しない。Plane/mergeは親担当。

静的確認記録: 変更4文書の追加/変更内容にあるrepositoryリンク29件の対象パス・固定SHA・アンカーを確認し、欠落0件。業務件数は4FR/12ACのまま。内部原稿名/メッセージID/作業pathの転載なし。`git diff --check`は問題なし。`python3 -B scripts/checks/verify-agent-task.py --base a28f81bfe99010c440ff42a51b8f31811ae1904f --plan` はdocs4件、checks=[]、unmapped=[]、executed_count=0で、試験実行ではない。
