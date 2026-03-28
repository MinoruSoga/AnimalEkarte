# BUG-035: ペット体重に負の値を指定しても登録される（バリデーション未実装）

## 種類
バグ（バックエンド — バリデーション未実装）

## 重要度
低

## 発見日
2026-03-28

## 再現手順
1. `PATCH /api/v1/pets/:id` に負の体重値を指定して送信
   ```json
   { "weight": -5.0 }
   ```

## 期待動作
- HTTP 400 Bad Request が返る
- エラーメッセージ: 「体重は0以上の値を入力してください」等

## 実際の動作
- HTTP 200 OK で負の体重値が登録される
- DBに weight = -5 が保存される

## 影響
- 体重として無効な値（負値）がDBに保存される
- 体重グラフや体重変化表示が不正確になる可能性

## 修正方針
### バックエンド
- `backend/internal/service/` の UpdatePet 処理で weight >= 0 の検証を追加
- 違反時は `errors.ErrInvalidInput` を返して HTTP 400 にする

### フロントエンド
- ペット編集フォームの体重フィールドに `min="0"` 属性を設定（HTMLバリデーション）

## 対象ファイル（推定）
- `backend/internal/service/pet_service.go`（または masters_service.go）
- フロントエンドのペット編集フォームコンポーネント
