package manualarticle

// raw_body_guard.go — EMR-502 既存記事 edit 専用の raw body guard。
// global SanitizeNullBytes より前に登録し、sanitizer が raw 0x00 を除去する
// 「前」に byte 単位で検査する。manual 系の他 route・他 domain の挙動は変更しない。
//
// 対象は PUT /api/v1/manual/articles/:category/:slug のみ（method + FullPath 一致、
// Content-Type 不問・chunked body も対象）。raw NUL (0x00) または JSON 構文不正は
// claim 前 HTTP 400。上限超過は既存 sanitizer と同じ 413。検査した bytes は
// 書換えずそのまま後段へ戻す（別 DTO / command の差込みはしない）。

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// manualEditRoutePath は guard 対象の登録済み route pattern。
// gin.FullPath() は実 path ではなく pattern を返すため literal 一致できる。
const manualEditRoutePath = "/api/v1/manual/articles/:category/:slug"

// RawBodyGuard は PUT manual article edit 専用の raw body 検査 middleware。
// maxBytes は既存 global sanitizer と同じ上限を composition 側から渡す
// （manualarticle は internal/middleware を import しない）。
// 一致しない request は body を読まず即時素通しする。
func RawBodyGuard(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPut || c.FullPath() != manualEditRoutePath {
			c.Next()
			return
		}
		raw, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes))
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("request body exceeds size limit"))
			} else {
				httpapi.RespondError(c, apperrors.WrapInvalidInput("request body unreadable"))
			}
			c.Abort()
			return
		}
		if bytes.IndexByte(raw, 0) >= 0 {
			httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid manual article edit request"))
			c.Abort()
			return
		}
		if !json.Valid(raw) {
			httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid manual article edit request"))
			c.Abort()
			return
		}
		// 検査した同じ bytes を書換えず後段へ戻す。sanitizer / handler が
		// 同一 byte 列を見る。
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))
		c.Request.ContentLength = int64(len(raw))
		c.Next()
	}
}
