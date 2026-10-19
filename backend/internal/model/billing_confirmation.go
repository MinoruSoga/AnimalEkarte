package model

// ConfirmationStatus は会計医師確認ステータス
type ConfirmationStatus string

const (
	ConfirmationStatusPending   ConfirmationStatus = "pending"
	ConfirmationStatusConfirmed ConfirmationStatus = "confirmed"
	ConfirmationStatusReturned  ConfirmationStatus = "returned"
)

// BillingConfirmation は会計医師確認（v7.0追加）
