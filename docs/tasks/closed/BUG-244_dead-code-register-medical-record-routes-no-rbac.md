# BUG-244: RegisterMedicalRecordRoutes が未使用デッドコードとして残存（RBAC なし）

## 概要
`medical_record_handler.go:277` に `RegisterMedicalRecordRoutes` 関数が存在するが、実際のルーティングでは一切呼ばれていない。
呼び出し元である `handler.go:108` の `registerMedicalRecordRoutesWithAuth` とほぼ同一の内容だが、**RBAC 権限チェックが一切含まれていない**。
将来の開発者が誤って `RegisterMedicalRecordRoutes` を呼び出した場合、カルテ機能全体の RBAC が無効化される。

## 再現手順
再現は不要（デッドコードのため現時点では実害なし）。
ただし以下で確認可能:

```bash
grep -n "RegisterMedicalRecordRoutes" backend/internal/handler/handler.go
# → 結果なし（呼ばれていない）
```

## 期待する動作
- デッドコードを削除し、`registerMedicalRecordRoutesWithAuth`（`handler.go:108`）のみが存在する
- RBAC なしのルート登録関数が誤って呼ばれる余地がない

## 現状コード

### `backend/internal/handler/medical_record_handler.go:277-293`
```go
// ← 呼び出し元がない。handler.go:77 では registerMedicalRecordRoutesWithAuth を使用
func (h *Handler) RegisterMedicalRecordRoutes(rg *gin.RouterGroup) {
    records := rg.Group("/medical-records")
    records.GET("", h.ListMedicalRecords)
    records.POST("", h.CreateMedicalRecord)        // ← RequirePermission なし
    records.GET("/:id", h.GetMedicalRecord)
    records.PATCH("/:id", h.UpdateMedicalRecord)   // ← RequirePermission なし
    records.DELETE("/:id", h.DeleteMedicalRecord)  // ← RequirePermission なし

    h.RegisterVitalRoutes(records)
    h.RegisterTreatmentRoutes(records)
    h.RegisterBillingReviewRoutes(records)
    h.RegisterRecordImageRoutes(records)
    h.RegisterTreatmentPlanMedicalRecordRoutes(records)
    h.RegisterClinicalPlanRoutes(records)
    h.RegisterCheckupRoutes(records)
    h.RegisterInquiryRoutes(records)
}
```

### 比較: 正しい実装（`handler.go:108-124`）
```go
func (h *Handler) registerMedicalRecordRoutesWithAuth(rg *gin.RouterGroup) {
    records := rg.Group("/medical-records")
    records.GET("", h.ListMedicalRecords)
    records.GET("/:id", h.GetMedicalRecord)
    records.POST("", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateMedicalRecord)
    records.PATCH("/:id", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateMedicalRecord)
    records.DELETE("/:id", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteMedicalRecord)
    // ... サブルート登録
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/medical_record_handler.go:277-293` | デッドコード — 削除対象 | 未修正 |
| `backend/internal/handler/handler.go:108-124` | 正しい実装 — そのまま維持 | 正常 |

## 修正方針

### 1. デッドコード削除 — `backend/internal/handler/medical_record_handler.go:277-293`

`RegisterMedicalRecordRoutes` 関数全体（277〜293行）を削除する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約
> **Handler**: `RespondError(c, err)` で統一レスポンス。
> `ShouldBindJSON` エラー: `RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))` （全31ハンドラ統一済み）

BUG-020 で RBAC 権限チェックが全リソースの write 操作に適用されたが、デッドコードとして残存した非 RBAC バージョンは混乱の元になる。

### プロジェクト内参照実装
- `handler.go:73-95` — `RegisterRoutes` で全ルートが `registerXxxWithAuth` 関数経由で登録されており、RBAC が適用されているパターン

## 優先度
**Low** — 現時点では呼び出されておらず実害なし。ただし削除しておかないと将来のリグレッションリスクになる。

## 関連チケット
- BUG-020: 各リソースの write 操作に権限チェックを適用（クローズ済み）

## 関連ファイル
- `backend/internal/handler/medical_record_handler.go:277-293` — 削除対象のデッドコード
- `backend/internal/handler/handler.go:108-124` — 正しい実装（維持）
