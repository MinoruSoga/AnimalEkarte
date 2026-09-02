package lstep

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const (
	csvPreflightBufferBytes = 4 * 1024
)

type csvLexicalState uint8

const (
	csvFieldStart csvLexicalState = iota
	csvUnquotedField
	csvQuotedField
	csvQuoteClosed
)

func spoolCSVToTemp(fileReader io.Reader) (*os.File, error) {
	file, err := os.CreateTemp("", "animal-ekarte-lstep-csv-*")
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create CSV staging file")
	}
	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}

	written, err := io.Copy(file, io.LimitReader(fileReader, maxCSVSizeBytes+1))
	if err != nil {
		cleanup()
		return nil, apperrors.WrapInvalidInput("failed to read CSV file")
	}
	if written > maxCSVSizeBytes {
		cleanup()
		return nil, apperrors.WrapInvalidInput("CSV file exceeds 50MB limit")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return nil, apperrors.Wrap(err, "failed to rewind CSV staging file")
	}
	return file, nil
}

func cleanupTempCSV(file *os.File) {
	if file == nil {
		return
	}
	name := file.Name()
	_ = file.Close()
	_ = os.Remove(name)
}

// preflightCSVShape performs an allocation-bounded RFC 4180 lexical scan before
// encoding/csv sees the input. This prevents a comma- or cell-heavy record from
// growing encoding/csv's internal field index slices before our limits apply.
func preflightCSVShape(input io.Reader) error {
	reader := bufio.NewReaderSize(input, csvPreflightBufferBytes)
	state := csvFieldStart
	columns := 1
	cellBytes := 0
	records := 0
	recordStarted := false
	expectedColumns := 0

	addCellByte := func() error {
		cellBytes++
		if cellBytes > maxCSVCellBytes {
			return fmt.Errorf("CSV cell exceeds %d byte limit", maxCSVCellBytes)
		}
		return nil
	}
	finishField := func() error {
		columns++
		if columns > maxCSVColumns {
			return fmt.Errorf("CSV column count exceeds %d column limit", maxCSVColumns)
		}
		cellBytes = 0
		state = csvFieldStart
		return nil
	}
	finishRecord := func() error {
		if !recordStarted {
			columns = 1
			cellBytes = 0
			state = csvFieldStart
			return nil
		}
		if expectedColumns == 0 {
			expectedColumns = columns
		} else if columns != expectedColumns {
			return fmt.Errorf("failed to parse CSV: wrong number of fields")
		}
		records++
		if records > maxCSVDataRows+1 {
			return fmt.Errorf("CSV row count exceeds %d row limit", maxCSVDataRows)
		}
		columns = 1
		cellBytes = 0
		state = csvFieldStart
		recordStarted = false
		return nil
	}
	consumeRecordEnd := func(current byte) (bool, error) {
		if current == '\n' {
			return true, finishRecord()
		}
		if current != '\r' {
			return false, nil
		}
		next, peekErr := reader.Peek(1)
		if peekErr == nil && next[0] == '\n' {
			_, _ = reader.ReadByte()
			return true, finishRecord()
		}
		return false, nil
	}

	for {
		current, err := reader.ReadByte()
		if errors.Is(err, io.EOF) {
			if !recordStarted && records == 0 {
				return fmt.Errorf("CSV file is empty")
			}
			if state == csvQuotedField {
				return fmt.Errorf("failed to parse CSV: unclosed quoted field")
			}
			if recordStarted {
				return finishRecord()
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV during preflight: %w", err)
		}
		switch state {
		case csvFieldStart:
			switch current {
			case ',':
				recordStarted = true
				if err := finishField(); err != nil {
					return err
				}
			case '"':
				recordStarted = true
				state = csvQuotedField
			default:
				if ended, endErr := consumeRecordEnd(current); ended || endErr != nil {
					if endErr != nil {
						return endErr
					}
					continue
				}
				recordStarted = true
				state = csvUnquotedField
				if err := addCellByte(); err != nil {
					return err
				}
			}

		case csvUnquotedField:
			switch current {
			case ',':
				if err := finishField(); err != nil {
					return err
				}
			case '"':
				return fmt.Errorf("failed to parse CSV: bare quote in unquoted field")
			default:
				if ended, endErr := consumeRecordEnd(current); ended || endErr != nil {
					if endErr != nil {
						return endErr
					}
					continue
				}
				if err := addCellByte(); err != nil {
					return err
				}
			}

		case csvQuotedField:
			if current != '"' {
				if current == '\r' {
					consumeOptionalLF(reader)
				}
				if err := addCellByte(); err != nil {
					return err
				}
				continue
			}
			next, peekErr := reader.Peek(1)
			if peekErr == nil && next[0] == '"' {
				_, _ = reader.ReadByte()
				if err := addCellByte(); err != nil {
					return err
				}
				continue
			}
			state = csvQuoteClosed

		case csvQuoteClosed:
			switch current {
			case ',':
				if err := finishField(); err != nil {
					return err
				}
			default:
				if ended, endErr := consumeRecordEnd(current); ended || endErr != nil {
					if endErr != nil {
						return endErr
					}
					continue
				}
				return fmt.Errorf("failed to parse CSV: unexpected data after closing quote")
			}
		}
	}
}

func newDecodedCSVReader(file *os.File) (io.Reader, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apperrors.Wrap(err, "failed to rewind CSV staging file")
	}
	prefix := make([]byte, len(utf8BOM))
	n, readErr := file.Read(prefix)
	if readErr != nil && readErr != io.EOF {
		return nil, apperrors.Wrap(readErr, "failed to inspect CSV encoding")
	}
	offset := int64(0)
	if n == len(utf8BOM) && bytes.Equal(prefix, utf8BOM) {
		offset = int64(len(utf8BOM))
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, apperrors.Wrap(err, "failed to inspect CSV encoding")
	}

	validUTF8, err := isValidUTF8Stream(file)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to inspect CSV encoding")
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, apperrors.Wrap(err, "failed to rewind decoded CSV")
	}
	if validUTF8 {
		return file, nil
	}
	return transform.NewReader(file, japanese.ShiftJIS.NewDecoder()), nil
}

func consumeOptionalLF(reader *bufio.Reader) {
	next, peekErr := reader.Peek(1)
	if peekErr == nil && next[0] == '\n' {
		_, _ = reader.ReadByte()
	}
}

func validateDecodedCSVRecord(record []string) error {
	if len(record) > maxCSVColumns {
		return fmt.Errorf("CSV column count exceeds %d column limit", maxCSVColumns)
	}
	for _, cell := range record {
		if len(cell) > maxCSVCellBytes {
			return fmt.Errorf("CSV cell exceeds %d byte limit", maxCSVCellBytes)
		}
	}
	return nil
}

func isValidUTF8Stream(input io.Reader) (bool, error) {
	reader := bufio.NewReaderSize(input, csvPreflightBufferBytes)
	for {
		r, size, err := reader.ReadRune()
		if errors.Is(err, io.EOF) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to read UTF-8 stream: %w", err)
		}
		if r == utf8.RuneError && size == 1 {
			return false, nil
		}
	}
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}
