package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockLineReservationSettingRepository struct {
	LineReservationSettingRepository
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	findAllFn        func(ctx context.Context) ([]model.LineReservationSetting, error)
	saveFn           func(ctx context.Context, clinicID uint64, setting *model.LineReservationSetting) error
}

func (m *mockLineReservationSettingRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{ClinicID: clinicID}, nil
}

func (m *mockLineReservationSettingRepository) FindAll(ctx context.Context) ([]model.LineReservationSetting, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return []model.LineReservationSetting{}, nil
}

func (m *mockLineReservationSettingRepository) Save(ctx context.Context, clinicID uint64, setting *model.LineReservationSetting) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, clinicID, setting)
	}
	return nil
}

func TestLineReservationSettingService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{ClinicID: clinicID}, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		res, err := svc.Get(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("not found returns nil without error", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
		}
		svc := NewLineReservationSettingService(repo)
		res, err := svc.Get(ctx, 1)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, err := svc.Get(ctx, 1)
		assert.Error(t, err)
	})
}

func TestLineReservationSettingService_Save(t *testing.T) {
	ctx := context.Background()

	t.Run("success create new setting", func(t *testing.T) {
		var callCount int
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				callCount++
				if callCount == 1 {
					return nil, apperrors.WrapNotFound("setting", "")
				}
				return &model.LineReservationSetting{ClinicID: clinicID}, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		input := &UpsertLineReservationSettingInput{
			Status:         "active",
			ClosedWeekdays: []byte("[1, 2]"),
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, res)
		assert.Equal(t, "active", res.Status)
	})

	t.Run("success update existing setting", func(t *testing.T) {
		existing := &model.LineReservationSetting{ClinicID: 1}
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return existing, nil
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		input := &UpsertLineReservationSettingInput{
			Status:         "active",
			ClosedWeekdays: []byte("[1, 2]"),
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.False(t, isNew)
		assert.NotNil(t, res)
		assert.NotNil(t, persisted)
	})

	// EMR-504 contract A / A03 §3.1: business_hours_by_weekday is nullable; an empty input and the JSON
	// literal null both mean "release the weekday override" and are stored as SQL NULL (a nil pointer).
	t.Run("business_hours_by_weekday: empty and null are stored as NULL, a value is kept", func(t *testing.T) {
		cases := []struct {
			name    string
			input   []byte
			wantNil bool
		}{
			{name: "nil input", input: nil, wantNil: true},
			{name: "empty input", input: []byte(""), wantNil: true},
			{name: "whitespace only", input: []byte("  "), wantNil: true},
			{name: "json null", input: []byte("null"), wantNil: true},
			{name: "object", input: []byte(`{"6":{"start":"0900","end":"1300"}}`), wantNil: false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				var persisted *model.LineReservationSetting
				repo := &mockLineReservationSettingRepository{
					findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
						return nil, apperrors.WrapNotFound("setting", "")
					},
					saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
						persisted = setting
						return nil
					},
				}
				svc := NewLineReservationSettingService(repo)
				_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
					Status:                 "active",
					ClosedWeekdays:         []byte("[]"),
					BusinessHoursByWeekday: tc.input,
				})
				if !assert.NoError(t, err) || !assert.NotNil(t, persisted) {
					return
				}
				if tc.wantNil {
					assert.Nil(t, persisted.BusinessHoursByWeekday)
					return
				}
				if assert.NotNil(t, persisted.BusinessHoursByWeekday) {
					assert.JSONEq(t, string(tc.input), string(*persisted.BusinessHoursByWeekday))
				}
			})
		}
	})

	t.Run("omitted additional_fields defaults to empty JSON array on create", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, isNew, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:         "running",
			ClosedWeekdays: []byte("[]"),
		})
		assert.NoError(t, err)
		assert.True(t, isNew)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, "[]", string(persisted.AdditionalFields))
		}
	})

	t.Run("json null additional_fields defaults to empty JSON array", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:           "running",
			AdditionalFields: []byte("null"),
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, "[]", string(persisted.AdditionalFields))
		}
	})

	t.Run("explicit additional_fields array is preserved", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		raw := []byte(`[{"key":"phone"}]`)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:           "running",
			AdditionalFields: raw,
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, `[{"key":"phone"}]`, string(persisted.AdditionalFields))
		}
	})

	t.Run("omitted not-null jsonb columns default instead of SQL null", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{Status: "running"})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, "[]", string(persisted.ClosedWeekdays))
			assert.JSONEq(t, "[]", string(persisted.ClosedDates))
			assert.JSONEq(t, "[]", string(persisted.BreakHours))
			assert.JSONEq(t, "[]", string(persisted.AdditionalFields))
			assert.JSONEq(t, `{"start":"0900","end":"1900"}`, string(persisted.BusinessHours))
			assert.Empty(t, persisted.BusinessHoursByWeekday)
		}
	})

	t.Run("json null closed_weekdays defaults to empty JSON array", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:         "running",
			ClosedWeekdays: []byte("null"),
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, "[]", string(persisted.ClosedWeekdays))
		}
	})

	t.Run("explicit closed_weekdays array is preserved", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:         "running",
			ClosedWeekdays: []byte("[0,6]"),
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, "[0,6]", string(persisted.ClosedWeekdays))
		}
	})

	t.Run("legacy closed weekdays are persisted as canonical numeric JSON without input mutation", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		closedWeekdays := []byte(`["0",2,"6"]`)
		original := append([]byte(nil), closedWeekdays...)

		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:         "running",
			ClosedWeekdays: closedWeekdays,
		})

		assert.NoError(t, err)
		assert.Equal(t, original, closedWeekdays)
		if assert.NotNil(t, persisted) {
			assert.Equal(t, `[0,2,6]`, string(persisted.ClosedWeekdays))
		}
	})

	t.Run("invalid closed weekdays are rejected before repository operations", func(t *testing.T) {
		tests := []struct {
			name  string
			input []byte
		}{
			{name: "weekday out of range", input: []byte(`[2,7]`)},
			{name: "whitespace-only malformed JSON", input: []byte(" \t\n")},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var repositoryCalls int
				repo := &mockLineReservationSettingRepository{
					findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
						repositoryCalls++
						return nil, apperrors.WrapNotFound("setting", "")
					},
					saveFn: func(_ context.Context, _ uint64, _ *model.LineReservationSetting) error {
						repositoryCalls++
						return nil
					},
				}
				svc := NewLineReservationSettingService(repo)
				inputJSON := append([]byte(nil), tt.input...)
				input := &UpsertLineReservationSettingInput{ClosedWeekdays: tt.input}

				setting, isNew, err := svc.Save(ctx, 1, input)

				assert.Nil(t, setting)
				assert.False(t, isNew)
				if !assert.Error(t, err) {
					return
				}
				assert.True(t, apperrors.IsInvalidInput(err))
				assert.NotContains(t, err.Error(), string(inputJSON))
				assert.Equal(t, 0, repositoryCalls)
				assert.Equal(t, inputJSON, input.ClosedWeekdays)
			})
		}
	})

	t.Run("invalid json fields", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil)
		input := &UpsertLineReservationSettingInput{
			ClosedWeekdays: []byte("bad-json"),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
	})

	// BE-refactor.md D10/F-2 フォローアップ（go-reviewer/security-reviewer 指摘）:
	// break_hours の unmarshal 失敗を予約作成側で fail-closed にしたため、保存側でも形状を
	// 検証しないと、構文的には有効だが形状が不正な値が保存された時点で当該医院の
	// 全 LINE 予約が拒否され続ける可用性劣化になる。保存時に 400 で弾くことを保証する。
	t.Run("break_hours: 構文は有効だが形状が不正（オブジェクト）→ 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`{}`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 配列だが数値フィールド（文字列でない）→ 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`[{"start":1200,"end":1300}]`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 配列・文字列だが HHMM 形式でないエントリ → 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`[{"start":"not-a-time","end":"1300"}]`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 正しい形状は保存できる（回帰防止）", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{ClinicID: clinicID}, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`[{"start":"1200","end":"1300"}]`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
	})

	t.Run("break_hours: 空は許容される（休憩時間なし）", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{ClinicID: clinicID}, nil
			},
		}
		svc := NewLineReservationSettingService(repo)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`[]`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
	})

	t.Run("db find error on save", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})

	t.Run("db save error", func(t *testing.T) {
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, _ *model.LineReservationSetting) error {
				return errors.New("save error")
			},
		}
		svc := NewLineReservationSettingService(repo)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})

	t.Run("returns write result without post-save reload (RSV-03)", func(t *testing.T) {
		// Second FindByClinicID would fail; Save must still succeed with write result.
		var findCount int
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				findCount++
				if findCount > 1 {
					return nil, errors.New("db load error")
				}
				return nil, apperrors.WrapNotFound("setting", "")
			},
		}
		svc := NewLineReservationSettingService(repo)
		res, isNew, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{Status: "active"})
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, res)
		assert.Equal(t, uint64(1), res.ClinicID)
		assert.Equal(t, 1, findCount, "must not re-fetch after Save")
	})
}
