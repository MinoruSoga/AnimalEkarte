package handler

import "testing"

func TestConfirmBillingConfirmationRequest_ToServiceInput(t *testing.T) {
	req := confirmBillingConfirmationRequest{ConfirmedBy: 10, Memo: ""}

	input := req.toServiceInput()

	if input.ConfirmedBy != req.ConfirmedBy {
		t.Errorf("ConfirmedBy = %d, want %d", input.ConfirmedBy, req.ConfirmedBy)
	}
	if input.Memo != req.Memo {
		t.Errorf("Memo = %q, want %q", input.Memo, req.Memo)
	}
}

func TestReturnBillingConfirmationRequest_ToServiceInput(t *testing.T) {
	req := returnBillingConfirmationRequest{ReturnedBy: 10, ReturnReason: "missing item", Memo: ""}

	input := req.toServiceInput()

	if input.ReturnedBy != req.ReturnedBy {
		t.Errorf("ReturnedBy = %d, want %d", input.ReturnedBy, req.ReturnedBy)
	}
	if input.ReturnReason != req.ReturnReason {
		t.Errorf("ReturnReason = %q, want %q", input.ReturnReason, req.ReturnReason)
	}
	if input.Memo != req.Memo {
		t.Errorf("Memo = %q, want %q", input.Memo, req.Memo)
	}
}
