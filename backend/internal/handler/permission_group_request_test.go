package handler

import "testing"

func TestCreatePermissionGroupRequest_ToServiceInput(t *testing.T) {
	input := createPermissionGroupRequest{
		Name:        "Admin",
		Description: "full access",
		Color:       "#ff0000",
		IsActive:    true,
		SortOrder:   1,
	}.toServiceInput()

	if input.Name != "Admin" {
		t.Fatalf("Name = %q, want Admin", input.Name)
	}
	if !input.IsActive {
		t.Fatalf("IsActive = false, want true")
	}
}

func TestUpdatePermissionGroupRequest_ToServiceInput(t *testing.T) {
	name := ""
	description := ""
	color := "#000000"
	sortOrder := 0
	isActive := false

	input := updatePermissionGroupRequest{
		Name:        &name,
		Description: &description,
		Color:       &color,
		SortOrder:   &sortOrder,
		IsActive:    &isActive,
	}.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
}

func TestSetPermissionGroupRulesRequest_ToServiceInput(t *testing.T) {
	input := setPermissionGroupRulesRequest{
		Rules: []permissionGroupRuleInput{
			{Resource: "owners", CanView: true, CanCreate: true},
			{Resource: "pets", CanEdit: true, CanDelete: true},
		},
	}.toServiceInput()

	if len(input) != 2 {
		t.Fatalf("len(input) = %d, want 2", len(input))
	}
	if input[0].Resource != "owners" || !input[0].CanCreate {
		t.Fatalf("input[0] = %#v", input[0])
	}
	if input[1].Resource != "pets" || !input[1].CanDelete {
		t.Fatalf("input[1] = %#v", input[1])
	}
}
