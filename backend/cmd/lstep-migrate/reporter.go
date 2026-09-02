package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
)

// sanitizeCSVCell はセルがスプレッドシート数式と解釈されるのを防ぐ。
// 先頭文字が = + - @ の場合、単一引用符 ' で前置する。
func sanitizeCSVCell(cell string) string {
	if cell != "" {
		switch cell[0] {
		case '=', '+', '-', '@':
			return "'" + cell
		}
	}
	return cell
}

// WriteCSVReport は進捗レコードを CSV で出力する。
func WriteCSVReport(w io.Writer, records []ProgressRecord, logger *slog.Logger) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"owner_id", "owner_name", "status", "tags_added", "tags_failed", "error_message"}); err != nil {
		return fmt.Errorf("csv header write failed: %w", err)
	}

	success, partial, failed := 0, 0, 0
	for _, r := range records {
		row := []string{
			fmt.Sprintf("%d", r.OwnerID),
			sanitizeCSVCell(r.OwnerName),
			sanitizeCSVCell(r.Status),
			fmt.Sprintf("%d", r.TagsAdded),
			fmt.Sprintf("%d", r.TagsFailed),
			sanitizeCSVCell(r.ErrorMessage),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("csv row write failed (owner=%d): %w", r.OwnerID, err)
		}
		switch r.Status {
		case "success":
			success++
		case "partial":
			partial++
		default:
			failed++
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("csv flush failed: %w", err)
	}

	logger.Info("migration complete",
		slog.Int("total", len(records)),
		slog.Int("success", success),
		slog.Int("partial", partial),
		slog.Int("failed", failed),
	)
	return nil
}
