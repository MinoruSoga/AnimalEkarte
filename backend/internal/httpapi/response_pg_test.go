package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassifyPgError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		want      string
		wantKnown bool
	}{
		{
			name:      "foreign_key_violation returns 参照先が存在しません",
			err:       &pgconn.PgError{Code: "23503"},
			want:      "参照先が存在しません",
			wantKnown: true,
		},
		{
			name:      "unique_violation returns 既に登録されています",
			err:       &pgconn.PgError{Code: "23505"},
			want:      "既に登録されています",
			wantKnown: true,
		},
		{
			name:      "numeric_value_out_of_range returns 数値が範囲外です",
			err:       &pgconn.PgError{Code: "22003"},
			want:      "数値が範囲外です",
			wantKnown: true,
		},
		{
			name:      "invalid_text_representation returns 入力値の形式が正しくありません",
			err:       &pgconn.PgError{Code: "22P02"},
			want:      "入力値の形式が正しくありません",
			wantKnown: true,
		},
		{
			name:      "check_violation on chk_care_plan_item_ref returns care-plan-specific message (BUG-403)",
			err:       &pgconn.PgError{Code: "23514", ConstraintName: "chk_care_plan_item_ref"},
			want:      "選択した種別に応じたマスタ(投薬=薬剤・処置検査=処置・持ち物=入院プラン)を選択してください",
			wantKnown: true,
		},
		{
			name:      "check_violation with unrecognized constraint name falls back to generic check message",
			err:       &pgconn.PgError{Code: "23514", ConstraintName: "some_other_check"},
			want:      "入力値が制約条件を満たしていません",
			wantKnown: true,
		},
		{
			name:      "wrapped pg error is unwrapped via errors.As",
			err:       fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23505"}),
			want:      "既に登録されています",
			wantKnown: true,
		},
		// BUG-2026-07-27-01: 未知コードは「クライアント入力起因」と断定できない。
		// 42703 (undefined_column) は model と稼働 DB スキーマの乖離＝サーバ側欠陥であり、
		// 400「入力値が正しくありません」として返すと利用者のせいに見えて診断が遅れた。
		{
			name:      "undefined_column (42703) is NOT client-attributable",
			err:       &pgconn.PgError{Code: "42703"},
			want:      "",
			wantKnown: false,
		},
		{
			name:      "undefined_table (42P01) is NOT client-attributable",
			err:       &pgconn.PgError{Code: "42P01"},
			want:      "",
			wantKnown: false,
		},
		{
			name:      "unknown pg error code is NOT client-attributable",
			err:       &pgconn.PgError{Code: "99999"},
			want:      "",
			wantKnown: false,
		},
		{
			name:      "non-pg error is NOT client-attributable",
			err:       errors.New("plain error"),
			want:      "",
			wantKnown: false,
		},
		{
			name:      "nil error is NOT client-attributable",
			err:       nil,
			want:      "",
			wantKnown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotKnown := classifyPgError(tt.err)
			if gotKnown != tt.wantKnown {
				t.Errorf("classifyPgError() known = %v, want %v", gotKnown, tt.wantKnown)
			}
			if gotKnown && got != tt.want {
				t.Errorf("classifyPgError() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestResolveErrorResponse_PgErrorStatusMapping は BUG-2026-07-27-01 の回帰ゲート。
// 旧障害モード: 未知の pg エラー (42703 undefined_column) が
// HTTP 400 "入力値が正しくありません" として返り、サーバ側スキーマ欠陥が
// 「利用者の入力ミス」に見えていた。
func TestResolveErrorResponse_PgErrorStatusMapping(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "known 23503 stays 400",
			err:         &pgconn.PgError{Code: "23503"},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "参照先が存在しません",
		},
		{
			name:        "known 23505 stays 400",
			err:         &pgconn.PgError{Code: "23505"},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "既に登録されています",
		},
		{
			name:        "known 22003 stays 400",
			err:         &pgconn.PgError{Code: "22003"},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "数値が範囲外です",
		},
		{
			name:        "known 22P02 stays 400",
			err:         &pgconn.PgError{Code: "22P02"},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "入力値の形式が正しくありません",
		},
		{
			name:        "known 23514 stays 400",
			err:         &pgconn.PgError{Code: "23514", ConstraintName: "some_other_check"},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "入力値が制約条件を満たしていません",
		},
		{
			// 実障害そのもの: pets.version が稼働 DB に存在せず Find が 42703 で落ちた。
			name:        "undefined_column 42703 maps to 500, not 400",
			err:         fmt.Errorf("database error: %w", &pgconn.PgError{Code: "42703", Message: `column pets.version does not exist`}),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "internal server error",
		},
		{
			name:        "undefined_table 42P01 maps to 500, not 400",
			err:         &pgconn.PgError{Code: "42P01"},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "internal server error",
		},
		{
			name:        "unknown code 99999 maps to 500, not 400",
			err:         &pgconn.PgError{Code: "99999"},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message, _ := ResolveErrorResponse(tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if message != tt.wantMessage {
				t.Errorf("message = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}

// TestResolveErrorResponse_UnknownPgErrorDoesNotLeakInternals は、500 に落ちた
// pg エラーの応答本文へ pg メッセージ・制約名・テーブル名・SQL が混入しないことを固定する。
func TestResolveErrorResponse_UnknownPgErrorDoesNotLeakInternals(t *testing.T) {
	err := fmt.Errorf("database error: %w", &pgconn.PgError{
		Code:           "42703",
		Message:        "column pets.version does not exist",
		ConstraintName: "pets_version_check",
		TableName:      "pets",
		Detail:         "SELECT pets.version FROM pets LEFT JOIN owners",
	})

	_, message, code := ResolveErrorResponse(err)

	for _, leaked := range []string{
		"pets", "version", "column", "SELECT", "42703",
		"pets_version_check", "does not exist",
	} {
		if strings.Contains(message, leaked) {
			t.Errorf("response message %q leaks internal detail %q", message, leaked)
		}
		if strings.Contains(code, leaked) {
			t.Errorf("response code %q leaks internal detail %q", code, leaked)
		}
	}
}

func TestIsPgError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "returns true for pg error",
			err:  &pgconn.PgError{Code: "23505"},
			want: true,
		},
		{
			name: "returns true for wrapped pg error",
			err:  fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23505"}),
			want: true,
		},
		{
			name: "returns false for non-pg error",
			err:  errors.New("plain error"),
			want: false,
		},
		{
			name: "returns false for nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPgError(tt.err); got != tt.want {
				t.Errorf("isPgError() = %v, want %v", got, tt.want)
			}
		})
	}
}
