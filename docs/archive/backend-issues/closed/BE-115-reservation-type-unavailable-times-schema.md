# BE-115: 予約不可時間 & 職種紐付けテーブル追加 + Go モデル生成

**Status**: Open
**Priority**: High
**Affects**: LINE予約 予約区分マスタ
**Date Created**: 2026-04-16
**Related**: TASK-001, BE-116, BE-117, FE-252

## Summary

LINE予約の予約区分毎に「予約不可時間帯」と「対応職種」を設定できるよう、
2つの新テーブルをDBスキーマに追加し、対応する Go モデルを実装する。
`make codegen` で `models.ts` を更新して FE-252 に引き渡す。

## 現状のコード

```go
// backend/internal/model/reservation_type.go（既存モデル抜粋）
type ReservationType struct {
    ID                    uint64  `gorm:"primaryKey"`
    ClinicID              uint64
    Name                  string
    IsActive              bool
    Description           string
    Color                 string
    SortOrder             int
    GroupID               *uint64
    DurationMinutes       int
    ReservationDayOption  string  // "none" | "weekday" | "saturday" | "anyday"
    IsInternal            bool
    // ... 他フィールド
}
// ※ 予約不可時間・職種紐付けフィールドは存在しない
```

```sql
-- backend/migrations/001_init.sql より（reservation_types テーブル末尾付近）
-- 予約不可時間カラム: なし
-- 職種紐付けカラム: なし
```

## 必要な変更

### 1. DB マイグレーション

`backend/migrations/001_init.sql` の末尾（既存テーブル定義の後）に追記:

```sql
-- =============================================
-- 予約区分予約不可時間
-- =============================================
CREATE TABLE reservation_type_unavailable_times (
    id                  BIGSERIAL PRIMARY KEY,
    clinic_id           BIGINT NOT NULL REFERENCES clinics(id),
    reservation_type_id BIGINT NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    unavailable_type    TEXT NOT NULL CHECK (unavailable_type IN ('weekly', 'specific')),
    -- weekly: 0=日曜, 1=月曜, ..., 6=土曜
    day_of_week         SMALLINT CHECK (
                            (unavailable_type = 'weekly' AND day_of_week BETWEEN 0 AND 6)
                            OR (unavailable_type = 'specific' AND day_of_week IS NULL)
                        ),
    specific_date       DATE CHECK (
                            (unavailable_type = 'specific' AND specific_date IS NOT NULL)
                            OR (unavailable_type = 'weekly' AND specific_date IS NULL)
                        ),
    -- "HH:MM" 形式で保存（VARCHAR(5)）
    -- TIME型を使わない理由: GORMがTIME列をstringにscanすると"HH:MM:SS"形式になり、
    -- timeslot_engine の minutesSinceMidnight（4文字HHMM専用）が必ずエラーになるため
    start_time          VARCHAR(5) NOT NULL CHECK (start_time ~ '^\d{2}:\d{2}$'),
    end_time            VARCHAR(5) NOT NULL CHECK (end_time ~ '^\d{2}:\d{2}$'),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_time_range CHECK (end_time > start_time)
    -- 論理削除なし（物理削除）: reservation_types の ON DELETE CASCADE で連動削除される
);

-- GetAvailableTimes で毎回全件取得するためのインデックス
CREATE INDEX idx_rtype_unavailable_clinic_type
    ON reservation_type_unavailable_times(clinic_id, reservation_type_id);
-- 部分インデックス: weekly / specific の絞り込み高速化
CREATE INDEX idx_rtype_unavailable_weekly
    ON reservation_type_unavailable_times(reservation_type_id, day_of_week)
    WHERE unavailable_type = 'weekly';
CREATE INDEX idx_rtype_unavailable_specific
    ON reservation_type_unavailable_times(reservation_type_id, specific_date)
    WHERE unavailable_type = 'specific';

-- =============================================
-- 予約区分 × 職種 中間テーブル（M:N）
-- =============================================
CREATE TABLE reservation_type_occupations (
    id                  BIGSERIAL PRIMARY KEY,
    clinic_id           BIGINT NOT NULL REFERENCES clinics(id),
    reservation_type_id BIGINT NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    occupation_id       BIGINT NOT NULL REFERENCES occupations(id) ON DELETE CASCADE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reservation_type_id, occupation_id)
    -- 論理削除なし（物理削除）:
    --   reservation_types 削除 → ON DELETE CASCADE で自動削除
    --   occupations 削除 → ON DELETE CASCADE で自動削除
    -- occupations テーブル自体に deleted_at なし（プロジェクトの物理削除ポリシーと一致）
);

CREATE INDEX idx_rtype_occupation_clinic
    ON reservation_type_occupations(clinic_id, reservation_type_id);
-- CountWorkingStaff のJOIN最適化
CREATE INDEX idx_rtype_occupation_occupation
    ON reservation_type_occupations(occupation_id);
```

### 2. Go モデル追加

`backend/internal/model/reservation_type.go` に追記:

```go
// UnavailableType は予約不可時間の種別
type UnavailableType string

const (
    UnavailableTypeWeekly   UnavailableType = "weekly"
    UnavailableTypeSpecific UnavailableType = "specific"
)

// ReservationTypeUnavailableTime は予約区分の予約不可時間帯
type ReservationTypeUnavailableTime struct {
    ID                 uint64          `gorm:"primaryKey"                json:"id"`
    ClinicID           uint64          `gorm:"not null"                  json:"clinic_id"`
    ReservationTypeID  uint64          `gorm:"not null"                  json:"reservation_type_id"`
    UnavailableType    UnavailableType `gorm:"not null"                  json:"unavailable_type"`
    DayOfWeek          *int8           `                                 json:"day_of_week,omitempty"`   // 0=Sun..6=Sat（weekly のみ）
    SpecificDate       *time.Time      `gorm:"type:date"                 json:"specific_date,omitempty"` // specific のみ
    StartTime          string          `gorm:"type:varchar(5);not null"  json:"start_time"`              // "HH:MM"（VARCHAR(5)で保存）
    EndTime            string          `gorm:"type:varchar(5);not null"  json:"end_time"`                // "HH:MM"（VARCHAR(5)で保存）
    CreatedAt          time.Time       `                                 json:"created_at"`
    UpdatedAt          time.Time       `                                 json:"updated_at"`
}

func (ReservationTypeUnavailableTime) TableName() string {
    return "reservation_type_unavailable_times"
}

// ReservationTypeOccupation は予約区分と職種の紐付け（M:N）
type ReservationTypeOccupation struct {
    ID                 uint64     `gorm:"primaryKey" json:"id"`
    ClinicID           uint64     `gorm:"not null"   json:"clinic_id"`
    ReservationTypeID  uint64     `gorm:"not null"   json:"reservation_type_id"`
    OccupationID       uint64     `gorm:"not null"   json:"occupation_id"`
    Occupation         *Occupation `gorm:"foreignKey:OccupationID" json:"occupation,omitempty"`
    CreatedAt          time.Time  `                  json:"created_at"`
}

func (ReservationTypeOccupation) TableName() string {
    return "reservation_type_occupations"
}
```

既存の `ReservationType` struct にリレーション追加:

```go
type ReservationType struct {
    // ... 既存フィールド ...
    UnavailableTimes []ReservationTypeUnavailableTime `gorm:"foreignKey:ReservationTypeID" json:"unavailable_times,omitempty"`
    Occupations      []ReservationTypeOccupation      `gorm:"foreignKey:ReservationTypeID" json:"occupations,omitempty"`
}
```

### 3. `make codegen` 実行

```bash
make codegen
# → frontend/src/types/generated/models.ts が自動更新される
```

## フロントエンド影響

- `make codegen` で `models.ts` に `ReservationTypeUnavailableTime` / `ReservationTypeOccupation` 型が追加される
- FE-252 はこの型を利用してフォームを構築する

## 完了条件

- [ ] `reservation_type_unavailable_times` テーブルが `001_init.sql` に追加されている
- [ ] `reservation_type_occupations` テーブルが `001_init.sql` に追加されている
- [ ] 両テーブルに適切なインデックスが設定されている
- [ ] Go モデル `ReservationTypeUnavailableTime` / `ReservationTypeOccupation` が実装されている
- [ ] `ReservationType` に `UnavailableTimes` / `Occupations` リレーションが追加されている
- [ ] `make codegen` が通り `models.ts` が更新されている
- [ ] `docker compose exec backend go build ./...` が通る
