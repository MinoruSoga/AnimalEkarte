# CODE-QUALITY-001: payment_method_master 全レイヤー品質修正

## 概要

`payment_method_master` は他マスタと異なる実装パターンが複数あり、コード規約から著しく逸脱している。Handler / Service / Repository の全レイヤーで修正が必要。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/handler/payment_method_master_handler.go` | 全体 |
| `backend/internal/service/payment_method_master_service.go` | 全体 |
| `backend/internal/repository/payment_method_master_repository.go` | L94-104 |

---

## 問題一覧

### [Handler] 1. リクエスト/レスポンス型定義がハンドラファイルに混在

`createPaymentMethodRequest` / `updatePaymentMethodRequest` がハンドラファイル内（L16-25）に定義されている。  
他の全マスタは `*_request.go` / `*_response.go` に分離しており、このファイルだけ規約違反。

**修正方針**: `payment_method_master_request.go` と `payment_method_master_response.go` を新設し型定義を移動。

---

### [Handler] 2. model を直接シリアライズ（response 型なし）

`ListPaymentMethods` / `CreatePaymentMethod` / `UpdatePaymentMethod` が `model.PaymentMethodMaster` を直接 `c.JSON` に渡している。  
他の全マスタは `toXxxResponse()` + 専用 Response 型でレイヤー間の分離を保っている。  
将来スキーマ変更時に API レスポンス形状が意図せず変わるリスクがある。

**修正方針**: `paymentMethodResponse` 型と `toPaymentMethodResponse()` を `payment_method_master_response.go` に追加。

---

### [Handler] 3. RBAC リソースキーの誤分類

支払方法マスタの CRUD に `model.ResourceClosingSettings`（締め設定用）の権限を使用している（L115-121）。  
締め設定の権限を持たないスタッフが支払方法を管理できない、または意図しない権限昇格が発生するリスクがある。

**修正方針**: `model.Resource` に `ResourcePaymentMethod`（または `ResourceMasterAccounting` 等）を追加し、適切なリソースキーに変更。

---

### [Handler] 4. Reorder エンドポイント欠如

全マスタが `PATCH /reorder` を持つが、`payment_method_master` にのみ `Reorder` が存在しない。  
UI で表示順を変更できない状態。

**修正方針**: Handler / Service / Repository の全レイヤーに `Reorder` を追加し、ルートに `PATCH /payment-methods/reorder` を登録する。

---

### [Service] 5. `apperrors.Wrap` 漏れ（List / Create / Delete）

```go
// 現状: エラーを裸で返している
func (s *paymentMethodMasterService) List(...) {
    return s.repo.FindAll(ctx, clinicID)  // Wrap なし
}
func (s *paymentMethodMasterService) Delete(...) {
    return s.repo.Delete(ctx, clinicID, id)  // Wrap なし
}
```

**修正方針**: 全メソッドで `apperrors.Wrap(err, "failed to ...")` でラップする。

---

### [Service] 6. Update の空フィールドチェックが他サービスと不統一

`Update` が `len(fields) == 0` のとき他の全サービスは `apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)` を返すが、  
`payment_method_master_service.go` のみ `FindByID` の結果をそのまま返す（NoOp 扱い）。

**修正方針**: 他サービスと同じパターンに統一。

---

### [Service] 7. `buildPaymentMethodUpdateFields` の引数が値型

```go
// 現状: 値渡し
func buildPaymentMethodUpdateFields(input UpdatePaymentMethodInput) map[string]any {

// 他の全サービス: ポインタ渡し
func buildXxxUpdateFields(input *UpdateXxxInput) map[string]any {
```

**修正方針**: `*UpdatePaymentMethodInput` に変更し、`Update` メソッドのシグネチャも統一。

---

### [Service] 8. ログ位置の不整合（Create / Delete）

`Create` でログが操作前に出力されており、失敗してもログが記録される。  
他の全サービスは操作成功後にログを出力している。

**修正方針**: `slog.InfoContext` の呼び出しを操作成功後に移動。

---

### [Service] 9. 削除前の依存チェック未実装

`Delete` に支払方法が使用中かチェックする処理がない。  
DB 側で FK 違反が発生するか、GORM soft delete により孤立レコードが残る可能性がある。

**修正方針**:
```go
count, err := s.repo.CountUsageByID(ctx, clinicID, id)
if count > 0 {
    return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
}
```

---

### [Repository] 10. `CountUsageByID` の `apperrors.FromGORM` 未使用

```go
// 現状: Wrap を使用（規約違反）
return 0, apperrors.Wrap(err, "failed to count...")

// 規約: Repository は FromGORM を使用
return 0, apperrors.FromGORM(err, "payment", "")
```

---

### [Repository] 11. `CountUsageByID` に `clinicID` フィルタ欠落

`CountUsageByID(ctx context.Context, id uint64)` の引数に `clinicID` がなく、テナント分離が適用されていない。  
他クリニックの使用状況を誤カウントするリスクがある。

**修正方針**: シグネチャを `CountUsageByID(ctx context.Context, clinicID, id uint64) (int64, error)` に変更し、クエリに `clinicID` フィルタを追加。

---

## 規約参照

- `.claude/CLAUDE.md`: エラー処理の統一（1節）、マスタ削除の FK 依存チェック（1b節）
- `.claude/rules/go-language.md`: GORM PATCH パターン（4節）
- `.claude/rules/error-handling.md`: Repository → Service のエラーフロー

## テスト

- 全 CRUD（List / Create / Update / Delete）のハンドラテスト
- Service の `Delete` 依存チェックのユニットテスト（使用中 / 未使用の両ケース）
- Repository の `CountUsageByID` のユニットテスト
