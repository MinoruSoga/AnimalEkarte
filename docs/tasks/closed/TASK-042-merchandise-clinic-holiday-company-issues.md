# TASK-042: merchandise_item / clinic_holiday / company ドメイン問題

## 優先度

HIGH（一部）/ MEDIUM

---

## 問題 1: merchandise_item_repository の Update が UpdateFields パターン未準拠

### ファイル
`backend/internal/repository/merchandise_item_repository.go:78-100`

### 問題
```go
// インターフェース
Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
// → 戻り値が error のみ。UpdateFields pattern（*model.MerchandiseItem, error）未準拠。
```

その結果 `merchandise_item_service.go` が Update 後に別途 `FindByID` を呼ぶ 2-query race パターンになっている。TASK-038 で occupation/inquiry_template の同問題を指摘済みだが merchandise_item は未対応。

### 修正案
```go
// インターフェース
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)

// 実装: Updates 後に FindByID して返す
// service 側: UpdateFields 戻り値を直接返し FindByID の 2-query を削除
```

---

## 問題 2: clinic_holiday_service の Set/Remove に slog.InfoContext なし

### ファイル
`backend/internal/service/clinic_holiday_service.go:36-57`

### 問題
```go
// Set()（Upsert に相当する重要操作）
func (s *clinicHolidayService) Set(ctx context.Context, clinicID uint64, date time.Time, reason string) (*model.ClinicHoliday, error) {
    holiday := &model.ClinicHoliday{...}
    result, err := s.repo.Upsert(ctx, holiday)
    if err != nil { ... }
    return result, nil  // slog なし
}

// Remove()
func (s *clinicHolidayService) Remove(ctx context.Context, clinicID uint64, date time.Time) error {
    if err := s.repo.Delete(ctx, clinicID, date); err != nil { ... }
    return nil  // slog なし
}
```

業務カレンダーの変更（休日設定・解除）はアクセス記録が必要な操作だが、両メソッドとも slog が欠落。

### 修正案
```go
// Set 成功後
slog.InfoContext(ctx, "clinic holiday set",
    slog.Uint64("clinic_id", clinicID),
    slog.String("date", date.Format("2006-01-02")))

// Remove 成功後
slog.InfoContext(ctx, "clinic holiday removed",
    slog.Uint64("clinic_id", clinicID),
    slog.String("date", date.Format("2006-01-02")))
```

---

## 問題 3: clinic_holiday_repository の FindByYearMonth が deleted_at IS NULL 欠落

### ファイル
`backend/internal/repository/clinic_holiday_repository.go:27-42`

### 問題
```go
q := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Order("date ASC")
// deleted_at IS NULL なし → 論理削除済みの休日も返される
```

休日を論理削除（Remove）した後、同月を再取得すると削除済みレコードが含まれる。論理削除が正しく機能しない。

### 修正案
```go
q := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Where("deleted_at IS NULL").
    Order("date ASC")
```

---

## 問題 4: company_service の Update slog に構造化コンテキストなし

### ファイル
`backend/internal/service/company_service.go:60`

### 問題
```go
slog.InfoContext(ctx, "company updated")
// → slog.Uint64 等の構造化フィールドなし
```

Company は singleton だが、他の service と一貫した構造化ログが必要。

### 修正案
```go
slog.InfoContext(ctx, "company updated", slog.Uint64("company_id", company.ID))
```

---

## 問題 5: company_repository の Update 戻り値が error のみ

### ファイル
`backend/internal/repository/company_repository.go:39-51`

### 問題
Company は singleton（ID 固定）であるため `UpdateFields` パターンの厳密適用は要件上難しいが、Update 後に service 側で Get() を別途呼んでいる（問題 1 同様の 2-query パターン）。

### 修正案
Company が singleton である性質を考慮し、`Update` の戻り値を `(*model.Company, error)` に変更し内部で Find して返す、または service 内の 2-query をコメントで「意図的」と明記する。後者なら Update 方針を ADR として記録する。
