package main

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/trimming"
)

type imageDeletionRunner interface {
	Start(ctx context.Context) func()
}

var newTrimmingImageDeletionWorker = func(db *gorm.DB, uploader infra.FileUploader) imageDeletionRunner {
	return trimming.NewImageDeletionWorker(
		trimming.NewImageDeletionRetryStore(db),
		uploader,
	)
}

func startTrimmingImageDeletionWorker(
	ctx context.Context,
	cfg *config.Config,
	db *gorm.DB,
	uploader infra.FileUploader,
) func() {
	if cfg == nil || !cfg.TrimmingImageCleanupEnabled {
		return func() {}
	}
	logger.Info("trimming image deletion retry worker enabled")
	return newTrimmingImageDeletionWorker(db, uploader).Start(ctx)
}
