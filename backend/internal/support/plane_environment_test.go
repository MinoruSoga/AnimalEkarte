package support

// plane_environment_test.go — 信頼済み環境名の external_id 名前空間（PR650-R3）と
// 固定 label 送信（PR650-R4）のリクエスト契約。
//
// external_id は "<環境名>:<報告ID>"。環境名は APP_ENV 由来の allowlist
// （development/staging/production）だけを使い、リクエスト値・URL からは
// 導出しない。件名・説明も環境で区別できる固定文とし、STG/prod の同一 #42 が
// 衝突しないことを固定する。label は設定された 2 UUID をそのまま送り、
// 未設定では labels キーを送らない。

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestPlaneClient_EnvironmentNamespacesTicketIdentity(t *testing.T) {
	for _, env := range []string{"development", "staging", "production"} {
		t.Run(env, func(t *testing.T) {
			captured := &capturedPlaneRequest{}
			server := newPlaneTestServer(t, http.StatusCreated,
				`{"id":"issue-1","identifier":"EMR-7","sequence_id":7}`, captured)
			defer server.Close()

			client := newTestPlaneClient(server, "EMR")
			client.environment = env
			client.labelIDs = []string{"label-src", "label-lane"}

			_, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 7, Title: "t"})
			require.NoError(t, err)

			assert.Equal(t, env+":7", captured.body["external_id"])
			name, _ := captured.body["name"].(string)
			assert.Contains(t, name, env)
			assert.Contains(t, name, "#7")
			desc, _ := captured.body["description_html"].(string)
			assert.Contains(t, desc, env)
			assert.Contains(t, desc, "#7")
			assert.Equal(t, []any{"label-src", "label-lane"}, captured.body["labels"])
		})
	}
}

// STG/prod の同一 #42 は external_id・件名・説明のすべてで区別される。
func TestPlaneClient_EnvironmentDistinguishesSameReportIDAcrossEnvs(t *testing.T) {
	bodies := make([]map[string]any, 0, 2)
	for _, env := range []string{"staging", "production"} {
		captured := &capturedPlaneRequest{}
		server := newPlaneTestServer(t, http.StatusCreated,
			`{"id":"issue-1","identifier":"EMR-42","sequence_id":42}`, captured)

		client := newTestPlaneClient(server, "EMR")
		client.environment = env
		_, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 42, Title: "t"})
		require.NoError(t, err)
		server.Close()
		bodies = append(bodies, captured.body)
	}

	assert.NotEqual(t, bodies[0]["external_id"], bodies[1]["external_id"])
	assert.NotEqual(t, bodies[0]["name"], bodies[1]["name"])
	assert.NotEqual(t, bodies[0]["description_html"], bodies[1]["description_html"])
}

// allowlist 外・空の環境名は development に正規化する（厳格な拒否は
// config.Validate の責務 — クライアントは外部入力を素通ししない）。
func TestPlaneClient_UnknownEnvironmentNormalizesToDevelopment(t *testing.T) {
	for _, env := range []string{"", "test", "E2E"} {
		captured := &capturedPlaneRequest{}
		server := newPlaneTestServer(t, http.StatusCreated,
			`{"id":"issue-1","identifier":"EMR-7","sequence_id":7}`, captured)

		client := newTestPlaneClient(server, "EMR")
		client.environment = env
		_, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 7, Title: "t"})
		require.NoError(t, err)
		server.Close()

		assert.Equal(t, "development:7", captured.body["external_id"], "environment %q", env)
	}
}

// 正規化は trim/lower 後に allowlist 判定する。
func TestPlaneClient_EnvironmentNormalization(t *testing.T) {
	captured := &capturedPlaneRequest{}
	server := newPlaneTestServer(t, http.StatusCreated,
		`{"id":"issue-1","identifier":"EMR-7","sequence_id":7}`, captured)
	defer server.Close()

	client := newTestPlaneClient(server, "EMR")
	client.environment = " Production "
	_, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 7, Title: "t"})
	require.NoError(t, err)
	assert.Equal(t, "production:7", captured.body["external_id"])
}

// label 未設定では labels キーを送らない（config 側の enabled 検証とは別に
// クライアントが空の label 配列を送りつけないことを固定する）。
func TestPlaneClient_LabelsOmittedWhenNotConfigured(t *testing.T) {
	captured := &capturedPlaneRequest{}
	server := newPlaneTestServer(t, http.StatusCreated,
		`{"id":"issue-1","identifier":"EMR-7","sequence_id":7}`, captured)
	defer server.Close()

	client := newTestPlaneClient(server, "EMR")
	_, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 7, Title: "t"})
	require.NoError(t, err)
	_, ok := captured.body["labels"]
	assert.False(t, ok, "labels key must be absent when no label IDs are configured")
}

// NewPlaneTicketCreator はオプションで環境名と label ID を受け取る。
func TestNewPlaneTicketCreator_AppliesOptions(t *testing.T) {
	creator := NewPlaneTicketCreator(
		"", "", "key", "ws", "proj", "EMR", "",
		time.Second,
		WithPlaneEnvironment("staging"),
		WithPlaneLabelIDs("src-uuid", "lane-uuid"),
	)
	require.NotNil(t, creator)
	client, ok := creator.(*planeClient)
	require.True(t, ok)
	assert.Equal(t, "staging", client.environment)
	assert.Equal(t, []string{"src-uuid", "lane-uuid"}, client.labelIDs)
}

// 片方だけの label 設定は採用しない（片ラベルだけの起票を防ぐ）。
func TestNewPlaneTicketCreator_IgnoresPartialLabelOption(t *testing.T) {
	creator := NewPlaneTicketCreator(
		"", "", "key", "ws", "proj", "EMR", "",
		time.Second,
		WithPlaneLabelIDs("src-uuid", ""),
	)
	require.NotNil(t, creator)
	client, ok := creator.(*planeClient)
	require.True(t, ok)
	assert.Empty(t, client.labelIDs)
}
