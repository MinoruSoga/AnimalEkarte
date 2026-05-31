package handler

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateMedicalRecordAddendumRequest_ToServiceInput(t *testing.T) {
	req := CreateMedicalRecordAddendumRequest{AfterText: "after", Reason: "reason"}

	input := req.toServiceInput(1, 2)

	if input.MedicalRecordID != 1 {
		t.Errorf("MedicalRecordID = %d, want 1", input.MedicalRecordID)
	}
	if input.AuthorUserID != 2 {
		t.Errorf("AuthorUserID = %d, want 2", input.AuthorUserID)
	}
	if input.AfterText != req.AfterText {
		t.Errorf("AfterText = %q, want %q", input.AfterText, req.AfterText)
	}
}

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

func TestUpdateLstepSettingsRequest_ToServiceInput(t *testing.T) {
	isSyncEnabled := false
	cpmVersion := "v2"
	noahLTV := int64(10000)
	lookbackDays := 30
	req := updateLstepSettingsRequest{
		LstepAPIKey:                  "key",
		IsSyncEnabled:                &isSyncEnabled,
		CPMVersion:                   &cpmVersion,
		CPMV1NoahLTV:                 &noahLTV,
		HealthPreventionLookbackDays: &lookbackDays,
	}

	input := req.toServiceInput()

	if input.LstepAPIKey != req.LstepAPIKey {
		t.Errorf("LstepAPIKey = %q, want %q", input.LstepAPIKey, req.LstepAPIKey)
	}
	if input.IsSyncEnabled == nil || *input.IsSyncEnabled {
		t.Errorf("IsSyncEnabled = %v, want false pointer", input.IsSyncEnabled)
	}
	if input.CPMVersion == nil || *input.CPMVersion != cpmVersion {
		t.Errorf("CPMVersion = %v, want %q", input.CPMVersion, cpmVersion)
	}
	if input.CPMV1NoahLTV == nil || *input.CPMV1NoahLTV != noahLTV {
		t.Errorf("CPMV1NoahLTV = %v, want %d", input.CPMV1NoahLTV, noahLTV)
	}
	if input.HealthPreventionLookbackDays == nil || *input.HealthPreventionLookbackDays != lookbackDays {
		t.Errorf("HealthPreventionLookbackDays = %v, want %d", input.HealthPreventionLookbackDays, lookbackDays)
	}
}
