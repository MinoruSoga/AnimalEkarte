package support

import "testing"

func TestIsSarashinaURL(t *testing.T) {
	cases := map[string]bool{
		"https://api.sarashina.ai/v1":          true,
		"http://api.sarashina.ai/v1":           true, // 未承認時は dev http でも検出
		"https://API.SARASHINA.AI/v1":          true,
		"https://api.sarashina.ai:8443/v1":     true,
		"https://api.sarashina.ai./v1":         true,
		"https://api.sarashina.ai.evil.com/v1": false,
		"https://evil.com/api.sarashina.ai":    false,
		"https://api.sarashina.ai.evil":        false,
		"":                                     false,
		"not a url":                            false,
	}
	for u, want := range cases {
		if got := IsSarashinaURL(u); got != want {
			t.Errorf("IsSarashinaURL(%q)=%v want %v", u, got, want)
		}
	}
}
