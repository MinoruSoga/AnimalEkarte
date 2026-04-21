# TASK-215: medicine_service.go — Update で薬剤名を変更しても inventory_item.name が同期されない

## 優先度
High

## 対象ファイル
`backend/internal/service/medicine_service.go`

## 問題概要
`Create` 時に `medicine.name` を使って `inventory_item` を自動生成しているが、
`Update` で `medicine.name` を変更しても `inventory_item.name` が更新されない。

この結果、`medicine.name` と `inventory_item.name` が乖離し、
在庫一覧で古い薬剤名が表示される不整合が発生する。

## 現状の動作

```
Create: medicine.name = "アモキシシリン" → inventory_item.name = "アモキシシリン" ✓
Update: medicine.name = "アモキシシリン250mg" → inventory_item.name = "アモキシシリン" のまま ✗
```

## 修正方針

`Update` メソッド内で `input.Name != nil` のとき、トランザクション内で `inventory_item.name` も同時更新する。

```go
func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
    // 既存の Update ロジック
    ...

    // name 変更時は inventory_item.name も同期
    if input.Name != nil {
        if err := s.inventoryRepo.UpdateNameByMedicineName(
            ctx, clinicID, oldMedicine.Name, *input.Name,
        ); err != nil {
            return nil, apperrors.Wrap(err, "failed to sync inventory item name")
        }
    }
}
```

または `inventory_item` の表示名を `medicine_id` FK 経由で medicine から JOIN 取得するよう設計変更する（根本解決）。

## 調査が必要な点
- `inventory_item` テーブルに `medicine_id` FK カラムが存在するか確認（存在すれば JOIN 方式が可能）
- 存在しない場合は `UpdateNameByMedicineName` メソッドを repository に追加する方式を採用

## 完了条件
- [ ] `Update` で `input.Name != nil` のとき `inventory_item.name` を同期する処理を追加
- [ ] 薬剤名変更後、在庫一覧の名称が更新されることをテストで確認
- [ ] `go test ./backend/internal/...` がパス
