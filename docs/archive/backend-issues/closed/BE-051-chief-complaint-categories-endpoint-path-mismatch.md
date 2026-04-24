# BE-051: 主訴マスタ 404エラー — エンドポイントパス不一致

## 問題

`/settings/interview/chief-complaint` ページで主訴マスタデータが取得できない。

```
GET /api/v1/masters/chief-complaint-categories → 404 Not Found
```

## 根本原因

フロントエンドとバックエンドでエンドポイントパスが不一致。

| 側 | パス |
|---|---|
| フロントエンド（呼び出し側） | `/v1/masters/chief-complaint-categories` |
| バックエンド（登録側） | `/v1/masters/chief-complaints` |

**バックエンド登録箇所**: `backend/internal/handler/staff_handler.go:274-278`

```go
masters.GET("/chief-complaints", h.ListChiefComplaints)
masters.POST("/chief-complaints", h.CreateChiefComplaint)
// ...
```

**フロントエンド呼び出し箇所**:
- `frontend/src/features/master/api/chief-complaint-categories.ts:55`
- `frontend/src/features/medical-records/api/get-chief-complaint-categories.ts:12`

## 修正方針

バックエンド側のルートパスをフロントエンドに合わせる（変更箇所が1ファイルで済む）。

```go
// 修正前
masters.GET("/chief-complaints", h.ListChiefComplaints)
masters.POST("/chief-complaints", h.CreateChiefComplaint)
masters.GET("/chief-complaints/:id", h.GetChiefComplaint)
masters.PATCH("/chief-complaints/:id", h.UpdateChiefComplaint)
masters.DELETE("/chief-complaints/:id", h.DeleteChiefComplaint)

// 修正後
masters.GET("/chief-complaint-categories", h.ListChiefComplaints)
masters.POST("/chief-complaint-categories", h.CreateChiefComplaint)
masters.GET("/chief-complaint-categories/:id", h.GetChiefComplaint)
masters.PATCH("/chief-complaint-categories/:id", h.UpdateChiefComplaint)
masters.DELETE("/chief-complaint-categories/:id", h.DeleteChiefComplaint)
```

## 対象ファイル

- `backend/internal/handler/staff_handler.go` — ルート登録（274-278行目）

## テスト

```bash
curl -s http://localhost:8080/api/v1/masters/chief-complaint-categories | jq .
```
