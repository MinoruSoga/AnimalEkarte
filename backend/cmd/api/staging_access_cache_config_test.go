package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Staging wrangler must not set a positive current-access cache TTL.
// The env gate in composition_auth.go stays; this test only reads the
// staging config and does not repeat that gate.
func TestStagingWranglerHasNoPositiveCurrentAccessCacheTTL(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "wrangler.jsonc"))
	require.NoError(t, err)

	var doc struct {
		Vars map[string]string `json:"vars"`
	}
	require.NoError(t, json.Unmarshal([]byte(stripJSONCComments(string(raw))), &doc))
	require.NotNil(t, doc.Vars)

	ttl, ok := doc.Vars["CURRENT_ACCESS_CACHE_TTL_SEC"]
	require.Falsef(t, ok, "staging wrangler sets CURRENT_ACCESS_CACHE_TTL_SEC=%q", ttl)
}

func stripJSONCComments(src string) string {
	var b strings.Builder
	b.Grow(len(src))
	inString := false
	escaped := false
	for i := 0; i < len(src); i++ {
		c := src[i]
		if inString {
			b.WriteByte(c)
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			b.WriteByte(c)
			continue
		}
		if c == '/' && i+1 < len(src) && src[i+1] == '/' {
			for i < len(src) && src[i] != '\n' {
				i++
			}
			if i < len(src) {
				b.WriteByte('\n')
			}
			continue
		}
		b.WriteByte(c)
	}
	return trailingComma.ReplaceAllString(b.String(), "$1")
}

var trailingComma = regexp.MustCompile(`,(\s*[}\]])`)
