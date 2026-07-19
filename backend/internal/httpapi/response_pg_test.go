package httpapi

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassifyPgError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "foreign_key_violation returns 参照先が存在しません",
			err:  &pgconn.PgError{Code: "23503"},
			want: "参照先が存在しません",
		},
		{
			name: "unique_violation returns 既に登録されています",
			err:  &pgconn.PgError{Code: "23505"},
			want: "既に登録されています",
		},
		{
			name: "numeric_value_out_of_range returns 数値が範囲外です",
			err:  &pgconn.PgError{Code: "22003"},
			want: "数値が範囲外です",
		},
		{
			name: "invalid_text_representation returns 入力値の形式が正しくありません",
			err:  &pgconn.PgError{Code: "22P02"},
			want: "入力値の形式が正しくありません",
		},
		{
			name: "unknown pg error code falls back to default message",
			err:  &pgconn.PgError{Code: "99999"},
			want: "入力値が正しくありません",
		},
		{
			name: "check_violation on chk_care_plan_item_ref returns care-plan-specific message (BUG-403)",
			err:  &pgconn.PgError{Code: "23514", ConstraintName: "chk_care_plan_item_ref"},
			want: "選択した種別に応じたマスタ(投薬=薬剤・処置検査=処置・持ち物=入院プラン)を選択してください",
		},
		{
			name: "check_violation with unrecognized constraint name falls back to generic check message",
			err:  &pgconn.PgError{Code: "23514", ConstraintName: "some_other_check"},
			want: "入力値が制約条件を満たしていません",
		},
		{
			name: "wrapped pg error is unwrapped via errors.As",
			err:  fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23505"}),
			want: "既に登録されています",
		},
		{
			name: "non-pg error falls back to default message",
			err:  errors.New("plain error"),
			want: "入力値が正しくありません",
		},
		{
			name: "nil error falls back to default message",
			err:  nil,
			want: "入力値が正しくありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyPgError(tt.err)
			if got != tt.want {
				t.Errorf("classifyPgError() = %q, want %q", got, tt.want)
			}
		})
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
