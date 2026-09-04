package testdb

// repotest_test.go — BE8-3: db_setup_test.go（旧 backend/internal/repository/db_setup_test.go）
// から移設した純粋ヘルパーの単体テスト。DB接続不要。

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoMigrateDoesNotClaimInitParity(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "testdb.go"))
	require.NoError(t, err)
	source := string(raw)
	require.Contains(t, source, "AutoMigrate does not apply 001_init.sql")
	require.NotContains(t, strings.ToLower(source), "apply 001_init.sql to testdb")
}

func TestQuotePostgresIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{
			name:       "simple identifier",
			identifier: "ekarte_db_test",
			want:       `"ekarte_db_test"`,
		},
		{
			name:       "embedded quote is escaped",
			identifier: `ekarte"db_test`,
			want:       `"ekarte""db_test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, quotePostgresIdentifier(tt.identifier))
		})
	}
}

func TestIsDedicatedTestDatabaseName(t *testing.T) {
	tests := []struct {
		name string
		db   string
		want bool
	}{
		{name: "ekarte_db_test is dedicated", db: "ekarte_db_test", want: true},
		{name: "suffix _test is dedicated", db: "anything_test", want: true},
		{name: "live ekarte_db is refused", db: "ekarte_db", want: false},
		{name: "empty is refused", db: "", want: false},
		{name: "testdb without suffix is refused", db: "testdb", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDedicatedTestDatabaseName(tt.db))
		})
	}
}

func TestIsDeadlockError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "40P01 is deadlock",
			err:  &pgconn.PgError{Code: "40P01"},
			want: true,
		},
		{
			name: "wrapped 40P01 is deadlock",
			err:  fmt.Errorf("truncate: %w", &pgconn.PgError{Code: "40P01"}),
			want: true,
		},
		{
			name: "lock timeout is not deadlock",
			err:  &pgconn.PgError{Code: "55P03"},
			want: false,
		},
		{
			name: "other pg error is not deadlock",
			err:  &pgconn.PgError{Code: "23503"},
			want: false,
		},
		{
			name: "plain error is not deadlock",
			err:  errors.New("plain"),
			want: false,
		},
		{
			name: "nil is not deadlock",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDeadlockError(tt.err))
		})
	}
}

func TestIsLockTimeoutError(t *testing.T) {
	assert.True(t, isLockTimeoutError(&pgconn.PgError{Code: "55P03"}))
	assert.True(t, isLockTimeoutError(fmt.Errorf("lock: %w", &pgconn.PgError{Code: "55P03"})))
	assert.False(t, isLockTimeoutError(&pgconn.PgError{Code: "40P01"}))
	assert.False(t, isLockTimeoutError(errors.New("plain")))
}

func TestIsSafeTruncateTableName(t *testing.T) {
	tests := []struct {
		name  string
		table string
		want  bool
	}{
		{name: "owners", table: "owners", want: true},
		{name: "snake_case", table: "permission_groups", want: true},
		{name: "empty", table: "", want: false},
		{name: "space", table: "owners CASCADE", want: false},
		{name: "semicolon", table: "owners;drop", want: false},
		{name: "uppercase", table: "Owners", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isSafeTruncateTableName(tt.table))
		})
	}
}

func TestCoreTruncateUsesSingleStatement(t *testing.T) {
	assert.Equal(
		t,
		"TRUNCATE TABLE billing_refunds, payments, billings, medical_records, owners CASCADE",
		coreTruncateSQL,
	)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "testdb.go"))
	require.NoError(t, err)
	source := string(raw)
	require.NotContains(t, source, `db.Exec("TRUNCATE TABLE billing_refunds CASCADE")`)
	require.NotContains(t, source, `db.Exec("TRUNCATE TABLE payments CASCADE")`)
	require.NotContains(t, source, `db.Exec("TRUNCATE TABLE billings CASCADE")`)
	require.NotContains(t, source, `db.Exec("TRUNCATE TABLE medical_records CASCADE")`)
	require.NotContains(t, source, `db.Exec("TRUNCATE TABLE owners CASCADE")`)
}

func TestIsDuplicateDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "duplicate_database is true",
			err:  &pgconn.PgError{Code: "42P04"},
			want: true,
		},
		{
			name: "wrapped duplicate_database is true",
			err:  fmt.Errorf("create failed: %w", &pgconn.PgError{Code: "42P04"}),
			want: true,
		},
		{
			name: "different pg error is false",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "plain error is false",
			err:  errors.New("plain error"),
			want: false,
		},
		{
			name: "nil is false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDuplicateDatabaseError(tt.err))
		})
	}
}

func TestEnumValuesEqual_TestEnumAppendedValues(t *testing.T) {
	t.Run("enumValuesEqual: identical values match", func(t *testing.T) {
		assert.True(t, enumValuesEqual([]string{"manual", "trimming"}, []string{"'manual'", "'trimming'"}))
	})
	t.Run("enumValuesEqual: different length does not match", func(t *testing.T) {
		assert.False(t, enumValuesEqual([]string{"manual"}, []string{"'manual'", "'trimming'"}))
	})
	t.Run("enumValuesEqual: same length different value does not match", func(t *testing.T) {
		assert.False(t, enumValuesEqual([]string{"manual", "hospitalization"}, []string{"'manual'", "'trimming'"}))
	})

	t.Run("enumAppendedValues: pure trailing append is detected (item_source G12-2 case)", func(t *testing.T) {
		appended, ok := enumAppendedValues(
			[]string{"medical_record", "manual", "hospitalization"},
			[]string{"'medical_record'", "'manual'", "'hospitalization'", "'trimming'"},
		)
		assert.True(t, ok)
		assert.Equal(t, []string{"'trimming'"}, appended)
	})
	t.Run("enumAppendedValues: no new values is not an append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'a'", "'b'"})
		assert.False(t, ok)
	})
	t.Run("enumAppendedValues: reordered values is not a pure append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'b'", "'a'", "'c'"})
		assert.False(t, ok)
	})
	t.Run("enumAppendedValues: removed value is not a pure append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'a'"})
		assert.False(t, ok)
	})
}
