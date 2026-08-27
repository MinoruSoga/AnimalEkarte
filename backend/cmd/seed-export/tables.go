package main

// bundleTables lists, per bundle, exactly the tables owned by that bundle in
// FK-safe load order. The order is a sub-sequence of 001_init.sql's CREATE
// TABLE order. Clinical rows are not seeded; they come from old_db handoff.
var bundleTables = map[string][]string{
	"002_master": {
		"companies",
		"animal_species",
		"lstep_auto_managed_prefixes",
		"lstep_condition_tag_mappings",
		"lstep_send_purpose_tag_prefixes",
		"clinics",
		"clinic_settings",
		"exam_types",
		"reservation_types",
		// clinics INSERT trigger also auto-inserts defaults; cmd/migrate
		// disables that trigger when this table is in the manifest so
		// explicit seed IDs remain authoritative.
		"payment_methods",
		"permission_groups",
		"permission_group_rules",
	},
}

var bundleOrder = []string{"002_master"}

const totalSeedTableCount = 12
