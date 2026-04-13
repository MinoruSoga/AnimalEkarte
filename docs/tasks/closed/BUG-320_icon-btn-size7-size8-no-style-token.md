# BUG-320: 薬品マスタ新規作成時に在庫アイテムが自動作成されない

## 優先度
HIGH

## 関連セクション
- 14.19 薬品マスタ詳細管理テスト - 「薬剤マスタ追加（在庫連携あり）→ 在庫一覧に自動追加」

## 問題の説明

薬品マスタの新規作成時に、対応する在庫アイテム（inventory_items）が自動作成されない。

### 期待動作
1. `/settings/medicine` で新規薬品を作成
2. フォーム送信 → POST /api/v1/masters/medicines
3. 薬品レコード作成と同時に、`inventory_items` テーブルにも自動的に対応レコードが作成される
4. 在庫管理ページ `/inventory` を開くと、新規薬品が選択肢に追加される

### 実際の動作
- ✅ 薬品マスタレコードは正常に作成される
- ❌ 在庫アイテムが自動作成されない
- ❌ 在庫管理ページに新規薬品が選択肢に表示されない

## 技術詳細

### Backend: 実装状況
- **ファイル**: `backend/internal/service/medicine_service.go`
- **メソッド**: `func (s *MedicineService) Create(ctx context.Context, input CreateMedicineInput) (*model.Medicine, error)`
- **問題**: `Create()` メソッドにはロジックなし。`inventory_items` テーブルへの自動作成が実装されていない

### 修正方法
```go
// medicine_service.go の Create() メソッド内
func (s *MedicineService) Create(ctx context.Context, input CreateMedicineInput) (*model.Medicine, error) {
  medicine, err := s.repo.Create(ctx, &model.Medicine{...})
  if err != nil {
    return nil, apperrors.Wrap(err, "failed to create medicine")
  }

  // 新規: 在庫アイテムを自動作成
  inventoryItem := &model.InventoryItem{
    ClinicID:        input.ClinicID,
    MedicineID:      medicine.ID,
    InventoryName:   medicine.Name,
    CurrentQuantity: 0,
    MinimumQuantity: 0,
    Unit:            "錠", // デフォルト
  }
  if err := s.inventoryRepo.Create(ctx, inventoryItem); err != nil {
    slog.ErrorContext(ctx, "failed to create inventory item", "medicine_id", medicine.ID, "error", err)
    // best-effort: 薬品は作成済みなので、エラーは警告レベル
  }

  return medicine, nil
}
```

### テスト追加
- `backend/internal/service/medicine_service_test.go`
- Table-driven test で `CreateMedicineWithAutoInventory` ケースを追加

## 修正優先度
- 今のセッション内で実装推奨
- `inventory` feature との整合が必須

## テスト確認項目
- [x] 薬品新規作成後に inventory_items に自動レコード作成
- [x] 在庫管理ページで新規薬品が選択肢に表示
- [x] InventoryItem.MedicineID が正しく紐付けられている
- [x] エラーハンドリング（在庫作成失敗時も薬品は作成されたまま）
