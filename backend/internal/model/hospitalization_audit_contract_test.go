package model

import "testing"

func TestAuditActionHospitalizationDischargeWithBilling(t *testing.T) {
	if got, want := AuditActionHospitalizationDischargeWithBilling, "hospitalization.discharge_with_billing"; got != want {
		t.Fatalf("AuditActionHospitalizationDischargeWithBilling = %q, want %q", got, want)
	}
}

func TestAuditResourceHospitalization(t *testing.T) {
	if got, want := AuditResourceHospitalization, "hospitalization"; got != want {
		t.Fatalf("AuditResourceHospitalization = %q, want %q", got, want)
	}
}
