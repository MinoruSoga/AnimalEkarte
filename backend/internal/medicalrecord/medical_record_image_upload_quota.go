package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Quota constants aligned with FE multi-file caps so legitimate UI uploads succeed.
const (
	medicalRecordImageUploadStaffMaxConcurrent  = 3
	medicalRecordImageUploadClinicMaxConcurrent = 9
	medicalRecordImageUploadStaffRatePerMinute  = 10
	medicalRecordImageUploadClinicRatePerMinute = 30
	medicalRecordImageUploadStaffByteBudget     = int64(50 * 1024 * 1024)  // 50 MiB / 60s
	medicalRecordImageUploadClinicByteBudget    = int64(150 * 1024 * 1024) // 150 MiB / 60s
	medicalRecordImageUploadRateWindow          = time.Minute
	medicalRecordImageUploadInFlightStale       = 5 * time.Minute
)

var (
	errMedicalRecordImageUploadConcurrency      = errors.New("too many concurrent medical record image uploads")
	errMedicalRecordImageUploadRateLimit        = errors.New("medical record image upload rate limit exceeded")
	errMedicalRecordImageUploadByteBudget       = errors.New("medical record image upload byte budget exceeded")
	errMedicalRecordImageUploadQuotaUnavailable = errors.New("upload quota unavailable")
)

// medicalRecordImageUploadQuotaStore reserves concurrency/rate/bytes before heavy multipart work.
// release must be called when the request finishes (success or failure).
type medicalRecordImageUploadQuotaStore interface {
	Acquire(ctx context.Context, clinicID, staffID uint64, declaredBytes int64) (release func(context.Context), err error)
}

// ---------------------------------------------------------------------------
// Shared-memory store (unit tests / single-process authoritative simulation)
// ---------------------------------------------------------------------------

type memoryQuotaLease struct {
	id            int64
	clinicID      uint64
	staffID       uint64
	declaredBytes int64
	acquiredAt    time.Time
	releasedAt    *time.Time
}

type memoryMedicalRecordImageUploadQuotaStore struct {
	mu     sync.Mutex
	nextID int64
	leases []memoryQuotaLease
	now    func() time.Time
}

func newMemoryMedicalRecordImageUploadQuotaStore() *memoryMedicalRecordImageUploadQuotaStore {
	return &memoryMedicalRecordImageUploadQuotaStore{
		now: time.Now,
	}
}

func (s *memoryMedicalRecordImageUploadQuotaStore) Acquire(
	_ context.Context,
	clinicID, staffID uint64,
	declaredBytes int64,
) (func(context.Context), error) {
	if declaredBytes < 0 {
		declaredBytes = 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if err := evaluateMedicalRecordImageUploadQuota(s.snapshotCounts(clinicID, staffID, now), declaredBytes); err != nil {
		return nil, err
	}

	s.nextID++
	id := s.nextID
	s.leases = append(s.leases, memoryQuotaLease{
		id:            id,
		clinicID:      clinicID,
		staffID:       staffID,
		declaredBytes: declaredBytes,
		acquiredAt:    now,
	})
	return func(context.Context) {
		s.mu.Lock()
		defer s.mu.Unlock()
		released := s.now()
		for i := range s.leases {
			if s.leases[i].id == id && s.leases[i].releasedAt == nil {
				s.leases[i].releasedAt = &released
				return
			}
		}
	}, nil
}

type medicalRecordImageUploadQuotaCounts struct {
	staffInFlight  int
	clinicInFlight int
	staffRate      int
	clinicRate     int
	staffBytes     int64
	clinicBytes    int64
}

func (s *memoryMedicalRecordImageUploadQuotaStore) snapshotCounts(
	clinicID, staffID uint64,
	now time.Time,
) medicalRecordImageUploadQuotaCounts {
	staleBefore := now.Add(-medicalRecordImageUploadInFlightStale)
	rateWindowStart := now.Add(-medicalRecordImageUploadRateWindow)
	var counts medicalRecordImageUploadQuotaCounts
	for _, lease := range s.leases {
		if lease.clinicID != clinicID {
			continue
		}
		inRateWindow := !lease.acquiredAt.Before(rateWindowStart)
		if inRateWindow {
			counts.clinicRate++
			counts.clinicBytes += lease.declaredBytes
			if lease.staffID == staffID {
				counts.staffRate++
				counts.staffBytes += lease.declaredBytes
			}
		}
		// In-flight: unreleased and not stale.
		if lease.releasedAt == nil && lease.acquiredAt.After(staleBefore) {
			counts.clinicInFlight++
			if lease.staffID == staffID {
				counts.staffInFlight++
			}
		}
	}
	return counts
}

func evaluateMedicalRecordImageUploadQuota(
	counts medicalRecordImageUploadQuotaCounts,
	declaredBytes int64,
) error {
	if counts.staffInFlight >= medicalRecordImageUploadStaffMaxConcurrent ||
		counts.clinicInFlight >= medicalRecordImageUploadClinicMaxConcurrent {
		return errMedicalRecordImageUploadConcurrency
	}
	if counts.staffRate >= medicalRecordImageUploadStaffRatePerMinute ||
		counts.clinicRate >= medicalRecordImageUploadClinicRatePerMinute {
		return errMedicalRecordImageUploadRateLimit
	}
	if counts.staffBytes+declaredBytes > medicalRecordImageUploadStaffByteBudget ||
		counts.clinicBytes+declaredBytes > medicalRecordImageUploadClinicByteBudget {
		return errMedicalRecordImageUploadByteBudget
	}
	return nil
}

// ---------------------------------------------------------------------------
// Postgres-backed store (multi-process / multi-replica)
// ---------------------------------------------------------------------------

type postgresMedicalRecordImageUploadQuotaStore struct {
	db *gorm.DB
}

// NewPostgresMedicalRecordImageUploadQuotaStore wires the authoritative quota store.
// A nil db is fail-closed (Acquire always errors).
func NewPostgresMedicalRecordImageUploadQuotaStore(db *gorm.DB) medicalRecordImageUploadQuotaStore {
	if db == nil {
		return failClosedMedicalRecordImageUploadQuotaStore{}
	}
	return &postgresMedicalRecordImageUploadQuotaStore{db: db}
}

type failClosedMedicalRecordImageUploadQuotaStore struct{}

func (failClosedMedicalRecordImageUploadQuotaStore) Acquire(
	context.Context, uint64, uint64, int64,
) (func(context.Context), error) {
	return nil, errMedicalRecordImageUploadQuotaUnavailable
}

type medicalRecordImageUploadQuotaRow struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	ClinicID      uint64     `gorm:"column:clinic_id"`
	StaffID       uint64     `gorm:"column:staff_id"`
	DeclaredBytes int64      `gorm:"column:declared_bytes"`
	AcquiredAt    time.Time  `gorm:"column:acquired_at"`
	ReleasedAt    *time.Time `gorm:"column:released_at"`
}

func (medicalRecordImageUploadQuotaRow) TableName() string {
	return "medical_record_image_upload_quota"
}

func (s *postgresMedicalRecordImageUploadQuotaStore) Acquire(
	ctx context.Context,
	clinicID, staffID uint64,
	declaredBytes int64,
) (func(context.Context), error) {
	if s == nil || s.db == nil {
		return nil, errMedicalRecordImageUploadQuotaUnavailable
	}
	if declaredBytes < 0 {
		declaredBytes = 0
	}

	var leaseID int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize quota decisions per clinic across processes/replicas.
		if clinicID > uint64(math.MaxInt64) {
			return fmt.Errorf("%w: clinic_id exceeds advisory lock range", errMedicalRecordImageUploadQuotaUnavailable)
		}
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(clinicID)).Error; err != nil {
			return fmt.Errorf("%w: advisory lock: %w", errMedicalRecordImageUploadQuotaUnavailable, err)
		}

		now := time.Now()
		staleBefore := now.Add(-medicalRecordImageUploadInFlightStale)
		rateWindowStart := now.Add(-medicalRecordImageUploadRateWindow)

		var counts medicalRecordImageUploadQuotaCounts
		// In-flight concurrency (unreleased, non-stale).
		if err := tx.Raw(`
			SELECT
				COUNT(*) FILTER (WHERE staff_id = ?) AS staff_in_flight,
				COUNT(*) AS clinic_in_flight
			FROM medical_record_image_upload_quota
			WHERE clinic_id = ?
			  AND released_at IS NULL
			  AND acquired_at > ?
		`, staffID, clinicID, staleBefore).Row().Scan(&counts.staffInFlight, &counts.clinicInFlight); err != nil {
			return fmt.Errorf("%w: count in-flight: %w", errMedicalRecordImageUploadQuotaUnavailable, err)
		}

		// Rate + byte budget over rolling window (includes released).
		if err := tx.Raw(`
			SELECT
				COUNT(*) FILTER (WHERE staff_id = ?) AS staff_rate,
				COUNT(*) AS clinic_rate,
				COALESCE(SUM(declared_bytes) FILTER (WHERE staff_id = ?), 0) AS staff_bytes,
				COALESCE(SUM(declared_bytes), 0) AS clinic_bytes
			FROM medical_record_image_upload_quota
			WHERE clinic_id = ?
			  AND acquired_at > ?
		`, staffID, staffID, clinicID, rateWindowStart).Row().Scan(
			&counts.staffRate,
			&counts.clinicRate,
			&counts.staffBytes,
			&counts.clinicBytes,
		); err != nil {
			return fmt.Errorf("%w: count rate/bytes: %w", errMedicalRecordImageUploadQuotaUnavailable, err)
		}

		if err := evaluateMedicalRecordImageUploadQuota(counts, declaredBytes); err != nil {
			return err
		}

		row := medicalRecordImageUploadQuotaRow{
			ClinicID:      clinicID,
			StaffID:       staffID,
			DeclaredBytes: declaredBytes,
			AcquiredAt:    now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("%w: insert lease: %w", errMedicalRecordImageUploadQuotaUnavailable, err)
		}
		leaseID = row.ID
		return nil
	})
	if err != nil {
		if errors.Is(err, errMedicalRecordImageUploadConcurrency) ||
			errors.Is(err, errMedicalRecordImageUploadRateLimit) ||
			errors.Is(err, errMedicalRecordImageUploadByteBudget) {
			return nil, err
		}
		// Infrastructure / unexpected → fail-closed.
		if errors.Is(err, errMedicalRecordImageUploadQuotaUnavailable) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %w", errMedicalRecordImageUploadQuotaUnavailable, err)
	}

	return func(releaseCtx context.Context) { //nolint:contextcheck // lease release may run after the request context is canceled
		if releaseCtx == nil {
			releaseCtx = context.Background()
		}
		_ = s.db.WithContext(releaseCtx).
			Model(&medicalRecordImageUploadQuotaRow{}).
			Where("id = ? AND released_at IS NULL", leaseID).
			Update("released_at", time.Now()).Error
	}, nil
}
