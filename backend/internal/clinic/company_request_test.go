package clinic

import "testing"

func TestUpdateCompanyRequest_ToServiceInput(t *testing.T) {
	name := ""
	email := "clinic@example.com"
	req := UpdateCompanyRequest{Name: &name, Email: &email}

	input := req.ToServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Errorf("Name = %v, want empty string pointer", input.Name)
	}
	if input.Email == nil || *input.Email != email {
		t.Errorf("Email = %v, want %q", input.Email, email)
	}
}
