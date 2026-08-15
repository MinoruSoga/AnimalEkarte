package model_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/animal-ekarte/backend/internal/model"
)

// buildDSN は環境変数からDSNを構築する。
// Docker Compose 内: DB_HOST=db（デフォルト）
// CI (GitHub Actions): DB_HOST=localhost
func buildDSN() string {
	host := getEnvDefault("DB_HOST", "db")
	port := getEnvDefault("DB_PORT", "5432")
	user := getEnvDefault("DB_USER", "ekarte_user")
	pass := getEnvDefault("DB_PASSWORD", "ekarte_password")
	name := getEnvDefault("DB_NAME", "ekarte_db")
	ssl := getEnvDefault("DB_SSL_MODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, name, ssl)
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// allModels は全GORMモデルを列挙する。
// 新しいモデル追加時はここに追記すること。
func allModels() []any {
	return []any{
		&model.Owner{},
		&model.Pet{},
		&model.Clinic{},
		&model.Company{},
		&model.Account{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ShiftEntry{},
		&model.AnimalSpecies{},
		&model.ReservationType{},
		&model.Occupation{},
		&model.Insurance{},
		&model.ChiefComplaintType{},
		&model.Medicine{},
		&model.Procedure{},
		&model.Vaccine{},
		&model.CheckupType{},
		&model.CheckupPackageImportReceipt{},
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.ExamReferenceRange{},
		&model.DiagnosisType{},
		&model.DiagnosisName{},
		&model.MedicalRecord{},
		&model.ClinicalPlan{},
		&model.Consultation{},
		&model.VitalRecord{},
		&model.Treatment{},
		&model.Examination{},
		&model.ExamResult{},
		&model.ExaminationRevision{},
		&model.ExaminationRevisionItem{},
		&model.Vaccination{},
		&model.Checkup{},
		&model.Prescription{},
		&model.MedicalRecordImage{},
		&model.Reservation{},
		&model.Inquiry{},
		&model.InquiryTemplate{},
		&model.Hospitalization{},
		&model.HospitalizationPlan{},
		&model.Cage{},
		&model.CarePlanItem{},
		&model.TreatmentPlan{},
		&model.DailyRecord{},
		&model.CareLog{},
		&model.StaffNote{},
		&model.TrimmingCourse{},
		&model.TrimmingOption{},
		&model.AppointmentTrimmingDetail{},
		&model.AppointmentTrimmingOption{},
		&model.InventoryItem{},
		&model.Estimate{},
		&model.EstimateItem{},
		&model.BillingConfirmation{},
		&model.Billing{},
		&model.BillingItem{},
		&model.Payment{},
		&model.BillingRefund{},
		&model.MerchandiseItem{},
		// LINE予約システム
		&model.LineReservationSetting{},
		&model.LineCustomer{},
		&model.StaffReservationExclusion{},
		&model.StaffReservationCapability{},
		&model.ShiftEntryBreak{},
		&model.ShiftTemplate{},
		&model.ShiftTemplateBreak{},
		// 権限管理
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
		&model.StaffPermissionGroup{},
		// 予約種別グループ
		&model.ReservationTypeGroup{},
		// クリニック休診日
		&model.ClinicHoliday{},
		// FEAT-368: 集計・締め機能
		&model.ClinicSettings{},
		&model.ClosingSpecialPeriod{},
		&model.PaymentMethodMaster{},
		&model.CashRegisterClose{},
		&model.CashRegisterCloseAdjustment{},
		// G12-1: 監査・欠落モデル追記（TestAllModelsExhaustive で網羅性を担保）
		&model.AuditLog{},
		&model.PaymentSplit{},
		&model.CheckupFieldResult{},
		&model.CheckupTypeField{},
		&model.MedicalRecordAddendum{},
		&model.MedicineDoseParam{},
		&model.PetChronicCondition{},
		&model.PetOwner{},
		&model.PasswordResetToken{},
		&model.TokenBlacklist{},
		&model.LineLinkToken{},
		&model.LineSendLog{},
		&model.SharedFile{},
		&model.TrimmingCourseType{},
		&model.ManualArticle{},
		&model.ManualArticleVersion{},
		&model.Campaign{},
		&model.CampaignTargetCategory{},
		&model.CampaignTargetItem{},
		&model.ClinicIntegration{},
		&model.LabImportJob{},
		&model.LabImportEvent{},
		&model.LabImportUsageReceipt{},
		&model.LabImportExamRetraction{},
		&model.LabImportExamRetractionItem{},
		&model.LabImportRevertReceipt{},
		&model.ReservationTypeAvailableSlot{},
		&model.ReservationTypeOccupation{},
		&model.ReservationTypeUnavailableTime{},
		// Lステップ系
		&model.LstepAutoManagedPrefix{},
		&model.LstepConditionTagMapping{},
		&model.LstepCsvImport{},
		&model.LstepDeliveryTriggerLog{},
		&model.LstepFriendAttributeSnapshot{},
		&model.LstepSendPurposeTagPrefix{},
		&model.LstepSettings{},
		&model.LstepSyncErrorCounter{},
		&model.LstepTagCache{},
		&model.LstepTagCodeMapping{},
		&model.LstepTriggerPriority{},
		// #239 Phase 1 (fb11108c8): multi-clinic identity links
		&model.OwnerIdentityGroup{},
		&model.OwnerIdentityGroupMember{},
		&model.PetIdentityGroup{},
		&model.PetIdentityGroupMember{},
	}
}

// pgTypeCategory はPostgreSQLの型をカテゴリに分類する。
// 厳密な型名比較ではなく、カテゴリレベルで一致を検証する。
// （GORMの型推論とPostgreSQLの型名表記は完全一致しないため）
func pgTypeCategory(dbType string) string {
	t := strings.ToLower(dbType)

	switch {
	case t == "bigint" || t == "integer" || t == "smallint" || t == "int4" || t == "int8" || t == "int2":
		return "integer"
	case strings.HasPrefix(t, "numeric") || strings.HasPrefix(t, "decimal") || t == "double precision" || t == "real" || t == "float8" || t == "float4":
		return "numeric"
	case t == "boolean" || t == "bool":
		return "boolean"
	case t == "text" || strings.HasPrefix(t, "character") || strings.HasPrefix(t, "varchar"):
		return "text"
	case strings.Contains(t, "timestamp"):
		return "timestamp"
	case t == "date":
		return "date"
	case t == "bytea":
		return "bytea"
	case strings.HasSuffix(t, "[]") || t == "array":
		return "array"
	case t == "inet":
		// G12-1 (X-3依存): PG組込型を明示カテゴリ化し、isEnumLikeの許容集合から除外する。
		// これにより audit_logs.ip_address (DB=inet, Go=string) の型不一致が検出可能になる。
		return "inet"
	case t == "uuid":
		return "uuid"
	default:
		// ENUM型やその他はそのまま返す
		return t
	}
}

// goTypeCategory はGoの型をDBカテゴリにマッピングする。
func goTypeCategory(goType reflect.Type, gormTag string) string {
	// gorm タグで明示的に型指定がある場合はそれを使う
	if tag := extractGormType(gormTag); tag != "" {
		return pgTypeCategory(tag)
	}

	// ポインタ型はデリファレンス
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	// pq.StringArray / pq.Int64Array 等の配列型
	if goType.Name() == "StringArray" || goType.Name() == "Int64Array" {
		return "array"
	}

	switch goType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "numeric"
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "text"
	default:
		name := goType.Name()
		if name == "Time" || name == "DeletedAt" {
			return "timestamp"
		}
		// カスタム型（enum等）は文字列ベース
		if goType.Kind() == reflect.String {
			return "text"
		}
		return strings.ToLower(name)
	}
}

// extractGormType は gorm タグから "type:xxx" を抽出する。
func extractGormType(tag string) string {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "type:") {
			return strings.TrimPrefix(part, "type:")
		}
	}
	return ""
}

// knownSchemaDriftAllowlist は既知だが未解消のスキーマ差分を一時的に許容するための
// allowlist。エントリを追加する際は、根拠となる Issue 番号を必ず併記すること。
// 実修正が完了したら該当エントリを削除し、検査を再度有効化すること。
//
// key は "ModelType.column_name"。型不一致（差分チェック1）と、DB にだけ存在する
// カラム（差分チェック2: Go モデルにフィールド無し）の双方で参照する。
var knownSchemaDriftAllowlist = map[string]string{
	// X-3 (audit-ip-inet-model-drift): audit_logs.ip_address は DB=inet, Go=*string(text) で
	// 型カテゴリ不一致（pgTypeCategory は "inet" と "text" を区別する）。
	// X-3 では IPAddress を string→*string に変更し、空文字列が ''::inet として INSERT され
	// 22P02 になる本来の欠陥（NULL許容性の欠如）を解消した。型カテゴリそのものの一致
	// （Go側をnetip.Addr等に変更する、またはDB側をtextに変更する）は本ユニットのスコープ外
	// として意図的に見送り、allowlist を維持する（*string でも PostgreSQL 側で text→inet の
	// 暗黙キャストが機能するため実害はない。実測は audit_real_ddl_test.go 参照）。
	"AuditLog.ip_address": "X-3",

	// cash_register_closes.deleted_at は 001_init.sql 由来のレガシー列。append-only 契約
	// （003_cash_register_close_append_only.sql + TestCashRegisterCloseRepository_AppendOnlyContract_NoDeleteMethod）
	// により app は Update/Delete/soft-delete しない。Go モデルに gorm.DeletedAt を足すと
	// GORM が全 query に deleted_at IS NULL を自動付与し append-only の前提が崩れるため、
	// 意図的にフィールドを持たない（cash_register_close.go コメントと一致）。DDL 側の列削除は
	// 適用済み migration 編集禁止のため行わない。
	"CashRegisterClose.deleted_at": "append-only: intentional unmapped legacy deleted_at (no gorm.DeletedAt)",
}

// knownNullabilityDriftAllowlist は Go=pointer(NULL許容) だが DB=NOT NULL(デフォルト無し) の
// 既知の不整合を一時的に許容するための allowlist。
// エントリを追加する際は、根拠となるコメント（原因・実修正の要否）を必ず併記すること。
//
// 以下2件は G12-1（TestAllModelsExhaustive / nullability チェック新設）により新たに
// 検出された、本ユニット追加前から存在する Go/DB 間のNULL許容性不整合である。
// このユニットは test-infra 追加（挙動保存）のみが範囲であり、モデル/マイグレーションの
// 実修正はスコープ外のため、検出ロジック追加を GREEN で完了させるために一時 pin する。
// 実修正（Go側を非pointerにする、または業務要件次第でDB側をNULL許容にする）は別issue化すること。
var knownNullabilityDriftAllowlist = map[string]string{
	// AuditLog.ClinicID は Go=*uint64 だが audit_logs.clinic_id は NOT NULL REFERENCES clinics(id)。
	// X-3 (audit-ip-inet-model-drift) で意図的に維持: service.validateAuditLog（audit_service.go）が
	// 永続化前に非nil/非ゼロを検証する経路を持ち、呼び出し元が検証前の構造体を一時的に組み立てる
	// 余地を残すため、DB の NOT NULL 制約を最終防衛線として Go 側は pointer のまま
	// gorm:"not null" タグのみ付与した（このチェックは Go 型の pointer 性のみを見るため、タグの
	// 有無に関わらず引き続き allowlist が必要）。DB レベルの拒否は
	// TestAuditLogRealDDL_NilClinicIDFails（internal/repository/audit_real_ddl_test.go）で保証。
	"AuditLog.clinic_id": "X-3: DB制約(NOT NULL)を最終防衛線とし、Go側は*uint64+gorm:not nullを維持（意図的）",

	// LstepCsvImport.UploadedByUserID pin は H-5 で解消（BE-refactor.md 第4期）。
	// 供給元（lstep_csv_import_handler.go の ImportLstepFriendAttributesCsv）を遡って確認した結果、
	// actorID は routegroup 全体に適用済みの middleware.Auth（Auth必須・常にJWTのuser_idを設定）+
	// extractStaffID（未設定/パース失敗時は即401 RespondError）を経由するため、nilで
	// ImportFriendAttributesCSV に到達する正当な経路は無い（構造的に非nil保証）。
	// Go側を UploadedByUserID uint64 + gorm:"not null" に変更しDBのNOT NULL制約と一致させたため、
	// このエントリは stale になり削除（残すと「許容不要な差分」が allowlist に残り続け、将来の
	// 別カラムでの誤用の温床になる）。
}

// fieldNullMeta はGoフィールドのNULL許容性判定に必要なメタ情報を保持する。
type fieldNullMeta struct {
	isPointer       bool // Goフィールドがポインタ型（nil許容）か
	isPrimaryKey    bool // gorm:"primaryKey" タグの有無
	isAutoIncr      bool // gorm:"autoIncrement" タグの有無
	isTimeOrSoft    bool // time.Time / gorm.DeletedAt か（ゼロ値・内部NULL処理を持つため対象外）
	explicitNotNull bool // gorm:"not null" タグの有無
}

// TestSchemaDrift はGoモデルとDBスキーマの差分を検出する。
//
// 検出する差分:
//   - DBにカラムが存在するがGoモデルにフィールドがない
//   - GoモデルにフィールドがあるがDBにカラムが存在しない
//   - カラムの型カテゴリ（integer/numeric/text/boolean/timestamp）が不一致
//   - Go=pointer(NULL許容) だが DB=NOT NULL(デフォルト無し) というNULL許容性の不整合
//
// 実行条件:
//   - Docker Compose でDBが起動していること（docker compose up db）
//   - マイグレーションが適用済みであること
//
// 実行方法:
//
//	docker compose exec backend go test ./internal/model/ -run TestSchemaDrift -v
func TestSchemaDrift(t *testing.T) {
	dsn := buildDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		t.Skipf("DB connection failed (skip in CI without DB): %v", err)
	}

	models := allModels()
	var drifts []string
	var warnings []string

	for _, m := range models {
		modelType := reflect.TypeOf(m)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}

		// テーブル名を取得
		tableName := ""
		if tn, ok := m.(interface{ TableName() string }); ok {
			tableName = tn.TableName()
		} else {
			t.Errorf("model %s does not implement TableName()", modelType.Name())
			continue
		}

		// DBからカラム情報を取得
		migrator := db.Migrator()
		if !migrator.HasTable(tableName) {
			drifts = append(drifts, fmt.Sprintf("[%s] テーブル %q がDBに存在しない", modelType.Name(), tableName))
			continue
		}

		dbColumns, err := migrator.ColumnTypes(m)
		if err != nil {
			t.Errorf("[%s] ColumnTypes failed: %v", modelType.Name(), err)
			continue
		}

		// DBカラムをmap化
		dbColMap := make(map[string]string) // column_name -> db_type
		dbColumnTypeMap := make(map[string]gorm.ColumnType)
		for _, col := range dbColumns {
			colType := col.DatabaseTypeName()
			dbColMap[col.Name()] = colType
			dbColumnTypeMap[col.Name()] = col
		}

		// Goモデルのフィールドを走査
		modelColMap := make(map[string]string) // column_name -> go_type_category
		modelFieldMeta := make(map[string]fieldNullMeta)
		for i := range modelType.NumField() {
			field := modelType.Field(i)

			// リレーションフィールド（foreignKey, many2many 等）はスキップ
			gormTag := field.Tag.Get("gorm")
			if strings.Contains(gormTag, "foreignKey") || strings.Contains(gormTag, "many2many") {
				continue
			}
			// 仮想フィールド（gorm:"-"）はスキップ
			if gormTag == "-" || strings.HasPrefix(gormTag, "-;") {
				continue
			}

			// 構造体型のリレーションフィールドをスキップ（Time, DeletedAt は除外）
			// スライス型はリレーションの場合が多いが、pq.StringArray や []byte(JSONB) は除外しない
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			isStringArray := fieldType.Name() == "StringArray" || fieldType.Name() == "Int64Array"
			isByteSlice := fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8
			if fieldType.Kind() == reflect.Slice && !isStringArray && !isByteSlice {
				continue
			}
			if fieldType.Kind() == reflect.Struct && fieldType.Name() != "Time" && fieldType.Name() != "DeletedAt" {
				continue
			}

			// json タグからカラム名を推論
			jsonTag := field.Tag.Get("json")
			var colName string
			if jsonTag == "" || jsonTag == "-" {
				// json:"-" のフィールドもDBカラムとしては存在する場合がある
				// （例: PasswordHash, DeletedAt）
				// Go フィールド名を snake_case に変換してカラム名とする
				colName = toSnakeCase(field.Name)
			} else {
				colName = strings.Split(jsonTag, ",")[0]
				if colName == "" {
					continue
				}
			}

			category := goTypeCategory(field.Type, gormTag)
			modelColMap[colName] = category
			modelFieldMeta[colName] = fieldNullMeta{
				isPointer:       field.Type.Kind() == reflect.Ptr,
				isPrimaryKey:    strings.Contains(gormTag, "primaryKey"),
				isAutoIncr:      strings.Contains(gormTag, "autoIncrement"),
				isTimeOrSoft:    fieldType.Name() == "Time" || fieldType.Name() == "DeletedAt",
				explicitNotNull: strings.Contains(gormTag, "not null") || strings.Contains(gormTag, "notNull"),
			}
		}

		// 差分チェック 1: Goモデルにあるがに存在しないカラム
		for colName, goCategory := range modelColMap {
			dbType, exists := dbColMap[colName]
			if !exists {
				drifts = append(drifts, fmt.Sprintf("[%s.%s] Goモデルにフィールドがあるが、テーブル %q にカラムが存在しない", modelType.Name(), colName, tableName))
				continue
			}

			// 型カテゴリ比較
			dbCategory := pgTypeCategory(dbType)
			if goCategory != dbCategory {
				// ENUM型はtext/enum両方あり得るので、Goが string ベースなら許容
				if isEnumLike(goCategory) || isEnumLike(dbCategory) {
					continue
				}
				key := fmt.Sprintf("%s.%s", modelType.Name(), colName)
				if issue, ok := knownSchemaDriftAllowlist[key]; ok {
					t.Logf("[allowlist:%s] %s 型不一致を既知issueとして許容: Go=%s, DB=%s (raw: %s)", issue, key, goCategory, dbCategory, dbType)
					continue
				}
				drifts = append(drifts, fmt.Sprintf("[%s.%s] 型不一致: Go=%s, DB=%s (raw: %s)", modelType.Name(), colName, goCategory, dbCategory, dbType))
			}
		}

		// 差分チェック 2: DBにあるがGoモデルにないカラム
		for colName := range dbColMap {
			if _, exists := modelColMap[colName]; !exists {
				key := fmt.Sprintf("%s.%s", modelType.Name(), colName)
				if issue, ok := knownSchemaDriftAllowlist[key]; ok {
					t.Logf("[allowlist:%s] %s DBのみのカラムを既知として許容（Goモデルにフィールド無し）", issue, key)
					continue
				}
				drifts = append(drifts, fmt.Sprintf("[%s] テーブル %q のカラム %q がGoモデルにフィールドとして定義されていない", modelType.Name(), tableName, colName))
			}
		}

		// 差分チェック 3: NULL許容性の不整合
		// 危険な方向のみ fail: Go=pointer(NULL許容) だが DB=NOT NULL(デフォルト無し)。
		// 逆方向（DB=nullable だが Go=非pointer）は warnings へ（fail させない）。
		for colName, meta := range modelFieldMeta {
			if meta.isPrimaryKey || meta.isAutoIncr || meta.isTimeOrSoft {
				continue
			}
			col, ok := dbColumnTypeMap[colName]
			if !ok {
				continue // 差分チェック1で報告済み
			}
			dbNullable, nullableOK := col.Nullable()
			if !nullableOK {
				continue
			}
			_, hasDefault := col.DefaultValue()

			key := fmt.Sprintf("%s.%s", modelType.Name(), colName)

			if meta.isPointer && !dbNullable && !hasDefault {
				if reason, allowed := knownNullabilityDriftAllowlist[key]; allowed {
					t.Logf("[allowlist:nullability:%s] %s NULL許容性不整合を許容: Go=pointer, DB=NOT NULL(デフォルト無し)", reason, key)
					continue
				}
				drifts = append(drifts, fmt.Sprintf("[%s] NULL許容性不整合: Go=pointer(NULL許容) だが DB=NOT NULL(デフォルト無し)", key))
				continue
			}

			if !meta.isPointer && dbNullable && !meta.explicitNotNull {
				warnings = append(warnings, fmt.Sprintf("[%s] NULL許容性注意: DB=nullable だが Go=非pointer", key))
			}
		}
	}

	if len(warnings) > 0 {
		t.Logf("NULL許容性の注意事項 (%d件、fail対象外):\n%s", len(warnings), strings.Join(warnings, "\n"))
	}

	if len(drifts) > 0 {
		t.Errorf("スキーマ差分を検出 (%d件):\n%s", len(drifts), strings.Join(drifts, "\n"))
	}
}

// tableNameReceiverTypes は internal/model 配下（_test.go を除く）を go/ast で走査し、
// `func (T) TableName() string` を実装する全型名を収集する。
// allModels() の手動保守漏れを機械的に検出するための基礎データとなる。
func tableNameReceiverTypes(t *testing.T) map[string]bool {
	t.Helper()

	result := make(map[string]bool)
	fset := token.NewFileSet()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read internal/model directory: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", name, err)
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name.Name != "TableName" {
				continue
			}
			if len(fn.Recv.List) != 1 {
				continue
			}
			typeName := receiverTypeName(fn.Recv.List[0].Type)
			if typeName != "" {
				result[typeName] = true
			}
		}
	}

	return result
}

// receiverTypeName はレシーバの型式（値/ポインタ）から型名を抽出する。
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	default:
		return ""
	}
}

// TestAllModelsExhaustive は internal/model 配下で TableName() を実装する全モデルが
// allModels() に登録されているかを go/ast ベースで機械検証する（G12-1）。
// 手動保守（コメントによる注意喚起のみ）では新規モデル追加時の登録漏れを防げないため、
// ソースコード走査による双方向突合で強制する。
//
// 実行方法:
//
//	docker compose exec backend go test ./internal/model/ -run TestAllModelsExhaustive -v
func TestAllModelsExhaustive(t *testing.T) {
	astModels := tableNameReceiverTypes(t)

	registered := make(map[string]bool)
	for _, m := range allModels() {
		typ := reflect.TypeOf(m)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		registered[typ.Name()] = true
	}

	var missing []string
	for name := range astModels {
		if !registered[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)

	var extra []string
	for name := range registered {
		if !astModels[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)

	if len(missing) > 0 {
		t.Errorf("allModels() に未登録のモデル (%d件、TableName()実装あり): %s\n"+
			"新規モデル追加時は allModels() への追記が必須です。", len(missing), strings.Join(missing, ", "))
	}
	if len(extra) > 0 {
		t.Errorf("allModels() に登録されているが TableName() 実装が見つからないモデル (%d件): %s", len(extra), strings.Join(extra, ", "))
	}
}

// toSnakeCase はPascalCaseをsnake_caseに変換する。
// 連続する大文字（ID, URL等）は一つの単語として扱う。
// 例: MedicalRecordID → medical_record_id（medical_record_i_d ではない）
func toSnakeCase(s string) string {
	runes := []rune(s)
	var result strings.Builder
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			// 先行文字が小文字の場合のみアンダースコアを挿入
			if i > 0 && runes[i-1] >= 'a' && runes[i-1] <= 'z' {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32) // toLower
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// isEnumLike はカスタムENUM型かどうかを判定する。
// PostgreSQLのENUM型名はアプリケーション定義なので、
// integer/numeric/text/boolean/timestamp/date/bytea 以外はENUMと見なす。
func isEnumLike(category string) bool {
	switch category {
	case "integer", "numeric", "boolean", "text", "timestamp", "date", "bytea", "inet", "uuid":
		return false
	default:
		return true
	}
}
