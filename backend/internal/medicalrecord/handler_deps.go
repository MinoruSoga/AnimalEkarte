package medicalrecord

import (
	"context"
	"io"

	"github.com/animal-ekarte/backend/internal/model"
)

// This file declares the vital / medical-record-image handlers' consumer-side views of
// dependencies that live outside this package, following the same "define interfaces where they
// are consumed" rule as service_deps.go. The composition root (cmd/api/main.go) supplies the
// concrete implementations by structural typing.

// medicalRecordGetter is the vital / image handlers' view of the medical-record service's
// tenant-ownership check (GetByID). It is the faithful port of the pre-move
// internal/handler.Handler.verifyMedicalRecordOwnership, which called svc.MedicalRecord.GetByID
// (clinic-scoped fetch + error log/wrap). The composition root injects the concrete
// service.MedicalRecordService, so the ownership behaviour — including the error-path log — stays
// byte-for-byte identical to the pre-move path. clinical_plan needs no getter: its handler never
// verified ownership (it delegates the clinic scope to the service).
type medicalRecordGetter interface {
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// fileUploader is the image handler's view of infra.FileUploader (S3 or local storage). Declaring
// it here keeps internal/medicalrecord free of an internal/infra import; the composition root
// passes the concrete infra.FileUploader in by structural typing.
type fileUploader interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
}
