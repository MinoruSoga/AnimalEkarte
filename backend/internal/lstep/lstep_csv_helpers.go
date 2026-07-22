package lstep

import (
	"fmt"
	"strings"
	"time"
)

// csvColumnAliases は実ヘッダー名（大文字小文字・空白を許容）を内部 key にマップする。
// 仕様未確定のため想定 alias で実装。実 CSV 取得後に Phase 4 テストで検証する。
var csvColumnAliases = map[string][]string{
	"line_user_id":    {"line id", "line_user_id", "userid", "user id"},
	"display_name":    {"表示名", "display_name", "name"},
	"registered_at":   {"登録日", "registered_at", "registered"},
	"tags":            {"タグ", "tags"},
	"scenarios":       {"シナリオ", "scenarios"},
	"traffic_source":  {"流入元", "traffic_source", "source"},
	"block_status":    {"ブロック", "block_status", "block"},
	"last_message_at": {"最終メッセージ", "last_message_at", "last_message"},
}

// resolveCsvHeaders はヘッダー行から内部 key → 列インデックスのマップを返す。
// 必須: line_user_id。見つからない場合 error。
func resolveCsvHeaders(headers []string) (map[string]int, error) {
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}
	found := make(map[string]int)
	for idx, raw := range headers {
		norm := normalize(raw)
		for key, aliases := range csvColumnAliases {
			for _, alias := range aliases {
				if norm == alias {
					if _, exists := found[key]; !exists {
						found[key] = idx
					}
					break
				}
			}
		}
	}
	if _, ok := found["line_user_id"]; !ok {
		return nil, fmt.Errorf("required column not found: line_user_id (expected one of: LINE ID, line_user_id, userId)")
	}
	return found, nil
}

// splitMultiValue は tags / scenarios の複数値カラムを ; or , で分割する。
func splitMultiValue(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.NewReplacer(";", ",").Replace(s)
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// csvImportErrorEntry は error_log の 1 エントリ。
type csvImportErrorEntry struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
	Detail string `json:"detail,omitempty"`
}

const (
	csvErrorReasonUnknownLineUserID = "unknown_line_user_id"
	csvErrorReasonParseError        = "parse_error"
	csvErrorReasonMissingLineID     = "missing_line_id"
)

// csvDateFormats は Lステップ CSV で使われる日時フォーマット候補。
var csvDateFormats = []string{
	time.DateOnly,
	"2006/01/02",
	time.DateTime,
	"2006/01/02 15:04:05",
	"2006-01-02T15:04:05",
}

// parseCsvTime は CSV の日時文字列をパースする。パース失敗時は nil を返す（フィールドスキップ）。
func parseCsvTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, format := range csvDateFormats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return &t
		}
	}
	return nil
}
