# BUG-422: マスタ削除前の存在確認パターンが不統一（FindByID 有無の混在）

## 概要

マスタサービスの `Delete` メソッドで、削除前の存在確認（テナント検証）を行う方式が
ファイルによって異なり、統一されていない。

| パターン | 採用ファイル |
|---------|------------|
| `FindByID(ctx, clinicID, id)` で存在確認後に削除 | merchandise_item_service.go（行 188-189） |
| 存在確認なし、CountUsage/Exists で依存チェックのみ | cage_service.go、insurance_service.go、exam_type_service.go 等 |

## 問題箇所

### パターン A（FindByID あり）
```go
// merchandise_item_service.go:188-198
func (s *merchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {  // ← 存在確認
        return apperrors.Wrap(err, "failed to get merchandise item")
    }
    count, err := s.repo.CountUsageByMerchandiseItemID(ctx, clinicID, id)
    // ...
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
```

### パターン B（FindByID なし）
```go
// cage_service.go:111-126
func (s *cageService) Delete(ctx context.Context, clinicID, id uint64) error {
    exists, err := s.hospitalizationRepo.ExistsByCageID(ctx, id)   // 依存チェックのみ
    if err != nil {
        return apperrors.Wrap(err, "failed to check hospitalization dependency")
    }
    if exists {
        return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
    }
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {
```

## 影響範囲

不統一なため、以下のリスクが存在:
- パターン B では削除対象が存在しない場合、Repository の `Delete` で 404 が返るが、
  存在確認の 404 とは異なるエラーメッセージになる可能性がある
- パターン B で Repository の `Delete` に clinicID 条件が正しく実装されているか依存する

## 推奨する統一方針

**パターン A（FindByID あり）を標準とする**理由:
1. clinicID によるテナント検証が Delete 実行前に明示的に行われる
2. 「存在しない」と「依存あり」で異なるエラーメッセージを返せる
3. 規約（`Delete/Update 前に FindByID で存在確認`）に準拠

**対応が必要なサービス（FindByID 欠落）**:
- cage_service.go
- insurance_service.go
- occupation_service.go
- chief_complaint_service.go（BUG-413 と関連）
- consultation_service.go
- その他 CountUsage 系のみで存在確認をしていないサービス

## 修正方針

全マスタサービスの Delete メソッド先頭に統一パターンを追加:

```go
func (s *xxxService) Delete(ctx context.Context, clinicID, id uint64) error {
    // Step 1: 存在確認（テナント検証含む）
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to get xxx")
    }

    // Step 2: FK 依存チェック
    count, err := s.repo.CountUsageByXxxID(ctx, clinicID, id)
    if count > 0 {
        return apperrors.WrapConflict("この項目は使用中のため削除できません")
    }

    // Step 3: 削除
    return apperrors.Wrap(s.repo.Delete(ctx, clinicID, id), "failed to delete xxx")
}
```

## 優先度

**Low** — 動作上の問題は Repository 実装次第だが、規約統一のため対応が必要。

## 関連チケット

- BUG-413（chief_complaint_service の clinicID 欠落）
- BUG-420（trimming_service の Update マルチテナント検証欠落）
