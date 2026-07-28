package billing

import (
	"strings"
	"testing"
)

func TestConfirmBillingConfirmationRequest_NormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name     string
		request  confirmBillingConfirmationRequest
		wantMemo string
		wantErr  bool
	}{
		{
			name:     "accepts empty optional memo",
			request:  confirmBillingConfirmationRequest{},
			wantMemo: "",
		},
		{
			name:     "trims memo",
			request:  confirmBillingConfirmationRequest{Memo: "  note \n"},
			wantMemo: "note",
		},
		{
			name:     "accepts memo at 1000 character boundary",
			request:  confirmBillingConfirmationRequest{Memo: strings.Repeat("界", billingConfirmationMemoMaxLength)},
			wantMemo: strings.Repeat("界", billingConfirmationMemoMaxLength),
		},
		{
			name:    "rejects memo over 1000 characters",
			request: confirmBillingConfirmationRequest{Memo: strings.Repeat("界", billingConfirmationMemoMaxLength+1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := tt.request.normalizeAndValidate()

			if tt.wantErr {
				if err == nil {
					t.Fatal("normalizeAndValidate() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeAndValidate() error = %v", err)
			}
			if normalized.Memo != tt.wantMemo {
				t.Errorf("Memo = %q, want %q", normalized.Memo, tt.wantMemo)
			}
		})
	}
}

func TestReturnBillingConfirmationRequest_NormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name             string
		request          returnBillingConfirmationRequest
		wantReturnReason string
		wantMemo         string
		wantErr          bool
	}{
		{
			name:    "rejects missing return reason",
			request: returnBillingConfirmationRequest{},
			wantErr: true,
		},
		{
			name:    "rejects whitespace-only return reason",
			request: returnBillingConfirmationRequest{ReturnReason: " \t\n "},
			wantErr: true,
		},
		{
			name:             "trims return reason and memo",
			request:          returnBillingConfirmationRequest{ReturnReason: "  reason \n", Memo: "  note  "},
			wantReturnReason: "reason",
			wantMemo:         "note",
		},
		{
			name: "accepts both fields at their character boundaries",
			request: returnBillingConfirmationRequest{
				ReturnReason: strings.Repeat("界", billingConfirmationReturnReasonMaxLength),
				Memo:         strings.Repeat("界", billingConfirmationMemoMaxLength),
			},
			wantReturnReason: strings.Repeat("界", billingConfirmationReturnReasonMaxLength),
			wantMemo:         strings.Repeat("界", billingConfirmationMemoMaxLength),
		},
		{
			name:    "rejects return reason over 500 characters",
			request: returnBillingConfirmationRequest{ReturnReason: strings.Repeat("界", billingConfirmationReturnReasonMaxLength+1)},
			wantErr: true,
		},
		{
			name: "rejects memo over 1000 characters",
			request: returnBillingConfirmationRequest{
				ReturnReason: "reason",
				Memo:         strings.Repeat("界", billingConfirmationMemoMaxLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := tt.request.normalizeAndValidate()

			if tt.wantErr {
				if err == nil {
					t.Fatal("normalizeAndValidate() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeAndValidate() error = %v", err)
			}
			if normalized.ReturnReason != tt.wantReturnReason {
				t.Errorf("ReturnReason = %q, want %q", normalized.ReturnReason, tt.wantReturnReason)
			}
			if normalized.Memo != tt.wantMemo {
				t.Errorf("Memo = %q, want %q", normalized.Memo, tt.wantMemo)
			}
		})
	}
}

func TestConfirmBillingConfirmationRequest_ToServiceInput(t *testing.T) {
	const authenticatedStaffID = uint64(10)
	req := confirmBillingConfirmationRequest{Memo: ""}

	input := req.toServiceInput(authenticatedStaffID)

	if input.ConfirmedBy != authenticatedStaffID {
		t.Errorf("ConfirmedBy = %d, want authenticated staff %d", input.ConfirmedBy, authenticatedStaffID)
	}
	if input.Memo != req.Memo {
		t.Errorf("Memo = %q, want %q", input.Memo, req.Memo)
	}
}

func TestReturnBillingConfirmationRequest_ToServiceInput(t *testing.T) {
	const authenticatedStaffID = uint64(10)
	req := returnBillingConfirmationRequest{ReturnReason: "missing item", Memo: ""}

	input := req.toServiceInput(authenticatedStaffID)

	if input.ReturnedBy != authenticatedStaffID {
		t.Errorf("ReturnedBy = %d, want authenticated staff %d", input.ReturnedBy, authenticatedStaffID)
	}
	if input.ReturnReason != req.ReturnReason {
		t.Errorf("ReturnReason = %q, want %q", input.ReturnReason, req.ReturnReason)
	}
	if input.Memo != req.Memo {
		t.Errorf("Memo = %q, want %q", input.Memo, req.Memo)
	}
}
