package lstep

import (
	"encoding/csv"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestPreflightCSVShape_MatchesDecodedCellBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "quoted cell at exact limit",
			input: `"` + strings.Repeat("x", maxCSVCellBytes) + `"` + "\n",
		},
		{
			name:  "escaped quotes decode to exact limit",
			input: `"` + strings.Repeat(`""`, maxCSVCellBytes) + `"` + "\n",
		},
		{
			name:  "quoted CRLF counts as one decoded byte",
			input: `"` + strings.Repeat("x", maxCSVCellBytes-1) + "\r\n" + `"` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, preflightCSVShape(strings.NewReader(tt.input)))

			reader := csv.NewReader(strings.NewReader(tt.input))
			record, err := reader.Read()
			require.NoError(t, err)
			require.Len(t, record, 1)
			assert.Len(t, record[0], maxCSVCellBytes)
			require.NoError(t, validateDecodedCSVRecord(record))
		})
	}
}

func TestPreflightCSVShape_IgnoresBlankLinesForRowLimit(t *testing.T) {
	input := "line_user_id\n\n" + strings.Repeat("U1\n", maxCSVDataRows)

	require.NoError(t, preflightCSVShape(strings.NewReader(input)))
}

func TestPreflightCSVShape_UsesStrictQuoteSyntax(t *testing.T) {
	err := preflightCSVShape(strings.NewReader("line_user_id\nU\"1\n"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare quote")
}

func TestCSVClientError_HidesParseDetailsAndPassesAppError(t *testing.T) {
	t.Run("csv parse error becomes fixed client message", func(t *testing.T) {
		inner := &csv.ParseError{StartLine: 2, Line: 2, Err: csv.ErrFieldCount}
		err := csvClientError(inner)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), csvFormatInvalidMsg)
		assert.NotContains(t, err.Error(), inner.Error())
		assert.NotContains(t, err.Error(), "wrong number of fields")
	})
	t.Run("already AppError is passed through", func(t *testing.T) {
		orig := apperrors.WrapInvalidInput("CSV row count exceeds limit")
		err := csvClientError(orig)
		require.Equal(t, orig, err)
		assert.Contains(t, err.Error(), "row count exceeds")
	})
}

func TestPreparedCSVStream_DecodesAfterPreflight(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "UTF-8 BOM",
			input: append([]byte{0xEF, 0xBB, 0xBF}, []byte("LINE ID,表示名\nU1,山田\n")...),
			want:  "LINE ID,表示名\nU1,山田\n",
		},
		{
			name:  "Shift_JIS",
			input: encodeShiftJIS(t, "LINE ID,表示名\nU1,山田\n"),
			want:  "LINE ID,表示名\nU1,山田\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := spoolCSVToTemp(strings.NewReader(string(tt.input)))
			require.NoError(t, err)
			defer cleanupTempCSV(file)

			preflightReader, err := newDecodedCSVReader(file)
			require.NoError(t, err)
			require.NoError(t, preflightCSVShape(preflightReader))

			decodedReader, err := newDecodedCSVReader(file)
			require.NoError(t, err)
			decoded, err := io.ReadAll(decodedReader)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(decoded))
		})
	}
}
