package lintscan

import (
	"testing"
)

// petGrandchildFinding is the pet-facing finding shape kept for existing test
// contracts. Detection is owned by the single registry in
// grandchild_parent_clinic_correlation_lint_test.go (grandchildParentTargets).
type petGrandchildFinding struct {
	file     string
	line     int
	function string
	model    string
	table    string
	detail   string
}

type petGrandchildStats struct {
	filesParsed int
	readsByKey  map[string]int
}

// analyzeFilePetGrandchildParentClinicCorrelation delegates to the unified
// grandchild parent-clinic correlation analyzer. The duplicate petGrandchildTargets
// registry was removed in SEC-SWEEP-02-S1; targets live only in grandchildParentTargets.
func analyzeFilePetGrandchildParentClinicCorrelation(filename string, src []byte) ([]petGrandchildFinding, petGrandchildStats, error) {
	mainFindings, mainStats, err := analyzeFileGrandchildParentClinicCorrelation(filename, src)
	if err != nil {
		return nil, petGrandchildStats{}, err
	}
	findings := make([]petGrandchildFinding, 0, len(mainFindings))
	for _, f := range mainFindings {
		findings = append(findings, petGrandchildFinding{
			file:     f.file,
			line:     f.line,
			function: f.function,
			model:    f.modelName,
			table:    f.childTable,
			detail:   f.detail,
		})
	}
	readsByKey := mainStats.readsByKey
	if readsByKey == nil {
		readsByKey = make(map[string]int)
	}
	return findings, petGrandchildStats{
		filesParsed: mainStats.filesParsed,
		readsByKey:  readsByKey,
	}, nil
}

func TestPetGrandchildParentClinicCorrelation_AnalyzerFixtures(t *testing.T) {
	fixtures := []struct {
		name     string
		src      string
		wantFind int
	}{
		{
			name: "daily record read missing hospitalization clinic correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.DailyRecord
	db.Where("daily_records.hospitalization_id = ?", 1).Find(&records)
}`,
			wantFind: 1,
		},
		{
			name: "addendum read missing medical record clinic correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.MedicalRecordAddendum
	db.Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.medical_record_id = ?", 1, 2).Find(&records)
}`,
			wantFind: 1,
		},
		{
			name: "daily record read with parent correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.DailyRecord
	db.Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").Where("daily_records.hospitalization_id = ?", 1).Find(&records)
}`,
			wantFind: 0,
		},
		{
			name: "addendum read with parent correlation",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.MedicalRecordAddendum
	db.Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id").Where("medical_record_addenda.medical_record_id = ?", 2).Find(&records)
}`,
			wantFind: 0,
		},
		{
			name: "historical read does not require pet lifecycle predicates",
			src: `package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.DailyRecord
	db.Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").Where("daily_records.hospitalization_id = ?", 1).Find(&records)
}`,
			wantFind: 0,
		},
		{
			name: "unrelated table ignored",
			src: `package fixture
func f(db DB) {
	db.Where("pets.id = ?", 1).Find(&pets)
}`,
			wantFind: 0,
		},
	}
	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			findings, _, err := analyzeFilePetGrandchildParentClinicCorrelation("repository.go", []byte(tt.src))
			if err != nil {
				t.Fatalf("analyze fixture: %v", err)
			}
			if len(findings) != tt.wantFind {
				t.Fatalf("findings = %d, want %d: %#v", len(findings), tt.wantFind, findings)
			}
		})
	}
}

func TestPetGrandchildParentClinicCorrelation_LocationAgnostic(t *testing.T) {
	src := []byte(`package fixture
import "github.com/animal-ekarte/backend/internal/model"
func f(db DB) {
	var records []model.MedicalRecordAddendum
	db.Where("medical_record_addenda.medical_record_id = ?", 2).Find(&records)
}`)
	for _, file := range []string{"repository.go", "medicalrecord/repository.go", "medicalrecord/sub/repository.go"} {
		t.Run(file, func(t *testing.T) {
			findings, _, err := analyzeFilePetGrandchildParentClinicCorrelation(file, src)
			if err != nil {
				t.Fatalf("analyze fixture: %v", err)
			}
			if len(findings) != 1 {
				t.Fatalf("findings = %d, want 1", len(findings))
			}
			if findings[0].file != file && file != "repository.go" {
				t.Fatalf("finding file = %q, want %q", findings[0].file, file)
			}
		})
	}
}

func TestPetGrandchildParentClinicCorrelation_RealSourceHasNoUncorrelatedReads(t *testing.T) {
	tree := moduleInternalSource(t)
	if len(tree) < 40 {
		t.Fatalf("only %d production files discovered; lint would be vacuous", len(tree))
	}

	totalStats := petGrandchildStats{readsByKey: make(map[string]int)}
	var findings []petGrandchildFinding
	for key, src := range tree {
		fileFindings, stats, err := analyzeFilePetGrandchildParentClinicCorrelation(legacyLintKey(key), src)
		if err != nil {
			t.Fatalf("analyze %s: %v", key, err)
		}
		findings = append(findings, fileFindings...)
		totalStats.filesParsed += stats.filesParsed
		for statKey, count := range stats.readsByKey {
			totalStats.readsByKey[statKey] += count
		}
	}

	for _, required := range []string{
		"dailyRecordRepository.FindByHospitalizationID|DailyRecord",
		"dailyRecordRepository.FindByHospitalizationIDAndDate|DailyRecord",
		"medicalRecordAddendumRepository.FindByID|MedicalRecordAddendum",
		"medicalRecordAddendumRepository.FindByMedicalRecordID|MedicalRecordAddendum",
		"medicalRecordImageRepository.FindByMedicalRecordID|MedicalRecordImage",
		"billingItemRepository.FindByID|BillingItem",
		"examinationRepository.FindAllItemsByExamID|ExamResult",
	} {
		if totalStats.readsByKey[required] == 0 {
			t.Fatalf("required read site %q was not detected; lint would be vacuous", required)
		}
	}

	// Shared site-level residual allowlist (same as main RealSource gate).
	mainFindings := make([]grandchildParentFinding, 0, len(findings))
	for _, finding := range findings {
		mainFindings = append(mainFindings, grandchildParentFinding{
			file:       finding.file,
			line:       finding.line,
			function:   finding.function,
			modelName:  finding.model,
			childTable: finding.table,
			detail:     finding.detail,
		})
		key := grandchildParentResidualSiteKey(finding.file, finding.function, finding.model, finding.table)
		for _, site := range grandchildParentClinicCorrelationResidualSites {
			if grandchildParentResidualSiteKeyFromSite(site) == key {
				t.Logf("residual pet-facing finding (follow-up repair unit): %s:%d %s %s(%s): %s",
					finding.file, finding.line, finding.function, finding.model, finding.table, finding.detail)
				break
			}
		}
	}
	for _, errMsg := range grandchildParentResidualGateErrors(mainFindings, grandchildParentClinicCorrelationResidualSites) {
		t.Error(errMsg)
	}
}
