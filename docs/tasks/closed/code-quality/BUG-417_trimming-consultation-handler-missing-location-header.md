# BUG-417: trimming_handler・consultation_handler の Create に Location ヘッダ欠落

## 概要

`trimming_handler.go` と `consultation_handler.go` の Create エンドポイントが 201 Created を返す際、
REST 規約で必須の `Location` ヘッダを設定していない。BUG-410（reservation_type_liff_handler）と同種の問題だが対象ファイルが異なる。

## 問題箇所

```go
// trimming_handler.go:139
c.JSON(http.StatusCreated, toTrimmingResponse(appt))
// Location ヘッダなし

// consultation_handler.go:139
c.JSON(http.StatusCreated, toConsultationResponse(consultation))
// Location ヘッダなし
```

## 期待する実装

```go
// 標準パターン（medicine_handler.go:87-88）
c.Header("Location", fmt.Sprintf("/v1/masters/medicines/%d", medicine.ID))
c.JSON(http.StatusCreated, toMedicineResponse(medicine))
```

## 修正方針

各ハンドラのルーティング定義（`*_routes.go`）でパスを確認した上で Location ヘッダを追加する。

```go
// trimming_handler.go — CreateTrimming
c.Header("Location", fmt.Sprintf("/v1/masters/trimming-appointments/%d", appt.ID))
c.JSON(http.StatusCreated, toTrimmingResponse(appt))

// consultation_handler.go — CreateConsultation
c.Header("Location", fmt.Sprintf("/v1/masters/consultations/%d", consultation.ID))
c.JSON(http.StatusCreated, toConsultationResponse(consultation))
```

## 影響ファイル

- `backend/internal/handler/trimming_handler.go` — 行 139
- `backend/internal/handler/consultation_handler.go` — 行 139

## 優先度

**Low** — REST 規約違反。1行ずつの修正で対応可能。

## 関連チケット

- BUG-410（reservation_type_liff_handler の同種問題）
