package clinic

// 共通医院時間の版照合（CAS）・no-op・休診 set/unset の判定（A03 §2.2/§2.3）。
// 呼出側（C06 Participant の Apply）は現在行を lock した状態で本関数を使い、changed=true のときだけ
// 業務 write・監査・成功 receipt を同 TX で行う。版の増分方式（trigger/アプリ）は①の DDL に従い、
// 本関数は次版を計算しない。Participant・Q・lookup allowlist への接続は①（489/446）受領後。
//
// owner_kind・intent 6個・schema_version は保存側の owner（PR #822–#824）が package model で公開する定数を使う
// ため、ここでは宣言しない。lookup の intent→元 write action の写像は、保存側（#822–#824）と C06 registry が
// main に入った後に結線する（この package は registry／lookup allowlist を複製・模倣しない）。

import (
	"errors"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/operation"
)

var (
	// ErrOperatingHoursAlreadyConfigured は initialize 時に行が既に存在する（別 key の初期化競合を含む）。
	ErrOperatingHoursAlreadyConfigured = errors.New("clinic operating hours already configured")
	// ErrOperatingHoursUnconfigured は update/休診 set/unset 時に行がない。初期化へ案内し他 field を生成しない。
	ErrOperatingHoursUnconfigured = errors.New("clinic operating hours not configured")
	// ErrOperatingHoursVersionConflict は expected_version が現在版と異なる。
	ErrOperatingHoursVersionConflict = errors.New("clinic operating hours version conflict")
	// ErrClosedDateReasonConflict は同日に異なる理由の休診が既にある。
	ErrClosedDateReasonConflict = errors.New("closed date already set with a different reason")
)

// OperatingHoursState は lock 下で読んだ現在の保存状態。unconfigured は行なし（版・Settings なし）。
type OperatingHoursState struct {
	configured bool
	version    uint64
	settings   OperatingHoursSettings
}

// UnconfiguredOperatingHours は行なし状態。
func UnconfiguredOperatingHours() OperatingHoursState {
	return OperatingHoursState{}
}

// ConfiguredOperatingHours は保存行の状態。version は正の bigint、settings は検証済み（ゼロ値は拒否）。
func ConfiguredOperatingHours(version uint64, settings OperatingHoursSettings) (OperatingHoursState, error) {
	if version == 0 || version > maxVersion {
		return OperatingHoursState{}, errors.New("clinic operating hours version must be a positive bigint")
	}
	if !settings.isValid() {
		return OperatingHoursState{}, errors.New("clinic operating hours settings must be validated")
	}
	return OperatingHoursState{configured: true, version: version, settings: settings}, nil
}

// maxVersion は正 bigint の上限（9223372036854775807）。
const maxVersion = 1<<63 - 1

// Configured は行があるか。
func (s OperatingHoursState) Configured() bool { return s.configured }

// Version は現在版（unconfigured は 0）。
func (s OperatingHoursState) Version() uint64 { return s.version }

// Settings は現在 Settings と有無。
func (s OperatingHoursState) Settings() (OperatingHoursSettings, bool) {
	return s.settings.clone(), s.configured
}

// OperatingHoursWrite は判定結果。Changed=false は CAS 通過後の no-op で、版・業務変更監査を増やさない。
type OperatingHoursWrite struct {
	Settings OperatingHoursSettings
	Changed  bool
}

// DecideOperatingHoursInitialize は initialize（行不在要求・全 field 明示）を判定する。
func DecideOperatingHoursInitialize(current OperatingHoursState, settings OperatingHoursFields) (OperatingHoursWrite, error) {
	if current.configured {
		return OperatingHoursWrite{}, ErrOperatingHoursAlreadyConfigured
	}
	next, err := NewOperatingHoursSettings(settings)
	if err != nil {
		return OperatingHoursWrite{}, err
	}
	return OperatingHoursWrite{Settings: next, Changed: true}, nil
}

// DecideOperatingHoursUpdate は update（現在版 CAS・changes 1 field 以上）を判定する。
// no-op でも CAS を検査する。
func DecideOperatingHoursUpdate(current OperatingHoursState, expectedVersion uint64, changes OperatingHoursFields) (OperatingHoursWrite, error) {
	if err := current.checkVersion(expectedVersion); err != nil {
		return OperatingHoursWrite{}, err
	}
	next, err := current.settings.ApplyChanges(changes)
	if err != nil {
		return OperatingHoursWrite{}, err
	}
	return OperatingHoursWrite{Settings: next, Changed: !next.Equal(current.settings)}, nil
}

// DecideClosedDateSet は日付限定 adapter の set を判定する。全営業時間・曜日・祝日方針は書き換えない。
func DecideClosedDateSet(current OperatingHoursState, expectedVersion uint64, date CalendarDate, reason string) (OperatingHoursWrite, error) {
	if err := current.checkVersion(expectedVersion); err != nil {
		return OperatingHoursWrite{}, err
	}
	next, changed, err := current.settings.WithClosedDate(date, reason)
	if err != nil {
		return OperatingHoursWrite{}, err
	}
	return OperatingHoursWrite{Settings: next, Changed: changed}, nil
}

// DecideClosedDateUnset は日付限定 adapter の unset を判定する。既にない日は no-op。
func DecideClosedDateUnset(current OperatingHoursState, expectedVersion uint64, date CalendarDate) (OperatingHoursWrite, error) {
	if err := current.checkVersion(expectedVersion); err != nil {
		return OperatingHoursWrite{}, err
	}
	next, changed, err := current.settings.WithoutClosedDate(date)
	if err != nil {
		return OperatingHoursWrite{}, err
	}
	return OperatingHoursWrite{Settings: next, Changed: changed}, nil
}

func (s OperatingHoursState) checkVersion(expected uint64) error {
	if !s.configured {
		return ErrOperatingHoursUnconfigured
	}
	if expected != s.version {
		return ErrOperatingHoursVersionConflict
	}
	return nil
}

// OperatingHoursRejection は判定エラーを既存 C06 の確定拒否 status へ写像する。
//
//   - 既に行がある初期化（別 key による同時初期化を含む）と版の不一致は version-conflict。
//   - 行がない update／休診 set／unset（ErrOperatingHoursUnconfigured）は「存在しない owner 対象」であり、
//     operation/contract.go の StatusScopeNotFound の注記（"a missing owner target uses validation"）に従い
//     validation。scope-not-found は FindBinding が binding を見つけない場合専用で owner の拒否ではない。
//   - 同日異理由の休診と入力不正は validation。
//
// それ以外は確定拒否ではない（ok=false）。
func OperatingHoursRejection(err error) (*operation.Rejection, bool) {
	switch {
	case errors.Is(err, ErrOperatingHoursAlreadyConfigured),
		errors.Is(err, ErrOperatingHoursVersionConflict):
		return &operation.Rejection{Status: operation.StatusVersionConflict}, true
	case errors.Is(err, ErrOperatingHoursUnconfigured),
		errors.Is(err, ErrClosedDateReasonConflict),
		apperrors.IsInvalidInput(err):
		return &operation.Rejection{Status: operation.StatusValidation}, true
	default:
		return nil, false
	}
}

// OperatingHoursResultRefs は成功 receipt の result_refs（exact `{clinic_id}`）。
func OperatingHoursResultRefs(clinicID uint64) map[string]string {
	return map[string]string{"clinic_id": strconv.FormatUint(clinicID, 10)}
}

// OperatingHoursCommittedVersions は成功 receipt の committed_versions（exact `{clinic_operating_hours}`）。
// 値は実保存版の正規十進文字列で、no-op は CAS を通過した現在版を渡す。
func OperatingHoursCommittedVersions(savedVersion uint64) map[string]string {
	return map[string]string{"clinic_operating_hours": strconv.FormatUint(savedVersion, 10)}
}
