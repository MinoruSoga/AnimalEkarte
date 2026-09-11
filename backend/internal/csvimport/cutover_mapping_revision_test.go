package csvimport

import (
	"strings"
	"testing"
)

func TestCutoverSourceProvenanceRoutes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		route    *string
		archive  *string
		recovery string
		accept   bool
	}{
		{"complete base", stringPointer("complete_base"), nil, "", true},
		{"reacquire", stringPointer("reacquire"), stringPointer(strings.Repeat("c", 64)), strings.Repeat("a", 64), true},
		{"complete base with archive", stringPointer("complete_base"), stringPointer(strings.Repeat("c", 64)), "", false},
		{"complete base with recovery", stringPointer("complete_base"), nil, strings.Repeat("a", 64), false},
		{"reacquire missing archive", stringPointer("reacquire"), nil, strings.Repeat("a", 64), false},
		{"reacquire missing recovery", stringPointer("reacquire"), stringPointer(strings.Repeat("c", 64)), "", false},
		{"missing route", nil, stringPointer(strings.Repeat("c", 64)), strings.Repeat("a", 64), false},
		{"unknown route", stringPointer("unknown"), nil, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				f.manifest.SourceIdentity.KnjoProvenanceRoute = tc.route
				f.manifest.SourceIdentity.KNJOArchiveSHA256 = tc.archive
				f.manifest.SourceEvidenceSHA256.KNJORecovery = tc.recovery
			})
			for _, expected := range []ExpectedCutoverSource{
				{ManifestSHA256: digest, ClinicCode: "hachioji", ClinicOrdinal: 1, RunID: "run-1"},
				stagingExpectedCutoverSource(digest),
			} {
				_, err := PreflightCutoverBundle(dir, expected)
				if (err == nil) != tc.accept {
					t.Fatalf("route acceptance (%s) = %v, want %v (error: %v)", expected.Provenance.Mode, err == nil, tc.accept, err)
				}
			}
		})
	}
}

func TestCutoverImmutableOutputRevision(t *testing.T) {
	const prefix = "sensitive-local/animalekarte-csv-export/hachioji/run-1-revisions/"
	for _, tc := range []struct {
		path   string
		accept bool
	}{
		{prefix + "review-20260907-01", true},
		{prefix, false},
		{prefix + "../other", false},
		{prefix + "one/two", false},
		{prefix + strings.Repeat("a", 65), false},
		{strings.Replace(prefix, "hachioji", "jouto", 1) + "review", false},
		{strings.Replace(prefix, "run-1", "run-2", 1) + "review", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				f.manifest.OutputDir = tc.path
			})
			for _, expected := range []ExpectedCutoverSource{
				{ManifestSHA256: digest, ClinicCode: "hachioji", ClinicOrdinal: 1, RunID: "run-1"},
				stagingExpectedCutoverSource(digest),
			} {
				_, err := PreflightCutoverBundle(dir, expected)
				if (err == nil) != tc.accept {
					t.Fatalf("revision acceptance (%s) = %v, want %v (error: %v)", expected.Provenance.Mode, err == nil, tc.accept, err)
				}
			}
		})
	}
}

func TestCutoverReviewedMappingRevision(t *testing.T) {
	for _, tc := range []struct {
		name   string
		digest string
		accept bool
	}{
		{"reviewed", "0d7f089990079af28c2ba454ee48188d55a67beab19a9f16a86dee93fec80597", true},
		{"superseded", "888dac89e9b262320cd3afb0c6d223ba5f9c5b943be43f9a56367e81bcad8131", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, digest := writeCutoverFixture(t, func(f *fixtureBundle) {
				f.manifest.StageMappingSHA256 = tc.digest
			})
			_, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{
				ManifestSHA256: digest,
				ClinicCode:     "hachioji",
				ClinicOrdinal:  1,
				RunID:          "run-1",
			})
			if (err == nil) != tc.accept {
				t.Fatalf("mapping revision acceptance = %v, want %v (error: %v)", err == nil, tc.accept, err)
			}
		})
	}
}
