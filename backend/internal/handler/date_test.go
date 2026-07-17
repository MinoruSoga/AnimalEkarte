package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func TestJsonDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string // raw JSON token, e.g. `"2026-07-09"`
		want    time.Time
		wantErr bool
	}{
		{
			name:  "YYYY-MM-DD形式を time.Local でパースする",
			input: `"2026-07-09"`,
			want:  time.Date(2026, 7, 9, 0, 0, 0, 0, time.Local),
		},
		{
			name:  "RFC3339形式をパースする",
			input: `"2026-07-09T12:34:56Z"`,
			want:  time.Date(2026, 7, 9, 12, 34, 56, 0, time.UTC),
		},
		{
			name:  "null は無視する(ゼロ値のまま)",
			input: `null`,
			want:  time.Time{},
		},
		{
			name:  "空文字は無視する(ゼロ値のまま)",
			input: `""`,
			want:  time.Time{},
		},
		{
			name:    "不正な形式は汎用エラーを返す",
			input:   `"not-a-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d jsonDate
			err := json.Unmarshal([]byte(tt.input), &d)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				assert.Equal(t, flexibleDateInvalidInputMsg+": "+apperrors.ErrInvalidInput.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			assert.True(t, tt.want.Equal(d.Time))
		})
	}
}

func TestParseDate(t *testing.T) {
	t.Run("nilポインタは(nil, nil)を返す", func(t *testing.T) {
		got, err := parseDate(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("YYYY-MM-DD形式を time.Local でパースする", func(t *testing.T) {
		s := "2026-07-09"
		got, err := parseDate(&s)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, time.Date(2026, 7, 9, 0, 0, 0, 0, time.Local).Equal(*got))
	})

	t.Run("RFC3339形式をパースする", func(t *testing.T) {
		s := "2026-07-09T12:34:56Z"
		got, err := parseDate(&s)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, time.Date(2026, 7, 9, 12, 34, 56, 0, time.UTC).Equal(*got))
	})

	t.Run("不正な形式は入力値を含まない汎用エラーを返す", func(t *testing.T) {
		s := "not-a-date"
		got, err := parseDate(&s)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.NotContains(t, err.Error(), "not-a-date")
		assert.Equal(t, flexibleDateInvalidInputMsg+": "+apperrors.ErrInvalidInput.Error(), err.Error())
	})
}
