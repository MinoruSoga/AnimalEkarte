package medicalrecord

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CheckupPackageManifest is the strict schema for a versioned import package.
// Clinic ID and actor ID are intentionally absent — request context is authority.
type CheckupPackageManifest struct {
	Namespace           string                   `json:"namespace"`
	Version             string                   `json:"version"`
	ClinicalApprovalRef string                   `json:"clinical_approval_ref"`
	Types               []CheckupPackageTypeDef  `json:"types"`
	Fields              []CheckupPackageFieldDef `json:"fields"`
}

type CheckupPackageTypeDef struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	ParentKey   *string `json:"parent_key,omitempty"`
	Description string  `json:"description"`
	Interval    string  `json:"interval"`
	TargetAge   string  `json:"target_age"`
	SortOrder   int     `json:"sort_order"`
	IsActive    bool    `json:"is_active"`
}

type CheckupPackageFieldDef struct {
	Key           string   `json:"key"`
	TypeKey       string   `json:"type_key"`
	Name          string   `json:"name"`
	FieldType     string   `json:"field_type"`
	Unit          string   `json:"unit"`
	MinValue      *string  `json:"min_value,omitempty"`
	MaxValue      *string  `json:"max_value,omitempty"`
	Options       []string `json:"options"`
	IsProvisional bool     `json:"is_provisional"`
	SortOrder     int      `json:"sort_order"`
}

type CanonicalCheckupPackage struct {
	Manifest CheckupPackageManifest
	Digest   string
}

// ParseAndCanonicalizeCheckupPackage parses strict JSON, validates, and returns
// a stable-key-ordered canonical form with content digest.
func ParseAndCanonicalizeCheckupPackage(raw []byte) (*CanonicalCheckupPackage, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var manifest CheckupPackageManifest
	if err := dec.Decode(&manifest); err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid checkup package manifest: %v", err))
	}
	if dec.More() {
		return nil, apperrors.WrapInvalidInput("invalid checkup package manifest: trailing content")
	}

	if err := validateCheckupPackageManifest(&manifest); err != nil {
		return nil, err
	}
	canonical, err := canonicalizeCheckupPackage(&manifest)
	if err != nil {
		return nil, err
	}
	digest, err := digestCheckupPackage(canonical)
	if err != nil {
		return nil, err
	}
	return &CanonicalCheckupPackage{Manifest: *canonical, Digest: digest}, nil
}

func validateCheckupPackageManifest(m *CheckupPackageManifest) error {
	m.Namespace = strings.TrimSpace(m.Namespace)
	m.Version = strings.TrimSpace(m.Version)
	m.ClinicalApprovalRef = strings.TrimSpace(m.ClinicalApprovalRef)
	if m.Namespace == "" {
		return apperrors.WrapInvalidInput("manifest namespace is required")
	}
	if m.Version == "" {
		return apperrors.WrapInvalidInput("manifest version is required")
	}
	if m.ClinicalApprovalRef == "" {
		return apperrors.WrapInvalidInput("clinical_approval_ref is required")
	}
	if len(m.Types) == 0 {
		return apperrors.WrapInvalidInput("manifest types must not be empty")
	}

	typeKeys, err := normalizeAndValidateCheckupTypes(m)
	if err != nil {
		return err
	}
	return validateCheckupPackageFields(m, typeKeys)
}

func normalizeAndValidateCheckupTypes(m *CheckupPackageManifest) (map[string]struct{}, error) {
	typeKeys := make(map[string]struct{}, len(m.Types))
	for i := range m.Types {
		t := &m.Types[i]
		t.Key = strings.TrimSpace(t.Key)
		t.Name = strings.TrimSpace(t.Name)
		t.Description = strings.TrimSpace(t.Description)
		t.Interval = strings.TrimSpace(t.Interval)
		t.TargetAge = strings.TrimSpace(t.TargetAge)
		if t.ParentKey != nil {
			pk := strings.TrimSpace(*t.ParentKey)
			if pk == "" {
				t.ParentKey = nil
			} else {
				t.ParentKey = &pk
			}
		}
		if t.Key == "" || t.Name == "" {
			return nil, apperrors.WrapInvalidInput("type key and name are required")
		}
		if _, dup := typeKeys[t.Key]; dup {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("duplicate type key %q", t.Key))
		}
		typeKeys[t.Key] = struct{}{}
	}
	for i := range m.Types {
		if m.Types[i].ParentKey != nil {
			if _, ok := typeKeys[*m.Types[i].ParentKey]; !ok {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf("unknown parent_key %q", *m.Types[i].ParentKey))
			}
			if *m.Types[i].ParentKey == m.Types[i].Key {
				return nil, apperrors.WrapInvalidInput("type cannot parent itself")
			}
		}
	}
	return typeKeys, nil
}

func validateCheckupPackageFields(m *CheckupPackageManifest, typeKeys map[string]struct{}) error {
	fieldKeys := make(map[string]struct{}, len(m.Fields))
	for i := range m.Fields {
		f := &m.Fields[i]
		f.Key = strings.TrimSpace(f.Key)
		f.TypeKey = strings.TrimSpace(f.TypeKey)
		f.Name = strings.TrimSpace(f.Name)
		f.FieldType = strings.TrimSpace(f.FieldType)
		f.Unit = strings.TrimSpace(f.Unit)
		if f.Key == "" || f.TypeKey == "" || f.Name == "" {
			return apperrors.WrapInvalidInput("field key, type_key, and name are required")
		}
		if _, ok := typeKeys[f.TypeKey]; !ok {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q references unknown type_key %q", f.Key, f.TypeKey))
		}
		if _, dup := fieldKeys[f.Key]; dup {
			return apperrors.WrapInvalidInput(fmt.Sprintf("duplicate field key %q", f.Key))
		}
		fieldKeys[f.Key] = struct{}{}
		if f.IsProvisional {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q must have is_provisional=false for apply", f.Key))
		}
		ft := model.CheckupFieldType(f.FieldType)
		if !ft.IsValid() {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q has invalid field_type %q", f.Key, f.FieldType))
		}
		if err := normalizeAndValidateCheckupFieldOptions(f, ft); err != nil {
			return err
		}
	}
	return nil
}

func normalizeAndValidateCheckupFieldOptions(f *CheckupPackageFieldDef, ft model.CheckupFieldType) error {
	normalized := make([]string, 0, len(f.Options))
	seenOpt := map[string]struct{}{}
	for _, opt := range f.Options {
		opt = strings.TrimSpace(opt)
		if opt == "" {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q has empty option", f.Key))
		}
		if _, dup := seenOpt[opt]; dup {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q has duplicate option %q", f.Key, opt))
		}
		seenOpt[opt] = struct{}{}
		normalized = append(normalized, opt)
	}
	f.Options = normalized

	switch ft {
	case model.CheckupFieldTypeNumber:
		if len(f.Options) != 0 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("number field %q must have empty options", f.Key))
		}
		minV, maxV, err := parseOptionalDecimalPair(f.MinValue, f.MaxValue)
		if err != nil {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q: %v", f.Key, err))
		}
		if minV != nil && maxV != nil && *minV > *maxV {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q min_value must be <= max_value", f.Key))
		}
	case model.CheckupFieldTypeSingleSelect, model.CheckupFieldTypeMultiSelect, model.CheckupFieldTypeChecklist:
		if len(f.Options) == 0 {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q requires non-empty options", f.Key))
		}
		if f.MinValue != nil || f.MaxValue != nil {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q must not set min/max", f.Key))
		}
	case model.CheckupFieldTypeBoolean, model.CheckupFieldTypeText:
		if len(f.Options) != 0 || f.MinValue != nil || f.MaxValue != nil {
			return apperrors.WrapInvalidInput(fmt.Sprintf("field %q must have empty options and min/max", f.Key))
		}
	}
	return nil
}

func parseOptionalDecimalPair(minS, maxS *string) (*float64, *float64, error) {
	var minV, maxV *float64
	if minS != nil {
		s := strings.TrimSpace(*minS)
		if s == "" {
			return nil, nil, fmt.Errorf("min_value empty string is invalid")
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid min_value")
		}
		// enforce decimal(10,4) string stability via reformat
		formatted := formatDecimal10_4(v)
		minSCopy := formatted
		*minS = minSCopy
		minV = &v
	}
	if maxS != nil {
		s := strings.TrimSpace(*maxS)
		if s == "" {
			return nil, nil, fmt.Errorf("max_value empty string is invalid")
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid max_value")
		}
		formatted := formatDecimal10_4(v)
		*maxS = formatted
		maxV = &v
	}
	return minV, maxV, nil
}

func formatDecimal10_4(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}

func canonicalizeCheckupPackage(m *CheckupPackageManifest) (*CheckupPackageManifest, error) {
	out := *m
	types := append([]CheckupPackageTypeDef(nil), m.Types...)
	fields := append([]CheckupPackageFieldDef(nil), m.Fields...)
	sort.SliceStable(types, func(i, j int) bool { return types[i].Key < types[j].Key })
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].Key < fields[j].Key })
	out.Types = types
	out.Fields = fields
	return &out, nil
}

func digestCheckupPackage(m *CheckupPackageManifest) (string, error) {
	// Deterministic JSON for digest: re-marshal with sorted maps via canonical struct order.
	payload, err := json.Marshal(m)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to marshal canonical checkup package")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}
