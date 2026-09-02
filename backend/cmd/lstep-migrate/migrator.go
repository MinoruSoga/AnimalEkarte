package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
)

// Config はバッチ実行の設定。
type Config struct {
	ClinicID        uint64
	DryRun          bool
	BatchSize       int
	RateLimitPerSec int
	SkipTier2       bool
	OwnerIDs        []uint64
	ResumeFrom      uint64
}

// ProgressRecord は1飼い主の進捗記録。
type ProgressRecord struct {
	OwnerID      uint64
	OwnerName    string
	Status       string
	TagsAdded    int
	TagsFailed   int
	ErrorMessage string
}

// progressRow は lstep_migration_progress テーブルの1行（GORM直接操作用）。
type progressRow struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement"`
	ClinicID     uint64     `gorm:"column:clinic_id;not null"`
	OwnerID      uint64     `gorm:"column:owner_id;not null"`
	Status       string     `gorm:"column:status;not null;default:pending"`
	TagsAdded    int        `gorm:"column:tags_added;not null;default:0"`
	TagsFailed   int        `gorm:"column:tags_failed;not null;default:0"`
	ErrorMessage *string    `gorm:"column:error_message"`
	StartedAt    *time.Time `gorm:"column:started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
}

func (progressRow) TableName() string { return "lstep_migration_progress" }

type migrationTagSync interface {
	SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error
	SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error
	SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error
}

// Migrator はLステップ初回一括同期を実行する。
type Migrator struct {
	cfg     Config
	db      *gorm.DB
	owners  owner.LstepRepository
	tagSync migrationTagSync
	logger  *slog.Logger
	records []ProgressRecord
	mu      sync.Mutex
}

// NewMigrator は Migrator を初期化して返す。
func NewMigrator(
	cfg Config,
	db *gorm.DB,
	owners owner.LstepRepository,
	tagSync migrationTagSync,
	logger *slog.Logger,
) *Migrator {
	return &Migrator{cfg: cfg, db: db, owners: owners, tagSync: tagSync, logger: logger}
}

// Run は一括同期を実行して進捗レコードを返す。
func (m *Migrator) Run(ctx context.Context) ([]ProgressRecord, error) {
	// 対象飼い主を取得
	allOwners, err := m.owners.FindAllWithLineUserID(ctx, m.cfg.ClinicID)
	if err != nil {
		return nil, fmt.Errorf("failed to list owners: %w", err)
	}

	// オプトアウト除外 + ownerIDs / resumeFrom フィルタ
	targets := m.filterOwners(allOwners)
	m.logger.Info("migration target owners",
		slog.Int("total_with_line_id", len(allOwners)),
		slog.Int("filtered", len(targets)),
		slog.Bool("dry_run", m.cfg.DryRun),
	)

	if len(targets) == 0 {
		m.logger.Info("no owners to migrate")
		return nil, nil
	}

	// 進捗テーブルを pending で初期化
	if !m.cfg.DryRun {
		if err := m.initProgressRows(ctx, targets); err != nil {
			return nil, fmt.Errorf("failed to init progress rows: %w", err)
		}
	}

	// レート制限器（オーナー単位）
	limiter := rate.NewLimiter(rate.Limit(m.cfg.RateLimitPerSec), m.cfg.BatchSize)
	// バッチ並列制御
	sem := semaphore.NewWeighted(int64(m.cfg.BatchSize))

	var wg sync.WaitGroup
	var loopErr error
	for i := range targets {
		owner := &targets[i]
		if err := sem.Acquire(ctx, 1); err != nil {
			loopErr = fmt.Errorf("semaphore acquire: %w", err)
			break
		}
		if err := limiter.Wait(ctx); err != nil {
			sem.Release(1)
			loopErr = fmt.Errorf("rate limiter wait: %w", err)
			break
		}
		wg.Go(func() {
			defer sem.Release(1)
			rec := m.processOwner(ctx, owner)
			m.mu.Lock()
			m.records = append(m.records, rec)
			m.mu.Unlock()
		})
	}
	wg.Wait()

	if loopErr != nil {
		return m.records, loopErr
	}
	return m.records, nil
}

// filterOwners はオプトアウト・ownerIDs・resumeFrom のフィルタを適用する。
func (m *Migrator) filterOwners(owners []model.Owner) []model.Owner {
	idSet := make(map[uint64]bool, len(m.cfg.OwnerIDs))
	for _, id := range m.cfg.OwnerIDs {
		idSet[id] = true
	}
	var result []model.Owner
	for i := range owners {
		o := &owners[i]
		if o.LstepOptOut {
			continue
		}
		if m.cfg.ResumeFrom > 0 && o.ID < m.cfg.ResumeFrom {
			continue
		}
		if len(idSet) > 0 && !idSet[o.ID] {
			continue
		}
		result = append(result, *o)
	}
	return result
}

// processOwner は1飼い主のタグ同期を実行して進捗を記録する。
func (m *Migrator) processOwner(ctx context.Context, owner *model.Owner) ProgressRecord {
	rec := ProgressRecord{OwnerID: owner.ID, OwnerName: owner.Name, Status: "success"}
	startedAt := time.Now()

	if !m.cfg.DryRun {
		if err := m.updateProgress(ctx, owner.ID, "pending", 0, 0, "", &startedAt, nil); err != nil {
			rec.Status = "failed"
			rec.ErrorMessage = "progress ledger write failed: " + err.Error()
			return rec
		}
	}

	m.runOwnerSyncMethods(ctx, owner, &rec)

	now := time.Now()
	switch {
	case rec.ErrorMessage == "" && rec.TagsFailed == 0:
		rec.Status = "success"
	case rec.TagsAdded > 0:
		rec.Status = "partial"
	default:
		rec.Status = "failed"
	}

	if m.cfg.DryRun {
		m.logger.Info("[dry-run] result",
			slog.Uint64("owner_id", owner.ID),
			slog.String("name", owner.Name),
			slog.String("status", rec.Status),
		)
		return rec
	}

	m.writeOwnerProgress(ctx, owner, &rec, startedAt, now)
	m.logger.Info("owner synced",
		slog.Uint64("owner_id", owner.ID),
		slog.String("status", rec.Status),
		slog.Int("added", rec.TagsAdded),
		slog.Int("failed", rec.TagsFailed),
	)
	return rec
}

type ownerSyncFn struct {
	name string
	fn   func() error
}

func (m *Migrator) ownerSyncMethods(ctx context.Context, owner *model.Owner) []ownerSyncFn {
	// Tier 1 同期メソッド（clinicID + ownerID のみ必要）
	tier1 := []ownerSyncFn{
		{"SyncOwnerAnimalClassificationTags", func() error {
			return m.tagSync.SyncOwnerAnimalClassificationTags(ctx, m.cfg.ClinicID, owner.ID)
		}},
		{"SyncPetBasicInfoTags", func() error {
			return m.tagSync.SyncPetBasicInfoTags(ctx, m.cfg.ClinicID, owner.ID)
		}},
		{"SyncVisitCompletionTags", func() error {
			return m.tagSync.SyncVisitCompletionTags(ctx, m.cfg.ClinicID, owner.ID)
		}},
		{"SyncCPMStageTag", func() error {
			return m.tagSync.SyncCPMStageTag(ctx, m.cfg.ClinicID, owner.ID)
		}},
	}

	// Tier 2 同期メソッド
	tier2 := []ownerSyncFn{
		{"SyncNextVisitTag", func() error {
			return m.tagSync.SyncNextVisitTag(ctx, m.cfg.ClinicID, owner.ID)
		}},
		{"SyncPrescriptionTag", func() error {
			return m.tagSync.SyncPrescriptionTag(ctx, m.cfg.ClinicID, owner.ID)
		}},
	}

	if m.cfg.SkipTier2 {
		return tier1
	}
	return append(tier1, tier2...)
}

func (m *Migrator) runOwnerSyncMethods(ctx context.Context, owner *model.Owner, rec *ProgressRecord) {
	var errs []string
	for _, s := range m.ownerSyncMethods(ctx, owner) {
		if m.cfg.DryRun {
			m.logger.Info("[dry-run] would call", slog.Uint64("owner_id", owner.ID), slog.String("method", s.name))
			rec.TagsAdded++
			continue
		}
		if err := s.fn(); err != nil {
			m.logger.Warn("sync failed",
				slog.Uint64("owner_id", owner.ID),
				slog.String("method", s.name),
				slog.String("error", err.Error()),
			)
			errs = append(errs, s.name+": "+err.Error())
			rec.TagsFailed++
		} else {
			rec.TagsAdded++
		}
	}
	if len(errs) > 0 {
		rec.ErrorMessage = join(errs, "; ")
	}
}

func (m *Migrator) writeOwnerProgress(
	ctx context.Context,
	owner *model.Owner,
	rec *ProgressRecord,
	startedAt, now time.Time,
) {
	if err := m.updateProgress(ctx, owner.ID, rec.Status, rec.TagsAdded, rec.TagsFailed, rec.ErrorMessage, &startedAt, &now); err != nil {
		// CMD-07: do not report success when the progress ledger diverges from the run result.
		if rec.Status == "success" {
			rec.Status = "partial"
		}
		if rec.ErrorMessage == "" {
			rec.ErrorMessage = "progress ledger write failed: " + err.Error()
		} else {
			rec.ErrorMessage = rec.ErrorMessage + "; progress ledger write failed: " + err.Error()
		}
		m.logger.Error("failed to update progress ledger",
			slog.Uint64("owner_id", owner.ID),
			slog.String("error", err.Error()),
		)
	}
}

func (m *Migrator) initProgressRows(ctx context.Context, owners []model.Owner) error {
	rows := make([]progressRow, 0, len(owners))
	for i := range owners {
		rows = append(rows, progressRow{ClinicID: m.cfg.ClinicID, OwnerID: owners[i].ID, Status: "pending"})
	}
	return m.db.WithContext(ctx).
		Clauses().
		CreateInBatches(rows, 100).Error
}

func (m *Migrator) updateProgress(ctx context.Context, ownerID uint64, status string, added, failed int, errMsg string, startedAt, completedAt *time.Time) error {
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	row := progressRow{
		ClinicID:     m.cfg.ClinicID,
		OwnerID:      ownerID,
		Status:       status,
		TagsAdded:    added,
		TagsFailed:   failed,
		ErrorMessage: errPtr,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
	}
	if err := m.db.WithContext(ctx).
		Where("clinic_id = ? AND owner_id = ?", m.cfg.ClinicID, ownerID).
		Assign(row).
		FirstOrCreate(&row).Error; err != nil {
		return fmt.Errorf("update progress owner_id=%d: %w", ownerID, err)
	}
	return nil
}

func join(parts []string, sep string) string {
	return strings.Join(parts, sep)
}
