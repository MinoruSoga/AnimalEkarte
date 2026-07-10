# BE-refactor.md — バックエンド リファクタリング計画（唯一の正本）

> 本ファイルが唯一の正本である。`BE-refactor-v2.md`（2026-07-10作成、当時「旧 BE-refactor.md は参照のみ・編集禁止」と規定していた）は
> 2026-07-11 に §2/§3 の R0〜R21 全項目を現 HEAD で再実測し、DONE 項目は本ファイルへ再掲せず（詳細は `git log` 参照）、
> OPEN 項目のみを本ファイル新設の「挙動保存残件」節へ統合したうえで `git rm` 済み。v2 の「旧 BE-refactor.md 編集禁止」規定は本統合をもって破棄する。
> X-9（resv-slot-phantom-toctou）・X-10（mr-version-check-not-atomic）・finalize-child-write-race は挙動変更トラックとして実装完了・CLOSED（詳細は git log 参照）。
> 前提: backend は複数回の系統的リファクタと監査で well-maintained と判定済みのコードベースである。

## 監査の方法と信頼性

- 対象: `backend/` 配下全体。15次元で並列監査 → 敵対的検証（Appendix A・H）。R 系は 2026-07-11 に現 HEAD で個別再実測（ファイル存在・grep・scoped `go test`）。
- 本編（挙動保存）と Appendix A（挙動変更・別トラック）に分離。CLOSED 済みの完了履歴は本ファイルから削除済み。

## サマリー

| 区分 | 件数 | 内訳 |
|---|---|---|
| Appendix A（挙動変更・別トラック） | 5件 | X-14〜X-18 |
| レビュー由来フォローアップ（未登録・別チケット推奨） | 8件 | H-1〜H-8 |
| 挙動保存残件（旧 BE-refactor-v2 R系・OPENのみ） | 6件 | R10〜R15 |
| 別トラック残件（旧 BE-refactor-v2 §4・未着手） | 4件 | G6-2, F3, F6, B-2 |

**旧 BE-refactor-v2.md の R1〜R9・R16〜R21（15件）は 2026-07-11 の現 HEAD 実測で完了確認済み**（`docker compose exec backend go test ./internal/repository/ ...` 等の scoped 実行で該当テストが全て PASS）。実装の詳細手順・コミット履歴は `git log` を参照。R16〜R19（api.yaml の masters/reservation-staffs/reservations/shifts/checkups/pets/medical-records/hospitalizations/owners/lstep/LINE webhook 系オペレーション文書化）は v2 が想定した番号順の別コミット群ではなく「G1-2 residual follow-up (2026-07-10)」という一括作業で達成されており、`internal/apicontract` の route drift gate は現在 PASS（allowlist 残存は意図的 pin のみ）。

### レビュー由来フォローアップ（本編未登録）

| ID | 内容 | 発見元 | 優先度 |
|---|---|---|---|
| H-1 | `UpdateStaffGroups` の staff_id 単位 DELETE が多施設所属スタッフの他クリニックグループ紐付けを意図せず削除しうる | G11-1 security-reviewer | HIGH — 別チケット化推奨 |
| H-2 | `UpdateExcludedReservationTypes`（reservation_staff_repository.go）の DELETE が `staff_id` のみでスコープされ `clinic_id` を含まない一方、INSERT は呼び出しクリニックの型IDのみ。`staff_reservation_exclusions` テーブル自体に `clinic_id` 列が無いため、多施設所属スタッフに対しては clinic A の正当な操作が clinic B の除外設定行を無警告で全削除する（H-1 と同型のクロステナント破壊）。兄弟の `UpdateReservationCapabilities`/`staff_reservation_capabilities` は自前 `clinic_id` 列を持ち `Where("clinic_id = ? AND staff_id = ?")` で正しくスコープされており非対称。 | G11-4 security-reviewer（`UpdateReservationCapabilities` との比較監査で発見） | HIGH — 別チケット化推奨（`staff_reservation_exclusions` への `clinic_id` 列追加 or DELETE を真の差分更新へ変更、要 migration） |
| H-3 | `billing_items.category` に索引が無く、`FindOwnersByCategoryPurchaseDate`（Lstep FEAT-383 配信ターゲティング、バッチ/cron想定）が `category = ?` 述語 + `billings` join で Seq Scan リスク。テーブル成長に伴い悪化。既存索引は `merchandise_item_id`/`treatment_id`/`appointment_id`/`trimming_course_id`/`trimming_option_id`/`billing_id`/`deleted_at` のみで `category` は対象外。`idx_billings_clinic_completed_at` も `WHERE status='completed'` の部分索引でこの3クエリ（status述語なし）はカバーしない。 | G11-5 database-reviewer | MEDIUM（パフォーマンス、要 migration・別チケット推奨: `CREATE INDEX idx_billing_items_category ON billing_items(category) WHERE deleted_at IS NULL`） |
| H-4 | `audit_logs.clinic_id` が Go では `*uint64`（nil許容）だが DDL では `bigint NOT NULL REFERENCES clinics(id)`。`gorm:"not null"` と実DDLテストは済だが、Go 型は `*uint64` のまま（コンパイル時保証なし）。実防御は `validateAuditLog`（nil/0 拒否）。 | G12-1 schema_drift nullability check | LOW（残作業は型を `uint64` 非ポインタ化するのみ・別チケット推奨） |
| H-5 | `lstep_csv_imports.uploaded_by_user_id` が Go では `*uint64`（nil許容）だが DDL では `bigint NOT NULL REFERENCES accounts(id)`。H-4 と同型のクラス。 | G12-1 schema_drift nullability check（新設） | MEDIUM（要 migration or model 修正・別チケット推奨） |
| H-6 | `backend/CODING_RULES.md` の §3.2/§5.1/§5.4/§6.1/§6.3 に、G1-6 で是正した README.md と同型の forbidden-pattern 教材コード（生の `gin.H{"error":...}` レスポンス、`uuid.UUID` ベースの `FindByID` シグネチャ例 — 実際は全モデル `uint64` PK、sentinel-error `errors.Is` 例示で `apperrors.FromGORM`/`RespondError` 未使用）が残存。§6 に `RequirePermission`/P5 ルートゲーティングの言及が一切ない。G1-6 の対象範囲（ディレクトリツリーのみ）を超える約400行規模の書き直しのため別ユニット化推奨。 | G1-6 実装エージェント | MEDIUM（オンボーディング文書の質・別チケット推奨） |
| H-7 | `reservationStaffService.Update` の所有権確認読み取り(`s.GetByID`)が tx 外で行われ、確認〜更新の間にスタッフが削除されると TOCTOU の窓が生じる。X-8 の修正対象（fields 更新+除外設定置換の原子性）とは独立した既存の設計であり、X-8 は悪化させていない（security-reviewer 確認済み）。低頻度の管理操作のため実害は限定的。 | X-8 security-reviewer | LOW（別チケット化検討・優先度低） |
| H-8 | finalize-child-write-race の修正は treatment.Create／examination.Create・Update／vital.Create・Update・Delete／checkup_field_result.ReplaceForCheckup の5経路のみ `LockDraftByID` 行ロックで保護した。同一クラスの check-then-act（親カルテ確定済みチェックが素の `FindByID` で tx 外・無ロック）が `treatment_service.go` の Update・Delete、`examination_service.go` の Delete、`prescription_service.go` の Delete に残存する（対応方針: 同じ `LockDraftByID` + 子書込 tx パターンを適用。examination/prescription の Delete は現在 `Transactor.WithTx` 未使用のため新規導入が必要）。HC-003/005/006 は "invariant" と定義されている以上、一部経路のみの保護は invariant の完全復元とは言えない。加えて `treatment_service.go` の `BulkUpdateSortOrder`（465行）は確定ステータスの gate 自体が無い（clinic 所有権確認のみで finalize チェックなし）ため、確定済みカルテでも治療の並び順を無条件に変更できる — race というより欠落した業務ルールチェックで、他の残存項目より単純だが影響は同型（security-reviewer 発見）。 | finalize-child-write-race healthcare-reviewer / security-reviewer（2026-07-11 セッション） | HIGH — 別チケット化推奨（silent close 不可、必ず追跡すること） |

---

## Appendix A: 挙動変更を伴う項目（別トラック・PO/責任者判断を要する）

以下5件は監査で実在を確認した defect だが、修正すると HTTP レスポンス・DB書込結果・権限判定・API契約のいずれかが観測可能な形で変わる。このため本計画（挙動保存リファクタ）の実行対象からは外し、個別 Issue として起票のうえ別トラックで扱うことを推奨する。severity 順に記載。

### X-14. master-FK write allowlistのknown-unguarded残28エントリにisolation test不在

- **ID**: `test-known-unguarded-master-fk-isolation-tests`
- **重要度**: P2 / **工数目安**: L
- **対象ファイル**: internal/service/master_fk_write_inventory_lint_test.go
- **依存関係**: 各エントリのガード実装(挙動変更)が先行必須。テストのみ先行するとCIがREDになる

**証拠(現HEAD実測: 2026-07-11)**

`grep -cE '^\s*\{"[A-Za-z]+\.[A-Za-z]+", statusKnownUnguarded,' backend/internal/service/master_fk_write_inventory_lint_test.go` → **28件**。残クラスタ: `billingItemService.CreateItem`(MerchandiseItemID残・DEAD field注記), `inquiryService.Save`, `labImportExaminationService.PersistBatch/PersistExam`, `labResultImportService.Commit`, `liffService.CreateReservation`(TrimmingCourseID/OptionID残), `medicalRecordService.CreateSubRecords`, `medicineService.Create/Update`(InventoryID/ParentID), `ownerService.CreateWithPets`・`petService.Create/Update`(InsuranceID), `reservationAdminService.Create`・`reservationService.Create/Update`・`reservationStaffService.Create/Update`(ExcludedTypeIDs)・`reservationTypeService.Create/Update`(GroupID)・`reservationValidators.ValidateAndCreate`, `staffService.Create/CreateWithAccount/Update`(OccupationID), `treatmentService.Create/Update`(InventoryID残), `trimmingCourseService.Update`(CourseTypeID)・`trimmingService.Create/Update`(CourseID/OptionIDs)。

**問題**

review網羅性gateは『名簿に載せる』ことしか担保せず(同ファイル冒頭に明記)、known-unguardedのまま滞留しているwrite経路はクロステナントmaster FK書き込みが実際に拒否されるかを誰も検証していない。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

別トラック(挙動変更)としてドメイン毎に1PRずつ段階実施: service層にFindByID(clinicID,…)ガード追加 → 既存 `cross_tenant_master_fk_write_test.go` パターンで「別クリニックFK指定→apperrors.WrapInvalidInput/NotFound拒否」テストを追加 → allowlistエントリをguardedへ更新。各バッチはTestMasterFKWriteInventoryのstatus突合がgateになるためallowlist更新漏れはCIで検出される。STGデータ監査(既存越境データの有無)を先行させること。これまでの実施バッチ（会計/campaign/self-ref ParentID群等）の実装詳細・件数推移はコミット履歴(`git log --oneline -- backend/internal/service/master_fk_write_inventory_lint_test.go`)を参照し、本ファイルには残件のみを保持する。**本セクションは残28件が存在するためCLOSEDではない。**

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -count=1
docker compose exec backend go test ./internal/service/ -run '_Rejects' -count=1
```

### X-15. P6逸脱: 状態トグル系 DELETE 4ルートが "delete" ではなく "edit" 権限でゲート(免除根拠コメントなし)

- **ID**: `p6-delete-routes-edit-permission`
- **重要度**: P3 / **工数目安**: S
- **対象ファイル**: internal/handler/handler.go (158,185); internal/handler/pet_handler.go (158,165)
- **依存関係**: PO判断(権限ポリシー)。案(b)の場合は FE の権限ガードに波及の可能性

**証拠(現HEAD検証済み)**

internal/handler/handler.go:158: `owners.DELETE("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeleteOwnerLstepOptOut)`(185 に co エイリアス同等行)。internal/handler/pet_handler.go:158: `pets.DELETE("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)`(165 に clinicPets エイリアス同等行)。対照: 同一リソースの handler.go:155 `owners.DELETE("/:id/line", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLine)`、handler.go:170 の lstep/tags DELETE も "delete"。セマンティクス裏付け: internal/handler/lstep_lifecycle_handler.go:80 「DELETE /owners/:id/lstep-opt-out — オーナーを Lステップ配信にオプトインする（BE-017）」、同39 「DELETE /pets/:id/death — ペット死亡取り消しを記録し CPM タグを再同期する」。等価操作の統合エンドポイント PatchOwnerLstepOptOut(PATCH, handler.go:160)も "edit"。ルート登録箇所の前後コメント(handler.go:156-157 / pet_handler.go:156)に権限選定の根拠記載なし。

**問題**

P6 は「DELETE ルートには delete 権限（edit は違反）」と規定し、全 DELETE ルート走査でこの4行のみが "edit"(LIFF公開 CancelLiffReservation は免除)。RBAC 上 delete 権限を剥奪してもこの4操作は edit 保持者に実行可能。ただし実体は資源削除ではなく状態解除(オプトイン/死亡取消)で、同一操作が PATCH(edit) でも実行できるため delete に揃えると同一業務操作の必要権限がエンドポイント表現で割れる。意図的設計の可能性が高いが明文化されておらず、P6 スキャン運用の恒常的偽陽性源になっている。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

PO/責任者に「undo系DELETE の要求権限を delete に上げるか、P6 例外として明文化するか」を確定させてから着手。(a) 例外維持(挙動保存): 4ルート行直上に「P6例外: 状態トグル(資源削除でない)のため PATCH 等価の edit を要求(PO決定日付)」コメントを追記し、.claude/refs/gin-architecture-compliance.md の P6 節と backend/internal/handler/CLAUDE.md に免除注記を追加。(b) delete化(挙動変更): handler.go:158,185 / pet_handler.go:158,165 の action を "delete" に変更し、internal/handler/lstep_lifecycle_handler_test.go の TestDeletePetDeath / TestDeleteOwnerLstepOptOut に権限不足403ケースを追加、FE 側の該当操作 can 判定への波及を grep 確認。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/handler/ -run 'TestDeletePetDeath|TestDeleteOwnerLstepOptOut|TestRegisterRoutes_NoPanic' -count=1
```

### X-16. 健診クリニック横断一覧がページネーションなし・フィルタ全optional(全件+5 Preload)、期限アラートも下限なし全滞留分を返す

- **ID**: `checkup-list-unbounded`
- **重要度**: P3 / **工数目安**: M
- **対象ファイル**: internal/repository/checkup_repository.go (44-70,107-124); internal/handler/checkup_request.go (89-112); internal/service/checkup_service.go (117-130)
- **依存関係**: FE同期必須。FindAlerts の下限はPO判断。checkups-vaccinations-missing-composite-index を先行推奨

**証拠(現HEAD検証済み)**

checkup_repository.go:44-70 FindByClinicID は `if filters.StartDate != nil { q = q.Where(...) }` と全フィルタが optional で、:65 `err := q.Order("date DESC").Find(&checkups).Error` に Limit/Offset なし。:48-52 で `Preload("CheckupType"...).Preload("Doctor"...).Preload("MedicalRecord"...).Preload("MedicalRecord.Pet"...).Preload("MedicalRecord.Pet.Owner"...)` の5 Preload。checkup_request.go:98-101 は `optionalStringQueryFilter(values.Get("start_date"))` でクエリ未指定を許容し必須化していない。FindAlerts(:116-118)は `next_date <= upperBound` のみで下限・LIMITなし＝過去の期限切れ全件を毎回返す。他の取引系一覧(owners/pets/medical_records/vaccinations/examinations/hospitalizations/estimates/treatments 等)は全て page/limit 実装済みで、checkups のみ例外。

**問題**

健診記録は診療ごとに増える成長テーブルであり、FE がフィルタを省略した場合(またはブラウザから直接叩かれた場合)にクリニック全履歴+5関連の全件シリアライズが走る。ページネーション規約(gin-api-design)からの逸脱でもある。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラック(API形状変更)として扱う。(1) FindByClinicID に page/limit を追加し vaccination_repository.go:35-70 と同型(buildBase closure + Count + Offset/Limit)へ。handler は owners 等の既存 page/limit クエリ規約に合わせ既定 limit を設定。(2) FE(健診管理ページ)のクエリ同期が必要。(3) FindAlerts は業務要件(過去滞留分をどこまで表示するか)を PO 確認の上で下限日付 or LIMIT を導入。先行して checkups-vaccinations-missing-composite-index のインデックスを入れれば当面の劣化は緩和される。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestCheckupRepository -count=1 && docker compose exec backend go test ./internal/service/ -run 'TestCheckupService' -count=1
```

### X-17. RequireXRequestedWithのエラーレスポンスがmiddleware共通スキーマ（code/message/timestamp）から逸脱し{"error":...}を返す

- **ID**: `csrf-error-schema-drift`
- **重要度**: P3 / **工数目安**: S
- **対象ファイル**: internal/middleware/csrf.go (22-27); internal/middleware/response.go (9-17)

**証拠(現HEAD検証済み)**

internal/middleware/csrf.go:22-26:
		if c.Request.Header.Get("X-Requested-With") == "" {
			err := apperrors.WrapForbidden("X-Requested-With header required for state-changing requests")
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			c.Abort()

internal/middleware/response.go:9-16:
// respondError はミドルウェア層共通のエラーレスポンスを返す。
// handler 層の RespondError と同一スキーマ（code/message/timestamp）を使用する。
func respondError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"code":      status,
		"message":   msg,
		"timestamp": time.Now(),

**問題**

middleware層は respondError で handler 層 RespondError と同一のエラーエンベロープ（code/message/timestamp）に統一されている（auth.go/liff_auth.go/rate_limit.go は全て準拠）が、csrf.go のみ {"error": "..."} 形式を返す。FE のエラーハンドラがスキーマ別分岐を強いられる一貫性負債。レスポンスボディ形状が変わるため挙動変更扱い。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) csrf.go:23-26 を respondError(c, http.StatusForbidden, "X-Requested-With header required for state-changing requests") の1行に置換（apperrors.WrapForbidden 生成は不要になる。ステータス403は不変）。(2) csrf_test.go のボディ検証を新スキーマに更新。(3) frontend 側で当該 403 の "error" キーをパースしている箇所がないか grep（axios interceptor が message キー前提なら影響なし）確認の上で適用。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/middleware/ -run RequireXRequestedWith -count=1
```

### X-18. password_reset の30sタイムアウトが smtp.SendMail に非伝播（goroutine が無期限にブロックしうる）

- **ID**: `KR-pwreset-smtp-timeout`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/service/password_reset_service.go (101-109, 178-210)
- **依存関係**: なし（単独ファイル完結）

**証拠(現HEAD検証済み)**

internal/service/password_reset_service.go:101-108:
	go func() { //nolint:gosec,contextcheck // fire-and-forget: request ctx キャンセル後も送信継続が必要なため context.Background を使用
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:gosec // 上記と同理由
		defer cancel()
		if sendErr := s.sendResetEmail(email, resetURL); sendErr != nil {
			slog.ErrorContext(bgCtx, "failed to send password reset email",
— bgCtx はログ出力にしか使われず sendResetEmail に渡らない。password_reset_service.go:206:
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
— smtp.SendMail は context/deadline を受け取らないため 30s タイムアウトは実質無効。

**問題**

既知残存項目1の再確認（現HEADで現存）。SMTP サーバが応答しない場合、送信 goroutine が OS の TCP タイムアウトまで（あるいは無期限に）ブロックし、cancel は名ばかりになる。fire-and-forget なのでリクエストには影響しないが、リセットメール未達がタイムアウトログすら出さずに滞留し、goroutine リークの温床になる。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラックで扱う。手順: 1) sendResetEmail(ctx context.Context, to, resetURL string) にシグネチャ変更。2) 実装を smtp.SendMail から (&net.Dialer{}).DialContext(ctx, "tcp", addr) + smtp.NewClient に置換し、conn.SetDeadline(deadline) を ctx の deadline から導出（Auth/Mail/Rcpt/Data/Quit を明示実行。STARTTLS 対応は現行 SendMail と同等の分岐を踏襲）。3) 呼び出し側 goroutine で sendResetEmail(bgCtx, ...) に変更し //nolint:contextcheck を除去。4) ハングするダミーリスナー（net.Listen して読み捨て）で「deadline 内に error 復帰する」回帰テストを追加（既存 TestPasswordResetService_SendResetEmail の隣）。影響範囲は本ファイルのみ（PasswordResetService interface は不変）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestPasswordResetService -count=1
```

---

## 挙動保存残件（旧 BE-refactor-v2 R系・OPENのみ）

以下6件は 2026-07-10 作成の `BE-refactor-v2.md`（R0〜R21・全項目挙動保存）のうち、2026-07-11 に現 HEAD で再実測して未着手（OPEN）と確認したものだけを移植した。R1〜R9・R16〜R21 は完了確認済みのため移植していない（詳細は `git log` 参照）。

### R10. G7-1が残した孤児メソッド CountWorkingStaffByReservationTypeID の削除

- **ID**: `dead-count-working-staff-singular`
- **重要度**: P3 / **工数目安**: S
- **対象ファイル**: internal/repository/reservation_type_occupation_repository.go（interface宣言:27, 実装:98）、同 `_test.go:171`、internal/service/reservation_type_service_test.go・liff_service_mock_test.go・reservation_type_service_occupation_test.go（mock 3箇所）
- **依存関係**: なし（単独削除可能）

**証拠(現HEAD実測: 2026-07-11)**

`grep -rn 'CountWorkingStaffByReservationTypeID\b' backend/internal --include='*.go' | grep -v _test.go` は repository のinterface宣言・実装のみがヒットし、本番呼び出し元は0件（`liff_service_availability.go:150`はバッチ版`CountWorkingStaffByReservationTypeIDs`（複数形）のみを呼ぶ）。G7-1（`755f3e42`）が本番呼び出しをバッチ版へ切替えた結果の孤児メソッドが、v2作成時点(2026-07-10)から未削除のまま現HEADにも残存している。

**対応方針**

interface宣言・実装・専用repositoryテスト（`reservation_type_occupation_repository_test.go:171`）・service層mock3箇所の該当メソッドを削除する。複数形`...TypeIDs`（バッチ版）は現役のため絶対に触らない。削除前に上記grepを再実行し単数形の残存参照が0件になることを確認する。

**検証コマンド(スコープ限定)**
```
grep -rn 'CountWorkingStaffByReservationTypeID\b' backend/internal --include='*.go' | grep -v '_test.go'
docker compose exec backend go test ./internal/repository/ -run TestReservationTypeOccupation -count=1
docker compose exec backend go test ./internal/service/ -run 'TestLiff' -count=1
```

### R11. lstep_csv_helpers.go の複合日時リテラル1件を time.DateTime へ統一

- **ID**: `lstep-csv-datetime-const`
- **重要度**: P3 / **工数目安**: XS
- **対象ファイル**: internal/service/lstep_csv_helpers.go:108
- **依存関係**: なし

**証拠(現HEAD実測)**

`csvDateFormats`（:105-111）の3番目の要素が `"2006-01-02 15:04:05"` のリテラル文字列のまま（Go 1.20+ の `time.DateTime` と完全一致）。隣接する `time.DateOnly`（:106）とは不統一。v2作成時点から未修正。

**対応方針**

`:108` のリテラルを `time.DateTime` に置換（値同一・挙動不変）。`:109-110` の真の複合レイアウトは触らない。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'Csv' -count=1
```

### R12. ConfigureTimeZone の JST 再導出を config.JST 再利用に統一

- **ID**: `configuretimezone-jst-reuse`
- **重要度**: P3 / **工数目安**: XS
- **対象ファイル**: internal/config/timezone.go:23-30
- **依存関係**: なし

**証拠(現HEAD実測)**

`ConfigureTimeZone()` が `time.LoadLocation(JapanTimeZone)` を独自に再実行しており、package init で panic-fail-fast 済みでキャッシュされている `JST` 変数（:14-20）を再利用していない。ドキュメントコメント（:11-14）が「各呼び出し箇所で再導出せずキャッシュ済み JST を使え」と明記した直後の関数がそれに反する。v2作成時点から未修正。

**対応方針（シグネチャ・戻り値型不変）**
```go
func ConfigureTimeZone() error {
    time.Local = JST
    return nil
}
```
`fmt` import が他で未使用になる場合は整理する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/config/ -count=1
```

### R13. G3-3未完遂分: 業務repoのUpdate系を updateScopedByID へ統合

- **ID**: `repo-update-helper-consolidation`
- **重要度**: P3 / **工数目安**: M
- **対象ファイル(12サイト)**: estimate_repository.go:81 / checkup_repository.go:149 / inventory_repository.go:72 / diagnosis_repository.go:70,202 / examination_repository.go:111 / permission_group_repository.go:72 / reservation_type_liff_repository.go:68 / reservation_staff_repository.go:87 / closing_special_period_repository.go:80 / shift_entry_repository.go:94 / reservation_repository.go:170 / accounting_repository.go:214
- **依存関係**: 保護テスト（accounting_complete_appointments_test.go 等）は現HEADで実装済み

**証拠(現HEAD実測: 2026-07-11)**

上記12サイトを実読した結果、全サイトが `updateScopedByID`（helpers.go:70）を未使用で、マスタ系19 repo（cage/checkup_type/consultation/exam_type/vaccine 等）で使われているのと同型の手書きボディ（clinicScope + Where(id) + Updates + FromGORM + RowsAffected==0→WrapNotFound）が残存する。v2作成時点(2026-07-10)から未着手のまま。

**対応方針**

各サイトを機械的にヘルパ呼び出しへ置換する。dbOrTx変種（inventory/examination/reservation_staff/shift_entry/accounting）は `updateScopedByID(dbOrTx(ctx, r.db), ...)` の形で ambient tx 参加を維持する。置換前に既存ボディとヘルパの挙動を①RowsAffected==0の扱い ②更新後再取得(FindByID)の有無とPreload ③Select/Omit句の有無 が完全一致するサイトのみ置換すること。除外（触ってはならない）: treatment/care_plan_item/clinical_plan（サブクエリ隔離）、billing_item（JOINスコープ）、medical_record_repository.goのUpdate（draft-status条件+Conflict変換）、reservation_type_liffのDelete（FK-conflict変換）、pet_chronic_condition_repository.go（本ファイル下部「別トラック残件」F3参照 — RowsAffected検査が無いため置換は挙動変更になる）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -count=1
```

### R14. G3-3未完遂分: 業務repoのDelete系を deleteScopedByID へ統合

- **ID**: `repo-delete-helper-consolidation`
- **重要度**: P3 / **工数目安**: S
- **対象ファイル(5サイト)**: estimate_repository.go:95 / checkup_repository.go:164 / inventory_repository.go:87 / medical_record_repository.go:238 / appointment_admin_repository.go:101（SoftDelete）
- **依存関係**: R13と同一ファイル群のため直後に実施推奨

**証拠・問題・対応方針**: R13と同一の統合・同一の手順・同一の除外リスト。**注意**: `medical_record_repository.go:238` の Delete は `Where("id = ? AND status = ?", id, model.MedicalRecordStatusDraft)` の追加述語を持つためヘルパ対象外の可能性が高い — 置換前に必ず実体を読むこと（現HEAD実測: draft条件は Update 側にも同型で存在することを確認済み）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -count=1
```

### R15. api.yaml: Billing スキーマに medical_record ネストプロパティを追記

- **ID**: `openapi-billing-medical-record-property`
- **重要度**: P3 / **工数目安**: XS
- **対象ファイル**: backend/docs/api.yaml（Billingスキーマ、現HEAD: :1472-1571 付近）
- **依存関係**: なし

**証拠(現HEAD実測)**

model `accounting.go` の `MedicalRecord *MedicalRecord json:"medical_record,omitempty"`（Preload時のみ直列化）に対応するプロパティが Billing スキーマに存在しない。owner/pet/items/payments/payment_splits/refunds は追記済みだが medical_record のみ欠落。

**対応方針**

Billing スキーマに `medical_record`（nullable, readOnly, description: Preload 時のみ）を、既存の owner/pet と同様の入れ子オブジェクトまたは `$ref` で追記する。実装コードは触らない。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/apicontract/ -count=1
```

---

## 別トラック残件（旧 BE-refactor-v2 §4・未着手・記録のみ）

以下は挙動変更または PO 判断を要するため本計画では実行しない。実行者はこの節に一切手を付けないこと。

| ID | 内容 | 現HEAD状態(2026-07-11実測) |
|---|---|---|
| G6-2 | repo内部tx13ファイルのdbOrTx化 + tx規約CLAUDE.md追記 | ブロッカーだった旧Appendix Aのtx非参加3件（X-9/X-10/finalize-child-write-race）はCLOSED済みだが、G6-2自体の対象13ファイルの特定・変換は未着手・未検証 |
| F3 | `pet_chronic_condition_repository.go:61,71` の Update/Delete に RowsAffected==0→NotFound 検査が無い（パッケージ内で唯一） | 現HEADでも検査なしのまま残存を確認。ヘルパ統合は「無言成功→NotFound」への挙動変更になるためPO判断が前提 |
| F6 | 死にコード群（`LstepTagService.BulkAddOwnerTag`、`SyncPetSpeciesTags`、`SyncSeniorTag`、`FindOwnersByCategoryPurchaseDate` 等）のkeep/delete判断 | 現HEADで全メソッドの存在を確認。Lstep Write API一時停止(2026-05)由来の意図的休眠の可能性がありオーナーの機能ロードマップ判断待ち |
| B-2 | Preload read-lint未登録の3マスタ | modelにGORM associationが無く構文的に登録不能（設計変更が前提）。対象ファイル未特定のため次回着手時に再調査要 |

（`G9-1`: main.go二段階DIの単一化は現HEADで完了確認済み — `cmd/api/main.go:95`「G9-1: 旧・二段階DIを単一段階に統合」。この表からは除外し、記録済み扱いとする。）
