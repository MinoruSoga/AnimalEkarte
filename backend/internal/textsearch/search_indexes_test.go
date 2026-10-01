package textsearch_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

const searchFoldMigrationPath = "../../migrations/015_search_fold_indexes.sql"

// foldedColumnPairs は migrations/015_search_fold_indexes.sql が作る
// translate() 式インデックスと、検索クエリが FoldedExpr 経由で生成する
// 式の対応表。ここに載っていない列にインデックスを追加してはいけないし、
// ここに載っている列の式が migration とズレてもいけない。
var foldedColumnPairs = []struct {
	table  string
	column string
	index  string
}{
	{"owners", "name", "idx_owners_name_fold_trgm"},
	{"owners", "name_kana", "idx_owners_name_kana_fold_trgm"},
	{"pets", "name", "idx_pets_name_fold_trgm"},
	{"pets", "name_kana", "idx_pets_name_kana_fold_trgm"},
	{"inquiries", "chief_complaint", "idx_inquiries_chief_complaint_fold_trgm"},
	{"treatments", "content", "idx_treatments_content_fold_trgm"},
	{"treatments", "memo", "idx_treatments_memo_fold_trgm"},
	{"procedures", "name", "idx_procedures_name_fold_trgm"},
	{"medicines", "name", "idx_medicines_name_fold_trgm"},
	{"consultations", "name", "idx_consultations_name_fold_trgm"},
	{"inventory_items", "name", "idx_inventory_items_name_fold_trgm"},
	{"vaccines", "name", "idx_vaccines_name_fold_trgm"},
}

// TestSearchFoldMigrationLiteralParity は migration 内の translate() リテラルが
// textsearch.FoldedExpr の生成式と文字単位で一致することを保証する。
// PostgreSQL の式インデックスはリテラル一致でバインドするため、
// カナ/空白文字セットを片側だけ変更すると静かに Seq Scan へ戻る。
// このテストはそのズレをコンパイル時に近い位置で検出するためのガード。
func TestSearchFoldMigrationLiteralParity(t *testing.T) {
	raw, err := os.ReadFile(searchFoldMigrationPath)
	require.NoError(t, err)
	migration := string(raw)

	for _, tc := range foldedColumnPairs {
		expr := textsearch.FoldedExpr(tc.column)
		require.Contains(t, migration, expr,
			"migration に %s の式が見つからない: %s", tc.index, expr)
		require.Contains(t, migration, tc.index,
			"migration にインデックス名 %s が見つからない", tc.index)
	}
	require.Contains(t, migration, textsearch.SpaceStripRegexp,
		"compact 式インデックスの空白除去 regexp が migration に見つからない")
	require.Contains(t, migration,
		"regexp_replace("+textsearch.FoldedExpr("name"),
		"idx_owners_name_compact_fold_trgm の式が FoldedExpr と不一致")
}

// TestSearchFoldIndexesUsedByPlanner は migration を実 DB に適用した上で、
// 検索クエリと同形の述語が各 GIN/btree 式インデックスを使う計画になることを
// enable_seqscan=off + EXPLAIN で検証する。
func TestSearchFoldIndexesUsedByPlanner(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{}, &model.Pet{}, &model.MedicalRecord{}, &model.Inquiry{},
		&model.Treatment{}, &model.Procedure{}, &model.Medicine{},
		&model.Consultation{}, &model.InventoryItem{}, &model.Vaccine{},
	))

	require.NoError(t, db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error)
	applySearchFoldMigration(t, db)

	// 本番(001_init.sql)に存在するが AutoMigrate テストスキーマに無い btree
	// インデックスを補う。ペア IN / FK 相関プローブの計画を本番同等にするための
	// 前提条件（PK 以外の FK 参照列は AutoMigrate では索引が作られない）。
	for _, stmt := range []string{
		"CREATE INDEX IF NOT EXISTS idx_medical_records_clinic_record_no ON medical_records(clinic_id, record_no)",
		"CREATE INDEX IF NOT EXISTS idx_medical_records_clinic_pet ON medical_records(clinic_id, pet_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_medical_records_clinic_owner ON medical_records(clinic_id, owner_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_treatments_medical_record_id ON treatments(medical_record_id)",
		"CREATE INDEX IF NOT EXISTS idx_inquiries_medical_record_id ON inquiries(medical_record_id)",
	} {
		require.NoError(t, db.Exec(stmt).Error)
	}

	probe := "%プローブ%"
	type plainProbe struct {
		table string
		where string
		index string
	}
	probes := []plainProbe{
		{"owners", "name ILIKE '" + probe + "'", "idx_owners_name_trgm"},
		{"owners", "phone ILIKE '" + probe + "'", "idx_owners_phone_trgm"},
		{"owners", "email ILIKE '" + probe + "'", "idx_owners_email_trgm"},
		{"owners", "id::text = '12345'", "idx_owners_id_text"},
		{"pets", "pet_number ILIKE '" + probe + "'", "idx_pets_pet_number_trgm"},
		{"medical_records", "record_no ILIKE '" + probe + "'", "idx_medical_records_record_no_trgm"},
	}
	for _, tc := range foldedColumnPairs {
		probes = append(probes, plainProbe{
			table: tc.table,
			where: textsearch.FoldedExpr(tc.column) + " ILIKE '" + probe + "'",
			index: tc.index,
		})
	}
	probes = append(probes, plainProbe{
		table: "owners",
		where: "regexp_replace(" + textsearch.FoldedExpr("name") + ", '" +
			textsearch.SpaceStripRegexp + "', '', 'g') ILIKE '" + probe + "'",
		index: "idx_owners_name_compact_fold_trgm",
	})

	// SET enable_seqscan はセッション単位なので、同一コネクション上で
	// SET → EXPLAIN を実行するため gorm.Connection でピン留めする。
	err := db.Connection(func(tx *gorm.DB) error {
		if err := tx.Exec("SET enable_seqscan = off").Error; err != nil {
			return err
		}
		for _, p := range probes {
			plan := explainPlan(t, tx, "SELECT 1 FROM "+p.table+" WHERE "+p.where)
			require.Contains(t, plan, p.index,
				"%s.%s の述語がインデックス %s を使わない plan:\n%s",
				p.table, p.where, p.index, plan)
		}

		// 複合形状: applyMedicalRecordSearch が生成するペア IN + UNION ALL 形の
		// 全体を再現し、どのノードにも Seq Scan が残らないことを確認する。
		// 各腕は trgm/btree インデックス駆動の ID 集合計算になり、
		// JOIN+OR の全行逐行評価に戻らない。
		// （腕の個別インデックス一致は上の probes が担保する）
		lit := "'%プローブ%'"
		composite := "SELECT 1 FROM medical_records" +
			" WHERE medical_records.clinic_id IN (1)" +
			" AND medical_records.deleted_at IS NULL" +
			" AND (medical_records.id, medical_records.clinic_id) IN (" +
			"SELECT mr_no.id, mr_no.clinic_id FROM medical_records mr_no" +
			" WHERE mr_no.deleted_at IS NULL" +
			" AND mr_no.record_no ILIKE " + lit + " ESCAPE '\\'" +
			" UNION ALL " +
			"SELECT mr_o.id, mr_o.clinic_id FROM medical_records mr_o" +
			" JOIN owners searched_owner" +
			" ON searched_owner.id = mr_o.owner_id" +
			" AND searched_owner.clinic_id = mr_o.clinic_id" +
			" AND searched_owner.deleted_at IS NULL" +
			" WHERE mr_o.deleted_at IS NULL AND (" +
			textsearch.FoldedExpr("searched_owner.name") + " ILIKE " + lit + " ESCAPE '\\'" +
			" OR " + textsearch.FoldedExpr("searched_owner.name_kana") + " ILIKE " + lit + " ESCAPE '\\')" +
			" UNION ALL " +
			"SELECT mr_p.id, mr_p.clinic_id FROM medical_records mr_p" +
			" JOIN pets searched_pet" +
			" ON searched_pet.id = mr_p.pet_id" +
			" AND searched_pet.clinic_id = mr_p.clinic_id" +
			" AND searched_pet.deleted_at IS NULL" +
			" WHERE mr_p.deleted_at IS NULL AND (" +
			textsearch.FoldedExpr("searched_pet.name") + " ILIKE " + lit + " ESCAPE '\\'" +
			" OR " + textsearch.FoldedExpr("searched_pet.name_kana") + " ILIKE " + lit + " ESCAPE '\\')" +
			" UNION ALL " +
			"SELECT searched_inquiry.medical_record_id, mr_i.clinic_id" +
			" FROM inquiries searched_inquiry" +
			" JOIN medical_records mr_i ON mr_i.id = searched_inquiry.medical_record_id" +
			" WHERE " + textsearch.FoldedExpr("searched_inquiry.chief_complaint") + " ILIKE " + lit + " ESCAPE '\\'" +
			" UNION ALL " +
			"SELECT searched_treatment.medical_record_id, mr_t.clinic_id" +
			" FROM treatments searched_treatment" +
			" JOIN medical_records mr_t ON mr_t.id = searched_treatment.medical_record_id" +
			" WHERE searched_treatment.deleted_at IS NULL AND (" +
			textsearch.FoldedExpr("searched_treatment.content") + " ILIKE " + lit + " ESCAPE '\\'" +
			" OR " + textsearch.FoldedExpr("searched_treatment.memo") + " ILIKE " + lit + " ESCAPE '\\'" +
			" OR (searched_treatment.procedure_id, mr_t.clinic_id) IN (" +
			"SELECT searched_procedure.id, searched_procedure.clinic_id FROM procedures searched_procedure" +
			" WHERE searched_procedure.deleted_at IS NULL" +
			" AND " + textsearch.FoldedExpr("searched_procedure.name") + " ILIKE " + lit + " ESCAPE '\\')" +
			" OR (searched_treatment.medicine_id, mr_t.clinic_id) IN (" +
			"SELECT searched_medicine.id, searched_medicine.clinic_id FROM medicines searched_medicine" +
			" WHERE searched_medicine.deleted_at IS NULL" +
			" AND " + textsearch.FoldedExpr("searched_medicine.name") + " ILIKE " + lit + " ESCAPE '\\')" +
			" OR (searched_treatment.consultation_id, mr_t.clinic_id) IN (" +
			"SELECT searched_consultation.id, searched_consultation.clinic_id FROM consultations searched_consultation" +
			" WHERE searched_consultation.deleted_at IS NULL" +
			" AND " + textsearch.FoldedExpr("searched_consultation.name") + " ILIKE " + lit + " ESCAPE '\\')" +
			" OR (searched_treatment.inventory_id, mr_t.clinic_id) IN (" +
			"SELECT searched_inventory.id, searched_inventory.clinic_id FROM inventory_items searched_inventory" +
			" WHERE searched_inventory.deleted_at IS NULL" +
			" AND " + textsearch.FoldedExpr("searched_inventory.name") + " ILIKE " + lit + " ESCAPE '\\')))"
		plan := explainPlan(t, tx, composite)
		require.NotContains(t, plan, "Seq Scan",
			"カルテ横断検索の複合計画に Seq Scan が残っている:\n%s", plan)
		return nil
	})
	require.NoError(t, err)
}

// applySearchFoldMigration は migration SQL を文単位で実行する。
// 共有テスト DB 上で再実行された場合の "already exists" は冪等として許容する。
func applySearchFoldMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	raw, err := os.ReadFile(searchFoldMigrationPath)
	require.NoError(t, err)
	for _, stmt := range strings.Split(string(raw), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		err := db.Exec(stmt).Error
		if err != nil && strings.Contains(err.Error(), "already exists") {
			continue
		}
		require.NoError(t, err)
	}
}

// explainPlan は EXPLAIN の出力を1つの文字列に連結して返す。
func explainPlan(t *testing.T, tx *gorm.DB, stmt string) string {
	t.Helper()
	rows, err := tx.Raw("EXPLAIN " + stmt).Rows()
	require.NoError(t, err)
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		require.NoError(t, rows.Scan(&line))
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	require.NoError(t, rows.Err())
	return plan.String()
}
