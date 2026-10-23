package lstep

import "encoding/json"

// nullableJSONBytes は nullable jsonb 列（*json.RawMessage）を読み出し用の値に変換する。
// NULL（nil ポインタ）は nil、保存されている値はそのまま返す。呼び出し側は
// len(...) > 0 で「NULL / 空」と「JSON あり」を区別し、従来どおり前者は応答から省略する。
func nullableJSONBytes(p *json.RawMessage) json.RawMessage {
	if p == nil {
		return nil
	}
	return *p
}

// nullableJSONPtr は書き込み用に []byte を nullable jsonb 列へ変換する。
// 空（nil / len 0）は NULL（nil ポインタ）にする。空スライスへのポインタは jsonb として
// 不正（空文字列）になるため決して作らない。
func nullableJSONPtr(b []byte) *json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	m := json.RawMessage(b)
	return &m
}
