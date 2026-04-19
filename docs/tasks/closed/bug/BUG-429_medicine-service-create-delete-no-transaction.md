# BUG-429: medicine_service の Create/Delete が複数テーブル書き込みに WithTx を使用していない

## 概要

`medicine_service.go` の `Create` と `Delete` メソッドが、薬剤マスタと在庫テーブルへの
複数テーブル書き込みを行うにもかかわらず、`transactor.WithTx` によるトランザクション制御を
使用していない。薬剤作成成功・在庫作成失敗の際にデータ不整合が発生しうる。
現在は「best-effort（警告ログのみ）」として在庫失敗を無視しているが、整合性の保証がない。

## 問題箇所

### Create（薬剤作成 + 在庫自動作成）

```go
// medicine_service.go:194-212（簡略）
func (s *medicineService) Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error) {
    // Step 1: 薬剤マスタに INSERT
    medicine, err := s.repo.Create(ctx, clinicID, &model.Medicine{...})
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to create medicine")
    }

    // Step 2: 在庫テーブルに INSERT（別テーブル）← WithTx なし
    inventoryItem := &model.InventoryItem{...}
    if err := s.inventoryRepo.Create(ctx, clinicID, inventoryItem); err != nil {
        // best-effort: 在庫作成失敗は medicine 作成エラーにしない
        slog.WarnContext(ctx, "failed to create inventory item (best-effort)", ...)
        // medicine は作成済みのまま返却 ← 在庫なし薬剤が存在しうる
    }
    return medicine, nil
}
```

### Delete（薬剤削除 + 在庫削除）

```go
// medicine_service.go:253-295（簡略）
func (s *medicineService) Delete(ctx context.Context, clinicID, id uint64) error {
    // Step 1: 依存チェック
    // Step 2: 薬剤マスタを DELETE
    if err := s.repo.Delete(ctx, clinicID, id); err != nil { ... }

    // Step 3: 在庫テーブルを DELETE（別テーブル）← WithTx なし
    if err := s.inventoryRepo.DeleteByMedicineID(ctx, clinicID, id); err != nil {
        slog.WarnContext(ctx, "failed to delete inventory item (best-effort)", ...)
        // 薬剤は削除済み・在庫は残存という不整合が発生しうる
    }
    return nil
}
```

## 他サービスとの比較（正しいトランザクション実装）

```go
// trimming_service.go（WithTx 使用例）
func (s *trimmingService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error) {
    var appt *model.Reservation
    if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
        // Step 1: 予約テーブルに INSERT
        var err error
        appt, err = s.reservation.Create(txCtx, clinicID, ...)
        if err != nil { return err }

        // Step 2: 関連テーブルに INSERT
        if err := s.trimmingDetail.Create(txCtx, appt.ID, ...); err != nil {
            return err  // ← ロールバックされる
        }
        return nil
    }); err != nil { ... }
    return appt, nil
}
```

## リスク評価

| 状況 | 現在の挙動 | 期待する挙動 |
|-----|-----------|-------------|
| 薬剤作成成功 + 在庫作成失敗 | 薬剤が作成され在庫なし（孤児） | 薬剤作成もロールバック |
| 薬剤削除成功 + 在庫削除失敗 | 薬剤なし・在庫レコード残存（孤児） | 薬剤削除もロールバック |

在庫の孤児レコードは「在庫一覧 UI から手動修復可能」とコメントにあるが、これは運用上の負担であり、システムの整合性保証ではない。

## 修正方針

`medicine_service` に `transactor` を注入し、Create/Delete をトランザクションでラップする。

```go
// medicine_service.go（修正後 Create）
func (s *medicineService) Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error) {
    var medicine *model.Medicine
    if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
        var err error
        medicine, err = s.repo.Create(txCtx, clinicID, &model.Medicine{...})
        if err != nil {
            return apperrors.Wrap(err, "failed to create medicine")
        }
        inventoryItem := &model.InventoryItem{MedicineID: medicine.ID, ...}
        if err := s.inventoryRepo.Create(txCtx, clinicID, inventoryItem); err != nil {
            return apperrors.Wrap(err, "failed to create inventory item")
            // ← 失敗したら medicine 作成もロールバック
        }
        return nil
    }); err != nil {
        return nil, err
    }
    return medicine, nil
}
```

## 影響ファイル

- `backend/internal/service/medicine_service.go` — Create（行 194-212）、Delete（行 253-295）
- `backend/internal/service/medicine_service.go` — medicineService 構造体への `transactor` フィールド追加
- `backend/cmd/api/main.go` — DI 配線（transactor を medicine_service に注入）

## 優先度

**Medium** — データ整合性の問題。現状は best-effort で運用しているが、本番データの孤児レコード発生リスクがある。
