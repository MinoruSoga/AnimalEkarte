# 予約ID・パッケージの限定静的監査（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-reservation-packages-20261023 |
| 種別 | 限定静的監査の受領記録 |
| 決定状態 | 既存採択仕様を参照。新業務・技術契約の採択なし |
| 実装状態 | 以下の固定範囲だけを分類。全票・全ACの実装完了認定なし |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23（資料受領・静的整形） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

基準main: 171b56575e7ee028e54358dad22ae28574853125（GitHubコネクタ確認）。製品実装・試験・DB・Plane書込みなし。

<a id="audit-reservation-packages-conclusion"></a>
## 結論
- 532/#820でS06/S07/S08/S10は補完済み。古い監査コメントを現残差として再起票しない。
- 521：既存返却17メソッドのDBOrTx inventory/原子性証拠補完。実行上の原子性欠陥を観測したという意味ではない。
- 523：既存P01 header戻るdirty guardが未接続。523/450の既存残件。
- 524：成功後application/common-planのread subscriber/再表示が未接続。これは本文が要求する契約mock完結の範囲で、runtime未提供だけを理由に残す部分ではない。runtime gateはfalseのまま保つ。
- 519：採用段階1のみ。旧client退役/配信証拠を受けてproducer string切替、その後numeric互換除去。既存519の依存残件で、直ちに切替を求めない。
- 520：reader/paired projectionは実装済み。PRが明記するDB障害注入fixture未追加を区別する。実DB/RLS受入未確認をproduction未実装とは扱わない。

<a id="audit-reservation-packages-coverage"></a>
## 読取coverage
Plane本文と全コメント: {"519":8,"520":1,"521":10,"523":7,"524":3}。主要実装PR775/776/820のメタデータ・discussion取得。関連PR検索、最新mainの主要sourceを直接取得。全repository/全runtime挙動の網羅保証ではなく、未精査枝はunverified。Ancillary handoff PRは全て全文再読したわけではない。

<a id="audit-reservation-packages-524-in-scope"></a>
## 524 in-scope根拠
最新本文: 「package検索/内訳preview/行別警告→明示apply→receipt照会→application/common-plan再表示まで公開契約mockで完結。」
最新AC: 「成功後application IDsを確認し共通plan一覧へ再取得。再取得失敗をapply失敗にして再追加しない。」
#775採用回答はruntime gateをfalseとするが、上記mock完成条件を取り消していない。3タブの実績consumerは別ownerであり、この残差の修正へ含めない。既存use-treatment-package-apply.test.tsx成功ケースはPOST/stateをassertするが再取得GET/再表示をassertしない。単にテスト実行されていないという所見ではない。

<a id="audit-reservation-packages-ac"></a>
## AC分類

<a id="audit-reservation-packages-519-ac1-producer-precision-dependency"></a>
### 519 — AC1 producer precision: dependency
Both IDs remain uint64 JSON number by adopted stage-1. Old client retirement/deployment evidence needed before two producer string switch; numeric compatibility removal follows separately.
根拠: [backend/internal/reservation/liff_response.go:126-144,264-270](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/liff_response.go#L126)
重複/引渡し: 519 existing rollout A

<a id="audit-reservation-packages-519-ac1-consumer-precision-ac2-no-collisions-implemented"></a>
### 519 — AC1 consumer precision / AC2 no collisions: implemented
Canonical int64 strings and positive safe legacy numbers normalize to string; unsafe numbers rejected; boundary fixture exists.
根拠: [frontend/line-reserve/src/api/schemas.ts:100-127](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/api/schemas.ts#L100); schemas.test.ts
重複/引渡し: #776

<a id="audit-reservation-packages-519-ac3-malformed-ids-implemented"></a>
### 519 — AC3 malformed IDs: implemented
Stage-1 permits only safe positive legacy numbers by superseding answer; original blanket number rejection is stage-3 criterion.
根拠: [frontend/line-reserve/src/api/schemas.ts:100-127](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/api/schemas.ts#L100)
重複/引渡し: #776

<a id="audit-reservation-packages-519-ac4-invalid-201-no-blind-post-implemented"></a>
### 519 — AC4 invalid 201 no blind POST: implemented
Creation-unknown path and reservation-list guidance retained.
根拠: [frontend/line-reserve/src/pages/ConfirmPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/pages/ConfirmPage.tsx)
重複/引渡し: #776

<a id="audit-reservation-packages-519-ac5-existing-cancel-notes-auth-and-ac6-no-scope-bypass-unverified"></a>
### 519 — AC5 existing cancel/notes/auth and AC6 no scope bypass: unverified
Response notes and scoped reservation retrieval present; this audit did not exhaust every cancellation handler branch.
根拠: [backend/internal/reservation/liff_response.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/liff_response.go); liff_service_reservations.go
重複/引渡し: existing519

<a id="audit-reservation-packages-519-ac7-old-new-staging-and-gate-implemented"></a>
### 519 — AC7 old/new staging and gate: implemented
false gate retained; full string producer completion not claimed.
根拠: [frontend/line-reserve/src/lib/revisit-intake-activation.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/lib/revisit-intake-activation.ts)
重複/引渡し: 471-473 activation dependency

<a id="audit-reservation-packages-520-authorized-same-snapshot-current-name-implemented"></a>
### 520 — authorized same snapshot current name: implemented
Appointment clinic/customer + nondeleted owner/pet + linked binding + same common pet/current revision in one SQL.
根拠: [backend/internal/reservation/appointment_admin_repository.go:188-227](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/appointment_admin_repository.go#L188); [backend/internal/pet/common_basic_name.go:9-30](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/pet/common_basic_name.go#L9)
重複/引渡し: #776 / 491 reader only

<a id="audit-reservation-packages-520-paired-nullable-public-identity-no-201-pair-implemented"></a>
### 520 — paired nullable public identity; no 201 pair: implemented
Paired string/null and limited legacy normalization; no customer-field name fallback.
根拠: [backend/internal/reservation/liff_response.go:130-144,248-270](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/liff_response.go#L130); [frontend/line-reserve/src/api/schemas.ts:130-170](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/api/schemas.ts#L130)
重複/引渡し: #776

<a id="audit-reservation-packages-520-target-match-display-unavailable-distinction-implemented"></a>
### 520 — target match display / unavailable distinction: implemented
Name bound to matching pet target; reader error propagates.
根拠: [frontend/line-reserve/src/pages/RevisitIntakePage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/pages/RevisitIntakePage.tsx); [backend/internal/reservation/liff_service_reservations.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/liff_service_reservations.go)
重複/引渡し: #776

<a id="audit-reservation-packages-520-db-failure-injection-fixture-missing"></a>
### 520 — DB failure injection fixture: missing
Explicitly acknowledged real DB fault injection fixture not added; production error propagation exists. Finite fixture gap, not runtime error-handling absence.
根拠: PR #776 residual section
重複/引渡し: existing520; no new feature

<a id="audit-reservation-packages-521-ac1-8-typed-branches-distinct-keys-scoped-active-source-implemented"></a>
### 521 — AC1 8 typed branches, distinct keys, scoped active source: implemented
Typed source authorization and validation; manual branch separately validated.
根拠: [backend/internal/medicalrecord/treatment_package_service.go:182-319](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/treatment_package_service.go#L182)
重複/引渡し: #775

<a id="audit-reservation-packages-521-ac2-replace-cas-history-full-reorder-implemented"></a>
### 521 — AC2 replace/CAS/history/full reorder: implemented
Full uncapped locked set now used, duplicate/missing/foreign IDs rejected before writes; soft-delete/history code retained.
根拠: [backend/internal/medicalrecord/treatment_package_service.go:393-779](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/treatment_package_service.go#L393); treatment_package_repository.go:213-232
重複/引渡し: 532/#820 fixes S06

<a id="audit-reservation-packages-521-ac3-nullable-financial-values-implemented"></a>
### 521 — AC3 nullable financial values: implemented
Definition validates source identity/active/access, not invented financial defaults.
根拠: [backend/internal/medicalrecord/treatment_package_service.go:204-319](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/treatment_package_service.go#L204)
重複/引渡し: #775

<a id="audit-reservation-packages-521-ac4-catalog-bootstrap-clinic-test-double-implemented"></a>
### 521 — AC4 catalog bootstrap/clinic test double: implemented
CSV catalog insertion in existing TX; atomicClinicStore catalog methods now present.
根拠: [backend/cmd/migrate/csvbundle.go:156-189](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/cmd/migrate/csvbundle.go#L156); [backend/internal/clinic/clinic_service_atomicity_test.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/clinic/clinic_service_atomicity_test.go)
重複/引渡し: 532/#820 fixes S10

<a id="audit-reservation-packages-521-ac5-unknown-create-no-same-name-recovery-implemented"></a>
### 521 — AC5 unknown create no same-name recovery: implemented
Unknown create blocks repeated save; no new master C06 rules.
根拠: [frontend/src/features/master/hooks/use-treatment-packages.ts:203-245](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/hooks/use-treatment-packages.ts#L203)
重複/引渡し: #775

<a id="audit-reservation-packages-521-dbortx-inventory-registration-atomicity-proof-missing"></a>
### 521 — DBOrTx inventory registration/atomicity proof: missing
Only new LockAllPackagesForReorder registered, original 15 package methods + 2 clinic catalog methods absent. Need register 17 existing methods and map each to ambient-TX atomicity proof, supplement missing proofs; not allowlist-only.
根拠: [backend/internal/lintscan/dbortx_inventory_lint_test.go:80-86](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/lintscan/dbortx_inventory_lint_test.go#L80),259; [backend/internal/medicalrecord/treatment_package_repository.go:61-390](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/medicalrecord/treatment_package_repository.go#L61); [backend/internal/clinic/clinic_repository.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/clinic/clinic_repository.go)
重複/引渡し: 521 comment0321dede already returned; #820 excludes legacy methods

<a id="audit-reservation-packages-523-ac1-load-empty-error-access-delayed-scope-implemented"></a>
### 523 — AC1 load/empty/error/access/delayed scope: implemented
403 latch, explicit fresh authorization recovery, draft hidden/retained; current source includes #820.
根拠: [frontend/src/features/master/routes/TreatmentPackageSettings.tsx:77-242](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/routes/TreatmentPackageSettings.tsx#L77)
重複/引渡し: 532/#820 fixes S07

<a id="audit-reservation-packages-523-ac2-8-types-exact-string-quantities-implemented"></a>
### 523 — AC2 8 types/exact string quantities: implemented
Typed drafts preserve quantities and component identity.
根拠: [frontend/src/features/master/lib/treatment-package-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/lib/treatment-package-form.ts)
重複/引渡し: #775

<a id="audit-reservation-packages-523-ac3-replace-draft-conflict-compare-implemented"></a>
### 523 — AC3 replace/draft/conflict compare: implemented
409 fresh GET, explicit compare/adopt/keep, stale response token; no automatic overwrite.
根拠: [frontend/src/features/master/routes/TreatmentPackageSettings.tsx:110-146,304-357](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/routes/TreatmentPackageSettings.tsx#L110)
重複/引渡し: 532/#820 fixes S08

<a id="audit-reservation-packages-523-ac3-dirty-header-back-missing"></a>
### 523 — AC3 dirty header back: missing
Header directly navigates settings; local discard wrappers cover editor switch but not shell back; beforeunload is not SPA navigation guard.
根拠: [frontend/src/features/master/components/MasterPageShell.tsx:65](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/components/MasterPageShell.tsx#L65); [frontend/src/hooks/use-side-peek-dirty.tsx:31-35](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/hooks/use-side-peek-dirty.tsx#L31); [frontend/src/hooks/use-unsaved-changes.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/hooks/use-unsaved-changes.ts); TreatmentPackageSettings.tsx
重複/引渡し: existing523 comment2286c5c6 /450 P01; explicitly excluded532

<a id="audit-reservation-packages-523-ac4-reorder-and-create-unknown-implemented"></a>
### 523 — AC4 reorder and create unknown: implemented
Conservative cap guard + permission/CAS errors, unknown create no blind resubmit.
根拠: [frontend/src/features/master/routes/TreatmentPackageSettings.tsx:398-452](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/routes/TreatmentPackageSettings.tsx#L398); use-treatment-packages.ts:203-245
重複/引渡し: #820 / #775

<a id="audit-reservation-packages-523-ac5-production-api-versus-mock-implemented"></a>
### 523 — AC5 production API versus mock: implemented
Uses API mutations, no production fixture-success fallback found in inspected route/hooks.
根拠: [frontend/src/features/master/hooks/use-treatment-packages.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/master/hooks/use-treatment-packages.ts)
重複/引渡し: #775

<a id="audit-reservation-packages-524-ac1-8-type-preview-warnings-null-total-implemented"></a>
### 524 — AC1 8-type preview/warnings/null total: implemented
All 8 labels, clinical/commercial display, row issues and unknown total preserved.
根拠: [frontend/src/features/medical-records/components/TreatmentPackagePreview.tsx:11-94](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/components/TreatmentPackagePreview.tsx#L11)
重複/引渡し: #775 contract mock

<a id="audit-reservation-packages-524-ac2-explicit-key-unknown-retry-identity-implemented"></a>
### 524 — AC2 explicit key/unknown/retry identity: implemented
Pending map, matching key/record, no repeated submit while unresolved; success versus non-success separated.
根拠: [frontend/src/features/medical-records/hooks/use-treatment-package-apply.ts:139-256](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/hooks/use-treatment-package-apply.ts#L139); lib/treatment-package-operation-state.ts:123-178
重複/引渡し: #775 contract mock

<a id="audit-reservation-packages-524-ac3-grant-context-stale-scope-implemented"></a>
### 524 — AC3 grant/context/stale scope: implemented
UI gate requires create/context/applicable; late outcomes scoped. Real edit-context supplied separately.
根拠: [frontend/src/features/medical-records/lib/treatment-package-operation-state.ts:92-110](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatment-package-operation-state.ts#L92); hooks/use-treatment-package-apply.ts:117-166
重複/引渡し: 494 context dependency

<a id="audit-reservation-packages-524-ac4-successful-application-common-plan-redisplay-missing"></a>
### 524 — AC4 successful application/common-plan redisplay: missing
Read hooks are only declarations in connector search. Success UI shows application ID only. invalidateQueries has no active read subscriber here, so mock success→GET applications/plan→redisplay/failure state is not wired. Hook success test asserts POST/state only, not subsequent GET or rendered plans.
根拠: [frontend/src/features/medical-records/hooks/use-treatment-package-apply.ts:39-57,180-197](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/hooks/use-treatment-package-apply.ts#L39),237-249; components/TreatmentPackagePicker.tsx:65-72; components/MedicalRecordDiagnosisPlan.tsx:422-430
重複/引渡し: existing524 in-scope mock; not532; BE dependency separate

<a id="audit-reservation-packages-524-ac5-no-execution-billing-side-effects-implemented"></a>
### 524 — AC5 no execution/billing side effects: implemented
Only package apply and queries; no execution/result/billing writers invoked in inspected slice.
根拠: [frontend/src/features/medical-records/components/TreatmentPackagePreview.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/components/TreatmentPackagePreview.tsx); hooks/use-treatment-package-apply.ts
重複/引渡し: downstream owners516/517/500

<a id="audit-reservation-packages-524-ac6-three-tab-handoff-dependency"></a>
### 524 — AC6 three-tab handoff: dependency
Common query invalidation present; actual tab consumers belong516/517/500 and must receive plan identity, not fabricate treatments.
根拠: [frontend/src/features/medical-records/hooks/use-treatment-package-apply.ts:191-197](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/hooks/use-treatment-package-apply.ts#L191)
重複/引渡し: 516/517/500

<a id="audit-reservation-packages-524-ac7-false-runtime-gate-no-fixture-fallback-implemented"></a>
### 524 — AC7 false runtime gate/no fixture fallback: implemented
Runtime false and editContext null intentional; do not activate as part of mock completion.
根拠: [frontend/src/features/medical-records/lib/treatment-package-operation-state.ts:23](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/medical-records/lib/treatment-package-operation-state.ts#L23); MedicalRecordDiagnosisPlan.tsx:365,424
重複/引渡し: 522/516/494 required subsets

<a id="audit-reservation-packages-all-runtime-build-test-db-uat-this-audit-not-run"></a>
### all — runtime/build/test/DB/UAT this audit: NOT_RUN
No tests/build/codegen/DB/browser were run. #820 reports limited prior FE tests, not whole domain acceptance.
根拠: audit procedure
重複/引渡し: not implementation failure
