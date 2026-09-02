package csvimport

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// PreflightCutoverBundle validates all source-side trust boundaries before any
// target database connection is opened. Errors contain table/column/row numbers
// only; source values are deliberately never included because CSVs contain PHI.
func PreflightCutoverBundle(sourceDir string, expected ExpectedCutoverSource) (CutoverBundle, error) {
	cleanDir, err := validateCutoverDirectory(sourceDir)
	if err != nil {
		return CutoverBundle{}, err
	}
	if err := validateExpectedCutoverSource(expected); err != nil {
		return CutoverBundle{}, err
	}

	manifestPath := filepath.Join(cleanDir, cutoverManifestName)
	manifestBytes, err := readOwnerOnlyRegularFile(manifestPath)
	if err != nil {
		return CutoverBundle{}, fmt.Errorf("manifest: %w", err)
	}
	if got := sha256Hex(manifestBytes); !strings.EqualFold(got, expected.ManifestSHA256) {
		return CutoverBundle{}, fmt.Errorf("manifest sha256 does not match the operator-supplied digest")
	}

	manifest, err := decodeCutoverManifest(manifestBytes)
	if err != nil {
		return CutoverBundle{}, err
	}
	if err := validateCutoverManifest(manifest, expected); err != nil {
		return CutoverBundle{}, err
	}
	if err := validateCutoverFiles(cleanDir, manifest, expected.Provenance); err != nil {
		return CutoverBundle{}, err
	}
	return CutoverBundle{SourceDir: cleanDir, Manifest: manifest, Provenance: expected.Provenance}, nil
}

func validateCutoverDirectory(sourceDir string) (string, error) {
	if sourceDir == "" || !filepath.IsAbs(sourceDir) {
		return "", fmt.Errorf("source directory must be an absolute path")
	}
	cleanDir := filepath.Clean(sourceDir)
	resolved, err := filepath.EvalSymlinks(cleanDir)
	if err != nil {
		return "", fmt.Errorf("resolve source directory: %w", err)
	}
	if resolved != cleanDir {
		return "", fmt.Errorf("source directory must not contain a symbolic link")
	}
	info, err := os.Lstat(cleanDir)
	if err != nil {
		return "", fmt.Errorf("inspect source directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("source path is not a directory")
	}
	if err := requireOwnerOnly(info, true); err != nil {
		return "", fmt.Errorf("source directory: %w", err)
	}
	return cleanDir, nil
}

func validateExpectedCutoverSource(expected ExpectedCutoverSource) error {
	if !validSHA256(expected.ManifestSHA256) {
		return fmt.Errorf("expected manifest sha256 must be exactly 64 hexadecimal characters")
	}
	if !clinicCodePattern.MatchString(expected.ClinicCode) {
		return fmt.Errorf("expected clinic code has an invalid format")
	}
	if expected.ClinicOrdinal < 1 || expected.ClinicOrdinal > 50 {
		return fmt.Errorf("expected clinic ordinal must be between 1 and 50")
	}
	if !runIDPattern.MatchString(expected.RunID) {
		return fmt.Errorf("expected run id has an invalid format")
	}
	return nil
}

func decodeCutoverManifest(contents []byte) (CutoverManifest, error) {
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	var manifest CutoverManifest
	if err := decoder.Decode(&manifest); err != nil {
		return CutoverManifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return CutoverManifest{}, err
	}
	return manifest, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("manifest contains multiple JSON values")
		}
		return fmt.Errorf("decode manifest trailer: %w", err)
	}
	return nil
}

func validateCutoverManifest(manifest CutoverManifest, expected ExpectedCutoverSource) error {
	if acceptsRehearsalManifest(expected.Provenance.Mode) {
		if manifest.Status != "PASS" && manifest.Status != "REHEARSAL_ONLY" {
			return fmt.Errorf("manifest status must be PASS or REHEARSAL_ONLY for rehearsal import")
		}
	} else if manifest.Status != "PASS" {
		return fmt.Errorf("manifest status must be PASS")
	}
	if err := validateCutoverProducerProvenance(manifest, expected.Provenance); err != nil {
		return err
	}
	if manifest.SourceLayer != "animalekarte_stage" {
		return fmt.Errorf("manifest source layer must be animalekarte_stage")
	}
	if manifest.Format != "csv-with-header" {
		return fmt.Errorf("manifest format must be csv-with-header")
	}
	if _, err := time.Parse(time.RFC3339Nano, manifest.GeneratedAt); err != nil {
		return fmt.Errorf("manifest generatedAt must be an RFC3339 timestamp")
	}
	if manifest.ClinicCode != expected.ClinicCode || manifest.ClinicOrdinal != expected.ClinicOrdinal || manifest.SourceRunID != expected.RunID {
		return fmt.Errorf("manifest clinic/run binding does not match expected source")
	}
	expectedBand := CutoverIDBand{
		Base:               (expected.ClinicOrdinal - 1) * clinicBandSize,
		EndExclusive:       expected.ClinicOrdinal * clinicBandSize,
		NonOwnerIDOffset:   (expected.ClinicOrdinal-1)*clinicBandSize + nonOwnerBandOffset,
		OwnerFloor:         (expected.ClinicOrdinal-1)*clinicBandSize + ownerBandOffset,
		ApplicationIDFloor: applicationIDFloor,
	}
	if manifest.ClinicBandBase != expectedBand.Base ||
		manifest.ClinicBandEndExclusive != expectedBand.EndExclusive ||
		manifest.StageIDOffset != expectedBand.NonOwnerIDOffset ||
		manifest.IDBand != expectedBand {
		return fmt.Errorf("manifest clinic band is inconsistent with clinic ordinal")
	}
	wantOutputDir := filepath.Join("sensitive-local", "animalekarte-csv-export", expected.ClinicCode, expected.RunID)
	outputDirMatches := manifest.OutputDir == wantOutputDir
	// Old DB local rehearsal exports append a controlled suffix (for example
	// "-rehearsal-current") while retaining the clinic/run binding. Formal and
	// staging imports remain exact-path only.
	if expected.Provenance.Mode == CutoverProvenanceLocalRehearsal {
		outputDirMatches = outputDirMatches || strings.HasPrefix(manifest.OutputDir, wantOutputDir+"-rehearsal-")
	}
	if !outputDirMatches || filepath.IsAbs(manifest.OutputDir) || filepath.Clean(manifest.OutputDir) != manifest.OutputDir {
		return fmt.Errorf("manifest output directory binding is invalid")
	}
	if manifest.ImportablePredicate != cutoverImportablePredicate {
		return fmt.Errorf("manifest importable predicate does not match the cutover contract")
	}
	if !reflect.DeepEqual(manifest.PlaceholderColumns, CutoverPlaceholderColumns()) {
		return fmt.Errorf("manifest placeholder inventory does not match the cutover contract")
	}

	specs := CutoverTableSpecs()
	if len(manifest.Tables) != len(specs) {
		return fmt.Errorf("manifest table order/count mismatch: got %d tables, want %d", len(manifest.Tables), len(specs))
	}
	for i, spec := range specs {
		table := manifest.Tables[i]
		if table.Table != spec.Name {
			return fmt.Errorf("manifest table order mismatch at position %d", i+1)
		}
		wantFile := spec.Name + ".csv"
		if table.File != wantFile || filepath.Base(table.File) != table.File || strings.ContainsAny(table.File, `/\\`) {
			return fmt.Errorf("manifest filename is unsafe or unexpected for table %s", spec.Name)
		}
		if table.RowCount < 0 {
			return fmt.Errorf("manifest row count must not be negative for table %s", spec.Name)
		}
		if !acceptsRehearsalManifest(expected.Provenance.Mode) &&
			(spec.Name == "payments" || spec.Name == "payment_splits") &&
			table.RowCount == 0 {
			return fmt.Errorf("manifest table %s must contain formal rows", spec.Name)
		}
		if !validSHA256(table.SHA256) {
			return fmt.Errorf("manifest sha256 is invalid for table %s", spec.Name)
		}
	}
	return nil
}

func acceptsRehearsalManifest(mode CutoverProvenanceMode) bool {
	return mode == CutoverProvenanceLocalRehearsal || mode == CutoverProvenanceStagingRehearsal
}

func validateCutoverProducerProvenance(manifest CutoverManifest, contract CutoverProvenanceContract) error {
	switch contract.Mode {
	case "", CutoverProvenanceFormal:
		// Formal provenance validation continues below.
	case CutoverProvenanceLocalRehearsal:
		return validateLocalRehearsalProducerProvenance(manifest)
	case CutoverProvenanceStagingRehearsal:
		if contract.Target.Environment != "staging" || contract.Target.Host == "" ||
			contract.Target.Database == "" || contract.Target.ClinicID <= 0 {
			return fmt.Errorf("staging rehearsal provenance requires an explicit staging target binding")
		}
		return validateStagingRehearsalProducerProvenance(manifest)
	default:
		return fmt.Errorf("unknown cutover provenance mode %q", contract.Mode)
	}
	if manifest.ManifestSchemaVersion != cutoverManifestSchema ||
		manifest.StageMappingSHA256 != cutoverStageMappingSHA256 ||
		manifest.CSVContractSHA256 != cutoverCSVContractSHA256 {
		return fmt.Errorf("manifest mapping contract binding is invalid")
	}
	if manifest.HandoffEligibility != "TRUSTED_CANDIDATE" {
		return fmt.Errorf("manifest handoff eligibility must be TRUSTED_CANDIDATE")
	}
	if manifest.SourceCompletenessStatus != "PASS" ||
		!manifest.SourceComplete ||
		!manifest.SourceProvenanceVerified {
		return fmt.Errorf("manifest source completeness must be fully verified")
	}
	if manifest.IncompleteSourceTables == nil || len(*manifest.IncompleteSourceTables) != 0 {
		return fmt.Errorf("manifest incomplete source table list must be empty")
	}
	if !stageBuildIDPattern.MatchString(manifest.StageBuildID) {
		return fmt.Errorf("manifest stage build ID is invalid")
	}
	if !validVerifiedSourceIdentity(manifest.SourceIdentity) {
		return fmt.Errorf("manifest source identity must be complete and verified")
	}
	if !validLayerDigests(manifest.SourceSummarySHA256) {
		return fmt.Errorf("manifest summary digest set is invalid")
	}
	if !validEvidenceDigests(manifest.SourceEvidenceSHA256) {
		return fmt.Errorf("manifest evidence digest set is invalid")
	}
	if !validOrderedLayerTimestamps(manifest.SourceSummaryGeneratedAt, manifest.GeneratedAt) {
		return fmt.Errorf("manifest summary generation timestamps are invalid or out of order")
	}
	return nil
}

func validateStagingRehearsalProducerProvenance(manifest CutoverManifest) error {
	// Shared staging is not a disposable local reset. It requires the same
	// complete producer evidence as a formal handoff, in addition to the exact
	// runtime target binding checked by validateCutoverProducerProvenance.
	if manifest.Status != "PASS" {
		return fmt.Errorf("staging rehearsal manifest status must be PASS")
	}
	if manifest.ManifestSchemaVersion != cutoverManifestSchema ||
		manifest.StageMappingSHA256 != cutoverStageMappingSHA256 ||
		manifest.CSVContractSHA256 != cutoverCSVContractSHA256 {
		return fmt.Errorf("staging rehearsal manifest mapping contract binding is invalid")
	}
	if manifest.HandoffEligibility != "TRUSTED_CANDIDATE" {
		return fmt.Errorf("staging rehearsal manifest handoff eligibility must be TRUSTED_CANDIDATE")
	}
	if manifest.SourceCompletenessStatus != "PASS" ||
		!manifest.SourceComplete ||
		!manifest.SourceProvenanceVerified {
		return fmt.Errorf("staging rehearsal manifest source completeness must be fully verified")
	}
	if manifest.IncompleteSourceTables == nil || len(*manifest.IncompleteSourceTables) != 0 {
		return fmt.Errorf("staging rehearsal manifest incomplete source table list must be empty")
	}
	if !stageBuildIDPattern.MatchString(manifest.StageBuildID) {
		return fmt.Errorf("staging rehearsal manifest stage build ID is invalid")
	}
	if !validVerifiedSourceIdentity(manifest.SourceIdentity) {
		return fmt.Errorf("staging rehearsal manifest source identity must be complete and verified")
	}
	if !validLayerDigests(manifest.SourceSummarySHA256) {
		return fmt.Errorf("staging rehearsal manifest summary digest set is invalid")
	}
	if !validEvidenceDigests(manifest.SourceEvidenceSHA256) {
		return fmt.Errorf("staging rehearsal manifest evidence digest set is invalid")
	}
	if !validOrderedLayerTimestamps(manifest.SourceSummaryGeneratedAt, manifest.GeneratedAt) {
		return fmt.Errorf("staging rehearsal manifest summary generation timestamps are invalid or out of order")
	}
	return nil
}

func validateLocalRehearsalProducerProvenance(manifest CutoverManifest) error {
	if manifest.ManifestSchemaVersion != cutoverManifestSchema {
		return fmt.Errorf("manifest schema version is invalid")
	}
	// CSV contract must still match; stage mapping may differ while old_db SQL
	// is ahead of the frozen consumer digest during local rehearsal.
	if manifest.CSVContractSHA256 != cutoverCSVContractSHA256 {
		return fmt.Errorf("manifest CSV contract digest is invalid")
	}
	if manifest.HandoffEligibility != "TRUSTED_CANDIDATE" &&
		manifest.HandoffEligibility != "REHEARSAL_ONLY" {
		return fmt.Errorf("manifest handoff eligibility must be TRUSTED_CANDIDATE or REHEARSAL_ONLY")
	}
	switch manifest.SourceCompletenessStatus {
	case "PASS", "PARTIAL", "UNVERIFIED":
	default:
		return fmt.Errorf("manifest source completeness status is unsupported for local rehearsal")
	}
	if !stageBuildIDPattern.MatchString(manifest.StageBuildID) {
		return fmt.Errorf("manifest stage build ID is invalid")
	}
	if !validSHA256(manifest.SourceSummarySHA256.Raw) ||
		!validSHA256(manifest.SourceSummarySHA256.Intermediate) ||
		!validSHA256(manifest.SourceSummarySHA256.Stage) {
		return fmt.Errorf("manifest summary digest set is invalid")
	}
	if !validOrderedLayerTimestamps(manifest.SourceSummaryGeneratedAt, manifest.GeneratedAt) {
		return fmt.Errorf("manifest summary generation timestamps are invalid or out of order")
	}
	return nil
}

func validVerifiedSourceIdentity(identity CutoverSourceIdentity) bool {
	return identity.Verified &&
		identity.SourceBackupSHA256 != nil &&
		validSHA256(*identity.SourceBackupSHA256) &&
		identity.SourceBackupSizeBytes != nil &&
		*identity.SourceBackupSizeBytes > 0 &&
		identity.BaseArchiveSHA256 != nil &&
		validSHA256(*identity.BaseArchiveSHA256) &&
		identity.KNJOArchiveSHA256 != nil &&
		validSHA256(*identity.KNJOArchiveSHA256)
}

func validLayerDigests(digests CutoverLayerDigests) bool {
	return validSHA256(digests.Raw) &&
		validSHA256(digests.Intermediate) &&
		validSHA256(digests.Stage)
}

func validEvidenceDigests(digests CutoverEvidenceDigests) bool {
	return validSHA256(digests.BaseLoad) && validSHA256(digests.KNJORecovery)
}

func validOrderedLayerTimestamps(timestamps CutoverLayerTimestamps, manifestGeneratedAt string) bool {
	raw, rawErr := time.Parse(time.RFC3339Nano, timestamps.Raw)
	intermediate, intermediateErr := time.Parse(time.RFC3339Nano, timestamps.Intermediate)
	stage, stageErr := time.Parse(time.RFC3339Nano, timestamps.Stage)
	manifest, manifestErr := time.Parse(time.RFC3339Nano, manifestGeneratedAt)
	return rawErr == nil &&
		intermediateErr == nil &&
		stageErr == nil &&
		manifestErr == nil &&
		!raw.After(intermediate) &&
		!intermediate.After(stage) &&
		!stage.After(manifest)
}

func validateCutoverFiles(sourceDir string, manifest CutoverManifest, provenance CutoverProvenanceContract) error {
	for i, spec := range CutoverTableSpecs() {
		path := filepath.Join(sourceDir, manifest.Tables[i].File)
		if err := validateCutoverCSV(path, spec, manifest.Tables[i], manifest.IDBand); err != nil {
			return err
		}
	}
	allowed := map[string]struct{}{cutoverManifestName: {}}
	for _, table := range manifest.Tables {
		allowed[table.File] = struct{}{}
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("list cutover source directory: %w", err)
	}
	for _, entry := range entries {
		if _, ok := allowed[entry.Name()]; !ok {
			return fmt.Errorf("unexpected file or directory in cutover source")
		}
	}
	return validateCutoverPaymentGraph(sourceDir, &manifest, provenance)
}

func validateCutoverCSV(path string, spec CutoverTableSpec, table CutoverManifestTable, band CutoverIDBand) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("table %s: inspect CSV: %w", spec.Name, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("table %s: CSV must not be a symbolic link", spec.Name)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("table %s: CSV must be a regular file", spec.Name)
	}
	if err := requireOwnerOnly(info, false); err != nil {
		return fmt.Errorf("table %s: %w", spec.Name, err)
	}

	f, err := openStableOwnerOnlyFile(path)
	if err != nil {
		return fmt.Errorf("table %s: open CSV: %w", spec.Name, err)
	}
	defer func() { _ = f.Close() }()
	openedInfo, statErr := f.Stat()
	if statErr != nil || openedInfo.Size() > maxCutoverCSVBytes {
		return fmt.Errorf("table %s: CSV exceeds the size limit", spec.Name)
	}
	hash := sha256.New()
	// Hash and validate the exact same opened bytes. Separate pathname opens
	// would allow a file swap between digest and structural validation.
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(f, maxCutoverCSVBytes+1), hash)))
	reader.FieldsPerRecord = len(spec.Columns)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("table %s: read CSV header: %w", spec.Name, err)
	}
	if !reflect.DeepEqual(header, spec.Columns) {
		return fmt.Errorf("table %s: CSV header does not match exact column order", spec.Name)
	}

	columnIndexes := make(map[string]int, len(spec.Columns))
	for i, column := range spec.Columns {
		columnIndexes[column] = i
	}
	var rowCount int64
	seenIDs := make(map[int64]struct{})
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return fmt.Errorf("table %s row %d: read CSV: %w", spec.Name, rowCount+2, readErr)
		}
		rowCount++
		idText := row[columnIndexes["id"]]
		id, parseErr := strconv.ParseInt(idText, 10, 64)
		if parseErr != nil {
			return fmt.Errorf("table %s column id row %d: primary ID must be an integer", spec.Name, rowCount+1)
		}
		if _, duplicate := seenIDs[id]; duplicate {
			return fmt.Errorf("table %s column id row %d: duplicate primary ID", spec.Name, rowCount+1)
		}
		seenIDs[id] = struct{}{}
		if err := validateCutoverRow(spec, row, columnIndexes, band, rowCount+1); err != nil {
			return err
		}
	}
	if rowCount != table.RowCount {
		return fmt.Errorf("table %s: row count mismatch: got %d, want %d", spec.Name, rowCount, table.RowCount)
	}
	actualDigest := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualDigest, table.SHA256) {
		return fmt.Errorf("table %s: CSV sha256 does not match manifest", spec.Name)
	}
	return nil
}

func validateCutoverRow(spec CutoverTableSpec, row []string, indexes map[string]int, band CutoverIDBand, csvLine int64) error {
	if err := validatePaymentMethodPlaceholder(spec, row, indexes, csvLine); err != nil {
		return err
	}
	if (spec.Name == "payments" || spec.Name == "payment_splits") && row[indexes["clinic_id"]] != "{{CLINIC_ID}}" {
		return fmt.Errorf("table %s column clinic_id row %d: clinic placeholder is required", spec.Name, csvLine)
	}
	for i, value := range row {
		matches := placeholderPattern.FindAllString(value, -1)
		for _, token := range matches {
			if !placeholderAllowed(spec.Name, spec.Columns[i], value, token) {
				return fmt.Errorf("table %s column %s row %d: unknown or misplaced placeholder", spec.Name, spec.Columns[i], csvLine)
			}
		}
		if (strings.Contains(value, "{{") || strings.Contains(value, "}}")) && len(matches) == 0 {
			return fmt.Errorf("table %s column %s row %d: malformed placeholder", spec.Name, spec.Columns[i], csvLine)
		}
	}
	for _, column := range spec.BandColumns {
		value := row[indexes[column]]
		if value == "" || placeholderPattern.MatchString(value) {
			continue
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("table %s column %s row %d: clinic band value must be an integer", spec.Name, column, csvLine)
		}
		floor := band.NonOwnerIDOffset
		if spec.Name == "owners" && column == "id" || column == "owner_id" {
			floor = band.OwnerFloor
		}
		if id < floor || id >= band.EndExclusive {
			return fmt.Errorf("table %s column %s row %d: id is outside the clinic band", spec.Name, column, csvLine)
		}
	}
	return nil
}

func placeholderAllowed(table, column, value, token string) bool {
	if value != token {
		return false
	}
	key := table + "." + column
	if table == "pets" && column == "animal_species_id" {
		key = "pets.animal_species_id (fallback only, when unresolved)"
	}
	if (table == "payments" || table == "payment_splits") && column == "payment_method_id" {
		return token == "{{PAYMENT_METHOD_CASH_ID}}" || token == "{{PAYMENT_METHOD_CREDIT_CARD_ID}}"
	}
	want, ok := CutoverPlaceholderColumns()[key]
	return ok && token == want
}

func validatePaymentMethodPlaceholder(spec CutoverTableSpec, row []string, indexes map[string]int, csvLine int64) error {
	if spec.Name != "payments" && spec.Name != "payment_splits" {
		return nil
	}
	method := row[indexes["method"]]
	token := row[indexes["payment_method_id"]]
	want := map[string]string{
		"cash":        "{{PAYMENT_METHOD_CASH_ID}}",
		"credit_card": "{{PAYMENT_METHOD_CREDIT_CARD_ID}}",
	}[method]
	if want == "" {
		return fmt.Errorf("table %s column method row %d: unsupported payment method", spec.Name, csvLine)
	}
	if token != want {
		return fmt.Errorf("table %s column payment_method_id row %d: payment method placeholder does not match method", spec.Name, csvLine)
	}
	return nil
}

func readOwnerOnlyRegularFile(path string) ([]byte, error) {
	file, err := openStableOwnerOnlyFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect file size: %w", err)
	}
	if info.Size() > maxCutoverManifestBytes {
		return nil, fmt.Errorf("file exceeds the manifest size limit")
	}
	return io.ReadAll(io.LimitReader(file, maxCutoverManifestBytes+1))
}

func requireOwnerOnly(info os.FileInfo, directory bool) error {
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("path must be owner-only")
	}
	if directory && info.Mode().Perm()&0o500 != 0o500 {
		return fmt.Errorf("owner-only directory must be readable and searchable")
	}
	return nil
}

func sha256Hex(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	if value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
