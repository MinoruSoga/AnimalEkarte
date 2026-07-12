package handler

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/config"
)

//go:embed testdata/route_snapshot.golden
var routeSnapshotGolden string

// TestRouteSnapshot は全ルートの (method, path, ハンドラ関数名) 組をゴールデンファイルと
// 比較する特性テスト（BE-refactor.md F-1）。ルート登録リファクタ（F-2/F-3）の回帰検出網 —
// handler_routes_test.go の TestRegisterRoutes_NoPanic は panic しないことしか検証しない。
//
// RequirePermission はクロージャのため r.Routes() には現れない — method+path+handler名の
// 固定だけで F-2/F-3 の回帰は検出できるため、permission までは固定しない（計画の指示通り）。
//
// このテストが意図的なルート変更で失敗した場合は testdata/route_snapshot.golden を
// 実際の出力に合わせて更新すること。
func TestRouteSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-route-snapshot",
		},
	}

	r := gin.New()
	h.RegisterRoutes(context.Background(), r)

	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, lastHandlerSegment(route.Handler)))
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	if os.Getenv("UPDATE_ROUTE_SNAPSHOT") == "1" {
		if err := os.WriteFile("testdata/route_snapshot.golden", []byte(got), 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}

	assert.Equal(t, routeSnapshotGolden, got, "route snapshot drifted — if this is an intentional route change, update testdata/route_snapshot.golden (or rerun with UPDATE_ROUTE_SNAPSHOT=1)")
}

// lastHandlerSegment は gin の RouteInfo.Handler が返す完全修飾関数名から末尾のハンドラ名
// のみを取り出す（例: ".../handler.(*Handler).GetOwner-fm" → "GetOwner"）。
// フルパスは環境（GOPATH/モジュールキャッシュの絶対パス）に依存し不安定なため使わない。
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
