package reservation

// service/be9_2d_mock_carriers_test.go の同名ヘルパーの最小限の複製（package跨ぎimport不能のため）。
func strPtr(s string) *string { return &s }

func ptrString(s string) *string { return &s }

func ptrUint64(v uint64) *uint64 { return &v }
