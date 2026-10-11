package medicalrecord

import (
	"errors"
	"strings"
	"testing"
)

func TestIDEXXObservedGrammarTokens(t *testing.T) {
	cases := []struct {
		name   string
		tokens []string
		want   []LabDeviceItem
	}{
		{name: "equals_numeric", tokens: []string{"WBC", "=", "5.50", "K/uL", "HGB", "=", "12.8", "g/dL"}, want: []LabDeviceItem{{Code: "WBC", ValueRaw: "5.50", Unit: "K/uL", Flag: ""}, {Code: "HGB", ValueRaw: "12.8", Unit: "g/dL", Flag: ""}}},
		{name: "equals_qualitative_felv_fiv", tokens: []string{"FeLV", "=", "NEG", "FIV", "=", "POS"}, want: []LabDeviceItem{{Code: "FeLV", ValueRaw: "NEG", Unit: "", Flag: ""}, {Code: "FIV", ValueRaw: "POS", Unit: "", Flag: ""}}},
		{name: "legacy_numeric", tokens: []string{"WBC", "5.50", "K/uL", "HCT", "38.5", "%"}, want: []LabDeviceItem{{Code: "WBC", ValueRaw: "5.50", Unit: "K/uL", Flag: ""}, {Code: "HCT", ValueRaw: "38.5", Unit: "%", Flag: ""}}},
		{name: "legacy_mixed_case_felv", tokens: []string{"FeLV", "NEG"}, want: []LabDeviceItem{{Code: "FeLV", ValueRaw: "NEG", Unit: "", Flag: ""}}},
		{name: "unknown_equals_inequality", tokens: []string{"XYZ123", "=", "<1.00", "mg/dL"}, want: []LabDeviceItem{{Code: "XYZ123", ValueRaw: "<1.00", Unit: "mg/dL", Flag: ""}}},
		{name: "equals_without_unit", tokens: []string{"WBC", "=", "5.50"}, want: []LabDeviceItem{{Code: "WBC", ValueRaw: "5.50", Unit: "", Flag: ""}}},
		{name: "missing_value", tokens: []string{"WBC", "="}, want: nil},
		{name: "repeated_equals", tokens: []string{"WBC", "=", "="}, want: nil},
		{name: "next_analyte_after_missing_value", tokens: []string{"WBC", "=", "RBC", "=", "6.20", "M/uL"}, want: []LabDeviceItem{{Code: "RBC", ValueRaw: "6.20", Unit: "M/uL", Flag: ""}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseIDEXXTokens(tc.tokens)
			if len(got) != len(tc.want) {
				t.Fatalf("item count mismatch: got %d want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("synthetic item fields differ at index %d", i)
				}
			}
		})
	}
}

func TestIDEXXObservedGrammarFeLVLexicalGate(t *testing.T) {
	if !isIDEXXAnalyteCode("FeLV") {
		t.Fatal("observed FeLV label is rejected by lexical gate")
	}
	for _, token := range []string{"", "personName", "FeLV=NEG", "FeLV/NEG"} {
		if isIDEXXAnalyteCode(token) {
			t.Fatal("unrelated mixed-case or malformed token was accepted")
		}
	}
}

func TestIDEXXObservedGrammarUnknownQualitativeNeedsReview(t *testing.T) {
	body := strings.Repeat(" ", 20) + "FeLV = NEG FIV = POS"
	payload := append(append([]byte{2}, []byte(body)...), 3)
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil || len(frames) != 1 || len(frames[0].Items) != 2 {
		t.Fatal("synthetic qualitative result shape not preserved")
	}
	if frames[0].Items[0].Code != "FeLV" || frames[0].Items[0].ValueRaw != "NEG" ||
		frames[0].Items[1].Code != "FIV" || frames[0].Items[1].ValueRaw != "POS" || !frames[0].NeedsReview {
		t.Fatal("qualitative labels/values or unknown-code review boundary changed")
	}
}

func TestIDEXXObservedGrammarIncompleteResultRejected(t *testing.T) {
	body := strings.Repeat(" ", 20) + "WBC ="
	payload := append(append([]byte{2}, []byte(body)...), 3)
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if !errors.Is(err, ErrInvalidLabDevicePayload) || len(frames) != 0 {
		t.Fatal("equals delimiter was accepted as a missing result value")
	}
}

func TestIDEXXObservedGrammarPortInquiryIsNotResult(t *testing.T) {
	payload := encodeIDEXXPIMS('1', '0', []byte("I 1 0101261200"))
	if len(payload) != 21 {
		t.Fatal("non-vacuous port-inquiry fixture length mismatch")
	}
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if !errors.Is(err, ErrInvalidLabDevicePayload) || len(frames) != 0 {
		t.Fatal("21-byte PIMS port inquiry was accepted as a measurement")
	}
}
