# BE-118: 統一予約基盤 — DBスキーマ変更 + Go モデル + make codegen

**Status**: Open
**Priority**: High
**Affects**: appointments, trimming_records, reservation_types
**Date Created**: 2026-04-16
**Related**: TASK-002, BE-119, BE-120, FE-253, FE-254

## Summary

`trimming_records` / `trimming_record_options` テーブルを廃止し、
`appointment_trimming_details` / `appointment_trimming_options` に置き換える。
`reservation_types` に `category` カラムを追加してトリミング区分を識別できるようにする。
Go モデルを更新し、`make codegen` で `models.ts` を再生成する。

## 現状のコード

```sql
-- backend/migrations/001_init.sql:70
CREATE TYPE trimming_status AS ENUM ('completed', 'reserved', 'in_progress');

-- backend/migrations/001_init.sql:369-389
CREATE TABLE reservation_types (
    id                       BIGSERIAL   PRIMARY KEY,
    clinic_id                bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name                     text        NOT NULL,
    is_active                boolean     NOT NULL DEFAULT true,
    description              text        NOT NULL DEFAULT '',
    color                    text        NOT NULL DEFAULT '#3B82F6',
    sort_order               integer              DEFAULT 0,
    group_id                 bigint               REFERENCES reservation_type_groups(id) ON DELETE SET NULL,
    reservation_display_name text        NOT NULL DEFAULT '',
    duration_minutes         int         NOT NULL DEFAULT 15,
    short_name               text        NOT NULL DEFAULT '',
    show_short_name          boolean     NOT NULL DEFAULT false,
    reservation_visible      boolean     NOT NULL DEFAULT true,
    reservation_comment      text        NOT NULL DEFAULT '',
    reservation_image_url    text        NOT NULL DEFAULT '',
    reservation_day_option   text        NOT NULL DEFAULT 'none',
    is_internal              boolean     NOT NULL DEFAULT false,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now()
);
-- ※ category カラムなし

-- backend/migrations/001_init.sql:718-740
CREATE TABLE trimming_records (
    id              BIGSERIAL        PRIMARY KEY,
    clinic_id       bigint           NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    date            date             NOT NULL,                        -- ← 日付のみ（時刻なし）
    pet_id          bigint                    REFERENCES pets(id) ON DELETE RESTRICT,
    style_request   text             NOT NULL DEFAULT '',
    staff_id        bigint                    REFERENCES staffs(id) ON DELETE SET NULL,
    status          trimming_status           DEFAULT 'reserved',
    course_id       bigint                    REFERENCES trimming_courses(id) ON DELETE SET NULL,
    body_weight     numeric(6,2),
    bw_unit         body_weight_unit          DEFAULT 'Kg',
    body_temperature numeric(4,1),
    used_shampoo    text             NOT NULL DEFAULT '',
    used_ribbon     text             NOT NULL DEFAULT '',
    remarks         text             NOT NULL DEFAULT '',
    style_image     text             NOT NULL DEFAULT '',
    completed_image text             NOT NULL DEFAULT '',
    created_at      timestamptz      NOT NULL DEFAULT now(),
    updated_at      timestamptz      NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

-- backend/migrations/001_init.sql:1145-1154
CREATE TABLE trimming_record_options (
    id                 BIGSERIAL PRIMARY KEY,
    trimming_record_id bigint    NOT NULL REFERENCES trimming_records(id) ON DELETE CASCADE,
    option_id          bigint    NOT NULL REFERENCES trimming_options(id) ON DELETE RESTRICT,
    sort_order         integer            DEFAULT 0,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);
```

```go
// backend/internal/model/reservation_type.go:18-45（現在のReservationTypeには category なし）
type ReservationType struct {
    // ... 既存フィールド
    // Category フィールドなし
}

// backend/internal/model/trimming.go:17-56（TrimmingRecord / TrimmingRecordOption が存在）
type TrimmingRecord struct { ... }
type TrimmingRecordOption struct { ... }
type TrimmingStatus string
```

## 必要な変更

### 1. DB マイグレーション（001_init.sql）

#### 1-a. 削除対象（物理削除 — DB リセット運用）

`001_init.sql` から以下を**完全に削除**する:

```sql
-- 削除: line 70
CREATE TYPE trimming_status AS ENUM ('completed', 'reserved', 'in_progress');

-- 削除: lines 718-740（trimming_records テーブル定義ごと）
CREATE TABLE trimming_records ( ... );

-- 削除: lines 1145-1154（trimming_record_options テーブル定義ごと）
CREATE TABLE trimming_record_options ( ... );

-- 削除: 関連インデックス
--   idx_trimming_records_clinic_date（line 1508-1511）
--   idx_trimming_record_options_unique（line 1326）
--   idx_trimming_records_staff_id（line 1423）
```

#### 1-b. 追加: reservation_type_category ENUM

`001_init.sql` の既存 ENUM 定義群（CREATE TYPE セクション）に追記:

```sql
CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming');
```

#### 1-c. 追加: reservation_types.category カラム

`reservation_types` テーブル定義（line 369-389）に `category` を追加:

```sql
CREATE TABLE reservation_types (
    id                       BIGSERIAL   PRIMARY KEY,
    clinic_id                bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name                     text        NOT NULL,
    is_active                boolean     NOT NULL DEFAULT true,
    description              text        NOT NULL DEFAULT '',
    color                    text        NOT NULL DEFAULT '#3B82F6',
    sort_order               integer              DEFAULT 0,
    group_id                 bigint               REFERENCES reservation_type_groups(id) ON DELETE SET NULL,
    reservation_display_name text        NOT NULL DEFAULT '',
    duration_minutes         int         NOT NULL DEFAULT 15,
    short_name               text        NOT NULL DEFAULT '',
    show_short_name          boolean     NOT NULL DEFAULT false,
    reservation_visible      boolean     NOT NULL DEFAULT true,
    reservation_comment      text        NOT NULL DEFAULT '',
    reservation_image_url    text        NOT NULL DEFAULT '',
    reservation_day_option   text        NOT NULL DEFAULT 'none',
    is_internal              boolean     NOT NULL DEFAULT false,
    category  reservation_type_category  NOT NULL DEFAULT 'general',   -- ★ 追加
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now()
);
```

#### 1-d. 追加: appointment_trimming_details テーブル

`001_init.sql` の `appointments` テーブル定義（line 671-692）の**直後**に追記:

```sql
-- ------------------------------------
-- XX. appointment_trimming_details（トリミング予約詳細 - appointments の1:1拡張）
-- ------------------------------------
CREATE TABLE appointment_trimming_details (
    id               BIGSERIAL        PRIMARY KEY,
    clinic_id        BIGINT           NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    appointment_id   BIGINT           NOT NULL UNIQUE REFERENCES appointments(id) ON DELETE CASCADE,
    course_id        BIGINT                    REFERENCES trimming_courses(id) ON DELETE SET NULL,
    style_request    TEXT             NOT NULL DEFAULT '',
    body_weight      NUMERIC(6,2),
    bw_unit          body_weight_unit          DEFAULT 'Kg',
    body_temperature NUMERIC(4,1),
    used_shampoo     TEXT             NOT NULL DEFAULT '',
    used_ribbon      TEXT             NOT NULL DEFAULT '',
    remarks          TEXT             NOT NULL DEFAULT '',
    style_image      TEXT             NOT NULL DEFAULT '',
    completed_image  TEXT             NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- トリミング詳細の主要アクセスパターン
CREATE INDEX idx_appt_trimming_clinic_appointment
    ON appointment_trimming_details(clinic_id, appointment_id);

-- ------------------------------------
-- XX+1. appointment_trimming_options（トリミング予約 × オプション M:N）
-- ------------------------------------
CREATE TABLE appointment_trimming_options (
    id             BIGSERIAL   PRIMARY KEY,
    appointment_id BIGINT      NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    option_id      BIGINT      NOT NULL REFERENCES trimming_options(id) ON DELETE RESTRICT,
    sort_order     INTEGER              DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (appointment_id, option_id)
);

CREATE INDEX idx_appt_trimming_options_appointment
    ON appointment_trimming_options(appointment_id);
```

### 2. Go モデル変更

#### 2-a. `backend/internal/model/trimming.go` — TrimmingRecord/TrimmingRecordOption/TrimmingStatus 削除

以下を**削除**する（TrimmingCourse, TrimmingOption, TargetSize等マスタ定義は維持）:

```go
// ↓ 削除
type TrimmingStatus string
const (
    TrimmingStatusCompleted  TrimmingStatus = "completed"
    TrimmingStatusReserved   TrimmingStatus = "reserved"
    TrimmingStatusInProgress TrimmingStatus = "in_progress"
)

// ↓ 削除
type TrimmingRecord struct { ... }
func (TrimmingRecord) TableName() string { ... }

// ↓ 削除
type TrimmingRecordOption struct { ... }
func (TrimmingRecordOption) TableName() string { ... }
```

以下を**追加**する（`trimming.go` の BodyWeightUnit 定義の後に追記）:

```go
// AppointmentTrimmingDetail はトリミング予約の詳細情報（appointments の1:1拡張）
type AppointmentTrimmingDetail struct {
    ID              uint64         `gorm:"primaryKey"                                      json:"id"`
    ClinicID        uint64         `gorm:"not null"                                        json:"clinic_id"`
    AppointmentID   uint64         `gorm:"uniqueIndex;not null"                            json:"appointment_id"`
    CourseID        *uint64        `                                                       json:"course_id,omitempty"`
    StyleRequest    string         `gorm:"default:''"                                      json:"style_request"`
    BodyWeight      *float64       `gorm:"column:body_weight;type:numeric(6,2)"            json:"body_weight,omitempty"`
    BWUnit          BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'"              json:"bw_unit"`
    BodyTemperature *float64       `gorm:"column:body_temperature;type:numeric(4,1)"       json:"body_temperature,omitempty"`
    UsedShampoo     string         `gorm:"default:''"                                      json:"used_shampoo"`
    UsedRibbon      string         `gorm:"default:''"                                      json:"used_ribbon"`
    Remarks         string         `gorm:"default:''"                                      json:"remarks"`
    StyleImage      string         `gorm:"default:''"                                      json:"style_image"`
    CompletedImage  string         `gorm:"default:''"                                      json:"completed_image"`
    CreatedAt       time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
    UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`

    // Relations
    Course  *TrimmingCourse  `gorm:"foreignKey:CourseID"                              json:"course,omitempty"`
    Options []TrimmingOption `gorm:"many2many:appointment_trimming_options;joinForeignKey:AppointmentID;joinReferences:OptionID" json:"options,omitempty"`
}

func (AppointmentTrimmingDetail) TableName() string { return "appointment_trimming_details" }

// AppointmentTrimmingOption はトリミング予約とオプションの M:N 中間テーブル
type AppointmentTrimmingOption struct {
    ID            uint64    `gorm:"primaryKey" json:"id"`
    AppointmentID uint64    `gorm:"not null"   json:"appointment_id"`
    OptionID      uint64    `gorm:"not null"   json:"option_id"`
    SortOrder     int       `gorm:"default:0"  json:"sort_order"`
    CreatedAt     time.Time `                  json:"created_at"`
}

func (AppointmentTrimmingOption) TableName() string { return "appointment_trimming_options" }
```

#### 2-b. `backend/internal/model/reservation_type.go` — Category 追加

```go
// ReservationTypeCategory は予約区分のカテゴリ
type ReservationTypeCategory string

const (
    ReservationTypeCategoryGeneral  ReservationTypeCategory = "general"
    ReservationTypeCategoryTrimming ReservationTypeCategory = "trimming"
)
```

`ReservationType` struct に追加:

```go
type ReservationType struct {
    // ... 既存フィールド（変更なし） ...
    Category ReservationTypeCategory `gorm:"type:reservation_type_category;not null;default:'general'" json:"category"`

    // ... 既存リレーション + TASK-001で追加するリレーション（変更なし） ...
}
```

#### 2-c. `backend/internal/model/reservation.go` — TrimmingDetail リレーション追加

`Appointment` struct の Relations セクションに追加:

```go
// Relations
Owner           *Owner                    `gorm:"foreignKey:OwnerID"           json:"owner,omitempty"`
Pet             *Pet                      `gorm:"foreignKey:PetID"             json:"pet,omitempty"`
ReservationType *ReservationType          `gorm:"foreignKey:ReservationTypeID" json:"reservation_type,omitempty"`
Doctor          *Staff                    `gorm:"foreignKey:DoctorID"          json:"doctor,omitempty"`
CreatedByStaff  *Staff                    `gorm:"foreignKey:CreatedBy"         json:"created_by_staff,omitempty" tygo:"-"`
LineCustomer    *LineCustomer             `gorm:"foreignKey:LineCustomerID"    json:"line_customer,omitempty"`
TrimmingDetail  *AppointmentTrimmingDetail `gorm:"foreignKey:AppointmentID"    json:"trimming_detail,omitempty"` // ★ 追加
```

### 3. `make codegen` 実行

```bash
make codegen
# → frontend/src/types/generated/models.ts が自動更新される
# 更新内容:
#   - ReservationTypeCategory 型・定数が追加される
#   - AppointmentTrimmingDetail 型が追加される
#   - AppointmentTrimmingOption 型が追加される
#   - TrimmingRecord, TrimmingRecordOption, TrimmingStatus が削除される
#   - ReservationType に category フィールドが追加される
#   - Appointment に trimming_detail フィールドが追加される
```

### 4. ビルド確認

```bash
docker compose exec backend go build ./...
```

## フロントエンド影響

- `TrimmingRecord`, `TrimmingRecordOption`, `TrimmingStatus` 型が `models.ts` から削除される
  → `frontend/src/features/trimming/` 内の当該型参照は FE-253 で修正する
- `ReservationType.category` が新規追加される
- `AppointmentTrimmingDetail` が新規追加される
- `Appointment.trimming_detail` が新規追加される

## 完了条件

- [ ] `001_init.sql` から `trimming_records`, `trimming_record_options`, `trimming_status` ENUM が削除されている
- [ ] `001_init.sql` に `reservation_type_category` ENUM が追加されている
- [ ] `reservation_types` テーブル定義に `category` カラムが追加されている
- [ ] `001_init.sql` に `appointment_trimming_details` テーブルが追加されている
- [ ] `001_init.sql` に `appointment_trimming_options` テーブルが追加されている
- [ ] 両テーブルに適切なインデックスが設定されている
- [ ] Go モデル `TrimmingRecord`, `TrimmingRecordOption`, `TrimmingStatus` が削除されている
- [ ] Go モデル `AppointmentTrimmingDetail`, `AppointmentTrimmingOption` が追加されている
- [ ] `ReservationType` に `Category ReservationTypeCategory` が追加されている
- [ ] `Appointment` に `TrimmingDetail *AppointmentTrimmingDetail` が追加されている
- [ ] `make codegen` が通り `models.ts` が更新されている
- [ ] `docker compose exec backend go build ./...` が通る
- [ ] DB リセット: `docker compose exec backend go run cmd/api/main.go` でスキーマが正常に作成される
