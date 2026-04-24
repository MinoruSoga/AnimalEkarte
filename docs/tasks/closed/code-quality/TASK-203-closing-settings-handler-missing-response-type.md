# TASK-203: closing_settings_handler.go — model / service struct を直接 c.JSON に渡している

## 優先度
Medium

## 対象ファイル
`backend/internal/handler/closing_settings_handler.go`

## 問題概要
以下のハンドラが専用 response 型を経由せず、model 型や service struct を直接 `c.JSON` に渡している。

| 行（概算） | ハンドラ | 渡している型 |
|-----------|---------|-------------|
| 53 | `GetClosingSettings` | `*ClosingSettingsResponse`（service 層の struct） |
| 78 | `UpdateClosingSettings` | `*model.ClinicSettings` |
| 93 | `ListSpecialPeriods` | `[]model.ClosingSpecialPeriod` |
| 121 | `CreateSpecialPeriod` | `*model.ClosingSpecialPeriod` |
| 153 | `UpdateSpecialPeriod` | `*model.ClosingSpecialPeriod` |

他ドメイン（cage, animal_species 等）はすべて `to{Entity}Response` 変換関数を経由しており、
DB スキーマ変更が API レスポンスに漏出しないよう保護されている。

## 修正方針
1. `closing_settings_response.go` を新規作成
2. 必要な response 型（`ClosingSettingsResponse`, `SpecialPeriodResponse`）を定義
3. `to{Entity}Response` 変換関数を実装
4. handler が変換関数経由でレスポンスを返すよう修正

## 注意
既に `ClosingSettingsResponse` が service 層に定義されている場合は、
handler 層の response 型と分離して再定義するか、handler の response.go に移動すること。
Service 層は HTTP レスポンス形状を知るべきでない。

## 完了条件
- [ ] `closing_settings_response.go` を作成し response 型を定義
- [ ] `to{Entity}Response` 変換関数を実装
- [ ] 全5ハンドラが変換関数経由でレスポンスを返すよう修正
- [ ] `go test ./backend/internal/...` がパス
