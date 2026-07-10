# BE-refactor.md — バックエンド リファクタリング計画（Appendix A / H フォローアップのみ残存）

> **本編（挙動保存トラック、G1/G2/G6/G9/G12/G13/G14 の全17件）は 2026-07-10 に完遂（Epic CLOSED）。**
> 残るのは Appendix A（挙動変更・別トラック、PO判断待ち）14件と、レビュー由来フォローアップ H-1〜H-7（別チケット化推奨）のみ。
> 前提: backend は 2026-07-02 完遂の D1-D13/R1-R3 計画で一度系統的にリファクタ済み・複数回の監査で well-maintained と判定済みのコードベースである。

## 監査の方法と信頼性

- 対象: `backend/` 配下全体。15次元で並列監査 → 敵対的検証。
- 本編（挙動保存）と Appendix A（挙動変更・別トラック）に分離。CLOSED 済みの完了履歴は本ファイルから削除済み。

## サマリー

| 区分 | 件数 | 内訳 |
|---|---|---|
| 本編（挙動保存） | **0件 — CLOSED 2026-07-10** | 全17件完遂（詳細は本ファイル履歴 / git log 参照） |
| Appendix A（挙動変更・別トラック） | 14件 | X-1, X-3〜X-5, X-9〜X-18 |
| レビュー由来フォローアップ（未登録・別チケット推奨） | 7件 | H-1, H-2, H-3, H-4, H-5, H-6, H-7 |

## 本編 CLOSED 履歴（2026-07-10）

```
G12-1 → G12-2 → G13-1 → G14-1: 2026-07-10 実装・CLOSED（allModels exhaustiveness gate / item_source enum parity gate /
  lab import 補償ログ / accountingService.Update allowlist精度修正）
G1-1 / G1-3 / G1-4 / G2-1 / G2-2: 2026-07-09 の前セッションで既に CLOSED 済みだったと判明（本ファイル未更新のstaleな記載を本セッションで是正）
G1-2: 前セッションで 203→82 まで既完了、残差63件中48件を2026-07-10に追加完了。残る15件（14件は同一ハンドラのエイリアス
  重複登録・1件はPOST /api/line/webhookのパース制約）は internal/apicontract/openapi_route_drift_test.go の
  allowlist にコメント付きでpin。CLOSED
G1-5 / G1-6: 2026-07-10 実装・CLOSED（stale prototype docs 削除 / onboarding docs 是正）
X-2 / X-6 / X-7 / X-8（Appendix A、G6-2/G9-1 の BLOCKED 解消を目的にユーザー承認済みスコープとして本 Run に含めた）:
  2026-07-10、RED→GREEN 実証 + security-reviewer 独立レビュー(CRITICAL/HIGH 0件)を経て CLOSED
G6-2 → G9-1: BLOCKED 解消後 2026-07-10 に実装・CLOSED（tx機構の残り9ファイル一括置換+CLAUDE.md規約追加、
  main.go 二段階DI統合）。G9-1 は go-reviewer 独立レビュー Approve。
```

### レビュー由来フォローアップ（本編未登録）

| ID | 内容 | 発見元 | 優先度 |
|---|---|---|---|
| H-1 | `UpdateStaffGroups` の staff_id 単位 DELETE が多施設所属スタッフの他クリニックグループ紐付けを意図せず削除しうる | G11-1 security-reviewer | HIGH — 別チケット化推奨 |
| H-2 | `UpdateExcludedReservationTypes`（reservation_staff_repository.go）の DELETE が `staff_id` のみでスコープされ `clinic_id` を含まない一方、INSERT は呼び出しクリニックの型IDのみ。`staff_reservation_exclusions` テーブル自体に `clinic_id` 列が無いため、多施設所属スタッフに対しては clinic A の正当な操作が clinic B の除外設定行を無警告で全削除する（H-1 と同型のクロステナント破壊）。兄弟の `UpdateReservationCapabilities`/`staff_reservation_capabilities` は自前 `clinic_id` 列を持ち `Where("clinic_id = ? AND staff_id = ?")` で正しくスコープされており非対称。 | G11-4 security-reviewer（`UpdateReservationCapabilities` との比較監査で発見） | HIGH — 別チケット化推奨（`staff_reservation_exclusions` への `clinic_id` 列追加 or DELETE を真の差分更新へ変更、要 migration） |
| H-3 | `billing_items.category` に索引が無く、`FindOwnersByCategoryPurchaseDate`（Lstep FEAT-383 配信ターゲティング、バッチ/cron想定）が `category = ?` 述語 + `billings` join で Seq Scan リスク。テーブル成長に伴い悪化。既存索引は `merchandise_item_id`/`treatment_id`/`appointment_id`/`trimming_course_id`/`trimming_option_id`/`billing_id`/`deleted_at` のみで `category` は対象外。`idx_billings_clinic_completed_at` も `WHERE status='completed'` の部分索引でこの3クエリ（status述語なし）はカバーしない。 | G11-5 database-reviewer | MEDIUM（パフォーマンス、要 migration・別チケット推奨: `CREATE INDEX idx_billing_items_category ON billing_items(category) WHERE deleted_at IS NULL`） |
| H-4 | `audit_logs.clinic_id` が Go では `*uint64`（nil許容）だが DDL では `bigint NOT NULL REFERENCES clinics(id)`。Go 側の nil 許容がコンパイル時には検出されない実行時 NULL 制約違反クラス（INSERT 失敗）を許す。 | G12-1 schema_drift nullability check（新設） | MEDIUM（要 migration or model 修正・別チケット推奨: audit_service.go の validateAuditLog が実運用上 nil を弾いている前提を型で保証するか、DB 側制約と model を整合） |
| H-5 | `lstep_csv_imports.uploaded_by_user_id` が Go では `*uint64`（nil許容）だが DDL では `bigint NOT NULL REFERENCES accounts(id)`。H-4 と同型のクラス。 | G12-1 schema_drift nullability check（新設） | MEDIUM（要 migration or model 修正・別チケット推奨） |
| H-6 | `backend/CODING_RULES.md` の §3.2/§5.1/§5.4/§6.1/§6.3 に、G1-6 で是正した README.md と同型の forbidden-pattern 教材コード（生の `gin.H{"error":...}` レスポンス、`uuid.UUID` ベースの `FindByID` シグネチャ例 — 実際は全モデル `uint64` PK、sentinel-error `errors.Is` 例示で `apperrors.FromGORM`/`RespondError` 未使用）が残存。§6 に `RequirePermission`/P5 ルートゲーティングの言及が一切ない。G1-6 の対象範囲（ディレクトリツリーのみ）を超える約400行規模の書き直しのため別ユニット化推奨。 | G1-6 実装エージェント | MEDIUM（オンボーディング文書の質・別チケット推奨） |
| H-7 | `reservationStaffService.Update` の所有権確認読み取り(`s.GetByID`)が tx 外で行われ、確認〜更新の間にスタッフが削除されると TOCTOU の窓が生じる。X-8 の修正対象（fields 更新+除外設定置換の原子性）とは独立した既存の設計であり、X-8 は悪化させていない（security-reviewer 確認済み）。低頻度の管理操作のため実害は限定的。 | X-8 security-reviewer | LOW（別チケット化検討・優先度低） |

---

## Appendix A: 挙動変更を伴う項目（別トラック・PO/責任者判断を要する）

以下14件は監査で実在を確認した defect だが、修正すると HTTP レスポンス・DB書込結果・権限判定・API契約のいずれかが観測可能な形で変わる。このため本計画（挙動保存リファクタ）の実行対象からは外し、個別 Issue として起票のうえ別トラックで扱うことを推奨する。severity 順に記載。

（X-2 lstep-nilcipher-stale-di / X-6 tx-medicine-inventory-nonparticipation / X-7 tx-clinic-create-nonparticipation / X-8 tx-reservation-staff-nonparticipation は 2026-07-10、G6-2/G9-1 の BLOCKED 解消を目的にユーザー承認済みスコープとして RED→GREEN 実証 + security-reviewer 独立レビュー(CRITICAL/HIGH 0件)を経て CLOSED。）

**特に優先度が高い1件（P1・データ破損系）**:
- `X-1 sanitize-multipart-binary-corruption`: カルテ画像・共有ファイルアップロードのバイナリが保存時に破壊される可能性（2026-03-31 導入以来のクラス）。

**マルチテナント書込ガード欠落（P1・#124/#125 と同型）**:
- `X-4 billing-item-trimming-fk-unguarded` / `X-5 campaign-target-item-fk-unguarded`

**監査ログ整合性（P1）**:
- `X-3 audit-ip-inet-model-drift`: 本番DDL(inet列)とmodel(string)の型不一致により、複数の監査ログ書込経路が実行時に失敗しうる。
### X-1. SanitizeNullBytesがmultipart/form-dataを含む全POST/PUT/PATCHボディにbytes.Mapを適用しバイナリアップロードを破壊する

- **ID**: `sanitize-multipart-binary-corruption`
- **重要度**: P1 / **工数目安**: S
- **対象ファイル**: internal/middleware/sanitize_null_bytes.go (25-70); cmd/api/main.go (286-287); internal/handler/medical_record_image_handler.go (124)

**証拠(現HEAD検証済み)**

internal/middleware/sanitize_null_bytes.go:26-31:
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPatch && method != http.MethodPut {
			c.Next()
			return
		}
（Content-Type による除外は一切ない）

sanitize_null_bytes.go:47-58:
		sanitized := bytes.Map(func(r rune) rune {
			switch {
			case r == 0x00: // NULL バイト
				return -1
			...
			case r >= 0x0E && r <= 0x1F: // SO〜US
				return -1

cmd/api/main.go:286-287:
	// BUG-067: POST/PATCH/PUT ボディから NULL バイトを除去（PostgreSQL エラー防止）
	r.Use(middleware.SanitizeNullBytes())

internal/handler/medical_record_image_handler.go:124: file, fileHeader, err := c.Request.FormFile("file")

frontend/src/features/medical-records/api/medical-record-images.ts:17-22 は new FormData() + axios.post(multipart/form-data) で生Fileを送信する。

**問題**

bytes.Map はボディを UTF-8 コードポイント列として解釈する。バイナリ（JPEG の 0xFF 0xD8、PNG シグネチャ等）は不正 UTF-8 として各バイトが U+FFFD（3バイト EF BF BD）に置換され、さらに PNG シグネチャの 0x1A は除去レンジ（0x0E-0x1F）に該当してドロップされる。カルテ画像アップロード（POST /medical-records/:id/images/upload）、共有ファイル（shared_file_handler.go:48）、Lステップ CSV インポート（Shift-JIS の場合）が全てこのグローバル middleware（main.go:287、ルート登録前に Use）を通過するため、保存されるバイナリは原理的に原本と一致しない。臨床記録画像の破壊 = データ損失リスク。middleware は 2026-03-31（0234bf8a）、画像ハンドラは 2026-04-11 以降に追加されており、導入以来この経路にバイナリ忠実性のテストは存在しない（sanitize_null_bytes_test.go は application/json のみ）。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) RED test 先行 — internal/middleware/sanitize_null_bytes_test.go に Content-Type: multipart/form-data で PNG シグネチャ [0x89,0x50,0x4E,0x47,0x0D,0x0A,0x1A,0x0A] を含むボディを通し、バイト完全一致を検証するケースを追加（現実装で FAIL することを確認 = 破壊の実証）。(2) SanitizeNullBytes 冒頭に Content-Type ガードを追加: strings.HasPrefix(ct, "application/json") の場合のみサニタイズし、それ以外（multipart/form-data, application/octet-stream 等）は素通しする（BUG-067 の対象は JSON テキストのため意図保存）。(3) 既存 JSON ケースのテストが GREEN のままであることを確認。(4) 実機確認として staging で画像アップロード→ダウンロードのバイト一致を手動検証（破壊が確認された場合、既存アップロード済み画像の破損調査は別トラック起票）。影響範囲: multipart を使う3エンドポイントのアップロード忠実性。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/middleware/ -run TestSanitizeNullBytes -count=1
```

### X-3. audit_logs.ip_address が DDL では inet、model では string — 空文字書込は実DDLで INSERT 失敗し、AutoMigrate製テストスキーマがそれを隠蔽

- **ID**: `audit-ip-inet-model-drift`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/model/audit_log.go (9-27); migrations/001_init.sql (2543-2563); internal/service/audit_service.go (171-296); internal/service/examination_service.go (341-364); internal/repository/audit_repository_test.go (22-29)
- **依存関係**: drift-test-nullability-gap と同根（あちらは検出網、こちらは実体修正）。同時に着手する場合は本件の real-DDL テストを先に入れる

**証拠(現HEAD検証済み)**

DDL (migrations/001_init.sql:2545,2553): `clinic_id    bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,` / `ip_address   inet         NULL,`。model (internal/model/audit_log.go:11,22): `ClinicID   *uint64         \`json:"clinic_id"\`` / `IPAddress string          \`json:"user_agent"\`の直上、`IPAddress string          \`json:"ip_address"\``（gormタグ無し=INSERTに常に含まれる）。IPAddress を設定しない書込経路が多数: audit_service.go:184-193 LogLstepOperationWithMetadata / :209-219 LogMedicalRecordChange / :236-247 LogVitalChange / :285-295 LogAddendumCreate、および fail-closed tx 監査 examination_service.go:345-359 の LogEntryTx（AuditLogInput に IPAddress 未設定）。PostgreSQL では `''::inet` は 22P02 invalid input syntax であり、空文字パラメータの inet 列への INSERT は失敗する。一方テストは audit_repository_test.go:25 `require.NoError(t, db.AutoMigrate(&model.AuditLog{}))` で string→text 列のスキーマを自前生成するため通過し、同 :56 `t.Run("clinic_id / actor_id が nil のシステム操作でも作成できる", ...)` は実DDLの clinic_id NOT NULL では不可能な挙動をテストとして固定している（ClinicID=nil は service 層 validateAuditLog audit_service.go:94-96 が拒否するため運用上は到達しないが、テストスキーマと実DDLの乖離の証左）。

**問題**

監査ログ書込の実行時整合が実DDLに対して未検証。ip_address が実DDL通り inet なら、(a) ベストエフォート監査（LSTEP/カルテ/バイタル/追記/薬量逸脱等）は slog.Warn で握り潰され監査証跡がサイレント消失（例: line_link_service.go:182-184）、(b) fail-closed tx 監査（#211 exam/checkup replace）は監査INSERT失敗→tx rollback→PUT が 500 になる。552本のテストは ekarte_db_test（AutoMigrate製・ip_address=text/clinic_id nullable）で走るため、この乖離クラスをどのテストも検出できない。医療システムの監査証跡要件（audit_logs は削除禁止テーブル）に直結する。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) まず実挙動を確定する再現テストを追加: internal/repository/audit_real_ddl_test.go を新設し、checkup_field_cascade_test.go と同じ手法で 001_init.sql の audit_logs DDL 原文（inet/NOT NULL含む）からテーブルを再作成（setupIsolatedTestDB 使用）し、IPAddress="" の repo.Create が失敗するか検証する（RED想定）。(2) 失敗が確認されたら bug トラックへ: model.AuditLog.IPAddress を `*string` 化し buildAuditLog（audit_service.go:74-88）で空文字→nil 正規化（inet 列へは NULL が入る）。LogAuthLogin/LogClinicSwitch の string 引数シグネチャは維持し内部で変換。JSON応答互換が必要なら handler/response 変換で nil→"" を維持（現状 audit_logs の read API は無く影響は限定的 — audit_repository.go は Create/CreateTx のみ）。(3) ClinicID には `gorm:"not null"` を付与しテストスキーマを実DDLの NOT NULL に整合させ、audit_repository_test.go の nil-clinic テストは「repo単体は通るが実DDLでは不可」という現状固定をやめ、real-DDL テスト側で NOT NULL 違反を検証する形に書き換える。(4) 既存 AutoMigrate ベースの setupAuditTestDB は real-DDL 版に置換。影響範囲: model/audit_log.go・service/audit_service.go・repository/audit_repository_test.go のみ（migration 不要 — DDL 側は正とする）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestAuditRepository|TestAuditLogRealDDL' -count=1 && docker compose exec backend go test ./internal/service/ -run TestAuditService -count=1
```

### X-4. billingItemService.CreateItem が trimming_course_id/trimming_option_id を未検証で永続化

- **ID**: `billing-item-trimming-fk-unguarded`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/service/billing_item_service.go (167-231); internal/service/master_fk_write_inventory_lint_test.go (192)

**証拠(現HEAD検証済み)**

billing_item_service.go:220-231: `item := &model.BillingItem{ ... TrimmingCourseID: input.TrimmingCourseID, TrimmingOptionID: input.TrimmingOptionID, ... }` の直前に billing のテナント所有権確認 (`s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID)`) はあるが、TrimmingCourseID/TrimmingOptionID 自体の所有権検証は無い。master_fk_write_inventory_lint_test.go:192 が自己申告で `{"billingItemService.CreateItem", statusKnownUnguarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "all three FKs persisted directly without FindByID (billing_item_service.go:230)."}` と記録済み。

**問題**

P3.1 write側の原則（『正本は各サイトのruntime isolation test』）に反し、他クリニックの trimming_course_id/trimming_option_id を明細に紐付け可能。billing_item は clinic_id を自前で持たず billings 経由スコープのため、書込時点で検証しないと以後の CountUsageByTrimmingCourseID 等のクロステナント集計汚染・データ整合性破壊につながる（#124/#125 と同型のクラス）。read側 Preload は model に TrimmingCourse/TrimmingOption 関連が定義されていないため直接の情報漏洩は未確認だが、write-side 整合性違反は独立した問題。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

billing_item_service.go の CreateItem で input.TrimmingCourseID/TrimmingOptionID が非nilの場合、既存の trimmingCourseRepo.FindByID(ctx, input.ClinicID, *id) / trimmingOptionRepo.FindByID 相当（未配線なら repository に追加）で所有権検証してから item に設定する（examinationService.ReplaceItems の #124 修正パターンに倣う）。cross_tenant_master_fk_write_test.go に `TestBillingItemService_CreateItem_RejectsCrossClinicTrimmingFK` を追加し、GREEN化後に allowlist の該当行を `statusGuarded` へ更新する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestBillingItemService -v
```

### X-5. campaignService.Create/Update が campaign_target_items.merchandise_item_id を未検証で永続化

- **ID**: `campaign-target-item-fk-unguarded`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/service/campaign_service.go (83-89,204-233,235-283); internal/repository/campaign_repository.go (40-70)

**証拠(現HEAD検証済み)**

campaign_service.go:83-89 `buildCampaignTargetItems(itemIDs []uint64) []model.CampaignTargetItem { ... targets = append(targets, model.CampaignTargetItem{MerchandiseItemID: id}) ... }` は id の所有権を検証せず、Create(224行目 `TargetItems: buildCampaignTargetItems(input.TargetItemIDs)`)・Update(272行目 `s.repo.ReplaceTargets(ctx, id, cats, itemIDs)`) の双方で無検証のまま merchandise_item_id を書き込む。model.CampaignTargetItem（campaign.go:52-58）は clinic_id 列を持たない純粋な junction。

**問題**

他クリニックの merchandise_item_id をキャンペーン対象に紐付け可能（テナント越境ID混入）。CalculateItemCampaignDiscount のマッチングロジックが誤発火し、割引適用の整合性が壊れる。read Preload(`Preload("TargetItems")`)自体は merchandise item 本体を展開しないため直接の名称/価格漏洩はないが、write-side 整合性違反として P3.1 の対象クラスに該当する。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

campaignService.Create/Update で input.TargetItemIDs の各要素を merchandiseItemRepo.FindByID(ctx, clinicID, id)（または既存の CountByIDs 一括検証）でクリニック所有権を確認してから buildCampaignTargetItems に渡す。repository.CampaignRepository に検証済みIDのみ受け付ける契約を明記し、cross_tenant_master_fk_write_test.go に isolation test を追加後、allowlist の 2 行を `statusGuarded` に更新する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestCampaignService -v
```

### X-9. 予約枠競合チェックの SELECT FOR UPDATE は空枠のファントム挿入を防げず、同時予約で重複/超過予約が成立する

- **ID**: `resv-slot-phantom-toctou`
- **重要度**: P2 / **工数目安**: M
- **対象ファイル**: internal/repository/reservation_repository.go (294-316); internal/service/reservation_capacity.go (33-59); internal/service/reservation_service.go (208-221); internal/service/reservation_validators.go (92-139); migrations/001_init.sql (2692-2695)
- **依存関係**: なし。dbortx_inventory_lint_test.go の allowlist 更新を同一コミットで行うこと

**証拠(現HEAD検証済み)**

reservation_repository.go:301-309 `err := dbOrTx(ctx, r.db).Raw(`\n\t\tSELECT id FROM appointments\n\t\tWHERE clinic_id = ?\n\t\t  AND deleted_at IS NULL\n\t\t  AND status NOT IN ('cancelled')\n\t\t  AND start_time < ?\n\t\t  AND end_time > ?\n\t\t  AND (? = 0 OR id != ?)\n\t\tFOR UPDATE`` — FOR UPDATE は既存行のみロックし、条件に合致する行が 0 件（空枠）の場合は何もロックしない。型別容量は reservation_capacity.go:51 `count, err := reservationRepo.CountByTypeAndStartTime(ctx, clinicID, reservationTypeID, startTime, excludeID)` の素の COUNT（FOR UPDATE すら無し、reservation_repository.go:318-329）。DB バックストップは migrations/001_init.sql:2693-2695 `CREATE UNIQUE INDEX uk_appointment_staff_time ON appointments (clinic_id, doctor_id, start_time) WHERE deleted_at IS NULL AND status != 'cancelled';` のみで、(a) doctor_id が NULL の容量枠予約、(b) 同一医師の start_time が異なる時間重複（10:00-10:30 vs 10:15-10:45）、(c) MaxConcurrent 上限の型別容量は表現できない。reservation_service.go:208-210 のコメント「SELECT FOR UPDATE + トランザクションで競合を防止」は空枠ケースでは成立しない。

**問題**

check-then-act（競合カウント→INSERT）で、READ COMMITTED のスナップショット外で並行 tx が INSERT した行は数えられずロックもできない（ファントム）。空き枠に対する同時 2 リクエスト（LINE 予約は不特定多数から到達）が双方 0 件/上限未満と判定し双方 INSERT → 医師二重予約・出勤医師数超過・予約区分 MaxConcurrent 超過が成立する。既存 FOR UPDATE は「既存行がある場合の直列化」しか提供せず、コメントが主張する防止保証と実装が乖離している。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

1) ReservationRepository に `AcquireBookingLock(ctx context.Context, clinicID uint64) error` を新設し `SELECT pg_advisory_xact_lock(hashtextextended('appointments:' || ?::text, 0))`（clinic 単位のトランザクションスコープ advisory lock、dbOrTx 参加）を実装。2) reservation_service.go の Create(WithTx 内先頭・enforceBookingConstraints 時のみ)・updateWithConflictCheck(WithTx 内先頭)、reservation_validators.go ValidateAndCreate(WithTx 内先頭) の 3 箇所で競合チェック前に取得。3) dbortx_inventory_lint_test.go の allowlist に新メソッドを登録。4) 並行性テストを internal/repository/ に追加（2 goroutine が同一空枠を同時予約し、ちょうど 1 件だけ成功することを実 DB で検証。ltv_repository_test.go:1562 の advisory lock 利用が参考）。clinic 単位ロックで粒度が粗い懸念は、予約書込レートが clinic あたり低頻度のため許容（過剰な粒度細分化は不要）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestReservation|TestDBOrTxInventory' -count=1 && docker compose exec backend go test ./internal/service/ -run TestReservation -count=1
```

### X-10. カルテ楽観ロックの version 照合が UPDATE の WHERE に含まれず、並行 draft 編集で lost update が成立する

- **ID**: `mr-version-check-not-atomic`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/service/medical_record_crud.go (214-264); internal/repository/medical_record_repository.go (236-250)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

medical_record_crud.go:230-233 「// version が指定されている場合は一致確認\n\tif input.Version != nil && existing.Version != *input.Version {\n\t\treturn nil, apperrors.WrapConflict("他のユーザーがこのカルテを変更しました。再読み込みしてください")\n\t}」→ 同 254 行 `fields["version"] = existing.Version + 1` → repo 側 medical_record_repository.go:237-241 「result := r.db.WithContext(ctx).\n\t\tModel(&model.MedicalRecord{}).\n\t\tScopes(clinicScope(clinicID)).\n\t\tWhere("id = ? AND status = ?", id, model.MedicalRecordStatusDraft).\n\t\tUpdates(fields)」— WHERE は id+clinic+status のみで version 述語が無い。service 層の一致確認は tx 外の FindByID(216 行) 読取に対する check-then-act。

**問題**

2 名のスタッフが同一 version=N を読んで同時に更新すると、両者とも service 層チェックを通過し、両 UPDATE が WHERE(id, status=draft) にマッチして成功する。後勝ちがフィールドを黙って上書きし、かつ両者とも version=N+1 を書くため、以後のクライアントも競合を検出できない（楽観ロックの目的である lost update 防止が並行時に機能しない）。status=draft 述語は finalize 済みへの編集は原子的に防いでいるが、draft 同士の競合は防げない。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

1) medical_record_repository.go の Update シグネチャに `expectedVersion *uint64`（または fields とは別引数）を追加し、非 nil 時に `Where("version = ?", *expectedVersion)` を付与。2) RowsAffected==0 の場合、tx 内で status を再読して「not draft」（既存 Conflict メッセージ維持）と「version 不一致」（既存の『他のユーザーがこのカルテを変更しました』メッセージ）を区別して返す。3) medical_record_crud.go:214-264 の Update から service 層 version 照合を repo 述語に一本化（input.Version==nil の従来挙動＝照合なしは保存）。4) 並行 2 goroutine で同一 version から同時更新し片方だけ成功するテストを internal/repository/ に追加。呼出面: medical_record repo Update の他呼出箇所のシグネチャ追随が必要（grep で確認のこと）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestMedicalRecord -count=1 && docker compose exec backend go test ./internal/service/ -run TestMedicalRecord -count=1
```

### X-11. カルテ確定ロック(HC-003/005/006)の親 status チェックが子エンティティ書込と非原子で、確定と同時の子追加/編集が確定済カルテに混入しうる

- **ID**: `finalize-child-write-race`
- **重要度**: P3 / **工数目安**: L
- **対象ファイル**: internal/service/treatment_service.go (220-303); internal/service/examination_service.go (152-160); internal/service/vital_service.go (106-113); internal/service/prescription_service.go (83-84); internal/service/checkup_field_result_service.go (128-135); internal/repository/medical_record_repository.go (236-248)
- **依存関係**: resv-slot-phantom-toctou と同じく dbortx_inventory_lint_test.go allowlist 更新を伴う。5 サービス横断のため実装は 2 コミット以上に分割推奨

**証拠(現HEAD検証済み)**

treatment_service.go:222-229 「parent, err := s.repos.MedicalRecord.FindByID(ctx, clinicID, medicalRecordID)\n\t...\n\tif parent.Status == model.MedicalRecordStatusFinalized {\n\t\treturn nil, apperrors.WrapConflict("確定済みカルテには治療を追加できません")\n\t}」— 素の FindByID（無ロック・tx 外）でチェックした後、242 行 `err = s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {` の別 tx で子 INSERT。examination_service.go:159-160・vital_service.go:112-113・prescription_service.go:83-84・checkup_field_result_service.go:134-135 も同型（FindByID→status チェック→書込、親行ロックなし）。finalize 側 medical_record_repository.go:240 の `Where("id = ? AND status = ?", id, model.MedicalRecordStatusDraft)` はカルテ本体のみ原子化し、子テーブルには波及しない。

**問題**

T1(子追加) が parent.Status=draft を確認 → T2(finalize) がコミット → T1 の子 INSERT がコミット、の順序が可能で、確定済み（改変ロック済み）カルテに監査上 finalize 後の子レコードが追記なし(addendum 経由でなく)で混入する。確定ロックは臨床データの改竄追跡性の要の不変条件（HC-003/005/006・EXAM-001）だが、その並行時の強制が check-then-act のみ。競合窓はミリ秒級で発生頻度は低いものの、発生時は silent で検出手段がない。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラック。1) medical_record_repository.go に `LockDraftByID(ctx, clinicID, id) (*model.MedicalRecord, error)`（dbOrTx + `Clauses(clause.Locking{Strength: "UPDATE"})` + status 返却）を新設し dbortx lint allowlist に登録。2) 子書込 5 サービス（treatment/examination/vital/prescription/checkup_field_result — treatment は既存 repos.Transaction 内へ、他は Transactor.WithTx 導入）で、tx 内先頭に LockDraftByID → finalized なら既存と同一メッセージの Conflict を返し、子書込を同一 tx 内へ移す。finalize 側 UPDATE は同一行への行ロックを要求するため、子 tx 保持中の finalize は自然に待機し順序整合する（finalize 側の変更は不要）。3) 並行テスト（finalize と子追加を同時実行し、finalize 後の子混入ゼロを検証）を 1 サービス分（treatment）追加し、残りはパターン踏襲。段階導入可: まず treatment/examination（金額・検査値を持つ高リスク 2 系統）のみでも価値がある。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestTreatment|TestExamination' -count=1 && docker compose exec backend go test ./internal/repository/ -run TestDBOrTxInventory -count=1
```

### X-12. 会計 Create/Update の会計完了時 appointment 完了化が tx 外で、部分コミット（billing 確定済み・予約カード残留/エラー返却）が起こる

- **ID**: `billing-complete-appt-post-tx`
- **重要度**: P3 / **工数目安**: M
- **対象ファイル**: internal/service/accounting_service_core.go (80-86, 195-200, 251-267); internal/repository/accounting_repository.go (312-349)
- **依存関係**: dbortx_inventory_lint_test.go allowlist 更新を同一コミットで行うこと

**証拠(現HEAD検証済み)**

accounting_service_core.go:195-198 「if input.Status != nil && *input.Status == model.BillingStatusCompleted {\n\t\tif err := s.completeAccountingAppointments(ctx, input.ClinicID, accounting); err != nil {\n\t\t\treturn nil, apperrors.Wrap(err, "failed to complete accounting appointments during update")\n\t\t}」— 148-182 行の WithTx（fields/payment/splits/監査を R1-2 で原子化済み）がコミットした後に、tx 外の ctx で呼ばれる。repo 側 accounting_repository.go:317 `result := r.db.WithContext(ctx).` / 333 行も同じく r.db 直参照で dbOrTx 非参加。Create 側も同型: accounting_service_core.go:80-83 「if billing.Status == model.BillingStatusCompleted {\n\t\tif err := s.completeAccountingAppointments(ctx, input.ClinicID, billing); err != nil {\n\t\t\treturn nil, apperrors.Wrap(err, "failed to complete accounting appointments during create")」— billing Create(73 行) コミット後に失敗するとエラーを返すが billing は残る。

**問題**

会計確定は WithTx でコミット済みなのに appointment 完了化が失敗すると、(a) 呼出元にはエラーが返り操作全体が失敗に見えるが billing は completed で確定済み（Update 側）、(b) Create 側は billing が残ったままエラーになり、medical_record_id NULL のトリミング/手動会計ではリトライで二重 billing を作れる（idx_billings_medical_record_id_unique は medical_record_id 非 NULL のみバックストップ）。R1-2 が塞いだ「三系統分裂の部分コミット」と同型の残余が、同一ユースケースの後段に残っている。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラック（失敗時の原子性が変わる）。1) accounting_repository.go の CompleteAccountingAppointments 内 2 箇所（317, 333 行）を dbOrTx(ctx, r.db) に変更し dbortx lint allowlist に登録。2) Update 側: 呼出を WithTx クロージャ末尾（logPostCloseEdit の後）へ移動し txCtx で呼ぶ。判定は input.Status ベースに変更（現在は再読後 accounting を使うが、tx 内では fields 適用済みのため同値）。3) Create 側: repo.Create + completeAccountingAppointments を Transactor.WithTx で括る（Billing Create は既に単文なので repo 変更不要、Create が dbOrTx 未参加なら参加化）。4) syncCPMStageTag は外部 LSTEP 同期なので従来どおり tx 外 best-effort を維持。5) accounting_repository_tx_atomicity_test.go に「appointment 完了化失敗で billing 更新もロールバックする」ケースを追加。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestAccounting.*Atomicity|TestDBOrTxInventory' -count=1 && docker compose exec backend go test ./internal/service/ -run TestAccounting -count=1
```

### X-13. SharedFile.DeletedAt が *time.Time のため repo.Delete が物理 DELETE — DDL/読取述語/ログ文言のソフトデリート意図と不整合

- **ID**: `sharedfile-harddelete-vs-softdelete-intent`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/model/shared_file.go (18); internal/repository/shared_file_repository.go (62-71); internal/service/shared_file_service.go (133-175); migrations/001_init.sql (1250-1272)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

model (shared_file.go:18): `DeletedAt  *time.Time \`gorm:"index"          json:"deleted_at"\`` — gorm.DeletedAt でないため GORM のソフトデリートは発火せず、shared_file_repository.go:62-66 の `Delete(&model.SharedFile{})` は物理 DELETE を発行する。一方で意図はソフトデリート: DDL は deleted_at 列を持ち（001_init.sql:1262）部分インデックス `WHERE deleted_at IS NULL`（:1269-1271）を張り、読取は全て手動述語 `Where("deleted_at IS NULL")`（repo :41,:53,:76）、service のエラーログは `"failed to soft-delete shared file"`（shared_file_service.go:144）/`"failed to soft-delete expired shared file"`（:171）と明記。同型の LstepTagCodeMapping は SoftDelete メソッドで `Update("deleted_at", now)` を明示実装しており（lstep_tag_code_mapping_repository.go:81-98）、SharedFile だけ意図と実装が食い違う。

**問題**

LINE個別送信ファイルのメタデータ（誰がどの飼主向けに何をアップロードしたか）が削除・期限切れクリーンアップ時に物理消去され、deleted_at 列・部分インデックス・読取述語が全て死んでいる。監査・追跡可能性の設計意図（業務データは deleted_at を持つ: migrations/CLAUDE.md 必須チェック）に反する。ソフトデリート化は挙動変更（行が残る・一意/容量特性が変わる）のため別トラック。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

behaviorChange トラック。選択肢を PO 判断可能な形で提示: 案A（意図に合わせる）= DeletedAt を gorm.DeletedAt に変更（json:"-" 化で API から deleted_at フィールドが消える点は要確認 — 現状 json:"deleted_at" を露出）。repo の手動 `deleted_at IS NULL` 述語は GORM 自動述語と重複するが無害なので残置可。FindExpired のクリーンアップが物理削除を意図するなら Unscoped().Delete を明示。案B（実装に合わせる）= ハードデリートを正と決め、model から DeletedAt を除去し新規 migration で列と部分インデックスを drop、service のログ文言を hard-delete に修正。どちらでも「意図と実装の一致」を repository テスト（Delete 後に Unscoped 検索で行の有無を検証）で固定する。影響範囲: model/shared_file.go・repository/shared_file_repository.go・service/shared_file_service.go（案Bのみ migrations 追加）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestSharedFile -count=1 && docker compose exec backend go test ./internal/service/ -run TestSharedFile -count=1
```

### X-14. master-FK write allowlistのknown-unguarded約47エントリにisolation test不在(名簿上も『NO dedicated isolation test』と明記)

- **ID**: `test-known-unguarded-master-fk-isolation-tests`
- **重要度**: P2 / **工数目安**: L
- **対象ファイル**: internal/service/master_fk_write_inventory_lint_test.go (143-208)
- **依存関係**: 各エントリのガード実装(挙動変更)が先行必須。テストのみ先行するとCIがREDになる

**証拠(現HEAD検証済み)**

master_fk_write_inventory_lint_test.go:143-145「// statusKnownUnguarded: reviewed; NO dedicated isolation test confirms ownership\n statusKnownUnguarded masterFKWriteStatus = "known-unguarded"」。エントリ例(同:191-192)「{"accountingService.Update", statusKnownUnguarded, []string{"PaymentMethodID"}, "PaymentMethodID resolved via clinic system_key→ID map (resolvePaymentMethodMasterID); not a FindByID guard and no isolation test — verify rejection of explicit foreign IDs."},\n{"billingItemService.CreateItem", statusKnownUnguarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "all three FKs persisted directly without FindByID (billing_item_service.go:230)."}」。grep計測でstatusKnownUnguarded言及49行(定義2行を除き約47エントリ)。repository/CLAUDE.mdの規約は「正本ガード = 各サイトの runtime isolation test」だが、これらのエントリにはその正本ガードが存在しない。

**問題**

review網羅性gateは『名簿に載せる』ことしか担保せず(同ファイル冒頭に明記)、known-unguardedのまま滞留しているwrite経路はクロステナントmaster FK書き込みが実際に拒否されるかを誰も検証していない。会計ドメイン(accountingService.Update / billingItemService.CreateItem)を含む点が特に優先度が高い。テストを書くとRED(ガード自体が無い)になるため、これはガード実装を伴う挙動変更トラック。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

別トラック(挙動変更)として段階実施。優先順: (1)会計: accountingService.Update PaymentMethodID / billingItemService.CreateItem の3FK — service層にFindByID(clinicID,…)ガード追加後、internal/repository/または既存cross_tenant_master_fk_write_test.goパターンで『別クリニックFK指定→apperrors.WrapInvalidInput/NotFound拒否』テストを追加、allowlistエントリをguardedへ更新。(2)campaign TargetItemIDs(repo ReplaceTargets unscoped)。(3)carePlanItem HospitalizationPlanID / hospitalization CageID。(4)self-ref ParentID群(checkupType/consultation/examType)。各バッチはTestMasterFKWriteInventoryのstatus突合がgateになるため、allowlist更新漏れはCIで検出される。一括ではなくドメイン毎に1PRずつ、STGデータ監査(既存越境データの有無)を先行させること(R1-3の教訓)。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -count=1 && docker compose exec backend go test ./internal/repository/ -run TestCrossTenantMasterFKWrite -count=1
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
