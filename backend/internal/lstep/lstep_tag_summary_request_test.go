package lstep

import (
	"net/url"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewLstepOwnersByTagQuery(t *testing.T) {
	query, err := newLstepOwnersByTagQuery(url.Values{
		"tag":      {"my_tag"},
		"q":        {"田中"},
		"page":     {"2"},
		"per_page": {"50"},
	})
	if err != nil {
		t.Fatalf("newLstepOwnersByTagQuery returned error: %v", err)
	}

	if query.TagName != "my_tag" {
		t.Errorf("TagName = %q, want my_tag", query.TagName)
	}
	if query.NameQuery != "田中" {
		t.Errorf("NameQuery = %q, want 田中", query.NameQuery)
	}
	if query.Page != 2 {
		t.Errorf("Page = %d, want 2", query.Page)
	}
	if query.PerPage != 50 {
		t.Errorf("PerPage = %d, want 50", query.PerPage)
	}
	input := query.toServiceInput()
	if input.TagName != query.TagName || input.NameQuery != query.NameQuery || input.Page != query.Page || input.PerPage != query.PerPage {
		t.Errorf("toServiceInput() = %+v, want query values %+v", input, query)
	}
}

func TestNewLstepOwnersByTagQuery_Defaults(t *testing.T) {
	query, err := newLstepOwnersByTagQuery(url.Values{"tag": {"my_tag"}})
	if err != nil {
		t.Fatalf("newLstepOwnersByTagQuery returned error: %v", err)
	}

	if query.Page != 1 {
		t.Errorf("Page = %d, want 1", query.Page)
	}
	if query.PerPage != 20 {
		t.Errorf("PerPage = %d, want 20", query.PerPage)
	}
}

func TestNewLstepOwnersByTagQuery_CSV(t *testing.T) {
	query, err := newLstepOwnersByTagQuery(url.Values{
		"tag":      {"my_tag"},
		"q":        {"田中"},
		"format":   {"csv"},
		"page":     {"invalid"},
		"per_page": {"invalid"},
	})
	if err != nil {
		t.Fatalf("newLstepOwnersByTagQuery returned error: %v", err)
	}

	if !query.isCSV() {
		t.Fatal("isCSV = false, want true")
	}
	if query.Page != 0 || query.PerPage != 0 {
		t.Errorf("CSV query should not parse paging: page=%d perPage=%d", query.Page, query.PerPage)
	}
}

func TestNewLstepOwnersByTagQuery_Invalid(t *testing.T) {
	cases := []struct {
		name   string
		values url.Values
	}{
		{"missing tag", url.Values{}},
		{"page zero", url.Values{"tag": {"my_tag"}, "page": {"0"}}},
		{"page non numeric", url.Values{"tag": {"my_tag"}, "page": {"abc"}}},
		{"per_page zero", url.Values{"tag": {"my_tag"}, "per_page": {"0"}}},
		{"per_page too large", url.Values{"tag": {"my_tag"}, "per_page": {"101"}}},
		{"per_page non numeric", url.Values{"tag": {"my_tag"}, "per_page": {"abc"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newLstepOwnersByTagQuery(tc.values); err == nil {
				t.Fatal("newLstepOwnersByTagQuery returned nil error")
			} else if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}
