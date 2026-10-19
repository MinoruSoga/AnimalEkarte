// Package seedbundle defines the shared CSV seed-bundle manifest schema used
// by both the producer (cmd/seed-export, which dumps a disposable DB's final
// row state to CSV) and the consumer (cmd/migrate, which loads those CSVs at
// startup). Keeping the schema in one place prevents the two independent
// binaries from drifting on JSON field names.
//
// Table ownership per bundle is assigned by "earliest seed file that first
// INSERT INTOs the table" (not "last file", and not "every file that touches
// it"). This was verified against the 200 FK edges among the 90 seeded tables
// in 001_init.sql: assigning a table to the LAST touching file produced 38 FK
// load-order violations (e.g. clinics touched by both 003 and 004 would load
// in the 004 bundle, after 003-only child tables that reference it already
// tried to load). Assigning to the EARLIEST touching file produced zero
// violations, because 004 (staging) only ever adds more rows to tables 003
// (demo) already created — it never introduces a new parent relationship.
package seedbundle

import "strings"

// BundleOrder is the fixed, FK-safe load order of seed bundles (directory
// names relative to backend/migrations/seeds/). It mirrors the historical
// migration filename order (002 before 003 before 004) now that the stub
// 00[2-4]_seed_*.sql files that used to carry that ordering have been
// deleted — this slice is the sole source of full load order for
// non-production environments and for tooling that enumerates every bundle
// on disk (seed-export, lint). Runtime migrate selection is
// BundleOrderForEnv / PostLoginBundleOrderForEnv.
var BundleOrder = []string{"002_master", "004_demo"}

// BundleOrderForEnv returns the seed bundles loaded BEFORE the login seed
// for the given application environment value (APP_ENV). Fail-closed: every
// environment receives master only, so repository credentials never land on
// production or shared staging via migrate.
func BundleOrderForEnv(env string) []string {
	out := make([]string, 1)
	out[0] = "002_master"
	return out
}

// PostLoginBundleOrderForEnv returns the seed bundles loaded AFTER the login
// seed (cmd/migrate phase 3). 004_demo carries clinic-1 demo business data
// that references login-seed staffs, so it must load after phase 3 and only
// on the dedicated STG demo environment. It contains no accounts/staffs, so
// the SEC-CS2-F01 credential rule (credentials never leave phase 3) is
// unchanged; production and local dev remain master-only.
func PostLoginBundleOrderForEnv(env string) []string {
	if strings.EqualFold(strings.TrimSpace(env), "staging") {
		return []string{"004_demo"}
	}
	return nil
}

// AllBundlesForEnv returns the full migrate load plan for an environment in
// load order (pre-login bundles, then post-login bundles). It is used by
// cmd/migrate's key-coverage check so expected keys match the plan exactly.
func AllBundlesForEnv(env string) []string {
	out := append([]string{}, BundleOrderForEnv(env)...)
	out = append(out, PostLoginBundleOrderForEnv(env)...)
	return out
}

// BundleMigrationKey returns the schema_migrations.filename key used to
// record a seed bundle's application. The "seeds/" prefix keeps this key
// space disjoint from real *.sql migration filenames by construction, so a
// legacy key from before the 2026-07 CSV-only migration (see
// LegacyStubFilenames) can never collide with it.
func BundleMigrationKey(bundleDir string) string {
	return "seeds/" + bundleDir
}

// LegacyStubFilenames are the pre-2026-07 stub SQL migration filenames that
// used to record seed-bundle application directly in schema_migrations
// (SELECT 1 no-op bodies; the actual CSV load was already keyed off these
// filenames — see the deleted seedbundle.BundleDirs). A database migrated by
// an older binary before these files were deleted from the repo will still
// carry these keys but will never gain the new "seeds/<bundle>" keys, since
// nothing on disk maps to them anymore. cmd/migrate checks for their
// presence at startup and, if any are found, baselines all legacy-equivalent
// seeds/<bundle> keys as applied (never just the bundle matching the specific
// legacy key found —
// see legacyTranslationTargets in cmd/migrate) rather than fail-fasting or
// silently reinterpreting/skipping seed application — see
// detectLegacySeedKeys in cmd/migrate.
var LegacyStubFilenames = []string{
	"002_seed_master.sql",
	"003_seed_demo.sql",
	"004_seed_staging.sql",
}

// ManifestEntry is one seeded table within a bundle, in FK-safe load order.
type ManifestEntry struct {
	Table   string `json:"table"`
	CSVFile string `json:"csvFile"`
}

// Manifest is the top-level structure of each bundle's manifest.json.
type Manifest struct {
	Bundle string          `json:"bundle"`
	Tables []ManifestEntry `json:"tables"`
}
