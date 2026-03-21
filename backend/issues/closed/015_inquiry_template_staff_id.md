# 問診定型文テンプレートへの staff_id 追加確認

## 概要
`inquiry_templates` テーブルに `staff_id` カラムが存在するか確認し、存在する場合はモデル・ハンドラのレスポンスに追加する。

現状調査: `model/inquiry_template.go` の `InquiryTemplate` struct に `staff_id` フィールドは存在しない。DB スキーマ（`001_init.sql`）を確認してカラムの有無を確認する必要がある。

## 優先度
medium

## 関連テーブル
- `inquiry_templates` (id, clinic_id NOT NULL, category DEFAULT '', title NOT NULL, content DEFAULT '', is_active DEFAULT true, sort_order DEFAULT 0, created_at, updated_at)
  - `staff_id` カラムの有無を `001_init.sql` で確認
- `staffs` (staff_id の参照先、存在する場合)

## 実装内容

### 事前調査
`backend/migrations/001_init.sql` を確認し `inquiry_templates` テーブルの定義を特定する。
- `staff_id` カラムが**存在する場合**: 以下の実装を行う
- `staff_id` カラムが**存在しない場合**: 本チケットはクローズ（不要）

### モデル
`model/inquiry_template.go` に追加（`staff_id` が DB に存在する場合のみ）:
```go
type InquiryTemplate struct {
    ...
    StaffID  *uint64 `json:"staff_id,omitempty"`  // 追加
    ...
    // Relations
    Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`  // 追加
}
```

### リポジトリ
`repository/inquiry_template_repository.go` を確認し、`Preload("Staff")` が未実装であれば追加する。

### サービス
`service/inquiry_template_service.go` を確認し:
- `CreateInquiryTemplateInput.StaffID *uint64` が存在するか確認・追加
- `UpdateInquiryTemplateInput.StaffID *uint64` が存在するか確認・追加

### ハンドラ
`handler/inquiry_template_request.go` を確認し:
- `createInquiryTemplateRequest.StaffID *uint64` を追加
- `updateInquiryTemplateRequest.StaffID *uint64` を追加

`handler/inquiry_template_response.go` を確認し:
- `inquiryTemplateResponse.StaffID *uint64` を追加
- `inquiryTemplateResponse.Staff *staffSummaryResponse` を追加
- `toInquiryTemplateResponse()` で `Staff: toStaffSummary(t.Staff)` を追加

`handler/inquiry_template_handler.go` を確認し、CreateInquiryTemplate / UpdateInquiryTemplate で `staff_id` が service input に渡されているか確認。

### ルート登録
変更不要。既存の問診テンプレートルートを使用。

## 完了条件
- DB に `staff_id` カラムが存在することを確認
- `POST /v1/inquiry-templates` で `staff_id` を受け付けられる
- `PATCH /v1/inquiry-templates/:id` で `staff_id` を更新できる
- `GET /v1/inquiry-templates` / `GET /v1/inquiry-templates/:id` のレスポンスに `staff_id` と `staff` オブジェクト（ネスト）が含まれる
- `staff_id` が null の場合は `staff` フィールドを省略する（`omitempty`）
- DB に `staff_id` カラムが存在しない場合は本チケットをクローズする
