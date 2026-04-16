# BUG-344: 予約フォームの担当者選択でその日出勤していないスタッフを非表示にする

## 概要
予約登録フォームの担当者選択は「アクティブなスタッフ全員」を表示しているが、選択した日にシフトが登録されていないスタッフも選択肢に含まれている。出勤していない担当者を選択すると、登録時に `errNoDoctorsOnDuty` / 容量超過エラーになるか、または担当者なし予約として誤登録される。選択肢を「選択日に出勤が登録されているスタッフのみ」に絞るべきである。

## 再現手順
1. `admin@example.com` / `password` でログイン
2. 予約カレンダー → 「予約登録」ボタンをクリック
3. 日付にシフト未登録の日（例: 翌月の未設定日）を選択する
4. 担当者フィールドのドロップダウンを開く
5. **結果**: シフト未登録の日でも全アクティブスタッフが選択肢に表示される

## 期待する動作
- 選択した日付に基づいて、シフト管理で出勤登録されているスタッフのみを担当者の選択肢に表示する
- 日付未選択の場合は全アクティブスタッフを表示する（現状と同じ）
- 出勤スタッフが 0 人の場合は「担当者なし（出勤スタッフなし）」のような案内テキストを表示する

## 現状コード

### `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:66-68,264-287`
```tsx
// line 66-68: アクティブフィルタのみ（シフトフィルタなし）
const staffItems = useMasterItems("staff");
const activeStaff = staffItems.filter((s) => s.status === "active");

// line 264-287: 全アクティブスタッフをそのまま表示
<Select value={formData.staffId} onValueChange={(v) => onChange("staffId", v)}>
  <SelectContent>
    {activeStaff.map((s) => (
      <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
    ))}
  </SelectContent>
</Select>
```

### バックエンド: 出勤者カウントは実装済みだが一覧 API がない
```go
// backend/internal/service/appointment_service.go:98-104
// CountOnDutyDoctors は「何人出勤か」の数を返すのみ
// 「誰が出勤しているか」のリストを返すエンドポイントが未実装
func checkCapacitySlotConflict(...) error {
    doctorCount, err := repo.CountOnDutyDoctors(ctx, clinicID, start)
    // ...
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:66-68,264-287` | 選択日に基づく出勤スタッフフィルタリング追加 | 未修正 |
| `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx` | 選択日を ReservationFormFields に渡す（既存） | 確認要 |
| `backend/internal/repository/shift_entry_repository.go` | `FindOnDutyStaffIDs(ctx, clinicID, date)` メソッド追加 | 未実装 |
| `backend/internal/service/appointment_service.go` または `shift_service.go` | 出勤スタッフ ID リスト取得ロジック追加 | 未実装 |
| `backend/internal/handler/reservation_handler.go` または `shift_handler.go` | `GET /v1/on-duty-staffs?date=YYYY-MM-DD` エンドポイント追加 | 未実装 |

## 修正方針

### 1. バックエンド: 出勤スタッフ一覧エンドポイントを追加
`backend/internal/repository/shift_entry_repository.go`
```go
// その日にシフトが登録されているスタッフの ID と名前を返す
func (r *shiftEntryRepository) FindOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
    var staffs []model.Staff
    err := r.db.WithContext(ctx).
        Joins("JOIN shift_entries ON shift_entries.staff_id = staffs.id"+
            " AND shift_entries.clinic_id = ? AND shift_entries.date = ?"+
            " AND shift_entries.deleted_at IS NULL", clinicID, date.Format("2006-01-02")).
        Where("staffs.clinic_id = ? AND staffs.deleted_at IS NULL", clinicID).
        Find(&staffs).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "on-duty staffs", date.Format("2006-01-02"))
    }
    return staffs, nil
}
```

`backend/internal/handler/shift_handler.go` または新規 `on_duty_handler.go`
```go
// GET /v1/on-duty-staffs?date=YYYY-MM-DD
func (h *Handler) GetOnDutyStaffs(c *gin.Context) {
    dateStr := c.Query("date")
    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput(fmt.Errorf("invalid date: %s", dateStr)))
        return
    }
    clinicID := getClinicID(c)
    staffs, err := h.service.GetOnDutyStaffs(c.Request.Context(), clinicID, date)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, staffs)
}
```

### 2. フロントエンド: 選択日に基づいて出勤スタッフを取得
```tsx
// frontend/src/features/reservations/api/get-on-duty-staffs.ts (新規)
export function useGetOnDutyStaffs(date: string | null) {
  return useQuery({
    queryKey: ["on-duty-staffs", date],
    queryFn: () => apiClient.get<Staff[]>(`/v1/on-duty-staffs?date=${date}`),
    enabled: !!date,
  });
}
```

```tsx
// ReservationFormFields.tsx: 選択日がある場合は出勤スタッフに絞る
const selectedDate = formData.start ? format(formData.start, "yyyy-MM-dd") : null;
const { data: onDutyStaffs } = useGetOnDutyStaffs(selectedDate);

const staffOptions = selectedDate && onDutyStaffs
  ? onDutyStaffs
  : activeStaff; // 日付未選択時は全アクティブスタッフ
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — マルチテナント: clinicScope（必須）
> JOIN 先テーブルの `clinic_id` フィルタが JOIN 条件に含まれているか

`shift_entries.clinic_id = ?` を JOIN 条件に含める（clinicScope 非使用の JOIN パターン）。

### `backend/CLAUDE.md` — JOIN を含む repository メソッドのレビューチェックリスト
> - [ ] JOIN 先テーブルの `clinic_id` フィルタが JOIN 条件に含まれているか
> - [ ] JOIN 先テーブルの `deleted_at IS NULL` が JOIN 条件に含まれているか

### `.claude/rules/error-handling.md` — Repository エラー変換
> **Repository**: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換。

### `.claude/rules/code-style.md` — `async-parallel`
> loader 内の独立フェッチは `Promise.all` / `Promise.allSettled` で並列実行

`useGetOnDutyStaffs` と他のマスタデータ取得が独立している場合は並列フェッチする。

### プロジェクト内参照実装
- `backend/internal/service/appointment_service.go:98-104` — `CountOnDutyDoctors` パターン（同じ shifts テーブルを参照）
- `backend/internal/repository/shift_entry_repository.go` — ShiftEntry Repository の既存 JOIN パターン

## 優先度
**High** — 出勤していない担当者を選択すると登録エラーになるか不正予約が入る可能性がある。バックエンドに `CountOnDutyDoctors` が実装済みのため、一覧返却への拡張コストは低い。

## 関連チケット
- BUG-343: 定休日カレンダー disabled（日付選択との連携）

## 関連ファイル
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:66-68,264-287` — 担当者 Select
- `backend/internal/service/appointment_service.go:98-104` — `CountOnDutyDoctors`（参照元）
- `backend/internal/repository/shift_entry_repository.go` — ShiftEntry Repository
- `backend/internal/handler/shift_handler.go` — 新エンドポイント追加先
