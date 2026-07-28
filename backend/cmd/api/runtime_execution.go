package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/logger"
)

type runtimeExecution struct {
	runner      serverRunner
	signalCtx   context.Context
	stopSignals context.CancelFunc
	port        string
}

func (e runtimeExecution) run() error {
	if e.stopSignals != nil {
		defer e.stopSignals()
	}
	logger.Info("server starting", slog.String("port", e.port))
	return e.runner.run(e.signalCtx)
}

func openRuntimeDatabase(
	cfg *config.Config,
) (*gorm.DB, *sql.DB, error) {
	db, err := dbconn.OpenGORM(cfg)
	if err != nil {
		logger.Error(
			"failed to connect to database",
			slog.String("error", err.Error()),
		)
		return nil, nil, err
	}
	logger.Info("database connected")
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(
			"failed to access database connection pool",
			slog.String("error", err.Error()),
		)
		return nil, nil, err
	}
	return db, sqlDB, nil
}

func closeRuntimeDatabase(sqlDB *sql.DB) {
	if sqlDB == nil {
		return
	}
	if err := sqlDB.Close(); err != nil {
		logger.Error(
			"failed to close database connection pool",
			slog.String("error", err.Error()),
		)
	}
}

type runtimeComponents struct {
	composition runtimeComposition
	uploader    infra.FileUploader
}

func initializeRuntimeComponents(
	cfg *config.Config,
	db *gorm.DB,
) (runtimeComponents, error) {
	integrationCipher, err := initializeIntegrationCipher(cfg)
	if err != nil {
		return runtimeComponents{}, err
	}
	sharedStorage, err := initializeSharedStorage(cfg)
	if err != nil {
		return runtimeComponents{}, err
	}
	uploader, err := initializeUploader(cfg)
	if err != nil {
		return runtimeComponents{}, err
	}
	return runtimeComponents{
		composition: newRuntimeComposition(runtimeCompositionDependencies{
			Config:            cfg,
			DB:                db,
			IntegrationCipher: integrationCipher,
			SharedStorage:     sharedStorage,
		}),
		uploader: uploader,
	}, nil
}

func prepareRuntimeExecution(
	cfg *config.Config,
	db *gorm.DB,
	sqlDB *sql.DB,
) (runtimeExecution, error) {
	components, err := initializeRuntimeComponents(cfg, db)
	if err != nil {
		return runtimeExecution{}, err
	}
	appCtx, appCancel := context.WithCancel(context.Background())
	router, err := newRouter(cfg)
	if err != nil {
		appCancel()
		return runtimeExecution{}, err
	}
	if err := components.composition.registerRoutes(
		appCtx,
		router,
		components.uploader,
		cfg.GinMode == "release",
	); err != nil {
		appCancel()
		logger.Error(
			"failed to register application routes",
			slog.String("error", err.Error()),
		)
		return runtimeExecution{}, err
	}
	server := newHTTPServer(cfg, router)
	signalCtx, stopSignals := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	return runtimeExecution{
		runner: serverRunner{
			server:           server,
			address:          server.Addr,
			backgroundCtx:    appCtx,
			cancelBackground: appCancel,
			shutdownTimeout:  130 * time.Second,
			drainers:         runtimeDrainers(components.composition),
			resources:        []resourceCloser{sqlDB},
		},
		signalCtx:   signalCtx,
		stopSignals: stopSignals,
		port:        cfg.Port,
	}, nil
}
