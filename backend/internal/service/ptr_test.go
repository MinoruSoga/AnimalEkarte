package service

// ptr is a generic *T literal builder. It previously lived in lab_import_service.go; BE9-2D
// sub-batch③ moved the lab services to internal/medicalrecord, so this test-only residual copy
// keeps the remaining consumers (accounting_service_test.go, cash_register_service_test.go)
// compiling. Only *_test.go files in this package use it.
func ptr[T any](v T) *T { return &v }
