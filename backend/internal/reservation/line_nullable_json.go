package reservation

import (
	"bytes"
	"encoding/json"
)

// nullableJSONBytes は nullable jsonb 列（*json.RawMessage）を読み出し用の値に変換する。
// NULL（nil ポインタ）は nil。呼び出し側は len(...) > 0 で「NULL / 空」と「JSON あり」を区別する。
func nullableJSONBytes(p *json.RawMessage) json.RawMessage {
	if p == nil {
		return nil
	}
	return *p
}

// nullableJSONPtr は書き込み用に []byte を nullable jsonb 列へ変換する。
// 空（nil / len 0 / 空白のみ）と JSON リテラル null は SQL NULL（nil ポインタ）にする
// （business_hours_by_weekday の null は曜日 override の全解除を意味し、この全解除以外で null は不可:
// docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md §3.1）。空スライスへのポインタは
// jsonb として不正（空文字列）になるため決して作らない。
func nullableJSONPtr(b []byte) *json.RawMessage {
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	m := json.RawMessage(b)
	return &m
}
