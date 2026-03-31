# BUG-035b: 体重の負値をバックエンドが受け入れる

## 概要
ペット情報の体重フィールドで負の値を入力した場合、
フロントエンドは「体重は0以上の値を入力してください」と表示するが、
バックエンドAPIには検証がなく、負の値でも保存される。

## フロントエンド状態
- ✅ フロントエンドバリデーション: 「体重は0以上の値を入力してください」が表示される（BUG-035 frontend部分は修正済み）

## 残存NG（バックエンド）
- `PATCH /api/v1/pets/:id` に `weight: -5` を直接送信すると保存成功してしまう
- バックエンドの service/repository 層に weight >= 0 の検証がない

## 期待する動作
- バックエンドでも `weight >= 0` を検証し、負値の場合は 400 Bad Request を返す
- エラーメッセージ: `"weight must be greater than or equal to 0"`

## 実装場所
- `backend/internal/service/pet_service.go` または `handler/pet_handler.go`
- UpdatePetInput の weight フィールドにバリデーション追加

## 優先度
Medium（フロントバリデーション回避で不正データが入る）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-035（体重バリデーション）
- テスト確認日: 2026-03-30
