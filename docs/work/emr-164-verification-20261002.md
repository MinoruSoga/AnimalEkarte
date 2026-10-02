# EMR-164: スプシ No.28 改善依頼 — 対応済み項目の検証記録

> 検証日: 2026-10-02 / 作業ツリー: worktree (claim ブランチ `claim/EMR-164`)
> 目的: スプシ No.28 で「対応済」と記載された 4 項目と保険分離会計パスを、現在のコードと決定論的テストで固定する。オンライン請求（アニコム/保険会社への送信）は本ユニットの対象外。

## 結論

スプシ No.28 の 4 項目は**すべて現行コードで実装済みであり、テストで固定済み**。保険適用・負担割合・自己負担額の分離計算も既存テスト群で網羅されている。**アニコム等へのオンライン請求は未実装であり、本チケットでは意図的に実装しない**（接続先・様式・認証情報がチケットに存在しないため。別チケットで仕様確定後に着手する）。

## 改善点 → コード証跡 → テスト名 マッピング

### ① 複数医院所属スタッフの拠点横断「閲覧のみ」

要件: 「複数医院に所属していれば横断で飼主情報やカルテが閲覧できる。閲覧のみ。編集は各医院の所属範囲内」

**コード証跡（読み側）**

- `backend/internal/medicalrecord/medical_record_handler.go:27-35` — `ListMedicalRecords` は `httpapi.ResolveListClinicIDsForPermission(c, medical-records, "view")` で一覧スコープを解決
- `backend/internal/medicalrecord/medical_record_handler.go:67-74` — `GetMedicalRecord` は `ResolveAllClinicIDsForPermission` で詳細スコープを解決（所属医院全体 → 権限付与医院に絞込）
- `backend/internal/httpapi/context.go:175-194` — `ResolveListClinicIDs`: `clinic_ids` パラメータが所属集合の部分集合であることを検証。未指定時は従来どおり選択医院のみ（後方互換）
- `backend/internal/httpapi/context.go:202-218` — `ResolveAllClinicIDs`: 詳細画面は全所属医院に拡張
- `backend/internal/httpapi/clinic_permission.go:118-146` — `FilterClinicIDsForPermission`: 所属医院 ∩ 権限付与医院の積で絞込。全滅なら 403
- `frontend/src/features/owners/loaders.ts:104-128` — 飼主一覧は URL `?clinics=1,2` を `clinic_ids` API パラメータへ転送

**コード証跡（書き側 = 閲覧のみの裏付け）**

- `backend/internal/medicalrecord/medical_record_handler.go:101,139,176,202` — Create/Update/UpdateRecommendationReason/Delete はすべて `httpapi.ExtractClinicID`（選択医院の単一値）に束縛。clinic_ids 拡張は行わない
- `backend/internal/medicalrecord/medical_record_owner_pet_clinic_isolation_test.go:183` — `TestMedicalRecordService_Update_RejectsCrossClinicOwnerPetAndMismatch`: 他医院の owner/pet 参照を伴う更新は拒否

**テスト（本ユニットで追加）**

- `medical_record_handler_multi_clinic_view_test.go`
  - `TestListMedicalRecords_MembershipABGrantsAB` / 「clinic_ids=1,2 は両医院をスコープに渡し横断カルテを返す」— 所属 A+B・両医院に view 権限 → 200、両医院のレコードを含む
  - 同 / 「clinic_ids 未指定は従来どおり選択医院のみにスコープする」— 後方互換の固定
  - `TestGetMedicalRecord_MembershipABGrantsAB` / 「選択医院Aのまま所属医院Bのカルテ詳細を閲覧できる」— `GetByIDForClinics` に `[1,2]` が渡り、clinic_id=2 のカルテが 200 で返る

**既存テスト（引用）**

- `medical_record_handler_selected_clinic_b_grant_a_test.go`
  - `TestListMedicalRecords_MembershipABGrantASelectedB` — 権限無し医院 B を単独指定 → 403。混在 `clinic_ids=1,2` は医院 A にフィルタされ 200
  - `TestGetMedicalRecord_MembershipABGrantASelectedB` — 詳細も同様に医院 A のみへ絞込
- `medical_record_repository_test.go`
  - `TestMedicalRecordRepository_FindAll_ClinicIsolation` — スコープ外医院の混入なし・空配列 fail-safe
  - `TestMedicalRecordRepository_FindByIDForClinics` — 「許可リストに所属医院が含まれれば取得できる / 含まれなければ NotFound / 空は fail-safe NotFound」
  - `TestDB_MedicalRecordRepositoryFindAllCorrelatesRelationsToEachParentClinic` — `FindAll(ctx, {A,B})` が両医院のレコードを返し、preload は親医院に相関（他医院 billing 混入なし）
- `medical_record_owner_pet_clinic_isolation_test.go` — `Create/Update` のクロス医院拒否（閲覧のみ契約の書込側）
- `backend/internal/pet/repository_test.go:204`（参照）— `#86 拠点横断: clinic_ids 複数指定で両医院のペットを返す`
- `frontend/src/features/owners/loaders.test.ts:246` — 「URL の page/search/species/include_deceased/clinics を backend にそのまま転送する」

### ② 飼主・ペット検索（よみがな・電話番号・ペット名・カナ揺れ）

要件: 「漢字苗字のみでしかヒットしない」不具合の解消。飼主よみがな・電話番号・ペット名・カナ揺れで検索ヒット。

**コード証跡**

- `backend/internal/medicalrecord/medical_record_repository_list_search.go:32-60` — カルテ一覧 free 検索は4腕 UNION: `record_no` / 飼主 `name`+`name_kana` / ペット `name`+`name_kana` / `inquiries.chief_complaint`。各文字列は `textsearch.NormalizeKana` + `FoldedExpr`（カナ+空白畳込み）で検索
- `backend/internal/pet/` 飼主・ペット検索も同一パターン（飼主名・よみがな・電話・ペット名・カナ・pet_number）

**テスト（owned パス内・既存）**

- `medical_record_kana_search_test.go` `TestMedicalRecordRepository_FindAll_KanaNameSearch` — 飼主/ペット × 保存ひらがな・カタカナ × クエリひらがな・カタカナの 8 通り
- `medical_record_space_search_test.go` `TestMedicalRecordRepository_FindAll_OwnerNameIdeographicSpaceFourWay` — 全角/半角スペース揺れ + 他医院除外
- `medical_record_search_treatment_test.go` `TestMedicalRecordRepository_FindAll_TreatmentSearchExcluded` / 「検索対象はカルテ番号・飼主名/カナ・ペット名/カナ・主訴の4条件」(:230-237) — 保存された `name_kana` 値（`ヨンジョウケン`/`シカイペット`）へのひらがなクエリ命中を含む
- `frontend/src/features/owners/loaders.test.ts:246` — `search` を API へ透過

**既存テスト（参照・owned パス外）**

- `backend/internal/pet/repository_test.go` `TestPetRepository_FindAll_Search`
  - 「existing phone search still matches」(:358-360) — **電話番号**
  - 「existing pet name search still matches」(:361-363) — **ペット名**
  - 「kana-normalized owner name still matches」(:364-374) — 保存 `name_kana`（ひらがな）にカタカナクエリで命中 = **飼主よみがな**
  - 姓名の全角/半角/連続/前後空白バリアント (:302-330)、pet_number 完全一致・部分一致 (:334-341)
- `backend/internal/pet/repository_kana_search_test.go` `TestPetRepository_FindAll_KanaNameSearch` — ペット名・飼主名のカナ 8 通り
- `backend/internal/owner/repository_kana_search_test.go` `TestOwnerRepository_FindAll_KanaSearchSymmetry` — 飼主名のカナ対称性

### ③ カルテヘッダーの年齢・性別・避妊去勢

要件: 「カルテヘッダーに年齢・性別・避妊去勢が表示される」

**コード証跡**

- `frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx:202-230` — `selectedPet.birthDate` / `gender` / `neuteredDate` / `breed` / `weight` / `insuranceName` / `insuranceDetails` / `dangerLevel` / `dangerReason` / `ownerIsDangerous` を `PatientContextHeader` へ伝搬
- `frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx:16-20` — `calcAge` は `calcAgePartsAt(birthDate, new Date())` で計算済み年齢を表示（`{birthDate}生（N歳Mヶ月）`）
- 同 :22-30 — `formatNeuteredStatus`: 雄+避妊去勢日→`去勢済`、雌→`避妊済`、その他→`避妊・去勢済`、日付無し→`—`
- 同 :215-248 — `性別` / `避妊去勢` / `品種` ラベル+値の描画、:249-256 体重、:279-285 `来院 N 回`、:286-301 保険名+補償率

**テスト（本ユニットで追加）**

- `frontend/src/features/medical-records/components/MedicalRecordStickyHeader.test.tsx`
  - 「生年月日から計算した年齢・性別・避妊去勢・品種などをヘッダーに表示する」— `PatientContextHeader` 実描画で `YYYY-MM-DD生（N歳Mヶ月） / 犬`、`性別`=雄、`避妊去勢`=去勢済、`品種`=柴犬、体重・#ペット番号・来院回数・保険名/補償率を検証
  - 「雌の避妊済表示と、避妊去勢日未記録時の「—」推測しない表示を出し分ける」— 性別分岐と未記録時の非推測表示

**既存テスト（引用）**

- `frontend/src/components/shared/PatientContextHeader/PatientContextHeader.test.tsx` — 「birthDate があれば『生（年齢）』形式で表示される」(:38)、「犬の性別・去勢済・品種を表示する」(:47)、「猫の性別・避妊済・品種を表示する」(:66)、「性別不明で避妊去勢日があれば中立な済表現」(:82)、「未設定値を推測しない」(:96)、年齢特性テスト (:114-134)、「visitCount があれば『来院 N 回』」(:183)
- `frontend/src/lib/calc-age.test.ts` — 年齢計算ユーティリティの特性固定

### ④ 危険度: スタッフには色+理由、飼主レポートには非表示

要件: 「危険度を高いにするとこちらから⚠️が見えてしまう」→ スタッフには段階別アイコン（色+形）と理由メモを出し、飼主向けレポートからは危険度を出さない。

**コード証跡（スタッフ側）**

- `frontend/src/components/shared/DangerBadge/DangerBadge.tsx:13-29` — `高`→`OctagonAlert` 赤系 (`C.bgDanger10`/`C.danger`/`C.borderDanger20`)、`中`→`TriangleAlert` 黄系 (`BADGE.yellow`)。色だけでなく形状も分離（色覚多様性対応）
- 同 :107-136 — Popover トリガーはアイコンのみ（平文なし）。クリックで `特記レベル: {高|中}` + 理由（未登録時は `内容未登録`）
- `frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx:217-219` — `ownerIsDangerous` / `petDangerLevel` / `petDangerReason` をヘッダーへ伝搬

**コード証跡（飼主側・非表示の境界）**

- `frontend/src/features/owner-report/api/get-owner-report-pets.ts:7-32` — owner-report の wire 型に `danger_level`/`danger_reason` キー自体が存在しない
- `frontend/src/features/owner-report/lib/owner-report-pet.ts:10,34` — `toPet` は `dangerLevel: undefined` / `dangerReason: undefined` で fail-closed 構築

**テスト（本ユニットで追加）**

- `MedicalRecordStickyHeader.test.tsx` 「ペット特記 高 は赤いアイコンバッジを出し、クリックで理由メモを開く」— トリガーの danger 系 class、`特記レベル: 高`、理由メモ `保定時に噛む` を Popover 内に検証
- 同 「飼主 is_dangerous は名前横のアイコンマークで出し、特記レベル低・未設定はバッジを出さない」— owner バッジ（role=img「特記」）と低レベルの非描画

**既存テスト（引用）**

- `frontend/src/features/owners/components/OwnersListTable.danger.test.tsx`
  - 「特記レベル 高 は赤いアイコンバッジを飼主名列に出し、メモ Popover を開閉できる」(:68)
  - 「特記レベル 中 は黄色アイコンバッジを出し、同じ中立名の Popover を開く」(:90)
  - 「特記レベル 低・未設定の行はバッジを何も出さず、直接文言も出ない」(:105)
  - 「飼主に特記フラグがあれば飼主名リンクの横にアイコンマークを出す」(:118)
- `frontend/src/features/owner-report/routes/OwnerReport.panels.test.tsx` 「レガシー EMR 準拠の飼主・ペット項目…」(:399~) — fixture に `dangerLevel: "中"` がある状態で `queryByText("危険度")` / `queryByText("中")` が非描画と明記 (#229)。`咬傷注意` は危険度ではなくペット `remarks` 由来の飼主向け項目として描画
- `frontend/src/features/owner-report/api/get-owner-report-pets.test.tsx:46-47,91-92` — wire が `danger_level:"high"`/`danger_reason` を返しても transform 後に `dangerLevel`/`dangerReason` プロパティを持たないことを固定（上流が送っても落とす多層防御）
- `frontend/src/components/shared/PatientContextHeader/PatientContextHeader.test.tsx:292-` — 特記マーク describe（ownerIsDangerous / petDangerLevel 高・中・低）

### ⑤ 保険分離会計パス（A2: 明細ごとの保険適用・負担割合・自己負担）

**コード証跡**

- `backend/internal/billing/accounting_complete_tx.go:193` — 明細ごとに `IsInsuranceApplicable` を保持
- 同 :231-241 — `has_insurance=false` で `insurance_amount≠0` は拒否 (:231-233)。`billing_amount = total_amount − insurance_amount − discount_amount` をサーバ側で不変条件から再計算 (:241)
- `backend/internal/billing/billing_item_request.go` / `billing_item_service.go` — `is_insurance_applicable` の request→service→DB パイプ
- `backend/internal/billing/accounting_service_builders.go` — `insurance_name` / `insurance_ratio` / `insurance_amount` を Payment へ構築
- `backend/internal/billing/insurance_service.go` + `insurance_handler.go` — 保険マスタ（名称・負担率 `coverage_rate`）の CRUD

**既存テスト（引用・追加なし）**

- `billing_item_request_test.go` `TestCreateBillingItemRequest_ToServiceInput` (:73, `IsInsuranceApplicable` true の伝搬 assert)、`TestUpdateBillingItemRequest_ToServiceInput` (:138)
- `estimate_response_test.go:93` — estimate response が `IsInsuranceApplicable` を round-trip
- `unbilled_revision_test.go:90` — unbilled revision diff で `IsInsuranceApplicable` 変更を検出
- `accounting_complete_test.go`
  - `TestAccountingService_CompleteAccounting_NormalizesNegativeInsuranceAmount` (:237) — 保険額を正 magnitude へ正規化
  - `TestAccountingService_CompleteAccounting_RejectsInsuranceAmountWithoutInsurance` (:283) — 保険無し会計への非ゼロ保険額を拒否
  - `TestAccountingService_CompleteAccounting_PartialSplitRejected` (:527) / `StaleScreenTotalRejected` (:562) / `EmptySplitsRejectedWhenBillingPositive` (:592) — 支払分割は再計算後の請求額と一致必須（自己負担額の不変条件）
  - `TestAccountingService_CompleteAccounting_DBSuccess_ServerTotals` (:1067) — DB 経路でも server totals
- `accounting_service_test.go` — `Update` 未送信フィールド保持（`InsuranceRatio` 0.5 維持 :1273、`BillingAmount` server 再計算 :1281）、`Update_NormalizesNegativeInsuranceAmount`（`請求額 = 合計 - |保険| - 割引`）、`Update_RejectsInsuranceAmountWithoutInsurance`
- `accounting_service_builders_test.go` `TestBuildPaymentFromInput` — `InsuranceRatio`/`InsuranceAmount`/`InsuranceName`/`BillingAmount` を payment へ伝搬 (:115-118)
- `accounting_service_correction_test.go` — 訂正時も `InsuranceName`（"アニコム"）を保持 (:106)
- `insurance_service_test.go` / `insurance_handler_test.go` — 保険マスタ（"アニコム損保" 等）の list/create

### ⑥ スコープ外: アニコム/保険会社へのオンライン請求

- 本ユニットの out-of-scope。実装は存在せず、今回も追加しない
- 証跡: `backend/internal`・`frontend/src` 全域で `anicom`/`オンライン請求`/`e-claim`/`レセプト`/`claim submit` 系の送信実装・エンドポイントは存在しない。`アニコム` の字面は保険マスタ名称（データ値）としてのみ出現（`insurance_service_test.go`、`category-config.ts`、`PatientContextHeader.stories.tsx` 等）
- 本 diff にも送信系コード・エンドポイント・UI を一切追加していない（`git diff` で確認）

## 検証コマンド

```sh
# backend（Docker・testdb は animalekarte_ekarte-network 上の db コンテナ）
docker run --rm --network animalekarte_ekarte-network \
  --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local \
  -e APP_ENV=test -e DB_HOST=db -e DB_PORT=5432 -e TZ=Asia/Tokyo \
  -e GOFLAGS=-p=4 -e GOMAXPROCS=4 \
  -v <worktree>/backend:/app \
  -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build \
  -w /app animalekarte-backend \
  test ./internal/medicalrecord/... ./internal/billing/...   # イメージ既定 entrypoint が migration/Air のため --entrypoint go を付与 (Makefile:119,141,384 の慣例)

# frontend（Docker）
docker run --rm \
  -v <worktree>/frontend:/app \
  -v ekarte-frontend-node-modules:/app/node_modules \
  -w /app animalekarte-frontend \
  npx vitest run src/features/medical-records src/features/owner-report src/features/owners
```

### 実行結果 (2026-10-02 実測)

| コマンド | 結果 |
|---|---|
| `go test ./internal/medicalrecord/...` | **PASS** — `ok .../internal/medicalrecord 74.083s`（本ユニット追加の `TestListMedicalRecords_MembershipABGrantsAB` / `TestGetMedicalRecord_MembershipABGrantsAB` を含む） |
| `go test ./internal/billing/...` | **PASS** — `ok .../internal/billing 33.032s` |
| `npx vitest run src/features/medical-records src/features/owner-report src/features/owners` | **PASS** — Test Files 127 passed (127) / Tests 1084 passed (1084) |
| `test -f docs/work/emr-164-verification-20261002.md` | **PASS** — 本ファイル |

**共有 test DB 競合について**: backend suite はデフォルトで全 agent・全 package が共有する `ekarte_db_test` を TRUNCATE し合う（`internal/testdb/truncate.go` は他 process の `idle in transaction` backend まで `pg_terminate_backend` する設計）。検証当日は兄弟キャンペーン agent が同一 DB で並走しており、共有 DB 上の初回実行は無関係ドメインで deadlock (40P01) / 接続切断 (57P01) / FK 違反が多発した。そこで `ekarte_db_emr164_test`（`_test` suffix により testdb の専用 DB ガードを通過）を一時作成し、`TEST_DATABASE_URL` でそこへ向けて上記 2 コマンドを実行・全 PASS。実行後に該当 DB は DROP 済み。失敗は EMR-164 の変更起因ではなくインフラ競合であることを、この隔離実行で切り分けた。

## 残リスク・フォローアップ

- **オンライン請求は別チケット待ち**: 接続先・様式・認証の仕様が確定してから実装。現行は保険分離の会計記録のみ
- `internal/pet`・`internal/owner`・`internal/httpapi`・`frontend/src/components/shared` は本ユニットの owned paths 外のため、そこの既存テストは「引用」で証跡とした（本ユニットで変更なし）
