//go:build tools

package tools

import (
	// oapi-codegen: api.yaml (OpenAPI 3.0) から Go 型を生成するツール
	// 使用方法: go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config oapi-codegen.yaml docs/api.yaml
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
