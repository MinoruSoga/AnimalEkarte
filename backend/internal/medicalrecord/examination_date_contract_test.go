package medicalrecord

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCreateExaminationDateJSONContract(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{name: "JST midnight RFC3339", date: "2026-10-15T00:00:00+09:00"},
		{name: "UTC RFC3339", date: "2026-10-14T15:00:00Z"},
		{name: "date-only form value is rejected", date: "2026-10-15", wantErr: true},
		{name: "empty date is rejected", date: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			body := `{"pet_id":42,"exam_type_id":5,"date":"` + tt.date + `"}`
			c.Request = httptest.NewRequest(http.MethodPost, "/examinations", strings.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			var req createExaminationRequest
			err := c.ShouldBindJSON(&req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ShouldBindJSON error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			want, err := time.Parse(time.RFC3339, "2026-10-15T00:00:00+09:00")
			if err != nil {
				t.Fatal(err)
			}
			if !req.Date.Equal(want) {
				t.Fatalf("bound date = %s, want instant %s", req.Date.Format(time.RFC3339), want.Format(time.RFC3339))
			}
		})
	}
}
