# BUG-LINE-001: LIFF コース取得 API パス不一致で 404

## 概要

LIFF App（LINE予約）のステップ2「コースを選択」画面で、コース一覧の取得が 404 エラーとなり予約フローが進行不能。

## 再現手順

1. `https://stg.noah-karte.com/line-reserve/3` にアクセス
2. ステップ1「お客様情報」を入力して「次へ」
3. ステップ2「コースを選択」→ **「コースの取得に失敗しました」** と表示

## 原因

Frontend と Backend のエンドポイントパスが不一致。

| 側 | パス | ステータス |
|----|------|-----------|
| **Frontend** (`liff-api.ts:37`) | `GET /api/liff/:clinicId/courses` | リクエスト送信 |
| **Backend route** (`reservation_line_routes.go:69`) | `GET /api/liff/:clinicId/types` | 登録済み |
| **Backend handler comment** (`liff_handler.go:51`) | `// GET /api/liff/:clinicId/courses` | コメントは `/courses` |

Backend のハンドラコメントは `/courses` と記載されているが、ルート登録が `/types` になっている。

## 影響

- LINE予約フロー全体がブロック（ステップ2以降に進めない）
- 城東（clinic_id=2）、八王子（clinic_id=3）両方に影響

## 修正案

`reservation_line_routes.go:69` のルート登録を `/types` → `/courses` に変更:

```go
// Before
authed.GET("/types", h.GetLiffTypes)

// After
authed.GET("/courses", h.GetLiffTypes)
```

## 優先度

**CRITICAL** — 予約フロー完全ブロック

## 確認環境

- staging: `https://stg.noah-karte.com/line-reserve/3`
- Network: `GET /api/liff/3/courses` → 404
