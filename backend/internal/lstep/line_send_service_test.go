package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lineinfra "github.com/animal-ekarte/backend/internal/infra/line"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock LineSendLogRepository ----

type mockLineSendLogRepo struct {
	createFn      func(ctx context.Context, log *model.LineSendLog) error
	findByOwnerFn func(ctx context.Context, clinicID, ownerID uint64, limit int) ([]*model.LineSendLog, error)
}

func (m *mockLineSendLogRepo) Create(ctx context.Context, log *model.LineSendLog) error {
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}
func (m *mockLineSendLogRepo) FindByOwner(ctx context.Context, clinicID, ownerID uint64, limit int) ([]*model.LineSendLog, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID, limit)
	}
	return []*model.LineSendLog{}, nil
}

// ---- mock SharedFileService ----

// mockSharedFileSvc は SharedFileService のテスト用モック。GetSignedURL のみ設定可能
// （line_send_service.go の Send は GetSignedURL しか呼ばないため他は固定挙動のままでよい）。
type mockSharedFileSvc struct {
	getSignedURLFn func(ctx context.Context, clinicID, id uint64) (string, error)
}

func (m *mockSharedFileSvc) GetSignedURL(ctx context.Context, clinicID, id uint64) (string, error) {
	if m.getSignedURLFn != nil {
		return m.getSignedURLFn(ctx, clinicID, id)
	}
	return "", nil
}

// ---- mock lineinfra.MessagingClient ----

// mockLineMessagingClient は lineinfra.MessagingClient のテスト用モック。
// line_send_service.go の Send は本来 lineinfra.NewMessagingClient で実際に LINE の
// Messaging API へ HTTP 通信する — テストでネットワークへ到達させないため、
// lineSendService.newLineClient（テスト容易性のためのシーム）経由でこのモックへ差し替える。
type mockLineMessagingClient struct {
	pushTextErr     error
	pushImageURLErr error
	pushFileURLErr  error
	pushTextFn      func(ctx context.Context, lineUserID, text string) error
	pushImageURLFn  func(ctx context.Context, lineUserID, originalURL, previewURL string) error
	pushFileURLFn   func(ctx context.Context, lineUserID, fileURL, altText string) error
}

func (m *mockLineMessagingClient) PushText(ctx context.Context, lineUserID, text string) error {
	if m.pushTextFn != nil {
		return m.pushTextFn(ctx, lineUserID, text)
	}
	return m.pushTextErr
}
func (m *mockLineMessagingClient) PushImageURL(ctx context.Context, lineUserID, originalURL, previewURL string) error {
	if m.pushImageURLFn != nil {
		return m.pushImageURLFn(ctx, lineUserID, originalURL, previewURL)
	}
	return m.pushImageURLErr
}
func (m *mockLineMessagingClient) PushFileURL(ctx context.Context, lineUserID, fileURL, altText string) error {
	if m.pushFileURLFn != nil {
		return m.pushFileURLFn(ctx, lineUserID, fileURL, altText)
	}
	return m.pushFileURLErr
}

var _ lineinfra.MessagingClient = (*mockLineMessagingClient)(nil)

// ---- helpers ----

func newLineSendSvc(ownerRepo lstepOwnerRepo, logRepo LineSendLogRepository) LineSendService {
	return NewLineSendService(
		&mockLstepSettingsService{},
		ownerRepo,
		&mockSharedFileSvc{},
		&mockLstepTagCacheRepository{},
		&mockAuditService{},
		logRepo,
		nil, // tagConfigRepo
	)
}

// lineSendTestDeps は Send() の全依存関係を差し替え可能にするテスト用の構成一式。
type lineSendTestDeps struct {
	settings      *mockLstepSettingsService
	ownerRepo     lstepOwnerRepo
	sharedFile    *mockSharedFileSvc
	tagCacheRepo  *mockLstepTagCacheRepository
	auditSvc      lstepAuditLogger
	logRepo       LineSendLogRepository
	tagConfigRepo LstepTagConfigRepository
	lineClient    lineinfra.MessagingClient
}

// newLineSendSvcFull は lineSendService を直接構築し、newLineClient シームへ
// mockLineMessagingClient を配線する（同一パッケージ内のみで非公開フィールドへアクセス可能）。
func newLineSendSvcFull(d lineSendTestDeps) LineSendService {
	if d.settings == nil {
		d.settings = &mockLstepSettingsService{}
	}
	if d.ownerRepo == nil {
		d.ownerRepo = &mockLstepOwnerRepo{}
	}
	if d.sharedFile == nil {
		d.sharedFile = &mockSharedFileSvc{}
	}
	if d.tagCacheRepo == nil {
		d.tagCacheRepo = &mockLstepTagCacheRepository{}
	}
	if d.auditSvc == nil {
		d.auditSvc = &mockAuditService{}
	}
	if d.logRepo == nil {
		d.logRepo = &mockLineSendLogRepo{}
	}
	client := d.lineClient
	if client == nil {
		client = &mockLineMessagingClient{}
	}
	return &lineSendService{
		lstepSettings: d.settings,
		ownerRepo:     d.ownerRepo,
		sharedFile:    d.sharedFile,
		tagCacheRepo:  d.tagCacheRepo,
		auditSvc:      d.auditSvc,
		logRepo:       d.logRepo,
		tagConfigRepo: d.tagConfigRepo,
		newLineClient: func(_ string) lineinfra.MessagingClient { return client },
	}
}

// ownerWithLineUserID は LINE 送信可能な（LineUserID 設定済み・opt-out していない）owner を返す。
func ownerWithLineUserID() *model.Owner {
	return &model.Owner{ID: 1, LineUserID: ptrString("U1234567890")}
}

// ---- tests ----

func TestGetSendLogs(t *testing.T) {
	tests := []struct {
		name    string
		logRepo *mockLineSendLogRepo
		wantLen int
		wantErr bool
	}{
		{
			name: "success returns logs",
			logRepo: &mockLineSendLogRepo{
				findByOwnerFn: func(_ context.Context, _, _ uint64, _ int) ([]*model.LineSendLog, error) {
					return []*model.LineSendLog{{ID: 1, Status: "sent"}, {ID: 2, Status: "failed"}}, nil
				},
			},
			wantLen: 2,
		},
		{
			name: "repo error",
			logRepo: &mockLineSendLogRepo{
				findByOwnerFn: func(_ context.Context, _, _ uint64, _ int) ([]*model.LineSendLog, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newLineSendSvc(&mockLstepOwnerRepo{}, tt.logRepo)
			logs, err := svc.GetSendLogs(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, logs, tt.wantLen)
			}
		})
	}
}

func TestSend_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newLineSendSvc(ownerRepo, &mockLineSendLogRepo{})
	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 99, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}

func TestSend_NoLineUserID(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := newLineSendSvc(ownerRepo, &mockLineSendLogRepo{})
	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}

func TestSend_LstepOptOut(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LstepOptOut: true}, nil
		},
	}
	svc := newLineSendSvc(ownerRepo, &mockLineSendLogRepo{})
	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}

func TestSend_CredentialsError(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	settings := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "", "", "", errors.New("credentials not configured")
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{settings: settings, ownerRepo: ownerRepo})

	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}

func TestSend_InvalidMessageType(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo})

	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "carrier_pigeon", Text: "hi"})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestSend_PdfUrlMissingFileID(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo})

	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "pdf_url"})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestSend_PdfUrlSignedURLError(t *testing.T) {
	fileID := uint64(5)
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	sharedFile := &mockSharedFileSvc{
		getSignedURLFn: func(_ context.Context, _, _ uint64) (string, error) {
			return "", errors.New("file not found")
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, sharedFile: sharedFile})

	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "pdf_url", FileID: &fileID})
	assert.Error(t, err)
}

func TestSend_TextSuccess(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	var pushedText string
	client := &mockLineMessagingClient{
		pushTextFn: func(_ context.Context, _, text string) error {
			pushedText = text
			return nil
		},
	}
	var loggedStatus string
	logRepo := &mockLineSendLogRepo{
		createFn: func(_ context.Context, log *model.LineSendLog) error {
			loggedStatus = log.Status
			return nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, lineClient: client, logRepo: logRepo})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "こんにちは"})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "こんにちは", pushedText)
	assert.Equal(t, "sent", loggedStatus)
	assert.Empty(t, result.TagAdded, "tagConfigRepo が未設定の場合はタグを付与しない")
}

func TestSend_TextPushError(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	client := &mockLineMessagingClient{pushTextErr: errors.New("line api error")}
	var loggedStatus string
	var loggedErrMsg *string
	logRepo := &mockLineSendLogRepo{
		createFn: func(_ context.Context, log *model.LineSendLog) error {
			loggedStatus = log.Status
			loggedErrMsg = log.ErrorMessage
			return nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, lineClient: client, logRepo: logRepo})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrBadGateway)
	assert.NotContains(t, err.Error(), "line api error", "LSA-09: raw upstream error must not appear in client response")
	assert.Contains(t, err.Error(), "LINE送信に失敗しました")
	assert.Nil(t, result)
	assert.Equal(t, "failed", loggedStatus)
	require.NotNil(t, loggedErrMsg)
	assert.Equal(t, "line_api_error", *loggedErrMsg, "LSA-09: persist classification code only")
}

func TestSend_ImageURLSuccess(t *testing.T) {
	fileID := uint64(5)
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	sharedFile := &mockSharedFileSvc{
		getSignedURLFn: func(_ context.Context, _, _ uint64) (string, error) {
			return "https://example.com/image.png", nil
		},
	}
	var capturedURL string
	client := &mockLineMessagingClient{
		pushImageURLFn: func(_ context.Context, _, originalURL, _ string) error {
			capturedURL = originalURL
			return nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, sharedFile: sharedFile, lineClient: client})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "image_url", FileID: &fileID})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://example.com/image.png", capturedURL)
}

func TestSend_PdfURLSuccess_UsesFileNameAsAltText(t *testing.T) {
	fileID := uint64(5)
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	sharedFile := &mockSharedFileSvc{
		getSignedURLFn: func(_ context.Context, _, _ uint64) (string, error) {
			return "https://example.com/doc.pdf", nil
		},
	}
	var capturedAltText string
	client := &mockLineMessagingClient{
		pushFileURLFn: func(_ context.Context, _, _, altText string) error {
			capturedAltText = altText
			return nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, sharedFile: sharedFile, lineClient: client})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "pdf_url", FileID: &fileID, FileName: "見積書.pdf"})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "見積書.pdf", capturedAltText)
}

func TestSend_PdfURLSuccess_DefaultsAltTextWhenFileNameEmpty(t *testing.T) {
	fileID := uint64(5)
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	sharedFile := &mockSharedFileSvc{
		getSignedURLFn: func(_ context.Context, _, _ uint64) (string, error) {
			return "https://example.com/doc.pdf", nil
		},
	}
	var capturedAltText string
	client := &mockLineMessagingClient{
		pushFileURLFn: func(_ context.Context, _, _, altText string) error {
			capturedAltText = altText
			return nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, sharedFile: sharedFile, lineClient: client})

	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "pdf_url", FileID: &fileID})

	assert.NoError(t, err)
	assert.Equal(t, "ファイル", capturedAltText)
}

func TestSend_WithPurposeTag(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
			return []*model.LstepSendPurposeTagPrefix{{Purpose: "estimate", TagPrefix: "見積送付_"}}, nil
		},
	}
	var upsertedTag string
	tagCacheRepo := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error {
			upsertedTag = tagName
			return nil
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, tagConfigRepo: tagConfigRepo, tagCacheRepo: tagCacheRepo})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi", Purpose: "estimate"})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.TagAdded)
	assert.Equal(t, result.TagAdded, upsertedTag)
}

func TestSend_TagCacheUpsertErrorDoesNotFailSend(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
			return []*model.LstepSendPurposeTagPrefix{{Purpose: "estimate", TagPrefix: "見積送付_"}}, nil
		},
	}
	tagCacheRepo := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error {
			return errors.New("db error")
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, tagConfigRepo: tagConfigRepo, tagCacheRepo: tagCacheRepo})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi", Purpose: "estimate"})

	assert.NoError(t, err, "タグ付与の失敗は送信全体を失敗させない")
	assert.Empty(t, result.TagAdded)
}

func TestSend_PurposePrefixFetchErrorDoesNotFailSend(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, tagConfigRepo: tagConfigRepo})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi", Purpose: "estimate"})

	assert.NoError(t, err, "タグプレフィックス取得の失敗は送信全体を失敗させない")
	assert.Empty(t, result.TagAdded)
}

func TestSend_LogRepoCreateErrorDoesNotFailSend(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	logRepo := &mockLineSendLogRepo{
		createFn: func(_ context.Context, _ *model.LineSendLog) error {
			return errors.New("db error")
		},
	}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, logRepo: logRepo})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})

	assert.NoError(t, err, "送信ログの保存失敗は送信全体を失敗させない")
	assert.NotNil(t, result)
}

func TestSend_AuditLogErrorDoesNotFailSend(t *testing.T) {
	ownerRepo := &mockLstepOwnerRepo{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return ownerWithLineUserID(), nil
		},
	}
	audit := &mockAuditService{logLstepOperationErr: errors.New("audit db down")}
	svc := newLineSendSvcFull(lineSendTestDeps{ownerRepo: ownerRepo, auditSvc: audit})

	result, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})

	assert.NoError(t, err, "監査ログの失敗は送信全体を失敗させない")
	assert.NotNil(t, result)
}

func TestPurposeTagNameFromPrefixes(t *testing.T) {
	sentAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	prefixes := []*model.LstepSendPurposeTagPrefix{
		{Purpose: "estimate", TagPrefix: "見積送付_"},
		{Purpose: "reminder", TagPrefix: "リマインド_"},
	}

	tests := []struct {
		name     string
		purpose  string
		prefixes []*model.LstepSendPurposeTagPrefix
		want     string
	}{
		{
			name:     "returns prefixed tag when purpose matches",
			purpose:  "estimate",
			prefixes: prefixes,
			want:     "見積送付_2026-07-02",
		},
		{
			name:     "returns second matching prefix",
			purpose:  "reminder",
			prefixes: prefixes,
			want:     "リマインド_2026-07-02",
		},
		{
			name:     "returns empty string when purpose does not match any prefix",
			purpose:  "unknown",
			prefixes: prefixes,
			want:     "",
		},
		{
			name:     "returns empty string when purpose is empty",
			purpose:  "",
			prefixes: prefixes,
			want:     "",
		},
		{
			name:     "returns empty string when prefixes is nil",
			purpose:  "estimate",
			prefixes: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := purposeTagNameFromPrefixes(tt.purpose, sentAt, tt.prefixes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLineSendTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "returns input unchanged when shorter than max",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "returns input unchanged when exactly max length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "truncates when longer than max",
			input:  "hello world",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "returns empty string for empty input",
			input:  "",
			maxLen: 5,
			want:   "",
		},
		{
			name:   "truncates multi-byte runes safely",
			input:  "こんにちは世界",
			maxLen: 5,
			want:   "こんにちは",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lineSendTruncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}
