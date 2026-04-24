# BUG-381: 薬剤マスタ削除で連携在庫が孤児化する

**作成日**: 2026-04-15
**Status**: OPEN
**Priority**: MEDIUM (data integrity / inventory 管理)
**Affects**: `backend/internal/service/medicine_service.go`, `backend/internal/service/inventory_service.go`, `features/inventory`, `features/medicines`

---

## 概要

薬剤マスタ登録時は BUG-320 の実装により在庫レコードが**自動作成**される。しかし薬剤マスタを削除しても、連携して作成された在庫レコードは削除されず**孤児レコードとして残存**する。

参照先 `medicine_id` が存在しない状態なのに在庫一覧に「医薬品」として表示され続け、データ整合性を破壊する。

## 再現手順 (ブラウザ検証 2026-04-15)

1. `admin@noavet.jp` で `/settings/medicine` へ遷移（24 件）
2. 「+ 新規登録」 → 薬品名「ブラウザテスト薬剤_在庫連携」・単価 `1200`・剤形 `錠剤`・単位 `1錠あたり` で保存
   - `POST /api/v1/masters/medicines` **201 Created** → 薬剤一覧 24→25 件
3. `/inventory` へ遷移
   - 「ブラウザテスト薬剤_在庫連携」が自動作成されている (カテゴリ `医薬品`、在庫数 0、最低在庫 0、単位 `錠`)
   - BUG-320 の仕様どおり ✅
4. `/settings/medicine` へ戻り、作成した薬剤を **削除** (DELETE **204** + 「削除しました」トースト)
   - 薬剤一覧 25→24 件
5. `/inventory` へ遷移 → **「ブラウザテスト薬剤_在庫連携」が在庫一覧に残っている** (残存確認スクリプトで判定: `still exists`)

## 期待動作（いずれかの方針を選択すること）

### 方針 A: カスケード削除 (推奨 / 新規作成運用)
薬剤マスタ DELETE 時に連携在庫もトランザクション内で削除する。BUG-320 で「マスタと在庫を一体のもの」として設計したなら削除も一体で扱うべき。

```go
// backend/internal/service/medicine_service.go (仮)
func (s *MedicineService) Delete(ctx context.Context, id uint64) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // FK 依存チェック（診療記録で使用中は 409）
        ...
        // 連携在庫を削除
        if err := s.inventoryRepo.DeleteByMedicineIDTx(ctx, tx, id); err != nil {
            return apperrors.Wrap(err, "failed to delete linked inventory")
        }
        return s.repo.DeleteTx(ctx, tx, id)
    })
}
```

### 方針 B: 参照チェックで 409 (安全側)
在庫履歴 (入荷・払出トランザクション) が残っている場合はデータ保全のため削除ブロック。

- 在庫トランザクション 0 件 → カスケード削除
- 在庫トランザクション ≥ 1 件 → 409「この薬剤は在庫履歴があるため削除できません」

### 方針 C: 薬剤 → 在庫の結合を nullable FK 化
`inventory_items.medicine_id` を nullable にし、マスタ削除時は `medicine_id = NULL` にセット。在庫は独立したリソースとして管理する。現在設計から大きな逸脱。

## 修正不要とする場合に必要な対応

UI 側で孤児在庫が判別できるようにする必要がある:
- 在庫一覧の該当行に「マスタ削除済み」バッジ表示
- 孤児在庫の手動削除 UI を用意

## 関連

- **BUG-320**: 薬剤マスタ追加時の在庫自動作成（今回の逆方向の対応が未実装）
- `backend/internal/service/medicine_service.go` Delete メソッド
- `backend/internal/repository/inventory_repository.go` に `DeleteByMedicineID` 系メソッドが無い可能性 → 確認必要

## 影響範囲

- **データ整合性**: 孤児レコードが累積し、在庫総数や医薬品カテゴリ集計が歪む
- **UX**: 削除したはずの薬剤が在庫画面で見える → 管理者混乱
- **監査**: 薬剤マスタと在庫の trace が切れる

## 後始末

ブラウザ検証時に作成された孤児在庫レコード:
- `inventory.name = "ブラウザテスト薬剤_在庫連携"` → DB から直接削除または UI 実装後にクリーンアップ
- `inventory.name = "検証用薬剤"` (過去テストの残骸と思われる) → 同様に要クリーンアップ

## 確認事項

- [ ] `inventory_items.medicine_id` の FK 制約設定 (CASCADE / RESTRICT / SET NULL) を DB スキーマで確認
- [ ] 方針 A〜C のどれを採用するか判断
- [ ] 物販マスタ (`merchandise_items`) にも同様の連携在庫問題がないか確認（BUG-320 は薬剤のみ？）
