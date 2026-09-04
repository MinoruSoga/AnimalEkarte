// cmd/seed-old-db loads old_db migration-output TSV files into the local dev DB.
//
// Safety boundary: exits non-zero if DB_HOST is not "db", "localhost", or "127.0.0.1".
// Run via: docker compose -f docker-compose.yml -f docker-compose.seed-old-db.yml run --rm seed-old-db
//
// Required env vars (shared with migrate):
//
//	DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
//
// Optional:
//
//	OLD_DB_MIGRATION_OUTPUT_DIR  path to migration-output dir (default /old-db-data)
//	OLD_DB_MANIFEST_PATH         path to new-schema-import-manifest.json
package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
)

// ManifestEntry is one entry in new-schema-import-manifest.json.
type ManifestEntry struct {
	PlanKey          string   `json:"planKey"`
	TargetTable      string   `json:"targetTable"`
	LegacyTable      string   `json:"legacyTable"`
	OutputFile       string   `json:"outputFile"`
	OutputHeaderFile string   `json:"outputHeaderFile"`
	TargetColumns    []string `json:"targetColumns"`
	// CrosswalkColumns are join-only metadata (__xw_*) the old_db generator
	// appends to recoverable child-detail TSVs so the seeder can resolve the
	// parent FK; they are not target columns and are dropped before INSERT.
	CrosswalkColumns []string `json:"crosswalkColumns"`
	RowCount         int      `json:"rowCount"`
	GitIgnored       bool     `json:"gitIgnored"`
}

// Manifest is the top-level structure of new-schema-import-manifest.json.
type Manifest struct {
	FileCount    int             `json:"fileCount"`
	TotalRows    int             `json:"totalRows"`
	Entries      []ManifestEntry `json:"entries"`
	SafetyPolicy string          `json:"safetyPolicy"`
}

// loadResult summarises the outcome for one manifest entry.
type loadResult struct {
	PlanKey string
	Table   string
	File    string
	Status  string // "loaded", "skipped", "error"
	Rows    int64
	Reason  string
}

type lookupCache struct {
	Pets             map[string]petLookup
	OwnerIDs         map[int64]bool
	AnimalSpeciesIDs map[int64]bool

	// Composite crosswalk caches keyed on the OLD legacy key tuple
	// (old pet_number + old record_no/sno), built with compositeKey(). These
	// resolve a child-detail row to its parent's NEW bigint id.
	//
	// record_no ALONE is non-unique in the legacy data (e.g. record_no
	// '00000001' repeats 15,101× across pets), so a record_no-keyed cache would
	// silently mis-link children to the wrong parent. The composite
	// (pet_no, record_no) is unique for medical_records (TBL_KARTE_DATA is 1:1)
	// and uniquely identifies the parent billing/exam for the line-item sources.
	MedicalRecords map[string]int64 // (pet_no, record_no) → medical_records.id
	Billings       map[string]int64 // (pet_no, record_no) → billings.id
	Exams          map[string]int64 // (pet_no, sno)       → exams.id

	FallbackAnimalSpeciesID int64
}

// qualifyRecordNo builds the pet-qualified record_no that is STORED in
// medical_records.record_no for the legacy seed: "<pet_no>-<record_no>".
//
// The new schema enforces UNIQUE(clinic_id, record_no), but legacy record_no
// (Ksk_KarteNo / TReat_Sno / HkSK_SNo) is unique only PER PET — record_no
// "00000001" repeats across 15,101 different pets. With every seeded row in the
// single synthetic clinic (all old_db rows resolve to one clinic), the raw record_no would collapse 425,545
// medical_records to 1,297 via ON CONFLICT. Qualifying with the pet number makes
// it unique within clinic 1, keeps it human-readable and traceable to the legacy
// chart, and — because it is also the composite cache key — lets every
// child-detail row resolve its parent unambiguously.
//
// This is also the in-memory composite cache key: it is globally unique, so the
// caches key on this single string and both the parent (cache build) and child
// (lookup) sides compute it identically.
func qualifyRecordNo(petNo, recordNo string) string {
	return petNo + "-" + recordNo
}

type petLookup struct {
	ID       int64
	ClinicID int64
	OwnerID  int64
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if err := config.ConfigureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}

	dbName := os.Getenv("DB_NAME")
	conn, err := dbconn.FromEnv()
	if err != nil || dbName == "" {
		return fmt.Errorf("missing required DB env vars (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)")
	}
	dbHost := conn.Host

	// Local-only safety guard
	if !dbconn.IsLocalHost(dbHost) {
		return fmt.Errorf(
			"SAFETY: DB_HOST=%q is not a known local host — refusing to seed against non-local DB",
			dbHost,
		)
	}

	outputDir := os.Getenv("OLD_DB_MIGRATION_OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "/old-db-data"
	}

	manifestPath := os.Getenv("OLD_DB_MANIFEST_PATH")
	if manifestPath == "" {
		manifestPath = filepath.Join(outputDir, "..", "..", "docs", "generated", "new-schema-import-manifest.json")
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("OLD_DB_MIGRATION_OUTPUT_DIR=%q does not exist — is the volume mounted?", outputDir)
	}

	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest from %s: %w", manifestPath, err)
	}

	logger.Info("Manifest loaded",
		slog.String("path", manifestPath),
		slog.Int("entries", len(manifest.Entries)),
		slog.Int("totalRows", manifest.TotalRows),
	)

	connStr := conn.DSN(dbName)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to open pgx pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("Connected to database", slog.String("host", dbHost), slog.String("dbname", dbName))

	// Pre-flight: verify all TSV files exist before touching the DB
	missing := []string{}
	for _, entry := range manifest.Entries {
		for _, f := range []string{entry.OutputFile, entry.OutputHeaderFile} {
			if _, err := os.Stat(filepath.Join(outputDir, f)); os.IsNotExist(err) {
				missing = append(missing, f)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing %d file(s) from %s:\n  %s",
			len(missing), outputDir, strings.Join(missing, "\n  "))
	}

	// Process entries in FK-safe dependency order.
	// The manifest is sorted alphabetically by planKey; we re-sort by table tier
	// so parent tables are always loaded before child tables.
	orderedEntries := orderByDependency(manifest.Entries)

	results := make([]loadResult, 0, len(orderedEntries))
	errCount := 0
	cache := &lookupCache{
		Pets:             map[string]petLookup{},
		MedicalRecords:   map[string]int64{},
		Billings:         map[string]int64{},
		Exams:            map[string]int64{},
		OwnerIDs:         map[int64]bool{},
		AnimalSpeciesIDs: map[int64]bool{},
	}

	// Re-runnable safety: child-detail / line-item tables have only a BIGSERIAL
	// PK and no natural unique key, so ON CONFLICT DO NOTHING is a no-op for them
	// and a second seed run on a non-fresh DB would duplicate every row. Truncate
	// them up front (reverse-dependency order) so the seed is idempotent even
	// without a full DB reset. Parents with natural keys (owners/pets/...) are
	// not truncated; their ON CONFLICT DO NOTHING handles re-runs.
	if err := truncateChildDetailTables(ctx, pool, logger); err != nil {
		return fmt.Errorf("truncate child-detail tables: %w", err)
	}

	for _, entry := range orderedEntries {
		res := processEntry(ctx, pool, logger, entry, outputDir, cache)
		results = append(results, res)
		if res.Status == "error" {
			errCount++
		}
		if res.Status == "loaded" {
			switch entry.TargetTable {
			case "animal_species":
				if err := refreshAnimalSpeciesCache(ctx, pool, cache); err != nil {
					return fmt.Errorf("refresh animal species lookup cache: %w", err)
				}
			case "owners":
				if err := refreshOwnerCache(ctx, pool, cache); err != nil {
					return fmt.Errorf("refresh owner lookup cache: %w", err)
				}
			case "pets":
				if err := refreshPetCache(ctx, pool, cache); err != nil {
					return fmt.Errorf("refresh pet lookup cache: %w", err)
				}
			case "medical_records":
				if err := refreshMedicalRecordCache(ctx, pool, cache); err != nil {
					return fmt.Errorf("refresh medical record lookup cache: %w", err)
				}
			case "billings":
				if err := refreshBillingCache(ctx, pool, cache); err != nil {
					return fmt.Errorf("refresh billing lookup cache: %w", err)
				}
			case "exams":
				if err := refreshExamCache(ctx, pool, cache); err != nil {
					return fmt.Errorf("refresh exam lookup cache: %w", err)
				}
			}
		}
	}

	// Summary
	logger.Info("─────────────────────────────────────────")
	loaded, skipped, errored := 0, 0, 0
	totalRows := int64(0)
	for _, r := range results {
		switch r.Status {
		case "loaded":
			loaded++
			totalRows += r.Rows
			logger.Info("✓ loaded", slog.String("planKey", r.PlanKey), slog.Int64("rows", r.Rows))
		case "skipped":
			skipped++
			logger.Warn("⏭ skipped", slog.String("planKey", r.PlanKey), slog.String("reason", r.Reason))
		case "error":
			errored++
			logger.Error("✗ error", slog.String("planKey", r.PlanKey), slog.String("reason", r.Reason))
		}
	}
	logger.Info("─────────────────────────────────────────")
	logger.Info("Totals",
		slog.Int("loaded", loaded),
		slog.Int("skipped", skipped),
		slog.Int("errors", errored),
		slog.Int64("totalRowsInserted", totalRows),
	)

	if errCount > 0 {
		return fmt.Errorf("%d file(s) failed — see errors above", errCount)
	}

	// Reset BIGSERIAL sequences for tables where we inserted explicit IDs.
	// Without this, the next application-level INSERT would conflict with existing IDs.
	if err := resetSequences(ctx, pool, logger); err != nil {
		return fmt.Errorf("sequence reset failed: %w", err)
	}

	logger.Info("✓ old-db seed completed successfully")
	return nil
}

// resetSequences advances BIGSERIAL sequences to max(id)+1 for all tables
// where explicit numeric IDs were inserted from the TSV.
func resetSequences(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	tables := []string{
		"animal_species",
		"clinics",
		"owners",
		"exam_types",
		"merchandise_items",
		"procedures",
	}
	for _, t := range tables {
		sql := fmt.Sprintf(
			"SELECT setval(pg_get_serial_sequence('public.%s', 'id'), COALESCE(MAX(id), 0) + 1, false) FROM public.%s",
			t, t,
		)
		var newVal int64
		if err := pool.QueryRow(ctx, sql).Scan(&newVal); err != nil {
			return fmt.Errorf("setval %s: %w", t, err)
		}
		logger.Info("sequence reset", slog.String("table", t), slog.Int64("nextVal", newVal))
	}
	return nil
}

func refreshPetCache(ctx context.Context, pool *pgxpool.Pool, cache *lookupCache) error {
	rows, err := pool.Query(ctx, `
		SELECT pet_number, id, clinic_id, owner_id
		FROM public.pets
		WHERE pet_number <> ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	next := map[string]petLookup{}
	for rows.Next() {
		var petNumber string
		var pet petLookup
		if err := rows.Scan(&petNumber, &pet.ID, &pet.ClinicID, &pet.OwnerID); err != nil {
			return err
		}
		if _, exists := next[petNumber]; !exists {
			next[petNumber] = pet
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	cache.Pets = next
	return nil
}

func refreshOwnerCache(ctx context.Context, pool *pgxpool.Pool, cache *lookupCache) error {
	rows, err := pool.Query(ctx, `SELECT id FROM public.owners`)
	if err != nil {
		return err
	}
	defer rows.Close()

	next := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		next[id] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	cache.OwnerIDs = next
	return nil
}

func refreshAnimalSpeciesCache(ctx context.Context, pool *pgxpool.Pool, cache *lookupCache) error {
	rows, err := pool.Query(ctx, `SELECT id FROM public.animal_species ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	next := map[int64]bool{}
	var fallback int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		next[id] = true
		if fallback == 0 {
			fallback = id
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	cache.AnimalSpeciesIDs = next
	cache.FallbackAnimalSpeciesID = fallback
	return nil
}

// refreshMedicalRecordCache builds the qualified-record_no → medical_records.id
// crosswalk. medical_records.record_no is stored pet-qualified
// ("<pet_no>-<record_no>", see qualifyRecordNo), which is globally unique, so the
// cache keys directly on the stored value. Child rows compute the same qualified
// key from (__xw_pet_no, __xw_record_no) and resolve their parent unambiguously.
func refreshMedicalRecordCache(ctx context.Context, pool *pgxpool.Pool, cache *lookupCache) error {
	m, err := buildRecordNoCache(ctx, pool, `
		SELECT mr.record_no, mr.id
		FROM public.medical_records mr
		WHERE mr.record_no <> ''
		ORDER BY mr.id
	`)
	if err != nil {
		return err
	}
	cache.MedicalRecords = m
	return nil
}

// refreshBillingCache builds qualified-record_no → billings.id for the two-hop
// billing_items → billings link. billings stores no record_no, but its
// medical_record_id is UNIQUE and resolved from the same visit number, so the key
// is recovered by joining back through medical_records (whose record_no is the
// qualified value). billings with a NULL medical_record_id cannot be keyed, and
// their line items are skipped rather than mis-linked or orphaned.
func refreshBillingCache(ctx context.Context, pool *pgxpool.Pool, cache *lookupCache) error {
	m, err := buildRecordNoCache(ctx, pool, `
		SELECT mr.record_no, b.id
		FROM public.billings b
		JOIN public.medical_records mr ON mr.id = b.medical_record_id
		WHERE mr.record_no <> ''
		ORDER BY b.id
	`)
	if err != nil {
		return err
	}
	cache.Billings = m
	return nil
}

// refreshExamCache builds qualified-record_no → exams.id for the two-hop
// exam_results → exams link, by joining exams back through medical_records.
func refreshExamCache(ctx context.Context, pool *pgxpool.Pool, cache *lookupCache) error {
	m, err := buildRecordNoCache(ctx, pool, `
		SELECT mr.record_no, e.id
		FROM public.exams e
		JOIN public.medical_records mr ON mr.id = e.medical_record_id
		WHERE mr.record_no <> ''
		ORDER BY e.id
	`)
	if err != nil {
		return err
	}
	cache.Exams = m
	return nil
}

// buildRecordNoCache runs a (record_no, id) query and returns a first-id-wins map
// keyed on the globally-unique qualified record_no. Shared by the three
// composite caches.
func buildRecordNoCache(ctx context.Context, pool *pgxpool.Pool, query string) (map[string]int64, error) {
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	next := map[string]int64{}
	for rows.Next() {
		var recordNo string
		var id int64
		if err := rows.Scan(&recordNo, &id); err != nil {
			return nil, err
		}
		if _, exists := next[recordNo]; !exists {
			next[recordNo] = id
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return next, nil
}

// childDetailTruncateOrder lists the child-detail / line-item tables the seeder
// loads, in reverse-dependency order (children before any table they reference).
// They have no natural unique key, so they are truncated up front to keep the
// seed idempotent across re-runs (see truncateChildDetailTables).
var childDetailTruncateOrder = []string{
	"exam_results",
	"billing_items",
	"treatments",
	"vital_records",
	"clinical_plans",
	"inquiries",
}

func truncateChildDetailTables(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	for _, t := range childDetailTruncateOrder {
		// RESTART IDENTITY keeps BIGSERIAL ids stable across re-seeds; CASCADE is
		// unnecessary (these are leaf tables) but harmless if a future FK is added.
		sql := fmt.Sprintf("TRUNCATE TABLE public.%s RESTART IDENTITY CASCADE", quoteIdent(t))
		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("truncate %s: %w", t, err)
		}
	}
	logger.Info("child-detail tables truncated for idempotent re-seed",
		slog.Int("tables", len(childDetailTruncateOrder)))
	return nil
}

// processEntry handles one manifest entry: validate header → stream rows → INSERT to target.
func processEntry(
	ctx context.Context,
	pool *pgxpool.Pool,
	logger *slog.Logger,
	entry ManifestEntry,
	outputDir string,
	cache *lookupCache,
) loadResult {
	res := loadResult{PlanKey: entry.PlanKey, Table: entry.TargetTable, File: entry.OutputFile}

	if reason := skipReason(entry); reason != "" {
		res.Status = "skipped"
		res.Reason = reason
		return res
	}

	headerPath := filepath.Join(outputDir, entry.OutputHeaderFile)
	tsvPath := filepath.Join(outputDir, entry.OutputFile)

	headerData, err := os.ReadFile(headerPath) //nolint:gosec // path from validated outputDir
	if err != nil {
		res.Status = "error"
		res.Reason = fmt.Sprintf("read header: %v", err)
		return res
	}
	rawCols := strings.Split(strings.TrimRight(string(headerData), "\n\r"), "\t")

	allowed := allowedColumns(entry.TargetTable)
	filteredCols, colIndexes, dropped := filterColumnsWithIndex(rawCols, allowed)
	if len(dropped) > 0 {
		logger.Warn("Dropping unmapped TSV columns",
			slog.String("planKey", entry.PlanKey),
			slog.String("dropped", strings.Join(dropped, ", ")),
		)
	}
	if len(filteredCols) == 0 {
		res.Status = "skipped"
		res.Reason = "all TSV columns filtered out (none map to target table)"
		return res
	}

	tsvFile, err := os.Open(tsvPath) //nolint:gosec // path from validated outputDir
	if err != nil {
		res.Status = "error"
		res.Reason = fmt.Sprintf("open tsv: %v", err)
		return res
	}
	defer tsvFile.Close() //nolint:errcheck

	rows, err := insertFromTSV(ctx, pool, logger, entry, filteredCols, colIndexes, tsvFile, cache)
	if err != nil {
		res.Status = "error"
		res.Reason = fmt.Sprintf("load: %v", err)
		return res
	}

	res.Status = "loaded"
	res.Rows = rows
	return res
}

// insertFromTSV builds INSERT SQL with typed casts and uses pgx CopyFrom for bulk insert.
func insertFromTSV(
	ctx context.Context,
	pool *pgxpool.Pool,
	logger *slog.Logger,
	entry ManifestEntry,
	cols []string,
	colIndexes []int,
	tsvFile *os.File,
	cache *lookupCache,
) (int64, error) {
	// Build the INSERT SQL template with per-column casts.
	// We use a VALUES approach via pgx.CopyFrom (which maps to COPY FROM).
	// The target-facing INSERT is built as:
	//   INSERT INTO table (col1, col2, ...) VALUES ($1::cast, $2::cast, ...) ON CONFLICT DO NOTHING
	// executed in batches via pool.SendBatch.
	//
	// readCols is what we extract from each TSV line into args (real columns +
	// any __xw_* crosswalk metadata + an appended FK column for crosswalk
	// children). insertNames/insertSrc is the subset that goes into the INSERT:
	// crosswalk metadata is dropped, the resolved FK is kept. insertSrc[k] is the
	// index into args supplying INSERT column k, so unused __xw_* slots never
	// reach pgx (which rejects supplied-but-unreferenced params).
	readCols := withCrosswalkFKColumn(entry, cols)
	insertNames, insertSrc := insertColumns(readCols)

	quotedCols := make([]string, len(insertNames))
	paramExprs := make([]string, len(insertNames))
	for k, c := range insertNames {
		quotedCols[k] = quoteIdent(c)
		paramExprs[k] = buildParamExpr(k+1, entry, c)
	}

	// Add synthetic columns (NOT NULL without DEFAULT that TSV doesn't supply).
	// Their params follow the insert columns, so pass len(insertNames) as the base.
	synCols, synExprs := syntheticColumns(entry, insertNames)
	for i, c := range synCols {
		quotedCols = append(quotedCols, quoteIdent(c))
		paramExprs = append(paramExprs, synExprs[i])
	}

	conflict := conflictClause()

	insertSQL := fmt.Sprintf(
		"INSERT INTO public.%s (%s) VALUES (%s)%s",
		quoteIdent(entry.TargetTable),
		strings.Join(quotedCols, ", "),
		strings.Join(paramExprs, ", "),
		conflict,
	)

	logger.Debug("INSERT template", slog.String("planKey", entry.PlanKey), slog.String("sql", insertSQL))

	// Stream the TSV and batch-execute
	scanner := bufio.NewScanner(tsvFile)
	scanner.Buffer(make([]byte, 16*1024*1024), 16*1024*1024) // 16 MB line buffer for long text fields

	const batchSize = 1000
	batch := &pgx.Batch{}
	insertedRows := int64(0)
	lineNum := 0

	flushBatch := func() (int64, error) {
		if batch.Len() == 0 {
			return 0, nil
		}
		br := pool.SendBatch(ctx, batch)
		inserted := int64(0)
		for i := 0; i < batch.Len(); i++ {
			tag, err := br.Exec()
			if err != nil {
				_ = br.Close()
				return 0, fmt.Errorf("batch exec at line ~%d: %w", lineNum, err)
			}
			inserted += tag.RowsAffected()
		}
		if err := br.Close(); err != nil {
			return 0, fmt.Errorf("batch close: %w", err)
		}
		batch = &pgx.Batch{}
		return inserted, nil
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")

		// Extract into args aligned with readCols. colIndexes maps each original
		// filtered column to its TSV field index; any appended FK column (index
		// >= len(colIndexes)) has no TSV source and starts nil, to be filled by
		// normalizeArgs from the composite cache.
		args := make([]any, len(readCols))
		for i, idx := range colIndexes {
			if idx < len(fields) {
				v := fields[idx]
				if v == `\N` || v == "" {
					args[i] = nil
				} else {
					args[i] = v
				}
			} else {
				args[i] = nil
			}
		}
		if shouldSkipRow(entry, readCols, args) {
			continue
		}
		if normalizeArgs(entry, readCols, args, cache) {
			continue
		}

		// Project args down to the INSERT columns (drops __xw_* metadata, keeps
		// the resolved FK), preserving order so they align with $1..$N.
		queueArgs := make([]any, len(insertSrc))
		for k, src := range insertSrc {
			queueArgs[k] = args[src]
		}

		batch.Queue(insertSQL, queueArgs...)

		if batch.Len() >= batchSize {
			inserted, err := flushBatch()
			if err != nil {
				return 0, err
			}
			insertedRows += inserted
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan TSV: %w", err)
	}
	inserted, err := flushBatch()
	if err != nil {
		return 0, err
	}
	insertedRows += inserted

	return insertedRows, nil
}
