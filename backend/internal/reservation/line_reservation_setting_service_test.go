package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

// testIntegrationKeyHex は 32 バイト（AES-256）のダミー暗号鍵。
const testIntegrationKeyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
		input := &UpsertLineReservationSettingInput{
			Status:          "active",
			ClosedWeekdays:  []byte("[1, 2]"),
			LineAccessToken: "token",
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, res)
		// R-05 Phase B: channel secret is never written from this path.
		assert.Empty(t, res.LineChannelSecret)
		assert.Equal(t, "token", res.LineAccessToken)
	})

	t.Run("success update existing setting keeps access token; never rewrites channel secret", func(t *testing.T) {
		existing := &model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: "old_secret",
			LineAccessToken:   "old_token",
		}
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
		input := &UpsertLineReservationSettingInput{
			Status:          "active",
			ClosedWeekdays:  []byte("[1, 2]"),
			LineAccessToken: "",
		}
		res, isNew, err := svc.Save(ctx, 1, input)
		assert.NoError(t, err)
		assert.False(t, isNew)
		assert.NotNil(t, res)
		// Access token empty-input preserve remains (co-travel token path).
		assert.Equal(t, "old_token", res.LineAccessToken)
		// Channel secret must not be copied/re-encrypted into the write model.
		assert.Empty(t, res.LineChannelSecret)
		if assert.NotNil(t, persisted) {
			assert.Empty(t, persisted.LineChannelSecret)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			Status:         "running",
			ClosedWeekdays: []byte("[0,6]"),
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			assert.JSONEq(t, "[0,6]", string(persisted.ClosedWeekdays))
		}
	})

	t.Run("invalid json fields", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(nil, testPlainEncrypt, testPlainDecrypt)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`{}`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 配列だが数値フィールド（文字列でない）→ 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, testPlainEncrypt, testPlainDecrypt)
		input := &UpsertLineReservationSettingInput{
			BreakHours: []byte(`[{"start":1200,"end":1300}]`),
		}
		_, _, err := svc.Save(ctx, 1, input)
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("break_hours: 配列・文字列だが HHMM 形式でないエントリ → 400", func(t *testing.T) {
		svc := NewLineReservationSettingService(nil, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
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
		svc := NewLineReservationSettingService(repo, testPlainEncrypt, testPlainDecrypt)
		res, isNew, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{Status: "active"})
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.NotNil(t, res)
		assert.Equal(t, uint64(1), res.ClinicID)
		assert.Equal(t, 1, findCount, "must not re-fetch after Save")
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

	t.Run("plaintext access token is encrypted at rest; channel secret is never written", func(t *testing.T) {
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
		svc := NewLineReservationSettingService(repo, testCipherEncrypt(cipher), testCipherDecrypt(cipher))
		_, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			LineAccessToken: "plain-token",
		})
		assert.NoError(t, err)
		if assert.NotNil(t, persisted) {
			// R-05 Phase B: channel secret write path removed — model must stay empty.
			assert.Empty(t, persisted.LineChannelSecret)
			// Access token still encrypts at rest.
			assert.NotEqual(t, "plain-token", persisted.LineAccessToken)
			gotToken, decErr := cipher.Decrypt(persisted.LineAccessToken)
			assert.NoError(t, decErr)
			assert.Equal(t, "plain-token", gotToken)
		}
	})

	t.Run("existing encrypted access token preserved without double-encryption when input empty", func(t *testing.T) {
		encSecret, err := cipher.Encrypt("kept-secret")
		assert.NoError(t, err)
		encToken, err := cipher.Encrypt("kept-token")
		assert.NoError(t, err)
		repo := newCapturingRepo(&model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: encSecret,
			LineAccessToken:   encToken,
		})
		svc := NewLineReservationSettingService(repo, testCipherEncrypt(cipher), testCipherDecrypt(cipher))
		res, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{
			LineAccessToken: "",
		})
		assert.NoError(t, err)
		if assert.NotNil(t, res) {
			// Channel secret must not be re-read/re-encrypted onto the write model.
			assert.Empty(t, res.LineChannelSecret)
			// Access token: 二重暗号化されず、1 回の復号で元の平文に戻る
			gotToken, decErr := cipher.Decrypt(res.LineAccessToken)
			assert.NoError(t, decErr)
			assert.Equal(t, "kept-token", gotToken)
		}
	})

	t.Run("legacy plaintext access token is re-encrypted; channel secret is left untouched by service", func(t *testing.T) {
		// 暗号化導入前に保存された平文 access token を模す（AES-GCM 復号に失敗する値）
		// 実クレデンシャルを使わず、明らかに合成の 32-hex ダミーを用いる
		const legacyPlainSecret = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		repo := newCapturingRepo(&model.LineReservationSetting{
			ClinicID:          1,
			LineChannelSecret: legacyPlainSecret,
			LineAccessToken:   "legacy-plain-token",
		})
		svc := NewLineReservationSettingService(repo, testCipherEncrypt(cipher), testCipherDecrypt(cipher))
		res, _, err := svc.Save(ctx, 1, &UpsertLineReservationSettingInput{})
		assert.NoError(t, err)
		if assert.NotNil(t, res) {
			// R-05 Phase B: service must not opportunistic-re-encrypt channel secret.
			assert.Empty(t, res.LineChannelSecret)

			assert.NotEqual(t, "legacy-plain-token", res.LineAccessToken)
			gotToken, decErr := cipher.Decrypt(res.LineAccessToken)
			assert.NoError(t, decErr)
			assert.Equal(t, "legacy-plain-token", gotToken)
		}
	})
}

// testPlainEncrypt/testPlainDecrypt は旧 cipher=nil 挙動（暗号化なし素通し）を再現する
// （lstep/line_credentials.go の encrypt/lstep.DecryptLineCredential の nil-cipher 分岐と同一契約）。
func testPlainEncrypt(value string) (string, error) { return value, nil }

func testPlainDecrypt(_ context.Context, value string) string { return value }

// testCipherEncrypt/testCipherDecrypt はproduction注入closure（composition rootが
// lstep/line_credentials.goのencrypt/decryptを包む）と同一契約の
// テスト側再現（nil/空文字素通し+実 cipher・H-4）。helper 本体の単体検証は
// service/line_credentials_test.go が正本。
func testCipherEncrypt(cipher *crypto.AESGCMCipher) func(string) (string, error) {
	return func(value string) (string, error) {
		if cipher == nil || value == "" {
			return value, nil
		}
		return cipher.Encrypt(value)
	}
}

func testCipherDecrypt(cipher *crypto.AESGCMCipher) func(context.Context, string) string {
	return func(_ context.Context, value string) string {
		if cipher == nil || value == "" {
			return value
		}
		plaintext, err := cipher.Decrypt(value)
		if err != nil {
			return value
		}
		return plaintext
	}
}
