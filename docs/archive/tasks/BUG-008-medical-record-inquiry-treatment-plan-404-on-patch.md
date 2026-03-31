# BUG-008: カルテ問診・治療プランの PATCH が 404 を返す

## 種類
バグ（バックエンド API 不備）

## 発見日
2026-03-23

## 再現手順
1. カルテ詳細 `/medical-records/:id` を開く
2. 問診（主訴）と診察/治療プランを入力
3. 「保存」ボタンをクリック

## 期待動作
- `PATCH /v1/medical-records/:id/inquiries` → HTTP 200
- `PATCH /v1/medical-records/:id/treatment-plans` → HTTP 200

## 実際の動作
- `PATCH /v1/medical-records/:id` → HTTP 200（主記録は保存成功）
- `PATCH /v1/medical-records/:id/inquiries` → **HTTP 404**
- `PATCH /v1/medical-records/:id/treatment-plans` → **HTTP 404**

UI に「更新対象が見つかりません。」トーストが表示される。

## 根本原因
カルテ新規作成時に `inquiries` / `treatment_plans` レコードが自動生成されない、
または PATCH ハンドラが既存レコードの有無を確認せず 404 を返している。
既存レコードが存在しない場合に INSERT（upsert）する実装が必要。

## 影響範囲
- 全カルテの問診・診察プラン保存が機能しない
- 重症度: **高**（カルテの中核機能が保存できない）

## 修正方針
以下いずれかの対応が必要：

**案A: PATCH をupsert化**
- `PATCH /v1/medical-records/:id/inquiries` でレコードが存在しなければ INSERT
- `PATCH /v1/medical-records/:id/treatment-plans` も同様

**案B: カルテ作成時に空レコードを同時生成**
- `POST /v1/medical-records` 成功時に `inquiries` / `treatment_plans` の空行を自動 INSERT

## 関連
- バックエンドイシュー: `backend/issues/open/` に BE-053 として起票すること
