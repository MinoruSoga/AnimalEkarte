package pet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// TestFirstVisitDateSerializesConsistentlyWithMedicalRecordDate は、pet 初診日
// エンドポイント (toPetFirstVisitResponse) が、同一の MedicalRecord.date 値を
// medical_record 詳細エンドポイント (medical_record_response.go の localTime(r.Date)) と
// byte 一致でシリアライズすることを保証する (SC1 クロスエンドポイント)。
//
// 規約 (canonical): MedicalRecord.date は medical_record 経路で localTime(r.Date) として
// datetime 配信される。初診日 (first_visit_date) は同じ date 値の派生であり、*time.Time のため
// localTimePtr で time.Local へ変換してから配信しなければならない。修正前は初診日経路が
// raw (`…T00:00:00Z`)、medical_record 経路が localTime (`…+09:00`) を返し経路間で割れて RED。
// first_visit_date は openapi 未宣言 (型 datetime 制約なし) のため tz 表現のみ是正する。
//
// withJSTLocal / marshalJSONField は pet_birthdate_consistency_test.go と共有 (再定義しない)。
func TestFirstVisitDateSerializesConsistentlyWithMedicalRecordDate(t *testing.T) {
	withJSTLocal(t)

	// pgx/v5 は Postgres date 列を UTC 深夜の time.Time として読み出す。
	// これが本番で GetFirstVisitDate / medical_record 詳細の双方に渡る実値。
	recordDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	// 初診日経路 (修正対象)
	firstVisit := toPetFirstVisitResponse(&recordDate)
	require.NotNil(t, firstVisit.FirstVisitDate)

	// medical_record 詳細経路の canonical シリアライズ = localTime(r.Date)。
	// 深夜 UTC の date 値を JST へ変換すると同日 09:00+09:00 になる。
	const want = `"2026-06-30T09:00:00+09:00"`
	canonical := httpapi.LocalTime(recordDate)

	assert.Equal(t, want, marshalJSONField(t, &canonical),
		"medical_record 経路の date は canonical (localTime) 表現でなければならない")
	assert.Equal(t, want, marshalJSONField(t, firstVisit.FirstVisitDate),
		"first_visit_date は medical_record 経路と byte 一致でなければならない")
}

// TestFirstVisitDateNullWhenNoRecord は、カルテが無い (nil) 場合に初診日が
// null のまま保たれる (捏造しない) ことを保証する。localTimePtr は nil を素通しする。
func TestFirstVisitDateNullWhenNoRecord(t *testing.T) {
	withJSTLocal(t)

	resp := toPetFirstVisitResponse(nil)
	assert.Nil(t, resp.FirstVisitDate, "カルテが無い場合 first_visit_date は null でなければならない")
}

// TestFirstVisitDateCalendarDateNeverRolls は tz 変換で初診日のカレンダー日付が
// ずれないこと (SC2) を年/月/閏日境界の格納値で保証する。date 列は常に深夜 UTC で
// 読み出されるため JST(+09:00) 変換は同日 09:00 へ前進するのみで日付を繰り上げない。
func TestFirstVisitDateCalendarDateNeverRolls(t *testing.T) {
	withJSTLocal(t)

	cases := []struct {
		name     string
		in       time.Time
		wantDate string
	}{
		{"midyear", time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "2026-06-30"},
		{"year_boundary_dec31", time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-12-31"},
		{"year_boundary_jan1", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), "2027-01-01"},
		{"leap_day", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), "2024-02-29"},
		{"month_boundary", time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC), "2026-03-31"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.in
			resp := toPetFirstVisitResponse(&in)
			require.NotNil(t, resp.FirstVisitDate)
			assert.Equal(t, tc.wantDate, resp.FirstVisitDate.Format("2006-01-02"),
				"初診日のカレンダー日付がずれてはならない")
		})
	}
}
