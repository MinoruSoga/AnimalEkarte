# BE Todo — バックエンド残タスク台帳

- **更新日**: 2026-07-12（対応済 PERF/SEED を削除。詳細・手順は git 履歴と `docs/tasks/closed/` が正本）
- **本書の規約**: 今期着手可能な PERF/SEED 系の未対応のみを記載する。対応済みは残さない。
- **別台帳**:
  - リファクタ系の今期着手可能残: `BE-refactor.md`
  - 次期送り・着手保留・任意検証: `BE-pending.md`（例: X-16②、STG クロステナント監査）
  - 本書と重複させない。

### 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 残タスク一覧

**エージェント実装可能な残タスクなし（2026-07-12 時点）。**

### 人間作業（未適用）

| ID | 内容 | 状態 |
|----|------|------|
| PERF-FOLLOWUP-01（適用残） | migration `003_add_pets_batch_living_count_index.sql` | **未適用**。checksum 問題なし。次回デプロイ時の `cmd/migrate` で自動適用（`db_reset` 不要）。ローカルは migrate 再実行で反映。 |

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。正本: `docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件のため本台帳に含めない。

---

### 既知バグの skip テスト台帳(2026-07-13 棚卸し)

`rg -n 't\.Skip\("known production bug' backend/internal --glob '*.go'` 実測 9 件（コード変更なし・台帳化のみ）。

| ファイル:行 | テスト名 | skip メッセージ要約 |
|---|---|---|
| `hospitalization_repository_test.go:221` | `TestHospitalizationRepository_FindByID_Success` | `FindByID` の `Preload("CarePlanItems"/"DailyRecords", "deleted_at IS NULL")` が実在しない `deleted_at` 列を参照し、常に `42703` で失敗 |
| `hospitalization_repository_test.go:296` | `TestHospitalizationRepository_Update_Success` | `Update` が内部で `FindByID` を呼ぶため、上記と同根の `deleted_at` 列不存在バグの影響を受ける |
| `hospitalization_repository_test.go:358` | `TestHospitalizationRepository_CountCarePlanItemsByHospitalizationID` | `care_plan_items.deleted_at` 列が `CREATE TABLE`/`model.CarePlanItem` のいずれにも存在せず、常に `42703` |
| `hospitalization_repository_test.go:387` | `TestHospitalizationRepository_CountDailyRecordsByHospitalizationID` | `daily_records.deleted_at` 列が同様に存在せず、常に `42703` |
| `pet_chronic_condition_repository_test.go:209` | `TestPetChronicConditionRepository_FindActiveConditionCodesByOwner` | `model.PetChronicCondition.IsActive`（`gorm:"not null;default:true"`）が zero 値 `false` を GORM `Create()` 時に自動省略し、常に DB デフォルト `true` で作成される（inactive を永続化不可能） |
| `medicine_repository_test.go:301` | `TestMedicineRepository_CountUsageByMedicineID_TreatmentUsage`（t.Run "同一クリニックの treatment 使用が1件カウントされる"） | `CountUsageByMedicineID` の第2クエリが存在しない `care_plan_items.deleted_at` を参照し常に `42703`。Medicine 削除時の FK 依存チェックが機能不全 |
| `medicine_repository_test.go:308` | `TestMedicineRepository_CountUsageByMedicineID_TreatmentUsage`（t.Run "別クリニックからは0件（clinic_id 隔離・JOIN スコープ）"） | 同上と同根の `care_plan_items.deleted_at` 列不存在バグ |
| `hospitalization_plan_repository_test.go:229` | `TestHospitalizationPlanRepository_CountUsageByHospitalizationPlanID` | 同根の `care_plan_items.deleted_at` 列不存在バグ（`hospitalization_repository.go` と同一原因） |
| `payment_method_master_repository_test.go:258` | `TestPaymentMethodMasterRepository_Reorder`（t.Run "指定した順序でdisplay_orderが1始まりに更新される"） | 共通ヘルパー `reorderByClinicID` がハードコードする `"sort_order"` 列が `payment_methods` テーブルに存在せず（実カラム名は `display_order`）、常に `42703` |
