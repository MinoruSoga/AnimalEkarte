package clinic

// 共通医院時間の読取 use case（A03 §2.2/§2.3 の GET 側）。
// 行なしは未設定（補完しない・エラーにもしない）。DB 障害と保存データ破損は未設定や空成功へ変換せず、
// 別のエラーとして返し、handler が 503 にする。旧 clinic_holidays は参照しない（§2.4）。

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ErrOperatingHoursUnavailable は保存行の読取に失敗した（DB 障害・context 取消など）ことを表す。
var ErrOperatingHoursUnavailable = errors.New("clinic operating hours could not be read")

// ErrOperatingHoursClinicIDOutOfRange は医院 ID が 1..MaxInt64 の外（0 を含む）であることを表す内部の誤用エラー。
// 未設定（unconfigured）とは区別し、handler は 503 にする。
var ErrOperatingHoursClinicIDOutOfRange = errors.New("clinic id is outside 1..9223372036854775807")

// OperatingHoursRepository は共通医院時間の data port。行なしは found=false（エラーではない）。
type OperatingHoursRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (row model.ClinicOperatingHours, found bool, err error)
}

// CommonOperatingHours は医院の共通医院時間の現在状態。Configured=false なら未設定で Version=0・Settings は invalid。
type CommonOperatingHours struct {
	Configured bool
	Version    int64
	Settings   OperatingHoursSettings
}

// OperatingHoursReader は handler が消費する読取 port。
type OperatingHoursReader interface {
	Get(ctx context.Context, clinicID uint64) (CommonOperatingHours, error)
}

// OperatingHoursReadService は OperatingHoursReader の実装。
type OperatingHoursReadService struct {
	repo OperatingHoursRepository
}

// NewOperatingHoursReadService constructs the read use case.
func NewOperatingHoursReadService(repo OperatingHoursRepository) *OperatingHoursReadService {
	return &OperatingHoursReadService{repo: repo}
}

// Get は clinicID の共通医院時間を返す。
// エラーは ErrOperatingHoursUnavailable（読取失敗）か ErrOperatingHoursStoredDataCorrupt（保存データ破損）。
func (s *OperatingHoursReadService) Get(ctx context.Context, clinicID uint64) (CommonOperatingHours, error) {
	if clinicID == 0 || clinicID > math.MaxInt64 {
		// 範囲外の医院 ID を未設定へ畳まない（wire は 1..MaxInt64 の正規十進のみ）。呼出側の誤用で、入力不正ではない。
		return CommonOperatingHours{}, fmt.Errorf("%w: %d", ErrOperatingHoursClinicIDOutOfRange, clinicID)
	}
	if s == nil || s.repo == nil {
		return CommonOperatingHours{}, fmt.Errorf("%w: repository is not configured", ErrOperatingHoursUnavailable)
	}
	row, found, err := s.repo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return CommonOperatingHours{}, fmt.Errorf("%w: %w", ErrOperatingHoursUnavailable, err)
	}
	if !found {
		return CommonOperatingHours{}, nil
	}
	if row.ClinicID != clinicID {
		return CommonOperatingHours{}, corruptOperatingHours("row clinic_id does not match the requested clinic")
	}
	settings, err := LoadOperatingHoursSettings(row)
	if err != nil {
		return CommonOperatingHours{}, err
	}
	return CommonOperatingHours{Configured: true, Version: row.Version, Settings: settings}, nil
}

// YearMonth は厳密 YYYY-MM（月 01..12）。ゼロ値は無効。
type YearMonth struct {
	text string
}

// ParseYearMonth は ^[0-9]{4}-(0[1-9]|1[0-2])$ だけを受ける。
func ParseYearMonth(raw string) (YearMonth, error) {
	invalid := apperrors.WrapInvalidInput("year_month は YYYY-MM 形式（月は 01〜12）で指定してください")
	if len(raw) != len("2006-01") || raw[4] != '-' {
		return YearMonth{}, invalid
	}
	for i := range len(raw) {
		if i == 4 {
			continue
		}
		if raw[i] < '0' || raw[i] > '9' {
			return YearMonth{}, invalid
		}
	}
	month := raw[5:]
	if month < "01" || month > "12" {
		return YearMonth{}, invalid
	}
	return YearMonth{text: raw}, nil
}

// String は YYYY-MM。ゼロ値は空文字。
func (m YearMonth) String() string { return m.text }

// ClosedDatesIn は month の JST 暦月に含まれる個別休診日を日付昇順で返す（複製・非 nil）。
// 未設定は空を返す（呼出側が state で未設定と区別する）。configured なのに Settings が invalid な値は
// 内部の誤用なので、空ではなくエラーにする。
func (c CommonOperatingHours) ClosedDatesIn(month YearMonth) ([]ClosedDate, error) {
	if month.text == "" {
		return nil, errors.New("year month was not built by ParseYearMonth")
	}
	if !c.Configured {
		return []ClosedDate{}, nil
	}
	all, err := c.Settings.ClosedDates()
	if err != nil {
		return nil, err
	}
	prefix := month.text + "-"
	out := make([]ClosedDate, 0, len(all))
	for _, cd := range all {
		if strings.HasPrefix(cd.Date.String(), prefix) {
			out = append(out, cd)
		}
	}
	return out, nil
}
