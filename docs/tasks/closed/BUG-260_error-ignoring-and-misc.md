# BUG-260: Count エラー無視・liff エラー無視・税率ハードコード・重複チェック等

## 概要

MEDIUM レベルの各種違反をまとめたチケット。

## 影響範囲

### 1. Count クエリのエラー無視（Rule 10）

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `repository/medicine_repository.go` | :109-117 | `Count(&count)` の `.Error` を評価していない |
| `repository/merchandise_item_repository.go` | :86-94 | 同上 |
| `repository/service_type_repository.go` | :71 | 同上 |

修正:
```go
if err := r.db.WithContext(ctx).Model(&model.Xxx{}).
    Where("id = ? AND clinic_id = ?", id, clinicID).
    Count(&count).Error; err != nil {
    return apperrors.FromGORM(err, "xxx", fmt.Sprintf("%d", id))
}
```

### 2. liff_service.go のエラー無視

| 行番号 | 問題 |
|--------|------|
| :250, :284 | `customer, _ := s.customerRepo.FindByID(...)` — 通知文脈のベストエフォートだが slog.Warn すべき |
| :341 | `breaks, _ := s.scheduleRepo.FindBreaksByEntryID(...)` — 同上 |
| :154 | `datesSettings, _ := ParseAvailableDatesSettings(...)` — JSON パースエラー無視 |

### 3. 税率ハードコード

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `service/hospitalization_service.go` | :215-227 | `TaxRate: 0.10` と `* 0.10` がハードコード。`item.TaxRate` を参照すべき |

### 4. 重複 FK チェック

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `repository/animal_species_repository.go` | :76-99 | Delete 内で pets 参照チェック。`animal_species_service.go` と重複 |

### 5. reservation_staff_service.go の N+1

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `service/reservation_staff_service.go` | GetByID | 全件取得 `FindAllByClinicID` → メモリフィルタリング。専用 FindByID メソッド推奨 |

### 6. staff_service.go の SetClinicAssignments 非トランザクション

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `service/staff_service.go` | :240-255 | Delete + Create が個別実行。途中失敗で孤立状態になる |

### 7. hospitalization_plan_repository.go の二重ラップ

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `repository/hospitalization_plan_repository.go` | :113-115 | トランザクション内で既にラップ済みのエラーをさらに `apperrors.Wrap` |

### 8. billing_item_service.go の recalculateTotals エラー無視

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `service/billing_item_service.go` | :196-201 | slog.Warn のみでエラーを上位に伝達しない。ベストエフォートならコメント明記 |

### 9. staff_handler.go の reloadErr 握りつぶし（2026-04-10 追加）

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `handler/staff_handler.go` | :118-120 | `reloadErr != nil` 時にエラーを無視して古い staff を返す。slog.WarnContext でログすべき |

### 10. reservation_notification_service.go の context 未使用（2026-04-10 追加）

| ファイル | 行番号 | 問題 |
|---------|--------|------|
| `service/reservation_notification_service.go` | :287 | `sendEmail(_ context.Context, ...)` で context を捨てている。SMTP ハング時に goroutine リーク |

## 優先度

**Medium** — 個別の影響は軽微だが、合計で17+箇所の規約違反。

## 関連チケット

- BUG-253: 親チケット
