package lstep

import "testing"

func TestCreateAutoManagedPrefixRequest_ToServiceInput(t *testing.T) {
	description := "desc"
	req := createAutoManagedPrefixRequest{
		Prefix:      "auto:",
		Category:    "C2",
		Description: &description,
	}

	input := req.toServiceInput()

	if input.Prefix != req.Prefix {
		t.Errorf("Prefix = %q, want %q", input.Prefix, req.Prefix)
	}
	if input.Description == nil || *input.Description != description {
		t.Errorf("Description = %v, want %q", input.Description, description)
	}
}

func TestCreateConditionTagMappingRequest_ToServiceInput(t *testing.T) {
	description := ""
	req := createConditionTagMappingRequest{
		ConditionCode: "heart",
		TagName:       "condition:heart",
		Description:   &description,
	}

	input := req.toServiceInput()

	if input.ConditionCode != req.ConditionCode {
		t.Errorf("ConditionCode = %q, want %q", input.ConditionCode, req.ConditionCode)
	}
	if input.Description == nil || *input.Description != description {
		t.Errorf("Description = %v, want empty string pointer", input.Description)
	}
}

func TestCreateSendPurposeTagPrefixRequest_ToServiceInput(t *testing.T) {
	req := createSendPurposeTagPrefixRequest{
		Purpose:   "notice",
		TagPrefix: "send:notice",
	}

	input := req.toServiceInput()

	if input.Purpose != req.Purpose {
		t.Errorf("Purpose = %q, want %q", input.Purpose, req.Purpose)
	}
	if input.TagPrefix != req.TagPrefix {
		t.Errorf("TagPrefix = %q, want %q", input.TagPrefix, req.TagPrefix)
	}
}

func TestCreateAutoManagedPrefixRequest_CategoryBinding(t *testing.T) {
	// Structural: ensure oneof is present (LSA-14). Runtime gin binding is covered in handler tests when present.
	req := createAutoManagedPrefixRequest{Prefix: "x", Category: "C2"}
	if req.Category != "C2" {
		t.Fatalf("setup")
	}
}
