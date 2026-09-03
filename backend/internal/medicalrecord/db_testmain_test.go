package medicalrecord

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if host, port, ok := isolatedDockerDBFallback(); ok {
		_ = os.Setenv("DB_HOST", host)
		_ = os.Setenv("DB_PORT", port)
	}
	os.Exit(m.Run())
}
