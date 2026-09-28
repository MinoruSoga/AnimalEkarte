package support

// realdb_chat_history_isolation_test.go — D3 support package (chat history)
//
// Proves support_chat_messages reads/writes stay scoped to (clinic_id, staff_id):
// rows of another clinic and of another staff member in the same clinic are
// invisible through the real repository + service + HTTP handler path.
// Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestRealDB_SupportChatHistoryClinicStaffIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.SupportChatMessage{}))
	testdb.Truncate(t, db, "support_chat_messages")

	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3support-chathistory realDB")
	ctx := context.Background()

	// clinic B の別スタッフ（同じ clinic でも他人の履歴は見えないことを検証する）
	other := &model.Staff{ClinicID: fx.ClinicB, Name: "D3support-chathistory other staff", StaffType: model.StaffTypeDoctor, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(other).Error)

	svc := NewService(NewRepository(db))

	const msgA = "D3support-chathistory-Aの質問"
	const msgB = "D3support-chathistory-Bの質問"
	const msgOther = "D3support-chathistory-他人の質問"
	require.NoError(t, svc.RecordChatExchange(ctx, fx.ClinicA, fx.StaffID, msgA, "Aの回答", nil))
	require.NoError(t, svc.RecordChatExchange(ctx, fx.ClinicB, fx.StaffID, msgB, "Bの回答",
		[]ChatSource{{Title: "画面別 会計", Category: "screens", Slug: "accounting"}}))
	require.NoError(t, svc.RecordChatExchange(ctx, fx.ClinicB, other.ID, msgOther, "他人の回答", nil))

	handler := NewHandler(svc, nil, nil, nil, nil, nil)
	const listPath = "/api/v1/support/chat/history"

	t.Run("list_returns_only_selected_clinic_and_staff_rows", func(t *testing.T) {
		// 選択clinic=B, ログイン=fx.StaffID のコンテキスト
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, listPath,
			testdb.ConfigureSelectedClinicBGrant(fx, "none", fx.ClinicB))
		handler.ChatHistory(c)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), msgB)
		assert.Contains(t, w.Body.String(), "accounting") // sources も復元される
		assert.NotContains(t, w.Body.String(), msgA)      // 同じスタッフでも他 clinic の履歴は出ない
		assert.NotContains(t, w.Body.String(), msgOther)  // 同じ clinic でも他スタッフの履歴は出ない
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.ClinicA})
	})

	t.Run("clear_deletes_only_selected_scope", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodDelete, listPath,
			testdb.ConfigureSelectedClinicBGrant(fx, "none", fx.ClinicB))
		handler.ClearChatHistory(c)
		c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder
		require.Equal(t, http.StatusNoContent, w.Code)

		// 消えたのは (B, fx.StaffID) だけ。(A, fx.StaffID) と (B, other) は残る
		remainingA, err := svc.ListChatHistory(ctx, fx.ClinicA, fx.StaffID)
		require.NoError(t, err)
		require.Len(t, remainingA, 2)
		remainingOther, err := svc.ListChatHistory(ctx, fx.ClinicB, other.ID)
		require.NoError(t, err)
		require.Len(t, remainingOther, 2)
	})
}
