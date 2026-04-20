# TASK-191: clinic_handler.go — モデル型を直接レスポンスに返却している

## 優先度
High

## 対象ファイル
`backend/internal/handler/clinic_handler.go`

## 問題概要
`ListClinics`・`ListClinicsByStaffID`・`GetClinic`・`CreateClinic` が
`model.Clinic` / `[]model.Clinic` を `c.JSON` に直接渡している。

DB スキーマ変更（カラム追加・削除・リネーム）が即座に API 破壊につながるリスクがある。
`toClinicResponse` 変換関数は定義済みだが、一部のハンドラで使用されていない。

## 現状コード

```go
// clinic_handler.go:29（NG）
c.JSON(http.StatusOK, clinics)   // []model.Clinic をそのまま返す

// clinic_handler.go:186（NG）
c.JSON(http.StatusCreated, result) // *model.Clinic をそのまま返す
```

## あるべき姿

```go
// ListClinics
c.JSON(http.StatusOK, lo.Map(clinics, func(c model.Clinic, _ int) ClinicResponse {
    return toClinicResponse(c)
}))

// CreateClinic
c.JSON(http.StatusCreated, toClinicResponse(*result))
```

`toClinicResponse` が定義済みであれば、全ハンドラで統一して使用するだけでよい。
未定義のレスポンスフィールドがあれば `clinic_response.go` に追加する。

## 完了条件
- [ ] `ListClinics` が `toClinicResponse` 経由で返却
- [ ] `ListClinicsByStaffID` が `toClinicResponse` 経由で返却
- [ ] `GetClinic` が `toClinicResponse` 経由で返却
- [ ] `CreateClinic` が `toClinicResponse` 経由で返却
- [ ] `go test ./backend/internal/...` がパス
