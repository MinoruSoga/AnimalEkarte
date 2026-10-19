package csvimport

type cutoverTargetUniqueRule struct {
	table         string
	column        string
	query         string
	nonemptyOnly  bool
	activeOnly    bool
	tentativeOnly bool
}

const cutoverMerchandiseNameQuery = `SELECT EXISTS (SELECT 1 FROM public.merchandise_items WHERE clinic_id = $1 AND is_active = true AND deleted_at IS NULL AND name = ANY($2::text[]))`

const cutoverOwnerEmailQuery = `SELECT EXISTS (SELECT 1 FROM public.owners WHERE clinic_id = $1 AND deleted_at IS NULL AND email IS NOT NULL AND email != '' AND email = ANY($2::text[]))`

const cutoverOwnerPhoneQuery = `SELECT EXISTS (SELECT 1 FROM public.owners WHERE clinic_id = $1 AND deleted_at IS NULL AND phone <> '' AND phone = ANY($2::text[]))`

const cutoverMedicalRecordNumberQuery = `SELECT EXISTS (SELECT 1 FROM public.medical_records WHERE clinic_id = $1 AND record_no = ANY($2::text[]))`

const cutoverEstimateNumberQuery = `SELECT EXISTS (SELECT 1 FROM public.estimates WHERE clinic_id = $1 AND estimate_no = ANY($2::text[]))`

const cutoverMedicalRecordAppointmentQuery = `SELECT EXISTS (SELECT 1 FROM public.medical_records WHERE clinic_id = $1 AND deleted_at IS NULL AND appointment_id IS NOT NULL AND appointment_id = ANY($2::text[]::bigint[]))`

const cutoverInquiryRecordQuery = `SELECT EXISTS (SELECT 1 FROM public.inquiries AS child JOIN public.medical_records AS parent ON parent.id = child.medical_record_id WHERE parent.clinic_id = $1 AND child.medical_record_id = ANY($2::text[]::bigint[]))`

const cutoverClinicalPlanRecordQuery = `SELECT EXISTS (SELECT 1 FROM public.clinical_plans AS child JOIN public.medical_records AS parent ON parent.id = child.medical_record_id WHERE parent.clinic_id = $1 AND child.medical_record_id = ANY($2::text[]::bigint[]))`

const cutoverTrimmingAppointmentQuery = `SELECT EXISTS (SELECT 1 FROM public.appointment_trimming_details AS child JOIN public.appointments AS parent ON parent.id = child.appointment_id WHERE parent.clinic_id = $1 AND child.appointment_id = ANY($2::text[]::bigint[]))`

const cutoverBillingRecordQuery = `SELECT EXISTS (SELECT 1 FROM public.billings AS child JOIN public.medical_records AS parent ON parent.id = child.medical_record_id WHERE parent.clinic_id = $1 AND child.medical_record_id IS NOT NULL AND child.medical_record_id = ANY($2::text[]::bigint[]))`

const cutoverPaymentBillingQuery = `SELECT EXISTS (SELECT 1 FROM public.payments AS child JOIN public.billings AS parent ON parent.id = child.billing_id WHERE parent.clinic_id = $1 AND child.billing_id = ANY($2::text[]::bigint[]))`

const cutoverTentativePetQuery = `SELECT EXISTS (SELECT 1 FROM public.appointments WHERE clinic_id = $1 AND status = 'tentative' AND deleted_at IS NULL AND pet_id = ANY($2::text[]::bigint[]))`

var cutoverTargetUniqueRules = []cutoverTargetUniqueRule{
	{table: "procedures", column: "name", query: cutoverProcedureNameQuery},
	{table: "vaccines", column: "name", query: cutoverVaccineNameQuery},
	{table: "merchandise_items", column: "name", query: cutoverMerchandiseNameQuery, activeOnly: true},
	{table: "owners", column: "email", query: cutoverOwnerEmailQuery, nonemptyOnly: true},
	{table: "owners", column: "phone", query: cutoverOwnerPhoneQuery, nonemptyOnly: true},
	{table: "medical_records", column: "record_no", query: cutoverMedicalRecordNumberQuery},
	{table: "estimates", column: "estimate_no", query: cutoverEstimateNumberQuery},
	{table: "medical_records", column: "appointment_id", query: cutoverMedicalRecordAppointmentQuery, nonemptyOnly: true},
	{table: "inquiries", column: "medical_record_id", query: cutoverInquiryRecordQuery, nonemptyOnly: true},
	{table: "clinical_plans", column: "medical_record_id", query: cutoverClinicalPlanRecordQuery, nonemptyOnly: true},
	{table: "appointment_trimming_details", column: "appointment_id", query: cutoverTrimmingAppointmentQuery, nonemptyOnly: true},
	{table: "billings", column: "medical_record_id", query: cutoverBillingRecordQuery, nonemptyOnly: true},
	{table: "payments", column: "billing_id", query: cutoverPaymentBillingQuery, nonemptyOnly: true},
	{table: "appointments", column: "pet_id", query: cutoverTentativePetQuery, nonemptyOnly: true, tentativeOnly: true},
}
