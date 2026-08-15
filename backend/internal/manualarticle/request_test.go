package manualarticle

import (
	"bytes"
	"testing"

	"github.com/gin-gonic/gin/binding"

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

func TestUpsertManualArticleRequest_BindingLimits(t *testing.T) {
	ok := UpsertManualArticleRequest{
		Title: "Title", Section: "section", BodyMarkdown: "# Body",
	}
	if err := binding.Validator.ValidateStruct(ok); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	tooLongTitle := UpsertManualArticleRequest{
		Title: string(bytes.Repeat([]byte("a"), 256)), Section: "s", BodyMarkdown: "b",
	}
	if err := binding.Validator.ValidateStruct(tooLongTitle); err == nil {
		t.Fatal("expected max title error")
	}
}
