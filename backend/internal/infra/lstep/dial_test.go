package lstep

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsForbiddenDialIP(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		forbidden bool
	}{
		{name: "nil", forbidden: true},
		{name: "loopback v4", ip: "127.0.0.1", forbidden: true},
		{name: "loopback v6", ip: "::1", forbidden: true},
		{name: "unspecified", ip: "0.0.0.0", forbidden: true},
		{name: "rfc1918 10/8", ip: "10.1.2.3", forbidden: true},
		{name: "link-local", ip: "169.254.1.1", forbidden: true},
		{name: "cgnat 100.64.0.1", ip: "100.64.0.1", forbidden: true},
		{name: "cgnat 100.127.255.255", ip: "100.127.255.255", forbidden: true},
		{name: "just below cgnat", ip: "100.63.255.255", forbidden: false},
		{name: "public resolver", ip: "8.8.8.8", forbidden: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ip net.IP
			if tt.ip != "" {
				ip = net.ParseIP(tt.ip)
				require.NotNil(t, ip)
			}
			assert.Equal(t, tt.forbidden, IsForbiddenDialIP(ip))
		})
	}
}

func TestHardenedDialContext_RejectsForbiddenResolvedAddresses(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{name: "loopback", ip: "127.0.0.1"},
		{name: "private", ip: "10.0.0.1"},
		{name: "cgnat", ip: "100.64.1.2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := lookupDialIPAddr
			lookupDialIPAddr = func(context.Context, string) ([]net.IPAddr, error) {
				return []net.IPAddr{{IP: net.ParseIP(tt.ip)}}, nil
			}
			t.Cleanup(func() { lookupDialIPAddr = orig })

			_, err := hardenedDialContext(context.Background(), "tcp", "api.lstep.jp:443")
			assert.ErrorIs(t, err, errBlockedDialAddress)
		})
	}
}
