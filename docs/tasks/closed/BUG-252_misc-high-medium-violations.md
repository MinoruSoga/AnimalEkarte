# BUG-252: その他 High/Medium 違反まとめ

## 概要

全ドメイン監査で検出された HIGH/MEDIUM 違反のうち、個別チケット化するほどではないが対応すべきもの。

---

## HIGH 違反

### 1. examination_handler — status enum バリデーション欠如

`CreateExamination` / `UpdateExamination` で `input.Status` を `model.ExaminationStatus()` に
ダイレクトキャストしており、`validateEnum` による検証がない。無効な文字列がそのまま DB に書き込まれる。

**ファイル**: `backend/internal/handler/examination_handler.go:113-115, 142-144`

**修正**:
```go
if input.Status != "" {
    s, err := validateEnum(input.Status,
        model.ExaminationStatusPending,
        model.ExaminationStatusInProgress,
        model.ExaminationStatusResultEntered,
        model.ExaminationStatusCompleted,
        model.ExaminationStatusConfirmed,
    )
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
        return
    }
    exam.Status = s
}
```

### 2. liff_service — `_ =` によるエラー無視（2箇所）

`best-effort` とコメントがあるが、`_ =` による明示的なエラー無視は規約禁止。

**ファイル**:
- `backend/internal/service/liff_service.go:231` — `_ = s.customerRepo.UpdateAdditionalFields(...)`
- `backend/internal/service/liff_service.go:257` — `apptForNotify, _ = s.adminRepo.FindByIDForNotify(...)`

**修正**: `slog.WarnContext` でログを出力してから無視する。
```go
if err := s.customerRepo.UpdateAdditionalFields(ctx, ...); err != nil {
    slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
}
```

### 3. liff_handler — Handler 内で `c.JSON` 直接使用

`ReservationLimitError` で `c.JSON(http.StatusConflict, ...)` を直接呼んでおり、
`RespondError` 統一ルールに反する。追加フィールド (`redirect_step`) が必要なため意図的だが、
`RespondError` の拡張またはドキュメント化が必要。

**ファイル**: `backend/internal/handler/liff_handler.go:200-205`

### 4. pet_repository — Count クエリのエラー無視

`Update` の `RowsAffected == 0` パスで追加 Count クエリを発行するが、
その `.Error` を無視している。

**ファイル**: `backend/internal/handler/pet_repository.go:119-127`

**修正**:
```go
var count int64
if err := r.db.WithContext(ctx).Model(&model.Pet{}).
    Where("id = ? AND clinic_id = ?", id, clinicID).
    Count(&count).Error; err != nil {
    return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", id))
}
```

### 5. line_messaging_service — インフラ層で slog 使用

`LineMessagingService` は外部 API クライアント（インフラ層）だが `slog.InfoContext` を使用。
ログは呼び出し元の service 層で取るべき。

**ファイル**: `backend/internal/service/line_messaging_service.go:80`

---

## MEDIUM 違反

### 6. Service 層で slog 欠如

以下のサービスの `Create` / `Delete` に `slog.InfoContext` がなく、監査証跡が残らない。

| サービス | メソッド |
|---------|---------|
| `examination_service.go` | Create |
| `consultation_service.go` | Create, Delete |
| `staff_service.go` | Create |
| `service_type_service.go` | Create, Update, Delete |

### 7. examination_handler — レスポンス変換なし

`ListExaminations` / `GetExamination` が `model.Examination` を直接 `c.JSON` に返しており、
`toXxxResponse` 変換関数を使っていない。`examination_response.go` は空ファイル。

**ファイル**: `backend/internal/handler/examination_handler.go:68,87,121,168`

### 8. clinic_service — 税率 0% が弾かれるバグ

`buildClinicUpdateFields` で `r > 0 && r <= 1` の条件により税率 0% が無視される。
0% を許容する場合は `r >= 0 && r <= 1` に変更が必要。

**ファイル**: `backend/internal/service/clinic_service.go:32-79`

### 9. reservation_staff_service — N+1 の `GetByID`

`GetByID` が `FindAllByClinicID` で全スタッフを取得してから線形探索。
`FindByClinicIDAndID` メソッドの追加を推奨。

**ファイル**: `backend/internal/service/reservation_staff_service.go:58-70`

### 10. liff_service — N+1 の `GetStaffs`

スタッフ一覧取得後、各スタッフに対して `FindExcludedServiceTypes` をループ内で呼び出し。

**ファイル**: `backend/internal/service/liff_service.go:95-114, 293-304`

### 11. shift_entry_repository — ユニーク制約検出の独自文字列マッチング

`strings.Contains(errStr, "23505")` を使用。他の repository は `isUniqueConstraintErr(err)` ヘルパーを使用。

**ファイル**: `backend/internal/repository/shift_entry_repository.go:80-82`

### 12. insurance_handler — id/clinicID 取得順の逆転

`UpdateInsurance` で `id` を先に取得し `clinicID` を後に取得。他ハンドラと逆順。

**ファイル**: `backend/internal/handler/insurance_handler.go:86-93`

### 13. animal_species_repository.Delete — service 層と重複する FK チェック

Repository 内で FK チェック（ペット件数確認）を行っているが、`animal_species_service.Delete` で
既に同じチェックを実施済み。Repository はデータアクセスのみに専念すべき。

**ファイル**: `backend/internal/repository/animal_species_repository.go:77-98`

### 14. reservation_schedule_repository — `updated_at` の手動セット

`entry.UpdatedAt` を `map[string]any` に手動で含めており、GORM の自動管理を無効化するリスク。

**ファイル**: `backend/internal/repository/reservation_schedule_repository.go:113`

### 15. medical_record — CreateSubRecords の戻り値なしシグネチャ

`CreateSubRecords` が `error` を返さない void シグネチャ。handler 側で inquiry / clinical_plan が
全て失敗しても 201 を返す。best-effort の設計意図をコメントで明記すべき。

**ファイル**: `backend/internal/handler/medical_record_handler.go:195-205`

### 16. pet_service — ペット番号の採番レースコンディション

`CountByOwner` + `Create` がトランザクション外。DB の unique 制約で保護されているか確認が必要。

**ファイル**: `backend/internal/service/pet_service.go:169-173`

### 17. pet_repository — `CountByAnimalSpeciesID` の clinic_id スコープ欠如

マスタデータのFK依存チェックで使われるが、他クリニックのペット数もカウントされる。
意図的な設計かを明確にしてコメントで明記すること。

**ファイル**: `backend/internal/repository/pet_repository.go:85-93`

## 優先度
**High** (HIGH 1-5) / **Medium** (MEDIUM 6-17)

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）
