package model_test

import (
	"fmt"
	"os"
	"reflect"
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
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.DiagnosisType{},
		&model.DiagnosisName{},
		&model.MedicalRecord{},
		&model.ClinicalPlan{},
		&model.Consultation{},
		&model.VitalRecord{},
		&model.Treatment{},
		&model.Examination{},
		&model.ExamResult{},
		&model.Vaccination{},
		&model.Checkup{},
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

	// pq.StringArray 等の配列型
	if goType.Name() == "StringArray" {
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

// TestSchemaDrift はGoモデルとDBスキーマの差分を検出する。
//
// 検出する差分:
//   - DBにカラムが存在するがGoモデルにフィールドがない
//   - GoモデルにフィールドがあるがDBにカラムが存在しない
//   - カラムの型カテゴリ（integer/numeric/text/boolean/timestamp）が不一致
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
		for _, col := range dbColumns {
			colType := col.DatabaseTypeName()
			dbColMap[col.Name()] = colType
		}

		// Goモデルのフィールドを走査
		modelColMap := make(map[string]string) // column_name -> go_type_category
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
			isStringArray := fieldType.Name() == "StringArray"
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
				drifts = append(drifts, fmt.Sprintf("[%s.%s] 型不一致: Go=%s, DB=%s (raw: %s)", modelType.Name(), colName, goCategory, dbCategory, dbType))
			}
		}

		// 差分チェック 2: DBにあるがGoモデルにないカラム
		for colName := range dbColMap {
			if _, exists := modelColMap[colName]; !exists {
				drifts = append(drifts, fmt.Sprintf("[%s] テーブル %q のカラム %q がGoモデルにフィールドとして定義されていない", modelType.Name(), tableName, colName))
			}
		}
	}

	if len(drifts) > 0 {
		t.Errorf("スキーマ差分を検出 (%d件):\n%s", len(drifts), strings.Join(drifts, "\n"))
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
	case "integer", "numeric", "boolean", "text", "timestamp", "date", "bytea":
		return false
	default:
		return true
	}
}
