package model

// type_aliases.go — 命名移行期間中の後方互換エイリアス。
// 完全移行後に削除すること。

// BillingConfirmation は BillingReview の別名（billing_review_repository.go が使用）。
type BillingConfirmation = BillingReview

// MedicalRecordImage は RecordImage の別名（record_image_repository.go が使用）。
type MedicalRecordImage = RecordImage

// Appointment は ReservationAppointment の別名（repository / service / handler 層が使用）。
type Appointment = ReservationAppointment
