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

func TestCashRegisterCloseVoidIsNotInRepositoryContract(t *testing.T) {
	for _, path := range []string{
		"../billing/cash_register_close_repository.go",
		"../billing/cash_register_service.go",
	} {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(src), "Void(") {
			t.Fatalf("cash-close void must not remain in repository or service contract: found in %s", path)
		}
	}
}
