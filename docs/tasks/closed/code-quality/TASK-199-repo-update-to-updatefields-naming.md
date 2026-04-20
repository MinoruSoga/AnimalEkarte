# TASK-199: reservation_staff_repository / reservation_type_liff_repository — Update → UpdateFields 命名統一

## 優先度
Medium

## 対象ファイル
- `backend/internal/repository/reservation_staff_repository.go`
- `backend/internal/repository/reservation_type_liff_repository.go`

## 問題概要
全 repository の Update 系メソッドは `UpdateFields` という命名で統一されているが、
上記2ファイルのみ `Update` という命名を使用しており不一致。

| ファイル | Interface メソッド名 |
|---------|---------------------|
| `reservation_staff_repository.go` | `Update(ctx, id, fields)` |
| `reservation_type_liff_repository.go` | `Update(ctx, clinicID, id, fields)` |
| その他すべて | `UpdateFields(...)` |

## 修正方針
- Interface の `Update` を `UpdateFields` にリネーム
- 実装メソッド名も同様にリネーム
- 呼び出し側（service 層）のメソッド名も追従修正

## 完了条件
- [ ] `ReservationStaffRepository.Update` → `UpdateFields` にリネーム
- [ ] `ReservationTypeLiffRepository.Update` → `UpdateFields` にリネーム
- [ ] 対応する service 側の呼び出しを修正
- [ ] `go test ./backend/internal/...` がパス
