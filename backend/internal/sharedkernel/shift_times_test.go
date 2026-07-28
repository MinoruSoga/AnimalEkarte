package sharedkernel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHHMM(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{name: "HH:MM", input: "09:05", wantH: 9, wantM: 5},
		{name: "HH:MM:SS strips seconds", input: "18:30:00", wantH: 18, wantM: 30},
		{name: "empty", input: "", wantErr: true},
		{name: "bad separator", input: "09-05", wantErr: true},
		{name: "no zero pad", input: "9:05", wantErr: true},
		{name: "non-numeric", input: "ab:cd", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m, err := ParseHHMM(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantH, h)
			assert.Equal(t, tt.wantM, m)
		})
	}
}
