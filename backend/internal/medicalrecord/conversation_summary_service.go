package medicalrecord

// conversation_summary_service.go — EMR-485 非表示会話要約の保存/読戻し service。
//
// 範囲: 入力検証と保存/読戻しの委譲のみ。route 登録・handler・DTO 変換は行わない
// （公開面へは結線しない。Phase2 の音声/外部サービス/文字起こし/手動入力formも作らない）。
// 保存契約（NULL/revision/source/期待版/再送）は repository の ConversationSummary /
// SaveConversationSummaryCommand のコメントを正本とする。

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SaveConversationSummaryInput は会話要約保存の入力 DTO。
//
// ExpectedRevision:
//   - nil : 版照合をスキップ（最終書き込み優先）
//   - 0   : 「未保存（3列すべて NULL）」を期待する
//   - n>0 : 現在の revision が n のときだけ保存する
//
// 再送例外: 保存済み内容が本入力の意図（summary+source）と一致する再送は
// 期待版が実際の版と異なっても冪等成功で返す（詳細は repository の
// SaveConversationSummaryCommand / Save の doc comment を正本とする）。
type SaveConversationSummaryInput struct {
	ClinicID         uint64
	MedicalRecordID  uint64
	Summary          string
	Source           string
	ExpectedRevision *int64
}

// ConversationSummaryService は来院カルテ単位の非公開会話要約の保存/読戻しを担う。
//
// 認可契約（EMR-485 監査3「現在認可/医院境界…をその保存subsetへだけ接続」）:
//   - 本service は公開経路（route/handler）を持たない。clinic/来院の取り違えと
//     他医院アクセスは repository の clinic_id 境界で拒否する（NotFound で存在を漏らさない）。
//   - staff 認可（permission）は HTTP 経線で検査する既存構成のため、公開API未選定の
//     今回時点では接続点が存在しない。Phase2 で route を結線する際は、当該 route に
//     選択 clinic スコープの medicalrecord 編集系 permission middleware を付与し、
//     本service へは認可済み clinicID のみを渡すこと。service 単体で権限を握らない。
type ConversationSummaryService interface {
	// Get は3列を返す。未保存は3列すべて nil（unknown）で返し、推定値を生成しない。
	Get(ctx context.Context, clinicID, medicalRecordID uint64) (*ConversationSummary, error)
	// Save は要約本文と source を保存し、採番後の revision を含む保存済み状態を返す。
	Save(ctx context.Context, input SaveConversationSummaryInput) (*ConversationSummary, error)
}

type conversationSummaryService struct {
	repo ConversationSummaryRepository
}

// NewConversationSummaryService は ConversationSummaryService を初期化して返す。
func NewConversationSummaryService(repo ConversationSummaryRepository) ConversationSummaryService {
	return &conversationSummaryService{repo: repo}
}

// Get は clinic_id + medical_record_id に限定した読戻しを行う。
func (s *conversationSummaryService) Get(ctx context.Context, clinicID, medicalRecordID uint64) (*ConversationSummary, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.WrapInternalServerError("会話要約の保存先が未設定です")
	}
	if clinicID == 0 || medicalRecordID == 0 {
		return nil, apperrors.WrapInvalidInput("医院とカルテを特定できません")
	}
	return s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
}

// Save は入力を正規化・検証してから repository に委譲する。
// source は識別子なので前後空白を除去して保存する。要約本文は内容保持のため改変しない。
func (s *conversationSummaryService) Save(ctx context.Context, input SaveConversationSummaryInput) (*ConversationSummary, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.WrapInternalServerError("会話要約の保存先が未設定です")
	}
	cmd := SaveConversationSummaryCommand{
		ClinicID:         input.ClinicID,
		MedicalRecordID:  input.MedicalRecordID,
		Summary:          input.Summary,
		Source:           strings.TrimSpace(input.Source),
		ExpectedRevision: input.ExpectedRevision,
	}
	if err := cmd.validate(); err != nil {
		return nil, err
	}

	saved, err := s.repo.Save(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if saved != nil && saved.Revision != nil {
		// 要約本文と source は非公開データなのでログに出さない（識別子と版のみ）。
		slog.InfoContext(ctx, "conversation_summary saved",
			slog.Uint64("clinic_id", cmd.ClinicID),
			slog.Uint64("medical_record_id", cmd.MedicalRecordID),
			slog.Int64("conversation_summary_revision", *saved.Revision))
	}
	return saved, nil
}
