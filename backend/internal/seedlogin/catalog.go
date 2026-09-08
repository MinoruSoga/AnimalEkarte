// Package seedlogin upserts synthetic STG/local demo login accounts at migrate time.
// The shared demo password is the public constant SharedPassword. Production
// APP_ENV skips this package and must not accept that password at login.
package seedlogin

import (
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	// BundleDir is the schema_migrations key suffix (seeds/003_login).
	// Not a CSV bundle; cmd/migrate records it after the SQL upsert phase.
	BundleDir = "003_login"
	// PermissionGroupGeneral and PermissionGroupExecutive are 002_master
	// names assigned to catalog demo logins. Hayashi is executive on every
	// catalog clinic; the other people stay general on their home clinic.
	PermissionGroupGeneral   = "一般"
	PermissionGroupExecutive = "執行"
	emailPattern             = "stg-staff-%d@example.test"
	// demoStaffBandSize must equal csvimport clinicBandSize. Synthetic login
	// staffs use clinicID * demoStaffBandSize as their ID base (that clinic's
	// cutover EndExclusive). Imported staffs occupy
	// [base+1_000_000, endExclusive), so these IDs cannot trip
	// CUTOVER_REF_BAND_OCCUPIED during make reset / csv-import-preflight.
	demoStaffBandSize = uint64(10_000_000)
)

// AccountSpec is one curated demo login row matching LoginForm DEMO_ACCOUNTS.
type AccountSpec struct {
	StaffID                 uint64
	ClinicID                uint64
	Name                    string
	StaffType               model.StaffType
	Email                   string
	ClinicLabel             string
	OccupationLabel         string
	PermissionGroupName     string
	AssignAllCatalogClinics bool
}

type personTemplate struct {
	suffix                  uint64
	name                    string
	occupationLabel         string
	permissionGroupName     string
	assignAllCatalogClinics bool
}

type clinicBand struct {
	clinicID uint64
	band     uint64
	label    string
}

var personTemplates = []personTemplate{
	{21, "林 文明", "獣医師", PermissionGroupExecutive, true},
	{3, "高橋 純子", "獣医師", PermissionGroupGeneral, false},
	{7, "鈴木 諒平", "獣医師", PermissionGroupGeneral, false},
	{8, "加藤 茉里", "獣医師", PermissionGroupGeneral, false},
	{25, "チャン ハン", "看護師", PermissionGroupGeneral, false},
	{31, "近喰 千瞳", "動物看護師", PermissionGroupGeneral, false},
	{34, "川野 称希", "動物看護師", PermissionGroupGeneral, false},
	{5, "冨田 美佳", "VT", PermissionGroupGeneral, false},
	{6, "井冨 和美", "VT", PermissionGroupGeneral, false},
	{9, "原 梨吏華", "スタッフ", PermissionGroupGeneral, false},
}

var clinicBands = []clinicBand{
	{1, 1 * demoStaffBandSize, "八王子病院"},
	{2, 2 * demoStaffBandSize, "城東センター病院"},
	{3, 3 * demoStaffBandSize, "ノア動物病院　敷島病院"},
	{4, 4 * demoStaffBandSize, "ノア動物病院　Hako bu neco"},
}

// Catalog returns the curated demo login set (LoginForm DEMO_ACCOUNTS).
func Catalog() []AccountSpec {
	out := make([]AccountSpec, 0, len(personTemplates)*len(clinicBands))
	for _, clinic := range clinicBands {
		for _, person := range personTemplates {
			staffID := clinic.band + person.suffix
			out = append(out, AccountSpec{
				StaffID:                 staffID,
				ClinicID:                clinic.clinicID,
				Name:                    person.name,
				StaffType:               staffTypeForOccupation(person.occupationLabel),
				Email:                   EmailForStaffID(staffID),
				ClinicLabel:             clinic.label,
				OccupationLabel:         person.occupationLabel,
				PermissionGroupName:     person.permissionGroupName,
				AssignAllCatalogClinics: person.assignAllCatalogClinics,
			})
		}
	}
	return out
}

// EmailForStaffID is the synthetic login email staff-attach also uses.
func EmailForStaffID(staffID uint64) string {
	return fmt.Sprintf(emailPattern, staffID)
}

// Emails returns catalog emails in catalog order.
func Emails() []string {
	catalog := Catalog()
	out := make([]string, 0, len(catalog))
	for _, row := range catalog {
		out = append(out, row.Email)
	}
	return out
}

func staffTypeForOccupation(label string) model.StaffType {
	if strings.TrimSpace(label) == "獣医師" {
		return model.StaffTypeDoctor
	}
	return model.StaffTypeNurse
}

func catalogClinicIDs() []uint64 {
	out := make([]uint64, 0, len(clinicBands))
	for _, clinic := range clinicBands {
		out = append(out, clinic.clinicID)
	}
	return out
}

func assignmentClinicIDs(spec AccountSpec) []uint64 {
	if spec.AssignAllCatalogClinics {
		return catalogClinicIDs()
	}
	return []uint64{spec.ClinicID}
}
