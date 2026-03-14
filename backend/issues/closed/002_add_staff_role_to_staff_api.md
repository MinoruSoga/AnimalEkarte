# Add staff_role field to Staff API responses

## 概要
Staff API レスポンスに `staff_role` フィールドを追加する。

現状調査の結果: `model/staff.go` には `StaffRole StaffRole` フィールドが存在する。`handler/staff_response.go` の `staffResponse` struct にも `StaffRole string` フィールドがあり、`toStaffResponse()` で `string(s.StaffRole)` に変換して返している。つまり **staff_role は既にレスポンスに含まれている**。

ただし `staffSummaryResponse`（他ハンドラで使う簡易版）に `staff_role` が含まれていない可能性がある。各所の summary struct を確認し、`staff_role` が漏れている箇所があれば追加する。

また `POST /v1/staffs` および `PATCH /v1/staffs/:id` のリクエストで `staff_role` を受け付けているかを確認し、未実装であれば追加する。

## 優先度
high

## 関連テーブル
- `staffs` (`staff_role staff_role_type NOT NULL`)
  - enum: `veterinarian` / `nurse` / `trimmer` / `reception` / `manager`

## 実装内容

### モデル
`model/staff.go` は変更不要。`StaffRole` / `StaffRoleVeterinarian` 等の定数が定義済み。

### リポジトリ
変更不要。

### サービス
`CreateStaffInput.StaffRole` および `UpdateStaffInput.StaffRole *string`（または `*model.StaffRole`）が存在するか確認し、未実装なら追加する。

`service/validators.go` に `validateStaffRole(role string) error` を追加し、enumバリデーションを行う。

### ハンドラ
`handler/staff_request.go`:
- `createStaffRequest.StaffRole string binding:"required"` が未実装なら追加
- `updateStaffRequest.StaffRole *string` が未実装なら追加

`handler/staff_response.go`:
- `staffResponse.StaffRole string` は実装済み（確認済み）
- `staffSummaryResponse`（存在する場合）に `StaffRole string` を追加

`handler/staff_handler.go`:
- `CreateStaff` で `req.StaffRole` を `model.StaffRole` に変換して service input に渡す
- `UpdateStaff` で同様に対応

### ルート登録
変更不要。

## 完了条件
- `GET /v1/staffs` および `GET /v1/staffs/:id` のレスポンスに `staff_role` フィールドが含まれる
- `POST /v1/staffs` で `staff_role` を必須フィールドとして受け付ける
- `PATCH /v1/staffs/:id` で `staff_role` を任意フィールドとして受け付け更新できる
- 不正な `staff_role` 値（enumに存在しない値）は 400 エラーを返す
- `staffSummaryResponse` にも `staff_role` が含まれる（他エンティティのネスト先でも参照可能）
