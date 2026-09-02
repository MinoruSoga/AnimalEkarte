package lstep

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

var (
	errBlockedDialAddress  = errors.New("blocked dial address")
	errRedirectDisallowed  = errors.New("redirect disallowed")
	lookupDialIPAddr       = defaultLookupDialIPAddr
	sharedCarrierGradeNAT4 = mustCIDR("100.64.0.0/10")
)

func mustCIDR(raw string) *net.IPNet {
	_, network, err := net.ParseCIDR(raw)
	if err != nil {
		panic(err)
	}
	return network
}

func defaultLookupDialIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return net.DefaultResolver.LookupIPAddr(ctx, host)
}

func denyRedirects(*http.Request, []*http.Request) error {
	return errRedirectDisallowed
}

func newHardenedHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: denyRedirects,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           hardenedDialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          4,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   timeout,
			ExpectContinueTimeout: time.Second,
		},
	}
}

func hardenedDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := lookupDialIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errBlockedDialAddress
	}
	for _, ipa := range ips {
		if IsForbiddenDialIP(ipa.IP) {
			return nil, errBlockedDialAddress
		}
	}
	d := &net.Dialer{Timeout: defaultTimeout}
	return d.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

// IsForbiddenDialIP reports addresses that must not be dialed for LSTEP egress.
// RFC1918 / loopback / link-local / multicast are covered by net.IP helpers;
// CGNAT 100.64.0.0/10 is not IsPrivate in Go and is blocked explicitly.
func IsForbiddenDialIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsUnspecified() ||
		ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() {
		return true
	}
	v4 := ip.To4()
	return v4 != nil && sharedCarrierGradeNAT4.Contains(v4)
}
