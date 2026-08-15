package main

import (
	"context"
	"flag"
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
	var (
		clinicID      = flag.Uint64("clinic-id", 0, "対象クリニックID（必須）")
		dryRun        = flag.Bool("dry-run", false, "ドライラン（DB 書き込みなし）")
		batchSize     = flag.Int("batch-size", 5, "並列実行数")
		rateLimitPerS = flag.Int("rate-limit-per-sec", 10, "1秒あたりの最大同期数")
		skipTier2     = flag.Bool("skip-tier-2", false, "Tier2 同期スキップ")
		ownerIDsFlag  = flag.String("owner-ids", "", "対象 ownerID のカンマ区切りリスト（省略時は全員）")
		resumeFrom    = flag.Uint64("resume-from", 0, "このID以降の飼い主のみ処理")
		reportPath    = flag.String("report", "lstep_migration_report.csv", "CSVレポート出力先")
	)
	flag.Parse()

	if *clinicID == 0 {
		fmt.Fprintln(os.Stderr, "error: --clinic-id は必須です")
		flag.Usage()
		return 1
	}
	if *batchSize < 1 {
		fmt.Fprintln(os.Stderr, "error: --batch-size は1以上を指定してください")
		return 1
	}
	if *rateLimitPerS < 1 {
		fmt.Fprintln(os.Stderr, "error: --rate-limit-per-sec は1以上を指定してください")
		return 1
	}

	ownerIDs, err := parseOwnerIDs(*ownerIDsFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: --owner-ids の解析に失敗しました: %v\n", err)
		return 1
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
		ClinicID:        *clinicID,
		DryRun:          *dryRun,
		BatchSize:       *batchSize,
		RateLimitPerSec: *rateLimitPerS,
		SkipTier2:       *skipTier2,
		OwnerIDs:        ownerIDs,
		ResumeFrom:      *resumeFrom,
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

	f, err := os.Create(*reportPath)
	if err != nil {
		log.Error("failed to create report file", slog.String("path", *reportPath), slog.String("error", err.Error()))
		return 1
	}
	defer func() { _ = f.Close() }()

	if err := WriteCSVReport(f, records, log); err != nil {
		log.Error("failed to write CSV report", slog.String("error", err.Error()))
		return 1
	}
	log.Info("report written", slog.String("path", *reportPath))
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
