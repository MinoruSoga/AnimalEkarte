package medicalrecord

import (
	"context"
	"errors"
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
	medicalRecordImageUploadStaffByteBudget     = 50 * 1024 * 1024  // 50 MiB / 60s
	medicalRecordImageUploadClinicByteBudget    = 150 * 1024 * 1024 // 150 MiB / 60s
	medicalRecordImageUploadRateWindow          = time.Minute
	medicalRecordImageUploadInFlightStale       = 5 * time.Minute
)

var (
	errMedicalRecordImageUploadConcurrency = errors.New("too many concurrent medical record image uploads")
	errMedicalRecordImageUploadRateLimit   = errors.New("medical record image upload rate limit exceeded")
	errMedicalRecordImageUploadByteBudget  = errors.New("medical record image upload byte budget exceeded")
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
	// RED stub: always allow until GREEN implements real checks.
	// (intentionally empty enforcement)
}

func newMemoryMedicalRecordImageUploadQuotaStore() *memoryMedicalRecordImageUploadQuotaStore {
	return &memoryMedicalRecordImageUploadQuotaStore{}
}

func (s *memoryMedicalRecordImageUploadQuotaStore) Acquire(
	_ context.Context,
	clinicID, staffID uint64,
	declaredBytes int64,
) (func(context.Context), error) {
	// RED: no enforcement — always succeeds so quota tests fail until GREEN.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	id := s.nextID
	s.leases = append(s.leases, memoryQuotaLease{
		id:            id,
		clinicID:      clinicID,
		staffID:       staffID,
		declaredBytes: declaredBytes,
		acquiredAt:    time.Now(),
	})
	return func(context.Context) {
		s.mu.Lock()
		defer s.mu.Unlock()
		now := time.Now()
		for i := range s.leases {
			if s.leases[i].id == id && s.leases[i].releasedAt == nil {
				s.leases[i].releasedAt = &now
				return
			}
		}
	}, nil
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

func (s *postgresMedicalRecordImageUploadQuotaStore) Acquire(
	context.Context, uint64, uint64, int64,
) (func(context.Context), error) {
	// RED stub — fail closed until GREEN implements Postgres lease logic.
	return nil, errMedicalRecordImageUploadQuotaUnavailable
}
