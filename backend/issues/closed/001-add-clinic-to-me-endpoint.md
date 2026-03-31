# 001: /me エンドポイントに clinic 情報を含める

## 概要

`GET /me` のレスポンスに、認証ユーザーが所属する医院の情報を追加する。

## 背景

現状、フロントエンドのサイドバー・会計書類ヘッダーなどで医院名を表示するために
`GET /v1/clinics`（全件取得）を呼び出し、先頭1件を使っている。

これは以下の問題を持つ:
- 全件取得して [0] を使うだけの無駄なクエリ
- マルチテナント環境で「自分の医院」の保証がない
- 認証情報と医院情報が別々にフェッチされキャッシュが分散する

## 修正内容

`/me` レスポンスに `clinic` フィールドを追加する。

```json
{
  "id": 1,
  "name": "田中 太郎",
  "email": "admin@example.com",
  "role": "admin",
  "clinic": {
    "id": 3,
    "name": "ノア動物病院",
    "logo_url": null,
    "phone_number": "03-1234-5678",
    "address": "東京都..."
  }
}
```

## 影響範囲

- `internal/handler/auth_handler.go` — `/me` ハンドラのレスポンス修正
- `internal/handler/auth_response.go` — `MeResponse` に `ClinicResponse` フィールド追加
- `internal/service/auth_service.go` — clinic 情報を取得して返す
- `backend/docs/api.yaml` — スキーマ更新

## 優先度

medium

## 関連イシュー

- frontend/issues/open/021-remove-use-clinic-info-hook.md
