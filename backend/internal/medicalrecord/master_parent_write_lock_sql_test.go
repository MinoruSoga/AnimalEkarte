package medicalrecord

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// SQL generation only; no database connection or concurrency acceptance claim.
func TestMasterParentWrite_LockSQLOrderAndScope(t *testing.T) {
	for _, entity := range []string{"consultation", "procedure"} {
		t.Run(entity, func(t *testing.T) {
			var output bytes.Buffer
			db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = sqlDB.Close() })
			ctx := persistence.WithTxValue(context.Background(), db)
			var locker masterWriteRowsLocker
			if entity == "consultation" {
				locker = NewConsultationRepository(db).(masterWriteRowsLocker)
			} else {
				locker = NewProcedureRepository(db).(masterWriteRowsLocker)
			}
			if err := locker.LockWriteRows(ctx, 7, []uint64{9, 2, 9}); err != nil {
				t.Fatal(err)
			}
			sql := output.String()
			first, second := strings.Index(sql, "id = 2"), strings.Index(sql, "id = 9")
			if first < 0 || second < first || strings.Count(sql, "FOR UPDATE") != 2 || strings.Count(sql, "clinic_id = 7") != 2 || strings.Count(sql, "\"deleted_at\" IS NULL") != 2 {
				t.Fatalf("unsafe lock SQL: %s", sql)
			}
			if err := locker.LockWriteRows(context.Background(), 7, []uint64{2}); err == nil {
				t.Fatal("missing ambient TX accepted")
			}
		})
	}
}
