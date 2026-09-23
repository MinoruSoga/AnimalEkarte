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
	"slices"
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
	manifestNotes, err := validateCutoverManifest(&manifest, expected)
	if err != nil {
		return CutoverBundle{}, err
	}
	if err := bindCutoverAccountSource(cleanDir, expected.AccountSourceDir, &manifest); err != nil {
		return CutoverBundle{}, err
	}
	fileNotes, err := validateCutoverFiles(cleanDir, manifest, expected.Provenance)
	if err != nil {
		return CutoverBundle{}, err
	}
	return CutoverBundle{
		SourceDir:      cleanDir,
		Manifest:       manifest,
		Provenance:     expected.Provenance,
		ToleratedDrift: append(manifestNotes, fileNotes...),
	}, nil
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

func validateCutoverManifest(manifest *CutoverManifest, expected ExpectedCutoverSource) ([]string, error) {
	mode := expected.Provenance.Mode
	if acceptsRehearsalManifest(mode) {
		if manifest.Status != "PASS" && manifest.Status != "REHEARSAL_ONLY" {
			return nil, fmt.Errorf("manifest status must be PASS or REHEARSAL_ONLY for rehearsal import")
		}
	} else if manifest.Status != "PASS" {
		return nil, fmt.Errorf("manifest status must be PASS")
	}
	notes, err := validateCutoverProducerProvenance(*manifest, expected.Provenance)
	if err != nil {
		return nil, err
	}
	if err := validateWindowZeroEvidence(*manifest, expected.Provenance); err != nil {
		return nil, err
	}
	if manifest.SourceLayer != "animalekarte_stage" {
		return nil, fmt.Errorf("manifest source layer must be animalekarte_stage")
	}
	if manifest.Format != "csv-with-header" {
		return nil, fmt.Errorf("manifest format must be csv-with-header")
	}
	if _, err := time.Parse(time.RFC3339Nano, manifest.GeneratedAt); err != nil {
		return nil, fmt.Errorf("manifest generatedAt must be an RFC3339 timestamp")
	}
	if manifest.ClinicCode != expected.ClinicCode || manifest.ClinicOrdinal != expected.ClinicOrdinal || manifest.SourceRunID != expected.RunID {
		return nil, fmt.Errorf("manifest clinic/run binding does not match expected source")
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
		return nil, fmt.Errorf("manifest clinic band is inconsistent with clinic ordinal")
	}
	wantOutputDir := filepath.Join("sensitive-local", "animalekarte-csv-export", expected.ClinicCode, expected.RunID)
	outputDirMatches := manifest.OutputDir == wantOutputDir
	// Reviewed rebuilds retain the source run and publish a new immutable revision.
	revisionPrefix := wantOutputDir + "-revisions/"
	if strings.HasPrefix(manifest.OutputDir, revisionPrefix) {
		outputDirMatches = runIDPattern.MatchString(strings.TrimPrefix(manifest.OutputDir, revisionPrefix))
	}
	// Old DB rehearsal exports append a controlled suffix (for example
	// "-rehearsal-current") while retaining the clinic/run binding. Formal
	// cutover and verified staging PASS bundles allow only the bound revision above.
	if expected.Provenance.Mode == CutoverProvenanceLocalRehearsal ||
		(expected.Provenance.Mode == CutoverProvenanceStagingRehearsal && isRehearsalOnlyProducer(*manifest)) {
		outputDirMatches = outputDirMatches || strings.HasPrefix(manifest.OutputDir, wantOutputDir+"-rehearsal-")
	}
	if !outputDirMatches || filepath.IsAbs(manifest.OutputDir) || filepath.Clean(manifest.OutputDir) != manifest.OutputDir {
		return nil, fmt.Errorf("manifest output directory binding is invalid")
	}
	// Drift tolerance is artifact-keyed: only REHEARSAL_ONLY handoffs admitted
	// under a rehearsal provenance mode may diverge from the frozen contract
	// surface, and every accepted divergence is recorded on the bundle.
	drift := rehearsalContractDrift(*manifest, mode)
	if manifest.ImportablePredicate != cutoverImportablePredicate {
		if !drift {
			return nil, fmt.Errorf("manifest importable predicate does not match the cutover contract")
		}
		notes = append(notes, "manifest.importablePredicate")
	}
	if !reflect.DeepEqual(manifest.PlaceholderColumns, CutoverPlaceholderColumns()) {
		if !drift {
			return nil, fmt.Errorf("manifest placeholder inventory does not match the cutover contract")
		}
		notes = append(notes, "manifest.placeholderColumns")
	}

	specs := CutoverTableSpecs()
	if err := normalizeCutoverManifestTables(manifest, specs, drift, &notes); err != nil {
		return nil, err
	}
	for i, spec := range specs {
		table := manifest.Tables[i]
		wantFile := spec.Name + ".csv"
		if table.File != wantFile || filepath.Base(table.File) != table.File || strings.ContainsAny(table.File, `/\\`) {
			return nil, fmt.Errorf("manifest filename is unsafe or unexpected for table %s", spec.Name)
		}
		if table.RowCount < 0 {
			return nil, fmt.Errorf("manifest row count must not be negative for table %s", spec.Name)
		}
		if !acceptsRehearsalManifest(mode) &&
			(spec.Name == "payments" || spec.Name == "payment_splits") &&
			table.RowCount == 0 {
			return nil, fmt.Errorf("manifest table %s must contain formal rows", spec.Name)
		}
		if !validSHA256(table.SHA256) {
			return nil, fmt.Errorf("manifest sha256 is invalid for table %s", spec.Name)
		}
	}
	return notes, nil
}

// normalizeCutoverManifestTables re-orders manifest tables into the immutable
// parent-before-child CutoverTableSpecs() order so every downstream consumer
// (files, reference graph, apply, resume, verify) iterates spec order even
// when a rehearsal producer emitted a different order. A reordered tables[]
// is tolerated only for rehearsal-grade bundles and recorded as drift;
// formal bundles must already match the contract order exactly. Unknown,
// duplicate, missing, or extra table names always fail closed.
func normalizeCutoverManifestTables(manifest *CutoverManifest, specs []CutoverTableSpec, drift bool, notes *[]string) error {
	if len(manifest.Tables) != len(specs) {
		return fmt.Errorf("manifest table order/count mismatch: got %d tables, want %d", len(manifest.Tables), len(specs))
	}
	byName := make(map[string]int, len(manifest.Tables))
	for i, table := range manifest.Tables {
		if _, duplicate := byName[table.Table]; duplicate {
			return fmt.Errorf("manifest lists table %s more than once", table.Table)
		}
		byName[table.Table] = i
	}
	specNames := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		specNames[spec.Name] = struct{}{}
	}
	var unknown []string
	for name := range byName {
		if _, ok := specNames[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	slices.Sort(unknown)
	if len(unknown) > 0 {
		return fmt.Errorf("manifest contains unknown table %s", unknown[0])
	}
	normalized := make([]CutoverManifestTable, len(specs))
	for i, spec := range specs {
		j, found := byName[spec.Name]
		if !found {
			return fmt.Errorf("manifest is missing table %s", spec.Name)
		}
		normalized[i] = manifest.Tables[j]
	}
	mismatch := -1
	for i, table := range manifest.Tables {
		if table.Table != specs[i].Name {
			mismatch = i
			break
		}
	}
	if mismatch < 0 {
		return nil
	}
	if !drift {
		return fmt.Errorf("manifest table order mismatch at position %d", mismatch+1)
	}
	*notes = append(*notes, "manifest.tables order")
	manifest.Tables = normalized
	return nil
}

func acceptsRehearsalManifest(mode CutoverProvenanceMode) bool {
	return mode == CutoverProvenanceLocalRehearsal || mode == CutoverProvenanceStagingRehearsal
}

func validateCutoverProducerProvenance(manifest CutoverManifest, contract CutoverProvenanceContract) ([]string, error) {
	switch contract.Mode {
	case "", CutoverProvenanceFormal:
		// Formal provenance validation continues below.
	case CutoverProvenanceLocalRehearsal:
		return validateLocalRehearsalProducerProvenance(manifest)
	case CutoverProvenanceStagingRehearsal:
		if contract.Target.Environment != "staging" || contract.Target.Host == "" ||
			contract.Target.Database == "" || contract.Target.ClinicID <= 0 {
			return nil, fmt.Errorf("staging rehearsal provenance requires an explicit staging target binding")
		}
		// Shared STG can load _old_db_handoff rehearsal bundles after the
		// cmd/csv-import-stg-uat env sentinel and target binding. Formal
		// csv-import still requires TRUSTED_CANDIDATE.
		if isRehearsalOnlyProducer(manifest) {
			return validateLocalRehearsalProducerProvenance(manifest)
		}
		return nil, validateStagingRehearsalProducerProvenance(manifest)
	default:
		return nil, fmt.Errorf("unknown cutover provenance mode %q", contract.Mode)
	}
	if manifest.ManifestSchemaVersion != cutoverManifestSchema ||
		manifest.StageMappingSHA256 != cutoverStageMappingSHA256 ||
		manifest.CSVContractSHA256 != cutoverCSVContractSHA256 {
		return nil, fmt.Errorf("manifest mapping contract binding is invalid")
	}
	if manifest.HandoffEligibility != "TRUSTED_CANDIDATE" {
		return nil, fmt.Errorf("manifest handoff eligibility must be TRUSTED_CANDIDATE")
	}
	if manifest.SourceCompletenessStatus != "PASS" ||
		!manifest.SourceComplete ||
		!manifest.SourceProvenanceVerified {
		return nil, fmt.Errorf("manifest source completeness must be fully verified")
	}
	if manifest.IncompleteSourceTables == nil || len(*manifest.IncompleteSourceTables) != 0 {
		return nil, fmt.Errorf("manifest incomplete source table list must be empty")
	}
	if !stageBuildIDPattern.MatchString(manifest.StageBuildID) {
		return nil, fmt.Errorf("manifest stage build ID is invalid")
	}
	if !validVerifiedSourceIdentity(manifest.SourceIdentity) {
		return nil, fmt.Errorf("manifest source identity must be complete and verified")
	}
	if !validLayerDigests(manifest.SourceSummarySHA256) {
		return nil, fmt.Errorf("manifest summary digest set is invalid")
	}
	if !validEvidenceDigests(manifest.SourceEvidenceSHA256, manifest.SourceIdentity) {
		return nil, fmt.Errorf("manifest evidence digest set is invalid")
	}
	if !validOrderedLayerTimestamps(manifest.SourceSummaryGeneratedAt, manifest.GeneratedAt) {
		return nil, fmt.Errorf("manifest summary generation timestamps are invalid or out of order")
	}
	return nil, nil
}

func validateStagingRehearsalProducerProvenance(manifest CutoverManifest) error {
	// Verified PASS / TRUSTED_CANDIDATE staging imports keep the same producer
	// evidence as formal cutover, plus the runtime target binding checked by
	// validateCutoverProducerProvenance. REHEARSAL_ONLY handoff uses
	// validateLocalRehearsalProducerProvenance instead.
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
	if !validEvidenceDigests(manifest.SourceEvidenceSHA256, manifest.SourceIdentity) {
		return fmt.Errorf("staging rehearsal manifest evidence digest set is invalid")
	}
	if !validOrderedLayerTimestamps(manifest.SourceSummaryGeneratedAt, manifest.GeneratedAt) {
		return fmt.Errorf("staging rehearsal manifest summary generation timestamps are invalid or out of order")
	}
	return nil
}

func validateLocalRehearsalProducerProvenance(manifest CutoverManifest) ([]string, error) {
	if manifest.ManifestSchemaVersion != cutoverManifestSchema {
		return nil, fmt.Errorf("manifest schema version is invalid")
	}
	// Stage mapping may differ while old_db SQL is ahead of the frozen
	// consumer digest during local rehearsal. A drifted CSV contract digest is
	// tolerated only for rehearsal-grade artifacts: the file-level gates
	// (header set equality, per-file sha256, row counts, reference graph)
	// stay authoritative, so the drift is recorded on the bundle instead of
	// being hidden. TRUSTED_CANDIDATE/PASS artifacts keep the strict check.
	var notes []string
	if manifest.CSVContractSHA256 != cutoverCSVContractSHA256 {
		if !isRehearsalOnlyProducer(manifest) {
			return nil, fmt.Errorf("manifest CSV contract digest is invalid")
		}
		notes = append(notes, "manifest.csvContractSha256")
	}
	if manifest.HandoffEligibility != "TRUSTED_CANDIDATE" &&
		manifest.HandoffEligibility != "REHEARSAL_ONLY" {
		return nil, fmt.Errorf("manifest handoff eligibility must be TRUSTED_CANDIDATE or REHEARSAL_ONLY")
	}
	switch manifest.SourceCompletenessStatus {
	case "PASS", "PARTIAL", "UNVERIFIED":
	default:
		return nil, fmt.Errorf("manifest source completeness status is unsupported for local rehearsal")
	}
	if !stageBuildIDPattern.MatchString(manifest.StageBuildID) {
		return nil, fmt.Errorf("manifest stage build ID is invalid")
	}
	if !validSHA256(manifest.SourceSummarySHA256.Raw) ||
		!validSHA256(manifest.SourceSummarySHA256.Intermediate) ||
		!validSHA256(manifest.SourceSummarySHA256.Stage) {
		return nil, fmt.Errorf("manifest summary digest set is invalid")
	}
	if !validOrderedLayerTimestamps(manifest.SourceSummaryGeneratedAt, manifest.GeneratedAt) {
		return nil, fmt.Errorf("manifest summary generation timestamps are invalid or out of order")
	}
	return notes, nil
}

func validVerifiedSourceIdentity(identity CutoverSourceIdentity) bool {
	baseValid := identity.Verified &&
		identity.SourceBackupSHA256 != nil &&
		validSHA256(*identity.SourceBackupSHA256) &&
		identity.SourceBackupSizeBytes != nil &&
		*identity.SourceBackupSizeBytes > 0 &&
		identity.BaseArchiveSHA256 != nil &&
		validSHA256(*identity.BaseArchiveSHA256)
	if !baseValid || identity.KnjoProvenanceRoute == nil {
		return false
	}
	switch *identity.KnjoProvenanceRoute {
	case "complete_base":
		return identity.KNJOArchiveSHA256 == nil
	case "reacquire":
		return identity.KNJOArchiveSHA256 != nil && validSHA256(*identity.KNJOArchiveSHA256)
	default:
		return false
	}
}

func validLayerDigests(digests CutoverLayerDigests) bool {
	return validSHA256(digests.Raw) &&
		validSHA256(digests.Intermediate) &&
		validSHA256(digests.Stage)
}

func validEvidenceDigests(digests CutoverEvidenceDigests, identity CutoverSourceIdentity) bool {
	if !validSHA256(digests.BaseLoad) || identity.KnjoProvenanceRoute == nil {
		return false
	}
	switch *identity.KnjoProvenanceRoute {
	case "complete_base":
		return digests.KNJORecovery == ""
	case "reacquire":
		return validSHA256(digests.KNJORecovery)
	default:
		return false
	}
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

func validateCutoverFiles(sourceDir string, manifest CutoverManifest, provenance CutoverProvenanceContract) ([]string, error) {
	// Header permutation tolerance is keyed on the same artifact+mode pair as
	// manifest drift so preflight and apply observe one policy.
	allowPermutation := rehearsalContractDrift(manifest, provenance.Mode)
	var notes []string
	for _, spec := range CutoverTableSpecs() {
		table, err := cutoverManifestTableByName(manifest, spec.Name)
		if err != nil {
			return nil, err
		}
		permuted, err := validateCutoverCSV(cutoverCSVPath(sourceDir, table), spec, table, manifest.IDBand, allowPermutation)
		if err != nil {
			return nil, err
		}
		if permuted {
			notes = append(notes, "csv header order: "+spec.Name)
		}
	}
	allowed := map[string]struct{}{cutoverManifestName: {}}
	if _, err := os.Lstat(filepath.Join(sourceDir, "accounts")); !os.IsNotExist(err) {
		allowed["accounts"] = struct{}{}
	}
	for _, table := range manifest.Tables {
		allowed[table.File] = struct{}{}
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("list cutover source directory: %w", err)
	}
	for _, entry := range entries {
		if _, ok := allowed[entry.Name()]; !ok {
			return nil, fmt.Errorf("unexpected file or directory in cutover source")
		}
	}
	if err := validateCutoverPaymentGraph(sourceDir, &manifest, provenance); err != nil {
		return nil, err
	}
	if err := validateCutoverReferenceGraph(sourceDir, manifest, allowPermutation); err != nil {
		return nil, err
	}
	return notes, nil
}

func validateCutoverCSV(path string, spec CutoverTableSpec, table CutoverManifestTable, band CutoverIDBand, allowPermutation bool) (bool, error) {
	path, err := resolveCutoverAccountCSV(path)
	if err != nil {
		return false, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("table %s: inspect CSV: %w", spec.Name, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("table %s: CSV must not be a symbolic link", spec.Name)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("table %s: CSV must be a regular file", spec.Name)
	}
	if err := requireOwnerOnly(info, false); err != nil {
		return false, fmt.Errorf("table %s: %w", spec.Name, err)
	}

	f, err := openStableOwnerOnlyFile(path)
	if err != nil {
		return false, fmt.Errorf("table %s: open CSV: %w", spec.Name, err)
	}
	defer func() { _ = f.Close() }()
	openedInfo, statErr := f.Stat()
	if statErr != nil || openedInfo.Size() > maxCutoverCSVBytes {
		return false, fmt.Errorf("table %s: CSV exceeds the size limit", spec.Name)
	}
	hash := sha256.New()
	// Hash and validate the exact same opened bytes. Separate pathname opens
	// would allow a file swap between digest and structural validation.
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(f, maxCutoverCSVBytes+1), hash)))
	// The header declares row arity (FieldsPerRecord 0): every data row must
	// match it exactly, and cutoverColumnIndexes resolves fields by name.
	header, err := reader.Read()
	if err != nil {
		return false, fmt.Errorf("table %s: read CSV header: %w", spec.Name, err)
	}
	columnIndexes, permuted, err := cutoverColumnIndexes(spec, header, allowPermutation)
	if err != nil {
		return false, err
	}

	var rowCount int64
	seenIDs := make(map[int64]struct{})
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return false, fmt.Errorf("table %s row %d: read CSV: %w", spec.Name, rowCount+2, readErr)
		}
		rowCount++
		idText := row[columnIndexes["id"]]
		id, parseErr := strconv.ParseInt(idText, 10, 64)
		if parseErr != nil {
			return false, fmt.Errorf("table %s column id row %d: primary ID must be an integer", spec.Name, rowCount+1)
		}
		if _, duplicate := seenIDs[id]; duplicate {
			return false, fmt.Errorf("table %s column id row %d: duplicate primary ID", spec.Name, rowCount+1)
		}
		seenIDs[id] = struct{}{}
		if err := validateCutoverRow(spec, header, row, columnIndexes, band, rowCount+1); err != nil {
			return false, err
		}
	}
	if rowCount != table.RowCount {
		return false, fmt.Errorf("table %s: row count mismatch: got %d, want %d", spec.Name, rowCount, table.RowCount)
	}
	actualDigest := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualDigest, table.SHA256) {
		return false, fmt.Errorf("table %s: CSV sha256 does not match manifest", spec.Name)
	}
	return permuted, nil
}

// cutoverColumnIndexes binds CSV header names to field positions so no
// consumer ever indexes rows positionally. Formal bundles keep exact-order
// enforcement; rehearsal-grade bundles may permute the same column set.
// Duplicate names, unknown columns, and missing contract columns each fail
// closed with the offending column named — never a cell value.
func cutoverColumnIndexes(spec CutoverTableSpec, header []string, allowPermutation bool) (map[string]int, bool, error) {
	exact := reflect.DeepEqual(header, spec.Columns)
	if !exact && !allowPermutation {
		return nil, false, fmt.Errorf("table %s: CSV header does not match exact column order", spec.Name)
	}
	indexes := make(map[string]int, len(header))
	for i, column := range header {
		if _, duplicate := indexes[column]; duplicate {
			return nil, false, fmt.Errorf("table %s: CSV header contains duplicate column %s", spec.Name, column)
		}
		indexes[column] = i
	}
	for _, column := range header {
		if !hasColumn(spec.Columns, column) {
			return nil, false, fmt.Errorf("table %s: CSV header contains unknown column %s", spec.Name, column)
		}
	}
	for _, column := range spec.Columns {
		if _, ok := indexes[column]; !ok {
			return nil, false, fmt.Errorf("table %s: CSV header is missing column %s", spec.Name, column)
		}
	}
	return indexes, !exact, nil
}

func validateCutoverRow(spec CutoverTableSpec, header []string, row []string, indexes map[string]int, band CutoverIDBand, csvLine int64) error {
	if err := validatePaymentMethodPlaceholder(spec, row, indexes, csvLine); err != nil {
		return err
	}
	if (spec.Name == "payments" || spec.Name == "payment_splits") && row[indexes["clinic_id"]] != "{{CLINIC_ID}}" {
		return fmt.Errorf("table %s column clinic_id row %d: clinic placeholder is required", spec.Name, csvLine)
	}
	// header[i] is the actual CSV column name for row[i]; the CSV reader pins
	// row arity to the header length so the two are always aligned.
	for i, value := range row {
		matches := placeholderPattern.FindAllString(value, -1)
		for _, token := range matches {
			if !placeholderAllowed(spec.Name, header[i], value, token) {
				return fmt.Errorf("table %s column %s row %d: unknown or misplaced placeholder", spec.Name, header[i], csvLine)
			}
		}
		if (strings.Contains(value, "{{") || strings.Contains(value, "}}")) && len(matches) == 0 {
			return fmt.Errorf("table %s column %s row %d: malformed placeholder", spec.Name, header[i], csvLine)
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
