# 一時バグ／障害メモ（ユーザー依頼 2026-09-13）

> **注意**: root `bug.md` は 2026-09-08 に `todo.md` の [製品 FAIL](todo.md#product-bugs) へ統合廃止済み。  
> 本ファイルはユーザー明示依頼により **一旦** 再作成したメモ台帳。製品 FAIL 正本は引き続き `todo.md#product-bugs`。  
> `BUG-LOCAL-HANDOFF-CSV-CONTRACT` は環境／handoff fixture の BLOCKED。  
> `BUG-*` はローカル（STG スナップショット）で切り分け済み。`PO-*` は仕様・データ修復の判断が必要。

更新日: 2026-09-13

修正プラン: [着手順序と依存関係](#bug-fix-order) · [検証と完了条件](#bug-fix-verification)。各項目の方針と詳細手順は下表から確認できる。すべて計画段階であり、修正済みではない。

## 索引

| ID | status | area | severity | 種別 | 修正プラン |
|:---|:---|:---|:---|:---|:---|
| BUG-LOCAL-HANDOFF-CSV-CONTRACT | OPEN | ops / local-db | Medium | handoff preflight BLOCKED | 現行契約でbundleを再生成し、取込前に検証する。[詳細](#plan-bug-local-handoff-csv-contract) |
| BUG-RES-DOCTOR-ID-ZERO | FIXED | reservation | High | **バグ断定**（FK / doctor_id=0） | 作成時の担当未指定をFE・BEで統一し、0をNULLへ正規化済み。[詳細](#plan-bug-res-doctor-id-zero) |
| BUG-RES-DIALOG-A11Y-CONSOLE | OPEN | reservation / a11y | Low | **バグ断定**（DialogContent Description 欠落コンソール警告） | 警告元と説明IDの対応を特定し、説明の参照切れを直す。[詳細](#plan-bug-res-dialog-a11y-console) |
| BUG-RES-STAFF-SELECT-ORPHAN-LABEL | FIXED | reservation / UI | High | **バグ断定**（担当者選択後に表示が消える） | 候補外でも表示名を保持し、確定 orphan は理由表示＋解除/再選択まで送信遮断。[詳細](#plan-bug-res-staff-select-orphan-label) |
| BUG-RES-AVAILABLE-TIMES-404 | FIXED | reservation | Medium | **バグ断定**（LINE設定欠落で院内API 404） | 院内は設定未登録を識別可能な unset（422 + code）にし案内付き手動時刻のみ許可。空枠/障害/LIFF必須は維持。[詳細](#plan-bug-res-available-times-404) |
| BUG-RES-DECEASED-STATUS-BYPASS | OPEN | reservation / pet | High | **バグ断定**（status=deceased なのに予約可） | 死亡判定の契約を確定し、不整合ペットへの新規writeを防ぐ。[詳細](#plan-bug-res-deceased-status-bypass) |
| PO-PET-DECEASED-DATA-BACKFILL | OPEN | data / pet | Medium | **PO確認**（不整合データの修復方針） | 対象・死亡日の根拠・監査・復旧を確定してからデータ修復する。[詳細](#plan-po-pet-deceased-data-backfill) |
| PO-STAFF-BLANK-NAME-LIST | OPEN | staff UX / data | Low | **PO確認**（空氏名の一覧表示方針） | 有効のみの初期表示と空氏名の代替表示をPO判断後に適用する。[詳細](#plan-po-staff-blank-name-list) |
| PO-OCCUPATION-MASTER-EMPTY | OPEN | master / data | Low | **PO確認**（職種マスタ0件の扱い） | 未登録の案内を整え、必須性と医院別初期登録をPO判断する。[詳細](#plan-po-occupation-master-empty) |
| NOTE-STAFF-STARTTIME-RDT | OPEN | staff console | Low | **調査**（startTime TypeError・アプリ外の疑い） | 拡張なしの環境と比較し、stackから原因を特定して修正対象を決める。[詳細](#plan-note-staff-starttime-rdt) |

---

### BUG-LOCAL-HANDOFF-CSV-CONTRACT: `make reset` 後の `_old_db_handoff` 取込が CSV contract digest 不一致で失敗する

- **現象**: ローカルで `make reset` を実行すると、backup → volume 再作成 → migrate + master seed → `/health` までは成功するが、段階5の `_old_db_handoff` 自動取込が preflight で失敗し、臨床 CSV（owners/pets 等）が投入されない。
- **エラー**: `manifest CSV contract digest is invalid`
- **実測（2026-09-13）**:
  - 失敗開始医院: `hakobuneco`（`run=jouto-intake-20260822-01`）
  - 結果として `jouto` / `shikishima` / `hakobuneco` いずれも clinical handoff 未投入
  - reset 後件数: clinics=4, staffs≈37（seed）, **owners=0, pets=0**
  - backup: `.local-db-backups/20260912T143638Z/`
- **原因（切り分け済み）**:
  - CSV 行データの破損というより、**manifest の `csvContractSha256` が現行 importer 契約と不一致**
  - manifest 側: `11cbd62696507efc2f7886046598b67f0f8b5762bdf158e6ef120b206f63b794`（`backend/migrations/seeds/_old_db_handoff/*/manifest.json`）
  - 現行コード期待: `19b2c5c270058b20c1fa816679c0430f236a0a181165ae1ed2257c64c83f6671`（`backend/internal/csvimport/cutover_contract.go` の `cutoverCSVContractSHA256`）
  - 検証箇所: `backend/internal/csvimport/cutover_contract_validate.go`（`validateLocalRehearsalProducerProvenance`）
  - 対象 bundle はいずれも `handoffEligibility: REHEARSAL_ONLY`（古い rehearsal）
- **影響**: ローカルを「seed + handoff で STG 相当の臨床データ」に揃えられない。予約再現などで実飼主データが必要な場合は、handoff 再生成か別経路が必要。
- **やってはいけないこと**: manifest の hash を手で現行値に書き換えて押し込むこと（契約ドリフトを隠蔽する）。
- **次アクション候補**:
  1. 現行 CSV contract に合う `_old_db_handoff` を再生成する
  2. または handoff なし（seed のみ）で日常開発し、必要な飼主・ペットは都度作成する
- **関連**: [LOCAL_DB_RESET.md](docs/ops/deploy/LOCAL_DB_RESET.md) · [OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) · `scripts/import-old-db-handoffs-on-reset.sh`

---

### BUG-RES-DOCTOR-ID-ZERO: 予約作成で担当医未選択なのに `doctor_id=0` が送られ「参照先が存在しません」になる

- **現象**: 院内予約作成（新規予約）で飼主・ペット・予約区分・日時を入れて「予約を登録／確定」すると、トースト **「参照先が存在しません」** で失敗することがある。担当者は「選択してください」（未選択）のまま。
- **ユーザー報告の流れ**:
  1. 当初「本日は医師が出勤していないため予約できません」（出勤医師 0 の日）
  2. スタッフ／シフト作成後も予約できず「参照先が存在しません」
  3. スクショ: 城詰 明子 / 獅王、予約区分「診察」、担当者未選択
- **再現（2026-09-13・ローカル＝STG スナップショット適用後）**:
  - 医院: 城東センター病院（clinic_id=2）
  - 飼主: 城詰　明子（owner_id=10426886）／ペット: 獅王（pet_id=11036310）
  - 予約区分: 診察（reservation_type_id=5）
  - `POST /api/v1/reservations` に **`doctor_id: 0`** を含めると **400** + `参照先が存在しません`
  - 同条件で **`doctor_id` 省略／null** なら **201** 成功
  - 同一 UI でも成功することがあり、**未選択時に 0 が載るか否かで結果が分岐**
- **原因（切り分け済み）**:
  - PostgreSQL `23503` → アプリ共通文言「参照先が存在しません」（`backend/internal/httpapi/response_pg.go`）
  - 破れた制約: **`fk_appointments_doctor_clinic`**  
    `FOREIGN KEY (doctor_id, clinic_id) REFERENCES staffs(id, clinic_id)`
  - INSERT 実測: `doctor_id=0`（staffs に id=0 は存在しない）
  - ログ例: `reservation_repository.go` insert … `violates foreign key constraint "fk_appointments_doctor_clinic" (SQLSTATE 23503)`
  - 担当医未選択は DB 上 **NULL** であるべきだが、リクエスト／永続化経路で **0** になっている
- **追加切り分け（2026-09-13）**:
  - FE: `transformToCreateRequest` の `data.doctor ? Number(data.doctor)` は **`"0"` を truthy** として `doctor_id: 0` を載せうる（`""` なら undefined）
  - BE Create: Update と異なり `DoctorID == 0` の NULL 正規化が無く、そのまま INSERT → FK `fk_appointments_doctor_clinic` で 400
- **影響**: 担当医なし予約が、出勤医師がいる日でも「参照先が存在しません」で失敗しうる。ユーザーには FK 詳細が見えず原因が分かりにくい。
- **修正方針候補**（実装済み・2026-09-13 attempt `att-bug-res-doctor-id-zero-20260913-001`）:
  1. FE: `normalizeCreateDoctorID` — 空/`"0"` は `doctor_id` 省略、正の十進 ID は保持、不正文字列は throw（fail-closed）
  2. BE: 共有 `normalizeCreateDoctorID` を Create / CreateBatch / admin Create で検証・競合・保存に適用（呼出し元 input は非破壊）。正の無効 ID は capability 検証で拒否（nil へ落とさない）
  3. PATCH 契約は変更なし（omit 保持 / 0→NULL）。回帰: `resolveUpdateParams` + `buildReservationUpdate`
- **検証**:
  - `docker compose exec frontend npx vitest run src/features/reservations/api/transforms.test.ts src/features/reservations/hooks/use-reservation-actions.test.ts` → PASS
  - `docker compose exec backend go test ./internal/reservation` → PASS
- **再確認（2026-09-13・UAT Bot・clinic_id=1 八王子デモ執行）**:
  - 証拠: `reports/uat-2026-09-13/staff-res-crud-v4-20260913-023652.json`
  - `POST /api/v1/reservations` に ISO `start_time`/`end_time` + `visit_type=first` + `reservation_type_id=1`（トリミング）で:
    - **`doctor_id: 0` → 400** `参照先が存在しません`（修正前の失敗モード）
    - **`doctor_id` 省略 → 201**（id=1000000018）
    - **`doctor_id: null` → 201**
  - 同日、省略作成の予約は **PATCH notes → 200**、**DELETE → 204**（予約 CRUD 自体は doctor 省略時に成立）
  - 修正後: FE は空/`"0"` で `doctor_id` 非送信。BE Create 系は 0→NULL。サービス直呼びでも FK `fk_appointments_doctor_clinic` を unset doctor で踏まない。

- **関連コード**:
  - FE: `frontend/src/features/reservations/api/transforms.ts`（`normalizeCreateDoctorID` / `transformToCreateRequest`）
  - FE: `frontend/src/features/reservations/hooks/use-reservation-save-actions.ts`
  - BE: `backend/internal/reservation/reservation_service_validate.go`（`normalizeCreateDoctorID`）
  - BE: `backend/internal/reservation/reservation_service.go`（Create / CreateBatch）
  - BE: `backend/internal/reservation/appointment_admin_service.go`（admin Create）
  - BE: `fk_appointments_doctor_clinic` / `response_pg.go` の FK 文言マッピング

---

## バグ断定

### BUG-RES-AVAILABLE-TIMES-404: 院内 `GET /reservations/available-times` が LINE 予約設定欠落で 404 になる

- **現象**: 予約作成ダイアログ操作中、`GET /api/v1/reservations/available-times?reservation_type_id=5&date=YYYY-MM-DD` が **404** `{"error":"not found"}` を返す（コンソール／Network に連続出現）。

- **コンソール再確認（2026-09-13・城東 clinic_id=2・新規予約で「診察」選択）**:
  - ユーザー報告どおり `use-reservation-types.ts:102` から  
    `GET /api/v1/reservations/available-times?reservation_type_id=5&date=2026-09-13` が **404** `{"error":"not found"}`（同一URLが連続）
  - API直叩き（`X-Clinic-ID: 2`）でも同404。`line_reservation_settings` 欠落経路と一致
  - 証拠メモ: `reports/uat-2026-09-13/followup-console-doctor-20260913-024602.json`
  - **UAT見逃し**: スタッフ／予約CRUD sweep時に予約区分を選ばずダイアログを開いただけだったため、このNetwork/コンソール404をその回の interesting に拾えていなかった（台帳自体には既存項目あり）
- **実測（2026-09-13・STG スナップショット local）**:
  - 城東ログイン後に上記 API → **404**
  - `line_reservation_settings` は **全医院 0 件**（clinic_id=2 も含む）
- **原因**:
  - 院内 available-times は `liffService.getAvailableTimes` → `settingRepo.FindByClinicID`（`line_reservation_settings`）に依存
  - 行が無いと GORM not found → **404**
  - 予約区分「診察」自体は存在し、枠なしなら本来は **空配列 200** が妥当（定休日時は `[]` を返す実装あり）。設定欠落が「not found」になるのは院内 UI にとって不適切
- **影響**: 予約作成自体は成功しうるが、空き枠取得が壊れ、エラーノイズ・枠表示欠落を起こす。STG 本体も同テーブル 0 件なら STG でも同症状の可能性
- **修正方針候補**（未実装）:
  1. 院内 available-times を LINE 設定欠落時にデフォルト営業時間／空枠へフォールバック（404 にしない）
  2. または migrate/seed/STG で `line_reservation_settings` を医院ごとに必須作成
- **関連**: `backend/internal/reservation/reservation_handler.go`（`GetReservationAvailableTimes`）· `liff_service_availability.go` · `line_reservation_setting_repository.go`

### BUG-RES-DECEASED-STATUS-BYPASS: `status=deceased` なのに `deceased_at` が null だと予約作成できてしまう

- **現象**: 受付・予約 UI 上は【死亡】と出るペットでも、予約作成 API が **201** になる場合がある。
- **実測**:
  - ペット「織」（pet_id=11036309）: `status=deceased`, **`deceased_at` IS NULL**
  - 同ペットで `POST /reservations` → **201 Created**（担当医なし・出勤ありの日）
  - 受付カードは `petStatus === deceased` で【死亡】表示・操作 disable（`AppointmentCard.tsx`）
- **原因（修正前）**:
  - BE 死亡 write ガード（`sharedkernel.ValidatePetNotDeceased`）は **`DeceasedAt != nil` のみ**を死亡とみなしていた
  - FE／表示は **`status === "deceased"`** を死亡扱い
- **修正（本ユニット）**: 死亡契約を **`status=deceased OR deceased_at != null`** に統一。日時捏造・DB backfill は行わない（backfill は `PO-PET-DECEASED-DATA-BACKFILL`）。
  - そのため **表示上は死亡・書込は生存扱い** の不整合が起きる
- **バグ断定の根拠**: オペレータから見て死亡と分かる個体に新規予約が通るのは、画面の死亡表示と write ガードの契約不一致（製品 FAIL）
- **関連**: `backend/internal/sharedkernel/pet_not_deceased.go` · `appointment_admin_service.go` · `frontend/src/features/reception/components/AppointmentCard.tsx` · `frontend/src/lib/transforms/pet.ts`
- **データ修復は別項 `PO-PET-DECEASED-DATA-BACKFILL`**

---


### BUG-RES-DIALOG-A11Y-CONSOLE: 新規予約ダイアログ表示時に DialogContent の Description 欠落コンソール警告

- **現象**: `/reservations?newReservation=1` で新規予約モーダルを開くと、ブラウザコンソールに Radix 警告が複数回出る。
  - `Warning: Missing \`Description\` or \`aria-describedby={undefined}\` for {DialogContent}.`
- **実測（2026-09-13）**:
  - 証拠: `reports/uat-2026-09-13/staff-res-crud-v4-20260913-023652.json`（`res_console` / `interesting_console`）
  - スタッフマスタ `/settings/staff` の同 sweep では **コンソール興味イベントなし**（スタッフ CRUD API/UI は PASS）
- **切り分けメモ**:
  - `ReservationFormModal` 本体は `aria-describedby={RESERVATION_FORM_DESCRIPTION_ID}` と sr-only `DialogDescription` を持つ実装がある
  - それでも警告が出るため、**入れ子の別 DialogContent**（例: `components/ui/command.tsx` の Command ダイアログは Description なし）や、Description マウント前の警告の可能性
- **影響**: 開発者コンソール汚染。a11y（スクリーンリーダー向け説明）欠落の兆候。予約 CRUD 自体の機能 FAIL ではない。
- **修正方針候補**:
  1. 警告元の DialogContent を特定し `DialogDescription`（sr-only 可）または `aria-describedby` を付与
  2. Command ダイアログ等の共有 UI も同様に揃える



### BUG-RES-STAFF-SELECT-ORPHAN-LABEL: 担当者を選んでもトリガーに名前が出ない（選択値が候補から外れるとプレースホルダに戻る）

- **現象**: 新規予約モーダルで担当者を選んでも、トリガーが「選択してください」のまま／選んだ直後や予約区分変更後に名前が表示されない。
- **ユーザー指摘（2026-09-13）**: 「担当者を選択しても表示されない。シンプルバグ。原因特定から」
- **原因（切り分け済み）**:
  1. 担当候補は `ReservationFormFields` で **出勤**（`/v1/shifts/on-duty-staffs`）∩ **対応可能コース**（`/v1/clinics/:id/reservation-staffs` の `capable_courses`）で絞り込む（`filterStaffCandidatesByCapability`。metadata 欠落は fail-closed で空）
  2. 城東・本日実測: 出勤3名（鈴木/三井/菊島）。診察(id=5)対応可は **鈴木・菊島のみ**。**三井は `capable_courses: []`** のため区分「診察」選択後は候補から消える
  3. 予約区分変更時 `onSelect` は `type` だけ更新し **`doctor` をクリアしない**（`ReservationTypeAndStaffFields.tsx`）
  4. `SearchableSelect` の表示ラベルは **現在の `options` から `value` 一致で解決**するだけ（`selectedLabel = options.find(...)?.label ?? ""`）。候補に無い value だと **空文字 → プレースホルダ「選択してください」** に見える。値自体は `formData.doctor` に残る
- **典型再現**:
  1. 城東・本日、新規予約を開く
  2. 先に担当「三井隆行」を選ぶ（区分未選択時は出勤者として出る）→ 名前表示
  3. 予約区分「診察」を選ぶ → 三井が capability 外で options から脱落 → **トリガーが未選択に見える**
- **API根拠（2026-09-13・`X-Clinic-ID: 2`）**:
  - `GET /api/v1/shifts/on-duty-staffs?date=2026-09-13` → 20000007, 40000037, 40000038
  - `GET /api/v1/clinics/2/reservation-staffs` → 20000007/40000038 のみ capable_courses に 5。40000037 は空配列
- **関連コード**:
  - `frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts`
  - `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx`（staffOptions）
  - `frontend/src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx`（type 変更で doctor 未クリア）
  - `frontend/src/components/ui/searchable-select.tsx`（options 外 value でラベル空）
- **修正方針候補**:
  1. 区分変更時、選択中 doctor が新 options に無ければ `doctor: ""` にクリア（または警告）
  2. SearchableSelect は options 外でも直前ラベル／フォールバック名を表示
  3. データ: 出勤医師の `capable_courses` を診察等に正しく紐付ける（運用／マスタ）


## PO確認

### PO-PET-DECEASED-DATA-BACKFILL: `status=deceased` かつ `deceased_at` null のペット大量不整合

- **実測**: pets（未削除）のうち  
  - `status=deceased` かつ `deceased_at` null: **5836**（城東だけでも **4699**）  
  - `status=deceased` かつ `deceased_at` あり: 6075  
  - `deceased_at` ありで status≠deceased: 0
- **問い（PO）**:
  1. 旧DB／handoff 由来の不整合を **`deceased_at` を埋めて死亡確定**するのか、**status を alive に戻す**のか
  2. 修復は STG のみか、本番 cutover 前の必須か
  3. `BUG-RES-DECEASED-STATUS-BYPASS` の恒久対策（status もガードに含めるか、SoT を deceased_at のみに揃えて FE も追従か）のどちらを採るか
- **注意**: credential／PHI を台帳に載せない。個体例は検証用 ID のみ

### PO-STAFF-BLANK-NAME-LIST: スタッフ一覧が空欄だらけに見える

- **実測（城東）**: 氏名空かつ無効のスタッフ **198**（種別はほぼ doctor）。一覧 API は `sort_order, name` のため空文字が先頭に来る → 「424件あるが表が真っ白」に見える
- **これはコード欠陥というより STG／移行データ＋ソート／フィルタの UX**
- **問い（PO）**:
  1. デフォルトで「有効のみ」にするか
  2. 空氏名を `(氏名未設定)` と表示するか
  3. 無効・空氏名の一括整理（非表示／削除）を許可するか

### PO-OCCUPATION-MASTER-EMPTY: 城東の職種マスタが 0 件

- **実測**: `occupations` where clinic_id=2 → **0 件**。スタッフ編集の職種 Select は有効職種のみ出すため選択肢が空
- **予約の出勤医師判定・FK には職種は使わない**（スタッフ種別 `doctor` が本線）
- **問い（PO）**: 職種マスタは運用必須か。必須なら STG／各医院への初期「獣医師」「動物看護師」投入を公式手順にするか

---


### NOTE-STAFF-STARTTIME-RDT: スタッフマスタで `startTime` TypeError（アプリ外の疑い）

- **ユーザー報告（2026-09-13）**:
  ```
  Uncaught TypeError: Cannot read properties of undefined (reading 'startTime')
      at et.reportAllChanges (<anonymous>:2:19429)
  ```
  スタッフマスタページを開いたときに発生、とのこと。
- **調査（2026-09-13・Chrome for Testing / CDP 9222）**:
  - `/settings/staff`（clinic 1/2）で **同 TypeError は未再現**（exception 0）
  - アプリソースに `reportAllChanges` は **存在しない**
  - スタックがすべて `<anonymous>` / `VM*` で、**React DevTools 系の `reportAllChanges` と一致する形**
  - 当該ブラウザに `__REACT_DEVTOOLS_GLOBAL_HOOK__` あり
- **判断**: 現状は **製品コード起因と断定できない**（拡張機能／DevTools の可能性が高い）。ユーザー環境（通常 Chrome + 拡張）での再現スタック（ファイルURL付き）があれば再判定。
- **次アクション**: シークレットウィンドウ（拡張OFF）でスタッフマスタを開き、同エラーが消えるか確認してもらう

## バグではない（記録のみ）

- **「本日は医師が出勤していないため予約できません」**: 対象日に有効医師の勤務シフトが無いときの仕様ガード。出勤 0 の日（例: 2026-03-23）で再現済み。バグではない

---

## 修正プラン（2026-09-13・未実装）

本節は上記9件（BUG6件・PO3件）の修正計画と、調査メモ1件の切り分け計画。計画作成中に追加された担当者表示バグと `startTime` 調査も対象に含めた。既存の再現記録と `OPEN` は維持する。上記の「候補」と現行コードが異なる場合は、本節の追加調査を踏まえて着手する。今回の確認範囲はローカル `main` / `7d0f3543c96ec61ac74b23b687b9af98e91726be` のコードと本メモであり、記載済みのDB件数・UAT結果を今回再測定したものではない。

- **成果物の範囲**: このファイルへの計画追記。製品コードの修正、データ更新、reset、migration、STG操作は未実施。
- **正本との関係**: 製品FAILの入口は引き続き [todo.md](todo.md#product-bugs)、実行管理の正本はLinear。本節はローカル案。Linear MCPの検索は実行できたが、対象IDとの対応を特定できておらず、外部チケット状態は **UNKNOWN**。登録・更新は別途承認後に行う。
- **業務目的**: 担当未選択の予約での再入力をなくす、死亡ペットへの新規業務登録を防ぐ、予約枠の設定不備を満枠と混同させない。不要な確認ダイアログや設定の二重管理は追加しない。
- **責任者・仕様判断**: 計画作成の依頼元は本セッションのユーザー。未確定の業務仕様・データ修復については、判断を担うPOの個人名と決定内容を実装着手前に記録する。以下の推奨案は承認済み仕様を意味しない。

<a id="bug-fix-order"></a>

### 着手順序と依存関係

| 順序 | 対象ID | 最初に行うこと | 実装・運用の前提 |
|:---|:---|:---|:---|
| 1・臨床安全 | BUG-RES-DECEASED-STATUS-BYPASS | 不整合ペットを拒否する回帰テストと、死亡判定の適用範囲を整理 | 死亡判定の契約確定。バックフィル完了は待たない |
| 2・予約失敗 | BUG-RES-DOCTOR-ID-ZERO | 単体・複数予約の作成時に0が残る経路をテストで固定 | 担当未選択の既存契約に沿って修正可能。1の判断待ち中に先行可 |
| 2と連続・担当者表示 | BUG-RES-STAFF-SELECT-ORPHAN-LABEL | 区分・日付変更後の候補と選択値の不一致を再現 | doctorの正規化と同じフォームを触るため順次実装 |
| 3・空き枠 | BUG-RES-AVAILABLE-TIMES-404 | 設定未登録・満枠・通信障害のAPI/UI契約を決める | 未設定時の時刻入力可否と応答形式の確定 |
| 4・a11y | BUG-RES-DIALOG-A11Y-CONSOLE | 現行画面で警告元と説明IDの対応を特定 | 再現するコンポーネントを特定してから修正 |
| 独立・環境 | BUG-LOCAL-HANDOFF-CSV-CONTRACT | producer/importer契約と再生成元を確認 | 再生成可能な元データ・producerが必要。取込は別の運用作業 |
| 判断後 | PO-PET-DECEASED-DATA-BACKFILL | 不整合を由来・死亡日根拠の有無で分類 | 死亡判定の契約、対象環境、修復・監査・復旧手順の承認 |
| 判断後 | PO-STAFF-BLANK-NAME-LIST | 表示改善とデータ整理を分けて決める | デフォルト絞り込み・代替表示の承認 |
| 判断後 | PO-OCCUPATION-MASTER-EMPTY | 職種の必須性と医院別の初期値を決める | マスタ責任者・正式な職種名の確定 |
| 独立・原因調査 | NOTE-STAFF-STARTTIME-RDT | 同じ操作を拡張なしの環境と比較 | 再現元URL・stackの特定。製品バグとは未確定 |

実装単位ごとに同名IDのclaimを確認・取得する。1〜3は `reservation` のサービスやテストが重なるため、同じファイルを同時編集しない。並行する場合は別worktreeと排他的な担当パスを使う。製品バグの回帰テストには合成データを用い、古いhandoffや実データの復旧を全項目の前提にしない。

<a id="plan-bug-res-deceased-status-bypass"></a>

### 1. BUG-RES-DECEASED-STATUS-BYPASS

**方針（実装済み・本ユニット）**: `status=deceased` または `deceased_at != null` のどちらかが死亡を示せば、新規予約を拒否する。日時不明の個体を生存扱いにせず、死亡日の推測補完も行わない。

1. `ValidatePetNotDeceased` を OR 契約へ更新（会計・入院・カルテ・検査も共有ヘルパー経由で継承）。履歴訂正の適用範囲は広げない。
2. 合成ペット `alive/null` / `deceased/null` / `deceased/dated` / `alive/dated` の RED→GREEN を sharedkernel・Create・CreateBatch・admin Create で固定。
3. FE は `isPetDeceasedForClinicalWrite`（選択 UI + 予約 submit）で同じ OR 契約。画面 disable のみに依存しない。
4. S01 コメントを OR 契約へ整合。データ backfill は別ユニット。

**主な対象**: 上記共通ガード、[reservation_service.go](backend/internal/reservation/reservation_service.go)、[appointment_admin_service.go](backend/internal/reservation/appointment_admin_service.go)、[予約フォーム](frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx)、[受付カード](frontend/src/features/reception/components/AppointmentCard.tsx)。

**完了条件**: 不整合ペットの単体予約は既存の死亡エラーで拒否され、予約・関連データの増分が0件。複数ペットのうち1頭が死亡でも全件rollbackされる。生存ペットの正常予約と、許可された履歴操作は成立する。実データのバックフィルは別項目として残す。

<a id="plan-bug-res-doctor-id-zero"></a>

### 2. BUG-RES-DOCTOR-ID-ZERO

**現行コードでの追加確認**: `/reservations` と `/reservations/batch` の入力は [reservation_request.go](backend/internal/reservation/reservation_request.go)。上記にある `appointment_admin_request.go` は別の管理者経路である。`Create` / `CreateBatch` は0を保存し、`buildReservationUpdate` は更新時の0をNULLにする。担当者所属・対応区分の検証も0を未指定扱いで通すため、検証と保存が不一致になっている。

1. [transforms.test.ts](frontend/src/features/reservations/api/transforms.test.ts) と [use-reservation-actions.test.ts](frontend/src/features/reservations/hooks/use-reservation-actions.test.ts) に、未選択・`"0"`・有効IDについて、実際の送信payloadを検証するREDを追加する。既存ペット単体、複数ペット、新規飼主からの予約を対象にする。
2. [transformToCreateRequest](frontend/src/features/reservations/api/transforms.ts) の作成用変換で、空値と互換用の `"0"` を未指定へ正規化する。負数・小数・数値でない値・安全に表現できないIDは入力エラーとし、未指定扱いで黙って保存しない。未選択を0にするフォーム初期値・選択解除経路が見つかった場合は、[担当者欄](frontend/src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx) とその入力元を修正する。
3. BEは作成用の正規化を共通化し、通常作成・一括作成・管理者作成の検証、競合判定、保存がすべて同じ `DoctorID` を参照するようにする。HTTP変換だけに依存せず、サービスを直接呼ぶ経路も覆う。呼出し元のinputを破壊的に変更しない。
4. 正のIDには、同医院への所属・予約区分への対応検証をwrite transaction内で維持する。無効・他院・非対応IDをnilへ落として成功させない。院内とLIFFのスタッフ公開条件は既存契約を維持する。
5. 作成とPATCHの意味を分けてテストする。作成は省略/null/0を未指定とし、PATCHは省略で既存担当を保持、0で担当解除という現行契約を保つ。`is_designated=true` と未指定の組合せは既存仕様を確認し、矛盾する場合は明示エラーにする。

| 入力・経路 | 期待結果 |
|:---|:---|
| 作成: `doctor_id` 省略 / null / 0 | 他の条件を満たせば201、保存値はNULL |
| 作成: 同医院・対応可能な正の担当者ID | 201、指定IDを保持 |
| 作成: 他院 / 存在しない / 非対応の担当者ID | 制約違反の前に既存の適切な入力・参照エラー、保存0件 |
| 作成: 負数 / 小数 / 不正文字列 | 入力エラー。未指定予約として作成しない |
| PATCH: `doctor_id` 省略 / 0 | それぞれ既存値保持 / NULLへ解除 |

**完了条件**: 担当者未選択の画面操作から送信payload・API応答・保存値まで確認し、FKエラーが再発しない。出勤医師0の日の既存ガード、枠競合、容量制約、一括作成の原子性も維持する。`"0"` が変換可能というコード上の事実と、実際にその値を発生させる画面操作の特定は別々に記録する。

<a id="plan-bug-res-available-times-404"></a>

### 3. BUG-RES-AVAILABLE-TIMES-404

**現行コードでの追加確認**: [GetReservationAvailableTimes](backend/internal/reservation/reservation_handler.go) は院内用 `GetStaffAvailableTimes` を優先するが、[getAvailableTimes](backend/internal/reservation/liff_service_availability.go) 内でLINE設定を読む。FEの [ReservationFormFields](frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx) は取得結果が `undefined` の場合には時刻候補を生成し、`[]` の場合には空の候補mapを作る。したがって、**404を一律に200 `[]`へ置換するだけでは、新規予約の時刻選択を損なう**。

**推奨案・実装前に契約確定**: 院内では「設定未登録」を「計算済みで空きなし」から区別できる応答にし、未登録時は案内付きで手動時刻入力を使う。既存の成功配列契約を維持できる専用エラーコード方式を第一候補とし、HTTPステータスとコードは既存エラー体系に合わせてAPI仕様で確定する。状態付き成功レスポンスへ変更する場合は、利用箇所と生成型の互換変更を同じ単位で行う。未設定の営業時間を9〜19時などへ暗黙補完する案は採らない。

1. 設定のnot-foundだけを院内用の「空き枠設定未登録」として識別する。予約区分・明示スタッフの不存在や他院参照、DB障害、設定JSON破損まで同じ状態へ変換しない。早期returnで医院・関連ID検証を飛ばさない。
2. [useGetReservationAvailableTimes](frontend/src/hooks/use-reservation-types.ts) とフォームで下表を実装する。未設定を単なる通信失敗として再試行し続けず、既存入力を保持する。設定未登録だけに手動入力の例外を限定する。
3. 設定済みの医院は既存の営業時間・休診日・シフト・休憩・予約不可時間・容量計算を利用する。LIFFの設定必須性、無効区分拒否、公開スタッフ条件は変えない。GETで設定行を自動作成しない。
4. handler/service/FEの回帰テストを追加し、医院切替時に他院の結果が残らないことも確認する。OpenAPI変更が必要なら、手書きの正本と呼出し側を更新し、生成物はユーザー実行の `make codegen` 後に整合確認する。

| 状態 | API/UIの受入条件 |
|:---|:---|
| LINE設定未登録・院内予約 | 未登録を識別でき、案内と手動入力が成立。保存時の出勤・休診・競合・容量・死亡ガードは維持 |
| 設定済み・空きあり | 計算された枠を表示し、選択して登録可能 |
| 設定済み・休診 / 満枠 | 200 `[]` を空きなしとして表示。全時間帯へフォールバックしない |
| DB障害 / 通信失敗 / 設定破損 | 取得エラーとして表示。未登録や空きなしへ置換しない |
| 不存在・他院の区分 / スタッフ、権限なし | 既存の参照・認可エラーを維持し、他院の存在を漏らさない |
| LIFF・設定未登録 | 院内の手動入力例外を流用せず、既存の予約制限を維持 |

**完了条件**: 設定なしの医院でも理由不明の404を出さず、担当未選択を含む院内予約が適切な時刻で登録できる。満枠・休診・障害時に予約可能と誤表示しない。設定seedの投入だけを製品修正の完了根拠にしない。

**実装メモ (2026-09-13 / att-bug-res-available-times-20260913-001)**: 院内 `GetStaffAvailableTimes`（`requireActive=false`）で `line_reservation_settings` not-found を `LINE_RESERVATION_SETTINGS_UNSET`（HTTP 422 + code）へ。LIFF `GetAvailableTimes` は従来どおり not-found。FE は code 検知時のみ案内＋`TIME_OPTIONS` 手動入力、200 `[]` は空、他エラーはエラー表示（全日フォールバック禁止）。available-times queryKey に clinicId を付与。seed/自動作成なし。検証: `docker compose exec backend go test ./internal/reservation`（ok）、`docker compose exec frontend npx vitest run src/hooks/use-reservation-types.test.ts src/components/shared/ReservationFormModal`（82 / 72 passed）。

<a id="plan-bug-res-dialog-a11y-console"></a>

### 4. BUG-RES-DIALOG-A11Y-CONSOLE

**現行コードでの追加確認**: [CommandDialog](frontend/src/components/ui/command.tsx) は既に `DialogTitle` と `DialogDescription` を持つ。[予約フォーム](frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx) と [ヘッダー](frontend/src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx)、[予約区分選択](frontend/src/components/shared/ReservationFormModal/ReservationTypePickerDialog.tsx) にも説明がある。「共有CommandにDescriptionがない」という上記の仮説は現行コードでは成立しない。

1. 現行コードが稼働しているブラウザで新規予約を開き、consoleのcomponent stackと各ダイアログの `aria-describedby`・説明要素ID・マウント時点を照合する。固定IDの重複、上書き、入れ子、開閉・再表示を調べる。旧ビルド由来ならその差を記録する。
2. 発生元を絞って、説明要素と参照IDが同じダイアログに対応するよう修正する。既存UIラッパーの標準的なID管理を優先する。consoleの抑制や説明の無条件削除を修正にしない。
3. モーダルの実コンポーネントを描画する回帰テストで、警告がないこととアクセシブルな名前・説明が得られることを検証する。Dialogをmockして警告を消したテストは受入証拠にしない。
4. ブラウザで初回表示・閉じて再表示・入れ子の予約区分選択・キーボード操作・フォーカス復帰を確認する。

**完了条件**: 再現していた同じ操作で対象警告0件、説明の参照切れ0件。現行コードで再現できない場合は `OPEN` のまま追加調査結果を記録し、静的にDescriptionがあるだけで修正済みにしない。

<a id="plan-bug-local-handoff-csv-contract"></a>

### 5. BUG-LOCAL-HANDOFF-CSV-CONTRACT

**方針**: importerの現行contractに対応したproducerからbundleを再生成する。manifestのdigestだけを置換したり、validatorを緩めたりしない。

1. [cutover_contract.go](backend/internal/csvimport/cutover_contract.go)、[cutover_contract_validate.go](backend/internal/csvimport/cutover_contract_validate.go)、[handoff手順](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) を基に、producerのrevision、列契約、provenance、元データ、対象医院を一覧化する。旧bundleは証拠として保持する。
2. 再生成元が揃えば、現行producerでCSV・manifest・hash・provenanceを一体で生成する。生成元を確認できない場合は環境項目を `BLOCKED` とし、予約の回帰検証は合成データで継続する。`REHEARSAL_ONLY` を本番投入可能へ昇格させない。
3. `stage-old-db-handoff.sh` / `check-old-db-handoff.sh` の既存経路で配置を確認し、取込前のpreflightで全対象bundleを検証する。古い契約・hash改ざん・医院不一致を拒否し、新契約が受理されることを確認する。fixture自体の更新で解決する場合は、検証器のコード変更を増やさない。
4. ユーザーが取込検証を行う段階で、専用ローカル環境のバックアップ・対象・復旧方法を確定し、既存手順に従って実行する。エージェントからreset、DBへの投入、migration適用を自動実行しない。

**完了条件**: まず全bundleのpreflight PASSを確認し、その後に承認されたローカル取込で医院別件数・owner/pet参照・再実行時の挙動を確認する。preflight PASSと取込成功は別の証拠とし、STG相当や本番cutover合格へ読み替えない。

<a id="plan-po-pet-deceased-data-backfill"></a>

### 6. PO-PET-DECEASED-DATA-BACKFILL

**推奨案**: 死亡日は根拠がある個体だけ補完し、根拠不明の個体は死亡表示と新規write拒否を保持したまま要確認に残す。現在日時による一括補完や、一律 `alive` への戻しは行わない。

1. POが死亡判定の正本、死亡日時不明の表現、対象環境（local/STG/本番cutover）、修復責任者を決める。日時必須を採る場合も、根拠がない個体を処理する方法を先に定義する。
2. 承認された読取調査で、医院・データ由来・根拠の有無別の件数を再集計する。既存メモの件数を現在値として流用しない。個人情報を含む原票はアクセス制限下で扱い、台帳には集計と証拠の参照だけを残す。
3. dry-runの変更前後差分、適用条件、更新済み行のスキップ、監査sink、並行更新の競合検出、再実行性、復旧方法を備えた修復手順を作る。実データ適用前に合成データで検証する。
4. 適用は対象環境と差分への明示承認後に実施し、医院別の修復件数・未解決件数・予約拒否の結果を照合する。死亡状態変更の既存監査を迂回しない。

**完了条件**: 承認対象の全件について修復または要確認の帰属が明確で、監査と再実行検証が成立する。未解決が残る場合は件数と担当を残し、本項目を完了扱いにしない。製品側の死亡ガード修正とは独立して判定する。

<a id="plan-po-staff-blank-name-list"></a>

### 7. PO-STAFF-BLANK-NAME-LIST

**推奨案**: 管理一覧は「有効のみ」を初期表示とし、「すべて／無効のみ」から過去スタッフにも到達できるようにする。空白・空文字の氏名は表示時だけ `(氏名未設定)` に置換する。既存スタッフの氏名や有効状態を表示都合で書き換えない。

1. POに初期フィルタと代替表示を確認する。無効スタッフの閲覧・編集や履歴の参照は残し、一括削除は本修正に含めない。
2. [staff-settings-model.ts](frontend/src/features/master/routes/staff-settings-model.ts) のstatusフィルタと一覧画面を再利用し、表示箇所にだけ代替名を適用する。現行 [staff_repository.go](backend/internal/staff/staff_repository.go) は全件を `sort_order, name` 順で返すため、まずFEの初期フィルタ・表示変更に限定する。着手時にページングへ変わっていた場合は、フィルタ・総件数・並び順を同じ条件で計算する。
3. 有効・無効、氏名あり・空文字・空白だけの組合せ、0件、ページ跨ぎ、検索、医院切替をテストする。予約の担当者候補や過去予約の担当者表示に副作用がないことを確認する。

**完了条件**: 初期表示が空欄で埋まらず、件数と行が一致する。無効スタッフに明示操作で到達でき、元データと過去予約の参照は維持される。データ削除・統合は別の判断と手順にする。

<a id="plan-po-occupation-master-empty"></a>

### 8. PO-OCCUPATION-MASTER-EMPTY

**推奨案**: まず0件時に「職種が未登録」であることを説明し、権限のある利用者を既存の職種マスタ登録へ案内する。職種未登録を理由に予約やスタッフ保存へ新しい必須制約を加えない。運用必須と決まった場合だけ、医院別の初期登録手順を整備する。

1. POが職種の必須性、正式名称、医院別差異、管理責任者を決める。初期「獣医師」「動物看護師」は候補であり、この計画だけで投入しない。
2. [use-staff-settings-lookups.ts](frontend/src/features/master/hooks/use-staff-settings-lookups.ts)、[staff-settings-model.ts](frontend/src/features/master/routes/staff-settings-model.ts) と [occupation_repository.go](backend/internal/staff/occupation_repository.go) を確認する。職種APIの実装は `staff` ドメインにあり、現行の `occupation_id` は任意。スタッフの既存職種値が無効の場合も、編集画面の表示用に保持できるか確認する。
3. 必須化する場合は、既存マスタ登録の仕組みで同一医院の重複登録を防ぐ手順を作る。既存行・編集済み名称を上書きせず、追加だけをdry-runで示す。コードで架空の選択肢を生成しない。
4. 0件・有効職種あり・無効職種のみ・登録権限なし・他院の職種指定をテストする。出勤医師判定のスタッフ種別と、予約区分に職種を紐付けた場合の判定を混同しない。

**完了条件**: 職種0件でも理由と次の操作が分かる。承認された医院だけに重複なく初期登録でき、他院の職種を割り当てられない。職種を任意のまま運用する場合は、その決定を記録して項目の扱いを確定する。

<a id="plan-bug-res-staff-select-orphan-label"></a>

### 9. BUG-RES-STAFF-SELECT-ORPHAN-LABEL

**現行コードでの確認**: [SearchableSelect](frontend/src/components/ui/searchable-select.tsx) は現在の候補から選択値のラベルを引くため、候補外の値はプレースホルダになる。一方、[予約区分の変更](frontend/src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx) は `type` だけを書き換え、`doctor` は残す。[候補フィルタ](frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts) は取得中も空配列を返すため、「取得中」と「取得済みで選択不可」を分ける必要がある。

**推奨案**: フォームで選択値・表示・送信値を一致させる。区分・日付の変更により担当が指定できなくなった場合は、その理由を表示して再選択または明示解除を求める。候補取得中に担当を勝手に消したり、表示名だけを残して非対応の担当者を有効に見せたりしない。

1. `ReservationFormModal.staff-candidates.test.tsx` と保存アクションのテストに、担当選択→非対応区分へ変更、日付変更→非出勤、対応可能な区分変更、候補の遅延取得・取得失敗のケースを追加する。DOMの表示と内部選択値、送信payloadを合わせて確認する。
2. `ReservationFormFields` で現在の医院・日付・区分に対する候補取得状態を管理し、取得済みの候補から選択の有効性を導出する。不整合時は担当者欄に「この条件では指定できない」旨と再選択・解除を提示し、解消前の送信を止める。取得中・失敗を確定した候補外判定に使わない。
3. 選択済みの名前は同医院の取得済みスタッフ情報から表示用に解決し、有効候補と表示用の値を区別する。フォーム内の対応を優先し、共有 `SearchableSelect` を変更する場合は必要なpropだけを追加して他のSelectも回帰確認する。候補への無条件再追加はしない。
4. 既存予約編集での過去担当の表示と、担当・区分・日時の変更時に必要な再検証を区別する。表示都合で保存済み担当を変更しない。明示解除後は `BUG-RES-DOCTOR-ID-ZERO` の未指定契約で送信する。

**完了条件**: 有効な担当者を選ぶと名前が表示され、payloadも同じIDになる。条件変更で無効になった担当者は未選択に偽装されず、再選択・解除まで送信できない。取得中・失敗で既存選択を破棄せず、履歴表示と医院分離を維持する。スタッフの対応区分を一括で増やしてUI不整合を隠す対応は行わない。

**実装メモ (2026-09-13 / att-bug-res-staff-orphan-20260913-001)**: `SearchableSelect` に `fallbackLabel` を追加。`ReservationFormFields` で候補 settle 後の confirmed orphan を導出し、表示名・理由・解除を `ReservationTypeAndStaffFields` に渡し、Modal submit で遮断。loading/error は orphan 扱いにしない。編集初期化で保存済み `doctor` を書き換えない。capable_courses の一括拡張なし。検証: `docker compose exec frontend npx vitest run src/components/shared/ReservationFormModal`（69 passed）。

<a id="plan-note-staff-starttime-rdt"></a>

### 10. NOTE-STAFF-STARTTIME-RDT（調査計画）

**方針**: 製品コード・ブラウザ拡張・開発ツールのどこで発生するかを特定してから修正対象を決める。`reportAllChanges` がアプリソースにないことだけでは、アプリ外と断定しない。

1. 承認されたブラウザ対象で、発生操作、ビルドrevision、ファイルURLを含むstack、発生時刻を記録する。個人情報や認証情報は証拠から除外する。
2. 同じビルド・同じ操作を、拡張のない専用プロファイルと報告環境で比較する。シークレットモードでも許可済み拡張が動くことがあるため、拡張の無効化状態を確認する。ユーザーの通常プロファイルや拡張設定を勝手に変更しない。
3. アプリ起因なら該当コード・再現条件・REDテストを持つ製品バグへ整理する。拡張起因と特定できた場合は拡張名・バージョン・比較結果と回避手順を記録する。再現できない場合は原因 `UNKNOWN` として必要な追加証拠を残す。

**調査完了条件**: 同条件の比較とstackで原因の帰属を説明できること。製品修正が必要なら別途その実装・検証が完了するまで閉じない。未再現だけを「バグなし」「修正済み」の根拠にしない。

<a id="bug-fix-verification"></a>

### 検証と完了報告

実装時は各項目のRED → GREEN → 関連回帰の順に確認し、対象worktreeがコンテナへマウントされていることを先に確認する。下記は実装後の限定検証候補であり、今回実行した結果ではない。

```bash
# 担当医のpayload・保存アクション
docker compose exec frontend npx vitest run src/features/reservations/api/transforms.test.ts src/features/reservations/hooks/use-reservation-actions.test.ts

# 予約フォーム: 未設定・満枠・死亡・a11yの回帰を追加した後
docker compose exec frontend npx vitest run src/components/shared/ReservationFormModal

# 予約の作成・設定欠落・分離・原子性と共通死亡ガード
docker compose exec backend go test ./internal/reservation ./internal/sharedkernel

# handoffの契約検証を変更した場合
docker compose exec backend go test ./internal/csvimport

# ドキュメント差分（未追跡ファイルは別途ベースライン比較も行う）
git diff --check -- bug.md
```

- 死亡ガードの共通化で影響する他ドメイン、スタッフ・職種、共有UIについては、確定した変更パスに対応する限定テストを追加実行する。名前だけ合うテストや0件実行をPASSとしない。
- FK・transaction・競合・医院分離は、既存の専用Docker実DBテスト手順に従って確認する。mockでの保存値確認だけでは実DB検証完了にしない。共有DBへ直接SQLを実行しない。
- ブラウザは承認されたURL・専用プロファイル・合成データを使い、元の失敗操作と保存後の再読込まで確認する。STG/UATは別途実施するまで未検証とする。
- coverageは [現行方針](docs/ops/coverage-policy.md) に従い、限定テストの結果から全体達成率を推測しない。全体test/build/lint、migration、codegenは自動実行しない。
- バグごとに、変更ファイル、対象revision、実行コマンドと終了状態、UI/API/DBの受入証拠、未実施事項を報告する。コード修正済み・限定テストPASS・UAT完了・データ修復完了を区別し、計画追記のみで `OPEN` を閉じない。

**今回の検証範囲**: 文書の差分・全10項目（調査メモ1件を含む）の網羅・追加した参照先・既存WIPの保全を確認する。文書のみの変更のため、実装テストとruntime検証は不要（未実施）。
