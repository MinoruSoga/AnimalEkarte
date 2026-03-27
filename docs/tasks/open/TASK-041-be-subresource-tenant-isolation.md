# TASK-041: BE サブリソースのテナント分離欠如修正

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: Critical
**領域**: Backend / Security

---

## 概要

認証済みユーザーが JWT の clinic_id に無関係な別テナントのサブリソースを参照・操作できる。
`extractClinicID(c)` を呼んでいない handler が4ファイルに存在する。

---

## 対象ファイルと影響範囲

### 1. `handler/record_image_handler.go`
- `ListRecordImages` (L16-34): 任意の medical_record_id の画像一覧を取得可能
- `CreateRecordImage` (L38-76): 別テナントのカルテに画像を追加可能
- `DeleteRecordImage` (L80-98): 別テナントの画像を削除可能

### 2. `handler/vital_handler.go`
- `ListVitals` (L15-33)
- `CreateVital` (L37-66)
- `UpdateVital` (L70-105)
- `DeleteVital` (L109-127)

### 3. `handler/treatment_plan_handler.go`
- 全メソッド（6つ）

### 4. `handler/billing_item_handler.go`
- `CreateBillingItem` (L15-54)
- `UpdateBillingItem` (L57-90)
- `DeleteBillingItem` (L93-105)

---

## 修正方針

各ハンドラで:
1. `clinicID, ok := extractClinicID(c)` を冒頭で呼ぶ
2. service 層に `clinicID` を渡す
3. service 層で親リソース（medical_record, billing 等）の `clinic_id` と一致するか検証

```go
// 修正例（record_image_handler.go）
func (h *Handler) ListRecordImages(c *gin.Context) {
    clinicID, ok := extractClinicID(c)  // ← 追加
    if !ok {
        return
    }
    recordID, err := strconv.ParseUint(c.Param("record_id"), 10, 64)
    // ... service.ListRecordImages(ctx, clinicID, recordID)
}
```

---

## 受入条件

- [ ] 上記4ファイル全ハンドラで `extractClinicID` を呼んでいる
- [ ] service 層で親リソースの clinic_id 検証を実施している
- [ ] 別テナントの ID を指定した場合 403 または 404 が返る（テスト追加）
- [ ] `docker compose exec backend go test ./...` 全テストパス
