package csvimport

import "fmt"

// CutoverProfile is selected by the operator, independently of provenance mode.
// Candidate selection never authorizes a formal producer or target cutover.
type CutoverProfile string

const (
	CutoverProfileLegacy21           CutoverProfile = "legacy-21"
	CutoverProfileMedicinesCandidate CutoverProfile = "medicines-22-candidate"
)

// ParseCutoverProfile resolves an empty selection to legacy21 and rejects
// unknown profiles. Selecting the medicines candidate does not grant formal
// activation; provenance binding enforces its rehearsal-only gate.
func ParseCutoverProfile(value string) (CutoverProfile, error) {
	switch CutoverProfile(value) {
	case "", CutoverProfileLegacy21:
		return CutoverProfileLegacy21, nil
	case CutoverProfileMedicinesCandidate:
		return CutoverProfileMedicinesCandidate, nil
	default:
		return "", fmt.Errorf("unknown cutover profile")
	}
}

func selectedCutoverProfile(profiles []CutoverProfile) CutoverProfile {
	if len(profiles) == 0 || profiles[0] == "" {
		return CutoverProfileLegacy21
	}
	return profiles[0]
}

func validCutoverProfileSelection(profiles []CutoverProfile) bool {
	if len(profiles) > 1 {
		return false
	}
	_, err := ParseCutoverProfile(string(selectedCutoverProfile(profiles)))
	return err == nil
}

func cutoverProfileStatus(profile CutoverProfile) string {
	if profile == CutoverProfileMedicinesCandidate {
		return "candidate_22_rehearsal_only"
	}
	return "formal_cutover_v1"
}

// bindCutoverManifestProfile validates membership before any database access.
// The profile is never inferred from the manifest's number of tables or pins.
func bindCutoverManifestProfile(manifest CutoverManifest, provenance CutoverProvenanceContract) (CutoverManifest, error) {
	profile, err := ParseCutoverProfile(string(provenance.Profile))
	if err != nil {
		return CutoverManifest{}, err
	}
	if manifest.profile != "" && manifest.profile != profile {
		return CutoverManifest{}, fmt.Errorf("cutover profile binding mismatch")
	}
	if err := validateCutoverCandidateActivation(manifest, provenance); err != nil {
		return CutoverManifest{}, err
	}
	notes := []string{}
	membership := manifest
	if err := normalizeCutoverManifestTables(&membership, CutoverTableSpecs(profile), rehearsalContractDrift(manifest, provenance.Mode), &notes); err != nil {
		return CutoverManifest{}, fmt.Errorf("cutover profile table membership: %w", err)
	}
	manifest.profile = profile
	return manifest, nil
}

// ValidateCutoverProfile checks the operator-selected profile before opening a target.
func ValidateCutoverProfile(manifest CutoverManifest, provenance CutoverProvenanceContract) error {
	_, err := bindCutoverManifestProfile(manifest, provenance)
	return err
}

// validateCutoverCandidateActivation blocks the unapproved 22-table profile.
// Existing digest pins identify the 21-table producer, not an activation receipt.
// Only an explicitly selected rehearsal mode with a rehearsal-only artifact
// may exercise this candidate. There is no manifest or environment override.
func validateCutoverCandidateActivation(manifest CutoverManifest, contract CutoverProvenanceContract) error {
	profile, err := ParseCutoverProfile(string(contract.Profile))
	if err != nil {
		return err
	}
	if profile != CutoverProfileMedicinesCandidate {
		return nil
	}
	if acceptsRehearsalManifest(contract.Mode) && isRehearsalOnlyProducer(manifest) {
		return nil
	}
	return fmt.Errorf("22-table cutover profile activation is blocked pending reviewed producer contract and receipt")
}
