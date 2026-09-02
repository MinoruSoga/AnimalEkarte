package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"log/slog"
	"testing"
)

func TestSanitizeCSVCell(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cell string
		want string
	}{
		{name: "empty stays empty", cell: "", want: ""},
		{name: "ordinary name", cell: "田中 太郎", want: "田中 太郎"},
		{name: "equals formula", cell: "=SUM(A1:A2)", want: "'=SUM(A1:A2)"},
		{name: "plus prefix", cell: "+1234", want: "'+1234"},
		{name: "minus prefix", cell: "-HYPERLINK(\"http://evil\")", want: "'-HYPERLINK(\"http://evil\")"},
		{name: "at prefix", cell: "@SUM(1+1)", want: "'@SUM(1+1)"},
		{name: "embedded equals is not prefixed", cell: "a=b", want: "a=b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sanitizeCSVCell(tt.cell)
			if got != tt.want {
				t.Fatalf("sanitizeCSVCell(%q) = %q, want %q", tt.cell, got, tt.want)
			}
		})
	}
}

func TestWriteCSVReport_NeutralizesFormulaCells(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	records := []ProgressRecord{
		{
			OwnerID:      42,
			OwnerName:    "=cmd|' /C calc'!A0",
			Status:       "failed",
			TagsAdded:    1,
			TagsFailed:   2,
			ErrorMessage: "@SUM(1+1)",
		},
	}

	if err := WriteCSVReport(&buf, records, logger); err != nil {
		t.Fatalf("WriteCSVReport: %v", err)
	}

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("parse report csv: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("row count = %d, want 2 (header + data)", len(rows))
	}
	header := rows[0]
	wantHeader := []string{"owner_id", "owner_name", "status", "tags_added", "tags_failed", "error_message"}
	if len(header) != len(wantHeader) {
		t.Fatalf("header = %v, want %v", header, wantHeader)
	}
	for i, col := range wantHeader {
		if header[i] != col {
			t.Fatalf("header[%d] = %q, want %q", i, header[i], col)
		}
	}

	row := rows[1]
	if row[0] != "42" {
		t.Fatalf("owner_id = %q, want 42", row[0])
	}
	if row[1] != "'=cmd|' /C calc'!A0" {
		t.Fatalf("owner_name was not formula-neutralized: %q", row[1])
	}
	if row[2] != "failed" {
		t.Fatalf("status = %q, want failed", row[2])
	}
	if row[5] != "'@SUM(1+1)" {
		t.Fatalf("error_message was not formula-neutralized: %q", row[5])
	}
}
