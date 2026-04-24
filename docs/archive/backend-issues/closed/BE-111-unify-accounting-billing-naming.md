# BE-111: accounting / billing 命名の整理

**Status**: Open  
**Priority**: Low  
**Type**: Refactor  
**Date Created**: 2026-04-19  
**Related**: FE-257

## 背景

BE 内部で `accounting` と `billing` が混在している。
DB テーブルは `billings`、モデルは `Billing` で統一されているが、
ハンドラファイル名・サービス名が `accounting_handler.go` / `AccountingService` となっており
FE の `accounting` 語彙と合わせた形になっている。

## 現状

| 場所 | 語 |
|------|-----|
| DB テーブル | `billings`（`billing` 語彙） |
| Go struct | `type Billing struct`（`billing` 語彙） |
| handler ファイル | `accounting_handler.go`（`accounting` 語彙） |
| service I/F | `AccountingService`（`accounting` 語彙） |
| repository I/F | `AccountingRepository`（`accounting` 語彙） |

## 判断

**現状維持を推奨**（Low priority・変更対効果が低い）。

理由:
- FE の `accounting` ディレクトリ・URL パスと handler/service の `accounting` 命名が一致しており、
  クロスレイヤーで見ると FE–BE 間の整合性はある。
- DB テーブル名 `billings` は GORM `TableName()` で吸収されているため、struct 名 `Billing` が
  handler 名 `Accounting` と違っていても実害は少ない。

## 対応すべき混在（Phase 1 として先行実施）

handler・service 内部で `billing` 語彙の変数名と `accounting` 語彙の変数名が混在している箇所を整理する:

```go
// NG: handler 内部で accounting と billing が混在
func (h *Handler) ListAccountings(c *gin.Context) {
    billings, total, err := h.svc.Accounting.List(...)  // svc.Accounting だが戻り値は billings
}
```

上記のような `billings` ローカル変数名を `accountings` に統一、あるいはコメントで意図を明示する。

## 完了条件

- [ ] handler 内部の変数名の混在を整理（コメント追加 or リネーム）
- [ ] Phase 2（FE パス変更含む大規模リネーム）は FE-257 と合わせて別途判断
