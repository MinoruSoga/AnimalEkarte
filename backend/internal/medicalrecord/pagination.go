// Package medicalrecord owns the medicalrecord domain's HTTP, application logic, and
// persistence code (BE9-2A target:medicalrecord, ADR-006). This is the BE9-2C first slice —
// the out-dom=0 master-CRUD vertical (diagnosis type/name, examination type, chief
// complaint type) from boundary map §3.7's sub-batch ①. Higher-risk paths
// (finalize-lock row protection, treatment/vital/clinical_plan, lab import saga,
// hospitalization/discharge-with-billing) live in this same medicalrecord package.
package medicalrecord

import "gorm.io/gorm"

// paginate converts a 1-origin page/limit pair into a GORM Offset/Limit scope. Consolidated
// here from the (byte-identical) local copies previously duplicated across diagnosis type
// and diagnosis name repositories. Reused by every medicalrecord repository that needs page/limit.
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}
