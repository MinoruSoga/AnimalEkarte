package trimming

import (
	"context"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-457: 短縮名・画像・曜日の保存契約。利用医院は既存 clinic_id（T1）。

type courseImageLifecycleFake struct {
	locked     []ImageKey
	deleted    map[string]bool
	inUseKeys  map[string]bool
	tombstoned []string
}

func (f *courseImageLifecycleFake) LockImageKeys(_ context.Context, keys []ImageKey) error {
	f.locked = append(f.locked, keys...)
	return nil
}

func (f *courseImageLifecycleFake) ImageKeyDeleted(_ context.Context, key ImageKey) (bool, error) {
	return f.deleted[key.value], nil
}

func (f *courseImageLifecycleFake) ImageKeyInUse(_ context.Context, _ uint64, key string, _ uint64) (bool, error) {
	return f.inUseKeys[key], nil
}

func (f *courseImageLifecycleFake) MarkImageKeyDeleted(_ context.Context, key ImageKey, _ uint64) error {
	if f.deleted == nil {
		f.deleted = map[string]bool{}
	}
	f.tombstoned = append(f.tombstoned, key.value)
	f.deleted[key.value] = true
	return nil
}

func newSupersedeUpdateRepo(existing *model.TrimmingCourse, captured *map[string]any) *mockTrimmingCourseRepository {
	return &mockTrimmingCourseRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
			*captured = buildTrimmingCourseUpdate(&cmd)
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID}, nil
		},
	}
}

func newAttributeCourseService(repo TrimmingCourseRepository, lifecycle ImageLifecycleRepository) TrimmingCourseService {
	if lifecycle == nil {
		return NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})
	}
	return NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{}, WithCourseImageLifecycle(lifecycle))
}

func captureCreateRepo(captured **model.TrimmingCourse) *mockTrimmingCourseRepository {
	return &mockTrimmingCourseRepository{
		createFn: func(_ context.Context, course *model.TrimmingCourse) error {
			*captured = course
			return nil
		},
	}
}

func TestTrimmingCourseService_CreateNormalizesAttributes(t *testing.T) {
	var got *model.TrimmingCourse
	lifecycle := &courseImageLifecycleFake{}
	svc := newAttributeCourseService(captureCreateRepo(&got), lifecycle)

	_, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{
		Name:            "シャンプー",
		ShortName:       "  シャン  ",
		ImageObjectKey:  " trimming-images/1/course.jpg ",
		AllowedWeekdays: []int{6, 1, 1, 3},
	})

	require.NoError(t, err)
	require.NotNil(t, got.ShortName)
	assert.Equal(t, "シャン", *got.ShortName)
	require.NotNil(t, got.ImageObjectKey)
	assert.Equal(t, "trimming-images/1/course.jpg", *got.ImageObjectKey)
	require.NotNil(t, got.AllowedWeekdays)
	assert.Equal(t, pq.Int64Array{1, 3, 6}, *got.AllowedWeekdays)
	assert.Equal(t, []ImageKey{{clinicID: 1, value: "trimming-images/1/course.jpg"}}, lifecycle.locked)
}

func TestTrimmingCourseService_CreateBlankOptionalAttributesStoreNull(t *testing.T) {
	var got *model.TrimmingCourse
	svc := newAttributeCourseService(captureCreateRepo(&got), nil)

	_, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{
		Name:            "シャンプー",
		ShortName:       "   ",
		ImageObjectKey:  "",
		AllowedWeekdays: []int{},
	})

	require.NoError(t, err)
	assert.Nil(t, got.ShortName, "blank short name must be NULL (CHECK rejects blank)")
	assert.Nil(t, got.ImageObjectKey, "missing image must not make the course unusable")
	assert.Nil(t, got.AllowedWeekdays, "no weekday selection means every weekday")
}

func TestTrimmingCourseService_CreateRejectsInvalidAttributes(t *testing.T) {
	tests := []struct {
		name  string
		input CreateTrimmingCourseInput
	}{
		{"weekday above 6", CreateTrimmingCourseInput{Name: "A", AllowedWeekdays: []int{7}}},
		{"negative weekday", CreateTrimmingCourseInput{Name: "A", AllowedWeekdays: []int{-1}}},
		{"foreign clinic image", CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "trimming-images/2/x.jpg"}},
		{"signed url instead of key", CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "https://cdn.example/x.jpg"}},
		{"non object reference", CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "legacy.jpg"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			repo := &mockTrimmingCourseRepository{createFn: func(context.Context, *model.TrimmingCourse) error {
				created = true
				return nil
			}}
			svc := newAttributeCourseService(repo, &courseImageLifecycleFake{})

			input := tt.input
			_, err := svc.Create(context.Background(), 1, &input)

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
			assert.False(t, created)
		})
	}
}

func TestTrimmingCourseService_CreateRejectsDeletedImageKey(t *testing.T) {
	repo := &mockTrimmingCourseRepository{createFn: func(context.Context, *model.TrimmingCourse) error {
		t.Fatal("create must not run for a tombstoned image")
		return nil
	}}
	lifecycle := &courseImageLifecycleFake{deleted: map[string]bool{"trimming-images/1/gone.jpg": true}}
	svc := newAttributeCourseService(repo, lifecycle)

	_, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "trimming-images/1/gone.jpg"})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "want conflict, got %v", err)
}

func TestTrimmingCourseService_CreateImageFailsClosedWithoutLifecycle(t *testing.T) {
	repo := &mockTrimmingCourseRepository{createFn: func(context.Context, *model.TrimmingCourse) error {
		t.Fatal("create must not run without image reference protection")
		return nil
	}}
	svc := newAttributeCourseService(repo, nil)

	_, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{Name: "A", ImageObjectKey: "trimming-images/1/a.jpg"})

	require.Error(t, err)
	assert.False(t, apperrors.IsInvalidInput(err))
}

func courseAttrStrPtr(v string) *string { return &v }

func updateCaptureRepo(captured *map[string]any) *mockTrimmingCourseRepository {
	return &mockTrimmingCourseRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID, Name: "A"}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
			*captured = buildTrimmingCourseUpdate(&cmd)
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID}, nil
		},
	}
}

func TestTrimmingCourseService_UpdateAttributeTriState(t *testing.T) {
	emptyDays := []int{}
	days := []int{5, 0}
	tests := []struct {
		name  string
		input UpdateTrimmingCourseInput
		want  map[string]any
		lock  bool
	}{
		{
			name:  "omitted attributes keep stored values",
			input: UpdateTrimmingCourseInput{Name: courseAttrStrPtr("B")},
			want:  map[string]any{colTrimmingCourseName: "B"},
		},
		{
			name: "empty values clear to NULL",
			input: UpdateTrimmingCourseInput{
				ShortName: courseAttrStrPtr(" "), ImageObjectKey: courseAttrStrPtr(""), AllowedWeekdays: &emptyDays,
			},
			want: map[string]any{
				colTrimmingCourseShortName:       nil,
				colTrimmingCourseImageObjectKey:  nil,
				colTrimmingCourseAllowedWeekdays: nil,
			},
		},
		{
			name: "values are normalized",
			input: UpdateTrimmingCourseInput{
				ShortName:       courseAttrStrPtr(" S "),
				ImageObjectKey:  courseAttrStrPtr("trimming-images/1/n.png"),
				AllowedWeekdays: &days,
			},
			want: map[string]any{
				colTrimmingCourseShortName:       "S",
				colTrimmingCourseImageObjectKey:  "trimming-images/1/n.png",
				colTrimmingCourseAllowedWeekdays: pq.Int64Array{0, 5},
			},
			lock: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fields map[string]any
			lifecycle := &courseImageLifecycleFake{}
			svc := newAttributeCourseService(updateCaptureRepo(&fields), lifecycle)

			input := tt.input
			_, err := svc.Update(context.Background(), 1, 10, &input)

			require.NoError(t, err)
			assert.Equal(t, tt.want, fields)
			assert.Equal(t, tt.lock, len(lifecycle.locked) == 1)
		})
	}
}

func TestTrimmingCourseService_UpdateRejectsForeignImage(t *testing.T) {
	var fields map[string]any
	svc := newAttributeCourseService(updateCaptureRepo(&fields), &courseImageLifecycleFake{})

	_, err := svc.Update(context.Background(), 1, 10, &UpdateTrimmingCourseInput{
		ImageObjectKey: courseAttrStrPtr("trimming-images/2/x.jpg"),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, fields)
}

func TestTrimmingCourseService_UpdatePreservesReplacedOriginal(t *testing.T) {
	captured := map[string]any{}
	existing := &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: courseAttrStrPtr("trimming-images/1/old.jpg")}
	lifecycle := &courseImageLifecycleFake{}
	svc := newAttributeCourseService(newSupersedeUpdateRepo(existing, &captured), lifecycle)

	newKey := "trimming-images/1/new.jpg"
	_, err := svc.Update(context.Background(), 1, 10, &UpdateTrimmingCourseInput{
		ImageObjectKey: &newKey,
		ActorStaffID:   7,
	})
	require.NoError(t, err)
	assert.Empty(t, lifecycle.tombstoned)
}

func TestTrimmingCourseService_UpdatePreservesClearedOriginal(t *testing.T) {
	captured := map[string]any{}
	existing := &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: courseAttrStrPtr("trimming-images/1/old.jpg")}
	lifecycle := &courseImageLifecycleFake{}
	svc := newAttributeCourseService(newSupersedeUpdateRepo(existing, &captured), lifecycle)

	_, err := svc.Update(context.Background(), 1, 10, &UpdateTrimmingCourseInput{
		ImageObjectKey: courseAttrStrPtr(""),
		ActorStaffID:   7,
	})
	require.NoError(t, err)
	assert.Empty(t, lifecycle.tombstoned)
}

func TestTrimmingCourseService_UpdateKeepsImageReferencedElsewhere(t *testing.T) {
	captured := map[string]any{}
	existing := &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: courseAttrStrPtr("trimming-images/1/shared.jpg")}
	lifecycle := &courseImageLifecycleFake{inUseKeys: map[string]bool{"trimming-images/1/shared.jpg": true}}
	svc := newAttributeCourseService(newSupersedeUpdateRepo(existing, &captured), lifecycle)

	newKey := "trimming-images/1/new.jpg"
	_, err := svc.Update(context.Background(), 1, 10, &UpdateTrimmingCourseInput{
		ImageObjectKey: &newKey,
		ActorStaffID:   7,
	})
	require.NoError(t, err)
	assert.Empty(t, lifecycle.tombstoned)
}

func TestTrimmingCourseService_UpdateSkipsSupersedeWhenImageUnchanged(t *testing.T) {
	captured := map[string]any{}
	existing := &model.TrimmingCourse{ID: 10, ClinicID: 1, ImageObjectKey: courseAttrStrPtr("trimming-images/1/same.jpg")}
	lifecycle := &courseImageLifecycleFake{}
	svc := newAttributeCourseService(newSupersedeUpdateRepo(existing, &captured), lifecycle)

	same := "trimming-images/1/same.jpg"
	_, err := svc.Update(context.Background(), 1, 10, &UpdateTrimmingCourseInput{
		Name:           courseAttrStrPtr("B"),
		ImageObjectKey: &same,
		ActorStaffID:   7,
	})
	require.NoError(t, err)
	assert.Empty(t, lifecycle.tombstoned)
}

func TestTrimmingCourseRequest_AttributeMapping(t *testing.T) {
	create := createTrimmingCourseRequest{
		Name: "A", ShortName: "S", ImageObjectKey: "trimming-images/1/a.jpg", AllowedWeekdays: []int{1},
	}
	in := create.toServiceInput()
	assert.Equal(t, "S", in.ShortName)
	assert.Equal(t, "trimming-images/1/a.jpg", in.ImageObjectKey)
	assert.Equal(t, []int{1}, in.AllowedWeekdays)

	days := []int{}
	update := updateTrimmingCourseRequest{
		ShortName: courseAttrStrPtr(""), ImageObjectKey: courseAttrStrPtr(""), AllowedWeekdays: &days,
	}
	out := update.toServiceInput()
	require.NotNil(t, out.ShortName)
	require.NotNil(t, out.ImageObjectKey)
	require.NotNil(t, out.AllowedWeekdays)
	assert.Empty(t, *out.AllowedWeekdays)
	assert.Nil(t, updateTrimmingCourseRequest{}.toServiceInput().AllowedWeekdays)
}

func TestTrimmingCourseResponse_ExposesAttributes(t *testing.T) {
	short := "S"
	key := "trimming-images/1/a.jpg"
	days := pq.Int64Array{0, 6}
	resp := toTrimmingCourseResponse(&model.TrimmingCourse{
		ID: 1, ClinicID: 1, Name: "A", ShortName: &short, ImageObjectKey: &key, AllowedWeekdays: &days,
	})
	assert.Equal(t, &short, resp.ShortName)
	assert.Equal(t, &key, resp.ImageObjectKey)
	assert.Equal(t, []int{0, 6}, resp.AllowedWeekdays)

	empty := toTrimmingCourseResponse(&model.TrimmingCourse{ID: 2, ClinicID: 1, Name: "B"})
	assert.Nil(t, empty.ShortName)
	assert.Nil(t, empty.ImageObjectKey)
	assert.Nil(t, empty.AllowedWeekdays, "NULL weekdays stay null on the wire (= every weekday)")
}
