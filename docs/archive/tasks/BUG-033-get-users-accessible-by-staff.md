# BUG-033: GET /api/v1/users が staff ユーザーに全ユーザー情報を返す（アクセス制御未実装）

## 種類
バグ（バックエンド — API アクセス制御未実装）

## 重要度
高

## 発見日
2026-03-28

## 再現手順
1. staff ユーザー（例: vet@example.com）でログイン
2. `GET /api/v1/users` を実行

## 期待動作
- HTTP 403 Forbidden が返る
- または staff は自分自身の情報のみ取得可能（自己参照のみ許可）

## 実際の動作
- HTTP 200 OK で全ユーザー（10件）の情報が返る
- staff ユーザーが他のスタッフのメールアドレス・権限グループ等の個人情報を閲覧可能

## 影響
- 情報漏洩リスク: staff が他のユーザーのメールアドレス・権限グループ情報を取得可能
- RBAC の根幹を崩す: staff が全ユーザー情報を取得できる状態
- `/settings/user-accounts` ページも staff でアクセス可能（UI 制御漏れ）

## 修正方針
### バックエンド
- `GET /api/v1/users` エンドポイントに clinic_admin 以上の権限チェックを追加
- staff ユーザーからのリクエストには HTTP 403 Forbidden を返す
- `backend/internal/handler/user_handler.go` の List ハンドラに認証チェックを追加

### フロントエンド
- `/settings/user-accounts` ページを staff が開いた際にアクセス拒否ページを表示
- サイドバーの「マスタ設定 > スタッフ/ユーザー管理」メニューを staff 向けに非表示

## 対象ファイル（推定）
- `backend/internal/handler/user_handler.go`
- `backend/internal/middleware/`（認証ミドルウェア）
- `frontend/src/features/` の user-accounts 関連コンポーネント

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-033（BE） | Backend | GET /v1/users を clinic_admin 以上のみに制限（RequireClinicAdmin ミドルウェア適用） |
