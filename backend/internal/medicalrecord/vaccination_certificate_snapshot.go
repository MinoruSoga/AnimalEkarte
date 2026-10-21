package medicalrecord

// vaccination_certificate_snapshot.go — EMR-486 schema-v1 certificate snapshot decoder.
//
// Stored snapshots are untrusted data: the reader must not guess, interpolate, or
// fill missing values from current master/owner/pet data. The decoder preserves
// JSON integers exactly (json.Decoder.UseNumber), rejects duplicate keys, unknown
// keys, missing required keys, lone surrogates and invalid UTF-8, and enforces the
// exact key sets fixed by docs/work/emr/emr-486-certificate-api-contract-20261020.md.
//
// Error mapping (kept in vaccination_certificate_handler.go):
//   - unparsable / wrong shape            -> 503 result-unavailable
//   - schema_version missing or != 1      -> 409 dependency-changed

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"time"
	"unicode/utf8"
)

// errCertificateInvalidSnapshot is the internal sentinel for a structurally
// invalid stored snapshot. decodeCertificateSnapshot maps it to
// errCertificateResultUnavailable before it reaches the HTTP boundary.
var errCertificateInvalidSnapshot = errors.New("invalid certificate snapshot")

// certificateSnapshot is the typed consumer view over a decoded v1 snapshot.
// Doc keeps the strict integer-preserving object graph: JCS hashing
// (snapshot_sha256 / render_manifest_sha256 / preview_fingerprint) requires
// the arbitrary JSON tree, and the DTO echoes the stored object verbatim.
// All other consumers read the typed fields below.
type certificateSnapshot struct {
	Doc              map[string]any
	Kind             string
	Source           certificateSource
	Owner            certificateOwner
	Pet              certificatePet
	Clinic           certificateClinic
	Inputs           certificateInputs
	DisplayIssueDate string
	RendererVersion  string
	FontAssetVersion string
	Templates        map[string]string
}

// Typed projection of the validated source/owner/pet/clinic/inputs objects.
// Nullable JSON fields are pointers; administered_on is a required date.
type certificateSource struct {
	VaccinationID   int64
	Version         int64
	VaccineID       int64
	VaccineName     *string
	AdministeredOn  string
	NextScheduledOn *string
	Lots            [4]*string
	DoctorName      *string
}

type certificateOwner struct {
	Name       *string
	NameKana   *string
	PostalCode *string
	Address    *string
	Phone      *string
}

type certificatePet struct {
	Name      *string
	Species   *string
	Breed     *string
	Sex       *string
	CoatColor *string
	AgeText   *string
	BirthDate *string
}

type certificateClinic struct {
	Name                    *string
	DirectorName            *string
	PostalCode              *string
	Address                 *string
	Phone                   *string
	Fax                     *string
	Email                   *string
	URL                     *string
	VeterinarianAssociation *string
	DesignatedVeterinarian  *string
	VeterinarianName        *string
}

type certificateInputs struct {
	CertificateNumber        *string
	InjectionTagNumber       *string
	DogRegistrationTagNumber *string
	DistinguishingFeatures   *string
	InjectionAmountText      *string
	BodySize                 *string
	RegistrationFee          *int64
	InjectionFee             *int64
	TagFee                   *int64
	Total                    *int64
}

var certificateSnapshotTopKeys = []string{
	"schema_version", "kind", "source", "owner", "pet", "clinic",
	"display_issue_date", "inputs", "eligibility", "render_manifest",
}

var certificateSourceKeys = []string{
	"vaccination_id", "version", "vaccine_id", "vaccine_name",
	"administered_on", "next_scheduled_on", "lots", "doctor_name",
}

var certificateOwnerKeys = []string{"name", "name_kana", "postal_code", "address", "phone"}

var certificatePetKeys = []string{"name", "species", "breed", "sex", "coat_color", "age_text", "birth_date"}

var certificateClinicKeys = []string{
	"name", "director_name", "postal_code", "address", "phone", "fax",
	"email", "url", "veterinarian_association", "designated_veterinarian", "veterinarian_name",
}

var certificateInputsKeys = []string{
	"certificate_number", "injection_tag_number", "dog_registration_tag_number",
	"distinguishing_features", "injection_amount_text", "body_size",
	"registration_fee", "injection_fee", "tag_fee", "total",
}

var certificateEligibilityKeys = []string{"rabies_eligible", "vaccine_id", "vaccine_version", "provenance"}

var certificateManifestKeys = []string{"renderer_version", "font_asset_version", "templates"}

var certificateInputStringKeys = []string{
	"certificate_number", "injection_tag_number", "dog_registration_tag_number",
	"distinguishing_features", "injection_amount_text",
}

var certificateInputIntKeys = []string{"registration_fee", "injection_fee", "tag_fee", "total"}

// decodeCertificateSnapshot strictly validates a stored snapshot blob.
func decodeCertificateSnapshot(raw []byte) (*certificateSnapshot, error) {
	doc, err := strictJSONObject(raw)
	if err != nil {
		return nil, errCertificateResultUnavailable
	}
	// Version gate first: a missing or non-1 schema_version is a
	// dependency-changed condition, not a shape failure. The snapshot bytes
	// must never be leaked in either error.
	sv, ok := doc["schema_version"].(json.Number)
	if !ok || !jsonIntegerLiteral(sv.String()) {
		return nil, errCertificateDependencyChanged
	}
	if svi, convErr := strconv.ParseInt(sv.String(), 10, 64); convErr != nil || svi != 1 {
		return nil, errCertificateDependencyChanged
	}
	if !requireExactKeys(doc, certificateSnapshotTopKeys) {
		return nil, errCertificateResultUnavailable
	}
	kind, _ := doc["kind"].(string)
	if kind != "normal" && kind != "rabies" {
		return nil, errCertificateResultUnavailable
	}
	source, ok := certNestedObject(doc, "source", certificateSourceKeys)
	if !ok || !validateCertificateSource(source) {
		return nil, errCertificateResultUnavailable
	}
	owner, ok := certNestedObject(doc, "owner", certificateOwnerKeys)
	if !ok || !validateStringOrNullFields(owner, certificateOwnerKeys) {
		return nil, errCertificateResultUnavailable
	}
	pet, ok := certNestedObject(doc, "pet", certificatePetKeys)
	if !ok || !validateCertificatePet(pet) {
		return nil, errCertificateResultUnavailable
	}
	clinic, ok := certNestedObject(doc, "clinic", certificateClinicKeys)
	if !ok || !validateStringOrNullFields(clinic, certificateClinicKeys) {
		return nil, errCertificateResultUnavailable
	}
	displayIssueDate, ok := doc["display_issue_date"].(string)
	if !ok || !isCertificateDate(displayIssueDate) {
		return nil, errCertificateResultUnavailable
	}
	inputs, ok := certNestedObject(doc, "inputs", certificateInputsKeys)
	if !ok || !validateCertificateInputs(inputs, kind) {
		return nil, errCertificateResultUnavailable
	}
	eligibility, ok := certNestedObject(doc, "eligibility", certificateEligibilityKeys)
	if !ok || !validateCertificateEligibility(eligibility) {
		return nil, errCertificateResultUnavailable
	}
	manifest, ok := certNestedObject(doc, "render_manifest", certificateManifestKeys)
	if !ok {
		return nil, errCertificateResultUnavailable
	}
	rendererVersion, fontAssetVersion, templates, ok := validateCertificateManifest(manifest, kind)
	if !ok {
		return nil, errCertificateResultUnavailable
	}
	return &certificateSnapshot{
		Doc:              doc,
		Kind:             kind,
		Source:           certificateTypedSource(source),
		Owner:            certificateTypedOwner(owner),
		Pet:              certificateTypedPet(pet),
		Clinic:           certificateTypedClinic(clinic),
		Inputs:           certificateTypedInputs(inputs),
		DisplayIssueDate: displayIssueDate,
		RendererVersion:  rendererVersion,
		FontAssetVersion: fontAssetVersion,
		Templates:        templates,
	}, nil
}

// ─── Typed projection ────────────────────────────────────────────────────────
//
// The validators above have already proven every field's wire type, so the
// unchecked coercions in these populate helpers cannot fail.

func certificateTypedSource(m map[string]any) certificateSource {
	vaccinationID, _ := certInt64Field(m, "vaccination_id")
	version, _ := certInt64Field(m, "version")
	vaccineID, _ := certInt64Field(m, "vaccine_id")
	administeredOn, _ := m["administered_on"].(string)
	var lots [4]*string
	if raw, ok := m["lots"].([]any); ok {
		for i := range lots {
			if i < len(raw) {
				lots[i] = certStringPtr(raw[i])
			}
		}
	}
	return certificateSource{
		VaccinationID:   vaccinationID,
		Version:         version,
		VaccineID:       vaccineID,
		VaccineName:     certStringPtr(m["vaccine_name"]),
		AdministeredOn:  administeredOn,
		NextScheduledOn: certStringPtr(m["next_scheduled_on"]),
		Lots:            lots,
		DoctorName:      certStringPtr(m["doctor_name"]),
	}
}

func certificateTypedOwner(m map[string]any) certificateOwner {
	return certificateOwner{
		Name:       certStringPtr(m["name"]),
		NameKana:   certStringPtr(m["name_kana"]),
		PostalCode: certStringPtr(m["postal_code"]),
		Address:    certStringPtr(m["address"]),
		Phone:      certStringPtr(m["phone"]),
	}
}

func certificateTypedPet(m map[string]any) certificatePet {
	return certificatePet{
		Name:      certStringPtr(m["name"]),
		Species:   certStringPtr(m["species"]),
		Breed:     certStringPtr(m["breed"]),
		Sex:       certStringPtr(m["sex"]),
		CoatColor: certStringPtr(m["coat_color"]),
		AgeText:   certStringPtr(m["age_text"]),
		BirthDate: certStringPtr(m["birth_date"]),
	}
}

func certificateTypedClinic(m map[string]any) certificateClinic {
	return certificateClinic{
		Name:                    certStringPtr(m["name"]),
		DirectorName:            certStringPtr(m["director_name"]),
		PostalCode:              certStringPtr(m["postal_code"]),
		Address:                 certStringPtr(m["address"]),
		Phone:                   certStringPtr(m["phone"]),
		Fax:                     certStringPtr(m["fax"]),
		Email:                   certStringPtr(m["email"]),
		URL:                     certStringPtr(m["url"]),
		VeterinarianAssociation: certStringPtr(m["veterinarian_association"]),
		DesignatedVeterinarian:  certStringPtr(m["designated_veterinarian"]),
		VeterinarianName:        certStringPtr(m["veterinarian_name"]),
	}
}

func certificateTypedInputs(m map[string]any) certificateInputs {
	return certificateInputs{
		CertificateNumber:        certStringPtr(m["certificate_number"]),
		InjectionTagNumber:       certStringPtr(m["injection_tag_number"]),
		DogRegistrationTagNumber: certStringPtr(m["dog_registration_tag_number"]),
		DistinguishingFeatures:   certStringPtr(m["distinguishing_features"]),
		InjectionAmountText:      certStringPtr(m["injection_amount_text"]),
		BodySize:                 certStringPtr(m["body_size"]),
		RegistrationFee:          certInt64Ptr(m, "registration_fee"),
		InjectionFee:             certInt64Ptr(m, "injection_fee"),
		TagFee:                   certInt64Ptr(m, "tag_fee"),
		Total:                    certInt64Ptr(m, "total"),
	}
}

func certNestedObject(doc map[string]any, key string, keys []string) (map[string]any, bool) {
	v, ok := doc[key].(map[string]any)
	if !ok || !requireExactKeys(v, keys) {
		return nil, false
	}
	return v, true
}

func requireExactKeys(obj map[string]any, keys []string) bool {
	if len(obj) != len(keys) {
		return false
	}
	for _, k := range keys {
		if _, ok := obj[k]; !ok {
			return false
		}
	}
	return true
}

func validateCertificateSource(source map[string]any) bool {
	for _, k := range []string{"vaccination_id", "version", "vaccine_id"} {
		if !isJSONInt64(source[k]) {
			return false
		}
	}
	if !isNullOrString(source["vaccine_name"]) || !isNullOrString(source["doctor_name"]) {
		return false
	}
	administered, ok := source["administered_on"].(string)
	if !ok || !isCertificateDate(administered) {
		return false
	}
	if next := source["next_scheduled_on"]; next != nil {
		s, ok := next.(string)
		if !ok || !isCertificateDate(s) {
			return false
		}
	}
	lots, ok := source["lots"].([]any)
	if !ok || len(lots) != 4 {
		return false
	}
	for _, lot := range lots {
		if !isNullOrString(lot) {
			return false
		}
	}
	return true
}

func validateCertificatePet(pet map[string]any) bool {
	if !validateStringOrNullFields(pet, certificatePetKeys) {
		return false
	}
	if birth := pet["birth_date"]; birth != nil {
		s, ok := birth.(string)
		if !ok || !isCertificateDate(s) {
			return false
		}
	}
	return true
}

func validateStringOrNullFields(obj map[string]any, keys []string) bool {
	for _, k := range keys {
		if !isNullOrString(obj[k]) {
			return false
		}
	}
	return true
}

func validateCertificateInputs(inputs map[string]any, kind string) bool {
	for _, k := range certificateInputStringKeys {
		if !isNullOrString(inputs[k]) {
			return false
		}
	}
	if bs := inputs["body_size"]; bs != nil {
		s, ok := bs.(string)
		if !ok || (s != "small" && s != "medium" && s != "large") {
			return false
		}
	}
	for _, k := range certificateInputIntKeys {
		if !isNullOrJSONInt64(inputs[k]) {
			return false
		}
	}
	if kind == "normal" {
		// Stored normal snapshots keep all 10 keys with every value null.
		for _, v := range inputs {
			if v != nil {
				return false
			}
		}
	}
	return true
}

func validateCertificateEligibility(eligibility map[string]any) bool {
	if eligible := eligibility["rabies_eligible"]; eligible != nil {
		if _, ok := eligible.(bool); !ok {
			return false
		}
	}
	if !isJSONInt64(eligibility["vaccine_id"]) {
		return false
	}
	if !isNullOrJSONInt64(eligibility["vaccine_version"]) {
		return false
	}
	provenance, ok := eligibility["provenance"].(string)
	return ok && (provenance == "master_snapshot" || provenance == "legacy_unresolved")
}

func validateCertificateManifest(manifest map[string]any, kind string) (rendererVersion, fontAssetVersion string, templates map[string]string, ok bool) {
	rendererVersion, ok1 := manifest["renderer_version"].(string)
	fontAssetVersion, ok2 := manifest["font_asset_version"].(string)
	rawTemplates, ok3 := manifest["templates"].(map[string]any)
	if !ok1 || !ok2 || !ok3 || rendererVersion == "" || fontAssetVersion == "" {
		return "", "", nil, false
	}
	templates = make(map[string]string, len(rawTemplates))
	for k, v := range rawTemplates {
		if k != "portrait" && k != "landscape" && k != "rabies" {
			return "", "", nil, false
		}
		s, isStr := v.(string)
		if !isStr || s == "" {
			return "", "", nil, false
		}
		templates[k] = s
	}
	if kind == "normal" {
		if _, hasPortrait := templates["portrait"]; !hasPortrait {
			return "", "", nil, false
		}
		if _, hasLandscape := templates["landscape"]; !hasLandscape {
			return "", "", nil, false
		}
	} else if _, hasRabies := templates["rabies"]; !hasRabies {
		return "", "", nil, false
	}
	return rendererVersion, fontAssetVersion, templates, true
}

// ─── Strict JSON plumbing ────────────────────────────────────────────────────

// strictJSONObject parses one complete JSON object, preserving integers as
// json.Number and rejecting duplicate keys and trailing data.
func strictJSONObject(raw []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	obj, err := decodeJSONObject(dec)
	if err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, errCertificateInvalidSnapshot
	}
	return obj, nil
}

func decodeJSONObject(dec *json.Decoder) (map[string]any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, errCertificateInvalidSnapshot
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, errCertificateInvalidSnapshot
	}
	obj := make(map[string]any)
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, errCertificateInvalidSnapshot
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, errCertificateInvalidSnapshot
		}
		if _, dup := obj[key]; dup {
			return nil, errCertificateInvalidSnapshot
		}
		value, err := decodeJSONValue(dec)
		if err != nil {
			return nil, err
		}
		obj[key] = value
	}
	if _, err := dec.Token(); err != nil {
		return nil, errCertificateInvalidSnapshot
	}
	return obj, nil
}

func decodeJSONArray(raw []byte) ([]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return nil, errCertificateInvalidSnapshot
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return nil, errCertificateInvalidSnapshot
	}
	var out []any
	for dec.More() {
		v, err := decodeJSONValue(dec)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if _, err := dec.Token(); err != nil {
		return nil, errCertificateInvalidSnapshot
	}
	if dec.More() {
		return nil, errCertificateInvalidSnapshot
	}
	return out, nil
}

func decodeJSONValue(dec *json.Decoder) (any, error) {
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, errCertificateInvalidSnapshot
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, errCertificateInvalidSnapshot
	}
	switch trimmed[0] {
	case '{':
		return strictJSONObject(trimmed)
	case '[':
		return decodeJSONArray(trimmed)
	default:
		return decodeJSONScalar(trimmed)
	}
}

func decodeJSONScalar(raw []byte) (any, error) {
	if !utf8.Valid(raw) {
		return nil, errCertificateInvalidSnapshot
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, errCertificateInvalidSnapshot
	}
	if dec.More() {
		return nil, errCertificateInvalidSnapshot
	}
	if _, ok := v.(string); ok {
		// encoding/json silently substitutes U+FFFD for lone \uD800-\uDFFF
		// escapes; schema v1 rejects them on the raw literal instead.
		if err := rejectLoneSurrogates(raw); err != nil {
			return nil, err
		}
	}
	return v, nil
}

// rejectLoneSurrogates scans a raw JSON string literal for unpaired \uXXXX
// surrogate escapes. Paired surrogate escapes (valid non-BMP text) pass.
func rejectLoneSurrogates(raw []byte) error {
	for i := 0; i+1 < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] != 'u' {
			i++
			continue
		}
		if i+6 > len(raw) {
			return errCertificateInvalidSnapshot
		}
		hi := certificateHex4(raw[i+2 : i+6])
		if hi < 0 {
			return errCertificateInvalidSnapshot
		}
		switch {
		case hi >= 0xD800 && hi <= 0xDBFF:
			if i+12 > len(raw) || raw[i+6] != '\\' || raw[i+7] != 'u' {
				return errCertificateInvalidSnapshot
			}
			lo := certificateHex4(raw[i+8 : i+12])
			if lo < 0xDC00 || lo > 0xDFFF {
				return errCertificateInvalidSnapshot
			}
			i += 11
		case hi >= 0xDC00 && hi <= 0xDFFF:
			return errCertificateInvalidSnapshot
		default:
			i += 5
		}
	}
	return nil
}

func certificateHex4(b []byte) int {
	if len(b) != 4 {
		return -1
	}
	n := 0
	for _, c := range b {
		var d int
		switch {
		case c >= '0' && c <= '9':
			d = int(c - '0')
		case c >= 'a' && c <= 'f':
			d = int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = int(c-'A') + 10
		default:
			return -1
		}
		n = n*16 + d
	}
	return n
}

// ─── Field predicates ────────────────────────────────────────────────────────

func isNullOrString(v any) bool {
	if v == nil {
		return true
	}
	_, ok := v.(string)
	return ok
}

// jsonIntegerLiteral reports whether a JSON number literal is an integer in
// canonical signed-decimal form (no fraction, exponent, plus sign, or leading
// zeros; "-0" is accepted here and normalized at projection time).
func jsonIntegerLiteral(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	if len(s) > 1 && s[0] == '0' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isJSONInt64(v any) bool {
	n, ok := v.(json.Number)
	if !ok || !jsonIntegerLiteral(n.String()) {
		return false
	}
	_, err := strconv.ParseInt(n.String(), 10, 64)
	return err == nil
}

func isNullOrJSONInt64(v any) bool {
	if v == nil {
		return true
	}
	return isJSONInt64(v)
}

// certInt64Field reads a required integer field as int64.
func certInt64Field(m map[string]any, key string) (int64, bool) {
	n, ok := m[key].(json.Number)
	if !ok || !jsonIntegerLiteral(n.String()) {
		return 0, false
	}
	v, err := strconv.ParseInt(n.String(), 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// certStringPtr reads a nullable string value into a *string (null -> nil).
func certStringPtr(v any) *string {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

// certText dereferences a nullable text field for rendering; nil -> "".
func certText(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// certInt64Ptr reads a nullable integer field.
func certInt64Ptr(m map[string]any, key string) *int64 {
	v, ok := certInt64Field(m, key)
	if !ok {
		return nil
	}
	return &v
}

// isCertificateDate enforces the exact YYYY-MM-DD wire form.
func isCertificateDate(s string) bool {
	if len(s) != len("2006-01-02") || s[4] != '-' || s[7] != '-' {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}
