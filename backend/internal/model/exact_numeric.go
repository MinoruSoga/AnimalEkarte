package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgtype"
)

// ExactNumeric preserves a PostgreSQL NUMERIC value without converting its decimal digits to binary float.
// Int/Exp describe the stored value; Valid distinguishes SQL NULL. It selects no rounding or allocation rule.
// Nullable columns use *ExactNumeric; API representation and calculator conversion remain owning-consumer work.
type ExactNumeric struct {
	pgtype.Numeric
}

// GormDataType prevents GORM from inferring a boolean from pgtype.Numeric.Valid.
// Column tags retain the actual NUMERIC precision and scale from the migration.
func (ExactNumeric) GormDataType() string { return "numeric" }

// Keep the existing pgx exact-decimal I/O contracts when adapting it for GORM.
var (
	_ sql.Scanner      = (*ExactNumeric)(nil)
	_ driver.Valuer    = ExactNumeric{}
	_ json.Marshaler   = ExactNumeric{}
	_ json.Unmarshaler = (*ExactNumeric)(nil)
)
