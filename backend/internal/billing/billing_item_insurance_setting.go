package billing

// insuranceApplicableFromMasterSetting maps an explicit master setting into a
// pre-adoption billing candidate. Legacy NULL remains unknown; it must not be
// replaced with a new-item default or an insurance-applicable value.
func insuranceApplicableFromMasterSetting(nonInsurance *bool) *bool {
	if nonInsurance == nil {
		return nil
	}
	applicable := !*nonInsurance
	return &applicable
}
