# 一時バグ／障害メモ（ユーザー依頼 2026-09-13）

> Task migrated to Plane `EMR-180`. This file remains supporting acceptance/evidence material; use Plane for current status.

> **注意**: root `bug.md` は 2026-09-08 に `todo.md` の [製品 FAIL](todo.md#product-bugs) へ統合廃止済み。  
> 本ファイルはユーザー明示依頼により **一旦** 再作成したメモ台帳。製品 FAIL 正本は引き続き `todo.md#product-bugs`。  
> `BUG-LOCAL-HANDOFF-CSV-CONTRACT` は環境／handoff fixture の BLOCKED。  
> OPEN 項目は2026-09-23にPlaneへ移行済み。以下の調査結果・実証データは履歴証拠として保持し、現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)とPlaneを参照する。

更新日: 2026-09-13

以下の調査・実証記録は根拠/履歴として保持する。OPENだった作業の現在状態と実行計画はPlaneで管理する（[移行記録](docs/work/plane-md-migration-20260923-receipt.md)）。過去の修正案やSTG receiptを現在の完了証拠として扱わない。

## 索引

| ID | status | area | severity | 種別 | 修正プラン |
|:---|:---|:---|:---|:---|:---|
| BUG-LOCAL-HANDOFF-CSV-CONTRACT | Plane MIG-19 | ops / local-db | Medium | handoff preflight BLOCKED | 現在の課題・状態は `MIG-19`（詳細は移行記録） |
| BUG-RES-DOCTOR-ID-ZERO | FIXED | reservation | High | **STG実機確認済み**（担当未選択・兼務先での登録） | `created_by` の主所属制約を追加修正。STG `916159093`で城東・担当者未選択の保存201、再読込・詳細再表示200を確認。[詳細](#plan-bug-res-doctor-id-zero) |
| BUG-RES-DIALOG-A11Y-CONSOLE | FIXED | reservation / a11y | Low | **バグ断定**（DialogContent Description 欠落コンソール警告） | 固定ID上書きで Radix DescriptionWarning が context.descriptionId を見失っていた。Radix 管理IDに戻し警告0。[詳細](#plan-bug-res-dialog-a11y-console) |
| BUG-RES-STAFF-SELECT-ORPHAN-LABEL | FIXED | reservation / UI | High | **バグ断定**（担当者選択後に表示が消える） | 候補外でも表示名を保持し、確定 orphan は理由表示＋解除/再選択まで送信遮断。[詳細](#plan-bug-res-staff-select-orphan-label) |
| BUG-RES-AVAILABLE-TIMES-404 | FIXED | reservation | Medium | **バグ断定**（LINE設定欠落で院内API 404） | 院内は設定未登録を識別可能な unset（422 + code）にし案内付き手動時刻のみ許可。空枠/障害/LIFF必須は維持。[詳細](#plan-bug-res-available-times-404) |
| BUG-RES-DECEASED-STATUS-BYPASS | FIXED | reservation / pet | High | **バグ断定**（status=deceased なのに予約可） | `status=deceased` OR `deceased_at` で新規write拒否。データ補完は別項。[詳細](#plan-bug-res-deceased-status-bypass) |
| PO-PET-DECEASED-DATA-BACKFILL | Plane EMR-108 | data / pet | Medium | **PO確認**（不整合データの修復方針） | 現在の課題・状態は `EMR-108`（詳細は移行記録） |
| PO-STAFF-BLANK-NAME-LIST | FIXED | staff UX / data | Low | **PO確認→実装**（空氏名の一覧表示方針） | Grill Recommended: 初期有効のみ＋表示 `(氏名未設定)`。DB書換・一括削除なし。[詳細](#plan-po-staff-blank-name-list) |
| PO-OCCUPATION-MASTER-EMPTY | FIXED | master / data | Low | **PO確認→実装**（職種マスタ0件の扱い） | Grill Recommended: 0件は未登録案内＋職種マスタへ誘導。`occupation_id` は任意のまま。偽選択肢・自動投入なし。[詳細](#plan-po-occupation-master-empty) |
| NOTE-STAFF-STARTTIME-RDT | Plane EMR-180 | staff console | Low | **調査**（startTime TypeError・アプリ外の疑い） | 現在の課題・状態は `EMR-180`（詳細は移行記録） |
| BUG-ACCT-CLOSE-PERM-DEFAULT | Plane EMR-71 | billing / permission | Medium | **バグ断定**（既定権限でレジ締め不能。edit は dead grant、必須の create は全グループ未付与） | 現在の課題・状態は `EMR-71`（詳細は移行記録） |
| BUG-S09-FIXTURE-TEARDOWN | Plane EMR-72 | testing / fixture | Low | **バグ断定**（締め実行済みの S09 合成 clinic が teardown で削除不能。append-only 台帳との矛盾） | 現在の課題・状態は `EMR-72`（詳細は移行記録） |
| BUG-AGG-NO-VISIT-REVENUE | Plane EMR-73 | aggregation / revenue | Medium | **バグ断定**（来院なし飼主が完了会計を持っても売上ランキングに一切出ない。`include_no_visit=false` 既定除外が売上タブにも適用） | 現在の課題・状態は `EMR-73`（詳細は移行記録） |
| BUG-TRIM-KANBAN-IN-CONSULTATION | Plane EMR-74 | trimming / reception | High | **バグ断定**（受付済トリミングカードの「カルテ作成」が `in_consultation` 遷移を試みるが、カルテ必須ガードで 409。トリミングは medical_record を持たないため永久に受付済のまま） | 現在の課題・状態は `EMR-74`（詳細は移行記録） |
| BUG-BILLING-UNBILLED-MR-EXCLUSION | Plane EMR-75 | billing / unbilled | Medium | **バグ断定**（カルテ連携 pending 会計から明細を soft-delete しても、billing が `medical_record_id` を保持する限り元の treatment が未請求候補に復帰しない。サイレント請求漏れ） | 現在の課題・状態は `EMR-75`（詳細は移行記録） |
| BUG-RES-OVERLAP-500 | Plane EMR-76 | reservation / API | Medium | **バグ断定**（同一スタッフ・時間帯重複の予約作成が DB 排他制約 `excl_appointments_doctor_timerange` の 500 として漏れる。完全一致は正しく 409） | 現在の課題・状態は `EMR-76`（詳細は移行記録） |
| BUG-LIFF-HEALTHCARD-OWNER-SYNC | Plane EMR-61 | liff / line_customers | High | **バグ断定**（LIFF トークン連携が `owners.line_user_id` のみ書き `line_customers.owner_id` を更新しない。health-card は `line_customers.owner_id` 経由解決のため、連携済み飼主のヘルスカードが「ペット情報はありません」のまま） | 現在の課題・状態は `EMR-61`（詳細は移行記録） |
| BUG-ACCT-INS-SIGN-MISMATCH | Plane EMR-62 | billing / insurance | High | **バグ断定**（保険適用会計の新規確定が FE/BE の `insurance_amount` 符号規約不整合で必ず 400。FE は負値送信だが BE は `billing=total−insurance_amount` で正値を前提） | 現在の課題・状態は `EMR-62`（詳細は移行記録） |
| BUG-ACCT-INS-EDIT-REWRITE | Plane EMR-63 | billing / insurance | Medium | **バグ断定**（確定済み会計の修正保存で、未変更の `insurance_amount`/`billing_amount` が recalc 値に上書きされ、保存後に保険表示が消える。`hasInsurance` が `insurance_amount < 0` 推定のため正値・0 の保存値は OFF 表示） | 現在の課題・状態は `EMR-63`（詳細は移行記録） |
| BUG-DIALOG-FOCUS-RESTORE | Plane EMR-64 | shared UI / a11y | Low | **バグ断定**（治療プラン検索ダイアログを Escape で閉じるとフォーカスが呼出元ボタンに戻らず `document.body` に落下。実クリック開閉・キーボード開閉の両経路で再現） | 現在の課題・状態は `EMR-64`（詳細は移行記録） |
| BUG-BILLING-TAX-TYPE-DROPPED | Plane EMR-65 | billing / master | High | **バグ断定**（マスタ登録の税区分（内税/非課税）が会計明細へ伝播せず全て外税10%で請求される。処置・診察の未請求候補と物販マスタ追加の双方で再現し DB も `excluded` 固定） | 現在の課題・状態は `EMR-65`（詳細は移行記録） |
| BUG-ACCT-DUP-COMPLETE-500 | Plane EMR-66 | billing / idempotency | Medium | **バグ断定**（同一カルテへの二重会計確定が意図した 409「このカルテには既に会計があります」に届かず 500。UNIQUE 競合後の replay 判定クエリが abort 済み tx 上で実行され 25P02 になる。逐次・真並行の双方で再現。行は作られずデータは守られるが、クライアントには不透明な 500 しか返らない） | 現在の課題・状態は `EMR-66`（詳細は移行記録） |
| BUG-MR-DOCTOR-HEADER-STALE | Plane EMR-67 | medical-record / UI | Medium | **バグ断定**（ヘッダー担当医が再読込でログインユーザー表示に戻る） | 現在の課題・状態は `EMR-67`（詳細は移行記録） |
| BUG-VITAL-NOTE-KEY-MISMATCH | Plane EMR-68 | medical-record / vitals | Medium | **バグ断定**（バイタルのメモが保存も表示もされない） | 現在の課題・状態は `EMR-68`（詳細は移行記録） |
| BUG-MR-VACCINE-FORM-NESTED | Plane EMR-69 | medical-record / vaccination | High | **バグ断定**（カルテ内の接種記録追加フォームが送信不能） | 現在の課題・状態は `EMR-69`（詳細は移行記録） |
| BUG-TRIM-EXCL-TIMERANGE-500 | Plane EMR-76 | trimming / reservation | High | **バグ断定**（同一担当の連続トリミング登録が 500 で失敗） | 現在の課題・状態は `EMR-76`（詳細は移行記録） |
| BUG-PRINT-PORTAL-HIDDEN-BLANK | Plane EMR-205 | print / shared UI | High | **バグ断定**（検査結果・日次会計・月次レポート・レジ締めの印刷が全面白紙。`hidden` 属性 + 印刷ポータルの unlayered `display:block!important` が Tailwind v4 preflight `[hidden]{display:none!important}`（`@layer base`）に敗北） | 詳細は [下記](#bug-print-portal-hidden-blank) |
| BUG-LIFF-VACCINE-DATE-RAW-ISO | Plane EMR-206 | liff / pet-health | Low | **バグ断定**（ペット健康カードのワクチン接種日・次回予定日が `2026-08-15T09:00:00+09:00` の RFC3339 生値で表示。最終来院日は `time.DateOnly` で整形済みのため不整合） | 詳細は [下記](#bug-liff-vaccine-date-raw-iso) |
| BUG-BUTTON-FOCUS-INVISIBLE | Plane EMR-207 | shared UI / a11y | Medium | **バグ断定**（共有 `Button` コンポーネントにフォーカス可視インジケータが無い。`outline-none` のみで `focus-visible:ring-*` が無く、キーボード Tab でフォーカスしても見た目が変化しない。nav リンク・input は可視） | 詳細は [下記](#bug-button-focus-invisible) |

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

以下は2026-09-13時点の計画記録。未完了項目の現在の作業内容と状態はPlaneを正本とし、移行済みの局所実行計画は削除した。上記コード/UAT観測は当時の証拠として保持する。

- **成果物の範囲**: このファイルへの計画追記。製品コードの修正、データ更新、reset、migration、STG操作は未実施。
- **正本との関係**: 本メモは根拠と履歴。2026-09-23以降の実行管理はPlaneへ移行済み（[移行記録](docs/work/plane-md-migration-20260923-receipt.md)）。
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
| 独立・環境 | Plane `MIG-19` | 現在の課題本文はPlane | 再生成可能な元データ・producerが必要。取込は別の運用作業 |
| 判断後 | Plane `EMR-108` | 現在の課題本文はPlane | 死亡判定の契約、対象環境、修復・監査・復旧手順の承認 |
| 判断後 | PO-STAFF-BLANK-NAME-LIST | 表示改善とデータ整理を分けて決める | デフォルト絞り込み・代替表示の承認 |
| 判断後 | PO-OCCUPATION-MASTER-EMPTY | 職種の必須性と医院別の初期値を決める | マスタ責任者・正式な職種名の確定 |
| 独立・原因調査 | Plane `EMR-180` | 現在の課題本文はPlane | 再現元URL・stackの特定。製品バグとは未確定 |

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

<a id="plan-bug-local-handoff-csv-contract">

### 5. BUG-LOCAL-HANDOFF-CSV-CONTRACT

> Current task detail is in Plane `MIG-19`. The local execution plan was removed; historical findings above remain.

### 6. PO-PET-DECEASED-DATA-BACKFILL

> Current task detail is in Plane `EMR-108`. The local execution plan was removed; historical findings above remain.

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

<a id="plan-note-staff-starttime-rdt">

### 10. NOTE-STAFF-STARTTIME-RDT（調査計画）

> Current task detail is in Plane `EMR-180`. The local execution plan was removed; historical findings above remain.

### 11. BUG-ACCT-CLOSE-PERM-DEFAULT

> Current task detail is in Plane `EMR-71`. The local execution plan was removed; historical findings above remain.

### 12. BUG-S09-FIXTURE-TEARDOWN

> Current task detail is in Plane `EMR-72`. The local execution plan was removed; historical findings above remain.

### 13. BUG-AGG-NO-VISIT-REVENUE

> Current task detail is in Plane `EMR-73`. The local execution plan was removed; historical findings above remain.

### 14. BUG-TRIM-KANBAN-IN-CONSULTATION

> Current task detail is in Plane `EMR-74`. The local execution plan was removed; historical findings above remain.

### 15. BUG-BILLING-UNBILLED-MR-EXCLUSION

> Current task detail is in Plane `EMR-75`. The local execution plan was removed; historical findings above remain.

### 16. BUG-RES-OVERLAP-500

> Current task detail is in Plane `EMR-76`. The local execution plan was removed; historical findings above remain.

### 17. BUG-LIFF-HEALTHCARD-OWNER-SYNC

> Current task detail is in Plane `EMR-61`. The local execution plan was removed; historical findings above remain.

### 18. BUG-ACCT-INS-SIGN-MISMATCH

> Current task detail is in Plane `EMR-62`. The local execution plan was removed; historical findings above remain.

### 19. BUG-ACCT-INS-EDIT-REWRITE

> Current task detail is in Plane `EMR-63`. The local execution plan was removed; historical findings above remain.

### 20. BUG-DIALOG-FOCUS-RESTORE

> Current task detail is in Plane `EMR-64`. The local execution plan was removed; historical findings above remain.

### 21. BUG-BILLING-TAX-TYPE-DROPPED

> Current task detail is in Plane `EMR-65`. The local execution plan was removed; historical findings above remain.

### 22. BUG-ACCT-DUP-COMPLETE-500

> Current task detail is in Plane `EMR-66`. The local execution plan was removed; historical findings above remain.

### BUG-AGG-NO-VISIT-REVENUE: 来院なし飼主が完了会計を持っても売上ランキングに表示されない

> Current task detail is in Plane `EMR-73`. The local execution plan was removed; historical findings above remain.

### BUG-TRIM-KANBAN-IN-CONSULTATION: 受付済トリミングカードが「診療中」へ遷移できず受付済に滞留する

> Current task detail is in Plane `EMR-74`. The local execution plan was removed; historical findings above remain.

### BUG-BILLING-UNBILLED-MR-EXCLUSION: カルテ連携会計から明細削除しても請求元 treatment が未請求候補に復帰しない

> Current task detail is in Plane `EMR-75`. The local execution plan was removed; historical findings above remain.

### BUG-RES-OVERLAP-500: 時間帯重複の予約作成が排他制約 500 で漏れる（完全一致は 409）

> Current task detail is in Plane `EMR-76`. The local execution plan was removed; historical findings above remain.

### BUG-LIFF-HEALTHCARD-OWNER-SYNC: LIFF 連携が成立しても `line_customers.owner_id` が更新されずヘルスカードが空のまま

> Current task detail is in Plane `EMR-61`. The local execution plan was removed; historical findings above remain.

### BUG-ACCT-INS-SIGN-MISMATCH: 保険適用会計の新規確定が FE/BE 符号規約不整合で必ず 400

> Current task detail is in Plane `EMR-62`. The local execution plan was removed; historical findings above remain.

### BUG-ACCT-INS-EDIT-REWRITE: 確定済み会計の無変更保存で保険金額・割合表示が保持されない

> Current task detail is in Plane `EMR-63`. The local execution plan was removed; historical findings above remain.

### BUG-DIALOG-FOCUS-RESTORE: 治療プラン検索ダイアログを閉じるとフォーカスが呼出元へ戻らず body に落下する

> Current task detail is in Plane `EMR-64`. The local execution plan was removed; historical findings above remain.

### BUG-BILLING-TAX-TYPE-DROPPED: マスタの税区分（内税/非課税）が会計明細へ伝播せず全て外税10%で請求される

> Current task detail is in Plane `EMR-65`. The local execution plan was removed; historical findings above remain.

### BUG-ACCT-DUP-COMPLETE-500: 同一カルテへの二重会計確定が意図した 409 に届かず 500 で漏れる

> Current task detail is in Plane `EMR-66`. The local execution plan was removed; historical findings above remain.

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

---

## 確認済み製品欠陥（UAT 2026-09-23 · V01 追加分）

`docs/ops/testing/scenarios/V01-clinical-forms.md` の UAT 実行で確定した製品欠陥。証拠は `reports/uat-2026-09-23/V01-clinical-forms.md`（gitignore 対象の日次レポート）。正本登録は `todo.md#product-bugs` と重複確認済み。S01–S33 由来の欠陥（BUG-LIFF-HEALTHCARD-OWNER-SYNC・BUG-ACCT-INS-SIGN-MISMATCH・BUG-ACCT-INS-EDIT-REWRITE・BUG-DIALOG-FOCUS-RESTORE・BUG-BILLING-TAX-TYPE-DROPPED・BUG-ACCT-DUP-COMPLETE-500）は上記の番号付き計画節を参照。

<a id="plan-bug-mr-doctor-header-stale"></a>

### BUG-MR-DOCTOR-HEADER-STALE（V01・Medium）

- **現象**: カルテヘッダーの「担当医」表示が、保存済み `doctor_id` ではなく常にログインユーザー名を表示する。担当医を変更すると即時 PATCH `{"doctor_id":…}` が発行され DB には正しく保存されるが、ブラウザ再読込後のヘッダーは再びログインユーザー名に戻る。
- **実証**: MR `1000000016` で担当医を 林文明 → 高橋純子（`doctor_id=10000003`）へ変更し PATCH 201・DB 永続を確認。再読込後も DB は `doctor_id=10000003` のままだが、ヘッダー表示は `担当医 林 文明`。
- **根因**: `MedicalRecordFormReadyPanels.tsx` の `staffName` が `useState(() => user?.displayName ?? "")` で初期化され、`useGetMedicalRecord` で取得した `record.doctor`（GET レスポンスに `doctor?: Staff` あり）から hydrate する経路がない。更新は `handleSelectStaff` 経由のローカル state のみ。
- **影響**: 別スタッフがカルテを開くと担当医が自分の名前に見える（誤帰属表示）。データ自体は正しいため表示限定だが、臨床帰属の誤認リスクあり。
- **証拠**: `reports/uat-2026-09-23/V01-clinical-forms.md`（§1 手順5）

<a id="plan-bug-vital-note-key-mismatch"></a>

### BUG-VITAL-NOTE-KEY-MISMATCH（V01・Medium）

- **現象**: バイタル記録のメモ（note）を入力して追加すると、保存は 201 で成功するが DB の `notes` カラムは空のまま。一覧のメモ列も常に `-` 表示。
- **実証**: MR `1000000016` の VitalsModal でメモ `V01メモ` を入力 → POST body `{"note":"V01メモ",…}` → `vital_records.notes` は空。既存メモの表示経路も FE が `vital.note` を読むのに対し BE response は `json:"notes"` のため表示不能。
- **根因**: wire key の不一致。FE `Vital`/`CreateVitalInput`/`UpdateVitalInput` は `note`（`medical-records/types/index.ts`）、BE `vital_request.go`/`vital_response.go` は `json:"notes"`。Go の unknown-key 無視によりエラーなく silent drop。
- **影響**: バイタルメモがユーザーの知らないうちに保存されず、既存メモも表示されない（軽度だが確実なデータ損失）。
- **証拠**: `reports/uat-2026-09-23/V01-clinical-forms.md`（§3 手順4）

<a id="plan-bug-mr-vaccine-form-nested"></a>

### BUG-MR-VACCINE-FORM-NESTED（V01・High）

- **現象**: カルテ編集画面の「予防接種」タブ →「記録を追加」で開くインライン接種フォームが一切送信できない。ワクチン未選択でもバリデーションエラー（`ワクチン種別を選択してください`）が表示されず、ワクチン＋接種日を入力しても POST が発行されない。
- **実証**: MR `1000000016` 予防接種タブで「接種記録を追加」を実クリック → 送信イベントは発火するが `fieldErrors` 未描画・`vaccinations` POST なし・DB 0 件。コンソールに `Running the JavaScript URL violates CSP 'script-src 'self''` が毎回記録される。内側 `<form>` の `action` 属性は React の `javascript:throw new Error('A React form was unexpectedly submitted…')` placeholder のまま。
- **根因**: `MedicalRecordFormReadyPanels.tsx:182` の外側 `<form action={form.formAction}>` がカルテ全体を包み、`MedicalRecordVaccination.tsx:96` の内側 `<form action={formAction}>` がネスト。HTML では form のネストは無効で、React の useActionState submit 横取りが内側フォームに効かず、既定送信が CSP でブロックされて無害に失敗する。
- **影響**: カルテ内からの接種記録追加が完全に使用不能（`/vaccinations/new` 独立ルートは正常・S29 実証済み）。機能喪失＋ユーザーへのフィードバックゼロ。
- **証拠**: `reports/uat-2026-09-23/V01-clinical-forms.md`（§6）

<a id="plan-bug-trim-excl-timerange-500"></a>

### BUG-TRIM-EXCL-TIMERANGE-500（V01・High）

- **現象**: 同一担当スタッフで同日に別ペットのトリミング（record_shortcut）を続けて登録すると、`POST /api/v1/trimmings` が **500 Internal Server Error** で失敗する。ユーザーにはトースト・インラインエラー等のフィードバックがなく保存できない。
- **実証**: 担当 `UAT担当医`（staff_id=1000000000）で豆助（pet 1000002）のトリミングを保存 → appointment `1000000015`（start 05:26:54+09 / end 06:56:54+09 の 90 分枠）作成。約 23 分後にマメラ（pet 1000001）で同一担当・同一コース系の 2 件目を保存 → 500×2 回。backend ログに `ERROR: conflicting key value violates exclusion constraint "excl_appointments_doctor_timerange" (SQLSTATE 23P01)`（`reservation_repository.go:259`）。2 件目の start_time は 05:49:29 付近で 1 件目の 90 分枠内に落下。
- **根因**: 2 層の問題。(a) FE `defaultRecordShortcutTimes`（`trimming-form-utils.ts`）は BUG-010 対策で「固定 10:00 → 現在 JST 時刻+90 分」にしたが、同一担当の連続登録は依然として時間枠が重複する（一意化は時刻文字列のみで枠の非重複は保証しない）。(b) BE は exclusion violation（23P01）を 409 conflict へマップしておらず、tx エラーがそのまま 500 として返る。
- **影響**: 同一スタッフが 90 分以内に複数トリミングを連続登録できない。エラーハンドリング不在のため UI は無音失敗に近く、V01 §12-6 の期待（一意な時刻が付き無関係な 2 件目がブロックされない）を満たさない。
- **証拠**: `reports/uat-2026-09-23/V01-clinical-forms.md`（§12 手順6）

---

<a id="bug-print-portal-hidden-blank"></a>

### BUG-PRINT-PORTAL-HIDDEN-BLANK（S37・High）

- **現象**: 印刷ポータルを使う帳票面（検査結果・日次会計・月次集計レポート・レジ締め明細）で印刷/PDF 出力を実行すると、アプリ本体は非表示になるが帳票本体も `display:none` のままで、**出力が白紙になる**。カルテ印刷（`MedicalRecordPrintView`）と領収書（`AccountingPrintArea`）は `hidden` Tailwind クラス方式のため正常。
- **実証**（`reports/uat-2026-09-23/S37-print-documents-layout.md`）: Playwright の `emulateMedia({ media: "print" })` で実測。
  - `/accounting?tab=daily` → `[data-testid="daily-print-area"]` が print メディアで `display:none`（`hidden` 属性付与のまま）。`#root` も `none` → 印刷面 0。
  - `/accounting/reports` → `[data-print-portal]`（monthly）が print で `display:none`、`#root` も `none`。
  - `/examinations/1000000000` → `[data-print-portal]`（examination）が print で `display:none`、`#root` も `none`。
  - `/accounting?tab=daily` の `page.pdf()` が **1,156 bytes**（実質 1 ページ白紙）。
  - 対照: `/medical-records/1000000019` の印刷面（`hidden print:block` クラス方式）は print で `display:block`・本文 498 文字 → 正常。
- **根因**: Tailwind v4.3.3 preflight（`tailwindcss/preflight.css:391`）が `[hidden]:where(:not([hidden='until-found'])) { display: none !important; }` を **`@layer base` 内**で出力する。CSS Cascade 5 では `!important` 宣言のレイヤー優先順位が反転するため、**先に宣言された `base` レイヤーの `!important` が、unlayered の `!important` に勝つ**。`PrintPortal.tsx`（`<div hidden … data-print-portal>` + unlayered `[data-print-portal]{display:block!important}`）と `DailyAccountingPrintArea.tsx`（`hidden` + unlayered `[data-testid="daily-print-area"]{display:block!important}`）はいずれも unlayered のため `display:none!important` に敗北する。実測でも unlayered `!important` は `none` のまま、`@layer base`/`@layer utilities` の同一宣言なら `block` になることを確認（Chrome 実測）。`hidden` Tailwind クラス（`.hidden`）は属性ではなくクラスのため preflight の対象外で、`print:block` が同じ utilities レイヤー内で後勝ちし正常動作する。
- **影響**: 検査結果・日次会計・月次レポート・レジ締めの印刷/PDF がすべて白紙。帳票運用（監査・締め・月次）が成立しない。High。
- **修正方針候補**: (a) 各印刷面の `hidden` 属性をやめ、`hidden print:block`（Tailwind クラス）方式に統一する（MR/領収書と同方式）、(b) 印刷ポータルの上書き規則を `@layer base`（または preflight より前のレイヤー）に置く、(c) `@media print` 内で `[hidden]` を `display:block!important` で上書きする規則を `@layer base` に追加する。回帰は「print メディアで各ポータルが `display:block` になること」を固定する。
- **証拠**: `reports/uat-2026-09-23/S37-print-documents-layout.md`

---

<a id="bug-liff-vaccine-date-raw-iso"></a>

### BUG-LIFF-VACCINE-DATE-RAW-ISO（S38・Low）

- **現象**: LIFF ペット健康カードのワクチン記録テーブルで「接種日」「次回予定日」が `2026-08-15T09:00:00+09:00` の RFC3339 生値で表示される（LINE 利用者に見える画面）。同じカードの「最終来院日」は `2026-08-15` 形式で整形されており、同一画面内で書式が不整合。
- **実証**: `http://localhost:3003/liff/?clinic_id=1`（mock lane、データは実バックエンド由来。飼主 `UATヘルス 飼主A` / 犬A）を 390×844 で表示。ワクチン行が `10%Ｐｒｏ-Ｈｅａｒｔ SR 12(10.1～20.0kg)` / `2026-08-15T09:00:00+09:00` / `2027-08-15T09:00:00+09:00`。S12 でも観測済み（未登録）。
- **根因**: `backend/internal/reservation/liff_response.go` の `liffHealthCardVaccineResponse` が `VaccinatedAt time.Time` / `NextDueAt *time.Time` をそのまま公開（`toLiffHealthCardResponse` で整形なし）。一方 `LastVisitDate` は同ファイル内で `time.DateOnly` 整形（294 行）。FE `frontend/liff/src/pages/PetHealthPage.tsx:173,175` も `v.vaccinated_at` / `v.next_due_at` を無整形で描画。
- **影響**: 利用者向けに機械可読タイムスタンプが露出。可読性・信頼感の低下（機能は動作）。Low。
- **修正方針候補**: バックエンドで `LastVisitDate` と同様に `time.DateOnly`（または FE で `formatJSTDate`）へ統一する。接種日は本来日付粒度のため、レスポンスを日付文字列に揃えるのが自然。回帰は「接種日/次回予定日が `YYYY-MM-DD` 形式であること」を固定。
- **証拠**: `reports/uat-2026-09-23/S38-liff-mobile-viewport.md`

---

<a id="bug-button-focus-invisible"></a>

### BUG-BUTTON-FOCUS-INVISIBLE（S39・Medium）

- **現象**: 共有 `Button` コンポーネントで描画されたボタン（更新・戻る・保存 等、主要操作の大半）に、キーボードフォーカス時の可視インジケータが無い。Tab でフォーカスしても見た目が一切変化しない。同じ画面でも nav リンク（`<a>`）はブラウザ既定の outline、input はブランド色 2px リングが出るため、ボタンだけが不可視。
- **実証**（`reports/uat-2026-09-23/S39-state-feedback-visibility.md`、Playwright 実測）:
  - `Button(更新)` / `Button(戻る)`: `:focus-visible = true`、`outline-style: none`、着色 box-shadow なし → 可視インジケータなし。
  - 対照 `nav a`: `:focus-visible = true`、`outline: auto 1px rgb(3,139,148)` → 可視。
  - 対照 `input#phone`: `:focus-visible = true`、`box-shadow: … rgb(3,139,148) 0 0 0 2px` → 可視。
- **根因**: `frontend/src/components/ui/button-variants.ts` の cva 基底クラスが `outline-none` のみで `focus-visible:ring-*` を持たない（input は `C.focusRingActionPrimary` = `focus:shadow-focus-primary` を使用しているが、Button には相当の指定が無い）。`globals.css` にもボタン向けの `:focus-visible` 代替規則は無い（`--shadow-focus-primary` 等の変数定義のみ）。
- **影響**: WCAG 2.2 AA 2.4.7（Focus Visible）/ 2.4.11 に不適合。キーボード操作時にどのボタンにフォーカスがあるか判別できず、誤操作・操作不能に近い状態。アプリ全体の主要操作に波及するため Medium。
- **修正方針候補**: `buttonVariants` の基底に `focus-visible:ring-2 focus-visible:ring-[#038B94] focus-visible:ring-offset-1`（既存 `--shadow-focus-primary` と同等）を追加する。回帰は「各 variant の Button が `:focus-visible` で着色リングを持つこと」を固定。
- **証拠**: `reports/uat-2026-09-23/S39-state-feedback-visibility.md`



