package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
)

func TestParseJSTDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "YYYY-MM-DDをJSTの当日0時として解釈する",
			input: "2026-04-17",
			want:  time.Date(2026, 4, 17, 0, 0, 0, 0, config.JST),
		},
		{
			name:    "RFC3339形式はエラーにする",
			input:   "2026-04-17T00:00:00+09:00",
			wantErr: true,
		},
		{
			name:    "空文字はエラーにする",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSTDate(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppointmentDayRange(t *testing.T) {
	jst := config.JST
	tests := []struct {
		name      string
		input     time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "UTC時刻をJSTの日単位レンジに変換する",
			input:     time.Date(2026, 4, 16, 15, 30, 0, 0, time.UTC), // JST 2026-04-17 00:30
			wantStart: time.Date(2026, 4, 17, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2026, 4, 18, 0, 0, 0, 0, jst),
		},
		{
			name:      "JST時刻は同日の開始から翌日開始までに丸める",
			input:     time.Date(2026, 4, 17, 23, 59, 59, 0, jst),
			wantStart: time.Date(2026, 4, 17, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2026, 4, 18, 0, 0, 0, 0, jst),
		},
		{
			name:      "JST深夜0時ちょうどは同日の開始になる",
			input:     time.Date(2026, 4, 17, 0, 0, 0, 0, jst),
			wantStart: time.Date(2026, 4, 17, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2026, 4, 18, 0, 0, 0, 0, jst),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := AppointmentDayRange(tt.input)

			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}
