package pet

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// withJSTLocal は本番 (config.ConfigureTimeZone) 同様に time.Local を JST(+09:00) に固定する。
// UTC ランナー上では localTimePtr が no-op になり経路間差分を再現できないため、
// クロスエンドポイント不整合を決定的に検出するには明示固定が必須。
//
// 注意: time.Local はパッケージグローバルなため本ヘルパーを使うテストは t.Parallel() 禁止。
// 並列化すると global 書き換えが競合し go test -race で DATA RACE を検出する。
func withJSTLocal(t *testing.T) {
	t.Helper()
	orig := time.Local
	time.Local = config.JST
	t.Cleanup(func() { time.Local = orig })
}

func marshalJSONField(t *testing.T, v *time.Time) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

// TestPetDateFieldsSerializeIdenticallyAcrossEndpoints は、同一 pet の
// birth_date / neutered_date / last_visit が pet詳細 (toResponse)・
// pet一覧 (toPetListResponse) の両経路で byte 一致のシリアライズを返すことを保証する (SC1)。
//
// 規約 (canonical): *time.Time 日付フィールドは localTimePtr で time.Local へ
// 変換してからマーシャルする。owner 経路はこれに準拠済みで正、pet 経路は raw で誤。
// 修正前は detail/list が `…T00:00:00Z`、nested が `…T09:00:00+09:00` を返し RED。
func TestPetDateFieldsSerializeIdenticallyAcrossEndpoints(t *testing.T) {
	withJSTLocal(t)

	// pgx/v5 は Postgres date 列を UTC 深夜の time.Time として読み出す
	// (格納形式 = ローカル暦日)。これが本番で各 toXxxResponse に渡る実値。
	birthDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
	neuteredDate := time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	pet := &model.Pet{
		ID:              7,
		OwnerID:         42,
		AnimalSpeciesID: 3,
		PetNumber:       "42-1",
		Name:            "ポチ",
		Gender:          model.PetGenderMale,
		Status:          model.PetStatusAlive,
		BirthDate:       &birthDate,
		NeuteredDate:    &neuteredDate,
		LastVisit:       &lastVisit,
	}

	detail := toResponse(pet)
	list := toPetListResponse(pet)

	// canonical の期待 byte をハードコードで固定する (循環参照回避)。
	// 両経路を独立に同一の期待値と照合する。
	// 深夜 UTC の date 値を JST へ変換すると同日 09:00+09:00 になる。
	const (
		wantBirth    = `"2020-01-15T09:00:00+09:00"`
		wantNeutered = `"2019-12-31T09:00:00+09:00"`
		wantLast     = `"2021-01-01T09:00:00+09:00"`
	)

	for _, p := range []struct {
		label                      string
		birth, neutered, lastVisit *time.Time
	}{
		{"pet詳細(toResponse)", detail.BirthDate, detail.NeuteredDate, detail.LastVisit},
		{"pet一覧(toPetListResponse)", list.BirthDate, list.NeuteredDate, list.LastVisit},
	} {
		assert.Equal(t, wantBirth, marshalJSONField(t, p.birth),
			"%s の birth_date は canonical 表現でなければならない", p.label)
		assert.Equal(t, wantNeutered, marshalJSONField(t, p.neutered),
			"%s の neutered_date は canonical 表現でなければならない (同一 root cause)", p.label)
		assert.Equal(t, wantLast, marshalJSONField(t, p.lastVisit),
			"%s の last_visit は canonical 表現でなければならない (同一 root cause)", p.label)
	}
}

// TestPetBirthDateCalendarDateNeverRolls は、tz 変換でカレンダー日付が
// ずれないこと (SC2) を、年/月/閏日境界の格納値で全経路について保証する。
// date 列は常に深夜 UTC で読み出されるため、JST(+09:00) 変換は同日内 09:00 へ
// 前進するのみで日付を繰り上げない。負オフセット帯ではないため繰り下げも無い。
func TestPetBirthDateCalendarDateNeverRolls(t *testing.T) {
	withJSTLocal(t)

	cases := []struct {
		name     string
		in       time.Time
		wantDate string
	}{
		{"midyear", time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC), "2020-01-15"},
		{"year_boundary_dec31", time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC), "2019-12-31"},
		{"year_boundary_jan1", time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), "2021-01-01"},
		{"leap_day", time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC), "2020-02-29"},
		{"month_boundary", time.Date(2020, 3, 31, 0, 0, 0, 0, time.UTC), "2020-03-31"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bd := tc.in
			pet := &model.Pet{BirthDate: &bd}

			detail := toResponse(pet)
			list := toPetListResponse(pet)

			require.NotNil(t, detail.BirthDate)
			require.NotNil(t, list.BirthDate)

			assert.Equal(t, tc.wantDate, detail.BirthDate.Format("2006-01-02"),
				"pet詳細のカレンダー日付がずれてはならない")
			assert.Equal(t, tc.wantDate, list.BirthDate.Format("2006-01-02"),
				"pet一覧のカレンダー日付がずれてはならない")
		})
	}
}
