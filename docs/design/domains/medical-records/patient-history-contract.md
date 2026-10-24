# EMR-451 H01 ペット別カルテ内履歴の技術引渡し

基準main `4b7ea65a3b692073bb41c4be52bd2c1a9be99d74`、2026-10-24。業務scopeは[カルテ内履歴](../../../spec/domains/medical-records/pages/medical-record-form.md)と[共通UI](../../../spec/shared/ui-contracts.md)・[適用台帳](../../../spec/shared/interaction-scope.md)の既決「このペット」「閲覧認可内全医院を初期値」「明示選択保持・失権除外」。以下は技術選定でありruntime実装・受入済みではなく、新FR/ACを作らない。

## H01専用callerと認可producer

`/medical-records/:id`（新規は `/medical-records/new?petId=`）→MedicalRecordFormReadyPanels→useGetPetMedicalHistory→MedicalRecordInterview/InterviewHistoryを対象とする。現GET /v1/medical-records＋pet_id＋X-Clinic-IDとFEの同local pet/clinic絞込みを専用readerへ置換する。一般管理一覧、owner-report `/pets/:id/treatment-history`、identity-links連携履歴、trimming履歴は別callerで一律変更しない。

認可constantはResourceMedicalRecords、wireはmedical-records:view。medical_recordsはDB名でgrant名ではない。現在所属かつ現在同resource:viewを持つ医院集合をserver生成し、GET /me必須medical_records_view_clinic_idsとFE型/変換まで同一実装PRで供給する。owners_view_clinic_ids・identity-links:view・membership単独で代用しない。[]は認可0件、欠落/null/解決異常は利用不可としてfetch停止。

同一個体の唯一の正本はcommon_pet_subjectsとcommon_pet_clinic_bindingsの確認済み(clinic_id,common_pet_id,local_pet_id)。認可済みseed pairをsubjectへ解決し、現在認可内のそのsubjectの院別pairを検索する。clinicとlocal petを独立INにせず相関pairで照合。同名/同owner/電話/旧active pet groupで未確認・不一致・欠損を補完しない。旧identitylink相関SQLは実装パターンだけの参考で旧groupをauthority/fallbackにしない。DB再作成前提で旧互換・移行・GET backfillは追加しない。491 subject/binding reader必要subsetを同PR供給又は固定採用SHAで受領し、491全面Done/basic writer全部待ちにはしない。

## 専用APIとbounded検索

GET `/api/v1/medical-records/patient-history` を静的routeとして/:idより先に登録。queryはseed_clinic_id/seed_pet_id/clinic_ids/q/exclude_record_id/limit/cursor。IDは正canonical decimal string、limit既定50・上限100、qはtrim後最大200文字、不正400。serverが選択院を現在認可集合で検査して不正院はfail closed、選択0院はfetchしない。seed現在認可・同一性・各行clinic/pet/recordも照合し、旧/meを恒久grantにせず毎request再認可、失権本文非開示。

対象pair・削除除外・終了条件・qをserver適用後にpage化。qは既決の治療内容/処方薬名/処置名に限定し一般一覧へ拡張しない。検索正規化索引、(clinic_id,pet_id,date,id)読出し索引、必要子owner結合索引を同PRへ含める。FE全record検索・無制限page走査・N+1全件detail取得を禁止。date DESC,id DESCのkeyset、limit+1でhas_moreを求め転送をboundedにする。cursorはseed/選択院/q/excludeに束縛し不一致拒否、scope変更で先頭へ。通常次pageありと取得失敗/不完全を分ける。

## 終了証拠と最小表示

[495 B0契約§0.1](../../../work/emr/emr-495-consultation-correction-contract-20261021.md)の同record最大revisionのconsultation_head/effective_end/evidence_stateをbatch readで再利用しknownかつeffective_end=trueだけを対象とする。終了取消/進行中/現在recordを除外、unknownを旧finalized/支払済みで代用しない。unknown除外があれば「終了状態未確認の記録あり」を返し完全な履歴なしと誤表示しない。一般管理一覧の進行中閲覧は維持。

responseはscope/identity/終了証拠の状態、items/has_more/next_cursor。itemはclinic_id/pet_id/medical_record_id/date、既存問診/担当表示/体重の必要字段、検索matchの治療内容/薬名/処置名だけ。金額/unit_price/請求/未納/入金/預り金なし。価格入りidentitylink DTOを流用しない。[共通基本情報](../../../spec/domains/owners-pets/lifecycle.md)の最新共有と院別診療/請求/預りの独立帰属を混同せず、閲覧で過去原本/証明書/会計を変更しない。

## FE・write境界と一つの実装PR

H01専用account/staff別医院選択保持とURL検索を接続。初回認可全院、復帰は明示選択保持、失権院除去、全失権/明示0院で全院fallback禁止。cache keyはactor/seed pair/認可・選択scope/q/cursor。scope/患者/actor変更時は旧request abort＋遅延応答discard。詳細は実record院/IDで再認可。横断readでcopy/edit権を増やさず、現copyの院/患者/版/編集権検査を維持。trimmingは別resource:view readerのまま。

一つの実装scopeへH01認可集合producer、491必要reader、495終了batch subset、索引/API/OpenAPI/DTO、FE選択/検索/page/詳細を収容。451は445/452へmappingを返し463飼主一覧と重複PRを作らない。共有hunkは単一editor又は固定採用SHAで統合、491 split A承認と既存D6を覆さずsubset/hunkを分ける。新API/requiredfieldは選定契約で、現api.yamlへ実装済みとして追記せずproducer/consumer同時置換を実装PRに含める。

fixtureは2院同subject/同名別subject/認可差/0院・失権/binding欠損矛盾/ended・undo・unknown/2ページ目match/上限/cursor不一致/途中失敗/遅延/account切替/戻る/他院詳細/readでwriteゼロ。全て定義だけ。製品/API YAML/DDL/生成・試験/build/DB/Docker/STG/外部更新/mergeはNOT_RUN。

固定mainの実接点：backend/internal/model/permission.go、frontend/src/types/generated/models.ts、frontend/src/hooks/use-medical-records.ts、backend/internal/identitylink/handler.go・repository.go、[binding構造](../../../architecture/erd-detail/common_pet_clinic_bindings.md)。これらの存在は今回の新reader供給済み証拠ではない。
