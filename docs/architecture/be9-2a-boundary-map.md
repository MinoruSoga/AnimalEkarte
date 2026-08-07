# BE9-2A: domain boundary map — target package boundaries、依存グラフ、ADR-006 の入力

> 対象: 旧BE-refactor.md BE9-2A（2026-07-24退役・経緯はgit履歴）。
> 実行日: 2026-07-19。手法: codegraph(callers/callees/explore) + grep/rg + git log を第一手段とする再実測（旧 BE8/§9 の file-prefix 分類はそのまま正本にせず、再実測との差分を§6に記録）。
> 分類マニフェスト（全761 production Go source row、未分類0件）: [be9-2a-classification-manifest.csv](be9-2a-classification-manifest.csv)。移行後target packageの物理file数とは別指標。
> **本docは分類とboundary inventoryの正本**。target設計の裁定は[ADR-006](adr/006-backend-domain-package-boundaries.md)。BE9移行は2026-07-24にcode complete（release pending）となり、release gateは[`q&a.html` OPS-13〜17](../../q&a.html#ops)を正本とする。
> **2026-07-24 final recensus**: 本文のcall-site、fan-in/out、file pathは明示がない限りBE9-2A開始時snapshotであり、現行作業listではない。13 target packageは全て移行済みで、BE9はcode complete / release pending。

## 0. 結論（Success Criteria 対応）

- 全 761 production Go source row（BE9-2A時点の移行元、`_test.go`・`cmd/_archive` 除外）が target package(13) / 現状維持 / 削除 のいずれかに分類され、未分類 0 件。`削除` は 0 件。移行時のfile分割・統合・composition追加により、現行filesystemの物理file数との一致は要求しない。
- target package 間の許可依存グラフは 13 ノード・45 エッジで **acyclic**（機械検証 = §5）。
- 実測で見つかった生の双方向結合（cycle）は 10 組（当初7組として起票し、round1 reviewで`LstepTagSyncService`経由の3組を追加）。BE9-2A時点ではreservation↔staffだけ設計裁定が必要だったが、2026-07-20に案A（staffをwrite owner）を採用し、reservation Phase 0で実装済み。現在は全10組の解消方式が確定している（§5、§7.1）。
- §9（旧 BE8-2、service のみの go/ast 実測）との差分は §6 に記録。最大の乖離は medicalrecord ドメイン（旧見積 96 file → 再実測 185 file）。

### 0.1 移行後live-tree recensus（2026-07-24）

classification manifest 761 rowはBE9-2A開始時のsource-path provenanceを保持するimmutable snapshotであり、移行後fileを追記・置換する台帳ではない。現行treeとの照合結果は次の通り。

- target bucket 601 row: 旧path現存 **0** / removed or moved **601**。
- keep bucket 160 row: present **136** / consolidation・責務移動・削除により旧pathなし **24**。
- 旧layer（2026-07-24 snapshot）: `internal/handler` directoryなし、`internal/service` production 0 / test-only 14、`internal/repository` production 0 / test-only 50。**その後3旧layer directoryは完全削除済み**（test residual も含め現行 tree に残存しない）。live lint は `backend/internal/lintscan/`。
- production Go import: 旧`internal/handler|service|repository`へのedge 0。
- composition: `cmd/api` production 22 file、target domainを直接importするfile 18。

| target package | 現行production Go file |
|---|---:|
| medicalrecord | 180 |
| lstep | 131 |
| reservation | 77 |
| billing | 71 |
| staff | 31 |
| auth | 26 |
| trimming | 27 |
| clinic | 24 |
| pet | 22 |
| owner | 17 |
| inventory | 12 |
| httpapi | 10 |
| identitylink | 7 |
| manualarticle | 6 |

> 2026-07-30: `identitylink` を #239 Phase 1 として target recensus に追加（14 target package）。production import は `cmd/api` composition root のみ。package 内 production 依存は `apperrors` / `audit` / `httpapi` / `model` / `persistence` / `textsearch`。

file分割、typed adapter、domain-owned composition/test helperの追加により、現行物理file数はsnapshotのtarget bucket行数と一致しない。移行完了判定は「旧pathの消滅」「現行target packageの存在」「production legacy import 0」「runtime/AST gate」で行い、行数一致だけで判定しない。

2026-07-24のfollow-up hardeningでは、LINE webhookのcross-clinic readを受信前identity解決に必要なchannel-secret走査だけへ縮小し、一意に署名一致したclinicへowner lookup/updateをscopeした。更新はexpected LINE user IDとLINE event timestampを含むCASとし、stale・duplicate・out-of-order・再連携前IDを安全なno-op、同時刻をunfollow優先とする。公開LIFF account linkはowner PIIを返さない`204 No Content`、LINE ID token検証はredirect追従禁止とした。billing confirmation/returnは`Content-Type: application/json`（charset parameter可）を必須とし、不一致を415、bodyを8 KiBのexact-key/string strict single-object JSON、trim後non-blankの`return_reason` 500文字、`memo` 1,000文字として境界で強制する。scheduler opsはCloudflare Access JWKSをWorker isolate内で10分cacheし、同時取得を集約、unknown `kid`/upstream failure後のrefreshを60秒cooldownしてfail closedにする。これらは現行実装の安全境界であり、fresh DB migration、production deploy、Access policy/edge rate limit、alert/recovery rehearsalのrelease gateを完了扱いにしない。

## 1. 分類マニフェスト概要

| bucket | files |
|---|---:|
| target:medicalrecord | 175 |
| keep（現状維持・cross-cutting） | 160 |
| target:lstep | 119 |
| target:reservation | 78 |
| target:billing | 65 |
| target:staff | 30 |
| target:auth | 25 |
| target:clinic | 25 |
| target:trimming | 23 |
| target:pet | 18 |
| target:owner | 13 |
| target:httpapi | 12 |
| target:inventory | 12 |
| target:manualarticle | 6 |
| **削除** | **0** |
| **合計** | **761** |

検証: `awk 'END { print NR - 1 }' docs/architecture/be9-2a-classification-manifest.csv` = 761。bucket合計=761、`unclassified\|未分類\|未定\|TBD`の分類値は0件。`target:reservation=78` / `target:staff=30`は2026-07-23の`liff_service_availability_staff.go`再分類を反映済み。現行filesystemの物理file数は移行後fileを含むため、このmanifest行数の検証には使用しない。

### 1.1 分類手法

1. **repository の 42 既存ドメインsubpackage**（BE8-4 成果物）は roll-up マッピングでそのまま target domain へ帰属（例: `repository/checkup` → medicalrecord、`repository/account` → auth）。
2. **handler/service/repository-root の flat file** は、①HTTPルート登録構造（`internal/handler/handler.go` の `Register*Routes`/`master_routes.go` の RBAC `model.Resource*` 定数）で実際のドメイン帰属を確定し、②ファイル名トークンの prefix ルール（最長一致優先）で機械的に割り当てた。旧 §9 のような filename-prefix-only ではなく、RBAC resource 名とroute registration構造を一次証拠として使っている点が再実測の実質的な改善。
3. **model/ package（85 file）は BE-refactor.md の明示方針により一括 `現状維持`** — 個々の型がどのドメインに概念的に属するかは §4.10 で別途タグ付けした（ファイル移動はしない）。
4. **cross-cutting 11 パッケージ**（config/dbconn/middleware/infra/timeutil/seedbundle/logger/csvimport/authjwt/apperrors/apicontract）と **facade root file**（handler.go/service.go/repositories.go/helpers.go/db.go/base.go/transactor.go）は `現状維持`。
5. 8並列エージェント調査で判明した誤分類は §2 の通り是正済み（本マニフェストは是正後の最終版）。

## 2. マニフェスト是正ログ（8並列エージェント調査で発見・複数エージェント収束の高信頼度のみ適用）

| ファイル | 誤分類→是正後 | 根拠 |
|---|---|---|
| `internal/handler/reservation_line_routes.go` | lstep → **reservation** | 16ルート中14がreservation admin/staff/type-liffハンドラ呼び出し、lstep呼び出しは2ルートのみ（line-customers）。bm-reservation-trimmingとbm-lstepが独立に同一結論（強い収束証拠） |
| `internal/handler/liff_{handler,request,response,validation}.go`、`internal/service/liff_service*.go`(9 file、`liff_service_availability_staff.go`除く) | lstep → **reservation** | `LiffService`構造体はreservation系repoのみで構成、Lstep*/LineLink*/LineMessaging*参照ゼロ。§9のliff out-dom=5/80もほぼreservationロジックだったと再確認。bm-reservation-trimmingとbm-lstep独立収束 |
| `internal/service/owner_service_line.go` | lstep → **owner** | 全メソッドのreceiverが`*ownerService`（ownerServiceの複数ファイル拡張）。lstepパッケージに置くとコンパイル不能 |
| `internal/service/lstep_health_tag_sync_{checkup,vaccine}.go`、`lstep_tag_sync_care{,_checkup,_chronic,_resync}.go`（6 file） | medicalrecord → **lstep** | 全メソッドのreceiverが`*lstepTagSyncService`（lstep_tag_sync_service.goで定義、target:lstep）。filename token（checkup/vaccine/care）ではなくreceiver型がGoのpackage境界を決める。bm-medrecord-aとbm-medrecord-bが独立に同一結論 |
| `internal/service/lstep_tag_sync_pet{,_basic,_exclusion}.go`（3 file） | pet → **lstep** | 全メソッドのreceiverが`*lstepTagSyncService`。L③aで分類訂正し、pet 21→18、lstep 107→110 |
| `internal/{handler,service,repository}/checkup_sync*`（9 production file） | medicalrecord → **lstep** | 健診データを読むが、application ownershipはLSTEPタグ同期。論点#7の案(a)を採用し、L③bで分類訂正してlstep 110→119、medicalrecord 184→175 |
| `internal/repository/accounting_repository_reports_daily.go` | medicalrecord → **billing** | receiverが`*accountingRepository`、`billings`/`payment_splits`テーブル操作（BUG-368日次集計） |
| `internal/repository/account_repository.go`、`internal/repository/account/repository.go`、`internal/service/account_service.go` | billing → **auth** | 機械分類ルールの`'account'`トークンが`'accounting'`と誤って同一視されていたバグ。実体は100%`model.Account`(ログイン認証情報)、billingとの関連ゼロ。`repository/account/repository.go`のdocコメントが"clinic_idを持たないグローバルなログインアカウント"と明記 |
| `internal/service/validators_contact.go` | keep(共有カーネル) → **owner** | `validateEmailFormat`/`validatePhoneFormat`/`validatePostalCodeFormat`の呼び出し元は`validators_owner.go`のみ（1 file/1 domain）。共有カーネルの2条件（複数ドメイン消費 or 下位層からの消費）どちらも満たさない |

上記は`be9-2a-classification-manifest.csv`へ反映済み（reasonカラムに"RE-CLASSIFIED"と根拠を記載）。

## 3. per-domain boundary map（9列）

凡例: **owned** = owned types/routes/queries, **deps** = 許可依存（このドメインが依存してよい他ドメイン）, **tx** = transaction使用, **tenant** = tenant boundary分類。

この節のcall-site、変更頻度、物理pathはBE9-2A時点のinventory snapshotであり、見出しの数値はclassification manifestのsource-row数である。分類訂正は§2、設計裁定はADR-006を正本とし、移行は2026-07-24に完了済み、完遂経緯はgit履歴を優先する。

### 3.1 owner (`target:owner` 13 source rows)

- **owned**: `model.Owner`(ClinicID not null)。Route: `/owners`(CRUD+line-user-id/line/lstep-opt-out/delivery-exclusion/delivery-caution/transfer-status/line-id-confirm/line-link-token)、alias `/clinics/:clinic_id/owners`。Repo: `FindAll/FindByID/FindByEmail/FindByPhone/FindByLineUserID/CreateWithPets/Update`、`ltv_repository.go`(owner+medical_records+billings/paymentsの読み取り専用LTV集計、C-10)。
- **consumers**: lstep(liff/line_*/lstep_analytics/csv_import/delivery_trigger/health_tag_sync/lifecycle)、medicalrecord、billing(discount rate解決)、pet。
- **deps**: clinic, httpapi。owner→billing(insuranceRepo)は owner 側が narrow interface を宣言し逆転（§5参照）。ltv_repositoryの3ドメイン横断raw SQL読み取りは schema-level exception としてADRに明記(Go importエッジではない)。**owner→lstep**: `owner_service.go:168` `tagSyncSvc LstepTagSyncService`(lstep所有interface)をfieldとして保持——**§5 cycle #8**（round1レビューで検出、owner側がTagSyncNotifier相当のinterfaceを宣言し逆転）。
- **change freq(120d)**: 84 commits。
- **fan-in/out**: fan-in 4ドメイン(lstep/medicalrecord/billing/pet)、fan-out 0(interface化後——owner→billing/owner→lstep とも解消済み)。
- **tx**: Yes — `owner_repository.go:176` `CreateWithPets`。
- **route**: Yes — `/owners` + alias。
- **tenant boundary**: clinic-scoped（Owner.ClinicID not null、曖昧性なし）。

### 3.2 pet (`internal/pet`, classification manifest 18 source rows)

- **owned**: `model.Pet`(ClinicID not null)、`model.AnimalSpecies`(**ClinicIDなし・global master**)、`model.PetChronicCondition`(ClinicID not null)。Route: `/pets`(CRUD+treatment-history/first-visit/death)+chronic-condition nested、`/masters/animal-species`(CRUD+reorder)。Repo: `FindAll/FindByID/FindLivingByOwner/CreateWithPets委譲/FindOwnersByPetBirthday`。
- **consumers**: medicalrecord(hospitalization/lab_import/medical_record)、lstep(checkup_sync/delivery_trigger/health_tag_sync/lifecycle/tag_sync)。
- **deps**: owner, clinic, httpapi。pet→owner はFKownership検証のみ、クリーンな一方向。pet↔staff結合は見つからず。**pet→lstep**: `pet_service.go:182`+`chronic_condition_service.go:62`が`LstepTagSyncService`をfieldとして保持——**§5 cycle #9**（round1レビューで検出、pet側がinterfaceを宣言し逆転）。
- **change freq**: 108 commits。
- **fan-in/out**: fan-in 2(medicalrecord,lstep)、fan-out 1(owner、interface化後——pet→lstepも解消済み)。
- **tx**: No（pet_service/chronic_condition/animal_species/pet_repository/pet_chronic_condition_repositoryいずれも`WithTx`/`Transaction(`ゼロ件）。
- **route**: Yes — `/pets` + `/masters/animal-species`。
- **tenant boundary**: **混在** — Pet/PetChronicConditionはclinic-scoped、AnimalSpeciesはglobal master（同一bucket内）。clinic-id-isolation lintがAnimalSpeciesクエリを誤検知しないようADRに明記が必要。

### 3.3 staff (`target:staff` 30 source rows)

- **owned**: `model.Staff`(ClinicID=home clinic)、`model.StaffClinicAssignment`(StaffID+ClinicID join、IsMainフラグ=マルチクリニック配属)、`model.ShiftEntry`/`ShiftTemplate`/`ShiftTemplateBreak`、`model.Occupation`。Route: `/masters/staffs`(CRUD+reorder+permission-groups+clinics+excluded/capable-reservation-types)、`/shifts`、`/shift-templates`、`/masters/occupations`。
- **consumers**: auth(login/session)、reservation(admin/appointment/reservation_service)、trimming。
- **deps**: clinic, httpapi。staff→account/auth(`CreateWithAccount`)は staff 側が narrow interface を宣言し逆転（§5参照）。staff→clinic(`StaffClinicAssignment.ClinicID`)。
- **change freq**: 158 commits（このクラスタで最高）。
- **fan-in/out**: fan-in 3(auth,reservation,trimming)、fan-out 1(clinic、interface化後)。
- **tx**: Yes — `staff_service_account.go`(`s.tx.WithTx` in `CreateWithAccount`)。
- **route**: Yes — `/masters/staffs`,`/shifts`,`/shift-templates`,`/masters/occupations`。
- **tenant boundary**: **cross-clinic-identity** — Staffはhome ClinicIDを持つが、StaffClinicAssignmentが明示的にマルチクリニック配属をモデル化（IsMainフラグ）。ShiftEntry/ShiftTemplate/Occupationは単純clinic-scoped。

### 3.4 auth (`target:auth` 25 source rows)

- **owned**: `model.Account`(**ClinicIDなし・グローバル識別子**)、`model.PermissionGroup`/`PermissionGroupRule`(ClinicID on group)、`StaffPermissionGroup`、`model.TokenBlacklist`(ClinicID/AccountIDなし)、`model.PasswordResetToken`(AccountID FK、ClinicIDなし)。Route: `/login`(rate-limited)、`/logout`、`/auth/refresh`、`/auth/forgot-password`、`/auth/reset-password`、`/masters/permission-groups`。Repo: `PermissionGroupRepository`、`AccountRepository`(§2是正)。
- **consumers**: 全13ドメインの route registration が `h.RequirePermission(...)` を呼ぶ普遍的RBAC fan-in — これは実際のauthビジネスロジック(login/session/password-reset)への依存ではなく、`middleware/auth.go`(既にkeep-tier)と同種のドキュメント化された例外として扱う。DAGの`domain→auth`エッジとしては数えない（§5参照）。
- **deps**: staff, clinic, httpapi。auth→staff(`NewAuthService(account,staff,effectivePermission)`)。auth→clinic(`auth_handler.go`の`h.svc.Clinic.ListClinics`、login/`/me`フロー)。
- **change freq**: 149 commits。
- **fan-in/out**: 業務fan-in 0(RBAC-gating経由を除く)、fan-out 2(staff,clinic)。
- **tx**: No（auth_service/password_reset_service/permission_group_service/token_blacklist_service/token_service/account_serviceいずれもゼロ件）。
- **route**: Yes — `/login`等 + 非route universal middleware export(`RequirePermission`)。
- **tenant boundary**: **混在、cross-clinic-identity優勢** — Account/TokenBlacklist/PasswordResetTokenはclinic非依存(identity/session)。PermissionGroup/PermissionGroupRuleはclinic-scoped(クリニックごとのRBAC定義)。

### 3.5 reservation (`target:reservation` 78 source rows)

- **owned**: `model.Reservation`(table `appointments`)、`ReservationType(Group/UnavailableTime/AvailableSlot/Occupation)`、`LineReservationSetting`、`StaffReservationCapability/Exclusion`。Route: `/reservations`、`/clinics/:clinic_id/{line-reservation-settings,reservation-types,reservation-staffs,reservations,line-customers}`、`/api/liff/:clinicId/*`(公開LIFF予約、§2是正でreservation化)、`/masters/reservation-type(-group)s`(+unavailable-times/available-slots/occupations)。Repo: `AcquireBookingLock`/`LockAndFindByID`(pg_advisory_xact_lock/FOR UPDATE)。
- **consumers**: billing(appointment charges)、medicalrecord(`AutoCreateFromReservation`)、lstep(`lstep_batch_service`のLINEバッチのみ、liff系は§2是正でreservation化済み)、trimming、staff。
- **deps**: owner, pet, staff, clinic, httpapi。owner/pet依存はraw SQL(`AssertOwnerInClinic`/`FindPetOwnerInClinic`)でありGo importエッジではない。**reservation↔trimming**は reservation 側(`reservation_validators.go`)が narrow interface を宣言し逆転(§5)。**reservation↔staffはstaffを唯一のwrite ownerとする案Aで解消済み**（reservation Phase 0、§7.1）。
- **change freq**: 259 commits。
- **fan-in/out**: fan-in 4ドメイン(billing,medicalrecord,staff,trimming)、fan-out 3(owner,pet,staff、+interface化前はtrimming)。
- **tx**: Yes — `AcquireBookingLock`(clinic-scoped `pg_advisory_xact_lock`)+`LockAndFindByID`(`FOR UPDATE`)。
- **route**: Yes — `/reservations`,`/clinics/:clinic_id/reservation-*`,`/api/liff/:clinicId/*`,`/masters/reservation-type(-group)s`。
- **tenant boundary**: clinic-scoped、曖昧性なし。

### 3.6 trimming (`target:trimming` 23 source rows)

**Historical BE9-2E landing状態（2026-07-23、code tip `297a23fc7`）**: 23 source rowのproduction bodyとdomain-owned testを`internal/trimming`へ収束した。当時は旧layerにroute/composition/tygoの実consumerを持つcompatibility surface 13件（handler 5、repository 4、service 4）をBE9-2F期限で残していた。これらは2026-07-24のBE9-2Fでconsumer切替・削除済みであり、現行残量ではない。route tuple/RBAC/OpenAPI、clinic isolation、status/error、public LIFF catalogの配置は維持した。

- **owned**: `model.TrimmingCourse`/`TrimmingOption`/`TrimmingCourseType`、`AppointmentTrimmingDetail/Option`(appointments 1:1拡張)。Route: `/trimmings`、`/masters/trimming-{courses,options,course-types}`、`/api/liff/:clinicId/trimming-{courses,options}`(read-only)。
- **consumers**: billing(TrimmingCourseID所有権チェック)、reservation(結合、後述)、lstep/LIFF(顧客向けカタログ、§2是正後はreservation内のliffファイル経由)。
- **deps**: reservation, pet, clinic, httpapi。`appointments`のwrite ownerはreservationであり、trimmingのwriteはconsumer-side interfaceの`CreateForTrimming` / `LockTrimmingByID` / `UpdateForTrimming` / `DeleteForTrimming`へ限定する。同interfaceにはtrimmingが必要とするread/validation/booking-lock capabilityも含むが、generic appointment writerや独立persistenceを持たない。
- **change freq**: 78 commits。
- **fan-in/out**: fan-in 2(billing,reservation)、fan-out 1(reservation)。
- **tx**: Yes。trimmingの`WithTx`内でreservation intentを呼び、appointment/detail/optionsを同時commit/rollbackする。新規Createは重複予約を禁止する条件でclinic booking advisory lockを取り、作成対象には既存appointment rowがない。既存appointmentへのattach/update/deleteはrow lockを使い、reservation type/course/option/staff assignment/capabilityの必要な`FOR SHARE`を同じtransactionへ参加させる。reservation intentはambient transaction欠落をfail-closedにする。
- **route**: Yes — `/trimmings`,`/masters/trimming-*`,`/api/liff/:clinicId/trimming-*`(read-only)。
- **tenant boundary**: clinic-scoped、曖昧性なし。

### 3.7 medicalrecord (`target:medicalrecord` 175 source rows — 最大domain、A(診断/検査/処方/lab)+B(治療/入院/看護)の2サブクラスタが同一Goパッケージ内で相互参照するため統合)

- **owned**: 主要model: `MedicalRecord`,`MedicalRecordAddendum`,`MedicalRecordImage`,`Checkup(Field/Type)`,`ChiefComplaintType`,`Consultation`,`DiagnosisType/Name`,`Examination(Type)`,`Inquiry(Template)`,`LabImportJob/Event`,`Prescription`,`Vaccination`,`Vaccine`,`Hospitalization`,`TreatmentPlan`,`DailyRecord`,`CareLog/CarePlanItem`,`HospitalizationPlan`,`Vital`,`Medicine`,`Procedure`,`Cage`,`ClinicalPlan`,`Treatment`。Route: `/medical-records/:id/{checkups,prescriptions,inquiries,addenda,images,vitals,treatments,treatment-plans,clinical-plan}`、`/checkups`(global)、`/examinations`,`/vaccinations`(top-level)、`/hospitalizations`(+discharge-with-billing/daily-records/care-plan-items/treatment-plans)、`/lab-imports`,`/lab-reports`、`/masters/{medicines,vaccines,consultations,examination-types,diagnosis-types,diagnosis-names,checkup-types,chief-complaint-types,inquiry-templates,cages,hospitalization-plans,procedures}`。
- **consumers**: lstep(checkup/vaccine/prescription/chronic tag sync trigger、consumer-interface経由)、billing(DischargeWithBilling)。
- **deps**: owner, pet, staff, reservation, billing, clinic, httpapi。**medicalrecord→billing**(hospitalization discharge がBilling/BillingItem行を作成)が実エッジ。**medicalrecord↔billing の逆方向**(billingItemService.treatmentRepo等がmedicalrecordを読む)は billing 側が narrow interface を宣言し逆転(§5)。**medicalrecord↔lstep**はmedicalrecord側がTagSyncNotifier相当のinterfaceを宣言し逆転——lstepがmedicalrecordのrepoを直接読むのは一方向として許容(lstepはDAG最上位)。lstep checkup-syncリポジトリの4ドメイン横断raw SQL読み取り(owners+pets+medical_records+billings+checkups)はschema-level exceptionとしてADRに明記。
- **change freq**: A=347 commits, B=223 commits（合計570、重複ファイルなし確認済み）。
- **fan-in/out**: `lockDraftMedicalRecord`のfan-in=20呼び出し/8ファイル/3domain(medicalrecord A+B, billing)。**round1レビュー(architect)で判明**: `LstepDeliveryTriggerService`(lstep所有)もcheckup/medical_recordサービスへ注入されておりmedicalrecord↔lstep結合面を広げるが、新規ドメインペアは追加しない(既存cycle #6の範囲内)。
- **tx**: **混在——安全ギャップあり**。checkup/prescription/examination/medical_record_image/treatment/vitalはfinalize-lock(`lockDraftMedicalRecord`row-lock)または相当のtx保護あり。inquiryは別方式(atomic WHERE status='draft')で意図的。**checkupはTOCTOUギャップあり**(FindByID読み取り→非ロックの別呼び出しでUpdate/Delete、`lockDraftMedicalRecord`未使用)。**vaccinationはガードなし**(MedicalRecordID紐付け時)。hospitalization discharge/treatment plan書き込みはtx保護あり。lab importは意図的なsaga pattern(単一tx wrapping ではない、per-row partial success)。
- **route**: Yes（§1参照）。
- **tenant boundary**: 全clinic-scoped(master含む、clinic固有カタログ)。自前ClinicIDを持たずmedical_record_id/hospitalization_id経由の間接scopeで保護される個別resourceとして`Inquiry`,`CarePlanItem`,`Treatment`,`ClinicalPlan`,`MedicalRecordImage`を確認し、subquery形式とcross-tenant testを確認済み。追加検出した`ExamResult`/`ExamTypeField`/`StaffNote`も安全な別patternであることを検証済み（§7.3）。
- **A+Bが1 packageへ統合される根拠の補足（round1レビュー(architect)の指摘を反映）**: 当初「`lockDraftMedicalRecord`という非exportヘルパーをA/B双方が呼ぶため同一package必須」という理由を主軸に記載していたが、この論法は**billingも同じ非exportヘルパーを呼んでいる**（`billing_confirmation_service.go`/`estimate_service.go`）ため、同じ理屈ならbillingもmedicalrecordへ統合すべきという誤った結論を導いてしまう（実際にはbillingは§5 cycle #3のinterface逆転で解決する対象であり、統合対象ではない）。A+B統合の正しい根拠は**呼び出し密度の非対称性**（A/B間はcheckup/treatment/vital等20箇所・8fileが直接同一packageの非exportシンボルを参照し合う濃い相互依存、billingとの結合は2fileのみで一方向に整流可能）と、**業務的凝集**（診断・検査・処方・治療・入院はいずれも「1回の受診/入院という単一の臨床記録」を構成する不可分な要素という業務ドメインの一体性）の2点であり、非exportヘルパー共有はその一証拠に過ぎない——単独の判断根拠にしない。
- **medicalrecord内部のsub-batch案（§8のlstep案と対になる参考情報、round1レビューのsuggestion）**: 185 fileは§9由来のsink-first原則に従えば概ね次の順で安全に抽出できる——①diagnosis/diagnosis_type/examination_type/chief_complaint_type等の純粋master CRUD（out-dom=0）、②checkup/vaccine/prescription/inquiry等の非確定期処理（lstep通知はinterface経由のため抽出を妨げない）、③lab_import/lab_report（sagaパターンで自己完結）、④treatment/vital/clinical_plan（`lockDraftMedicalRecord`row-lock中核、抽出時にA全体が先に必要）、⑤hospitalization/discharge-with-billing（billingとの実エッジを含む最終段）。具体的なfile単位の割り当てはBE9-2C/2D着手時に確定する（本ADRの時点では概略のみ）。

### 3.8 billing (`target:billing` 65 source rows)

- **owned**: `Billing`,`BillingItem`,`BillingConfirmation`(**ClinicIDなし、medical_record_id経由の間接scope**),`BillingRefund`,`Campaign`,`Insurance`,`PaymentMethodMaster`,`CashRegisterClose`,`Account`は含まない(§2是正でauthへ)、`Estimate`。Route: `/accountings`,`/estimates`,`/billing-items`,`/cash-register`,`/payment-methods`,`/masters/{insurances,campaigns}`、`/records/.../billing-confirmations`(medicalrecordのrouteグループ内にmount——ルーティング上の結合)。
- **consumers**: medicalrecord/hospitalization(`DischargeWithBilling`)、owner/pet(InsuranceID検証)、lstep(tag automation向け読み取り)。
- **deps**: owner, reservation, trimming, inventory, clinic, staff, httpapi。billing→medicalrecordは narrow interface 経由(§5、Go importなし)。billing→trimming/reservationは所有権チェックのread-onlyで一方向(逆方向なし、cycleではない)。**billing→lstep**: `accounting_service.go:135` `tagSyncSvc LstepTagSyncService`(lstep所有interface)をfieldとして保持——**§5 cycle #10**（round1レビューで検出、billing側がinterfaceを宣言し逆転）。
- **change freq**: 264 commits。
- **fan-in/out**: `AccountingRepository`6consumer、`InsuranceRepository`3consumer(うち2はowner/pet外部)、`PaymentMethodMasterRepository`4consumer。
- **tx**: Yes、広範(`s.transactor.WithTx`)。refund_serviceは行ロック+tx内再集計で原子性を確保。**ギャップ**: `cashRegisterService.Close`がcheck-then-createでtx/lockなし(同一期間の同時締め処理race)。
- **route**: Yes（§1参照）。
- **tenant boundary**: clinic-scoped(Billing/BillingRefund/Campaign/Insurance/PaymentMethodMaster/CashRegisterClose/Estimate)。`BillingConfirmation`、`BillingItem`、`Payment`は親経由の間接scope。BE9-2Aで検出した`billing_item_repository.go`のUpdate/Delete防御ギャップはBUG-417としてsubquery形式とcross-tenant testで修正済み（`2634f58fe`、§7.4）。2026-07-24 hardeningではconfirmation/return actorを認証済みstaff contextから導出し、`Content-Type: application/json`（charset parameter可）以外を415、bodyを8 KiB上限・exact lowercase key/string値だけのstrict single objectへ限定、case variant・null・非string・unknown/trailing JSONを拒否し、trim後non-blankの`return_reason` 500文字・`memo` 1,000文字を強制した。
- **discount_permission.go所在の確認**: authドメイン分類が正しい(5ハンドラファイル×3ドメインから呼ばれる純粋RBACヘルパー、billing固有ロジックなし)。billingへ移すとowner/medicalrecordハンドラがbillingをimportする逆結合が生じるため現状(auth)を維持。

### 3.9 inventory (`target:inventory` 12 source rows)

- **owned**: `model.Inventory`,`MerchandiseItem`。Route: `/inventory`、`/masters/merchandise-items`。
- **consumers**: medicalrecord/medicine(medicineService.inventoryRepo、**medicine CRUDと同一txでInventory行を書き込む co-transactional cross-domain write**)、billing(campaignService.merchandiseItemRepo、read-only所有権チェック)。
- **deps**: clinic, httpapi。inventoryService/merchandiseItemService自体は他ドメインへ一切依存しない(確認済み)——**近純粋シンクだが、medicine→inventoryのco-transactional書き込みがあるため「書き込み者ゼロの葉」ではない**。transactorのabstractionはsplit後もMedicine+Inventory両repoを同一txスコープで公開し続ける必要がある。
- **change freq**: 83 commits。
- **fan-in/out**: fan-in 2(medicalrecord/medicine、billing/campaign)、fan-out 0。
- **tx**: No own transactor——medicineServiceのWithTx内へゲストとして参加するのみ。
- **route**: Yes — `/inventory`,`/masters/merchandise-items`。
- **tenant boundary**: clinic-scoped(MerchandiseItem.ClinicID not null確認済み、Inventoryはrepo method引数として明示的clinicID受け渡しパターン、モデルファイル直接確認は次フェーズ推奨)。

### 3.10 lstep (`target:lstep` classification manifest 119 source rows)

- **owned**: `LstepSettings`,`LstepTagCache`,`LstepTagCodeMapping`,`LstepTriggerPriority`,`LstepDeliveryTriggerLog`,`LstepSyncErrorCounter`,`LstepCsvImport`,`LstepFriendAttributeSnapshot`,`LstepAutoManagedPrefix`(global-master),`LstepConditionTagMapping`(global-master),`LstepSendPurposeTagPrefix`(global-master)、`LineCustomer`,`LineLinkToken`,`LineSendLog`,`SharedFile`。Route: `/lstep-settings`,`/lstep-tags`,`/lstep/csv-imports`,`/clinics/:clinic_id/lstep/checkup-sync`、`/api/liff/:clinicId/*`は§2是正でreservationへ移動、`POST /api/line/webhook`(JWTなし、HMAC-SHA256署名検証)、`/owners/:id/lstep/*`。
- **consumers**: なし(lstepはDAG最上位、他ドメインから逆参照されない)。**round1レビュー(architect)で判明した訂正**: lstepが所有する`LstepTagSyncService`interfaceは owner(`owner_service.go:168`)/pet(`pet_service.go:182`,`chronic_condition_service.go:62`)/billing(`accounting_service.go:135`)/medicalrecord(cycle #6、複数ファイル)の**4ドメイン**からfieldとして直接保持されている——これは全て生cycleであり(§5 cycle #6,#8,#9,#10)、medicalrecordの1件のみが当初文書化されていた。4ドメインとも「消費側(owner/pet/billing/medicalrecord)がTagSyncNotifier相当のinterfaceを宣言し逆転」という同一方式で解消し、lstep側はGo importを受けない——解消後の実態としては「consumers: なし」で正しいが、**解消前のraw cycle censusが不完全だった**点を明記する。
- **deps（BE9-2A snapshot）**: owner, pet, staff, reservation, medicalrecord, billing, clinic, httpapi。当時の`lstepTagSyncService`は8〜9domain幅のrepository injectionを保持していた。ADR-006はLINE transportを含む単一`internal/lstep`を裁定し、実装ではconsumer-side narrow interfaceへ縮小している。liffはreservationへ統合済み。独立`internal/line`は作らない。
- **change freq**: 264 commits/120日、全internal/コミットの21.4%がこの15.2%のfile群に集中——著しく活発。
- **fan-in/out**: `LstepTagSyncService`のfan-in=9呼び出し元/4+file、**カバリングテストなし**(リスクフラグ)。
- **tx**: Yes(liff由来のbooking作成、CSV import、lifecycle)。配信バッチ(delivery/batch系)は意図的にtxなし・continue-on-errorのbest-effort設計。
- **route**: Yes（§1参照）。
- **tenant boundary**: **混在、3分類**——clinic-scoped多数、global-master(AutoManagedPrefix/ConditionTagMapping/SendPurposeTagPrefix)、**cross-clinic-identity(重要edge case)**: `POST /api/line/webhook`はclinic IDなしで届くため、受信前identity解決に限って全`LineReservationSetting`のchannel secretを読む。これはADR-002の「no unscoped read」に対する限定例外であり、ownerを全clinic横断で検索する権限ではない。署名は一意に一致したclinic IDだけを受理し、異なるclinicで同じsecretが一致する曖昧系はfail closedとする。以後のowner lookupとfollow/unfollow updateはそのclinic IDへscopeし、owner未登録のtyped NotFoundだけをno-op、真のlookup/update errorはnon-2xx retryへ伝播する。更新CASは`clinic_id + owner id + expected line_user_id`を必須とし、正数かつ受信時刻+5分以内のLINE event timestampを保存済みfollow/unfollow時刻と比較する。followは両保存時刻より新しい場合だけ、unfollowは保存followと同時刻以上かつ保存unfollowより新しい場合だけ適用するため、stale・duplicate・out-of-order・再連携前IDは`RowsAffected == 0`の安全なno-op、同時刻はunfollow優先となる。公開LIFF account linkはowner PIIを返さない`204 No Content`、LINE ID token検証のoutbound requestはredirectを追従しない。

**L⑥後の実装境界（2026-07-22・core `849c27524` / final composition `962ce70e3`）**: SharedFileを含むroute/use case/persistence/testとproduction compositionは`internal/lstep`へ収束した。target側のtyped `Application`を`cmd/api`が組み立て、legacy `service.Services` / root `repository.Repositories`のLSTEP ownershipは0。SharedFile route 4本とPOSTのOR権限（owners edit / medical-records create / medical-records edit）、JWT clinic/staff scope、OpenAPI/storage/error contractを保存した。consumer 0のroot facade 16本と旧service adapter 3本は削除し、owner/pet/chronic-condition等の実consumerを持つ期限付きcompatibility surfaceだけをBE9-2E/2Fへ残した。

### 3.11 clinic (`target:clinic` 25 source rows)

- **owned**: `model.Clinic`,`Company`(**シングルトン、id=1固定、global-master——multi-tenant-spanningではない**、model comment/migration comment/コードのid=1 hardcodeで3重確認),`ClinicHoliday`,`ClosingSpecialPeriod`,`ClinicSettings`。Route: `/clinics`,`/company`,`/clinic-holidays`,`/closing-settings`(+special-periods/holidays、ClinicHolidayハンドラへ委譲)。Repo: `CountBlockingReferencesByClinicID`(11ドメインのテーブルへのraw `Table()`削除ガードscan)。
- **consumers**: billing(cash_register/accounting_report、holidayRepo/clinicRepo)、lstep(clinicRepo)、auth(login/`/me`フローの`ListClinics`)。**reservation/trimmingがClinicHoliday/ClosingSettingsを直接消費する証拠は見つからず**、BE9の依存edgeには追加しない。予約可否へ反映すべきかは製品論点としてBE9外で扱う（§7.2）。
- **deps**: httpapi。**clinic↔auth**: clinicService.CreateClinicが`PermissionGroupRepository`をtx内で使用(デフォルト権限グループbootstrap)——clinic側がnarrow interfaceを宣言し逆転(§5)、残るエッジはauth→clinic(3.4参照)のみ。
- **change freq**: 113 commits(主にrepo全体のリファクタスイープ、clinic固有機能変更は少なめ)。
- **fan-in/out**: fan-in 3ドメイン(billing×2,lstep×1)+auth(service経由)、fan-out 1(auth、interface化前)。
- **tx**: Yes — `CreateClinic`(clinic行+デフォルト権限グループ2件を1txで作成、**クロスドメインtx境界**)、`Delete`(nested tx)。
- **route**: Yes（§1参照）。
- **tenant boundary**: clinicパッケージ自体がtenant boundaryの**定義者**(全`clinic_id` FKが`clinics.id`を参照、ADR-002)。ClinicHoliday/ClosingSpecialPeriod/ClinicSettingsは通常のclinic-scoped行。Companyはglobal-masterシングルトン。

### 3.12 manualarticle (`target:manualarticle` 6 source rows)

- **owned**: `ManualArticle`,`ManualArticleVersion`(**ClinicIDなし、明示的にclinic横断共通コンテンツ**)。Route: `/manual/articles`(GET=`ResourceManualEdit view`で全認証staffに公開、PUT/DELETE=edit/delete権限)。
- **consumers**: なし(DI wiring以外)——確認済みleaf。
- **deps**: httpapi のみ。
- **change freq**: 13 commits(低churn、leaf/低risk分類と整合)。
- **fan-in/out**: fan-in 0、fan-out 0。
- **tx**: No。監査ログ書き込みがhandler層でbest-effort・tx外(`slog.WarnContext`のみ、ロールバックなし)——低リスクだが記録。
- **route**: Yes — `/manual/articles`(+`/:category/:slug`,`/versions`)。
- **tenant boundary**: global-master(clinic横断共通、リスクなし)。

### 3.13 httpapi (`target:httpapi` 12 source rows)

汎用Gin response/binding/validation helper。生成的で高fan-inかつビジネスドメイン知識ゼロという条件を満たす: `response.go`(RespondError 92 file)、`bind_errors.go`(parseBindError 81 file)、`context_helpers.go`(extractClinicID 85 file、extractStaffID 20 file)、`query_helpers.go`(parseIDParam 72 file)、`slice_helpers.go`(mapSlice 63 file)、`time_response.go`(localTime 59 file)、`date.go`、`list_query_request.go`、`response_pg.go`。

**BE9-2A時点で検出した混入（後続移行への引き継ぎ記録）**:
- `validation.go`の`checkDoctorClinicAssignment`/`verifyStaffClinicMembership`/`resolveStaffWithClinic`はstaff-domain business logic(staff serviceを直接呼ぶ)——httpapiが持つべきでない後方依存。staff/reservationへ移すべき。
- `validators.go`の`validateTaxType`は`model.TaxType`をimportしbilling専用enumを検証(唯一の消費者は`accounting_request.go`)——billingへ移すべき。`validatePassword`はauth専用ポリシー(唯一消費者`auth_password_handler.go`)——低優先度。
- `master_routes.go`は「multi-domain route aggregator」として既にマニフェストがフラグ済み(BE9-2Bで所有ドメインごとに分割)。

fan-out: 0(依存先を持たない純粋リーフ)。route: N/A(自身はルート登録しない、他ドメインのroute実装を支える)。tx: N/A。tenant boundary: N/A(ビジネスデータを持たない)。

## 4. 現状維持(keep)バケットの内訳

### 4.1 cross-cutting package roster(12パッケージ)

3種の維持理由を区別する(すべて「広く消費されている」で一括りにすると不正確):

**(a) 真の実行時cross-cutting、複数ドメインから直接import**: `apperrors`(387 file、全13ドメイン)、`model`(412 file、全13ドメイン)、`config`(30 file、lstep/reservation/billing/auth/staff)、`middleware`(route-wiring層、実質全route)、`infra`+`crypto`/`httpx`/`line`/`lstep`サブパッケージ(lstep/reservation/medicalrecord/pet)、`timeutil`(5 file、billing/reservation/medicalrecord/httpapi)、`authjwt`(直接消費者2 fileのみだが**下位層の腕で維持**——middlewareがauthjwtに依存するため、authドメインへ畳み込むとmiddleware→auth-domainの逆依存が生じる)。

**(b) cmd/tooling専用、実行時(handler/service/repository)消費者ゼロ**: `dbconn`(cmd/seed-export,migrate,stage-importのみ——**APIサーバー本体は`repository.NewDB`という別経路でDB接続**、dbconnはCLIツール専用のDSNビルダー)、`seedbundle`(cmd/seed-export,migrateのみ)、`logger`(cmd/api,cmd/lstep-migrateのbootstrap専用)、`csvimport`(**単一呼び出し元**`cmd/seed-export/main.go`のみ——demo/seed用でlstep機能の輸入経路とは意図的に分離。「internal/配下は単一呼び出し元禁止」という将来ルールを採用するなら唯一の例外として明示的に許容すべき)。

**(c) lint/contract専用、production消費者ゼロ**: `apicontract`(productionコードは`doc.go`のみ、2つの`_test.go`がgo/astでhandlerソースをdocs/api.yamlと静的照合——0 importerは静的解析パッケージとして正しい姿)。

12パッケージいずれも特定ビジネスドメインへの誤配置は見つからなかった。

### 4.2 service/repository root内の共有カーネルファイル

| file(s) | fan-in | 判定 |
|---|---|---|
| `service/validators.go`+`validators_master.go`+`validators_name.go` | 39 file/9ドメイン | §9のin-dom=37(validators.go単体のAST識別子参照カウント)と同オーダー、precedent確認。最大breadthの共有カーネル |
| `service/update_fields.go` | 7 file/2ドメイン(medicalrecord,reservation) | §9のaudit precedent同型、維持 |
| `service/audit_diff.go`+`audit_service.go`+`replace_audit_tail.go` | 25 file/4ドメイン(medicalrecord,lstep,billing,owner) | §9のin-dom=16(絶対数は下回るが≥2ドメインfan-inを確認、方向性はprecedent通り)。§4.3で詳述 |
| `service/go_safe.go` | 3 file/3ドメイン(auth,medicalrecord,reservation) | 少数だが多ドメイン、維持 |
| `service/smtp_sender.go` | 2 file/2ドメイン(auth,reservation) | 少数だが多ドメイン、維持 |
| `repository/base.go`+`helpers.go`+`transactor.go`(非export) | flat root 44 file | **repository/repohelpers/{scope,tx}.go(export版)との重複判明**——subpackageが親packageの非exportシンボルを呼べないため作られた移行期debt。root完全分割後は非export版が死コードになる、BE9-2F削除対象としてADRに明記 |
| `repository/repohelpers/{scope,tx,junction}.go` | `DBOrTx`単体13 file/7ドメイン | 広範、`repository/audit`サブパッケージも既に使用 |
| `repository/junction_helpers.go` | 2 file/2ドメイン(auth,reservation) | 少数だが多ドメイン、維持 |
| `repository/kana_normalize.go` | 4 file/3ドメイン(owner,pet,medicalrecord) | 維持 |
| `repository/repotest/repotest.go` | 32 `_test.go` file | テスト専用カーネル、期待通り |

### 4.3 audit(安全重要、SAFETY-CRITICAL)

- **(a) tenant-scoped**: Yes。`model.AuditLog.ClinicID *uint64 gorm:"not null"`(`audit_log.go:14`)——全audit_log行がDBレベルでclinic_id必須。
- **(b) fail-closed**: **部分的、意図的にガバナンスされたgap**（偶発ではない）。`AuditService.LogEntryTx`→`AuditRepository.CreateTx`は`repohelpers.DBOrTx`経由で呼び出し元のambient txに参加(呼び出し元rollbackがaudit書き込みも巻き戻す、fail-closed、#211参照、9箇所)。`AuditService.Log`/`LogEntry`→`AuditRepository.Create`は非tx・best-effort(失敗時`slog.ErrorContext`でログのみ、業務書き込みの自動ロールバックなし、9箇所)。ガバナンス層は`audit_tx_inventory_lint_test.go`——臨床結果行(ExamResult/CheckupFieldResult、`gorm.DeletedAt`なし)のhard-deleteを`status`フィールド付きallowlistでインベントリ化し、`pending-migration`件数=0を`TestClinicalResultAuditTxInventory_StatusesAreLive`でCI強制(この特定surfaceについては100%達成済み、全audit書き込みがtx保護されているという主張ではなく、hard-delete surfaceに限定したnarrow-scope保証)。
- **(c) 昇格推奨**: **Yes** — BE9-1のpackage非依存safety lint実装後、`internal/audit`への昇格(repository/auditサブパッケージ+service層の`audit_service.go`/`audit_diff.go`/`replace_audit_tail.go`統合)を推奨。`repository/audit`は既にexported `repohelpers`カーネルに依存する独立subpackageであり、tx境界lintは`go:embed *.go */*.go`で1階層subpackage分割を既に想定済み(`preload_clinic_scope_lint_test.go:380`のコメントで確認)。ADR-002紐付けのtenant/fail-closed保証は単一package境界を持つべきで、§9自身が「②共有カーネル」(3つのテンプレートシンクの直後に抽出)と位置付けている——現状のようにservice/repository両rootに分散したまま放置すべきではない。

### 4.4 model/ package(85 file)のドメインタグ付け(ファイル移動なし、参考情報)

85 fileはBE-refactor.mdの明示方針により**一括`現状維持`**(GORM associationのcycleを実測せずに分割しない)。概念的な帰属先ドメインは以下の通り(曖昧なもの7件は§3の各ドメインエージェント結果で解消済み — 括弧内に解消結果を記載):

- **medicalrecord**: checkup_field/record/type、chief_complaint_type、clinical_plan、consultation、cpm_v1/v2_thresholds、diagnosis、dormant_thresholds、examination_record/type、health_prevention_thresholds、hospitalization(+plan)、inquiry(+template)、lab_import/report、medical_record(+addendum/image)、medicine(+dose_param)、prescription、procedure、treatment、vaccination_record、vaccine、vital、cage(§3.7で確定)
- **billing**: accounting、billing_confirmation、billing_refund、cash_register_close、estimate、insurance(§3.8で確定)、payment_method_master、campaign(§3.8で確定)
- **lstep**: line_customer/link_token/send_log、lstep_*(11 file)、shared_file(§3.10で確定、doc comment "LINE個別送信用"根拠)
- **reservation**: reservation(+type/type_group)、line_reservation_setting（2026-07-20の論点#4でreservation帰属に確定。概念タグのみで、model file自体は`現状維持`）。
- **staff**: occupation、shift_entry_break、staff、staff_reservation_capability/exclusion(reservation-scheduling側からも参照されるが所有はstaffのまま)
- **clinic**: clinic、clinic_holiday/integration/settings、closing_special_period、company
- **trimming**: trimming、trimming_course_type、trimming_master
- **auth**: account(§2是正で確定)、password_reset_token、permission(+group)、token_blacklist
- **pet**: animal_species、pet、pet_chronic_condition
- **owner**: owner
- **inventory**: inventory、merchandise_item
- **manualarticle**: manual_article
- **shared/audit kernel**: audit_log(最大消費者はmedicalrecordだが、構造的には§4.3のauditカーネルに属する)

### 4.5 composition-root facadeのfan-out

| facade | struct | fan-out | 備考 |
|---|---|---:|---|
| `internal/service/service.go` | `Services` | **108 field** | ドメインserviceごとに1 field、コードベース最大のfan-out。BE9-2Fで全108行を解体する必要 |
| `internal/repository/repositories.go` | `Repositories` | **95 field** | 同型、1階層下 |
| `internal/handler/handler.go` | `Handler` | struct自体は5 field(薄い)。fan-outは別経路: **95 file**が共有`(h *Handler)` receiverでメソッド定義、`RegisterRoutes`が**44**の`Register*Routes(...)`を呼ぶ | BE9-2Fは44のroute登録呼び出しの解体 **と** 95-fileの共有receiverパターンの解体(各ドメインのhandlerメソッドをドメイン所有のhandler structへ移す)の両方が必要——単純な移動ではない |

## 5. 許可依存グラフ（DAG）と生cycleの解消方式

### 5.0 acyclicity機械検証（round2 santa dual-reviewで要求されたコマンド+出力の明示）

DESIGNされた許可依存グラフ（13 target package、下記10cycleの解消後に残るedgeのみ）に対し、使い捨てPythonスクリプト（Kahnのtopological sortアルゴリズム、scratchpadのみに配置・repo非コミット——本タスクの制約「使い捨てmeasurement scriptはscratchpadのみ・repo非コミット」に従う）でcycle detectionを実行した。**検証対象はDESIGNされた許可グラフであり、生の現行コード参照グラフではない**——後者は10組のraw cycleを含むため意図的にacyclicではない（advisorの助言通り、acyclicity証明はDESIGNされたグラフに対してのみ行う）。

```
$ python3 be92a_toposort.py
nodes: 13, edges: 45
cycle detection: 0 cycles (graph is a DAG)
topological order (dependency-first):
  1. httpapi        6. pet          11. billing
  2. clinic         7. staff        12. medicalrecord
  3. inventory       8. auth         13. lstep
  4. manualarticle   9. reservation
  5. owner          10. trimming
```

edge定義（`EDGES[X] = X が依存してよい対象のリスト`、Kahn法でnode=13・indegree/outdegreeから位相順序を導出、残余node（cycle参加node）が0件であることを`remaining`配列の空チェックで確認）は、下記§5.1の10cycle解消後に残る一方向edgeと1対1対応する。

### 5.1 生cycleの解消方式一覧

| # | cycle | 解消方式 | 残るエッジ |
|---|---|---|---|
| 1 | staff↔auth（staffがAccount作成、authがStaffServiceを要求） | staffが`AccountCreator`相当のnarrow interfaceを宣言、authの具象`AccountRepository`/`AccountService`が構造的に満たす | auth→staff |
| 2 | owner↔billing（ownerがInsuranceRepositoryを要求、billingがownerRepoを読む） | ownerが`InsuranceExistenceChecker`相当のnarrow interfaceを宣言、billingの具象`InsuranceRepository`が満たす | billing→owner |
| 3 | medicalrecord↔billing（billingItemService/billingConfirmationService/estimateServiceがmedicalrecordを読む、`lockDraftMedicalRecord`をimport） | billingが`MedicalRecordOwnershipChecker`/`DraftLocker`相当のnarrow interfaceを宣言、medicalrecordの具象repo/exported helperが満たす | medicalrecord→billing（hospitalization discharge時にBilling/BillingItem行を作成） |
| 4 | reservation↔trimming（`reservation_validators.go`がTrimmingCourseRepository/TrimmingOptionRepositoryをimport） | reservationが`masterFKOwnershipChecker`相当のnarrow interfaceを宣言、trimmingの具象repoが満たす | trimming→reservation（ロック機構借用含む） |
| 5 | clinic↔auth（clinicServiceがPermissionGroupRepositoryを使用、authがClinic.ListClinicsを呼ぶ） | clinicが`PermissionGroupWriter`相当のnarrow interfaceを宣言、authの具象`PermissionGroupRepository`が満たす | auth→clinic |
| 6 | medicalrecord↔lstep（medicalrecordがlstepのTagSyncを呼ぶ、lstepがmedicalrecordのrepoを直接読む） | medicalrecordが`TagSyncNotifier`相当のnarrow interfaceを宣言、lstepの具象`LstepTagSyncService`が満たす | lstep→medicalrecord（repository injection、8-9ドメイン幅） |
| 7 | **reservation↔staff（BE9-2A時点では共有テーブルへ2つの独立write path）** | 2026-07-20に案Aを採用。staffを唯一のwrite ownerとし、reservationはconsumer-side writerを宣言してstaff具象を注入。reservation Phase 0で実装済み | reservation→staff |
| 8 | owner↔lstep（`owner_service.go:168` `tagSyncSvc LstepTagSyncService`をfield保持、lstepはownerRepoを直接保持） | ownerが`TagSyncNotifier`相当のnarrow interfaceを宣言、lstepの具象`LstepTagSyncService`が満たす | lstep→owner（repository injection） |
| 9 | pet↔lstep（`pet_service.go:182`+`chronic_condition_service.go:62`が`LstepTagSyncService`をfield保持、lstepはpetRepoを直接保持） | petが`TagSyncNotifier`相当のnarrow interfaceを宣言、lstepの具象`LstepTagSyncService`が満たす | lstep→pet（repository injection） |
| 10 | billing↔lstep（`accounting_service.go:135` `tagSyncSvc LstepTagSyncService`をfield保持、lstepはbillingItemRepo/accountRepoを直接保持） | billingが`TagSyncNotifier`相当のnarrow interfaceを宣言、lstepの具象`LstepTagSyncService`が満たす | lstep→billing（repository injection） |

**round1 santa dual-review（architect、2026-07-19）で追加検出**: 当初の生cycle censusは`LstepTagSyncService`（lstep所有interface）がmedicalrecordだけでなくowner/pet/billingの計4ドメインからfieldとして直接保持されていることを見落としていた（#8-10として追加）。3件とも#6(medicalrecord↔lstep)と**同一の解消方式**（消費側がnotifier interfaceを宣言）で解消可能であり、13 package taxonomy・トポロジカル順序（§後述）には影響しない——census網羅性の是正であり、DAG構造自体の誤りではない。Goはimport cycleを物理的に拒否するため、この見落としが実害化する経路は「コンパイルエラー」であり「サイレントなtenant isolation不具合」ではない。

10組中9組（#1-6,#8-10）はconsumer-side interfaceで解消可能。#7はpublic contract変更を要したため案Aを裁定し、reservation Phase 0でstaff write ownerへ収束済み。したがって、現時点で解消方式未確定のcycleは0件。

### 5.2 許可依存の機械ガード（ARCH-A6）

文書上の許可グラフと実装のドリフトを防ぐため、**production domain→domain Go import** を CI / scoped test で固定する。

| 項目 | 正本 |
|------|------|
| 許可 edge 表 | `backend/internal/lintscan/domain_import_allowlist_lint_test.go` の `domainImportAllowlist` |
| 機械チェック | 同ファイル `TestDomainImportAllowlistLint`（`go/parser` ImportsOnly、production `.go` のみ） |
| 対象外 | `model` / `apperrors` / `audit` / `persistence` / `sharedkernel` / `middleware` 等の cross-cutting import（domain 間 edge ではない） |
| 同一 domain 内 | top-level が同じ `internal/<domain>` への import は gate 対象外 |

**新規 domain edge を足すとき（必須）**

1. 本 boundary map §3（deps）と §5（残るエッジ / cycle 解消）を同じ PR で更新する。必要なら ADR-006 も。
2. `domainImportAllowlist` に importer → dep を追加する（allowlist 自体が acyclic であること — 同 test が検証）。
3. 逆方向が「consumer-side interface で解消済み」の edge を **具象 import で復活させない**（A6-5）。例: `medicalrecord`↛`lstep`、`billing`↛`medicalrecord`、`reservation`↛`trimming`、`owner`↛`billing`/`lstep`、`clinic`↛`auth`。
4. レビューチェック: 許可外 import が無いこと + silent 部分成功や write-owner 迂回が無いこと（[cross-domain-orchestration-catalog](cross-domain-orchestration-catalog.md) も更新対象なら同 PR）。

**四半期境界監査（A6-4）**

1. `code-review-graph`（または同等）で domain community / bridge edge を一覧する。
2. 観測された production import を `domainImportAllowlist` と突合し、差分は「allowlist 更新 + 文書」か「具象 import の interface 化」で閉じる。
3. 結果を短いメモ（PR または STATUS 1 行）に残す。フル `go test ./...` は不要 — `docker compose exec backend go test ./internal/lintscan/ -run DomainImportAllowlist -count=1` で足りる。

## 6. §9（旧BE8-2、serviceのみのgo/ast実測）との差分

| domain | BE-refactor.md旧見積(filename-prefixのみ) | classification manifest source rows（RBAC resource + route構造 + 8並列エージェント検証） | 差分の理由 |
|---|---:|---:|---|
| LSTEP/LINE | 106 files | **119 source rows**（初回106、後続の分類訂正を反映） | 旧見積はfilename-prefixのみ。再実測後もtag-sync receiver誤分類とcheckup_sync帰属を実装時に訂正し、現行manifestでは119に確定 |
| medical record/clinical | 96 files | **175 source rows**（初回185、後続の分類訂正を反映） | 旧見積は文字通り"medical_record"-prefixファイルのみをカウント。checkup_syncをlstepへ帰属訂正した後も最大ドメインである点は不変 |
| reservation/appointment/trimming | 79 files(合算) | reservation 77 + trimming 23 = **100 source rows**(trimmingを独立target domainへ分離) | trimmingを独立package化する決定 + liff/reservation_line_routes 13 sourceの吸収 |
| billing/accounting | 51 files | **65 source rows** | insurance/campaign/estimate/refund等のroll-inを実測で確認(account_*の3 sourceはむしろauthへ除外、正味+14) |
| owner/pet/staff/vital | 34 files(合算) | owner 13 + pet 18 + staff 31 = **62 source rows**(vitalはmedicalrecordへ確定、合算対象外) | 旧見積のリテラルprefixカウントに対し、RBAC resource名・route nesting構造による検証で大幅増。petのLSTEP tag-sync receiver 3 sourceは§2でlstepへ訂正済み |

**根本原因**: 旧見積表(BE-refactor.md L60-67)はfilename-prefixの単純カウントで、RBAC resource定数(`model.Resource*`)やroute registration構造(`handler.go`のnested route/master_routes.goの権限グループ分け)との突合を行っていなかった。本再実測はこれらを一次証拠として使用し、加えて8並列エージェントによる独立検証(特にreceiver型によるGoパッケージ境界の確認)で機械的prefix-matchingの限界(lstep_tag_sync_care*.goのようなreceiver-vs-filename不一致)を補正した。

## 7. 設計論点の解決・検証記録

<a id="be9-2a-reservation-staff"></a>

### 7.1 reservation↔staff 共有テーブル書き込み（Resolved: 案A）

**BE9-2A時点の問題**: `internal/repository/reservation_staff_repository.go`(target:reservation)が`model.Staff`(`staffs`テーブル)へ直接Create/Update/Delete/UpdateSortOrderを実行——ドキュメントコメント自身が「予約用ラッパー」と認めている。`internal/repository/staff_repository.go`(target:staff)も同一テーブルへ独立にフルCRUDを実装。同型の問題が`model.ShiftEntry`(`shift_entries`テーブル)にも存在——`reservation_schedule_repository.go`(reservation)と`repository/shiftentry/repository.go`(staff)が独立実装。

これは「どちらの方向にimportするか」では解決しない——**同一テーブルへの2つの独立書き込み経路**というデータ所有権の設計問題であり、tenant boundaryやpublic contractの変更が必要な真の依存cycle。

**検討した選択肢**:
- **(A)** staffを唯一の書き込み者とし、reservation固有カラム(`staff_type`,`reservation_visible`,`reservation_comment`,`sort_order`)の更新もstaff-domain-exportedメソッド経由に統一する（public contract変更: reservationはstaffの新規メソッドを呼ぶ）。
- **(B)** reservation固有カラムをstaffの主テーブルから分離し、reservation所有の別テーブル(例: `reservation_staff_settings`)へ切り出す（データモデル変更）。

**解決（2026-07-20）**: 案Aを採用し、staffを`staffs`/`shift_entries`の唯一のwrite ownerとした。reservation Phase 0（`3dc35694e`）でreservation側をconsumer-side writerへ変換し、staff具象を注入した。clinic scope、cross-tenant拒否、transaction atomicityの回帰testを追加済み。案Bのschema分割は採用しない。

### 7.2 clinic ↔ reservation/trimming の消費関係が実測で確認できなかった

brief(タスク前提)は「reservation/trimmingがClinicHoliday/ClosingSettingsを直接参照する」ことを想定していたが、BE9-2A調査と2026-07-20の再grepはいずれも直接参照0件だった。存在しない依存edgeは追加しないと裁定済み。休診日・臨時休業が予約可否へどう反映されるべきかはBE9外のproduct論点であり、要件化する場合だけ責任者名付きの別Issueで扱う。

<a id="be9-2a-indirect-isolation"></a>

### 7.3 間接isolation resourceの検証記録（round1 reviewで列挙漏れを検出・確認済み）

§3.7で挙げた5種（Inquiry/CarePlanItem/Treatment/ClinicalPlan/MedicalRecordImage）はいずれもコード上正しく保護されcross-tenant testも存在する。機械的な全数スイープで追加検出した次のresourceも確認を完了しており、生きた漏洩はない：

- `ExamResult`（examination_record.go）、`ExamTypeField`（examination_type.go）、`StaffNote`（hospitalization.go）も同様に自前ClinicIDを持たない——本docの2回目のレビュー往復で確認した。3件とも上記5種とは**異なる、しかし同等に安全な**パターン: `ExamResult`は`ReplaceItemsByExamID`(`examination_repository.go:177-196`)が親`Examination`行の`clinic_id`存在確認をtx内で先に行ってから`exam_id`で一括delete/re-createする（pre-check + tx方式、Inquiryと同型）。`ExamTypeField`は`ExaminationType`master配下の読み取り専用フィールド定義（`examination_service.go`でvalidFieldIDs検証にのみ使用、独自のCRUD routeを持たない）。`StaffNote`は`AddStaffNote`(`daily_record_service.go:228`)による追記専用（`clinicID`+`hospitalizationID`を明示的に受け取り、update/delete methodは存在しない——変更surfaceがないため回帰リスクも無い）。3件とも生きた漏洩なし、確認完了。
- `StaffReservationExclusion`（staff_reservation_exclusion.go）は自前ClinicIDを持たず、BE9-2A時点の`reservation_staff_repository.go`内で実装されていた。これは**新規発見の脆弱性ではなく、既存lintのsite-exceptionとして追跡済みの低severity読み取りギャップ**。書き込みパスは`validateClinicScopedMasterIDs`で保護され、専用の分離testも存在する。§7.1のwrite-owner cycleはその後解消済み。

**検証結果**: `ExamResult`/`ExamTypeField`/`StaffNote`は上記の通り確認済み。`internal/model`の全structをClinicID有無で分類し、global-master/間接isolationのどちらにも該当しないものを機械検出する追加checkは、BE9-1の完了条件ではなく将来のlint強化候補として残す。実装する場合は別taskでscopeとfalse-positive方針を定める。

<a id="be9-2a-bug-417"></a>

### 7.4 billing_item_repository.go のUpdate/Delete防御ギャップ（Resolved: BUG-417）

**BE9-2A時点の問題**: `internal/repository/billing_item_repository.go`のUpdate/Delete実装が`.Joins("JOIN billings ON ...billings.clinic_id=?...")`を`.Updates()`/`.Delete()`へ連結する形式で書かれていた。GORMの`Joins()`はSELECT系callbackのみで消費され、UPDATE/DELETE系callbackへは伝播しないため、repository防御が成立していなかった。

**BE9-2A時点の安全性**: serviceのUpdateItem/DeleteItemが事前にclinic-scoped `FindByID`でgateしていたため、当時も生きた漏洩ではなかった。ただしrepository防御は事実上no-opで、cross-tenant testも0件だったため、将来の別callerがgateを迂回するriskをBUG-417として起票した。

**解決（2026-07-21）**: BUG-417として、Treatment等と同じsubquery形式へ是正し、Update/Deleteのcross-tenant分離testを追加した（`2634f58fe`）。billing B①以降の前提gateを充足済み。

**BE9-2A時点のtier2記録**: `EstimateItem`、`CampaignTargetCategory`/`CampaignTargetItem`、`ShiftEntryBreak`/`ShiftTemplateBreak`、`AppointmentTrimmingOption`の書き込みパスは当時のinventoryでは未検証だった。現行の合否判断は各移行batchのruntime isolation testを優先する（移行は2026-07-24完了・経緯はgit履歴）。

## 8. lstepドメイン内部のBE9-2A時点リスク比較（履歴、DAGノードではない）

BE9-2Aでは、`target:lstep` 119 source rowを機能群ごとに比較して次のsink-first候補を置いた。その後ADR-006でLINE transportを含む**単一`internal/lstep`**を採用したため、この表はpackage分割案ではなく、実装batch順を決めた当時の依存リスク記録として読む。

| 順序 | sub-batch | out-dom | 根拠 |
|---|---|---|---|
| 1(最安全) | settings | 0 | 純粋config CRUD |
| 2 | tag_config | 0 | global-master config、cross-domain読み取りゼロ |
| 3 | csv_analytics | low | 自テーブル+アップロードCSVのみ読み取り |
| 4 | delivery_trigger | low-mid | 外部LSTEP APIと通信(自前client)、settings+tag_cacheのみに依存 |
| 5(最後) | tag_sync_core | 最高(9 repo依存) | owner/pet/medicalrecord/reservation/billingが独立importable packageになってから抽出 |

独立`internal/line`は作らない。liffはreservationへ統合済み。

**最終実装状態（2026-07-24）**: L①`6bae6095d`、L②`2ef112227`、L③a`d333d63ac`、L③b`ba5767e88`+`5fdfa11fa`、L④`62a09f62e`+`860bd5020`、L⑥`849c27524`+`962ce70e3`は完遂。L⑤は`0fd34c7b7`+`f8a4df073`+`4e8fb5b91`でcode landing完遂 / release pending、BE9-2E-0は`de15c7903`で完遂した。現行`internal/lstep`はproduction Go 131 file、manifestの`target:lstep` 119 source rowは旧path実在0件。L⑤のfresh DB migration実適用はrelease gateとして残る。Session A/Bのfrontier計画は全domain移行完了により履歴化し、詳細はgit履歴、release gateの正本は[`q&a.html` OPS-13〜17](../../q&a.html#ops)とする。

## 9. 実測手法の限界（正直な明記）

- 8並列エージェントはcodegraph MCP(利用可能、2814 file/約4万node)を第一手段とし、grep/git logで補強した。codegraph_callers/calleesは共通メソッド名(`FindByID`等)の解決が曖昧になるケースがあり、その場合はgrepベースの近似に切り替えたと各エージェントが明記している——fan-in/out数値は概算であり、正確な呼び出しグラフの完全証明ではない。
- 変更頻度(`git log --since="120 days ago"`)は直近120日のみの近似指標。
- raw-SQL経由の依存(例: lstep checkup-syncリポジトリの4ドメイン横断JOIN、clinicの`CountBlockingReferencesByClinicID`の11ドメインテーブルscan)はGo importグラフに現れないため、§5のDAGエッジとしては数えていない——schema-levelの実質的結合として本docに明記するに留めた。
