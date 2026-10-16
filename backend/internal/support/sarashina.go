package support

import (
	"net/url"
	"strings"
)

// IsSarashinaURL は baseURL が信頼済み Sarashina エンドポイントかを
// パース済みホスト名の完全一致で判定する（部分文字列一致・リダイレクトは使わない）。
func IsSarashinaURL(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	// スキームに関わらず宛先ホストで判定する（未承認時は dev の http でも門を閉じる）。
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	return host == "api.sarashina.ai"
}
