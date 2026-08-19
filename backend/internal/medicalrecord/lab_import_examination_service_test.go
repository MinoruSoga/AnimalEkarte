package medicalrecord

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestBuildExamResults_InvalidNumericBoundsRemainUnassessed(t *testing.T) {
	numericNaN := math.NaN()
	invertedMinimum, invertedMaximum := 10.0, 1.0
	tests := []struct {
		name   string
		refMin *float64
		refMax *float64
	}{
		{name: "NaN minimum", refMin: &numericNaN},
		{name: "inverted bounds", refMin: &invertedMinimum, refMax: &invertedMaximum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := buildExamResults(1, []LabExamItemInput{{
				Name:            "BUN",
				InspectionValue: "5",
				RefMin:          tt.refMin,
				RefMax:          tt.refMax,
			}})

			if len(results) != 1 {
				t.Fatalf("buildExamResults() length = %d, want 1", len(results))
			}
			if results[0].Status != model.ExaminationResultStatusNormal {
				t.Errorf("Status = %q, want %q", results[0].Status, model.ExaminationResultStatusNormal)
			}
			if results[0].IsAbnormal {
				t.Error("IsAbnormal = true, want false")
			}
			if toExamResultResponse(&results[0]).IsAssessed {
				t.Error("IsAssessed = true, want false")
			}
		})
	}
}

// ------------------------------------
// Stubs — no DB required
// ------------------------------------

// stubExamRepo は ExaminationRepository のインメモリ stub。
// exam_id は自動インクリメント。
// createFn が設定されている場合は createErr よりも優先される。
type stubExamRepo struct {
	exams      map[uint64]*model.Examination
	results    map[uint64][]model.ExamResult // keyed by examID
	nextID     uint64
	createErr  error
	replaceErr error
	deleteErr  error
	createFn   func(exam *model.Examination) error // nil なら createErr + auto-increment
}

func newStubExamRepo() *stubExamRepo {
	return &stubExamRepo{
		exams:   make(map[uint64]*model.Examination),
		results: make(map[uint64][]model.ExamResult),
		nextID:  1,
	}
}

func (r *stubExamRepo) Create(_ context.Context, exam *model.Examination) error {
	if r.createFn != nil {
		return r.createFn(exam)
	}
	if r.createErr != nil {
		return r.createErr
	}
	exam.ID = r.nextID
	r.nextID++
	cp := *exam
	r.exams[exam.ID] = &cp
	return nil
}

func (r *stubExamRepo) ReplaceItemsByExamID(_ context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
	if r.replaceErr != nil {
		return nil, 0, r.replaceErr
	}
	exam, ok := r.exams[examID]
	if !ok || exam.ClinicID != clinicID {
		return nil, 0, apperrors.WrapNotFound("exam", "")
	}
	deleted := int64(len(r.results[examID]))
	saved := make([]model.ExamResult, len(items))
	for i := range items {
		saved[i] = items[i]
		saved[i].ExamID = examID
	}
	r.results[examID] = saved
	return saved, deleted, nil
}

func (r *stubExamRepo) FindAll(_ context.Context, _ uint64, _, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
	return nil, 0, nil
}

func (r *stubExamRepo) FindByID(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
	exam, ok := r.exams[id]
	if !ok || exam.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("exam", "")
	}
	cp := *exam
	return &cp, nil
}

func (r *stubExamRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
	return nil, nil
}

// Delete は P2-7 (PR #186 review) 回帰テスト用に、実際に r.exams / r.results から
// exam を除去する（GORM soft-delete の deleted_at IS NULL スコープを模す）。
func (r *stubExamRepo) Delete(_ context.Context, clinicID, id uint64) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	exam, ok := r.exams[id]
	if !ok || exam.ClinicID != clinicID {
		return apperrors.WrapNotFound("exam", "")
	}
	delete(r.exams, id)
	delete(r.results, id)
	return nil
}

func (r *stubExamRepo) CountItemsByExamID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (r *stubExamRepo) FindAllItemsByExamID(_ context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	exam, ok := r.exams[examID]
	if !ok || exam.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("exam", "")
	}
	items := r.results[examID]
	out := make([]model.ExamResult, len(items))
	copy(out, items)
	return out, nil
}

func (r *stubExamRepo) FindByJobID(_ context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error) {
	var out []*model.Examination
	for _, e := range r.exams {
		if e.ClinicID == clinicID && e.JobID != nil && *e.JobID == jobID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

// stubDupChecker は LabImportDuplicateChecker のインメモリ stub。
type stubDupChecker struct {
	isDup    bool
	checkErr error
}

func (c *stubDupChecker) IsDuplicate(_ context.Context, _ LabExamPersistInput) (bool, error) {
	return c.isDup, c.checkErr
}


// passthroughLabImportTransactor は unit test 用の WithTx 素通し（ambient tx 不要な stub repo 経路）。
type passthroughLabImportTransactor struct{}

func (passthroughLabImportTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// stubRollbackLabImportTransactor は stubExamRepo の in-memory 状態を snapshot/restore して
// WithTx 失敗時の rollback を模擬する（MRC-05: Create+Replace 原子性の unit 近似）。
type stubRollbackLabImportTransactor struct {
	repo *stubExamRepo
}

func (t *stubRollbackLabImportTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	examsSnap := make(map[uint64]*model.Examination, len(t.repo.exams))
	for id, exam := range t.repo.exams {
		cp := *exam
		examsSnap[id] = &cp
	}
	resultsSnap := make(map[uint64][]model.ExamResult, len(t.repo.results))
	for id, items := range t.repo.results {
		cp := make([]model.ExamResult, len(items))
		copy(cp, items)
		resultsSnap[id] = cp
	}
	nextID := t.repo.nextID
	err := fn(ctx)
	if err != nil {
		t.repo.exams = examsSnap
		t.repo.results = resultsSnap
		t.repo.nextID = nextID
	}
	return err
}

// ------------------------------------
// Contract tests
// ------------------------------------

// matchingMedicalRecordRepo は指定 pet_id を持つ medical_record を返す。
// pet と medical_record の相関 fail-closed を通過させる happy-path 用。
func matchingMedicalRecordRepo(petID uint64) medicalRecordFinder {
	return &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
		p := petID
		return &model.MedicalRecord{ID: id, PetID: &p}, nil
	}}
}

// TestLabImportExaminationService_PersistExam_Happy は
// 正常パスで exam + exam_results が永続化され、結果サマリが正しいことを検証する。
func TestLabImportExaminationService_PersistExam_Happy(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{isDup: false}
	jobID := uuid.New()
	petID := uint64(42)
	mrID := uint64(100)
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), matchingMedicalRecordRepo(petID), passthroughLabImportTransactor{}).(*labImportExaminationService)
	// BUN normal range 8-30 mg/dL (canine synthetic reference)
	bunMin := 8.0
	bunMax := 30.0
	// CRE normal range 0.5-1.8 mg/dL
	creMin := 0.5
	creMax := 1.8
	// ALT normal range 10-100 U/L
	altMin := 10.0
	altMax := 100.0

	input := LabExamPersistInput{
		ClinicID:        1,
		PetID:           &petID,
		MedicalRecordID: &mrID,
		ExamTypeID:      5,
		Date:            time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Machine:         "Fuji DRI-CHEM",
		JobID:           jobID,
		Items: []LabExamItemInput{
			// BUN 12.3 in [8, 30] → normal
			{Name: "BUN", InspectionValue: "12.3", Unit: "mg/dL", RefMin: &bunMin, RefMax: &bunMax, SortOrder: 1},
			// CRE 0.3 < refMin 0.5 → low
			{Name: "CRE", InspectionValue: "0.3", Unit: "mg/dL", RefMin: &creMin, RefMax: &creMax, SortOrder: 2},
			// ALT 150 > refMax 100 → high
			{Name: "ALT", InspectionValue: "150", Unit: "U/L", RefMin: &altMin, RefMax: &altMax, SortOrder: 3},
		},
	}

	res, err := svc.persistExam(context.Background(), input)
	if err != nil {
		t.Fatalf("PersistExam: unexpected error: %v", err)
	}
	if res.Duplicate {
		t.Error("expected Duplicate=false for a fresh exam")
	}
	if res.ExamID == 0 {
		t.Error("expected non-zero ExamID after persist")
	}
	if res.ItemCount != 3 {
		t.Errorf("expected ItemCount=3, got %d", res.ItemCount)
	}
	if res.JobID != jobID {
		t.Error("JobID must be propagated from input to result")
	}

	// exam header persisted with correct clinic scope
	saved, err := examRepo.FindByID(context.Background(), 1, res.ExamID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if saved.ClinicID != 1 {
		t.Errorf("expected clinic_id=1, got %d", saved.ClinicID)
	}
	if *saved.PetID != petID {
		t.Errorf("expected pet_id=%d, got %d", petID, *saved.PetID)
	}
	if *saved.MedicalRecordID != mrID {
		t.Errorf("expected medical_record_id=%d, got %d", mrID, *saved.MedicalRecordID)
	}
	if saved.Status != model.ExaminationStatusResultEntered {
		t.Errorf("expected status=result_entered, got %s", saved.Status)
	}
	// Phase 4B.2: job_id FK must be persisted on the exam row
	if saved.JobID == nil {
		t.Error("expected JobID to be non-nil on saved exam (Phase 4B.2 exams.job_id FK)")
	} else if *saved.JobID != jobID {
		t.Errorf("expected saved.JobID=%s, got %s", jobID, *saved.JobID)
	}

	// exam_results persisted correctly
	items, err := examRepo.FindAllItemsByExamID(context.Background(), 1, res.ExamID)
	if err != nil {
		t.Fatalf("FindAllItemsByExamID: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// BUN: in range → normal/false
	if items[0].Status != model.ExaminationResultStatusNormal {
		t.Errorf("BUN: expected status=normal, got %s", items[0].Status)
	}
	if items[0].IsAbnormal {
		t.Error("BUN: expected is_abnormal=false")
	}

	// CRE: below ref_min=1 → low/true
	if items[1].Status != model.ExaminationResultStatusLow {
		t.Errorf("CRE: expected status=low, got %s", items[1].Status)
	}
	if !items[1].IsAbnormal {
		t.Error("CRE: expected is_abnormal=true")
	}

	// ALT: above ref_max=10 → high/true
	if items[2].Status != model.ExaminationResultStatusHigh {
		t.Errorf("ALT: expected status=high, got %s", items[2].Status)
	}
	if !items[2].IsAbnormal {
		t.Error("ALT: expected is_abnormal=true")
	}

	// all items bound to correct exam
	for _, it := range items {
		if it.ExamID != res.ExamID {
			t.Errorf("item ExamID=%d, expected %d", it.ExamID, res.ExamID)
		}
	}
}

// TestLabImportExaminationService_PersistExam_RejectsPetMedicalRecordMismatch は
// 同一 clinic 内でも pet_id と medical_record.pet_id が不一致なら fail-closed で拒否し、
// exam を永続化しないことを検証する（#249 residual: lab import patient correlation）。
func TestLabImportExaminationService_PersistExam_RejectsPetMedicalRecordMismatch(t *testing.T) {
	const (
		clinicID     = uint64(1)
		requestPetID = uint64(42)
		recordPetID  = uint64(99)
		mrID         = uint64(100)
	)

	t.Run("rejects mismatched pet_id and medical_record.pet_id", func(t *testing.T) {
		examRepo := newStubExamRepo()
		recordPet := recordPetID
		mrRepo := &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, PetID: &recordPet}, nil
		}}
		svc := NewLabImportExaminationService(
			examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), mrRepo, passthroughLabImportTransactor{},
		).(*labImportExaminationService)

		pet := requestPetID
		mr := mrID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:        clinicID,
			PetID:           &pet,
			MedicalRecordID: &mr,
			ExamTypeID:      5,
			Date:            time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			JobID:           uuid.New(),
		})
		if err == nil {
			t.Fatal("expected error when pet_id does not match medical_record.pet_id")
		}
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected NotFound (no existence leak), got: %v", err)
		}
		if out != nil {
			t.Error("expected nil result on pet/medical_record mismatch")
		}
		if len(examRepo.exams) != 0 {
			t.Errorf("exam must not be persisted on mismatch, got %d", len(examRepo.exams))
		}
	})

	t.Run("rejects medical_record with nil pet_id when request supplies pet_id", func(t *testing.T) {
		examRepo := newStubExamRepo()
		// okMedicalRecordRepo returns MedicalRecord without PetID — correlation must fail closed.
		svc := NewLabImportExaminationService(
			examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{},
		).(*labImportExaminationService)

		pet := requestPetID
		mr := mrID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:        clinicID,
			PetID:           &pet,
			MedicalRecordID: &mr,
			ExamTypeID:      5,
			Date:            time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			JobID:           uuid.New(),
		})
		if err == nil {
			t.Fatal("expected error when medical_record has no pet_id but request supplies pet_id")
		}
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected NotFound, got: %v", err)
		}
		if out != nil {
			t.Error("expected nil result")
		}
		if len(examRepo.exams) != 0 {
			t.Errorf("exam must not be persisted, got %d", len(examRepo.exams))
		}
	})

	t.Run("accepts matching pet_id and medical_record.pet_id", func(t *testing.T) {
		examRepo := newStubExamRepo()
		svc := NewLabImportExaminationService(
			examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), matchingMedicalRecordRepo(requestPetID), passthroughLabImportTransactor{},
		).(*labImportExaminationService)

		pet := requestPetID
		mr := mrID
		out, err := svc.persistExam(context.Background(), LabExamPersistInput{
			ClinicID:        clinicID,
			PetID:           &pet,
			MedicalRecordID: &mr,
			ExamTypeID:      5,
			Date:            time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			JobID:           uuid.New(),
		})
		if err != nil {
			t.Fatalf("unexpected error for matching correlation: %v", err)
		}
		if out == nil || out.ExamID == 0 {
			t.Fatal("expected persisted exam for matching pet/medical_record")
		}
		if len(examRepo.exams) != 1 {
			t.Errorf("expected 1 exam, got %d", len(examRepo.exams))
		}
	})
}

// TestLabImportExaminationService_PersistBatch_RejectsPetMedicalRecordMismatch は
// バッチ内の不一致行が RowError になり、他行の処理を中断せず exam を書かないことを検証する。
func TestLabImportExaminationService_PersistBatch_RejectsPetMedicalRecordMismatch(t *testing.T) {
	const (
		clinicID     = uint64(1)
		requestPetID = uint64(42)
		recordPetID  = uint64(99)
	)
	examRepo := newStubExamRepo()
	recordPet := recordPetID
	mrRepo := &mockMedicalRecordRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
		return &model.MedicalRecord{ID: id, ClinicID: clinicID, PetID: &recordPet}, nil
	}}
	svc := NewLabImportExaminationService(
		examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), mrRepo, passthroughLabImportTransactor{},
	)

	pet := requestPetID
	mr := uint64(100)
	okPet := uint64(7)
	results, err := svc.PersistBatch(context.Background(), []LabExamPersistInput{
		{
			ClinicID:        clinicID,
			PetID:           &pet,
			MedicalRecordID: &mr,
			ExamTypeID:      5,
			Date:            time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			JobID:           uuid.New(),
		},
		{
			// medical_record なし — 相関チェック対象外で成功する
			ClinicID:   clinicID,
			PetID:      &okPet,
			ExamTypeID: 6,
			Date:       time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC),
			JobID:      uuid.New(),
		},
	})
	if err != nil {
		t.Fatalf("PersistBatch system error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].RowError == nil {
		t.Fatal("expected RowError on mismatched first row")
	}
	if !apperrors.IsNotFound(results[0].RowError) {
		t.Errorf("expected NotFound RowError, got: %v", results[0].RowError)
	}
	if results[1].RowError != nil {
		t.Errorf("second row without medical_record should succeed, got: %v", results[1].RowError)
	}
	if results[1].ExamID == 0 {
		t.Error("second row should persist an exam")
	}
	if len(examRepo.exams) != 1 {
		t.Errorf("expected only the non-mismatched exam, got %d", len(examRepo.exams))
	}
}

// TestLabImportExaminationService_PersistExam_NoItems は
// items が空でも exam header だけ正常に保存されることを確認する。
func TestLabImportExaminationService_PersistExam_NoItems(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{}
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	input := LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
		Items:      []LabExamItemInput{},
	}

	res, err := svc.persistExam(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExamID == 0 {
		t.Error("expected non-zero ExamID")
	}
	if res.ItemCount != 0 {
		t.Errorf("expected ItemCount=0, got %d", res.ItemCount)
	}
	if res.Duplicate {
		t.Error("expected Duplicate=false")
	}
}

// TestLabImportExaminationService_PersistExam_Duplicate は
// 重複チェックが true を返した場合に exam が作成されず Duplicate=true を返すことを確認する。
func TestLabImportExaminationService_PersistExam_Duplicate(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{isDup: true}
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	input := LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
		Items: []LabExamItemInput{
			{Name: "BUN", InspectionValue: "10"},
		},
	}

	res, err := svc.persistExam(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Duplicate {
		t.Error("expected Duplicate=true when duplicate checker returns true")
	}
	if res.ExamID != 0 {
		t.Errorf("expected ExamID=0 (no exam created), got %d", res.ExamID)
	}
	if len(examRepo.exams) != 0 {
		t.Errorf("expected 0 exams created on duplicate, got %d", len(examRepo.exams))
	}
}

// TestLabImportExaminationService_PersistExam_SameDayDifferentContentNotDuplicate は
// Issue #249 R-3: 同日・同検査種別でも内容が異なれば両方 persist されることを検証する。
func TestLabImportExaminationService_PersistExam_SameDayDifferentContentNotDuplicate(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &examRepoBackedDupChecker{repo: examRepo}
	svc := NewLabImportExaminationService(
		examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{},
	).(*labImportExaminationService)

	petID := uint64(42)
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	base := LabExamPersistInput{
		ClinicID:   1,
		PetID:      &petID,
		ExamTypeID: 5,
		Date:       date,
		Machine:    "Fuji",
	}

	first := base
	first.JobID = uuid.New()
	first.Items = []LabExamItemInput{{Name: "BUN", InspectionValue: "12.0", Unit: "mg/dL", SortOrder: 1}}
	res1, err := svc.persistExam(context.Background(), first)
	if err != nil {
		t.Fatalf("first persist: %v", err)
	}
	if res1.Duplicate || res1.ExamID == 0 {
		t.Fatalf("first: expected new exam, got duplicate=%v examID=%d", res1.Duplicate, res1.ExamID)
	}

	// 同日・同 type・異なる値 → 新規保存
	second := base
	second.JobID = uuid.New()
	second.Items = []LabExamItemInput{{Name: "BUN", InspectionValue: "25.5", Unit: "mg/dL", SortOrder: 1}}
	res2, err := svc.persistExam(context.Background(), second)
	if err != nil {
		t.Fatalf("second persist: %v", err)
	}
	if res2.Duplicate {
		t.Error("same-day different content must NOT be duplicate")
	}
	if res2.ExamID == 0 || res2.ExamID == res1.ExamID {
		t.Errorf("second: expected distinct new exam, got examID=%d (first=%d)", res2.ExamID, res1.ExamID)
	}
	if len(examRepo.exams) != 2 {
		t.Errorf("expected 2 exams persisted, got %d", len(examRepo.exams))
	}
}

// TestLabImportExaminationService_PersistExam_FullIdenticalContentIsDuplicate は
// Issue #249 R-3: 完全同一ペイロードの再インポートのみスキップすることを検証する。
func TestLabImportExaminationService_PersistExam_FullIdenticalContentIsDuplicate(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &examRepoBackedDupChecker{repo: examRepo}
	svc := NewLabImportExaminationService(
		examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{},
	).(*labImportExaminationService)

	petID := uint64(42)
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	items := []LabExamItemInput{{Name: "BUN", InspectionValue: "12.0", Unit: "mg/dL", SortOrder: 1}}
	first := LabExamPersistInput{
		ClinicID: 1, PetID: &petID, ExamTypeID: 5, Date: date, Machine: "Fuji",
		JobID: uuid.New(), Items: items,
	}
	res1, err := svc.persistExam(context.Background(), first)
	if err != nil {
		t.Fatalf("first persist: %v", err)
	}
	if res1.Duplicate {
		t.Fatal("first should not be duplicate")
	}

	// 完全同一（JobID だけ違う）→ skip
	reimport := first
	reimport.JobID = uuid.New()
	res2, err := svc.persistExam(context.Background(), reimport)
	if err != nil {
		t.Fatalf("reimport: %v", err)
	}
	if !res2.Duplicate {
		t.Error("full-identical re-import must be Duplicate=true")
	}
	if res2.ExamID != 0 {
		t.Errorf("duplicate skip must not create exam, got ExamID=%d", res2.ExamID)
	}
	if len(examRepo.exams) != 1 {
		t.Errorf("expected still 1 exam after identical re-import, got %d", len(examRepo.exams))
	}
}

// TestLabImportExaminationService_PersistExam_ClinicScopeEnforced は
// 異なる clinic_id では exam が取得できないことを確認する（FK / tenant scope）。
func TestLabImportExaminationService_PersistExam_ClinicScopeEnforced(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{}
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	input := LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
	}
	res, err := svc.persistExam(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// wrong clinic cannot read the exam
	_, err = examRepo.FindByID(context.Background(), 999, res.ExamID)
	if err == nil {
		t.Error("expected error when accessing exam with wrong clinic_id")
	}
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected not-found error for wrong clinic, got: %v", err)
	}
}

// TestLabImportExaminationService_PersistExam_MissingClinicID は
// clinic_id=0 が InvalidInput を返すことを確認する（ガード）。
func TestLabImportExaminationService_PersistExam_MissingClinicID(t *testing.T) {
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   0,
		ExamTypeID: 3,
		Date:       time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for clinic_id=0")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput, got: %v", err)
	}
}

// TestLabImportExaminationService_PersistExam_MissingExamTypeID は
// exam_type_id=0 が InvalidInput を返すことを確認する。
func TestLabImportExaminationService_PersistExam_MissingExamTypeID(t *testing.T) {
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 0,
		Date:       time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for exam_type_id=0")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput, got: %v", err)
	}
}

// TestLabImportExaminationService_PersistExam_MissingDate は
// ゼロ日付が InvalidInput を返すことを確認する。
func TestLabImportExaminationService_PersistExam_MissingDate(t *testing.T) {
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Time{},
	})
	if err == nil {
		t.Fatal("expected error for zero date")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput, got: %v", err)
	}
}

// TestLabImportExaminationService_PersistExam_DupCheckError は
// 重複チェックがエラーを返した場合にエラーが伝播することを確認する。
func TestLabImportExaminationService_PersistExam_DupCheckError(t *testing.T) {
	dupChecker := &stubDupChecker{checkErr: errors.New("db error")}
	svc := NewLabImportExaminationService(newStubExamRepo(), dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error when duplicate check fails")
	}
}

// TestLabImportExaminationService_PersistExam_CreateRepoError は
// exam 作成がエラーを返した場合にエラーが伝播することを確認する。
func TestLabImportExaminationService_PersistExam_CreateRepoError(t *testing.T) {
	examRepo := newStubExamRepo()
	examRepo.createErr = errors.New("db error")
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error when exam create fails")
	}
}

// TestLabImportExaminationService_PersistExam_ReplaceItemsError は
// item 保存がエラーを返した場合にエラーが伝播することを確認する。
func TestLabImportExaminationService_PersistExam_ReplaceItemsError(t *testing.T) {
	examRepo := newStubExamRepo()
	examRepo.replaceErr = errors.New("db error")
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
		Items:      []LabExamItemInput{{Name: "BUN", InspectionValue: "10"}},
	})
	if err == nil {
		t.Fatal("expected error when replace items fails")
	}
}

// TestLabImportExaminationService_PersistBatch_Happy は
// PersistBatch が複数 exam を順次保存し、ジョブ接続点が全行に保持されることを確認する。
func TestLabImportExaminationService_PersistBatch_Happy(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{}
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	jobID := uuid.New()
	petA := uint64(10)
	petB := uint64(20)

	inputs := []LabExamPersistInput{
		{ClinicID: 1, PetID: &petA, ExamTypeID: 1, Date: time.Now(), JobID: jobID},
		{ClinicID: 1, PetID: &petB, ExamTypeID: 2, Date: time.Now(), JobID: jobID},
	}

	results, err := svc.PersistBatch(context.Background(), inputs)
	if err != nil {
		t.Fatalf("PersistBatch: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, res := range results {
		if res.Duplicate {
			t.Errorf("result[%d]: unexpected Duplicate=true", i)
		}
		if res.ExamID == 0 {
			t.Errorf("result[%d]: expected non-zero ExamID", i)
		}
		if res.JobID != jobID {
			t.Errorf("result[%d]: JobID not propagated", i)
		}
	}
	// distinct ExamIDs
	if results[0].ExamID == results[1].ExamID {
		t.Error("expected distinct ExamIDs for different batch rows")
	}
}

// TestLabImportExaminationService_PersistBatch_WithDuplicate は
// バッチ内の重複行がスキップされ、非重複行は保存されることを確認する。
func TestLabImportExaminationService_PersistBatch_WithDuplicate(t *testing.T) {
	examRepo := newStubExamRepo()
	callCount := 0
	dupChecker := &dynamicDupChecker{fn: func(_ context.Context, _ LabExamPersistInput) (bool, error) {
		callCount++
		// 2 行目だけ重複
		return callCount == 2, nil
	}}
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	jobID := uuid.New()
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 1, Date: time.Now(), JobID: jobID},
		{ClinicID: 1, ExamTypeID: 2, Date: time.Now(), JobID: jobID},
	}

	results, err := svc.PersistBatch(context.Background(), inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Duplicate {
		t.Error("row 0: expected Duplicate=false")
	}
	if !results[1].Duplicate {
		t.Error("row 1: expected Duplicate=true")
	}
	if len(examRepo.exams) != 1 {
		t.Errorf("expected exactly 1 exam created (1 duplicate skipped), got %d", len(examRepo.exams))
	}
}

// TestLabImportExaminationService_JobID_Propagated は
// JobID フィールドが lab_import_jobs との接続点として全結果に保持されることを確認する。
func TestLabImportExaminationService_JobID_Propagated(t *testing.T) {
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)
	jobID := uuid.New()

	res, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      jobID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.JobID != jobID {
		t.Errorf("expected JobID=%s, got %s", jobID, res.JobID)
	}
}

// TestLabImportExaminationService_NoExternalIO はサービスが外部ネットワーク・認証情報を
// 参照しないことを確認する（compile-time: インターフェース経由のみで完結）。
// この test は構造的に no-op だが、外部依存がないことをドキュメント化する。
func TestLabImportExaminationService_NoExternalIO(t *testing.T) {
	// LabImportExaminationService は ExaminationRepository と LabImportDuplicateChecker の
	// インターフェースにのみ依存する。ネットワーク I/O・credentials・環境変数への参照がなく、
	// 外部 Dr.Wan デバイスへのアクセスも持たない。
	// Phase BLOCKED: Dr.Wan MDB 接続は外部スキーマ確認後のみ許可。
	t.Log("no external IO: LabImportExaminationService depends only on injected repository interfaces")
}

// TestLabImportExaminationService_PersistBatch_RowErrorContinues は
// バッチ内の 1 行が失敗しても残りの行が処理されることを確認する。
// 失敗行は RowError に記録され、成功行は ExamID が設定される。
func TestLabImportExaminationService_PersistBatch_RowErrorContinues(t *testing.T) {
	callCount := 0
	examRepo := newStubExamRepo()
	// 2 行目だけ create エラー
	examRepo.createFn = func(exam *model.Examination) error {
		callCount++
		if callCount == 2 {
			return errors.New("db error on row 2")
		}
		exam.ID = uint64(callCount)
		cp := *exam
		examRepo.exams[exam.ID] = &cp
		return nil
	}
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	jobID := uuid.New()
	inputs := []LabExamPersistInput{
		{ClinicID: 1, ExamTypeID: 1, Date: time.Now(), JobID: jobID},
		{ClinicID: 1, ExamTypeID: 2, Date: time.Now(), JobID: jobID},
		{ClinicID: 1, ExamTypeID: 3, Date: time.Now(), JobID: jobID},
	}

	results, err := svc.PersistBatch(context.Background(), inputs)
	if err != nil {
		t.Fatalf("PersistBatch should not return function-level error for per-row failure: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results (all rows including failed), got %d", len(results))
	}
	if results[0].RowError != nil {
		t.Errorf("row 0: expected RowError=nil, got %v", results[0].RowError)
	}
	if results[1].RowError == nil {
		t.Error("row 1: expected RowError != nil for failed row")
	}
	if results[2].RowError != nil {
		t.Errorf("row 2: expected RowError=nil (processing continued after row 1 failure), got %v", results[2].RowError)
	}
}

// TestLabImportExaminationService_PersistExam_DBDuplicateTreatedAsDuplicate は
// examRepo.Create が AlreadyExists を返した場合に Duplicate=true として扱うことを確認する。
// これは TOCTOU 安全ネット（IsDuplicate の競合ウィンドウで発生した重複）。
func TestLabImportExaminationService_PersistExam_DBDuplicateTreatedAsDuplicate(t *testing.T) {
	examRepo := newStubExamRepo()
	examRepo.createErr = apperrors.WrapAlreadyExists("exam", "")
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{isDup: false}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), passthroughLabImportTransactor{}).(*labImportExaminationService)

	res, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
	})
	if err != nil {
		t.Fatalf("expected no function error for DB duplicate, got: %v", err)
	}
	if !res.Duplicate {
		t.Error("expected Duplicate=true when DB returns AlreadyExists on Create")
	}
	if res.ExamID != 0 {
		t.Errorf("expected ExamID=0 (no exam created), got %d", res.ExamID)
	}
}

// examRepoBackedDupChecker は full-identical 一致を stubExamRepo.exams / results の
// 現在の状態から動的に判定する。本物の LabImportDuplicateCheckerDB を模し、孤児 exam の
// 削除有無が retry の可否に直結することを検証するために使う（P2-7 / R-3 回帰テスト専用）。
type examRepoBackedDupChecker struct {
	repo *stubExamRepo
}

func (c *examRepoBackedDupChecker) IsDuplicate(_ context.Context, input LabExamPersistInput) (bool, error) {
	normalised := time.Date(input.Date.Year(), input.Date.Month(), input.Date.Day(), 0, 0, 0, 0, time.UTC)
	for _, exam := range c.repo.exams {
		if exam.ClinicID != input.ClinicID || exam.ExamTypeID != input.ExamTypeID {
			continue
		}
		examDate := time.Date(exam.Date.Year(), exam.Date.Month(), exam.Date.Day(), 0, 0, 0, 0, time.UTC)
		if !examDate.Equal(normalised) {
			continue
		}
		if !labImportNullableUint64Equal(exam.PetID, input.PetID) {
			continue
		}
		// Items は stub の results map から供給し、本番と同一の full-match 判定を使う。
		candidate := *exam
		candidate.Items = c.repo.results[exam.ID]
		if labImportExamFullyMatches(&candidate, input) {
			return true, nil
		}
	}
	return false, nil
}

// TestLabImportExaminationService_PersistExam_OrphanExamRolledBackOnReplaceItemsError は
// MRC-05 / X-06: exam Create と ReplaceItems を同一 WithTx に収めた結果、
// Replace 失敗時に exam が rollback され（孤児が残らない）、同一キーでの retry が可能であることを検証する。
func TestLabImportExaminationService_PersistExam_OrphanExamRolledBackOnReplaceItemsError(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &examRepoBackedDupChecker{repo: examRepo}
	tx := &stubRollbackLabImportTransactor{repo: examRepo}
	svc := NewLabImportExaminationService(examRepo, dupChecker, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), tx).(*labImportExaminationService)

	petID := uint64(42)
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	baseInput := LabExamPersistInput{
		ClinicID:   1,
		PetID:      &petID,
		ExamTypeID: 3,
		Date:       date,
		Items:      []LabExamItemInput{{Name: "BUN", InspectionValue: "10"}},
	}

	// 1 回目: exam 作成成功 → item 保存失敗 → tx rollback
	examRepo.replaceErr = errors.New("db error on item insert")
	firstInput := baseInput
	firstInput.JobID = uuid.New()
	if _, err := svc.persistExam(context.Background(), firstInput); err == nil {
		t.Fatal("expected error when replace items fails")
	}

	// (a) rollback により孤児 exam が残らない
	if len(examRepo.exams) != 0 {
		t.Fatalf("expected orphan exam to be rolled back after item save failure, got %d exams", len(examRepo.exams))
	}

	// (b) 同一 clinic/type/date/pet の retry は duplicate として誤スキップされない
	examRepo.replaceErr = nil
	retryInput := baseInput
	retryInput.JobID = uuid.New()
	res, err := svc.persistExam(context.Background(), retryInput)
	if err != nil {
		t.Fatalf("retry: unexpected error: %v", err)
	}
	if res.Duplicate {
		t.Error("retry: expected Duplicate=false — transaction rollback must allow retry to proceed")
	}
	if res.ExamID == 0 {
		t.Error("retry: expected exam to be created on retry")
	}
}

// TestLabImportExaminationService_PersistExam_ReplaceItemsErrorPropagates は
// item 保存失敗が呼び出し元へ伝播し、補償 Delete 経路に依存しないことを確認する。
func TestLabImportExaminationService_PersistExam_ReplaceItemsErrorPropagates(t *testing.T) {
	examRepo := newStubExamRepo()
	examRepo.replaceErr = errors.New("db error on item insert")
	tx := &stubRollbackLabImportTransactor{repo: examRepo}
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{}, okExamTypeRepo(), okPetRepo(), okMedicalRecordRepo(), tx).(*labImportExaminationService)

	_, err := svc.persistExam(context.Background(), LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
		Items:      []LabExamItemInput{{Name: "BUN", InspectionValue: "10"}},
	})
	if err == nil {
		t.Fatal("expected error when replace items fails")
	}
	if !strings.Contains(err.Error(), "failed to save exam items") {
		t.Errorf("expected item-save error to propagate, got: %v", err)
	}
	if len(examRepo.exams) != 0 {
		t.Fatalf("expected no orphan exam after replace failure, got %d", len(examRepo.exams))
	}
}

// ------------------------------------
// dynamicDupChecker — callCount ベースの動的 stub
// ------------------------------------

type dynamicDupChecker struct {
	fn func(ctx context.Context, input LabExamPersistInput) (bool, error)
}

func (c *dynamicDupChecker) IsDuplicate(ctx context.Context, input LabExamPersistInput) (bool, error) {
	return c.fn(ctx, input)
}
