# BUG-266: Model json タグ欠落 + シークレット露出

## 概要

2つの独立した問題:
1. `vital.go` の `VitalRecord` に json タグが全フィールドで欠落。API レスポンスが PascalCase で返る。
2. `reservation_setting.go` の `LineChannelSecret` / `LineAccessToken` が json シリアライズ可能な状態。API レスポンスで LINE シークレットが漏洩する。

## 脆弱性分類（reservation_setting のみ）
- **CWE-200**: Exposure of Sensitive Information to an Unauthorized Actor
- **OWASP**: A02:2021 Cryptographic Failures
- **影響**: LINE チャネルシークレット・アクセストークンが API レスポンスに含まれ、外部に漏洩

## 影響範囲

### 1. `backend/internal/model/vital.go:12-32`

全フィールドに `json` タグがない:
```go
type VitalRecord struct {
    ID              uint64 `gorm:"primaryKey;autoIncrement"`     // ← json タグなし
    PetID           uint64 `gorm:"not null"`                     // ← json タグなし
    MedicalRecordID *uint64                                      // ← json タグなし
    ...
}
```

**比較**: 他の全モデル（`examination.go`, `vaccination.go` 等）は全フィールドに `json:"snake_case"` タグがある。

### 2. `backend/internal/model/reservation_setting.go:33,35`

```go
LineChannelSecret string `gorm:"not null;default:''" json:"line_channel_secret"` // ← シークレットが露出
LineAccessToken   string `gorm:"not null;default:''" json:"line_access_token"`   // ← シークレットが露出
```

## 修正方針

### 1. vital.go — json タグ追加

```go
type VitalRecord struct {
    ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
    PetID           uint64         `gorm:"not null"                 json:"pet_id"`
    MedicalRecordID *uint64        `                                json:"medical_record_id"`
    DailyRecordID   *uint64        `                                json:"daily_record_id"`
    RecordedAt      time.Time      `gorm:"not null;default:now()"   json:"recorded_at"`
    StaffID         *uint64        `                                json:"staff_id"`
    Temperature     *float64       `gorm:"type:numeric"             json:"temperature"`
    HeartRate       *int           `gorm:"type:integer"             json:"heart_rate"`
    RespirationRate *int           `gorm:"type:integer"             json:"respiration_rate"`
    Weight          *float64       `gorm:"type:numeric"             json:"weight"`
    WeightUnit      BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'" json:"weight_unit"`
    Notes           string         `gorm:"not null;default:''"      json:"notes"`
    CreatedAt       time.Time      `gorm:"not null;default:now()"   json:"created_at"`
    UpdatedAt       time.Time      `gorm:"not null;default:now()"   json:"updated_at"`

    Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
    MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"-"`
    DailyRecord   *DailyRecord   `gorm:"foreignKey:DailyRecordID"   json:"-"`
    Staff         *Staff         `gorm:"foreignKey:StaffID"         json:"staff,omitempty"`
}
```

### 2. reservation_setting.go — シークレットフィールドを json:"-" に変更

```go
LineChannelSecret string `gorm:"not null;default:''" json:"-"`
LineAccessToken   string `gorm:"not null;default:''" json:"-"`
```

handler/service 層でシークレットが必要な場合は、専用の内部 DTO を使用し、API レスポンス DTO からは除外する。

## 準拠すべきプロジェクト規約

### `.claude/rules/security.md` — Logging / Secrets
> Never log sensitive data (passwords, tokens)

### `.claude/rules/code-style.md` — Go Naming Conventions
> JSON tags must use snake_case

## 優先度

**Medium（reservation_setting）** — handler 層の response DTO で LineChannelSecret/LineAccessToken は既に除外されている（`reservation_setting_response.go:38` にコメントあり）。ただし model に `json:"-"` がないため、他の場所でモデルを直接シリアライズした場合に漏洩するリスクが残る。防御的措置として `json:"-"` を付けるべき。
**High（vital.go）** — フロントエンドが snake_case を期待するため、API 互換性が壊れる。

## 関連チケット

- BUG-261: 第3回監査 親チケット
