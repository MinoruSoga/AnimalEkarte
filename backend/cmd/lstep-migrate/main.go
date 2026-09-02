package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	appCrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/pet"
)

func main() {
	os.Exit(run())
}

func run() int {
	opts, code := parseMigrateCLI()
	if code != 0 {
		return code
	}

	logger.Init(logger.Config{Level: slog.LevelInfo, Format: "json", Output: os.Stdout})
	log := slog.Default()

	if err := config.ConfigureTimeZone(); err != nil {
		log.Error("timezone configuration failed", slog.String("error", err.Error()))
		return 1
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Error("config validation failed", slog.String("error", err.Error()))
		return 1
	}

	db, err := dbconn.OpenGORM(cfg)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		return 1
	}
	log.Info("database connected")

	petWriter := pet.NewWriter(db)
	ownerRepo := owner.NewRepository(db, pet.NewOwnerRegistrationAdapter(petWriter))
	petRepo := pet.NewRepositoryWithWriter(db, petWriter)

	var cipher *appCrypto.AESGCMCipher
	if cfg.IntegrationEncryptionKey != "" {
		cipher, err = appCrypto.NewAESGCMCipher(cfg.IntegrationEncryptionKey)
		if err != nil {
			log.Error("failed to initialize cipher", slog.String("error", err.Error()))
			return 1
		}
	}

	settingsSvc := lstep.NewLstepSettingsService(
		lstep.NewLstepSettingsRepository(db),
		lstep.NewLstepSyncSettingsRepository(db),
		cipher,
		nil,
		nil,
	)
	tagSyncSvc := lstep.NewLstepTagSyncService(
		settingsSvc,
		ownerRepo,
		medicalrecord.NewVaccinationRepository(db),
		medicalrecord.NewMedicalRecordRepository(db),
		billing.NewAccountingRepository(db),
		lstep.NewLstepTagCacheRepository(db),
		petRepo,
		medicalrecord.NewPrescriptionRepository(db),
		medicalrecord.NewCheckupRepository(db),
		lstep.NewLstepSyncErrorCounterRepository(db),
		lstep.NewLstepTagCodeMappingRepository(db),
		billing.NewBillingItemRepository(db),
		lstep.NewLstepTagConfigRepository(db),
	)

	migCfg := Config{
		ClinicID:        opts.clinicID,
		DryRun:          opts.dryRun,
		BatchSize:       opts.batchSize,
		RateLimitPerSec: opts.rateLimitPerS,
		SkipTier2:       opts.skipTier2,
		OwnerIDs:        opts.ownerIDs,
		ResumeFrom:      opts.resumeFrom,
	}

	m := NewMigrator(migCfg, db, ownerRepo, tagSyncSvc, log)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	records, err := m.Run(ctx)
	if err != nil {
		log.Error("migration failed", slog.String("error", err.Error()))
		return 1
	}

	if len(records) == 0 {
		log.Info("no records to report")
		return 0
	}

	f, err := os.Create(opts.reportPath)
	if err != nil {
		log.Error("failed to create report file", slog.String("path", opts.reportPath), slog.String("error", err.Error()))
		return 1
	}
	defer func() { _ = f.Close() }()

	if err := WriteCSVReport(f, records, log); err != nil {
		log.Error("failed to write CSV report", slog.String("error", err.Error()))
		return 1
	}
	log.Info("report written", slog.String("path", opts.reportPath))
	return 0
}

func parseOwnerIDs(s string) ([]uint64, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	ids := make([]uint64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid owner ID %q: %w", p, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
