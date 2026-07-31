package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIdentityLinksMigration_RLSStructure pins #239 identity-link RLS + immutability
// structure in the 004 archive block inside 001_init.sql (section 10). Application
// runtime scope remains the first boundary; RLS is defense-in-depth once FORCE is enabled.
func TestIdentityLinksMigration_RLSStructure(t *testing.T) {
	t.Parallel()

	sql := readIdentityLinkMigration(t)

	required := []string{
		"CREATE TABLE owner_identity_groups",
		"CREATE TABLE owner_identity_group_members",
		"CREATE TABLE pet_identity_groups",
		"CREATE TABLE pet_identity_group_members",
		"uq_owner_identity_active_member",
		"uq_pet_identity_active_member",
		"app_private.prevent_identity_group_created_clinic_id_update",
		"trg_owner_identity_groups_created_clinic_immutable",
		"trg_pet_identity_groups_created_clinic_immutable",
		"created_clinic_id is immutable",
		"SELECT app_private.apply_rls_policy(\n    'owner_identity_groups'",
		"SELECT app_private.apply_rls_policy(\n    'owner_identity_group_members'",
		"SELECT app_private.apply_rls_policy(\n    'pet_identity_groups'",
		"SELECT app_private.apply_rls_policy(\n    'pet_identity_group_members'",
		"app_private.has_clinic_access(created_clinic_id)",
		"app_private.has_clinic_access(clinic_id)",
		"tenant_owner_identity_groups_isolation",
		"tenant_owner_identity_group_members_isolation",
		"tenant_pet_identity_groups_isolation",
		"tenant_pet_identity_group_members_isolation",
		// CASCADE DELETE is forbidden project-wide for tenant tables
		"ON DELETE RESTRICT",
	}
	for _, snippet := range required {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("identity-links migration missing required RLS/structure snippet:\n%s", snippet)
		}
	}

	// Explicitly reject cascade on these tables.
	if strings.Contains(sql, "ON DELETE CASCADE") {
		t.Fatal("identity-links migration must not use ON DELETE CASCADE")
	}

	// Do not re-add parent UNIQUE that already exists earlier in 001_init (comments may mention names).
	if strings.Contains(sql, "ADD CONSTRAINT uq_owners_clinic_id_id") ||
		strings.Contains(sql, "ADD CONSTRAINT uq_pets_clinic_id_id") ||
		strings.Contains(sql, "UNIQUE (clinic_id, id)") {
		t.Fatal("identity-links migration must not re-add parent owner/pet (clinic_id, id) UNIQUE")
	}
}

// TestIdentityLinksMigration_ArchiveHash pins the section-10 004 body SHA-256 so
// consolidation cannot silently rewrite the archived identity-link DDL.
func TestIdentityLinksMigration_ArchiveHash(t *testing.T) {
	t.Parallel()

	sql := readIdentityLinkMigration(t)
	const wantSHA = "17c0028a0c294ab278b6e8b9512ddbfd5f5e2977229a3280a4e816fc41ecf829"
	if !strings.Contains(sql, "-- Source SHA-256: "+wantSHA) {
		t.Fatalf("identity-links archive missing expected Source SHA-256 %s", wantSHA)
	}
	if !strings.Contains(sql, "-- Source file: 004_add_identity_links.sql") {
		t.Fatal("identity-links archive missing Source file marker for 004_add_identity_links.sql")
	}
}

func readIdentityLinkMigration(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", "001_init.sql")
	b, err := os.ReadFile(path) //nolint:gosec // fixed relative migration path
	if err != nil {
		t.Fatalf("read 001_init migration for identity-links archive: %v", err)
	}
	const marker = "-- Source file: 004_add_identity_links.sql\n"
	full := string(b)
	start := strings.Index(full, marker)
	if start < 0 {
		t.Fatal("identity-links archive block missing from 001_init.sql (expected section 10 Source file marker)")
	}
	rest := full[start:]
	// Body ends at the next archive Source file marker (005) or EOF.
	next := strings.Index(rest[len(marker):], "\n-- Source file:")
	if next < 0 {
		return rest
	}
	return rest[:len(marker)+next+1]
}
