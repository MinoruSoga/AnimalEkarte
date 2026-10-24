package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type SendLineMessageInput struct {
	OwnerID     uint64
	StaffID     uint64
	MessageType string
	Text        string
	FileID      *uint64
	FileName    string
	Purpose     string
}

type SendLineMessageResult struct {
	SentAt   time.Time
	TagAdded string
}

type LineSendService interface {
	Send(ctx context.Context, clinicID uint64, input *SendLineMessageInput) (*SendLineMessageResult, error)
}

type lineSendService struct{}

// NewLineSendService は手動 LINE 送信 service を返す。
//
// EMR-504: 旧 line_send_logs は DROP 済みで、置換先の配送記録（line_dispatch_intents / attempts / results）
// の writer・目的別確認・provider 結果の照合はまだ接続されていない。記録なしの直接 Push は
// 「結果不明を未送信にしない」契約に反するため、Send は固定の 422 で fail-closed する（成功を装わない）。
// 再開条件は dispatch の C1/C2 接続（docs/work/emr/emr-504-c-dispatch-adapter-contract-20261021.md §2, §5, §8）。
func NewLineSendService() LineSendService {
	return &lineSendService{}
}

func (s *lineSendService) Send(_ context.Context, _ uint64, _ *SendLineMessageInput) (*SendLineMessageResult, error) {
	return nil, apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineDispatchPending,
		"LINE送信の配送記録がまだ接続されていないため送信できません",
	)
}
