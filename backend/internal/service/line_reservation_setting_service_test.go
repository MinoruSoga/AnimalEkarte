package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// testIntegrationKeyHex は 32 バイト（AES-256）のダミー暗号鍵。
const testIntegrationKeyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type mockLineReservationSettingRepository struct {
	repository.LineReservationSettingRepository
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
		svc := NewLineReservationSettingService(repo, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
		input := &UpsertLineReservationSettingInput{
			Status:            "active",
			ClosedWeekdays:    []byte("[1, 2]"),
			LineChannelSecret: "secret",
			LineAccessToken:   "token",
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, res)
	})

	t.Run("success update existing setting keeping sensitive keys", func(t *testing.T) {
		existing := &model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: "old_secret",
			LineAccessToken:   "old_token",
		}
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return existing, nil
			},
		}
		svc := NewLineReservationSettingService(repo, nil)
		input := &UpsertLineReservationSettingInput{
			Status:            "active",
			ClosedWeekdays:    []byte("[1, 2]"),
			LineChannelSecret: "",
			LineAccessToken:   "",
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.False(t, isNew)
		assert.NotNil(t, res)
	})

	t.Run("invalid json fields", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, nil)
		input := &UpsertLineReservationSettingInput{
			ClosedWeekdays: []byte("bad-json"),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
	})

	// BE-refactor.md D10/F-2 フォローアップ（go-reviewer/security-reviewer 指摘）:
	// break_hours の unmarshal 失敗を予約作成側で fail-closed にしたため、保存側でも形状を
	// 検証しないと、構文的には有効だが形状が不正な値が保存された時点で当該クリニックの
	// 全 LINE 予約が拒否され続ける可用性劣化になる。保存時に 400 で弾くことを保証する。
	t.Run("break_hours: 構文は有効だが形状が不正（オブジェクト）→ 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, nil)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`{}`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 配列だが数値フィールド（文字列でない）→ 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, nil)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`[{"start":1200,"end":1300}]`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 配列・文字列だが HHMM 形式でないエントリ → 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
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
		svc := NewLineReservationSettingService(repo, nil)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})

	t.Run("db reload error after save", func(t *testing.T) {
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
		svc := NewLineReservationSettingService(repo, nil)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.Error(t, err)
	})
}

// TestLineReservationSettingService_Save_Encryption は cipher 有効時の暗号化挙動を検証する（H-4）。
func TestLineReservationSettingService_Save_Encryption(t *testing.T) {
	ctx := context.Background()
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	if err != nil {
		t.Fatalf("failed to build test cipher: %v", err)
	}

	// newCapturingRepo は 1 回目の FindByClinicID で existing を返し（nil ならば NotFound）、
	// Save で永続化された setting を捕捉し、以降の FindByClinicID（Save 内の再読み込み）でそれを返す。
	// Save の戻り値がこの捕捉済み setting になるため、呼び出し側は戻り値を検証する。
	newCapturingRepo := func(existing *model.LineReservationSetting) *mockLineReservationSettingRepository {
		var persisted *model.LineReservationSetting
		return &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				if persisted != nil {
					return persisted, nil
				}
				if existing != nil {
					return existing, nil
				}
				return nil, apperrors.WrapNotFound("setting", "")
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
	}

	t.Run("plaintext secrets are encrypted at rest on save", func(t *testing.T) {
		var persisted *model.LineReservationSetting
		var findCount int
		repo := &mockLineReservationSettingRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				findCount++
				if findCount == 1 {
					return nil, apperrors.WrapNotFound("setting", "")
				}
				return persisted, nil
			},
			saveFn: func(_ context.Context, _ uint64, setting *model.LineReservationSetting) error {
				persisted = setting
				return nil
			},
		}
		svc := NewLineReservationSettingService(repo, cipher)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			LineChannelSecret: "plain-secret",
			LineAccessToken:   "plain-token",
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			// 平文が DB 値としてそのまま残らない
			assert.NotEqual(t, "plain-secret", persisted.LineChannelSecret)
			assert.NotEqual(t, "plain-token", persisted.LineAccessToken)
			// round-trip: 復号すると元の平文に戻る
			gotSecret, decErr := cipher.Decrypt(persisted.LineChannelSecret)
			assert.NoError(t, decErr)
			assert.Equal(t, "plain-secret", gotSecret)
			gotToken, decErr := cipher.Decrypt(persisted.LineAccessToken)
			assert.NoError(t, decErr)
			assert.Equal(t, "plain-token", gotToken)
		}
	})

	t.Run("existing encrypted value preserved without double-encryption when input empty", func(t *testing.T) {
		encSecret, err := cipher.Encrypt("kept-secret")
		assert.NoError(t, err)
		encToken, err := cipher.Encrypt("kept-token")
		assert.NoError(t, err)
		repo := newCapturingRepo(&model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: encSecret,
			LineAccessToken:   encToken,
		})
		svc := NewLineReservationSettingService(repo, cipher)
		res, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			LineChannelSecret: "",
			LineAccessToken:   "",
		})
		assert.NoError(t, err)
		if assert.NotNil(t, res) {
			// 二重暗号化されず、1 回の復号で元の平文に戻る
			gotSecret, decErr := cipher.Decrypt(res.LineChannelSecret)
			assert.NoError(t, decErr)
			assert.Equal(t, "kept-secret", gotSecret)
			gotToken, decErr := cipher.Decrypt(res.LineAccessToken)
			assert.NoError(t, decErr)
			assert.Equal(t, "kept-token", gotToken)
		}
	})

	t.Run("legacy plaintext existing is re-encrypted when input empty", func(t *testing.T) {
		// 暗号化導入前に保存された平文レコードを模す（AES-GCM 復号に失敗する値）
		repo := newCapturingRepo(&model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: "5344ef84eb7072b5894f7e087db28827",
			LineAccessToken:   "legacy-plain-token",
		})
		svc := NewLineReservationSettingService(repo, cipher)
		res, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.NoError(t, err)
		if assert.NotNil(t, res) {
			// 機会的再暗号化: 平文のまま残らず、暗号化される
			assert.NotEqual(t, "5344ef84eb7072b5894f7e087db28827", res.LineChannelSecret)
			gotSecret, decErr := cipher.Decrypt(res.LineChannelSecret)
			assert.NoError(t, decErr)
			assert.Equal(t, "5344ef84eb7072b5894f7e087db28827", gotSecret)

			assert.NotEqual(t, "legacy-plain-token", res.LineAccessToken)
			gotToken, decErr := cipher.Decrypt(res.LineAccessToken)
			assert.NoError(t, decErr)
			assert.Equal(t, "legacy-plain-token", gotToken)
		}
	})
}
