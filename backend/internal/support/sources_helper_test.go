package support

import "encoding/json"

// sourcesPtr は保存済み sources（*json.RawMessage）のテスト用リテラルを作る。
func sourcesPtr(raw json.RawMessage) *json.RawMessage { return &raw }
