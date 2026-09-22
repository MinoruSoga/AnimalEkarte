package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CurrentAccess のキャッシュは権限/clinic 無効化の反映をTTL分遅らせるため、
// CURRENT_ACCESS_CACHE_TTL_SEC の env ゲート経由でのみ有効化されなければ
// ならない(3148d229f の認可ギャップ修正を無条件配線で再発させない)。
func TestAuthCompositionCachesCurrentAccessOnlyBehindEnvGate(t *testing.T) {
	source, err := os.ReadFile("composition_auth.go")
	require.NoError(t, err)
	src := string(source)
	assert.Contains(t, src, "CURRENT_ACCESS_CACHE_TTL_SEC")
	// 無条件配線の形(キャッシュを変数・フィールドへ直接代入)を禁止する。
	// 許容されるのは currentAccessCacheTTL() > 0 の条件ブロック内のみ。
	assert.NotContains(t, src, "currentAccess: auth.NewCachedCurrentAccessResolver")
	assert.NotContains(t, src, "currentAccess:    auth.NewCachedCurrentAccessResolver")
	abs, err := filepath.Abs("composition_auth.go")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(abs, "composition_auth.go"))
}

func TestCurrentAccessCacheTTL(t *testing.T) {
	t.Run("empty disables cache", func(t *testing.T) {
		t.Setenv("CURRENT_ACCESS_CACHE_TTL_SEC", "")
		assert.Equal(t, time.Duration(0), currentAccessCacheTTL())
	})
	t.Run("non-numeric disables cache", func(t *testing.T) {
		t.Setenv("CURRENT_ACCESS_CACHE_TTL_SEC", "abc")
		assert.Equal(t, time.Duration(0), currentAccessCacheTTL())
	})
	t.Run("zero or negative disables cache", func(t *testing.T) {
		t.Setenv("CURRENT_ACCESS_CACHE_TTL_SEC", "0")
		assert.Equal(t, time.Duration(0), currentAccessCacheTTL())
		t.Setenv("CURRENT_ACCESS_CACHE_TTL_SEC", "-5")
		assert.Equal(t, time.Duration(0), currentAccessCacheTTL())
	})
	t.Run("positive seconds enables cache", func(t *testing.T) {
		t.Setenv("CURRENT_ACCESS_CACHE_TTL_SEC", "30")
		assert.Equal(t, 30*time.Second, currentAccessCacheTTL())
	})
}
