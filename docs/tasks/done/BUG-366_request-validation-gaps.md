# BUG-366: リクエストバリデーション不足（金額 min=0 / ID パース一貫性）

## 優先度: MEDIUM

## 概要

複数の `*_request.go` で金額フィールドに `min=0` バリデーションがなく、
マイナス金額がDBに保存される可能性がある。
また ID パラメータのパースが `parseIDParam` と `strconv.ParseUint` 直接で混在。

## 修正内容

### 1. 金額フィールドに min=0 追加（非ポインタ型の Create リクエストのみ）

対象ファイル（Create リクエストの非ポインタ金額フィールド）:
- estimate_request.go: Subtotal, TaxTotal, TotalAmount
- createRefundRequest: 既に `min=1` あり（対応済み）

**注意**: ポインタ型（PATCH リクエスト）には `min=0` を追加しない。
Gin の binding バリデーターは nil を 0 として扱うため、
「未指定 = バリデーションエラー」になりフロントが壊れる。

保険料（InsuranceAmount）・値引き（DiscountAmount）はマイナスが正常値なので対象外。

### 2. ID パラメータパースを parseIDParam に統一

直接 `strconv.ParseUint` しているハンドラを `parseIDParam` に置換:
- refund_handler.go（billingID パース — これは parseIDParam("id") で代用可）
- その他 parseIDParam 未使用箇所

### 3. Service 層の金額バリデーション確認

accounting_service.go には既に BUG-142 で `TotalAmount < 0` チェックがある。
estimate_service.go にも同様のチェックがあるか確認し、なければ追加。
