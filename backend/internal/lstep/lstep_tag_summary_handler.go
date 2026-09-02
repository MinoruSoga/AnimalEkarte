package lstep

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// csvExportHeaderWriter defers CSV Content-Type/Disposition until the first body write
// so pre-stream validation errors can still return JSON via RespondError.
type csvExportHeaderWriter struct {
	gin.ResponseWriter
	onFirstWrite func()
	started      bool
}

func (w *csvExportHeaderWriter) Write(p []byte) (int, error) {
	if !w.started {
		w.started = true
		if w.onFirstWrite != nil {
			w.onFirstWrite()
		}
	}
	return w.ResponseWriter.Write(p)
}

func (w *csvExportHeaderWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// GetLstepTagSummary godoc
// GET /api/clinics/:clinic_id/lstep/tag-summary — タグ別飼い主数集計を返す（BE-020）。
func (h *Handler) GetLstepTagSummary(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	result, err := h.tagSummary.GetTagSummary(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTagSummaryResponse(result))
}

// SearchLstepOwnersByTag godoc
// GET /api/clinics/:clinic_id/lstep/owners — タグで飼い主を絞り込む（BE-020）。
// ?format=csv でCSVダウンロード。
func (h *Handler) SearchLstepOwnersByTag(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query, err := newLstepOwnersByTagQuery(c.Request.URL.Query())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if query.isCSV() {
		date := time.Now().In(time.Local).Format(time.DateOnly)
		filename := fmt.Sprintf("lstep-%s-%s.csv", query.TagName, date)
		// RFC 5987: 日本語タグ名を含むファイル名の文字化けを防ぐ（#179 ③）。
		// 非対応クライアント用に ASCII フォールバック(filename=)、対応クライアント用に
		// UTF-8 パーセントエンコード版(filename*=UTF-8'') を併記する。
		asciiFallback := fmt.Sprintf("lstep-%s.csv", date)
		// filename= は HTTP quoted-string。asciiFallback は純 ASCII（日付のみ）なので
		// %q による Go クオートは二重引用符の付与と等価で、HTTP quoted-string として安全。
		// Headers are set only after export validates and before body bytes are written
		// by ExportOwnersByTagCSV. Failures before first write still use JSON RespondError.
		// Failures after stream start must not stack a JSON body on the CSV stream.
		exportErr := h.tagSummary.ExportOwnersByTagCSV(
			c.Request.Context(),
			clinicID,
			query.TagName,
			query.NameQuery,
			&csvExportHeaderWriter{
				ResponseWriter: c.Writer,
				onFirstWrite: func() {
					c.Header("Content-Type", "text/csv; charset=utf-8")
					c.Header("Content-Disposition", fmt.Sprintf(
						"attachment; filename=%q; filename*=UTF-8''%s",
						asciiFallback, url.PathEscape(filename),
					))
				},
			},
		)
		if exportErr != nil {
			respondLstepCSVExportError(c, clinicID, query.TagName, exportErr)
			return
		}
		return
	}

	result, err := h.tagSummary.ListOwnersByTag(c.Request.Context(), clinicID, query.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTagOwnerListResponse(result))
}

func respondLstepCSVExportError(c *gin.Context, clinicID uint64, tagName string, exportErr error) {
	if !c.Writer.Written() {
		httpapi.RespondError(c, exportErr)
		return
	}
	slog.ErrorContext(c.Request.Context(), "lstep tag owner csv export failed after stream start",
		"clinic_id", clinicID, "tag", tagName, "error", exportErr)
	c.Abort()
}
