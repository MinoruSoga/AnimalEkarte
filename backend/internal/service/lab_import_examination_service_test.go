package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

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

func (r *stubExamRepo) ReplaceItemsByExamID(_ context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, error) {
	if r.replaceErr != nil {
		return nil, r.replaceErr
	}
	exam, ok := r.exams[examID]
	if !ok || exam.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("exam", "")
	}
	saved := make([]model.ExamResult, len(items))
	for i := range items {
		saved[i] = items[i]
		saved[i].ExamID = examID
	}
	r.results[examID] = saved
	return saved, nil
}

func (r *stubExamRepo) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
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

func (r *stubExamRepo) Delete(_ context.Context, _, _ uint64) error { return nil }

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

// stubDupChecker は LabImportDuplicateChecker のインメモリ stub。
type stubDupChecker struct {
	isDup    bool
	checkErr error
}

func (c *stubDupChecker) IsDuplicate(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (bool, error) {
	return c.isDup, c.checkErr
}

// ------------------------------------
// Contract tests
// ------------------------------------

// TestLabImportExaminationService_PersistExam_Happy は
// 正常パスで exam + exam_results が永続化され、結果サマリが正しいことを検証する。
func TestLabImportExaminationService_PersistExam_Happy(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{isDup: false}
	svc := NewLabImportExaminationService(examRepo, dupChecker)

	jobID := uuid.New()
	petID := uint64(42)
	mrID := uint64(100)
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

	res, err := svc.PersistExam(context.Background(), input)
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

// TestLabImportExaminationService_PersistExam_NoItems は
// items が空でも exam header だけ正常に保存されることを確認する。
func TestLabImportExaminationService_PersistExam_NoItems(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{}
	svc := NewLabImportExaminationService(examRepo, dupChecker)

	input := LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
		Items:      []LabExamItemInput{},
	}

	res, err := svc.PersistExam(context.Background(), input)
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
	svc := NewLabImportExaminationService(examRepo, dupChecker)

	input := LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
		Items: []LabExamItemInput{
			{Name: "BUN", InspectionValue: "10"},
		},
	}

	res, err := svc.PersistExam(context.Background(), input)
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

// TestLabImportExaminationService_PersistExam_ClinicScopeEnforced は
// 異なる clinic_id では exam が取得できないことを確認する（FK / tenant scope）。
func TestLabImportExaminationService_PersistExam_ClinicScopeEnforced(t *testing.T) {
	examRepo := newStubExamRepo()
	dupChecker := &stubDupChecker{}
	svc := NewLabImportExaminationService(examRepo, dupChecker)

	input := LabExamPersistInput{
		ClinicID:   1,
		ExamTypeID: 3,
		Date:       time.Now(),
		JobID:      uuid.New(),
	}
	res, err := svc.PersistExam(context.Background(), input)
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
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{})

	_, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{})

	_, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{})

	_, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(newStubExamRepo(), dupChecker)

	_, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{})

	_, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{})

	_, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(examRepo, dupChecker)

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
	dupChecker := &dynamicDupChecker{fn: func(_ context.Context, _ uint64, _ uint64, _ time.Time, _ *uint64) (bool, error) {
		callCount++
		// 2 行目だけ重複
		return callCount == 2, nil
	}}
	svc := NewLabImportExaminationService(examRepo, dupChecker)

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
	svc := NewLabImportExaminationService(newStubExamRepo(), &stubDupChecker{})
	jobID := uuid.New()

	res, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{})

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
	svc := NewLabImportExaminationService(examRepo, &stubDupChecker{isDup: false})

	res, err := svc.PersistExam(context.Background(), LabExamPersistInput{
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

// ------------------------------------
// dynamicDupChecker — callCount ベースの動的 stub
// ------------------------------------

type dynamicDupChecker struct {
	fn func(ctx context.Context, clinicID, examTypeID uint64, date time.Time, petID *uint64) (bool, error)
}

func (c *dynamicDupChecker) IsDuplicate(ctx context.Context, clinicID, examTypeID uint64, date time.Time, petID *uint64) (bool, error) {
	return c.fn(ctx, clinicID, examTypeID, date, petID)
}
