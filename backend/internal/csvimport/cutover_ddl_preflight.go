package csvimport

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type cutoverNumericClass uint8

const (
	cutoverNumericInvalid cutoverNumericClass = iota
	cutoverNumericFinite
	cutoverNumericNaN
	cutoverNumericPosInf
	cutoverNumericNegInf
)

var cutoverIntegerLiteral = regexp.MustCompile(`^[+-]?[0-9]+$`)
var cutoverKatakanaLiteral = regexp.MustCompile(`[\x{30A1}-\x{30F6}]`)

const cutoverNumericMaxExponent = int64(1<<31-1) / 2

type cutoverDDLEnumRule struct {
	column   string
	domain   string
	nullable bool
}

var cutoverDDLEnumDomains = map[string][]string{
	"membership_type":       {"non_member", "member", "deceased", "transferred"},
	"pet_gender":            {"male", "female", "unknown"},
	"pet_status":            {"alive", "deceased"},
	"anesthesia_type":       {"none", "local", "sedation", "general"},
	"tax_type":              {"included", "excluded", "exempt"},
	"visit_type":            {"first", "revisit"},
	"reservation_status":    {"confirmed", "pending", "cancelled", "checked_in", "in_consultation", "accounting", "completed", "no_show", "tentative"},
	"reservation_source":    {"manual", "line"},
	"medical_record_status": {"draft", "finalized"},
	"body_weight_unit":      {"Kg", "g"},
	"item_category":         {"examination", "test", "procedure", "surgery", "medicine", "food", "goods", "other", "vaccine", "trimming", "hotel", "training", "house_call"},
	"estimate_status":       {"draft", "sent", "approved", "rejected"},
	"next_schedule_type":    {"3weeks", "4weeks", "1year", "other"},
	"vaccine_species":       {"dog", "cat", "both"},
}

var cutoverDDLEnumColumns = map[string][]cutoverDDLEnumRule{
	"owners":            {{"membership_type", "membership_type", false}},
	"pets":              {{"gender", "pet_gender", false}, {"status", "pet_status", false}},
	"procedures":        {{"anesthesia", "anesthesia_type", true}, {"tax_type", "tax_type", false}},
	"appointments":      {{"visit_type", "visit_type", false}, {"status", "reservation_status", true}, {"source", "reservation_source", false}},
	"medical_records":   {{"status", "medical_record_status", true}, {"visit_type", "visit_type", true}},
	"vital_records":     {{"weight_unit", "body_weight_unit", true}},
	"billing_items":     {{"category", "item_category", false}, {"tax_type", "tax_type", false}},
	"estimates":         {{"status", "estimate_status", true}},
	"estimate_items":    {{"category", "item_category", false}, {"tax_type", "tax_type", false}},
	"merchandise_items": {{"category", "item_category", false}, {"tax_type", "tax_type", false}},
	"vaccinations":      {{"next_schedule_type", "next_schedule_type", true}},
	"vaccines":          {{"species", "vaccine_species", true}},
}

var cutoverDDLRequiredColumns = map[string][]string{
	"appointments":                 {"start_time", "end_time"},
	"medical_records":              {"date"},
	"vital_records":                {"medical_record_id", "recorded_at"},
	"billings":                     {"scheduled_date"},
	"exams":                        {"date"},
	"vaccinations":                 {"date"},
	"estimates":                    {"created_at"},
	"payments":                     {"created_at"},
	"payment_splits":               {"created_at"},
	"inquiries":                    {"medical_record_id"},
	"clinical_plans":               {"medical_record_id"},
	"appointment_trimming_details": {"appointment_id"},
}

type cutoverDDLBooleanRule struct {
	column   string
	nullable bool
}

var cutoverDDLBooleanColumns = map[string][]cutoverDDLBooleanRule{
	"staffs":            {{"is_active", false}, {"reservation_visible", false}},
	"procedures":        {{"is_active", false}, {"is_surgery", false}},
	"medicines":         {{"is_active", false}, {"is_non_insurance", false}},
	"merchandise_items": {{"is_active", false}},
	"owners":            {{"is_dangerous", false}, {"dm_preference", true}},
	"billing_items":     {{"is_insurance_applicable", true}},
	"estimate_items":    {{"is_insurance_applicable", true}},
	"vaccines":          {{"is_active", false}},
}

var cutoverKanaColumns = map[string][]string{
	"owners": {"name_kana"},
	"pets":   {"name_kana"},
}

var cutoverDDLQuantityColumns = map[string][]string{
	"billing_items":  {"quantity"},
	"estimate_items": {"quantity"},
}

func validateCutoverDDLRow(spec CutoverTableSpec, row []string, indexes map[string]int, csvLine int64) error {
	for _, column := range cutoverDDLRequiredColumns[spec.Name] {
		if strings.TrimSpace(row[indexes[column]]) == "" {
			return fmt.Errorf("table %s column %s row %d: value must not be empty", spec.Name, column, csvLine)
		}
	}
	for _, rule := range cutoverDDLEnumColumns[spec.Name] {
		value := row[indexes[rule.column]]
		if value == "" {
			if rule.nullable {
				continue
			}
			return fmt.Errorf("table %s column %s row %d: value must not be empty", spec.Name, rule.column, csvLine)
		}
		if !slices.Contains(cutoverDDLEnumDomains[rule.domain], value) {
			return fmt.Errorf("table %s column %s row %d: value is outside the enum domain", spec.Name, rule.column, csvLine)
		}
	}
	for _, column := range cutoverKanaColumns[spec.Name] {
		if cutoverKatakanaLiteral.MatchString(row[indexes[column]]) {
			return fmt.Errorf("table %s column %s row %d: value must not contain katakana", spec.Name, column, csvLine)
		}
	}
	for _, column := range cutoverDDLQuantityColumns[spec.Name] {
		if err := validateCutoverScaledQuantity(spec.Name, column, row[indexes[column]], csvLine); err != nil {
			return err
		}
	}
	if spec.Name == "vital_records" {
		if err := validateCutoverVitalRow(row, indexes, csvLine); err != nil {
			return err
		}
	}
	for _, rule := range cutoverDDLBooleanColumns[spec.Name] {
		value := row[indexes[rule.column]]
		if value == "" && rule.nullable {
			continue
		}
		if _, ok := cutoverPGBoolean(value); !ok {
			return fmt.Errorf("table %s column %s row %d: value must be a boolean literal", spec.Name, rule.column, csvLine)
		}
	}
	return validateCutoverDDLScalarRow(spec, row, indexes, csvLine)
}

func validateCutoverScaledQuantity(table, column, text string, csvLine int64) error {
	if text == "" {
		return fmt.Errorf("table %s column %s row %d: value must not be empty", table, column, csvLine)
	}
	value, class := parseCutoverPGNumeric(text)
	switch class {
	case cutoverNumericInvalid:
		return fmt.Errorf("table %s column %s row %d: value must be a numeric literal", table, column, csvLine)
	case cutoverNumericNaN:
		return nil
	case cutoverNumericPosInf, cutoverNumericNegInf:
		return fmt.Errorf("table %s column %s row %d: value exceeds the numeric(10,1) precision bound", table, column, csvLine)
	}
	if value.sign <= 0 || !cutoverScaledQuantityPositive(value) {
		return fmt.Errorf("table %s column %s row %d: value must be a positive finite numeric(10,1)", table, column, csvLine)
	}
	return nil
}

func cutoverScaledQuantityPositive(d cutoverDecimal) bool {
	if d.decPos > 9 || d.decPos < -1 {
		return false
	}
	var scaled int64
	need := d.decPos + 1
	next := byte('0')
	if need > 0 {
		for i := int64(0); i < need; i++ {
			digit := byte('0')
			if i < int64(len(d.digits)) {
				digit = d.digits[i]
			}
			scaled = scaled*10 + int64(digit-'0')
		}
		if int64(len(d.digits)) > need {
			next = d.digits[need]
		}
	} else if need == 0 && len(d.digits) > 0 {
		next = d.digits[0]
	}
	if next >= '5' {
		scaled++
	}
	return scaled >= 1 && scaled <= 9_999_999_999
}

func validateCutoverVitalRow(row []string, indexes map[string]int, csvLine int64) error {
	if text := row[indexes["temperature"]]; text != "" {
		value, class := parseCutoverPGNumeric(text)
		if class != cutoverNumericFinite ||
			cutoverDecimalCompare(value, "3", 2) < 0 || cutoverDecimalCompare(value, "5", 2) > 0 {
			return fmt.Errorf("table vital_records column temperature row %d: value must be between 30 and 50 inclusive", csvLine)
		}
	}
	if text := row[indexes["heart_rate"]]; text != "" {
		if err := validateCutoverVitalInteger(text, 499); err != nil {
			return fmt.Errorf("table vital_records column heart_rate row %d: value must be an integer between 1 and 499", csvLine)
		}
	}
	if text := row[indexes["respiration_rate"]]; text != "" {
		if err := validateCutoverVitalInteger(text, 199); err != nil {
			return fmt.Errorf("table vital_records column respiration_rate row %d: value must be an integer between 1 and 199", csvLine)
		}
	}
	if text := row[indexes["weight"]]; text != "" {
		value, class := parseCutoverPGNumeric(text)
		switch class {
		case cutoverNumericInvalid, cutoverNumericNegInf:
			return fmt.Errorf("table vital_records column weight row %d: value must be a numeric greater than zero", csvLine)
		case cutoverNumericFinite:
			if value.sign <= 0 {
				return fmt.Errorf("table vital_records column weight row %d: value must be a numeric greater than zero", csvLine)
			}
			if value.decPos > cutoverPGNumericMaxIntegerDigits || int64(len(value.digits))-value.decPos > cutoverPGNumericMaxFractionalDigits {
				return fmt.Errorf("table vital_records column weight row %d: value exceeds PostgreSQL numeric bounds", csvLine)
			}
		}
	}
	return nil
}

func validateCutoverVitalInteger(text string, upper int64) error {
	if !cutoverIntegerLiteral.MatchString(strings.TrimSpace(text)) {
		return fmt.Errorf("not an integer")
	}
	value, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if err != nil || value < 1 || value > upper {
		return fmt.Errorf("out of bounds")
	}
	return nil
}

type cutoverDecimal struct {
	sign   int
	digits string
	decPos int64
	dscale int64
}

func parseCutoverPGNumeric(text string) (cutoverDecimal, cutoverNumericClass) {
	trimmed := strings.TrimSpace(text)
	switch strings.ToLower(trimmed) {
	case "nan":
		return cutoverDecimal{}, cutoverNumericNaN
	case "infinity", "inf", "+infinity", "+inf":
		return cutoverDecimal{}, cutoverNumericPosInf
	case "-infinity", "-inf":
		return cutoverDecimal{}, cutoverNumericNegInf
	}
	body := trimmed
	sign := 1
	if strings.HasPrefix(body, "+") {
		body = body[1:]
	} else if strings.HasPrefix(body, "-") {
		sign = -1
		body = body[1:]
	}
	mantissa := body
	exponent := ""
	hasExponent := false
	if i := strings.IndexAny(body, "eE"); i >= 0 {
		mantissa = body[:i]
		exponent = body[i+1:]
		hasExponent = true
	}
	intPart, fracPart := mantissa, ""
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		if strings.IndexByte(mantissa[dot+1:], '.') >= 0 {
			return cutoverDecimal{}, cutoverNumericInvalid
		}
		intPart, fracPart = mantissa[:dot], mantissa[dot+1:]
	}
	if intPart == "" && fracPart == "" {
		return cutoverDecimal{}, cutoverNumericInvalid
	}
	for i := 0; i < len(intPart)+len(fracPart); i++ {
		var c byte
		if i < len(intPart) {
			c = intPart[i]
		} else {
			c = fracPart[i-len(intPart)]
		}
		if c < '0' || c > '9' {
			return cutoverDecimal{}, cutoverNumericInvalid
		}
	}
	exp := int64(0)
	if hasExponent {
		expNeg := false
		if exponent != "" && (exponent[0] == '+' || exponent[0] == '-') {
			expNeg = exponent[0] == '-'
			exponent = exponent[1:]
		}
		if exponent == "" {
			return cutoverDecimal{}, cutoverNumericInvalid
		}
		for i := 0; i < len(exponent); i++ {
			c := exponent[i]
			if c < '0' || c > '9' {
				return cutoverDecimal{}, cutoverNumericInvalid
			}
			exp = exp*10 + int64(c-'0')
			if exp > cutoverNumericMaxExponent {
				return cutoverDecimal{}, cutoverNumericInvalid
			}
		}
		if expNeg {
			exp = -exp
		}
	}
	digits := intPart + fracPart
	leadingZeros := 0
	for leadingZeros < len(digits) && digits[leadingZeros] == '0' {
		leadingZeros++
	}
	significant := digits[leadingZeros:]
	dscale := max(int64(0), int64(len(fracPart))-exp)
	if significant == "" {
		return cutoverDecimal{sign: 0, dscale: dscale}, cutoverNumericFinite
	}
	decPos := int64(len(intPart)) + exp - int64(leadingZeros)
	return cutoverDecimal{sign: sign, digits: significant, decPos: decPos, dscale: dscale}, cutoverNumericFinite
}

func cutoverDecimalCompare(d cutoverDecimal, boundDigits string, boundDecPos int64) int {
	if d.sign <= 0 {
		return -1
	}
	if d.decPos != boundDecPos {
		if d.decPos < boundDecPos {
			return -1
		}
		return 1
	}
	n := max(len(d.digits), len(boundDigits))
	for i := 0; i < n; i++ {
		a, b := byte('0'), byte('0')
		if i < len(d.digits) {
			a = d.digits[i]
		}
		if i < len(boundDigits) {
			b = boundDigits[i]
		}
		if a != b {
			if a < b {
				return -1
			}
			return 1
		}
	}
	return 0
}

func cutoverPGBoolean(text string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "1", "t", "tr", "tru", "true", "y", "ye", "yes", "on":
		return true, true
	case "0", "f", "fa", "fal", "fals", "false", "n", "no", "of", "off":
		return false, true
	default:
		return false, false
	}
}

type cutoverUniqueState struct {
	activeNames   map[string]struct{}
	textKeys      map[string]map[string]struct{}
	idKeys        map[string]map[uint64]struct{}
	tentativePets map[uint64]struct{}
}

func newCutoverUniqueState(table string) *cutoverUniqueState {
	state := &cutoverUniqueState{}
	switch table {
	case "merchandise_items":
		state.activeNames = map[string]struct{}{}
	case "owners":
		state.textKeys = map[string]map[string]struct{}{"email": {}, "phone": {}}
	case "medical_records":
		state.textKeys = map[string]map[string]struct{}{"record_no": {}}
		state.idKeys = map[string]map[uint64]struct{}{"appointment_id": {}}
	case "estimates":
		state.textKeys = map[string]map[string]struct{}{"estimate_no": {}}
	case "inquiries", "clinical_plans":
		state.idKeys = map[string]map[uint64]struct{}{"medical_record_id": {}}
	case "appointment_trimming_details":
		state.idKeys = map[string]map[uint64]struct{}{"appointment_id": {}}
	case "appointments":
		state.tentativePets = map[uint64]struct{}{}
	default:
		return nil
	}
	return state
}

func (state *cutoverUniqueState) check(spec CutoverTableSpec, row []string, indexes map[string]int, csvLine int64) error {
	for column, seen := range state.textKeys {
		value := row[indexes[column]]
		if spec.Name == "owners" && value == "" {
			continue
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("table %s column %s row %d: duplicate value", spec.Name, column, csvLine)
		}
		seen[value] = struct{}{}
	}
	for column, seen := range state.idKeys {
		value := row[indexes[column]]
		if value == "" {
			continue
		}
		id, ok := cutoverReferenceID(value)
		if !ok {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("table %s column %s row %d: duplicate value", spec.Name, column, csvLine)
		}
		seen[id] = struct{}{}
	}
	if spec.Name == "merchandise_items" {
		if active, _ := cutoverPGBoolean(row[indexes["is_active"]]); active {
			name := row[indexes["name"]]
			if _, duplicate := state.activeNames[name]; duplicate {
				return fmt.Errorf("table %s column name row %d: duplicate name", spec.Name, csvLine)
			}
			state.activeNames[name] = struct{}{}
		}
	}
	if spec.Name == "appointments" && row[indexes["status"]] == "tentative" {
		if petID, ok := cutoverReferenceID(row[indexes["pet_id"]]); ok {
			if _, duplicate := state.tentativePets[petID]; duplicate {
				return fmt.Errorf("table %s column pet_id row %d: duplicate tentative appointment", spec.Name, csvLine)
			}
			state.tentativePets[petID] = struct{}{}
		}
	}
	return nil
}

var cutoverRecordPetPairTables = map[string]struct{}{
	"vital_records": {}, "exams": {}, "vaccinations": {},
}

func validateCutoverRecordPetPair(table string, row []string, indexes map[string]int, recordPets map[uint64]*uint64, csvLine int64) error {
	recordID, okRecord := cutoverReferenceID(row[indexes["medical_record_id"]])
	petID, okPet := cutoverReferenceID(row[indexes["pet_id"]])
	if !okRecord || !okPet {
		return nil
	}
	recordPet, exists := recordPets[recordID]
	if !exists || recordPet == nil || *recordPet != petID {
		return fmt.Errorf("table %s column pet_id row %d: pet does not match the referenced medical record", table, csvLine)
	}
	return nil
}
