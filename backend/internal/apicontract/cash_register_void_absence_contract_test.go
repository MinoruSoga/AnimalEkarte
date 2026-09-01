package apicontract

import (
	"os"
	"strings"
	"testing"
)

func TestCashRegisterCloseVoidIsNotAPublicOperation(t *testing.T) {
	src, err := os.ReadFile("../../docs/api.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}
	for _, forbidden := range []string{
		"/cash-register/closes/{id}/void:",
		"VoidCashRegisterCloseRequest:",
		"CashRegisterVoidResponse:",
		"voidCashRegisterClose",
	} {
		if strings.Contains(string(src), forbidden) {
			t.Fatalf("cash-close void must not be a public OpenAPI operation: found %q", forbidden)
		}
	}
}
