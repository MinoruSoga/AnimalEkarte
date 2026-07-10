# BE-refactor.md — バックエンド リファクタリング計画（Appendix A / H フォローアップのみ残存）

> 残るのは Appendix A 5件（X-14〜X-18、P3中心）と、レビュー由来フォローアップ H-1〜H-7（別チケット化推奨）のみ。
> 本編（挙動保存トラック）および Appendix A の CLOSED 済み項目は本ファイルから削除済み（詳細は git log 参照）。
> X-9（resv-slot-phantom-toctou）・X-10（mr-version-check-not-atomic）・finalize-child-write-race は挙動変更トラックとして実装完了・CLOSED（詳細は git log 参照）。
> 前提: backend は 2026-07-02 完遂の D1-D13/R1-R3 計画で一度系統的にリファクタ済み・複数回の監査で well-maintained と判定済みのコードベースである。

## 監査の方法と信頼性

- 対象: `backend/` 配下全体。15次元で並列監査 → 敵対的検証。
- 本編（挙動保存）と Appendix A（挙動変更・別トラック）に分離。CLOSED 済みの完了履歴は本ファイルから削除済み。

## サマリー

| 区分 | 件数 | 内訳 |
|---|---|---|
| Appendix A（挙動変更・別トラック） | 5件 | X-14〜X-18 |
| レビュー由来フォローアップ（未登録・別チケット推奨） | 8件 | H-1, H-2, H-3, H-4, H-5, H-6, H-7, H-8 |

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

以下8件は監査で実在を確認した defect だが、修正すると HTTP レスポンス・DB書込結果・権限判定・API契約のいずれかが観測可能な形で変わる。このため本計画（挙動保存リファクタ）の実行対象からは外し、個別 Issue として起票のうえ別トラックで扱うことを推奨する。severity 順に記載。

### X-14. master-FK write allowlistのknown-unguarded約47エントリにisolation test不在(名簿上も『NO dedicated isolation test』と明記)

- **ID**: `test-known-unguarded-master-fk-isolation-tests`
- **重要度**: P2 / **工数目安**: L
- **対象ファイル**: internal/service/master_fk_write_inventory_lint_test.go (143-208)
- **依存関係**: 各エントリのガード実装(挙動変更)が先行必須。テストのみ先行するとCIがREDになる

**証拠(現HEAD検証済み)**

master_fk_write_inventory_lint_test.go:143-145「// statusKnownUnguarded: reviewed; NO dedicated isolation test confirms ownership\n statusKnownUnguarded masterFKWriteStatus = "known-unguarded"」。エントリ例(同:191-192)「{"accountingService.Update", statusKnownUnguarded, []string{"PaymentMethodID"}, "PaymentMethodID resolved via clinic system_key→ID map (resolvePaymentMethodMasterID); not a FindByID guard and no isolation test — verify rejection of explicit foreign IDs."},\n{"billingItemService.CreateItem", statusKnownUnguarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "all three FKs persisted directly without FindByID (billing_item_service.go:230)."}」。repository/CLAUDE.mdの規約は「正本ガード = 各サイトの runtime isolation test」だが、これらのエントリにはその正本ガードが存在しない。

**問題**

review網羅性gateは『名簿に載せる』ことしか担保せず(同ファイル冒頭に明記)、known-unguardedのまま滞留しているwrite経路はクロステナントmaster FK書き込みが実際に拒否されるかを誰も検証していない。会計ドメイン(accountingService.Update / billingItemService.CreateItem)を含む点が特に優先度が高い。テストを書くとRED(ガード自体が無い)になるため、これはガード実装を伴う挙動変更トラック。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

別トラック(挙動変更)として段階実施。優先順: (1)会計: accountingService.Update PaymentMethodID / billingItemService.CreateItem の3FK — service層にFindByID(clinicID,…)ガード追加後、internal/repository/または既存cross_tenant_master_fk_write_test.goパターンで『別クリニックFK指定→apperrors.WrapInvalidInput/NotFound拒否』テストを追加、allowlistエントリをguardedへ更新。(2)campaign TargetItemIDs(repo ReplaceTargets unscoped)。(3)carePlanItem HospitalizationPlanID / hospitalization CageID。(4)self-ref ParentID群(checkupType/consultation/examType)。各バッチはTestMasterFKWriteInventoryのstatus突合がgateになるため、allowlist更新漏れはCIで検出される。一括ではなくドメイン毎に1PRずつ、STGデータ監査(既存越境データの有無)を先行させること(R1-3の教訓)。

**進捗(Session 6, 2026-07-11時点)**

バッチ(3) carePlanItem HospitalizationPlanID / hospitalization CageID が完了。`carePlanItemService.validateMasterFKs` に `hospPlanRepo.FindByID(ctx, clinicID, HospitalizationPlanID)` を追加(medicine/procedureと同一エラークラス)、`hospitalizationService.Create/Update` に `repos.Cage.FindByID(ctx, clinicID, CageID)` を追加(CageID非nil時)。isolation test 4件追加: `TestCarePlanItemService_Create_RejectsCrossClinicHospitalizationPlanFK` / `TestCarePlanItemService_Update_RejectsCrossClinicHospitalizationPlanFK` / `TestHospitalizationService_Create_RejectsCrossClinicCageFK` / `TestHospitalizationService_Update_RejectsCrossClinicCageFK`(いずれも `internal/service/cross_tenant_master_fk_write_test.go`)。allowlist該当4エントリを `statusKnownUnguarded` → `statusGuarded` に更新。ガード実装を一時的に無効化した状態でこれら4テストがREDになることを確認済み(regressionを実際に検出する回帰テストであることの確認)。

**残件数の実測訂正**: 本セクション記載の見積り(約47)および過去に参照された「42」という数値は、直近の並行バッチ(X-11〜X-13等)によるallowlist変動を反映しておらずいずれも不正確だった。実測(`grep -c statusKnownUnguarded`)は本バッチ着手前 **46件**、本バッチ完了後 **42件**(-4、対象4エントリのみ)。優先順(2)campaign TargetItemIDs は既にallowlist上 `statusGuarded`(X-5で先行完了済み、本文の優先順メモが古い)。残る対象は主に(1)会計・(4)self-ref ParentID群および reservation/staff/treatment/trimming 等の残 known-unguarded 42件。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -count=1
docker compose exec backend go test ./internal/service/ -run '_Rejects' -count=1
```
(旧記載の `./internal/repository/ -run TestCrossTenantMasterFKWrite` はパス・テスト名とも現HEADと不一致だったため訂正。isolation testは `internal/service/cross_tenant_master_fk_write_test.go` にあり、対象テスト関数名は `TestCrossTenantMasterFKWrite` ではなく個々のサービス別 `Test*_*Rejects*` 群。)

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
