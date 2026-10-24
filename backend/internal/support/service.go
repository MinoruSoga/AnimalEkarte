package support

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateBugReportInput はサービス層への作成入力（handler で検証済みの値）
type CreateBugReportInput struct {
	Title         string
	Detail        string
	PageURL       string
	RoutePath     string
	UserAgent     string
	Viewport      string
	AppVersion    string
	ScreenshotKey *string
}

// Service はバグ報告とヘルプチャット履歴のユースケースインターフェース。
// バグ報告の一覧・作成は全医院横断だが、status 更新・Plane 起票・削除は
// 報告元 clinicID スコープ（共有ボードは閲覧のみ公開）。Create は provenance
// として clinic_id/reporter_staff_id を記録する。チャット履歴のみ
// clinic×staff スコープ。
type Service interface {
	Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error)
	ListAll(ctx context.Context) ([]BugReportWithReporter, error)
	// Get は報告元 clinicID スコープで 1 件を返す。他医院は NotFound（404）。
	Get(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error)
	EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	// ListChatExchanges は全医院の質問+回答ペアを新しい順で返す（共有一覧ページ用）
	ListChatExchanges(ctx context.Context) ([]ChatExchange, error)
	// ListChatRooms は全医院の発言ストリームを疑似会話ルーム単位で返す（共有一覧ページ用）
	ListChatRooms(ctx context.Context) ([]ChatRoom, error)
	RecordChatExchange(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error
	ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error
	// SyncPlaneTicketStates は scheduled job 用 — Plane 側 completed → resolved の
	// 一方向同期。全医院横断で動き、各報告の clinic_id で UpdateStatus する。
	SyncPlaneTicketStates(ctx context.Context) PlaneSyncResult
}

// codeSupportDeliveryPending は、置換先の配送記録（support_delivery_* の intent / attempt / result）の
// writer・reader がまだ接続されていないため停止した Plane 操作を示す 422 の安定コード（EMR-549 C1）。
// 共有定数を apperrors へ足さず、owner である support パッケージに閉じる（medicalrecord と同じ方針）。
const codeSupportDeliveryPending = "support_delivery_pending"

// defaultPlanePostSaveTimeout は Plane 外部呼出 budget の既定値（PLANE_TIMEOUT_MS の既定 10000ms に合わせる）。
// 送信停止中は参照されないが、composition の WithPlaneTimeout 配線を壊さないため seam を残す。
const defaultPlanePostSaveTimeout = 10 * time.Second

type service struct {
	repo     Repository
	tickets  TicketCreator     // nil = Plane 連携無効（ローカル保存のみ）
	notifier BugReportNotifier // nil = Discord 通知無効
	// audit は配送結果の監査記録先（nil = 記録しない）。Plane 送信停止中は参照されない。
	audit AuditLogger
	// planeTimeout は Plane 外部呼出 timeout。送信停止中は参照されない。
	planeTimeout time.Duration
}

// ServiceOption は NewService へのオプション依存を注入する。
type ServiceOption func(*service)

// WithAuditLogger は配送結果を記録する監査ロガーを設定する。
// 監査書込み自体は best-effort の方針を維持する。Plane 送信が停止中
// （codeSupportDeliveryPending）の間は記録対象がなく、参照されない。
func WithAuditLogger(logger AuditLogger) ServiceOption {
	return func(s *service) { s.audit = logger }
}

// WithPlaneTimeout は Plane 外部呼出の timeout を設定する。
// 0 以下は無視して既定値を維持する。Plane 送信が停止中の間は参照されない。
func WithPlaneTimeout(timeout time.Duration) ServiceOption {
	return func(s *service) {
		if timeout > 0 {
			s.planeTimeout = timeout
		}
	}
}

// NewService は Service を初期化して返す
func NewService(repo Repository, tickets TicketCreator, notifier BugReportNotifier, opts ...ServiceOption) Service {
	s := &service{
		repo:         repo,
		tickets:      tickets,
		notifier:     notifier,
		planeTimeout: defaultPlanePostSaveTimeout,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error) {
	report := model.SupportBugReport{
		ClinicID:        clinicID,
		ReporterStaffID: reporterStaffID,
		Title:           input.Title,
		Detail:          input.Detail,
		PageURL:         input.PageURL,
		RoutePath:       input.RoutePath,
		UserAgent:       input.UserAgent,
		Viewport:        input.Viewport,
		AppVersion:      input.AppVersion,
		ScreenshotKey:   input.ScreenshotKey,
		Status:          model.SupportBugReportStatusOpen,
	}
	// Discord 通知は DB 保存前に best-effort で実行する（報告原文の外部転写を優先）。
	// 失敗しても報告作成は続行する（Discord 障害で報告自体を失わせない）。
	// ID は未採番のため通知に含まれない。
	if s.notifier != nil {
		if err := s.notifier.NotifyBugReport(ctx, &report); err != nil {
			slog.WarnContext(ctx, "failed to send bug report discord notification",
				"clinic_id", clinicID, "reporter_staff_id", reporterStaffID, "error", err)
		}
	}
	if err := s.repo.Create(ctx, &report); err != nil {
		return nil, err
	}
	// 保存後の Plane 自動起票は停止中（EMR-549 C1）。送信前確認 intent・宛先ごとの試行 claim・
	// 結果記録の writer が未接続のまま外部へ送ると、孤立 issue と結果不明の再送を生む
	// （technical-spec.md#support-selected-delivery-contract 2・3・5）。報告の院内保存は成功させ、
	// 応答の plane_delivery_state は unavailable（未送信の証明ではない）のまま返す。
	if s.tickets != nil {
		slog.WarnContext(ctx, "plane auto-ticket is not sent: delivery contract is not connected",
			"code", codeSupportDeliveryPending, "report_id", report.ID, "clinic_id", clinicID)
	}
	return &report, nil
}

func (s *service) ListAll(ctx context.Context) ([]BugReportWithReporter, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) Get(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	return s.repo.FindByIDForClinic(ctx, clinicID, id)
}

func (s *service) UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error) {
	if !model.IsValidSupportBugReportStatus(string(status)) {
		return nil, apperrors.WrapInvalidInput("invalid status")
	}
	if err := s.repo.UpdateStatus(ctx, clinicID, id, status); err != nil {
		return nil, err
	}
	return s.repo.FindByIDForClinic(ctx, clinicID, id)
}

// EnsurePlaneTicket は一覧からの手動起票・再送。
// 旧実装は support_bug_reports の Plane 列（plane_issue_id / plane_issue_url / plane_sync_error）を
// claim 先にしていたが、列は D23 の配送宛先／試行／結果（support_delivery_*）へ移され削除済み。
// 送信前確認 intent・宛先ごとの試行 claim・結果記録の writer が未接続のまま外部へ起票すると、
// 孤立 issue と結果不明の再送を生む（technical-spec.md#support-selected-delivery-contract 3・5・7）。
// そのため報告元医院スコープの確認後に固定の 422 で停止する。成功を装わず、列も復元しない。
func (s *service) EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	// 他医院・不存在は従来どおり NotFound（AC-support-02: 報告元以外は操作できない）。
	if _, err := s.repo.FindByIDForClinic(ctx, clinicID, id); err != nil {
		return nil, err
	}
	return nil, apperrors.WrapUnprocessableWithCode(codeSupportDeliveryPending,
		"Plane への起票は、送信前の確認と配送記録（宛先・送信試行・結果）の接続まで利用できません")
}

// Delete は報告を論理削除する。返り値は削除前に読み取った行で、
// handler がスクショの後始末と監査記録に使う。
func (s *service) Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	report, err := s.repo.FindByIDForClinic(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SoftDeleteBugReport(ctx, clinicID, id); err != nil {
		return nil, err
	}
	return report, nil
}

// ListChatHistory は指定スタッフの会話履歴を古い順で返す。
func (s *service) ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error) {
	return s.repo.ListChatHistory(ctx, clinicID, staffID)
}

// RecordChatExchange は回答が成立した「質問+回答」のペアを1回の INSERT で保存する。
// LLM 呼び出しが失敗したやり取りは呼び出し側から渡されない前提。
func (s *service) RecordChatExchange(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error {
	var sourcesJSON *json.RawMessage // nil = sources なし（NULL を保存する）
	if len(sources) > 0 {
		b, err := json.Marshal(sources)
		if err != nil {
			return apperrors.Wrap(err, "failed to encode chat sources")
		}
		raw := json.RawMessage(b)
		sourcesJSON = &raw
	}
	return s.repo.CreateChatMessages(ctx, []*model.SupportChatMessage{
		{ClinicID: clinicID, StaffID: staffID, Role: model.SupportChatRoleUser, Content: userMessage},
		{ClinicID: clinicID, StaffID: staffID, Role: model.SupportChatRoleAssistant, Content: assistantReply, Sources: sourcesJSON},
	})
}

// ChatExchange は一覧表示用の質問+回答ペア（回答行に結合したスタッフ/医院名つき）。
type ChatExchange struct {
	UserMessage      model.SupportChatMessage
	AssistantMessage model.SupportChatMessage
	StaffName        string
	ClinicName       string
}

// ListChatExchanges は全医院の履歴を user→assistant のペアにして新しい順で返す。
// RecordChatExchange が質問+回答を同一 batch で保存するため、同一 clinic×staff 内で
// assistant の直前にある user 行がその質問。対が欠けた行（窓の境界など）は除外する。
func (s *service) ListChatExchanges(ctx context.Context) ([]ChatExchange, error) {
	messages, err := s.repo.ListAllChatMessages(ctx)
	if err != nil {
		return nil, err
	}
	// 新しい順で来るので古い順に反転してからペア化する
	type exchangeKey struct{ clinicID, staffID uint64 }
	pending := make(map[exchangeKey]*ChatMessageWithMeta, len(messages))
	exchanges := make([]ChatExchange, 0, len(messages)/2)
	for i := len(messages) - 1; i >= 0; i-- {
		m := &messages[i]
		key := exchangeKey{clinicID: m.ClinicID, staffID: m.StaffID}
		switch m.Role {
		case model.SupportChatRoleUser:
			pending[key] = m
		case model.SupportChatRoleAssistant:
			user, ok := pending[key]
			if !ok {
				continue
			}
			delete(pending, key)
			exchanges = append(exchanges, ChatExchange{
				UserMessage:      user.SupportChatMessage,
				AssistantMessage: m.SupportChatMessage,
				StaffName:        m.StaffName,
				ClinicName:       m.ClinicName,
			})
		}
	}
	// 古い順で組み立てたので新しい順に反転して返す
	for i, j := 0, len(exchanges)-1; i < j; i, j = i+1, j-1 {
		exchanges[i], exchanges[j] = exchanges[j], exchanges[i]
	}
	return exchanges, nil
}

// chatRoomGap は会話区切りとして扱う無発言時間。同一 clinic×staff の発言が
// この間隔を超えて空いたら別ルームとみなす（会話区切りはDBスキーマに存在しないためのヒューリスティック）。
const chatRoomGap = 30 * time.Minute

// ChatRoomTurn はルーム内の1発言（user / assistant）。
type ChatRoomTurn struct {
	ID        uint64
	Role      model.SupportChatMessageRole
	Content   string
	Sources   json.RawMessage
	CreatedAt time.Time
}

// ChatRoom は同一 clinic×staff の時間的に連続した会話まとまり（擬似ルーム）。
type ChatRoom struct {
	// ID は導出キー "clinicID-staffID-先頭メッセージID"。DB に永続化されない。
	ID            string
	ClinicName    string
	StaffName     string
	StartedAt     time.Time
	LastMessageAt time.Time
	Turns         []ChatRoomTurn
}

// ListChatRooms は全医院のサポートチャット履歴を、同一 clinic×staff の発言ストリームを
// chatRoomGap を区切りにした「擬似会話ルーム」単位で返す。ルームは最終発言の新しい順、
// ルーム内の発言は古い順。閲覧範囲は ListChatExchanges と同じ共有ボード方針。
func (s *service) ListChatRooms(ctx context.Context) ([]ChatRoom, error) {
	messages, err := s.repo.ListAllChatMessages(ctx)
	if err != nil {
		return nil, err
	}
	type key struct{ clinicID, staffID uint64 }
	rooms := make([]ChatRoom, 0, len(messages)/2)
	open := map[key]int{}
	// ListAllChatMessages は新しい順 — 古い順に走査して発言をルームへ積む
	for i := len(messages) - 1; i >= 0; i-- {
		m := &messages[i]
		k := key{m.ClinicID, m.StaffID}
		idx, ok := open[k]
		if !ok || m.CreatedAt.Sub(rooms[idx].LastMessageAt) > chatRoomGap {
			idx = len(rooms)
			rooms = append(rooms, ChatRoom{
				ID:         fmt.Sprintf("%d-%d-%d", m.ClinicID, m.StaffID, m.ID),
				ClinicName: m.ClinicName,
				StaffName:  m.StaffName,
				StartedAt:  m.CreatedAt,
				Turns:      []ChatRoomTurn{},
			})
			open[k] = idx
		}
		rooms[idx].Turns = append(rooms[idx].Turns, ChatRoomTurn{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			Sources:   storedSourcesBytes(m.Sources),
			CreatedAt: m.CreatedAt,
		})
		rooms[idx].LastMessageAt = m.CreatedAt
	}
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].LastMessageAt.After(rooms[j].LastMessageAt)
	})
	return rooms, nil
}

// ClearChatHistory は指定スタッフの会話履歴をすべて soft delete する。
func (s *service) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	return s.repo.ClearChatHistory(ctx, clinicID, staffID)
}

// PlaneSyncResult は scheduled plane_sync ジョブの件数サマリ。
// scheduler.Result の契約に合わせ Processed == Succeeded + Failed を維持する
// （配送結果 reader 未接続・fetch/update の失敗だけが Failed — 未完了スキップは Succeeded に含める）。
type PlaneSyncResult struct {
	Processed int
	Succeeded int
	Failed    int
}

// SyncPlaneTicketStates は Plane 側の状態を報告へ揃える scheduled job の入口。
// 旧実装は support_bug_reports.plane_issue_id（削除済み）を起票済みの印にしていた。置換先は
// support_delivery_results（provider='plane' の最新 result_state='succeeded' の external_id）だが、
// その reader と結果不明を保留する照合（technical-spec.md#support-selected-delivery-contract 5・7）が
// 未接続のため、外部 ID の無い報告を「未送信」と推測して空振りさせず、実行ごとに 1 件の失敗として
// 報告する（Processed == Succeeded + Failed を維持）。連携無効（tickets が TicketStateReader を
// 実装しない = nil 含む）は対象外のため zero result で no-op。
func (s *service) SyncPlaneTicketStates(ctx context.Context) PlaneSyncResult {
	if _, ok := s.tickets.(TicketStateReader); !ok {
		return PlaneSyncResult{}
	}
	slog.ErrorContext(ctx, "plane sync: delivery results reader is not connected; state sync skipped",
		"code", codeSupportDeliveryPending)
	return PlaneSyncResult{Processed: 1, Failed: 1}
}
