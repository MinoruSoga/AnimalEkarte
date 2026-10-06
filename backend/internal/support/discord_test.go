package support

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestNewDiscordNotifier_EmptyURLDisabled(t *testing.T) {
	assert.Nil(t, NewDiscordNotifier("", 5*time.Second))
	assert.Nil(t, NewDiscordNotifier("   ", 5*time.Second))
}

func TestDiscordNotifier_NotifyBugReport(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &received))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier := NewDiscordNotifier(server.URL, 5*time.Second)
	report := &model.SupportBugReport{
		ClinicID:        3,
		ReporterStaffID: 7,
		Title:           "受付でエラー @everyone",
		Detail:          "保存ボタンを押すと500になる",
		RoutePath:       "/reception",
		PageURL:         "https://app.example.com/reception",
		UserAgent:       "UA-test",
		Viewport:        "1280x720",
		AppVersion:      "1.2.3",
	}
	require.NoError(t, notifier.NotifyBugReport(context.Background(), report))

	require.NotNil(t, received)
	assert.Equal(t, "AnimalEkarte バグ報告", received["username"])

	mentions, ok := received["allowed_mentions"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, mentions["parse"], "報告文面中のメンションが Discord 側で発火しないこと")

	embeds, ok := received["embeds"].([]any)
	require.True(t, ok)
	require.Len(t, embeds, 1)
	embed := embeds[0].(map[string]any)
	assert.Contains(t, embed["title"], "受付でエラー")
	assert.Equal(t, "保存ボタンを押すと500になる", embed["description"])
	assert.NotEmpty(t, embed["timestamp"])

	fields, ok := embed["fields"].([]any)
	require.True(t, ok)
	byName := make(map[string]string, len(fields))
	for _, f := range fields {
		fm := f.(map[string]any)
		byName[fm["name"].(string)] = fm["value"].(string)
	}
	assert.Equal(t, "clinic#3", byName["医院"])
	assert.Equal(t, "staff#7", byName["報告者"])
	assert.Equal(t, "/reception", byName["画面"])
	assert.Equal(t, "なし", byName["スクリーンショット"])
}

func TestDiscordNotifier_UpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := NewDiscordNotifier(server.URL, 5*time.Second)
	err := notifier.NotifyBugReport(context.Background(), &model.SupportBugReport{Title: "t", Detail: "d"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestDiscordNotifier_TruncatesOversizedContent(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier := NewDiscordNotifier(server.URL, 5*time.Second)
	report := &model.SupportBugReport{
		Title:  strings.Repeat("あ", 500),
		Detail: strings.Repeat("詳", 5000),
	}
	require.NoError(t, notifier.NotifyBugReport(context.Background(), report))

	embed := received["embeds"].([]any)[0].(map[string]any)
	assert.LessOrEqual(t, len([]rune(embed["title"].(string))), discordEmbedTitleMaxRunes)
	assert.LessOrEqual(t, len([]rune(embed["description"].(string))), discordEmbedDescriptionMaxRunes)
}
