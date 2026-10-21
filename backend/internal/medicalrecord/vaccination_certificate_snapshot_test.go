package medicalrecord

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCertificateSnapshotJSON is a valid v1 normal snapshot used across tests.
const testCertificateSnapshotJSON = `{
	"schema_version": 1,
	"kind": "normal",
	"source": {
		"vaccination_id": 100, "version": 2, "vaccine_id": 7,
		"vaccine_name": "混合ワクチン", "administered_on": "2026-10-01",
		"next_scheduled_on": "2027-10-01", "lots": ["A1", "B2", null, null],
		"doctor_name": "田中"
	},
	"owner": {"name": "山田", "name_kana": "ヤマダ", "postal_code": "100-0001", "address": "東京都", "phone": "03-0000-0000"},
	"pet": {"name": "ポチ", "species": "犬", "breed": "柴", "sex": "male", "coat_color": "茶", "age_text": "3歳", "birth_date": "2023-04-01"},
	"clinic": {"name": "テスト動物病院", "director_name": "院長", "postal_code": "150-0001", "address": "渋谷区", "phone": "03-1111-2222", "fax": "03-1111-3333", "email": "c@example.jp", "url": "https://example.jp", "veterinarian_association": "協会", "designated_veterinarian": "指定医", "veterinarian_name": "獣医"},
	"display_issue_date": "2026-10-20",
	"inputs": {"certificate_number": null, "injection_tag_number": null, "dog_registration_tag_number": null, "distinguishing_features": null, "injection_amount_text": null, "body_size": null, "registration_fee": null, "injection_fee": null, "tag_fee": null, "total": null},
	"eligibility": {"rabies_eligible": false, "vaccine_id": 7, "vaccine_version": 3, "provenance": "master_snapshot"},
	"render_manifest": {"renderer_version": "ekarte-certificate-pdf/1", "font_asset_version": "adobe-japan1-heisei-kakugo-w5-unembedded/1", "templates": {"portrait": "normal-portrait/1", "landscape": "normal-landscape/1"}}
}`

// testCertificateRabiesSnapshot returns a valid v1 rabies snapshot for the
// given vaccination id/version.
func testCertificateRabiesSnapshot(vaccinationID, version int64) string {
	return fmt.Sprintf(`{
	"schema_version": 1,
	"kind": "rabies",
	"source": {
		"vaccination_id": %d, "version": %d, "vaccine_id": 7,
		"vaccine_name": "狂犬病", "administered_on": "2026-10-01",
		"next_scheduled_on": null, "lots": [null, null, null, null],
		"doctor_name": "田中"
	},
	"owner": {"name": "山田", "name_kana": "ヤマダ", "postal_code": "100-0001", "address": "東京都", "phone": "03-0000-0000"},
	"pet": {"name": "ポチ", "species": "犬", "breed": "柴", "sex": "male", "coat_color": "茶", "age_text": "3歳", "birth_date": "2023-04-01"},
	"clinic": {"name": "テスト動物病院", "director_name": "院長", "postal_code": "150-0001", "address": "渋谷区", "phone": "03-1111-2222", "fax": "03-1111-3333", "email": null, "url": null, "veterinarian_association": "協会", "designated_veterinarian": "指定医", "veterinarian_name": "獣医"},
	"display_issue_date": "2026-10-20",
	"inputs": {"certificate_number": "R-123", "injection_tag_number": "T-9", "dog_registration_tag_number": "D-5", "distinguishing_features": "なし", "injection_amount_text": "1mL", "body_size": "medium", "registration_fee": 3000, "injection_fee": 3000, "tag_fee": 550, "total": 6550},
	"eligibility": {"rabies_eligible": true, "vaccine_id": 7, "vaccine_version": 3, "provenance": "master_snapshot"},
	"render_manifest": {"renderer_version": "ekarte-certificate-pdf/1", "font_asset_version": "adobe-japan1-heisei-kakugo-w5-unembedded/1", "templates": {"rabies": "rabies-a4-landscape/1"}}
}`, vaccinationID, version)
}

func mustDecodeSnapshot(t *testing.T, raw string) *certificateSnapshot {
	t.Helper()
	snap, err := decodeCertificateSnapshot([]byte(raw))
	require.NoError(t, err)
	return snap
}

// certErrCode asserts err is a certificateError and returns its wire code.
func certErrCode(t *testing.T, err error) string {
	t.Helper()
	require.Error(t, err)
	var ce *certificateError
	require.True(t, errors.As(err, &ce), "expected *certificateError, got %T", err)
	return ce.code
}

func TestDecodeCertificateSnapshot_Valid(t *testing.T) {
	snap := mustDecodeSnapshot(t, testCertificateSnapshotJSON)
	assert.Equal(t, "normal", snap.Kind)
	assert.Equal(t, "2026-10-20", snap.DisplayIssueDate)
	assert.Equal(t, "ekarte-certificate-pdf/1", snap.RendererVersion)
	assert.Equal(t, "adobe-japan1-heisei-kakugo-w5-unembedded/1", snap.FontAssetVersion)
	assert.Equal(t, "normal-portrait/1", snap.Templates["portrait"])
	assert.Equal(t, "山田", certText(snap.Owner.Name))
	assert.Len(t, snap.Source.Lots, 4)
	assert.Equal(t, "A1", certText(snap.Source.Lots[0]))
	assert.Nil(t, snap.Source.Lots[2])
	rabies := mustDecodeSnapshot(t, testCertificateRabiesSnapshot(100, 2))
	assert.Equal(t, "rabies", rabies.Kind)
	assert.Equal(t, "rabies-a4-landscape/1", rabies.Templates["rabies"])
	require.NotNil(t, rabies.Inputs.Total)
	assert.Equal(t, int64(6550), *rabies.Inputs.Total)
}

func TestDecodeCertificateSnapshot_SchemaVersionGate(t *testing.T) {
	_, err := decodeCertificateSnapshot([]byte(`{}`))
	assert.Equal(t, "dependency-changed", certErrCode(t, err), "missing schema_version")
	for _, doc := range []string{
		`{"schema_version":2}`,
		`{"schema_version":"1"}`,
		`{"schema_version":1.0}`,
		`{"schema_version":null}`,
	} {
		_, err := decodeCertificateSnapshot([]byte(doc))
		assert.Equal(t, "dependency-changed", certErrCode(t, err), doc)
	}
}

func TestDecodeCertificateSnapshot_ShapeFailures(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"not an object", `[1,2]`},
		{"malformed json", `{"schema_version":1,`},
		{"duplicate key", `{"schema_version":1,"schema_version":1}`},
		{"trailing close brace", `{"schema_version":1}}`},
		{"trailing close bracket", `{"schema_version":1}]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeCertificateSnapshot([]byte(tc.raw))
			assert.Equal(t, "result-unavailable", certErrCode(t, err))
		})
	}
	mutations := []struct {
		name string
		in   string
	}{
		{"unknown top key", `{"schema_version":1,"kind":"normal","source":{},"owner":{},"pet":{},"clinic":{},"display_issue_date":"2026-10-20","inputs":{},"eligibility":{},"render_manifest":{},"extra":1}`},
		{"missing top key", `{"schema_version":1,"kind":"normal","owner":{},"pet":{},"clinic":{},"display_issue_date":"2026-10-20","inputs":{},"eligibility":{},"render_manifest":{}}`},
		{"bad kind", `{"schema_version":1,"kind":"other","source":{},"owner":{},"pet":{},"clinic":{},"display_issue_date":"2026-10-20","inputs":{},"eligibility":{},"render_manifest":{}}`},
		{"lots wrong length", `{"schema_version":1,"kind":"normal","source":{"vaccination_id":1,"version":1,"vaccine_id":1,"vaccine_name":null,"administered_on":"2026-01-01","next_scheduled_on":null,"lots":[null,null,null],"doctor_name":null},"owner":{"name":null,"name_kana":null,"postal_code":null,"address":null,"phone":null},"pet":{"name":null,"species":null,"breed":null,"sex":null,"coat_color":null,"age_text":null,"birth_date":null},"clinic":{"name":null,"director_name":null,"postal_code":null,"address":null,"phone":null,"fax":null,"email":null,"url":null,"veterinarian_association":null,"designated_veterinarian":null,"veterinarian_name":null},"display_issue_date":"2026-10-20","inputs":{"certificate_number":null,"injection_tag_number":null,"dog_registration_tag_number":null,"distinguishing_features":null,"injection_amount_text":null,"body_size":null,"registration_fee":null,"injection_fee":null,"tag_fee":null,"total":null},"eligibility":{"rabies_eligible":null,"vaccine_id":1,"vaccine_version":null,"provenance":"master_snapshot"},"render_manifest":{"renderer_version":"r","font_asset_version":"f","templates":{"portrait":"p","landscape":"l"}}}`},
		{"unknown template key", `{"schema_version":1,"kind":"normal","source":{"vaccination_id":1,"version":1,"vaccine_id":1,"vaccine_name":null,"administered_on":"2026-01-01","next_scheduled_on":null,"lots":[null,null,null,null],"doctor_name":null},"owner":{"name":null,"name_kana":null,"postal_code":null,"address":null,"phone":null},"pet":{"name":null,"species":null,"breed":null,"sex":null,"coat_color":null,"age_text":null,"birth_date":null},"clinic":{"name":null,"director_name":null,"postal_code":null,"address":null,"phone":null,"fax":null,"email":null,"url":null,"veterinarian_association":null,"designated_veterinarian":null,"veterinarian_name":null},"display_issue_date":"2026-10-20","inputs":{"certificate_number":null,"injection_tag_number":null,"dog_registration_tag_number":null,"distinguishing_features":null,"injection_amount_text":null,"body_size":null,"registration_fee":null,"injection_fee":null,"tag_fee":null,"total":null},"eligibility":{"rabies_eligible":null,"vaccine_id":1,"vaccine_version":null,"provenance":"master_snapshot"},"render_manifest":{"renderer_version":"r","font_asset_version":"f","templates":{"portrait":"p","landscape":"l","bogus":"x"}}}`},
		{"normal inputs non-null", `{"schema_version":1,"kind":"normal","source":{"vaccination_id":1,"version":1,"vaccine_id":1,"vaccine_name":null,"administered_on":"2026-01-01","next_scheduled_on":null,"lots":[null,null,null,null],"doctor_name":null},"owner":{"name":null,"name_kana":null,"postal_code":null,"address":null,"phone":null},"pet":{"name":null,"species":null,"breed":null,"sex":null,"coat_color":null,"age_text":null,"birth_date":null},"clinic":{"name":null,"director_name":null,"postal_code":null,"address":null,"phone":null,"fax":null,"email":null,"url":null,"veterinarian_association":null,"designated_veterinarian":null,"veterinarian_name":null},"display_issue_date":"2026-10-20","inputs":{"certificate_number":"X","injection_tag_number":null,"dog_registration_tag_number":null,"distinguishing_features":null,"injection_amount_text":null,"body_size":null,"registration_fee":null,"injection_fee":null,"tag_fee":null,"total":null},"eligibility":{"rabies_eligible":null,"vaccine_id":1,"vaccine_version":null,"provenance":"master_snapshot"},"render_manifest":{"renderer_version":"r","font_asset_version":"f","templates":{"portrait":"p","landscape":"l"}}}`},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeCertificateSnapshot([]byte(tc.in))
			assert.Equal(t, "result-unavailable", certErrCode(t, err))
		})
	}
}

func TestCertificateValidity_Priority(t *testing.T) {
	assert.Equal(t, "source_deleted", certificateValidity(true, 1, 1, true))
	assert.Equal(t, "source_deleted", certificateValidity(true, 5, 1, false))
	assert.Equal(t, "source_changed", certificateValidity(false, 5, 1, true))
	assert.Equal(t, "superseded", certificateValidity(false, 5, 5, true))
	assert.Equal(t, "current", certificateValidity(false, 5, 5, false))
}

func TestAvailableCertificateLayouts(t *testing.T) {
	normal := map[string]string{"portrait": "p", "landscape": "l"}
	assert.Equal(t, []string{"portrait", "landscape"}, availableCertificateLayouts("normal", normal))
	partial := map[string]string{"landscape": "l"}
	assert.Equal(t, []string{"landscape"}, availableCertificateLayouts("normal", partial))
	rabies := map[string]string{"rabies": "r", "portrait": "p"}
	assert.Equal(t, []string{"rabies"}, availableCertificateLayouts("rabies", rabies), "kinds must never mix")
	assert.Empty(t, availableCertificateLayouts("normal", map[string]string{}))
}
