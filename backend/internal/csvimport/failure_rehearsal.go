package csvimport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	SyntheticFailureFixtureID      = "F8_G4_TRANSACTION_ROLLBACK_V1"
	SyntheticFailureCheckpoint     = "G4_TARGET_VERIFIED"
	SyntheticFailureStage          = "TRANSACTION"
	SyntheticFailureMarker         = "SYNTHETIC_FK_VIOLATION_AFTER_COPY"
	syntheticFailureSQLState       = "23503"
	syntheticFailureOwnerID        = int64(300_000)
	syntheticFailurePetID          = int64(1_000_000)
	syntheticFailureMissingOwnerID = int64(9_999_999)
	failureRollbackTimeout         = 15 * time.Second
)

var commitPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)

// SyntheticFailureInput binds one fixed synthetic rollback rehearsal to the
// exact release and disposable database identity attested by the host runner.
type SyntheticFailureInput struct {
	ClinicCode                   string
	ClinicOrdinal                int64
	RunID                        string
	TargetReleaseCommit          string
	TargetDatabaseIdentitySHA256 string
	FixtureManifestSHA256        string
	FailureRuntimeReportSHA256   string
	Seeds                        CutoverSeedIDs
}

// BandCount is one aggregate count in the immutable F6 table order.
type BandCount struct {
	Table    string `json:"table"`
	RowCount int64  `json:"rowCount"`
}

// BandCountsEvidence is aggregate-only evidence. It contains no row values,
// identifiers, credentials, paths, timestamps, or free text.
type BandCountsEvidence struct {
	SchemaVersion int         `json:"schemaVersion"`
	TableCount    int         `json:"tableCount"`
	Tables        []BandCount `json:"tables"`
	TotalRowCount int64       `json:"totalRowCount"`
}

// CanonicalJSON returns the exact newline-terminated bytes whose SHA-256 is
// embedded in the failure apply report.
func (e BandCountsEvidence) CanonicalJSON() ([]byte, error) {
	return canonicalSyntheticEvidence(e)
}

// TransactionEvidence records only fixed aggregate state transitions.
type TransactionEvidence struct {
	SchemaVersion                int    `json:"schemaVersion"`
	FixtureID                    string `json:"fixtureId"`
	ClinicCode                   string `json:"clinicCode"`
	ClinicOrdinal                int64  `json:"clinicOrdinal"`
	RunID                        string `json:"runId"`
	TargetReleaseCommit          string `json:"targetReleaseCommit"`
	TargetDatabaseIdentitySHA256 string `json:"targetDatabaseIdentitySha256"`
	InjectionCheckpoint          string `json:"injectionCheckpoint"`
	InjectionStage               string `json:"injectionStage"`
	InjectionMarker              string `json:"injectionMarker"`
	CopiedRowCount               int64  `json:"copiedRowCount"`
	ObservedSQLState             string `json:"observedSqlState"`
	TransactionStarted           bool   `json:"transactionStarted"`
	TransactionRolledBack        bool   `json:"transactionRolledBack"`
	BeforeBandCountsSHA256       string `json:"beforeBandCountsSha256"`
	AfterBandCountsSHA256        string `json:"afterBandCountsSha256"`
}

// CanonicalJSON returns the exact newline-terminated bytes whose SHA-256 is
// embedded in the failure apply report.
func (e TransactionEvidence) CanonicalJSON() ([]byte, error) {
	return canonicalSyntheticEvidence(e)
}

// SyntheticFailureResult is the aggregate receipt from the database runner.
type SyntheticFailureResult struct {
	StartedAt                 time.Time
	FailureInjectedAt         time.Time
	CompletedAt               time.Time
	BeforeBandCounts          BandCountsEvidence
	AfterBandCounts           BandCountsEvidence
	TransactionEvidence       TransactionEvidence
	BeforeBandCountsSHA256    string
	AfterBandCountsSHA256     string
	TransactionEvidenceSHA256 string
	BandRowCountBefore        int64
	BandRowCountAfter         int64
}

type failureRehearsalTarget interface {
	cutoverQuerier
}

type failureRehearsalTx interface {
	cutoverTransaction
}

type failureRehearsalBegin func(context.Context) (failureRehearsalTx, error)

// RunSyntheticFailureRehearsal runs the fixed G4 failure against a caller-
// attested disposable local database. It has no caller-selected SQL or fixture.
func RunSyntheticFailureRehearsal(
	ctx context.Context,
	pool *pgxpool.Pool,
	input SyntheticFailureInput,
) (SyntheticFailureResult, error) {
	return runSyntheticFailureRehearsal(
		ctx,
		pool,
		func(ctx context.Context) (failureRehearsalTx, error) {
			tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
			if err != nil {
				return nil, err
			}
			return pgxCutoverTransaction{Tx: tx}, nil
		},
		input,
	)
}

func runSyntheticFailureRehearsal(
	ctx context.Context,
	target failureRehearsalTarget,
	begin failureRehearsalBegin,
	input SyntheticFailureInput,
) (SyntheticFailureResult, error) {
	if err := validateSyntheticFailureInput(input); err != nil {
		return SyntheticFailureResult{}, err
	}
	manifest := syntheticFailureManifest(input)
	before, err := queryBandCounts(ctx, target, manifest.IDBand)
	if err != nil {
		return SyntheticFailureResult{}, err
	}
	if before.TotalRowCount != 0 {
		return SyntheticFailureResult{}, fmt.Errorf("synthetic failure target band must be empty before transaction")
	}
	beforeBytes, err := before.CanonicalJSON()
	if err != nil {
		return SyntheticFailureResult{}, err
	}
	beforeSHA := sha256Bytes(beforeBytes)

	startedAt := time.Now().UTC()
	failureInjectedAt, err := injectSyntheticFailureRollback(ctx, begin, input, manifest)
	if err != nil {
		return SyntheticFailureResult{}, err
	}

	after, err := queryBandCounts(ctx, target, manifest.IDBand)
	if err != nil {
		return SyntheticFailureResult{}, err
	}
	afterBytes, err := after.CanonicalJSON()
	if err != nil {
		return SyntheticFailureResult{}, err
	}
	afterSHA := sha256Bytes(afterBytes)
	if after.TotalRowCount != 0 || !bytes.Equal(beforeBytes, afterBytes) {
		return SyntheticFailureResult{}, fmt.Errorf("synthetic failure rollback changed the target band")
	}
	if err := PreflightCutoverTarget(ctx, target, manifest, input.Seeds); err != nil {
		return SyntheticFailureResult{}, fmt.Errorf("post-failure target preflight failed")
	}

	transactionEvidence := TransactionEvidence{
		SchemaVersion:                1,
		FixtureID:                    SyntheticFailureFixtureID,
		ClinicCode:                   input.ClinicCode,
		ClinicOrdinal:                input.ClinicOrdinal,
		RunID:                        input.RunID,
		TargetReleaseCommit:          input.TargetReleaseCommit,
		TargetDatabaseIdentitySHA256: input.TargetDatabaseIdentitySHA256,
		InjectionCheckpoint:          SyntheticFailureCheckpoint,
		InjectionStage:               SyntheticFailureStage,
		InjectionMarker:              SyntheticFailureMarker,
		CopiedRowCount:               1,
		ObservedSQLState:             syntheticFailureSQLState,
		TransactionStarted:           true,
		TransactionRolledBack:        true,
		BeforeBandCountsSHA256:       beforeSHA,
		AfterBandCountsSHA256:        afterSHA,
	}
	transactionBytes, err := transactionEvidence.CanonicalJSON()
	if err != nil {
		return SyntheticFailureResult{}, err
	}
	return SyntheticFailureResult{
		StartedAt:                 startedAt,
		FailureInjectedAt:         failureInjectedAt,
		CompletedAt:               time.Now().UTC(),
		BeforeBandCounts:          before,
		AfterBandCounts:           after,
		TransactionEvidence:       transactionEvidence,
		BeforeBandCountsSHA256:    beforeSHA,
		AfterBandCountsSHA256:     afterSHA,
		TransactionEvidenceSHA256: sha256Bytes(transactionBytes),
		BandRowCountBefore:        before.TotalRowCount,
		BandRowCountAfter:         after.TotalRowCount,
	}, nil
}

func injectSyntheticFailureRollback(
	ctx context.Context,
	begin failureRehearsalBegin,
	input SyntheticFailureInput,
	manifest CutoverManifest,
) (time.Time, error) {
	tx, err := begin(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("begin synthetic failure transaction")
	}
	rolledBack := false
	defer func() {
		if !rolledBack {
			_ = rollbackFailureTransaction(ctx, tx)
		}
	}()

	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '10s'`); err != nil {
		return time.Time{}, fmt.Errorf("configure synthetic failure lock timeout")
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, cutoverAdvisoryLockKey); err != nil {
		return time.Time{}, fmt.Errorf("acquire synthetic failure cutover lock")
	}
	if err := lockCutoverTables(ctx, tx); err != nil {
		return time.Time{}, fmt.Errorf("lock synthetic failure cutover tables")
	}
	if err := validateCutoverTarget(ctx, tx, manifest, input.Seeds, true); err != nil {
		return time.Time{}, fmt.Errorf("validate synthetic failure target")
	}
	if err := copySyntheticFailureOwner(ctx, tx, input.Seeds); err != nil {
		return time.Time{}, err
	}

	_, injectionErr := tx.Exec(
		ctx,
		`INSERT INTO pets (id, clinic_id, owner_id, name, animal_species_id)
VALUES ($1, $2, $3, $4, $5)`,
		syntheticFailurePetID,
		input.Seeds.ClinicID,
		syntheticFailureMissingOwnerID,
		"F8 synthetic pet",
		input.Seeds.AnimalSpeciesID,
	)
	failureInjectedAt := time.Now().UTC()
	var pgError *pgconn.PgError
	if !errors.As(injectionErr, &pgError) || pgError.Code != syntheticFailureSQLState {
		return time.Time{}, fmt.Errorf("synthetic injection did not produce the required foreign-key violation")
	}
	if err := rollbackFailureTransaction(ctx, tx); err != nil {
		return time.Time{}, fmt.Errorf("rollback synthetic failure transaction")
	}
	rolledBack = true
	return failureInjectedAt, nil
}

func rollbackFailureTransaction(ctx context.Context, tx failureRehearsalTx) error {
	rollbackContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), failureRollbackTimeout)
	defer cancel()
	return tx.Rollback(rollbackContext)
}

func validateSyntheticFailureInput(input SyntheticFailureInput) error {
	if !clinicCodePattern.MatchString(input.ClinicCode) {
		return fmt.Errorf("clinic code has an invalid format")
	}
	if input.ClinicOrdinal != 1 {
		return fmt.Errorf("synthetic failure rehearsal requires clinic ordinal 1")
	}
	if !runIDPattern.MatchString(input.RunID) {
		return fmt.Errorf("synthetic failure run ID has an invalid format")
	}
	if !commitPattern.MatchString(input.TargetReleaseCommit) {
		return fmt.Errorf("target release commit must be 40 lowercase hexadecimal characters")
	}
	for digest, label := range map[string]string{
		input.TargetDatabaseIdentitySHA256: "target database identity",
		input.FixtureManifestSHA256:        "fixture manifest",
		input.FailureRuntimeReportSHA256:   "failure runtime report",
	} {
		if !validSHA256(digest) {
			return fmt.Errorf("%s SHA-256 is invalid", label)
		}
	}
	if input.Seeds.ClinicID <= 0 ||
		input.Seeds.AnimalSpeciesID <= 0 ||
		input.Seeds.ExamTypeID <= 0 ||
		input.Seeds.TrimmingReservationTypeID <= 0 ||
		input.Seeds.CashPaymentMethodID <= 0 ||
		input.Seeds.CreditCardPaymentMethodID <= 0 {
		return fmt.Errorf("all six explicit target seed IDs must be positive")
	}
	return nil
}

// ValidateSyntheticFailureInput validates CLI boundary values before a target
// connection is opened.
func ValidateSyntheticFailureInput(input SyntheticFailureInput) error {
	return validateSyntheticFailureInput(input)
}

func syntheticFailureManifest(input SyntheticFailureInput) CutoverManifest {
	band := CutoverIDBand{
		Base:               0,
		EndExclusive:       clinicBandSize,
		NonOwnerIDOffset:   nonOwnerBandOffset,
		OwnerFloor:         ownerBandOffset,
		ApplicationIDFloor: applicationIDFloor,
	}
	manifest := CutoverManifest{
		ClinicCode:             input.ClinicCode,
		ClinicOrdinal:          input.ClinicOrdinal,
		SourceRunID:            input.RunID,
		ClinicBandBase:         band.Base,
		ClinicBandEndExclusive: band.EndExclusive,
		IDBand:                 band,
	}
	for _, spec := range CutoverTableSpecs() {
		manifest.Tables = append(manifest.Tables, CutoverManifestTable{
			Table: spec.Name,
		})
	}
	return manifest
}

func queryBandCounts(
	ctx context.Context,
	target cutoverQuerier,
	band CutoverIDBand,
) (BandCountsEvidence, error) {
	evidence := BandCountsEvidence{
		SchemaVersion: 1,
		TableCount:    len(CutoverTableSpecs()),
		Tables:        make([]BandCount, 0, len(CutoverTableSpecs())),
	}
	for _, spec := range CutoverTableSpecs() {
		floor := band.NonOwnerIDOffset
		if spec.Name == "owners" {
			floor = band.OwnerFloor
		}
		query := fmt.Sprintf(
			`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2`,
			pgx.Identifier{spec.Name}.Sanitize(),
		)
		var count int64
		if err := target.QueryRow(ctx, query, floor, band.EndExclusive).Scan(&count); err != nil {
			return BandCountsEvidence{}, fmt.Errorf("count synthetic failure band table %s", spec.Name)
		}
		if count < 0 {
			return BandCountsEvidence{}, fmt.Errorf("count synthetic failure band table %s returned a negative value", spec.Name)
		}
		evidence.Tables = append(evidence.Tables, BandCount{Table: spec.Name, RowCount: count})
		evidence.TotalRowCount += count
	}
	return evidence, nil
}

func copySyntheticFailureOwner(
	ctx context.Context,
	tx failureRehearsalTx,
	seeds CutoverSeedIDs,
) error {
	var contents bytes.Buffer
	writer := csv.NewWriter(&contents)
	if err := writer.Write([]string{"id", "clinic_id", "name"}); err != nil {
		return fmt.Errorf("build synthetic owner copy")
	}
	if err := writer.Write([]string{
		fmt.Sprintf("%d", syntheticFailureOwnerID),
		fmt.Sprintf("%d", seeds.ClinicID),
		"F8 synthetic owner",
	}); err != nil {
		return fmt.Errorf("build synthetic owner copy")
	}
	writer.Flush()
	if writer.Error() != nil {
		return fmt.Errorf("build synthetic owner copy")
	}
	tag, err := tx.CopyFrom(
		ctx,
		strings.NewReader(contents.String()),
		`COPY "owners" ("id", "clinic_id", "name") FROM STDIN WITH (FORMAT csv, HEADER true)`,
	)
	if err != nil {
		return fmt.Errorf("copy synthetic owner")
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("copy synthetic owner returned an unexpected row count")
	}
	return nil
}

func canonicalSyntheticEvidence[T BandCountsEvidence | TransactionEvidence](value T) ([]byte, error) {
	contents, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode synthetic failure evidence: %w", err)
	}
	return append(contents, '\n'), nil
}

func sha256Bytes(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
