# BE-005: 問診更新API - DB未保存

## 問題
医療記録（カルテ）の「問診」タブで主訴詳細フィールドを更新して保存ボタンをクリックすると、フロントエンドは「更新しました」と表示するが、バックエンドがDBに保存していない。

## エラー発生箇所
- **エンドポイント**: `PATCH /api/v1/medical-records/:id/inquiries`
- **フロントエンド通知**: "カルテを更新しました" (表示される)
- **実際のDB**: inquiriesテーブルの更新時刻が古いまま（更新されていない）

## 問題の詳細
1. UIで主訴詳細フィールドに新しい値を入力
   - 入力値: "# どんな症状\n嘘吐、下痢"
2. 保存ボタンをクリック
3. フロントエンドに成功通知が表示される
4. **しかし** DB確認時に更新されていない
   - inquiries.updated_at: 2026-03-14 18:11:29 (古いまま)
   - テスト実施時刻: 2026-03-16 02:26:50 (新しい)

## 再現手順
1. 医療記録編集画面を開く（例: 記録ID 17）
2. 「問診」タブをクリック
3. 「主訴詳細」フィールドを編集
4. 「保存」ボタンをクリック
5. "カルテを更新しました" 通知が表示される
6. DBを確認すると、inquiries.updated_at が変更されていない

## 根本原因推定
- バックエンドの PATCH エンドポイント実装が不完全
- データベースのUPDATE文が実行されていない
- フロントエンド側で誤った成功判定（200以外のステータスを成功と判定している可能性）

## テスト環境
- 記録ID: 17
- ペットID: 15
- テスト日時: 2026-03-16

## DB確認結果
```sql
SELECT id, medical_record_id, chief_complaint, updated_at FROM inquiries WHERE medical_record_id = 17;
-- 結果:
--  id | medical_record_id | chief_complaint | updated_at
-- ----|-------------------|-----------------|---------------------------------
--  17 |                17 | 歯石除去        | 2026-03-14 18:11:29.151312+00
-- (更新されていない)
```

## リクエスト例
```json
{
  "chief_complaint": "# どんな症状\n嘘吐、下痢",
  "history": "",
  "current_medications": "",
  ...その他フィールド...
}
```

## 対応
1. バックエンドの PATCH エンドポイント実装を確認・修正
2. GORM/SQLの UPDATE処理が正常に実行されているか確認
3. エラーハンドリングを実装
4. フロントエンドのレスポンス検証ロジックを修正

---

## 🔍 実装コード確認結果（2026-03-16）

### Inquiry API 全層確認

#### handler: inquiry_handler.go
- ❌ **ファイルが存在しない**
- 注意: `inquiry_template_handler.go` は問診テンプレートマスタ用（別物）

#### service: inquiry_service.go
- ❌ **ファイルが存在しない**
- `service.go` の Services struct に `Inquiry` フィールドなし（行8-54）

#### repository: inquiry_repository.go
- ❌ **ファイルが存在しない**
- `repositories.go` の Repositories struct に `Inquiry` フィールドなし

#### model: inquiry.go
- ✅ `model.Inquiry` struct は存在（13フィールド: chief_complaint, history, current_medications, etc.）
- ✅ `TableName() = "inquiries"`
- ✅ `MedicalRecord` モデルに `Inquiry *Inquiry` リレーション定義済み（medical_record.go:35）

#### DB テーブル: inquiries
- ✅ 001_init.sql にテーブル定義あり
- ✅ 002_seed_master.sql にシードデータあり

### 根本原因（特定済み）

**Inquiry の handler/service/repository が全層未実装。API エンドポイントが存在しない。**

1. **Inquiry API が全層未実装**
   - `inquiry_handler.go` — 存在しない
   - `inquiry_service.go` — 存在しない
   - `inquiry_repository.go` — 存在しない
   - DI コンテナ（service.go, repositories.go）に Inquiry が未登録

2. **FE が間違ったエンドポイントにデータを送信している**
   - `useMedicalRecordForm.ts:103` は `PATCH /v1/medical-records/:id` に `chief_complaint` を送信
   - `updateMedicalRecordRequest` (medical_record_request.go:17-25) に `chief_complaint` フィールドは存在しない
   - Gin が未知フィールドを無視 → データは捨てられる → 200 返却

### 修正方針

#### BE 新規実装（clinical_plan を参照実装とする）

**新規ファイル 7 件:**

1. `repository/inquiry_repository.go`
   ```go
   type InquiryRepository interface {
       FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.Inquiry, error)
       Create(ctx context.Context, inquiry *model.Inquiry) error
       Update(ctx context.Context, id uint64, fields map[string]any) error
       Delete(ctx context.Context, id uint64) error
   }
   ```
   → `clinical_plan_repository.go` と同パターン

2. `service/inquiry_service.go`
   ```go
   type UpdateInquiryInput struct {
       ChiefComplaintCategoryID *uint64
       ChiefComplaint           *string
       History                  *string
       CurrentMedications       *string
       AllergyInfo              *string
       LastMeal                 *string
       LastDefecation           *string
       LastUrination            *string
       Appetite                 *model.AppetiteLevel
       WaterIntake              *model.WaterIntakeLevel
       OwnerObservations        *string
       Notes                    *string
       StaffID                  *uint64
   }

   type InquiryService interface {
       GetOrCreate(ctx context.Context, medicalRecordID uint64) (*model.Inquiry, error)
       Update(ctx context.Context, medicalRecordID uint64, input *UpdateInquiryInput) (*model.Inquiry, error)
       Delete(ctx context.Context, medicalRecordID uint64) error
   }
   ```
   → `clinical_plan_service.go` と同パターン（GetOrCreate + Update by medicalRecordID）

3. `handler/inquiry_handler.go`
   ```go
   // GET  /medical-records/:id/inquiry    → GetInquiry (GetOrCreate)
   // PATCH /medical-records/:id/inquiry   → UpdateInquiry
   // DELETE /medical-records/:id/inquiry  → DeleteInquiry
   func (h *Handler) RegisterInquiryRoutes(rg *gin.RouterGroup) {
       rg.GET("/:id/inquiry", h.GetInquiry)
       rg.PATCH("/:id/inquiry", h.UpdateInquiry)
       rg.DELETE("/:id/inquiry", h.DeleteInquiry)
   }
   ```

4. `handler/inquiry_request.go` — updateInquiryRequest struct
5. `handler/inquiry_response.go` — toInquiryResponse 関数

**既存ファイル修正 3 件:**

6. `repository/repositories.go` — `Inquiry InquiryRepository` フィールド追加 + `NewInquiryRepository(db)` 初期化追加
7. `service/service.go` — `Inquiry InquiryService` フィールド追加 + `NewInquiryService(repos.Inquiry)` 初期化追加
8. `handler/medical_record_handler.go:202` — `h.RegisterInquiryRoutes(records)` 追加

#### FE 修正
1. `features/medical-records/api/` に inquiry 用の API 関数を追加
2. `useMedicalRecordForm.ts` から `chief_complaint` フィールドを削除
3. 問診タブに独自の保存フローを実装（`PATCH /v1/medical-records/:id/inquiry` を呼ぶ）
