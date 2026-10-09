package medicalrecord

// medical_record_addendum_response_test.go — 追記応答 DTO の author_name マッピング契約。
//
// 保護する不変条件:
//   - AuthorStaff が解決できた場合のみ author_name に氏名が載る。
//   - AuthorStaff nil（削除・異動・他医院）または氏名が空文字の場合、author_name は
//     nil のまま — JSON に "author_name" キー自体を出さない（omitempty）。
//     FE は author_name ?? "スタッフID: <author_user_id>" でフォールバックするため、
//     空文字を送ると ID 表示が消えてしまう。

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func makeAddendumModel(authorStaff *model.Staff) *model.MedicalRecordAddendum {
	return &model.MedicalRecordAddendum{
		ID:              1,
		MedicalRecordID: 10,
		ClinicID:        1,
		AuthorUserID:    5,
		BeforeText:      "before",
		AfterText:       "after",
		Reason:          "誤記修正",
		CreatedAt:       time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		AuthorStaff:     authorStaff,
	}
}

func TestToMedicalRecordAddendumResponse_AuthorName(t *testing.T) {
	t.Run("AuthorStaff解決時は氏名がauthor_nameに載る", func(t *testing.T) {
		res := toMedicalRecordAddendumResponse(makeAddendumModel(&model.Staff{ID: 5, ClinicID: 1, Name: "田中 太郎"}))
		require.NotNil(t, res.AuthorName)
		assert.Equal(t, "田中 太郎", *res.AuthorName)
	})

	t.Run("AuthorStaff未解決（削除・異動・他医院）はauthor_nameがnil", func(t *testing.T) {
		res := toMedicalRecordAddendumResponse(makeAddendumModel(nil))
		assert.Nil(t, res.AuthorName)
		assert.Equal(t, uint64(5), res.AuthorUserID)
	})

	t.Run("氏名が空文字のスタッフはauthor_nameを送らない（IDフォールバックを潰さない）", func(t *testing.T) {
		res := toMedicalRecordAddendumResponse(makeAddendumModel(&model.Staff{ID: 5, ClinicID: 1, Name: ""}))
		assert.Nil(t, res.AuthorName)
	})
}

func TestToMedicalRecordAddendumResponse_AuthorNameJSONContract(t *testing.T) {
	t.Run("解決時はauthor_nameキーが出力される", func(t *testing.T) {
		res := toMedicalRecordAddendumResponse(makeAddendumModel(&model.Staff{ID: 5, ClinicID: 1, Name: "田中 太郎"}))
		b, err := json.Marshal(res)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		assert.Equal(t, "田中 太郎", m["author_name"])
	})

	t.Run("未解決時はauthor_nameキー自体が省略される", func(t *testing.T) {
		res := toMedicalRecordAddendumResponse(makeAddendumModel(nil))
		b, err := json.Marshal(res)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		_, ok := m["author_name"]
		assert.False(t, ok, "author_name must be omitted, not null/empty, so FE ?? fallback renders author_user_id")
		assert.Equal(t, float64(5), m["author_user_id"])
	})
}

func TestToMedicalRecordAddendumListResponse_AuthorName(t *testing.T) {
	addenda := []*model.MedicalRecordAddendum{
		makeAddendumModel(&model.Staff{ID: 5, ClinicID: 1, Name: "田中 太郎"}),
		makeAddendumModel(nil),
	}
	res := toMedicalRecordAddendumListResponse(addenda)
	require.Len(t, res, 2)
	require.NotNil(t, res[0].AuthorName)
	assert.Equal(t, "田中 太郎", *res[0].AuthorName)
	assert.Nil(t, res[1].AuthorName)
}
