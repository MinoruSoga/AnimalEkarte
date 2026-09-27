package medicalrecord

// checkup_package_assets_test.go — backend/checkup-packages/ 配下の
// 出荷済み manifest JSON が import API の strict schema を通ることを固定する。
// パース・正規化・digest 計算のみを検証する純粋テスト（DB 不要）。

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShippedCheckupPackageManifests_Validate(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "checkup-packages", "*.json"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "checkup-packages manifests must exist")

	for _, p := range paths {
		raw, err := os.ReadFile(p)
		require.NoError(t, err)

		canonical, err := ParseAndCanonicalizeCheckupPackage(raw)
		require.NoErrorf(t, err, "manifest %s must pass strict manifest validation", filepath.Base(p))
		require.NotNil(t, canonical)
		assert.NotEmptyf(t, canonical.Digest, "manifest %s digest", filepath.Base(p))
	}
}
