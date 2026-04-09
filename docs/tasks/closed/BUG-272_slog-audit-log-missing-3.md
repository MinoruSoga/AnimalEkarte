# BUG-272: Service 層 slog 監査ログ欠落（第3波）

## 概要

BUG-257（第1波）、BUG-263（第2波）で修正されなかった CUD メソッドの slog.InfoContext 欠落。

## 影響範囲

| ファイル | 行 | メソッド |
|----------|-----|---------|
| `staff_clinic_assignment_service.go` | 44 | Create |
| `staff_clinic_assignment_service.go` | 51 | Update（Assign） |
| `staff_clinic_assignment_service.go` | 58 | Delete（Unassign） |
| `reservation_admin_service.go` | 139 | Delete |
| `reservation_course_service.go` | 120 | Update |
| `reservation_staff_service.go` | 88 | Create |
| `reservation_staff_service.go` | 125 | Delete |
| `reservation_setting_service.go` | 103 | Upsert |

**合計: 8箇所 / 5ファイル**

## 修正方針

各メソッドの成功パス末尾に追加:
```go
slog.InfoContext(ctx, "resource_name action",
    slog.Uint64("resource_id", id),
    slog.Uint64("clinic_id", clinicID))
```

## 優先度

**High** — 監査証跡の欠落。前回修正でカバーされなかった残存。

## 関連チケット

- BUG-257: slog 第1波、BUG-263: slog 第2波
- BUG-270: 第4回監査 親チケット
