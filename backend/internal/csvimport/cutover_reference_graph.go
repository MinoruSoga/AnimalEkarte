package csvimport

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type cutoverReferenceKind uint8

const (
	cutoverCSVParent cutoverReferenceKind = iota + 1
	cutoverClinicSeed
	cutoverSpeciesSeed
	cutoverPlaceholderSeed
	cutoverUnavailableParent
)

type cutoverCSVReference struct {
	column   string
	parent   string
	nullable bool
	kind     cutoverReferenceKind
}

// cutoverCSVReferences classifies every reference supplied by the 21-table
// contract, independently of producer eligibility predicates. Empty CSV fields
// become SQL NULL; they are allowed only for explicitly nullable references.
// Non-imported masters are either pinned seeds or must stay empty, never guessed.
func cutoverCSVReferences() map[string][]cutoverCSVReference {
	return map[string][]cutoverCSVReference{
		"staffs":                       {{"clinic_id", "clinics", false, cutoverClinicSeed}},
		"procedures":                   {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"parent_id", "procedures", true, cutoverCSVParent}},
		"merchandise_items":            {{"clinic_id", "clinics", false, cutoverClinicSeed}},
		"owners":                       {{"clinic_id", "clinics", false, cutoverClinicSeed}},
		"pets":                         {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"owner_id", "owners", false, cutoverCSVParent}, {"animal_species_id", "animal_species", false, cutoverSpeciesSeed}},
		"medical_records":              {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"owner_id", "owners", true, cutoverCSVParent}, {"pet_id", "pets", true, cutoverCSVParent}, {"appointment_id", "appointments", true, cutoverCSVParent}, {"doctor_id", "staffs", true, cutoverCSVParent}, {"entered_by", "staffs", true, cutoverCSVParent}},
		"inquiries":                    {{"medical_record_id", "medical_records", false, cutoverCSVParent}, {"staff_id", "staffs", true, cutoverCSVParent}},
		"clinical_plans":               {{"medical_record_id", "medical_records", false, cutoverCSVParent}},
		"vital_records":                {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"medical_record_id", "medical_records", true, cutoverCSVParent}, {"pet_id", "pets", false, cutoverCSVParent}, {"staff_id", "staffs", true, cutoverCSVParent}},
		"appointments":                 {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"owner_id", "owners", true, cutoverCSVParent}, {"pet_id", "pets", true, cutoverCSVParent}, {"reservation_type_id", "reservation_types", false, cutoverPlaceholderSeed}, {"doctor_id", "staffs", true, cutoverCSVParent}},
		"appointment_trimming_details": {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"appointment_id", "appointments", false, cutoverCSVParent}},
		"billings":                     {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"medical_record_id", "medical_records", true, cutoverCSVParent}, {"owner_id", "owners", true, cutoverCSVParent}, {"pet_id", "pets", true, cutoverCSVParent}},
		"billing_items":                {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"billing_id", "billings", false, cutoverCSVParent}},
		"payments":                     {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"billing_id", "billings", false, cutoverCSVParent}, {"payment_method_id", "payment_methods", false, cutoverPlaceholderSeed}, {"paid_by", "staffs", true, cutoverCSVParent}},
		"payment_splits":               {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"billing_id", "billings", false, cutoverCSVParent}, {"payment_method_id", "payment_methods", false, cutoverPlaceholderSeed}, {"paid_by", "staffs", true, cutoverCSVParent}},
		"estimates":                    {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"medical_record_id", "medical_records", true, cutoverCSVParent}, {"owner_id", "owners", true, cutoverCSVParent}, {"pet_id", "pets", true, cutoverCSVParent}, {"created_by", "staffs", true, cutoverCSVParent}},
		"estimate_items":               {{"estimate_id", "estimates", false, cutoverCSVParent}, {"consultation_id", "consultations", true, cutoverUnavailableParent}, {"procedure_id", "procedures", true, cutoverCSVParent}, {"medicine_id", "medicines", true, cutoverUnavailableParent}, {"merchandise_item_id", "merchandise_items", true, cutoverCSVParent}},
		"exams":                        {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"medical_record_id", "medical_records", true, cutoverCSVParent}, {"pet_id", "pets", true, cutoverCSVParent}, {"exam_type_id", "exam_types", false, cutoverPlaceholderSeed}},
		"exam_results":                 {{"exam_id", "exams", false, cutoverCSVParent}},
		"vaccines":                     {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"inventory_id", "inventory_items", true, cutoverUnavailableParent}, {"parent_id", "vaccines", true, cutoverCSVParent}},
		"vaccinations":                 {{"clinic_id", "clinics", false, cutoverClinicSeed}, {"medical_record_id", "medical_records", true, cutoverCSVParent}, {"pet_id", "pets", true, cutoverCSVParent}, {"vaccine_id", "vaccines", false, cutoverCSVParent}, {"doctor_id", "staffs", true, cutoverCSVParent}},
	}
}

// validateCutoverReferenceGraph scans each CSV once, retaining only numeric ID
// sets for referenced parents, not source rows or per-edge copies. Parent tables
// precede children; self references are checked after that complete table so a
// later row may be the parent. Every scan rehashes its exact opened bytes.
func validateCutoverReferenceGraph(sourceDir string, manifest CutoverManifest, allowPermutation bool) error {
	specs := CutoverTableSpecs()
	references := cutoverCSVReferences()
	if err := validateCutoverReferenceInventory(specs, references); err != nil {
		return err
	}
	parents := make(map[string]map[uint64]struct{})
	for _, refs := range references {
		for _, ref := range refs {
			if ref.kind == cutoverCSVParent {
				parents[ref.parent] = nil
			}
		}
	}
	for _, spec := range specs {
		table, err := cutoverManifestTableByName(manifest, spec.Name)
		if err != nil {
			return fmt.Errorf("source reference graph: %w", err)
		}
		ids := make(map[uint64]struct{})
		selfParents := make(map[uint64]cutoverSelfParentRef)
		selfColumn, hasSelfRef := cutoverSelfReferenceColumn(spec.Name)
		selfEdges := make(map[uint64]cutoverSelfEdge)
		_, retainIDs := parents[spec.Name]
		err = streamCutoverReferenceCSV(sourceDir, spec, table, allowPermutation, func(row []string, header []string, indexes map[string]int, csvLine int64) error {
			if err := validateCutoverRow(spec, header, row, indexes, manifest.IDBand, csvLine); err != nil {
				return err
			}
			if retainIDs {
				id, ok := cutoverReferenceID(row[indexes["id"]])
				if !ok {
					return fmt.Errorf("table %s column id: source reference ID is invalid", spec.Name)
				}
				if _, duplicate := ids[id]; duplicate {
					return fmt.Errorf("table %s column id: duplicate source reference ID", spec.Name)
				}
				ids[id] = struct{}{}
			}
			for _, ref := range references[spec.Name] {
				value := row[indexes[ref.column]]
				if err := validateCutoverCSVReference(spec.Name, ref, value, parents, selfParents, csvLine); err != nil {
					return err
				}
			}
			if hasSelfRef {
				id, ok := cutoverReferenceID(row[indexes["id"]])
				if !ok {
					return fmt.Errorf("table %s column id: source reference ID is invalid", spec.Name)
				}
				if value := row[indexes[selfColumn]]; value != "" {
					if parent, ok := cutoverReferenceID(value); ok {
						selfEdges[id] = cutoverSelfEdge{column: selfColumn, parent: parent, csvLine: csvLine}
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		for id, selfRef := range selfParents {
			if _, found := ids[id]; !found {
				return fmt.Errorf("table %s column %s row %d: source CSV reference to %s is missing",
					spec.Name, selfRef.column, selfRef.csvLine, spec.Name)
			}
		}
		if err := validateCutoverSelfReferenceAcyclic(spec.Name, selfEdges); err != nil {
			return err
		}
		if retainIDs {
			parents[spec.Name] = ids
		}
	}
	return nil
}

// cutoverSelfParentRef records where a self-referencing row declared its
// parent so a miss can name the declaring row, never the identifier value.
type cutoverSelfParentRef struct {
	column  string
	csvLine int64
}

// cutoverSelfEdge is one declared child→parent link inside a self-referencing
// table; csvLine lets a cycle report point at the declaring row.
type cutoverSelfEdge struct {
	column  string
	parent  uint64
	csvLine int64
}

// cutoverSelfReferenceColumn returns the CSV column that references the table
// itself (procedures.parent_id, vaccines.parent_id), if the contract declares
// one. Apply uses it to order staged inserts because the FK is not deferrable.
func cutoverSelfReferenceColumn(table string) (string, bool) {
	for _, ref := range cutoverCSVReferences()[table] {
		if ref.kind == cutoverCSVParent && ref.parent == table {
			return ref.column, true
		}
	}
	return "", false
}

// validateCutoverSelfReferenceAcyclic walks each declared parent chain. A
// self-reference cycle can never be satisfied by any insert order under the
// non-deferrable FK, so the bundle must fail here, before apply commits a
// partial clinic band on the per-table path.
func validateCutoverSelfReferenceAcyclic(table string, edges map[uint64]cutoverSelfEdge) error {
	visited := make(map[uint64]uint64, len(edges))
	var stamp uint64
	for node := range edges {
		stamp++
		for current := node; ; {
			edge, isChild := edges[current]
			if !isChild {
				break
			}
			if visited[current] == stamp {
				return fmt.Errorf("table %s column %s row %d: self-reference cycle in source CSV",
					table, edge.column, edge.csvLine)
			}
			visited[current] = stamp
			current = edge.parent
		}
	}
	return nil
}

func validateCutoverCSVReference(table string, ref cutoverCSVReference, value string, parents map[string]map[uint64]struct{}, selfParents map[uint64]cutoverSelfParentRef, csvLine int64) error {
	if value == "" && ref.nullable {
		return nil
	}
	valid := false
	switch ref.kind {
	case cutoverCSVParent:
		id, ok := cutoverReferenceID(value)
		if ok && ref.parent == table {
			if _, seen := selfParents[id]; !seen {
				selfParents[id] = cutoverSelfParentRef{column: ref.column, csvLine: csvLine}
			}
			return nil
		}
		_, exists := parents[ref.parent][id]
		valid = ok && exists
	case cutoverClinicSeed:
		valid = value == "{{CLINIC_ID}}"
	case cutoverSpeciesSeed:
		id, ok := cutoverReferenceID(value)
		valid = value == "{{FALLBACK_ANIMAL_SPECIES_ID}}" || ok && id >= 1 && id <= 6
	case cutoverPlaceholderSeed:
		valid = placeholderAllowed(table, ref.column, value, value)
	case cutoverUnavailableParent:
		return fmt.Errorf("table %s column %s row %d: non-imported source parent %s must be empty", table, ref.column, csvLine, ref.parent)
	}
	if !valid {
		return fmt.Errorf("table %s column %s row %d: source CSV reference to %s is missing or invalid", table, ref.column, csvLine, ref.parent)
	}
	return nil
}

func cutoverReferenceID(value string) (uint64, bool) {
	// Match existing ParseInt semantics, including equivalent +/leading-zero
	// spellings, without string-key aliases or uint64 overflow acceptance.
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return uint64(id), true
}

func validateCutoverReferenceInventory(specs []CutoverTableSpec, references map[string][]cutoverCSVReference) error {
	if len(references) != len(specs) {
		return fmt.Errorf("source reference inventory: table coverage mismatch")
	}
	available := make(map[string]bool, len(specs))
	for _, spec := range specs {
		available[spec.Name] = true
		refs, found := references[spec.Name]
		if !found {
			return fmt.Errorf("source reference inventory: table is unclassified")
		}
		columns := make(map[string]bool, len(spec.Columns))
		for _, column := range spec.Columns {
			if strings.HasSuffix(column, "_id") || column == "paid_by" || column == "entered_by" || column == "created_by" {
				columns[column] = false
			}
		}
		for _, ref := range refs {
			seen, exists := columns[ref.column]
			if !exists || seen || ref.parent == "" || ref.kind < cutoverCSVParent || ref.kind > cutoverUnavailableParent {
				return fmt.Errorf("source reference inventory: invalid classification")
			}
			if ref.kind == cutoverCSVParent && !available[ref.parent] {
				return fmt.Errorf("source reference inventory: parent order mismatch")
			}
			columns[ref.column] = true
		}
		for _, classified := range columns {
			if !classified {
				return fmt.Errorf("source reference inventory: column is unclassified")
			}
		}
	}
	return nil
}

func streamCutoverReferenceCSV(sourceDir string, spec CutoverTableSpec, table CutoverManifestTable, allowPermutation bool, visit func(row []string, header []string, indexes map[string]int, csvLine int64) error) error {
	file, err := openStableOwnerOnlyFile(cutoverCSVPath(sourceDir, table))
	if err != nil {
		return fmt.Errorf("table %s: source reference CSV cannot be opened safely", spec.Name)
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || info.Size() > maxCutoverCSVBytes {
		return fmt.Errorf("table %s: source reference CSV size is invalid", spec.Name)
	}
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(file, maxCutoverCSVBytes+1), hash)))
	reader.ReuseRecord = true
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("table %s: source reference CSV header changed", spec.Name)
	}
	indexes, _, err := cutoverColumnIndexes(spec, header, allowPermutation)
	if err != nil {
		return fmt.Errorf("table %s: source reference CSV header changed", spec.Name)
	}
	// ReuseRecord may overwrite the header slice once data rows stream, so the
	// visit callback receives a stable clone for name-based column checks.
	header = append([]string(nil), header...)
	var count int64
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return fmt.Errorf("table %s: source reference CSV cannot be parsed", spec.Name)
		}
		count++
		if err := visit(row, header, indexes, count+1); err != nil {
			return err
		}
	}
	if count != table.RowCount || !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), table.SHA256) {
		return fmt.Errorf("table %s: source reference CSV digest or row count changed", spec.Name)
	}
	return nil
}
