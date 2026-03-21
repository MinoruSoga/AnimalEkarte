# BE-004: 見積書作成・更新API - DB未保存

## 問題
医療記録（カルテ）の「見積書」タブで見積書件名を入力して保存ボタンをクリックすると、フロントエンドは「更新しました」と表示するが、バックエンドがDBに保存していない。

## エラー発生箇所
- **エンドポイント**: `POST /api/v1/medical-records/:id/estimates` (推定) または `PATCH /api/v1/medical-records/:id/estimates/:estimate-id`
- **フロントエンド通知**: "カルテを更新しました" (表示される)
- **実際のDB**: estimatesテーブルに記録17用のデータが存在しない

## 問題の詳細
1. UIで見積書件名フィールドに値を入力
   - 入力値: "シロのワクチン接種・検査料金見積"
2. 保存ボタンをクリック
3. フロントエンドに成功通知が表示される
4. **しかし** DB確認時に記録が存在しない

## 再現手順
1. 医療記録編集画面を開く（例: 記録ID 17）
2. 「見積書」タブをクリック
3. 「見積書件名」フィールドに値を入力
4. 「保存」ボタンをクリック
5. "カルテを更新しました" 通知が表示される
6. DBを確認すると、estimatesテーブルに記録がない

## 根本原因推定
- バックエンドの POST/PATCH エンドポイント実装が不完全
- レスポンス処理でエラーをキャッチしていない
- トランザクション管理の問題（ロールバックが発生している可能性）
- フロントエンド側で誤った成功判定

## テスト環境
- 記録ID: 17
- ペットID: 15
- テスト日時: 2026-03-16

## DB確認結果
```sql
SELECT COUNT(*) FROM estimates WHERE medical_record_id = 17;
-- 結果: 0件（存在しない）
```

## リクエスト例
```json
{
  "title": "シロのワクチン接種・検査料金見積"
}
```

## 対応
1. バックエンドの POST/PATCH エンドポイント実装を確認・修正
2. トランザクション処理を確認
3. エラーハンドリングを実装
4. フロントエンドのレスポンス検証ロジックを修正

---

## 🔍 実装コード確認結果（2026-03-16）

### CreateEstimate flow analysis

#### handler: estimate_handler.go:82-122
- ✅ `c.ShouldBindJSON(&req)` でリクエストをバインド
- ✅ `service.CreateEstimateInput` に変換
- ✅ `h.svc.Estimate.Create(ctx, clinicID, input)` を呼び出し
- ✅ エラー時は `RespondError(c, err)` で返却
- ✅ 成功時は `c.JSON(http.StatusCreated, toEstimateResponse(estimate))`
- ✅ `slog.InfoContext(ctx, "estimate created", ...)` でログ出力

#### service: estimate_service.go:70-103
- ✅ `input.Title == ""` バリデーション
- ✅ `model.Estimate` struct を構築
- ✅ `s.repo.Create(ctx, estimate)` を呼び出し
- ✅ エラー時は `return nil, err`
- ✅ 成功時は `s.repo.FindByID(ctx, clinicID, estimate.ID)` で再取得して返却
- ✅ `slog.InfoContext(ctx, "estimate created", ...)` でログ出力

#### repository: estimate_repository.go:72-77
- ✅ `r.db.WithContext(ctx).Create(estimate).Error` で DB INSERT を実行
- ✅ エラー時は `apperrors.Wrap(err, "create estimate")` でラップして返却

#### API ルート登録: handler.go:74 + estimate_handler.go:191-198
- ✅ `h.RegisterEstimateRoutes(protected)` で登録
- ✅ `POST /api/v1/estimates` → `h.CreateEstimate`
- ✅ `PATCH /api/v1/estimates/:id` → `h.UpdateEstimate`

### 根本原因（特定済み）

**バックエンドは完全に正常動作する。問題はフロントエンド側にある。**

1. **見積書タブの「保存」ボタンに onClick ハンドラがない**
   - `MedicalRecordEstimate.tsx:122-128` の `<Button>保存</Button>` に `onClick` が未接続
   - ボタンをクリックしても何も起きない（API 呼び出しなし）

2. **カルテ全体の保存ボタンは見積書データを送信しない**
   - `useMedicalRecordForm.ts:101-116` の `handleSave` は `PATCH /v1/medical-records/:id` のみ呼ぶ
   - 見積書データは `updateMedicalRecord` リクエストに含まれない

3. **見積書データはモック（ハードコード）のまま**
   - `MedicalRecordEstimate.tsx:17-32` の items は `useState` ローカルステートで、API 連携なし
   - `useCreateEstimate` hook は `features/estimates/api/create-estimate.ts` に実装済みだが、`MedicalRecordEstimate` コンポーネントで使われていない

### 修正方針

#### FE 修正（medical-records feature）
1. `MedicalRecordEstimate.tsx` の保存ボタンに `onClick` ハンドラを追加
2. `useCreateEstimate` / `useUpdateEstimate` を呼び出す保存フローを実装
3. 見積書件名を入力 → 保存 → `POST /v1/estimates` に `{ title, medical_record_id, ... }` を送信
4. 注意: `estimates` は cross-feature なので、`app/pages/` で合成するか props 注入パターンを使う

#### 修正が不要なもの
- ❌ バックエンド（handler/service/repository）— 全て正常
- ❌ DB スキーマ — 正常
- ❌ ルート登録 — 正常
