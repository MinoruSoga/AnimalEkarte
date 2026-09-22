# 一時バグ／障害メモ（ユーザー依頼 2026-09-13）

> **注意**: root `bug.md` は 2026-09-08 に `todo.md` の [製品 FAIL](todo.md#product-bugs) へ統合廃止済み。  
> 本ファイルはユーザー明示依頼により **一旦** 再作成したメモ台帳。製品 FAIL 正本は引き続き `todo.md#product-bugs`。  
> `BUG-LOCAL-HANDOFF-CSV-CONTRACT` は環境／handoff fixture の BLOCKED。  
> `BUG-*` はローカル（STG スナップショット）で切り分け済み。`PO-*` は仕様・データ修復の判断が必要。

更新日: 2026-09-13

修正プラン: [着手順序と依存関係](#bug-fix-order) · [検証と完了条件](#bug-fix-verification)。各項目の方針と詳細手順は下表から確認できる。実装・STG配備・実機での解消確認は区別して各項目に記録する。

## 索引

| ID | status | area | severity | 種別 | 修正プラン |
|:---|:---|:---|:---|:---|:---|
| BUG-LOCAL-HANDOFF-CSV-CONTRACT | OPEN | ops / local-db | Medium | handoff preflight BLOCKED | 現行契約でbundleを再生成し、取込前に検証する。[詳細](#plan-bug-local-handoff-csv-contract) |
| BUG-RES-DOCTOR-ID-ZERO | FIXED | reservation | High | **STG実機確認済み**（担当未選択・兼務先での登録） | `created_by` の主所属制約を追加修正。STG `916159093`で城東・担当者未選択の保存201、再読込・詳細再表示200を確認。[詳細](#plan-bug-res-doctor-id-zero) |
| BUG-RES-DIALOG-A11Y-CONSOLE | FIXED | reservation / a11y | Low | **バグ断定**（DialogContent Description 欠落コンソール警告） | 固定ID上書きで Radix DescriptionWarning が context.descriptionId を見失っていた。Radix 管理IDに戻し警告0。[詳細](#plan-bug-res-dialog-a11y-console) |
| BUG-RES-STAFF-SELECT-ORPHAN-LABEL | FIXED | reservation / UI | High | **バグ断定**（担当者選択後に表示が消える） | 候補外でも表示名を保持し、確定 orphan は理由表示＋解除/再選択まで送信遮断。[詳細](#plan-bug-res-staff-select-orphan-label) |
| BUG-RES-AVAILABLE-TIMES-404 | FIXED | reservation | Medium | **バグ断定**（LINE設定欠落で院内API 404） | 院内は設定未登録を識別可能な unset（422 + code）にし案内付き手動時刻のみ許可。空枠/障害/LIFF必須は維持。[詳細](#plan-bug-res-available-times-404) |
| BUG-RES-DECEASED-STATUS-BYPASS | FIXED | reservation / pet | High | **バグ断定**（status=deceased なのに予約可） | `status=deceased` OR `deceased_at` で新規write拒否。データ補完は別項。[詳細](#plan-bug-res-deceased-status-bypass) |
| PO-PET-DECEASED-DATA-BACKFILL | OPEN | data / pet | Medium | **PO確認**（不整合データの修復方針） | 対象・死亡日の根拠・監査・復旧を確定してからデータ修復する。[詳細](#plan-po-pet-deceased-data-backfill) |
| PO-STAFF-BLANK-NAME-LIST | FIXED | staff UX / data | Low | **PO確認→実装**（空氏名の一覧表示方針） | Grill Recommended: 初期有効のみ＋表示 `(氏名未設定)`。DB書換・一括削除なし。[詳細](#plan-po-staff-blank-name-list) |
| PO-OCCUPATION-MASTER-EMPTY | FIXED | master / data | Low | **PO確認→実装**（職種マスタ0件の扱い） | Grill Recommended: 0件は未登録案内＋職種マスタへ誘導。`occupation_id` は任意のまま。偽選択肢・自動投入なし。[詳細](#plan-po-occupation-master-empty) |
| NOTE-STAFF-STARTTIME-RDT | OPEN | staff console | Low | **調査**（startTime TypeError・アプリ外の疑い） | 拡張なしの環境と比較し、stackから原因を特定して修正対象を決める。[詳細](#plan-note-staff-starttime-rdt) |
| BUG-ACCT-CLOSE-PERM-DEFAULT | OPEN | billing / permission | Medium | **バグ断定**（既定権限でレジ締め不能。edit は dead grant、必須の create は全グループ未付与） | `cash-register-close` の既定付与を create へ修正するか、締め endpoint の要求スコープを edit に揃えるかを裁定し実装。[詳細](#plan-bug-acct-close-perm-default) |
| BUG-S09-FIXTURE-TEARDOWN | OPEN | testing / fixture | Low | **バグ断定**（締め実行済みの S09 合成 clinic が teardown で削除不能。append-only 台帳との矛盾） | teardown の scoped 削除設計を裁定（closes/adjustments/holidays を含めるか、trigger 無効化の運用経路を持つか）。[詳細](#plan-bug-s09-fixture-teardown) |
| BUG-AGG-NO-VISIT-REVENUE | OPEN | aggregation / revenue | Medium | **バグ断定**（来院なし飼主が完了会計を持っても売上ランキングに一切出ない。`include_no_visit=false` 既定除外が売上タブにも適用） | `include_no_visit` フィルタの適用範囲を最終来院タブ限定に修正し、売上軸では来院有無に依らず算入する。[詳細](#plan-bug-agg-no-visit-revenue) |
| BUG-TRIM-KANBAN-IN-CONSULTATION | OPEN | trimming / reception | High | **バグ断定**（受付済トリミングカードの「カルテ作成」が `in_consultation` 遷移を試みるが、カルテ必須ガードで 409。トリミングは medical_record を持たないため永久に受付済のまま） | `validateInConsultationHasMedicalRecord` を trimming category で免除するか、FE が trimming 作成 POST に `status: in_consultation` を載せる経路に統一する。[詳細](#plan-bug-trim-kanban-in-consultation) |
| BUG-BILLING-UNBILLED-MR-EXCLUSION | OPEN | billing / unbilled | Medium | **バグ断定**（カルテ連携 pending 会計から明細を soft-delete しても、billing が `medical_record_id` を保持する限り元の treatment が未請求候補に復帰しない。サイレント請求漏れ） | `FindUnbilledByPetID` の billing 単位除外を明細単位の除外に限定し、削除された明細の請求元を再候補化する。[詳細](#plan-bug-billing-unbilled-mr-exclusion) |
| BUG-RES-OVERLAP-500 | OPEN | reservation / API | Medium | **バグ断定**（同一スタッフ・時間帯重複の予約作成が DB 排他制約 `excl_appointments_doctor_timerange` の 500 として漏れる。完全一致は正しく 409） | 排他制約違反（23P01 / exclusion violation）を conflict 409 へマッピングする。[詳細](#plan-bug-res-overlap-500) |
| BUG-LIFF-HEALTHCARD-OWNER-SYNC | OPEN | liff / line_customers | High | **バグ断定**（LIFF トークン連携が `owners.line_user_id` のみ書き `line_customers.owner_id` を更新しない。health-card は `line_customers.owner_id` 経由解決のため、連携済み飼主のヘルスカードが「ペット情報はありません」のまま） | `LinkAccount` で line_customers 行を FindOrCreate＋owner_id 紐付けするか、手動 link-owner を正式な第2段として仕様化するか裁定。[詳細](#plan-bug-liff-healthcard-owner-sync) |
| BUG-ACCT-INS-SIGN-MISMATCH | OPEN | billing / insurance | High | **バグ断定**（保険適用会計の新規確定が FE/BE の `insurance_amount` 符号規約不整合で必ず 400。FE は負値送信だが BE は `billing=total−insurance_amount` で正値を前提） | FE/BE どちらかの符号規約に統一し、保険付き会計の complete を回帰固定。[詳細](#plan-bug-acct-ins-sign-mismatch) |
| BUG-ACCT-INS-EDIT-REWRITE | OPEN | billing / insurance | Medium | **バグ断定**（確定済み会計の修正保存で、未変更の `insurance_amount`/`billing_amount` が recalc 値に上書きされ、保存後に保険表示が消える。`hasInsurance` が `insurance_amount < 0` 推定のため正値・0 の保存値は OFF 表示） | 編集保存で保険フィールドを変更しない限り保存値を維持する契約へ修正し、`hasInsurance` を保存フラグ由来にする。[詳細](#plan-bug-acct-ins-edit-rewrite) |
| BUG-DIALOG-FOCUS-RESTORE | OPEN | shared UI / a11y | Low | **バグ断定**（治療プラン検索ダイアログを Escape で閉じるとフォーカスが呼出元ボタンに戻らず `document.body` に落下。実クリック開閉・キーボード開閉の両経路で再現） | Radix Dialog の focus restoration が効いていない経路を特定し、閉じた後にフォーカスがトリガーへ戻ることを固定する。[詳細](#plan-bug-dialog-focus-restore) |
| BUG-BILLING-TAX-TYPE-DROPPED | OPEN | billing / master | High | **バグ断定**（マスタ登録の税区分（内税/非課税）が会計明細へ伝播せず全て外税10%で請求される。処置・診察の未請求候補と物販マスタ追加の双方で再現し DB も `excluded` 固定） | 未請求候補の集約と物販追加の両経路で master の tax_type/tax_rate をスナップショット引継ぎする契約へ修正。[詳細](#plan-bug-billing-tax-type-dropped) |
| BUG-ACCT-DUP-COMPLETE-500 | OPEN | billing / idempotency | Medium | **バグ断定**（同一カルテへの二重会計確定が意図した 409「このカルテには既に会計があります」に届かず 500。UNIQUE 競合後の replay 判定クエリが abort 済み tx 上で実行され 25P02 になる。逐次・真並行の双方で再現。行は作られずデータは守られるが、クライアントには不透明な 500 しか返らない） | 競合解決クエリを別接続/tx で実行するか、制約名（`idx_billings_medical_record_id_unique` / completion_request_id）を見て直接 409/replay に分岐する。[詳細](#plan-bug-acct-dup-complete-500) |

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

- **2026-09-13 追加報告・再オープン**: STGの城東で、担当者未選択でも同じエラーが続くとユーザー報告。既存の0→NULL修正が配備されたことと、予約登録が成功することは別の確認事項。
- **追加原因（実DDLで再現済み、STG実エラー制約名は未取得）**: 担当者とは別に認証スタッフIDを `created_by` に保存する。`fk_appointments_created_by_clinic` が `staffs(id, clinic_id)`、つまり主所属を要求するため、所属が許可された兼務先での予約も拒否する。既存002 migrationはカルテの `entered_by` のみを修正し、この予約制約は残っていた。
- **追加修正（PR #396、STG反映PR #397）**: 新規003 migrationで登録者を単独staff FK（削除RESTRICT）に変更。通常・複数ペット・管理画面の保存で、有効な登録者の医院所属またはシステム管理者権限を同一transaction内で検証・共有ロックする。登録者IDと履歴表示は保持する。
- **隔離DB検証（2026-09-13）**: ユーザー承認済みの使い捨てPostgreSQL 18で、001→003・master/login seed投入と再実行が成功。実DDLの旧FKで `doctor_id=NULL` の登録失敗（23503、上記制約名）を再現。新[回帰テスト](backend/internal/reservation/reservation_created_by_fk_test.go)で3登録経路×省略/null/0、未所属・無効actor拒否、権限変更とのロック競合、他院FK維持、履歴取得を検証。新guardのstatement coverageは86.7%。既存の予約repository・医院分離テスト19件もPASS。これはSTG実操作の受入証跡ではない。
- **STG実機検証（2026-09-13）**: 主所属が八王子（医院1）、城東（医院2）にも所属する非管理者のデモスタッフで検証。配備前は担当者を送信しない予約POSTが400「参照先が存在しません」。STG `916159093`の新コンテナへの切替完了後、同条件の合成飼主・ペットを使った画面保存が201となり、ブラウザ再読込後のカレンダーと詳細に表示された。単件GETも200、担当者未指定、登録者ID保持、医院2への保存を確認。
- **デプロイジョブとの区別**: [Backend Deploy 34739608245](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34739608245)は、配布後に旧コンテナ（DDLは001/002のみ）で開始したmigrationがrollout中に終了コード143で失敗。新コンテナへの切替は05:13:17 UTCに100%完了し、起動時migration後のhealth 200と上記予約成功を確認した。予約バグの解消と、デプロイ手順にrollout完了待ちがない問題は別として扱う。
- **完了条件達成**: 実DDLで旧制約による失敗と修正後の登録成功、未所属者の拒否・他院FKの隔離を検証し、STGの城東・担当者未選択で保存・再表示まで確認済み。

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
- **原因（live / real-Dialog 特定済み）**:
  - Culprit: `ReservationFormModal` → `DialogContent` + `ReservationModalHeader` の `DialogDescription`
  - 固定 `id="reservation-form-description"` を `DialogDescription` に渡し、同IDを `DialogContent` の `aria-describedby` にも指定していた
  - Radix `@radix-ui/react-dialog` の `DescriptionWarning` は **context.descriptionId**（自動採番）の要素存在を見る。custom `id` が context ID を上書きするため `document.getElementById(context.descriptionId)` が失敗し警告が出る
  - アクセシブル説明自体は custom id 経路で付いていたが、警告は残る（静的 Description 存在だけでは FIXED にできない理由）
  - Plan §4 の「CommandDialog 欠落」仮説は不成立（現行 CommandDialog は Title+Description あり）
- **修正（2026-09-13 / att-bug-res-dialog-a11y-20260913-001）**:
  - custom description id / 手動 `aria-describedby` をやめ、Radix の Description 配線に委譲（`ChangePasswordDialog` と同パターン）
  - 回帰: `ReservationFormModal.dialog-a11y.test.tsx`（実 Dialog、DialogContent mock なし、Missing Description 警告0 + accessible name/description、入れ子 type-picker）
- **影響**: 開発者コンソール汚染。a11y（スクリーンリーダー向け説明）欠落の兆候。予約 CRUD 自体の機能 FAIL ではない。



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
  1. デフォルトで「有効のみ」にするか → **採用（Recommended）**
  2. 空氏名を `(氏名未設定)` と表示するか → **採用（表示のみ）**
  3. 無効・空氏名の一括整理（非表示／削除）を許可するか → **本ユニット対象外**
- **修正（2026-09-13 / att-po-staff-blank-20260913-001）**: Staff settings 初期 `status is active`。空・空白氏名は一覧/aria のみ `(氏名未設定)`。create/update payload・DB・予約候補ピッカーは未変更。

### PO-OCCUPATION-MASTER-EMPTY: 城東の職種マスタが 0 件

- **実測**: `occupations` where clinic_id=2 → **0 件**。スタッフ編集の職種 Select は有効職種のみ出すため選択肢が空
- **予約の出勤医師判定・FK には職種は使わない**（スタッフ種別 `doctor` が本線）
- **問い（PO）**: 職種マスタは運用必須か。必須なら STG／各医院への初期「獣医師」「動物看護師」投入を公式手順にするか
- **決定（Grill Recommended / 2026-09-13）**: 空マスタ時は未登録案内＋既存 `/settings/occupations` への導線のみ。`occupation_id` はスタッフ保存・予約で必須化しない。偽の Select 選択肢や自動シードは行わない。医院別初期登録は別承認まで対象外。

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

### BUG-ACCT-CLOSE-PERM-DEFAULT: 既定権限モデルで `cash-register-close:create` が全グループ未付与、レジ締めが実行不能

- **発見経緯**: 2026-09-22 UAT S08 手順10（締め後訂正）の前提としてレジ締めを実行したところ、執行アカウント（林 文明）で「この操作を行う権限がありません」。
- **現象**: `POST /api/v1/cash-register/closes` は `cash-register-close:create` を要求するが、既定 seed ではどの権限グループも `can_create` を持たない。結果として新規 clinic では、権限グループを手動編集しない限り誰もレジ締めを実行できない。
- **証跡**:
  - `backend/internal/billing/routes.go` — `cr.POST("/closes", h.requirePermission(...CashRegisterClose, "create"), ...)`
  - `backend/migrations/seeds/002_master/accounts/permission_group_rules.csv` — 全9グループが `cash-register-close` に `can_view=t, can_create=f`（執行系 1/3/5 は `can_edit=t`）
  - `backend/internal/clinic/clinic_service.go` `defaultPermissionRuleTable` — 新規 clinic でも exec は view+edit のみ（create なし）
  - `cash-register-close:edit` を消費する endpoint は routes に存在しない（preview/closes GET=view、closes POST=create のみ）。既定の edit 付与は dead grant
  - FE は `usePermission(ResourceCashRegisterClose).canCreate` でゲート（`CashRegisterClosePage.tsx`）。FE/BE のスコープ自体は一致
- **設計上の矛盾**: テーブル内コメントは「設定系フォールバック（cash-register-close …）: 執行=view+edit（create/delete 不可、hospital-settings と同型）」。しかし hospital-settings は更新系で edit が機能するのに対し、レジ締めは create 操作であり、edit だけ付与しても利用不能。fail-closed 意図なら edit も落とすのが整合（identity-links 先例）
- **回避（本 UAT で実施済み）**: `/settings/permission-groups` で権限グループ「執行」の「レジ締め 作成」を ON → 保存 → 再ログインで締め実行可能になった（執行は master-permission を保持するため self-service 可）
- **影響**: 新規 clinic または未調整 seed で、日次の締め業務が既定のまま実行不能。設定マスタ編集を必須とする運用前提がドキュメント化されていない
- **修正方針候補**: (a) 既定付与を `create` へ修正（seed CSV + defaultPermissionRuleTable）、(b) 締め実行の要求スコープを `edit` に変更、(c) 意図的 fail-closed なら dead edit を除去し運用手順へ明記。(a)/(b) は既存 clinic への権限影響を要評価
- **関連**: UAT 記録 `reports/uat-2026-09-22/S08-accounting-corrections.md`。`todo.md#product-bugs` 正本にも登録

### BUG-S09-FIXTURE-TEARDOWN: S09 合成 clinic の teardown が append-only 台帳と矛盾し削除不能

- **発見経緯**: 2026-09-22 UAT S09 完了後の後処理で `synthetic-closing-fixture teardown --clinic-id 926065` を実行したところ失敗。
- **現象**: `fk_cash_register_close_adjustments_billing_clinic`（RESTRICT）により `billings` 削除で SQLSTATE 23001。`cash_register_closes` / `cash_register_close_adjustments` は `prevent_cash_register_close*_mutation` トリガで append-only（UPDATE/DELETE 禁止）のため手動削除も不可。
- **原因（切り分け済み）**: `DeleteSyntheticClosingFixture`（`backend/internal/billing/synthetic_closing_fixture.go`）の scoped 削除リストが `PaymentSplit/Payment/BillingItem/Billing/PaymentMethodMaster/ClinicSettings/Pet/Owner/StaffClinicAssignment` のみで、`cash_register_closes`・`cash_register_close_adjustments`・`clinic_holidays` を含まない。W-013 の immutability 設計（正しい製品仕様）と teardown が不整合。
- **影響**: S09 前提の「破棄手順を事前に確認」が成立しない。締めを1件でも作った合成 clinic は完全削除不能で、使い捨て DB に恒久的に残る。
- **回避**: 合成 clinic は `s09-clinic-<id>` 命名と id 920000 台で隔離済み。本 UAT では残置。完全削除には (a) teardown 側で closes/adjustments の削除順序と trigger 無効化を扱う、(b) teardown を「clinic 非活性化＋残置」に変更、(c) DB スナップショット復元を破棄手順とする、のいずれかが必要。
- **関連**: `reports/uat-2026-09-22/S09-closing-time-boundaries.md`。`todo.md#product-bugs` 正本にも登録

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

**旧修正時の確認**: `/reservations` と `/reservations/batch` の入力は [reservation_request.go](backend/internal/reservation/reservation_request.go)。`appointment_admin_request.go` は別の管理者経路。作成時の0→NULL正規化は既に実装・STG配備済みだが、上記の追加報告により登録者 `created_by` の主所属FKを追加修正する。以下1〜5は最初のdoctor_id正規化修正の計画として保持する。

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

**実装メモ (2026-09-13 / att-bug-res-dialog-a11y-20260913-001)**: Culprit は `ReservationFormModal`/`ReservationModalHeader` の固定 description ID。Radix `DescriptionWarning` は context.descriptionId を探すため custom id 上書きで警告。手動 `aria-describedby` と custom id を削除し Radix 配線に委譲。検証: `docker compose exec frontend npx vitest run src/components/shared/ReservationFormModal/ReservationFormModal.dialog-a11y.test.tsx`（2 passed）、関連 picker/init-values（8 passed）。console 抑制なし・DialogContent mock なし。

1. ~~現行コードが稼働しているブラウザで新規予約を開き…~~ → real-Dialog vitest で警告再現・原因特定済み（UAT `res_console` と一致）
2. ~~発生元を絞って…~~ → 固定ID上書きを除去済み
3. ~~回帰テスト…~~ → `ReservationFormModal.dialog-a11y.test.tsx` 追加済み
4. 入れ子 type-picker 開閉・Escape 後に親 dialog 残存を vitest で確認済み。ブラウザ実機はログイン要のため RTL を Mode 3 正本とする。

**完了条件**: 再現していた同じ操作で対象警告0件、説明の参照切れ0件。現行コードで再現できない場合は `OPEN` のまま追加調査結果を記録し、静的にDescriptionがあるだけで修正済みにしない。 → **達成（FIXED）**

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

**実装メモ (2026-09-13 / att-po-staff-blank-20260913-001)**: `STAFF_DEFAULT_ACTIVE_FILTERS` + `useMasterCRUD({ initialFilters })` でスタッフ設定のみ有効初期表示。`formatStaffListDisplayName` を `StaffSettingsRow` の表示/aria に適用。payload・DB書換・一括削除・予約候補フィルタなし。検証: `docker compose exec frontend npx vitest run src/features/master/routes/staff-settings-model.test.ts src/features/master/routes/StaffSettings.test.tsx src/features/master/hooks/use-master-crud.test.ts`（63 passed）。

**完了条件**: 初期表示が空欄で埋まらず、件数と行が一致する。無効スタッフに明示操作で到達でき、元データと過去予約の参照は維持される。データ削除・統合は別の判断と手順にする。 → **達成（FIXED）**

<a id="plan-po-occupation-master-empty"></a>

### 8. PO-OCCUPATION-MASTER-EMPTY

**推奨案**: まず0件時に「職種が未登録」であることを説明し、権限のある利用者を既存の職種マスタ登録へ案内する。職種未登録を理由に予約やスタッフ保存へ新しい必須制約を加えない。運用必須と決まった場合だけ、医院別の初期登録手順を整備する。

1. POが職種の必須性、正式名称、医院別差異、管理責任者を決める。初期「獣医師」「動物看護師」は候補であり、この計画だけで投入しない。
2. [use-staff-settings-lookups.ts](frontend/src/features/master/hooks/use-staff-settings-lookups.ts)、[staff-settings-model.ts](frontend/src/features/master/routes/staff-settings-model.ts) と [occupation_repository.go](backend/internal/staff/occupation_repository.go) を確認する。職種APIの実装は `staff` ドメインにあり、現行の `occupation_id` は任意。スタッフの既存職種値が無効の場合も、編集画面の表示用に保持できるか確認する。
3. 必須化する場合は、既存マスタ登録の仕組みで同一医院の重複登録を防ぐ手順を作る。既存行・編集済み名称を上書きせず、追加だけをdry-runで示す。コードで架空の選択肢を生成しない。
4. 0件・有効職種あり・無効職種のみ・登録権限なし・他院の職種指定をテストする。出勤医師判定のスタッフ種別と、予約区分に職種を紐付けた場合の判定を混同しない。

**実装メモ (2026-09-13 / att-po-occupation-empty-20260913-001)**: `StaffBasicInfoSection` で `allOccupations.length === 0` のとき「職種が未登録です」＋ `paths.settings.occupations.getHref()` へのリンクを表示し、空 Select を出さない。有効職種があるときは従来どおり active-only Select。選択中の無効職種は同一医院マスタ内の既存行だけ表示保持（架空ラベルなし）。`buildStaffCreateRequest` / `buildStaffUpdateRequest` の `occupation_id: data.jobTitleId ?? undefined` は変更なし（任意）。自動シードなし。検証: `docker compose exec frontend npx vitest run src/features/master/components/StaffBasicInfoSection.test.tsx src/features/master/routes/staff-settings-model.test.ts`（13 passed）。

**完了条件**: 職種0件でも理由と次の操作が分かる。承認された医院だけに重複なく初期登録でき、他院の職種を割り当てられない。職種を任意のまま運用する場合は、その決定を記録して項目の扱いを確定する。 → **達成（FIXED・任意運用＋空案内。医院別初期登録は別途）**

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

<a id="plan-bug-acct-close-perm-default"></a>

### 11. BUG-ACCT-CLOSE-PERM-DEFAULT

**方針**: まず設計意図を裁定する（締め実行に必要なスコープは `create` か `edit` か）。既定モデルが exec へ機能しない `edit` を付与し `create` を全グループから外している現状は、意図的 fail-closed にしても内部矛盾。

1. `POST /cash-register/closes` の要求スコープと `permission_group_rules.csv` / `defaultPermissionRuleTable` の既定付与を突合し、(a) 既定へ create 追加・(b) endpoint を edit 要求へ変更・(c) fail-closed 維持で dead edit 除去+運用明記、のいずれかを決める。
2. (a)/(b) を選ぶ場合、既存 clinic の権限グループへの移行影響（既に締め運用中の環境で権限が落ちないこと）を確認し、必要なら移行 seed を用意する。
3. 修正後、既定権限のままの新規 clinic で執行アカウントが締めを実行できること（または意図的拒否が文書どおりであること）を UAT レベルで再検証する。

**完了条件**: 既定権限モデルで締め実行の可否が設計意図と一致し、dead grant が解消されていること。S08 手順10 の再実行で締め→締め後訂正→監査記録（post_close フラグ）まで既定プロビジョニングのみで完走できること。

<a id="plan-bug-s09-fixture-teardown"></a>

### 12. BUG-S09-FIXTURE-TEARDOWN

**方針**: append-only 監査台帳（W-013）は製品正として維持し、fixture 側の破棄設計を直す。immutable テーブルの行を消す実装は採らない。

1. `DeleteSyntheticClosingFixture` の方針を裁定: (a) closes/adjustments を除く全行削除後に clinic を `is_active=false` で残置する「論理破棄」へ変更、(b) DB 全体リセット前提の運用に変えて teardown を締め済み clinic で失敗する旨を明示、のどちらかが安全。trigger 無効化による物理削除は監査設計を壊すため非推奨。
2. 選んだ方針で teardown を修正し、「締め→締め後編集→teardown」の順で回収が成立することを隔離 DB で検証する。
3. `clinic_holidays` など scoped 削除から抜けている他の fixture 生成物も棚卸しして対象に含める。

**完了条件**: S09 手順を全実行した合成 clinic に対して teardown がエラーなく完走する（または意図的残置が文書化され clinic 非活性化まで行われる）こと。`cash_register_closes`/`cash_register_close_adjustments` の append-only 制約を迂回しないこと。

<a id="plan-bug-agg-no-visit-revenue"></a>

### 13. BUG-AGG-NO-VISIT-REVENUE

**方針**: `include_no_visit` 除外は「最終来院」軸の UI 選択肢（来院なしを含む）を制御するものであり、売上ランキング・来院回数クエリへ適用するのは意味論的に誤り。`filterLTVRows` で no_visit 除外を `LastVisitBucket` 指定クエリに限定する。

1. `filterLTVRows`（`backend/internal/owner/ltv_repository_query.go`）の no_visit 除外を、`params.LastVisitBucket != ""` または last_visit 系ソート指定時に限定するよう修正。売上/来院軸では除外しない。
2. 回帰テスト: 完了会計を持つ no_visit 飼主が売上ランキング (`year` + `amount_basis` 指定) に出ること、`last_visit_bucket=over_1y` + `include_no_visit=false` では従来どおり除外されることを table-driven test で固定。
3. 仕様正本 36 との整合（「全顧客を対象とした売上貢献度」）を再確認。CPM ステージ集計への影響（no_visit 飼主の CPM 分類）も別途確認。

**完了条件**: `medical_record_id` なしの完了会計のみを持つ飼主が、売上ランキングで `annual_amount` 付きで表示されること。最終来院タブの「来院なしを含む」チェック動作が従来どおり機能すること。

<a id="plan-bug-trim-kanban-in-consultation"></a>

### 14. BUG-TRIM-KANBAN-IN-CONSULTATION

**方針**: 「in_consultation には medical_record が必要」というガードを trimming appointment に適用しない設計へ裁定する。候補は (a) `validateInConsultationHasMedicalRecord` で `reservation_types.category = 'trimming'` を免除（対象は `appointment_trimming_details` の存在を確認するか単に免除）、(b) FE `use-trimming-form` が既存 appointment への detail 作成 POST に `status: in_consultation` を載せる（BE `createTrimmingDetailForExistingInTx` は既に honor する）、(c) カンバン側で trimming は受付済→会計待ちの直行遷移を正式化。カルテ必須ガード自体は medical category で維持する。

1. 遷移契約を裁定: trimming の `checked_in → in_consultation` は detail 作成をもって成立させるか（§2.4/§5.2-G）、不要とするか。仕様正本 `reservation-to-record-flow.md` と S11 記述を更新対象に含める。
2. 選んだ契約で実装: (a) なら validator に category 分岐を追加し `appointment_trimming_details` 存在を要件化、(b) なら `buildCreateTrimmingRequest` に status 送信を追加。両方やる場合は冪等にする。
3. 回帰: 受付済 trimming カード→「トリミングカルテ作成」→保存で appointment が `in_consultation` になること、medical appointment のカルテなし遷移が従来どおり 409 で拒否されることを固定する。

**完了条件**: 受付済トリミングカードが UI 操作のみで「診療中→会計待ち」へ進めること。S11 手順2/4 を迂回なしで完走できること。

<a id="plan-bug-billing-unbilled-mr-exclusion"></a>

### 15. BUG-BILLING-UNBILLED-MR-EXCLUSION

**方針**: `FindUnbilledByPetID` の L115（`b.medical_record_id = mr.id AND b.status != 'cancelled'` の billing 単位除外）を見直す。「1 カルテ 1 会計」の意図なら明細削除ではなく会計 cancel を正規解放経路として明文化するか、明細単位除外（L114 のみ）へ変更して削除明細の請求元を再候補化するかを裁定する。

1. 仕様裁定: pending 会計からの明細削除を「その請求元の請求放棄」とみなすか「再候補化」とみなすか。S11 A3 の記述は再候補化を前提。
2. 再候補化を採る場合: L115 を削除または「billing に当該 MR の明細が1件でも残る場合のみ除外」へ緩める。trimming/vaccination 側の unbilled クエリも同じ二段除外があるか棚卸しする。
3. 回帰: MR 連携 pending 会計→明細 soft-delete→unbilled-details 復帰、billing cancel→復帰、精算済み明細削除→409 の3ケースを固定。

**完了条件**: 明細削除で請求元が未請求候補に復帰し、かつ精算済み会計の明細削除が引き続き拒否されること。

<a id="plan-bug-res-overlap-500"></a>

### 16. BUG-RES-OVERLAP-500

**方針**: `excl_appointments_doctor_timerange`（SQLSTATE 23P01 exclusion_violation）を `response_pg.go` のエラーマッピングで 409 Conflict + 既存の重複メッセージへ変換する。`CheckSlotConflict` の事前チェックは維持し、DB 制約到達時も同じ契約で返す。

1. `response_pg.go`（または apperrors.FromGORM の PG code 変換）に 23P01 → Conflict のマッピングを追加。制約名で「時間帯重複」メッセージを選択する。
2. 回帰テスト: 同一スタッフ部分重複 POST → 409、完全一致 → 409、別スタッフ同一時刻 → 201 を固定。
3. LIFF / バッチ作成経路でも同じ制約違反が 409 になることを確認。

**完了条件**: 時間帯重複が常に 409 + 業務メッセージで返り、500 + DB 制約名の漏出がなくなること。

<a id="plan-bug-liff-healthcard-owner-sync"></a>

### 17. BUG-LIFF-HEALTHCARD-OWNER-SYNC

**方針**: 「連携済み」の意味を一本化する。`owners.line_user_id`（L-step/通知向け）と `line_customers.owner_id`（health-card/予約顧客解決）の2系統を、検証済み LIFF トークン連携の成立時に同期するか、スタッフ手動 link-owner を必須の第2段として仕様・シナリオ・画面文言に明記するかを裁定する。

1. 仕様裁定: 医療情報開示前のスタッフ確認を意図的ゲートとするか。S12 は「連携完了→ヘルスカード閲覧」を一段で記述。
2. 同期を採る場合: `LinkAccount` 内で `line_customers` を `line_user_id` で FindOrCreate し `owner_id` を同一 Tx で更新。clinic_id 一致・既存 owner_id との競合（別飼主に紐付済み）の扱いを定義。解除時の逆同期も定義する。
3. 手動2段を採る場合: 04-owners-form / 38-liff-pet-health / S12 シナリオに「LIFF 連携後はスタッフが LINE 顧客の飼主紐付けを別途実施」と明記し、FE の連携済み表示に未紐付け顧客の注意を出す。
4. 回帰: 連携成立→ health-card が飼主ペットを返すこと、別飼主データが混ざらないこと、連携解除で再び空に戻ることを固定。

**完了条件**: LIFF トークン連携完了後にヘルスカードが当該飼主のペットを返すか、あるいは2段運用が仕様正本に明文化され画面が矛盾しないこと。

<a id="plan-bug-acct-ins-sign-mismatch"></a>

### 18. BUG-ACCT-INS-SIGN-MISMATCH

**方針**: `insurance_amount`（および同じ規約リスクを持つ `discount_amount`）の符号規約を FE/BE で一本化する。BE complete 経路（`billing = total − insurance − discount`）は正値前提が自然なため、FE 送信側を正値化（送信時に符号反転または絶対値）し、`PaymentInfo.insuranceAmount` の画面内表示契約（負値）とは分離するのが第一候補。BE を負値前提に変える案は既存の正値データ・update 経路との整合確認が必要。

1. API 契約を確定: `POST /accountings/complete` の `insurance_amount` を正値（保険負担額の絶対値）と明文化し、負値は 400 で拒否するか双方受理するか決める。OpenAPI/型生成に反映。
2. FE `use-accounting-completion-action.ts` の complete payload で `insurance_amount` を正値で送信（画面表示の負値との変換を境界で明示）。`has_insurance=false` 時は従来どおり null。
3. 回帰: 保険 50%/70% 付き会計の complete が 201 で `billing_amount = total − insurance` となること、保険なし会計が従来どおり成立すること、二重送信防止（Idempotency-Key）が維持されることを固定。`discount_amount` を送る経路があれば同じ検証を行う。

**完了条件**: UI から保険 50%/70% の新規会計が確定でき、API の `billing_amount`・`payments.insurance_amount` が画面内訳と一致すること。S15 手順7 を迂回なしで完走できること。

<a id="plan-bug-acct-ins-edit-rewrite"></a>

### 19. BUG-ACCT-INS-EDIT-REWRITE

**方針**: 確定済み会計の修正保存で「変更した項目だけが変わる」を保証する。`hasInsurance` を `insurance_amount < 0` 推定ではなく保存された `has_insurance` フラグ（または `insurance_ratio` の存在）から初期化し、編集保存ではユーザーが保険・明細を変更していない限り保存済みの `insurance_amount`/`billing_amount` を維持する（差分送信または変更検知ゲート）。

1. `use-accounting-detail-state.ts:197` の初期化を `has_insurance` フラグ優先に修正し、`insurance_amount` 非負・ratio のみ・金額不整合の各保存状態でスイッチ表示が保存値と一致することを vitest で固定。
2. `use-accounting-completion-action.ts` の update 経路で、保険関連フィールド（ratio/switch/明細の is_insurance_applicable）が未変更なら `insurance_amount`/`billing_amount` を送信しないか保存値を送る契約へ変更。変更ありの場合のみ recalc 値を送る。
3. 回帰: レガシー 90% 会計で理由のみ変更→保存→再読込で ratio/金額/保険表示が不変、明細変更→正しく recalc、正値保存データ（BE 規約）の再表示で保険 ON・保存金額が表示されることを固定。手順6/7 の再実行で確認。

**完了条件**: 確定済み会計の無変更保存で `payments` の金額・割合が保存値のまま維持され、再読込表示が保存値と一致すること。`has_insurance=true` と保険スイッチ表示が乖離しないこと。

<a id="plan-bug-dialog-focus-restore"></a>

### 20. BUG-DIALOG-FOCUS-RESTORE

**方針**: `TreatmentSearchDialog`（および同パターンで外部 `open` 制御・`lazy` 化された他ダイアログ）で、閉じた後のフォーカスが呼出元トリガーへ戻ることを保証する。Radix の既定復帰が効かない経路をまず特定する。

1. `TreatmentsTab` のトリガーが `DialogTrigger` ではなく外部 state ボタンである点と、`lazy` 遅延マウントで Radix FocusScope が復帰対象（開く直前のフォーカス要素）を捕捉できていない点を確認する。`DialogTrigger` 構成へ移行するか、`onCloseAutoFocus` でトリガー ref へ明示復帰するかを裁定する。
2. 選んだ方式で実装し、vitest（実 Dialog）で「開く → Escape → activeElement がトリガー」を固定。同一ダイアログで項目選択による閉鎖でも復帰先が妥当（現行は追加行の数量入力へ移る継続編集仕様）であることを確認する。
3. 同一パターン（外部 `open` 制御＋非 DialogTrigger）の他ダイアログを棚卸しし、共通 `Dialog` の修正で済むか個別対応かを決める。
4. （任意・別裁定）長い一覧の矢印キー移動: listbox/roving tabindex 導入の要否を PO/デザインと決める。Tab 順送りは現状の正式経路として維持。

**完了条件**: 実ブラウザで「ボタンで開く → Escape/閉じる → `document.activeElement` が呼出元ボタン」となること。S18 手順6 を迂回なしで完走できること。

### 21. BUG-BILLING-TAX-TYPE-DROPPED

**方針**: マスタの税区分・税率を会計明細へスナップショットとして引き継ぐ。確定済み会計は遡って再計算しない（S20 手順7 で確認済みの不変性を維持）。

1. BE: `billing_item_unbilled.go` の `treatmentToUnbilledBillingItem`・vaccination/exam/trimming 候補生成が、マスタの `tax_type`/`tax_rate` を参照するよう SQL と item 組立てを修正。マスタに税区分を持たない source（vaccine 等）は既定値のままとし、欠落時のフォールバック契約をテストで固定。
2. FE: `get-merchandise-items.ts` の transform に `tax_type` を含め、`use-accounting-item-actions.ts` の `handleAddItem` がマスタ値を `tax_type`/`tax_rate` として送信するよう修正。手入力項目は現行どおり外税10%既定で可。
3. 回帰: 「内税/非課税マスタ → 未請求候補/物販追加 → 明細行の課税区分・税額・合計」が正しいこと、および確定済み会計の明細が変わらないことを FE/BE テストと DB 値で固定。

**完了条件**: S20 手順5 を再実行し、内税・非課税項目が正しい税区分で明細化され、合計が手計算と一致すること。既存確定会計の金額が変わらないこと。

### 22. BUG-ACCT-DUP-COMPLETE-500

**方針**: `POST /accountings/complete` の UNIQUE 競合分岐が、abort 済みトランザクション上で後続クエリを実行しない構造にする。

1. `createCompleteBillingHeader`（`accounting_complete_tx.go`）で `repo.Create` が UNIQUE 競合（23505）を返した場合、`FindByCompletionRequestID` を同じ `txCtx` で呼ぶと Postgres は 25P02 で全コマンドを拒否する。競合解決クエリを（a）トランザクション外の別接続で実行、（b）エラーの制約名を見て `completion_request_id` 衝突のみ tx 外 replay に回す、（c）`ON CONFLICT DO NOTHING`＋別 tx 読取、のいずれかに修正する。
2. `medical_record_id` 衝突（別キーの同一カルテ二重会計）→ 409「このカルテには既に会計があります」、`completion_request_id` 衝突（同キー並行リクエスト）→ digest 一致なら replay 200 / 不一致なら 409、の分岐を実 DB テストで固定。
3. 回帰: 並行2リクエスト（同キー同 payload→両方 200/同一 billing、同キー異 payload→片方 409、別キー同一 MR→勝者 201・敗者 409）を実 DB の integration test または UAT 手順で固定。

**完了条件**: S21 手順4・異常系A2 を再実行し、二重会計試行が 409 で拒否されること（500 にならない）。既存行が壊れないこと。

### BUG-AGG-NO-VISIT-REVENUE: 来院なし飼主が完了会計を持っても売上ランキングに表示されない

- **発見経緯**: 2026-09-22 UAT S10 手順7（見積のみの飼主 B を確認）の拡張検証。飼主 B（UAT集計 飼主B, id=1000000002）に `medical_record_id` なしの完了会計 ¥3,300（billing 1000000012）を作成したところ、売上ランキングの全フィルタ組合せ（`include_zero` 真偽両方）で表示されない。
- **現象**: `GET /api/v1/clinics/1/owners/aggregations?search=UAT集計 飼主B&include_zero=true` → `owners=[]`。`include_no_visit=true` を付けた場合のみ `annual_amount=3300` で出現。
- **原因（切り分け済み）**: `filterLTVRows`（`backend/internal/owner/ltv_repository_query.go`）で `!params.IncludeNoVisit && *row.LastVisitBucket == "no_visit"` の行を無条件除外。来院履歴のない飼主は `vs.last_visit_date IS NULL` → bucket=`no_visit` となり、完了会計を持っていても除外される。`include_no_visit` は最終来院タブの「来院なしを含む」用パラメータだが、売上・来院回数クエリにも同じ既定除外が適用されている。FE の売上タブ既定 params（`aggregation-dashboard-model.ts`）には `include_no_visit` がなく、UI から回避不能。
- **影響**: 物販のみ等の「来院記録なし・売上あり」顧客が売上ランキングから恒久的に脱落し、年間診療費の一覧合計が実績を過小計上する。仕様正本 36 は「クリニックの全顧客を対象とした売上貢献度（LTV）」と定義しており矛盾。
- **再現**: 飼主作成 → `POST /accountings/complete`（`medical_record_id` なし、`owner_id` 指定）→ 集計 API で `search` 一致しても `include_no_visit=true` なしでは 0 件。
- **関連**: `reports/uat-2026-09-22/S10-customer-aggregation-consistency.md`。`todo.md#product-bugs` 正本にも登録

### BUG-TRIM-KANBAN-IN-CONSULTATION: 受付済トリミングカードが「診療中」へ遷移できず受付済に滞留する

- **発見経緯**: 2026-09-22→23 UAT S11 手順2/4。受付済（`checked_in`）トリミング appointment のカードから「トリミングカルテ作成」を押下。
- **現象**: ボタンには「※カルテ作成と同時に『診療中』へ移動します」と明記されるが、実際には `PATCH /api/v1/reservations/:id {status: in_consultation}` が **409** `診療を開始するにはカルテが必要です` で拒否され、カードは受付済のまま。トリミング記録の保存（`appointment_trimming_details`）を行っても status は変わらない。backend ログで 409 を2回確認（request_id `167894b8`, `3970757a`、対象 appointment 1000000006）。
- **原因（切り分け済み）**:
  - `validateInConsultationHasMedicalRecord`（`backend/internal/reservation/reservation_service_validate.go:286`）は status→`in_consultation` 遷移時に `medical_records` の COUNT>0 を**予約区分を問わず**要求する。トリミング appointment は `appointment_trimming_details` を持ち `medical_records` を持たないため、件数 0 で常に Conflict。
  - FE 経路も断線: カードの `onConfirm`（`ReceptionDialogActionButtons.tsx` 受付済分岐）が汎用 PATCH で in_consultation を試みて 409。一方、trimming 作成 POST 経路（`createTrimmingDetailForExistingInTx`）は `input.Status` を honor するが、FE の `use-trimming-form.ts` は `hasExistingAppointment` 時に `status` を送信しない（`record_shortcut` 新規のみ送信）ため detail 作成も遷移を起こさない。
  - 結果: 受付済 → 診療中 の正規遷移経路がトリミングには存在しない。ドラッグは `受付済→診療中` 直行を禁止し、ボタン列は `受付済→診療中` のみ（`NEXT_COLUMN_TITLE`）。**迂回**: `PATCH status=accounting`（受付済→会計待ちの飛び級）は 200 で通る（本 UAT で使用）が、in_consultation を経ない飛び級は状態機械の想定外。
- **影響**: トリミングの通常運用（受付→施術→会計待ち）が UI ボタン操作では完走しない。S11 の「カルテ作成が診療中への契機」および手順4「明示的な完了操作で会計待ちへ」の仕様記述と矛盾。
- **関連**: `reports/uat-2026-09-23/S11-trimming-combined-accounting.md`。`todo.md#product-bugs` 正本にも登録

### BUG-BILLING-UNBILLED-MR-EXCLUSION: カルテ連携会計から明細削除しても請求元 treatment が未請求候補に復帰しない

- **発見経緯**: 2026-09-23 UAT S11 異常系 A3（未精算の統合会計から一方の明細を削除→再候補化を確認）。
- **現象**: `medical_record_id` 付き pending 会計（billing 1000000016）に treatment_id=3 の明細を追加後、`DELETE /api/v1/billing-items/:id` で soft-delete（204）しても、`GET /billing-items/unbilled-details?pet_id=` がその treatment を返さない（0件）。会計を cancel（status=cancelled）すると復帰することを確認。
- **原因（切り分け済み）**: `FindUnbilledByPetID`（`backend/internal/medicalrecord/treatment_repository.go:114-115`）の除外条件が2段ある。
  - L114: `billing_items` 行単位（`bi.treatment_id = treatments.id AND bi.deleted_at IS NULL`）— soft-delete された明細はここを通過する
  - L115: `billings` 単位（`b.medical_record_id = mr.id AND b.status != 'cancelled'`）— **会計が `medical_record_id` を保持する限り、そのカルテの全 treatment が明細の有無に関係なく除外される**
  - 明細削除後も billing 行は残るため L115 が効き続け、treatment は未請求候補に復帰しない。
- **影響**: S11 異常系 A3 の仕様記述「明細削除は soft-delete。未請求クエリは deleted_at IS NULL のため削除行は再候補になる」と矛盾。pending/統合会計から明細を外すと請求元がサイレントに未請求一覧から消え、会計全体を cancel しない限り再請求できない（請求漏れリスク）。
- **補足**: `trimming` 明細側（`FindUnbilledTrimmingItemsByPetID`）は `appointment_id` 紐付けのため別経路。手入力明細（マスタ参照なし）には影響しない。
- **関連**: `reports/uat-2026-09-23/S11-trimming-combined-accounting.md`。`todo.md#product-bugs` 正本にも登録

### BUG-RES-OVERLAP-500: 時間帯重複の予約作成が排他制約 500 で漏れる（完全一致は 409）

- **発見経緯**: 2026-09-23 UAT S11 fixture 作成中、同一スタッフ（UAT担当医）に時間帯が一部重複する予約を POST。
- **現象**: 完全一致の同一時刻予約は `POST /reservations` → **409** `同じ時間帯に既に予約があります`（正しい）。一方、部分重複（例: 10:30-11:30 vs 既存 10:00-11:00）は PostgreSQL 排他制約 `excl_appointments_doctor_timerange` の SQLSTATE 23P01 が素通りして **500** `ERROR: conflicting key value violates exclusion constraint` が返る。
- **原因（切り分け済み）**: `CheckSlotConflict` の事前チェックは完全一致を捕捉するが、レースまたは部分重複パターンで DB 排他制約に到達し、`response_pg.go` のエラーマッピングに 23P01（exclusion violation）の conflict 変換がない。
- **影響**: クライアントには区別不能な 500 で返り、再試行・メッセージ表示の分岐を阻害。予約 UI からは通常発生しないが、並行予約・LIFF 同時予約では起こりうる。
- **関連**: `reports/uat-2026-09-23/S11-trimming-combined-accounting.md`。`todo.md#product-bugs` 正本にも登録

### BUG-LIFF-HEALTHCARD-OWNER-SYNC: LIFF 連携が成立しても `line_customers.owner_id` が更新されずヘルスカードが空のまま

- **発見経緯**: 2026-09-23 UAT S12。LIFF モック連携成功後の health-card 解決経路をコード突合。
- **現象**: `owners.line_user_id` が設定された（=院内 UI「連携済み」）状態で `line_customers.owner_id` が NULL のままだと、`GET /api/liff/:clinicId/health-card` は LINE 表示名 + `pets:[]` を返し、画面は「ペット情報はありません」。実測: `owners.line_user_id='mock-line-user-id'` + `line_customers.owner_id=NULL` で「テストユーザー / ペット情報はありません」、その後 `PATCH /line-customers/1/link-owner` で owner 紐付けすると同じ LINE 顧客のヘルスカードに飼主名+ペット+ワクチン記録が表示。
- **原因（切り分け済み）**:
  - `LinkAccount`（`line_link_service.go:259-326`）の Tx は `owners.line_user_id` 更新・token consume・audit のみ。`line_customers` への FindOrCreate / `owner_id` 書込みが存在しない。
  - health-card 解決（`liff_service_health_card.go:34-47`）は `line_customers.owner_id` → `Owner.Pets` のみを見る。`owners.line_user_id` からの逆引きはない。
  - `line_customers.owner_id` の唯一の書き手は `UpdateOwnerLink`（`line_customer_repository.go:120`、スタッフ手動 `PATCH /line-customers/:id/link-owner` 経由）。
  - DB トリガ・バッチ同期も存在しない（pg_trigger 確認済み）。
- **影響**: QR/URL 連携を完了した飼主全員がヘルスカードで「ペット情報はありません」を見る。院内は「連携済み」と表示され、誰も第二段の手動紐付けが必要だと気付かない。S12 シナリオ目的そのものが成立しない。
- **注意**: SEC-CS2-F02 で name+phone 自動紐付けは意図的に廃止済みだが、本件は検証済み単回トークン経由の明示的連携であり同一の脅威モデルではない。意図的な2段ゲートの可能性もあるため裁定要。
- **関連**: `reports/uat-2026-09-23/S12-liff-pet-health.md`。`todo.md#product-bugs` 正本にも登録

### BUG-ACCT-INS-SIGN-MISMATCH: 保険適用会計の新規確定が FE/BE 符号規約不整合で必ず 400

- **発見経緯**: 2026-09-23 UAT S15 手順7（新規会計を 50% で保存）。UI で保険 ON・支払 ¥1,200 を入力して「会計を確定する」を押下 → `支払い内訳の合計（1200）が請求金額（3200）と一致しません`。
- **現象**: 保険 ON の新規会計が UI から一切確定できない。画面の請求額表示は正しい（¥1,200）のに、BE が請求額を ¥3,200 と計算し支払内訳一致検証で 400。
- **原因（切り分け済み）**: `insurance_amount` の符号規約が FE/BE で不一致。
  - FE: `calculations.ts:112` `Math.floor(target * ratio) * -1` で負値を生成し、`use-accounting-completion-action.ts:271-300` がそのまま `insurance_amount` に送信。型注釈も「保険負担額（マイナスのみ）」（`types/index.ts:61`）。
  - BE: `accounting_complete_tx.go:211` `billingAmount = totalAmount - insuranceAmount - discountAmount` は**正値**の保険負担額を前提。負値 −1000 を渡すと `2200 − (−1000) = 3200`。
  - 直 API 検証: `insurance_amount:-1000` → 400（請求額3200と不一致）、`+1000` → 201（billing 1000000027 作成）。FE 送信経路は常に負値のため UI からは確定不能。
- **影響**: 保険窓口精算が新規会計で実質使用不能。画面上の内訳は正しいのに確定だけが失敗するため、利用者には理由が分からない。
- **関連**: `reports/uat-2026-09-23/S15-insurance-rate-preservation.md`。`todo.md#product-bugs` 正本にも登録

### BUG-ACCT-INS-EDIT-REWRITE: 確定済み会計の無変更保存で保険金額・割合表示が保持されない

- **発見経緯**: 2026-09-23 UAT S15 手順6（レガシー 90% 会計で割合以外の変更のみ保存）および手順7（保存→再読込の保持確認）。
- **現象（2つの失敗モード）**:
  - (a) 保存値が現在の recalc と異なるレガシー会計（`insurance_amount=-220`, `billing_amount=1980`, ratio 0.9、保険適用明細なし）で、修正理由のみ入力して保存 → `insurance_amount` が **0**、`billing_amount` が **2200** にサイレント上書き。再読込後は `has_insurance=true`・`insurance_ratio=0.9` が残るのに保険スイッチ OFF・割合非表示になり、実質保険が表示から消える。
  - (b) BE 規約（正値）で保存された会計（billing 1000000027: `insurance_amount=+1000`, ratio 0.5, `billing_amount=1200`）を詳細で開くと、保険スイッチ OFF・保険負担額行なし・請求額が recalc の ¥2,200 表示・「残り ¥1,000 未入力」の幻影未収が出る。保存時の金額と再読込表示が一致しない。
- **原因（切り分け済み）**:
  - `use-accounting-detail-state.ts:197` が `hasInsurance = (insurance_amount ?? 0) < 0` で推定しており、保存された `has_insurance` フラグ・`insurance_ratio` を見ない。正値（BE 規約）や 0 の保存値は常に OFF 表示。
  - `use-accounting-completion-action.ts:318-321` の編集保存（`updateAccounting`）は `insurance_amount`/`billing_amount` を現在明細からの recalc 値で無条件送信するため、ユーザーが保険・金額に触れなくても保存値が上書きされる（`insurance_amount` が 0 のときは `null` 送信）。
  - 結果として `has_insurance=true` + `insurance_ratio=0.9` + `insurance_amount=0` の内部矛盾レコードが生成される。
- **影響**: 「変更した項目だけが変わる」契約違反。レガシー移行データの金額が無関係な修正で静かに書き換わり、保険適用の表示自体も消失する。S15 手順6/7 の期待結果に合致しない。
- **関連**: `reports/uat-2026-09-23/S15-insurance-rate-preservation.md`。`todo.md#product-bugs` 正本にも登録

### BUG-DIALOG-FOCUS-RESTORE: 治療プラン検索ダイアログを閉じるとフォーカスが呼出元へ戻らず body に落下する

- **発見経緯**: 2026-09-23 UAT S18 手順6（Escape/閉じるで閉じた後の focus restoration 確認）。
- **現象**: カルテ治療タブの「マスタから追加」で `TreatmentSearchDialog` を開き、Escape で閉じると `document.activeElement` が `document.body` になる。キーボード操作の利用者は位置を失い、次の Tab は文書先頭からやり直しになる。
- **実測（2026-09-23・Chromium 系実ブラウザ・1366×625）**:
  - 実クリックで開く → Escape → `activeElement = BODY`（呼出元ボタンでない）
  - 呼出元ボタンを `focus()` 済みの状態で Enter キー押下により開く → Escape → 同じく `BODY`（キーボード経路でも復帰しない）
  - ダイアログは `lazy()` インポートの `TreatmentSearchDialog`（`TreatmentsTab.tsx:7-9`）で、トリガーは Radix `DialogTrigger` ではなく外部 `open` state で制御する通常ボタン
- **原因（切り分け）**: `ui/dialog.tsx` の `DialogContent` は `onCloseAutoFocus` を上書きしていないため Radix 既定の focus restore が期待されるが、実測では復帰しない。Radix の復帰先は「ダイアログが開く直前にフォーカスを持っていた要素」だが、`lazy` + 外部 `open` 制御の組合せで FocusScope が復帰対象を記録できていない可能性。`DialogTrigger` 経由でないため trigger 参照も存在しない。
- **関連するキーボード操作ギャップ（同一ダイアログ）**: 一覧項目はネイティブ `<button>` のみで ArrowUp/ArrowDown の roving や listbox 意味論がない。Tab で全項目を順送りする設計のため、長い一覧（実測 scrollHeight ≈ 389,000px）ではキーボードだけでの到達が事実上困難。検索絞込みとの併用が事実上の前提になる。
- **影響**: キーボードのみの利用者が項目選択/キャンセル後に画面位置を失う。S18 手順6 の「フォーカスが呼び出し元へ戻る（focus restoration）」に合致しない。マウス利用では影響なし。
- **関連**: `reports/uat-2026-09-23/S18-treatment-search-dialog-height.md`。`todo.md#product-bugs` 正本にも登録

### BUG-BILLING-TAX-TYPE-DROPPED: マスタの税区分（内税/非課税）が会計明細へ伝播せず全て外税10%で請求される

- **発見経緯**: 2026-09-23 UAT S20 手順5（税区分バリエーションの会計伝播確認）。
- **現象**: マスタで `tax_type=included`（内税）または `exempt`（非課税）として登録した項目が、会計明細では全て `外税 10%` として計算・保存される。
- **実測（2026-09-23・clinic 1・billing 1000000028）**:
  - マスタ登録: 処置 `S20 処置内税`（included, ¥1,100）、`S20 処置非課税`（exempt, ¥1,000）、物販 `S20 商品内税`（included, ¥1,100）、`S20 商品非課税`（exempt, ¥1,000）
  - 未請求候補（カルテ連携）の表示: 全行「外税 10%」。物販・その他追加ピッカーからの追加も全行「外税 10%」
  - DB `billing_items`（id 1000000025–1000000031）: 全行 `tax_type=excluded, tax_rate=0.10`（source=medical_record / manual 双方）
  - 期待合計（税区分が正しく伝播する場合）: ¥5,850 ／ 実際の請求: ¥6,270（消費税の過剰計上 ¥420）
- **原因（切り分け）**:
  - カルテ連携経路: `backend/internal/billing/billing_item_unbilled.go` の `treatmentToUnbilledBillingItem`（同ファイル内 vaccination/exam/trimming も同様）が `TaxType: model.TaxTypeExcluded, TaxRate: sharedkernel.DefaultTaxRate` をハードコードし、マスタの tax_type/tax_rate を参照していない
  - 物販経路: `frontend/src/features/accounting/api/get-merchandise-items.ts` の transform が `tax_type` を落とし、`use-accounting-item-actions.ts` の `handleAddItem` が `tax_type: "excluded"` を固定送信する
- **影響**: 非課税・内税品目を正しく登録しても会計で外税10%が課され、患者側への請求税額が実際より多くなる。会計明細行の課税区分は編集可能だが、修正を手動で行わない限り誤請求のまま確定する。
- **関連**: `reports/uat-2026-09-23/S20-master-to-accounting-path.md`。`todo.md#product-bugs` 正本にも登録

### BUG-ACCT-DUP-COMPLETE-500: 同一カルテへの二重会計確定が意図した 409 に届かず 500 で漏れる

- **発見経緯**: 2026-09-23 UAT S21 手順4（別 Idempotency-Key・同一 medical_record_id で complete 再送）。
- **現象**: 会計確定済みのカルテへ別キーで再度 complete を送ると、設計どおりの 409「このカルテには既に会計があります」ではなく `500 internal server error` が返る。真並行（別キー2リクエスト同時送信）でも敗者が同じく 500。
- **実測（2026-09-23・clinic 1）**:
  - 逐次: MR 1000000011 に billing 1000000030 確定後、別キー `…5502` で complete → `{"error":"internal server error"}` 500
  - 並行: MR 1000000012 へ別キー2本同時送信 → 勝者 201（billing 1000000032）・敗者 500
  - データ保全: 二重会計は作られない（`idx_billings_medical_record_id_unique` で INSERT 拒否・tx rollback）。レスポンス契約のみ破損
  - サーバーログ: `duplicate key value violates unique constraint "idx_billings_medical_record_id_unique" (SQLSTATE 23505)` → 直後の replay 判定クエリが `current transaction is aborted (SQLSTATE 25P02)` で失敗し `"failed to resolve completion unique conflict"` として 500 化
- **原因（切り分け済み）**: `accounting_complete_tx.go` の `createCompleteBillingHeader` が `repo.Create` の UNIQUE 失敗後に `FindByCompletionRequestID` を**同じ abort 済みトランザクション上**で実行する。Postgres は tx 内エラー後の全コマンドを拒否するため、replay/409 分岐（`accounting_complete_tx.go:149-156`）は実環境では到達不能な dead path。tx 前の冪等 lookup（`accounting_complete.go:199`）だけが機能しており、UNIQUE 側の競合は全て 500 に落ちる。
- **影響**: 「このカルテには既に会計があります」409 の契約が成立しない。クライアントは原因を判別できず、再送しても同じ 500。並行確定（2人のスタッフが同じカルテを同時確定等）でも敗者に意味不明な 500 が返る。監査・復旧上の誤解を招く。
- **関連**: `reports/uat-2026-09-23/S21-accounting-concurrency-idempotency.md`。`UAT-R2-EXCLUSIVE-LOCK` の「異なるkeyの同一カルテ会計を実DBで確認」項目の実測結果として接続。`todo.md#product-bugs` 正本にも登録

## バグではない（記録のみ）

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
