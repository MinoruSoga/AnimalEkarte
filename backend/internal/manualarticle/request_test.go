package manualarticle

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestUpsertManualArticleRequest_ToServiceInput(t *testing.T) {
	req := UpsertManualArticleRequest{
		Title:        "Title",
		OrderValue:   1.5,
		Section:      "section",
		BodyMarkdown: "# Body",
	}

	input := req.toServiceInput(model.ManualCategoryScreens, "slug")

	if input.Category != model.ManualCategoryScreens {
		t.Errorf("Category = %q, want %q", input.Category, model.ManualCategoryScreens)
	}
	if input.Slug != "slug" {
		t.Errorf("Slug = %q, want slug", input.Slug)
	}
	if input.BodyMarkdown != req.BodyMarkdown {
		t.Errorf("BodyMarkdown = %q, want %q", input.BodyMarkdown, req.BodyMarkdown)
	}
}
