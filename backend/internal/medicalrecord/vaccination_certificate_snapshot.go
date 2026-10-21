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
	"io"
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
	Doc              certJSONValue
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
	sv := doc.Obj["schema_version"]
	if sv.Kind != certJSONNumber || !jsonIntegerLiteral(sv.Num.String()) {
		return nil, errCertificateDependencyChanged
	}
	if svi, convErr := strconv.ParseInt(sv.Num.String(), 10, 64); convErr != nil || svi != 1 {
		return nil, errCertificateDependencyChanged
	}
	if !requireExactKeys(doc.Obj, certificateSnapshotTopKeys) {
		return nil, errCertificateResultUnavailable
	}
	kind := ""
	if kv := doc.Obj["kind"]; kv.Kind == certJSONString {
		kind = kv.Str
	}
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
	displayIssueDate := ""
	if dv := doc.Obj["display_issue_date"]; dv.Kind == certJSONString {
		displayIssueDate = dv.Str
	}
	if !isCertificateDate(displayIssueDate) {
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
	// Contract: only rabies_eligible == true permits a rabies edition, so a
	// stored false/unknown on a rabies snapshot is a contradiction. Report it
	// as result-unavailable and never consult the current master here.
	if kind == "rabies" {
		if eligible := eligibility.Obj["rabies_eligible"]; eligible.Kind != certJSONBool || !eligible.Bool {
			return nil, errCertificateResultUnavailable
		}
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
		Source:           certificateTypedSource(source.Obj),
		Owner:            certificateTypedOwner(owner.Obj),
		Pet:              certificateTypedPet(pet.Obj),
		Clinic:           certificateTypedClinic(clinic.Obj),
		Inputs:           certificateTypedInputs(inputs.Obj),
		DisplayIssueDate: displayIssueDate,
		RendererVersion:  rendererVersion,
		FontAssetVersion: fontAssetVersion,
		Templates:        templates,
	}, nil
}

// ─── Typed projection ────────────────────────────────────────────────────────
//
// The validators above have already proven every field's wire type, so the
// unchecked reads in these populate helpers cannot fail.

func certificateTypedSource(m map[string]certJSONValue) certificateSource {
	vaccinationID, _ := certInt64Field(m, "vaccination_id")
	version, _ := certInt64Field(m, "version")
	vaccineID, _ := certInt64Field(m, "vaccine_id")
	administeredOn := m["administered_on"].Str
	var lots [4]*string
	if raw := m["lots"]; raw.Kind == certJSONArray {
		for i := range lots {
			if i < len(raw.Arr) {
				lots[i] = certStringPtr(raw.Arr[i])
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

func certificateTypedOwner(m map[string]certJSONValue) certificateOwner {
	return certificateOwner{
		Name:       certStringPtr(m["name"]),
		NameKana:   certStringPtr(m["name_kana"]),
		PostalCode: certStringPtr(m["postal_code"]),
		Address:    certStringPtr(m["address"]),
		Phone:      certStringPtr(m["phone"]),
	}
}

func certificateTypedPet(m map[string]certJSONValue) certificatePet {
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

func certificateTypedClinic(m map[string]certJSONValue) certificateClinic {
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

func certificateTypedInputs(m map[string]certJSONValue) certificateInputs {
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

func certNestedObject(doc certJSONValue, key string, keys []string) (certJSONValue, bool) {
	v, ok := doc.Obj[key]
	if !ok || v.Kind != certJSONObject || !requireExactKeys(v.Obj, keys) {
		return certJSONValue{}, false
	}
	return v, true
}

func requireExactKeys(obj map[string]certJSONValue, keys []string) bool {
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

func validateCertificateSource(source certJSONValue) bool {
	// Stored IDs follow api.yaml CertificateId/CertificateSourceVersion:
	// integer/int64 with minimum 1.
	for _, k := range []string{"vaccination_id", "version", "vaccine_id"} {
		if !isPositiveJSONInt64(source.Obj[k]) {
			return false
		}
	}
	if !isNullOrString(source.Obj["vaccine_name"]) || !isNullOrString(source.Obj["doctor_name"]) {
		return false
	}
	administered := source.Obj["administered_on"]
	if administered.Kind != certJSONString || !isCertificateDate(administered.Str) {
		return false
	}
	if next := source.Obj["next_scheduled_on"]; next.Kind != certJSONNull {
		if next.Kind != certJSONString || !isCertificateDate(next.Str) {
			return false
		}
	}
	lots := source.Obj["lots"]
	if lots.Kind != certJSONArray || len(lots.Arr) != 4 {
		return false
	}
	for _, lot := range lots.Arr {
		if !isNullOrString(lot) {
			return false
		}
	}
	return true
}

func validateCertificatePet(pet certJSONValue) bool {
	if !validateStringOrNullFields(pet, certificatePetKeys) {
		return false
	}
	if birth := pet.Obj["birth_date"]; birth.Kind != certJSONNull {
		if birth.Kind != certJSONString || !isCertificateDate(birth.Str) {
			return false
		}
	}
	return true
}

func validateStringOrNullFields(obj certJSONValue, keys []string) bool {
	for _, k := range keys {
		if !isNullOrString(obj.Obj[k]) {
			return false
		}
	}
	return true
}

func validateCertificateInputs(inputs certJSONValue, kind string) bool {
	for _, k := range certificateInputStringKeys {
		if !isNullOrString(inputs.Obj[k]) {
			return false
		}
	}
	if bs := inputs.Obj["body_size"]; bs.Kind != certJSONNull {
		if bs.Kind != certJSONString || (bs.Str != "small" && bs.Str != "medium" && bs.Str != "large") {
			return false
		}
	}
	for _, k := range certificateInputIntKeys {
		if !isNullOrJSONInt64(inputs.Obj[k]) {
			return false
		}
	}
	if kind == "normal" {
		// Stored normal snapshots keep all 10 keys with every value null.
		for _, v := range inputs.Obj {
			if v.Kind != certJSONNull {
				return false
			}
		}
	}
	return true
}

func validateCertificateEligibility(eligibility certJSONValue) bool {
	if eligible := eligibility.Obj["rabies_eligible"]; eligible.Kind != certJSONNull {
		if eligible.Kind != certJSONBool {
			return false
		}
	}
	if !isPositiveJSONInt64(eligibility.Obj["vaccine_id"]) {
		return false
	}
	if !isNullOrPositiveJSONInt64(eligibility.Obj["vaccine_version"]) {
		return false
	}
	provenance := eligibility.Obj["provenance"]
	return provenance.Kind == certJSONString && (provenance.Str == "master_snapshot" || provenance.Str == "legacy_unresolved")
}

func validateCertificateManifest(manifest certJSONValue, kind string) (rendererVersion, fontAssetVersion string, templates map[string]string, ok bool) {
	rv := manifest.Obj["renderer_version"]
	fv := manifest.Obj["font_asset_version"]
	rawTemplates := manifest.Obj["templates"]
	if rv.Kind != certJSONString || fv.Kind != certJSONString || rawTemplates.Kind != certJSONObject || rv.Str == "" || fv.Str == "" {
		return "", "", nil, false
	}
	templates = make(map[string]string, len(rawTemplates.Obj))
	for k, v := range rawTemplates.Obj {
		if k != "portrait" && k != "landscape" && k != "rabies" {
			return "", "", nil, false
		}
		if v.Kind != certJSONString || v.Str == "" {
			return "", "", nil, false
		}
		templates[k] = v.Str
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
	return rv.Str, fv.Str, templates, true
}

// ─── Strict JSON plumbing ────────────────────────────────────────────────────

// strictJSONObject parses one complete JSON object into a certJSONValue,
// preserving integers as json.Number and rejecting duplicate keys and
// trailing data. The end of the input is verified by reading the next token:
// dec.More() cannot be used here because it also reports false on a stray
// trailing '}' or ']'.
func strictJSONObject(raw []byte) (certJSONValue, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	obj, err := decodeJSONObject(dec)
	if err != nil {
		return certJSONValue{}, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	return obj, nil
}

func decodeJSONObject(dec *json.Decoder) (certJSONValue, error) {
	tok, err := dec.Token()
	if err != nil {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	obj := certJSONValue{Kind: certJSONObject, Obj: make(map[string]certJSONValue)}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return certJSONValue{}, errCertificateInvalidSnapshot
		}
		key, ok := keyTok.(string)
		if !ok {
			return certJSONValue{}, errCertificateInvalidSnapshot
		}
		if _, dup := obj.Obj[key]; dup {
			return certJSONValue{}, errCertificateInvalidSnapshot
		}
		value, err := decodeJSONValue(dec)
		if err != nil {
			return certJSONValue{}, err
		}
		obj.Obj[key] = value
	}
	if _, err := dec.Token(); err != nil {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	return obj, nil
}

func decodeJSONArray(raw []byte) (certJSONValue, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	out := certJSONValue{Kind: certJSONArray}
	for dec.More() {
		v, err := decodeJSONValue(dec)
		if err != nil {
			return certJSONValue{}, err
		}
		out.Arr = append(out.Arr, v)
	}
	if _, err := dec.Token(); err != nil {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	if dec.More() {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	return out, nil
}

func decodeJSONValue(dec *json.Decoder) (certJSONValue, error) {
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return certJSONValue{}, errCertificateInvalidSnapshot
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

func decodeJSONScalar(raw []byte) (certJSONValue, error) {
	if !utf8.Valid(raw) {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	if dec.More() {
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
	switch t := tok.(type) {
	case nil:
		return certJSONValue{Kind: certJSONNull}, nil
	case bool:
		return certJSONValue{Kind: certJSONBool, Bool: t}, nil
	case json.Number:
		return certJSONValue{Kind: certJSONNumber, Num: t}, nil
	case string:
		// encoding/json silently substitutes U+FFFD for lone \uD800-\uDFFF
		// escapes; schema v1 rejects them on the raw literal instead.
		if err := rejectLoneSurrogates(raw); err != nil {
			return certJSONValue{}, err
		}
		return certJSONValue{Kind: certJSONString, Str: t}, nil
	default:
		return certJSONValue{}, errCertificateInvalidSnapshot
	}
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

func isNullOrString(v certJSONValue) bool {
	return v.Kind == certJSONNull || v.Kind == certJSONString
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

func isJSONInt64(v certJSONValue) bool {
	if v.Kind != certJSONNumber || !jsonIntegerLiteral(v.Num.String()) {
		return false
	}
	_, err := strconv.ParseInt(v.Num.String(), 10, 64)
	return err == nil
}

func isNullOrJSONInt64(v certJSONValue) bool {
	return v.Kind == certJSONNull || isJSONInt64(v)
}

// isPositiveJSONInt64 reports whether v is a JSON integer >= 1, matching the
// api.yaml minimum for stored IDs. Fee fields keep the signed isJSONInt64.
func isPositiveJSONInt64(v certJSONValue) bool {
	if v.Kind != certJSONNumber || !jsonIntegerLiteral(v.Num.String()) {
		return false
	}
	n, err := strconv.ParseInt(v.Num.String(), 10, 64)
	return err == nil && n >= 1
}

func isNullOrPositiveJSONInt64(v certJSONValue) bool {
	return v.Kind == certJSONNull || isPositiveJSONInt64(v)
}

// certInt64Field reads a required integer field as int64.
func certInt64Field(m map[string]certJSONValue, key string) (int64, bool) {
	n := m[key]
	if n.Kind != certJSONNumber || !jsonIntegerLiteral(n.Num.String()) {
		return 0, false
	}
	v, err := strconv.ParseInt(n.Num.String(), 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// certStringPtr reads a nullable string value into a *string (null -> nil).
func certStringPtr(v certJSONValue) *string {
	if v.Kind != certJSONString {
		return nil
	}
	s := v.Str
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
func certInt64Ptr(m map[string]certJSONValue, key string) *int64 {
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
