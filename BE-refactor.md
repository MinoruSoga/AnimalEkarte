# BE-refactor.md — バックエンド リファクタリング計画（Appendix A / H フォローアップのみ残存）

> 残るのは Appendix A 8件（X-11〜X-18、P3中心）と、レビュー由来フォローアップ H-1〜H-7（別チケット化推奨）のみ。
> 本編（挙動保存トラック）および Appendix A の CLOSED 済み項目は本ファイルから削除済み（詳細は git log 参照）。
> X-9（resv-slot-phantom-toctou）・X-10（mr-version-check-not-atomic）は挙動変更トラックとして実装完了・CLOSED（詳細は git log 参照）。
> 前提: backend は 2026-07-02 完遂の D1-D13/R1-R3 計画で一度系統的にリファクタ済み・複数回の監査で well-maintained と判定済みのコードベースである。

## 監査の方法と信頼性

- 対象: `backend/` 配下全体。15次元で並列監査 → 敵対的検証。
- 本編（挙動保存）と Appendix A（挙動変更・別トラック）に分離。CLOSED 済みの完了履歴は本ファイルから削除済み。

## サマリー

| 区分 | 件数 | 内訳 |
|---|---|---|
| Appendix A（挙動変更・別トラック） | 8件 | X-11〜X-18 |
| レビュー由来フォローアップ（未登録・別チケット推奨） | 7件 | H-1, H-2, H-3, H-4, H-5, H-6, H-7 |

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

---

## Appendix A: 挙動変更を伴う項目（別トラック・PO/責任者判断を要する）

以下8件は監査で実在を確認した defect だが、修正すると HTTP レスポンス・DB書込結果・権限判定・API契約のいずれかが観測可能な形で変わる。このため本計画（挙動保存リファクタ）の実行対象からは外し、個別 Issue として起票のうえ別トラックで扱うことを推奨する。severity 順に記載。

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
