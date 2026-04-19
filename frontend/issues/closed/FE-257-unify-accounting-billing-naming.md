# FE-257: accounting / billing 命名を統一

**Status**: Open  
**Priority**: Low  
**Type**: Refactor  
**Date Created**: 2026-04-19  
**Related**: BE-257（BE 側対応と同時進行を推奨）

## 背景

「会計」エンティティを `accounting` と `billing` の2語で呼び分けている。
FE feature・ルートパスは `accounting` を使用しているが、
BE のモデル名・DB テーブルは `Billing` / `billings` であり、
クロスレイヤーで語彙が一致していない。

## 現状の不統一

| 場所 | 語 |
|------|-----|
| DB テーブル | `billings`、`billing_items`、`billing_confirmations` |
| BE Go struct | `type Billing struct`、`type BillingItem struct` |
| BE handler | `accounting_handler.go`（`accounting` を使用） |
| BE API ルートパス | `/v1/accountings` |
| FE feature ディレクトリ | `features/accounting/` |
| FE `paths.ts` キー | `accounting` |
| FE 型名 | `interface Accounting`（types）|
| FE 型名（混在） | `BillingConfirmation`（medical-records/types）|

## 対応方針

**`billing` に統一することを推奨**。

理由: DB テーブル・BE モデルが既に `billing` で確立されており、変更対象が少ない。
また、会計（accounting）は「月次集計・経理業務全般」の意味で使われることが多く、
個別の請求レコード（billing）と混同されやすい。

ただし、FE の feature ディレクトリ・URL パス変更はユーザー影響が大きいため、
**実施は後回し（Low priority）とし、BillingConfirmation / BillingItem などの内部型名を先に整理**する。

## 変更対象（段階的対応）

### Phase 1（Low コスト・先行実施）
- `medical-records/types/index.ts` の `BillingConfirmation` 型は既に `billing` 語彙なので維持
- BE handler 内で `accounting` と `billing` が混在している変数名を整理

### Phase 2（URL 変更を伴う・後回し）
- FE feature ディレクトリ `accounting` → `billing`
- `paths.ts` の `accounting` キー → `billing`
- `/accounting` → `/billing` URL 変更 + リダイレクト

### BE 側（BE-257 として別起票）
- `accounting_handler.go` → `billing_handler.go` リネームを検討
- BE 内部で `accounting` 語彙が残っている箇所を `billing` に統一

## 完了条件

- [ ] Phase 1: FE/BE 内部の型・変数名の混在を解消
- [ ] Phase 2: FE feature・パス名変更（別イシューで実施可）
- [ ] lint / 型チェック / ビルドが通る

## 注意事項

- DB テーブル名 `billings` / `billing_items` は変更しない
- `BillingConfirmation` は既に `billing` 語彙なので現状維持
