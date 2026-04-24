# TASK-198: clinic_holiday_handler.go / closing_settings_handler.go — 型定義ファイル分離

## 優先度
Medium

## 対象ファイル
- `backend/internal/handler/clinic_holiday_handler.go`
- `backend/internal/handler/closing_settings_handler.go`

## 問題概要
他のすべてのハンドラは request 型を `*_request.go`、response 型を `*_response.go` に分離している。
以下2ファイルのみ、型定義が handler ファイル本体にインライン記述されており規約が異なる。

### `clinic_holiday_handler.go`
`clinicHolidayResponse` struct と `toClinicHolidayResponse` 関数が
`clinic_holiday_handler.go` 内に直接定義されている。
→ `clinic_holiday_response.go` へ移動すべき。

### `closing_settings_handler.go`
`updateClinicSettingsRequest`・`createSpecialPeriodRequest`・`updateSpecialPeriodRequest` が
`closing_settings_handler.go` 内に直接定義されている。
→ `closing_settings_request.go` へ移動すべき。

## 修正方針
- 新規ファイル `clinic_holiday_response.go` を作成して型を移動
- 新規ファイル `closing_settings_request.go` を作成して型を移動
- handler ファイルから型定義を削除

## 完了条件
- [ ] `clinic_holiday_response.go` を作成し `clinicHolidayResponse` と `toClinicHolidayResponse` を移動
- [ ] `closing_settings_request.go` を作成し3つのリクエスト型を移動
- [ ] `go test ./backend/internal/...` がパス
