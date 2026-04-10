# BUG-279: 安全に削除可能なデッドコード一覧

## 概要

設計変更やリファクタリングにより不要になったコードの一覧。
全項目は grep で呼び出し元ゼロを確認済み。

## カテゴリ A: 確実にデッドコード（即削除可）

### 1. 空の response ファイル（9ファイル）

`package handler` のみで型定義なし。ハンドラーはモデルを直接返している。

| ファイル |
|----------|
| `handler/cage_response.go` |
| `handler/examination_response.go` |
| `handler/hospitalization_response.go` |
| `handler/inquiry_response.go` |
| `handler/inventory_response.go` |
| `handler/procedure_response.go` |
| `handler/reservation_response.go` |
| `handler/vaccination_response.go` |
| `handler/vaccine_response.go` |

### 2. 未使用 Register*Routes メソッド（8メソッド）

`registerXxxRoutesWithAuth` ラッパーに置き換え済みで、旧メソッドは呼ばれていない。

| ファイル | メソッド |
|----------|---------|
| `handler/accounting_handler.go:194` | RegisterAccountingRoutes |
| `handler/examination_handler.go:210` | RegisterExaminationRoutes |
| `handler/estimate_handler.go:183` | RegisterEstimateRoutes |
| `handler/hospitalization_handler.go:252` | RegisterHospitalizationRoutes |
| `handler/inventory_handler.go:193` | RegisterInventoryRoutes |
| `handler/owner_handler.go:190` | RegisterOwnerRoutes |
| `handler/trimming_handler.go:198` | RegisterTrimmingRoutes |
| `handler/vaccination_handler.go:220` | RegisterVaccinationRoutes |

### 3. ReservationAdminFilter 構造体

`repository/reservation_admin_repository.go:15` — API 仕様変更で直接引数方式に変更済み。

### 4. FindByStaffAndTimeSlot メソッド

`repository/reservation_repository.go:22,140` — トランザクション内 `FOR UPDATE` ロックによる重複チェックに置き換え済み。
削除時は test mock からも削除が必要（`reservation_service_test.go:54`, `staff_service_test.go:88`, `service_type_service_test.go:102`）。

### 5. handler.go:208 末尾の不完全コメント

`registerPermissionGroupRoutesWithAuth` の dangling コメント。関数本体なし。

### 6. PasswordResetToken モデル

`model/auth.go:34-55` — パスワードリセット機能は未実装。TableName, IsExpired, IsUsed, IsValid メソッドを含む。
DB テーブル `password_reset_tokens` が存在しない場合は安全に削除可能。

## カテゴリ B: 判断が必要（スキャフォールド / 将来用）

### 7. StaffClinicAssignment.FindByClinicID / Update / Delete

`service/staff_clinic_assignment_service.go:14,17,18` + 実装（36-72行）

- **現在**: `FindByStaffID` と `Create` のみ使用
- **将来**: スタッフ管理画面で「クリニック別スタッフ一覧」「所属解除」を実装する際に必要
- **判断**: 今すぐ必要ないなら削除し、必要時に再実装。インターフェースが小さくなりテスト mock も簡潔になる

### 8. ExtractLiffClinicID / ExtractLiffDisplayName

`middleware/liff_auth.go:192,202` — テストコードでのみ使用。本番ハンドラーは別のヘルパーを使用。
テスト用に公開関数として残すか、テスト内に移動するかの判断が必要。

## 削除手順

1. カテゴリ A の 1〜5 を削除（空ファイル + 未使用メソッド）
2. `go vet ./...` で確認
3. カテゴリ A の 4（FindByStaffAndTimeSlot）を削除 + test mock 更新
4. `go test ./...` で確認
5. カテゴリ B は製品判断後に対応

## 優先度

**Medium** — 機能的な影響はないが、コードベースの可読性と保守性に寄与。

## 関連チケット

- BUG-277: デッドコード監査 親チケット
