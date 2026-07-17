package main

// bundleTables lists, per bundle, exactly the tables owned by that bundle in
// FK-safe load order (a table's parents always appear before it, both within
// a bundle and across bundle boundaries since 002 loads before 003 before
// 004). The order is a sub-sequence of 001_init.sql's CREATE TABLE order,
// which is itself a valid topological order of the FK graph — filtering a
// topological order to a subset preserves the topological property.
//
// Ownership assignment (earliest seed file touching each table) and FK-safety
// of this exact table order were both verified mechanically against
// 001_init.sql and 002/003/004's INSERT INTO statements before this file was
// written; see the package doc in internal/seedbundle for the method and
// result. Do not hand-edit this list without re-running that verification —
// silent FK violations only surface as a COPY failure against a freshly
// created disposable DB, not at review time.
var bundleTables = map[string][]string{
	"002_master": {
		"companies",
		"animal_species",
		"lstep_auto_managed_prefixes",
		"lstep_condition_tag_mappings",
		"lstep_send_purpose_tag_prefixes",
	},
	"003_demo": {
		"clinics",
		"clinic_integrations",
		"occupations",
		"accounts",
		"staffs",
		"owners",
		"lstep_tag_cache",
		"lstep_trigger_priorities",
		"lstep_delivery_trigger_log",
		"lstep_csv_imports",
		"lstep_friend_attribute_snapshots",
		"inventory_items",
		"exam_types",
		"exam_type_fields",
		"vaccines",
		"medicines",
		"insurances",
		"cages",
		"reservation_type_groups",
		"reservation_types",
		"staff_reservation_capabilities",
		"consultations",
		"procedures",
		"hospitalization_plans",
		// FK parent of trimming_courses.course_type_id. Must load before
		// trimming_courses. clinics INSERT trigger also auto-inserts defaults;
		// cmd/migrate disables that trigger when this table is in the manifest
		// so explicit seed IDs remain authoritative.
		"trimming_course_types",
		"trimming_courses",
		"trimming_options",
		"diagnosis_types",
		"diagnosis_names",
		"checkup_types",
		// checkup_type_fields の旧 seed（003_seed_demo.sql J-12b）は DO $$ ブロック内
		// INSERT だったため、トップレベル INSERT INTO 走査ベースの初回検証から漏れた。
		// FK 親は clinics / checkup_types（いずれも上に列挙済み）。
		"checkup_type_fields",
		"chief_complaint_types",
		"inquiry_templates",
		"pets",
		"pet_chronic_conditions",
		"staff_clinic_assignments",
		"permission_groups",
		"permission_group_rules",
		"staff_permission_groups",
		"line_customers",
		"shared_files",
		"line_send_logs",
		"appointments",
		"hospitalizations",
		"appointment_trimming_options",
		"medical_records",
		"prescriptions",
		"medical_record_addenda",
		"vaccinations",
		"checkups",
		"exams",
		"inquiries",
		"clinical_plans",
		"treatments",
		"treatment_plans",
		"medical_record_images",
		"billing_confirmations",
		"estimates",
		"exam_results",
		"daily_records",
		"vital_records",
		"care_plan_items",
		"merchandise_items",
		"estimate_items",
		"care_logs",
		"staff_notes",
		"billings",
		"billing_items",
		// FK parent of payments/payment_splits.payment_method_id. Must load
		// before payments. clinics INSERT trigger also auto-inserts defaults;
		// cmd/migrate disables that trigger when this table is in the manifest
		// so explicit seed IDs remain authoritative.
		"payment_methods",
		"payments",
		"payment_splits",
		"billing_refunds",
		"shift_entries",
		"clinic_holidays",
		"clinic_settings",
		"closing_special_periods",
		"cash_register_closes",
		"audit_logs",
		"line_reservation_settings",
		"staff_reservation_exclusions",
		"shift_entry_breaks",
		"shift_templates",
		"shift_template_breaks",
		"reservation_type_unavailable_times",
		"reservation_type_occupations",
		"manual_articles",
		"manual_article_versions",
	},
	"004_staging": {
		"appointment_trimming_details",
	},
}

// bundleOrder is the fixed load order of bundles (mirrors migration filename
// order: 002 before 003 before 004).
var bundleOrder = []string{"002_master", "003_demo", "004_staging"}

// totalSeedTableCount is a guard-rail constant checked at runtime against
// len(bundleTables flattened) so a future hand-edit that drops or duplicates
// a table entry fails loudly instead of silently shipping an incomplete
// export. 93 is the verified unique seeded-table count (5 + 87 + 1) — not
// 139, which was the sum of raw INSERT-touched-table counts per file before
// dedup, and not the 108 CREATE TABLE count in 001_init.sql (which includes
// tables no seed file ever populates). The original count of 90 missed
// checkup_type_fields, whose old seed lived inside a DO $$ block (J-12b) that
// the top-level INSERT INTO scan did not see. 91→92 when trimming_course_types
// was added; 92→93 when payment_methods was added as the explicit FK parent
// of payments/payment_splits.payment_method_id.
const totalSeedTableCount = 93
