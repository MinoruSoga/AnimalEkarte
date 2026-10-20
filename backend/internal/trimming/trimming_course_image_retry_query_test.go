package trimming

import (
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SQL generation contract only; no connection, fixtures or real DB acceptance.
func TestCourseImageRetry_ProofFiltersPrecedeBatchLimit(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return pendingCourseImageDeletionQuery(tx, 7, 100).Find(&[]struct{ ObjectKey string }{})
	})
	for _, required := range []string{"JOIN trimming_course_image_uploads", "u.uploader_staff_id = d.actor_staff_id", "d.clinic_id = 7", "d.storage_deleted_at IS NULL", "u.uploaded_at IS NOT NULL", "u.submitted_at IS NULL", "ORDER BY d.last_attempt_at ASC NULLS FIRST, d.object_key ASC", "LIMIT 100"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("unsafe/unfair retry SQL missing %q: %s", required, sql)
		}
	}
	if strings.Index(sql, "u.submitted_at IS NULL") > strings.Index(sql, "LIMIT 100") {
		t.Fatal("proof filtering applied after batch selection")
	}
}
