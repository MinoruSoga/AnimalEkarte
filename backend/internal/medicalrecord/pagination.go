// Package medicalrecord owns the medicalrecord domain's HTTP, application logic, and
// persistence code (BE9-2A target:medicalrecord, ADR-006). This is the BE9-2C first slice —
// the out-dom=0 master-CRUD vertical (diagnosis type/name, examination type, chief
// complaint type) from boundary map §3.7's sub-batch ①. The domain's higher-risk paths
// (finalize-lock row protection, treatment/vital/clinical_plan, lab import saga,
// hospitalization/discharge-with-billing) remain in internal/service|handler|repository
// pending BE9-2D.
package medicalrecord

import "gorm.io/gorm"

// paginate converts a 1-origin page/limit pair into a GORM Offset/Limit scope. Consolidated
// here from the (byte-identical) local copies previously duplicated across
// internal/repository/diagnosistype and internal/repository/diagnosisname — repohelpers has
// no generic Paginate helper yet (rule-of-three hoist tracked as a repohelpers follow-up;
// see BE-refactor.md §6 step 2), and duplicating the same function twice in one package
// would not compile. Reused by every medicalrecord repository that needs page/limit.
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}
