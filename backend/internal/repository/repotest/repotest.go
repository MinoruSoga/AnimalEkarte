// Package repotest is a temporary BE9-2F compatibility surface.
// New tests import internal/testdb directly.
package repotest

import (
	"testing"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// EnumType is a compatibility alias for testdb.EnumType.
type EnumType = testdb.EnumType

// SharedTestSchemaEnumTypes exposes the canonical testdb schema list.
var SharedTestSchemaEnumTypes = testdb.SharedTestSchemaEnumTypes

// EnumValueRe exposes the canonical testdb enum parser.
var EnumValueRe = testdb.EnumValueRe

// EnsureAutoMigrated delegates to testdb.
func EnsureAutoMigrated(db *gorm.DB, models ...any) error {
	return testdb.EnsureAutoMigrated(db, models...)
}

// MarkAutoMigrated delegates to testdb.
func MarkAutoMigrated(models ...any) {
	testdb.MarkAutoMigrated(models...)
}

// CloseSharedTestDB delegates to testdb.
func CloseSharedTestDB() {
	testdb.CloseSharedTestDB()
}

// SetupTestDB delegates to testdb.
func SetupTestDB(t *testing.T) *gorm.DB {
	return testdb.SetupTestDB(t)
}

// EnsureClinicSettingsTable delegates to testdb.
func EnsureClinicSettingsTable(t *testing.T, db *gorm.DB) {
	testdb.EnsureClinicSettingsTable(t, db)
}

// MakeTestOwner delegates to testdb.
func MakeTestOwner(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
) *model.Owner {
	return testdb.MakeTestOwner(t, db, clinicID, name)
}

// SetupIsolatedTestDB delegates to testdb.
func SetupIsolatedTestDB(t *testing.T) *gorm.DB {
	return testdb.SetupIsolatedTestDB(t)
}
