# BUG-068: カルテ処置行 PATCH で quantity=0 を直接 API 送信すると 500 エラー

## 種類
バグ（バックエンドバリデーション欠如）

## 重要度
低

## 発見日
2026-03-29

## 再現手順

1. カルテ処置タブに処置行を追加する
2. API を直接呼び出す: `PATCH /api/v1/medical-records/:id/treatments/:treatmentId` に `{"quantity": 0}` を送信する

## 期待動作

- HTTP 400 Bad Request で `{"error": "quantity must be greater than 0"}` を返す

## 実際の動作

- HTTP 500 Internal Server Error が返る
- バックエンドが quantity ≤ 0 のバリデーションを実装していないため、DB 制約（CHECK 制約または業務ロジック）で内部エラーになる

## 補足

フロントエンドの `TreatmentRow.tsx` には `min=0.1` が設定されており、UI 経由では quantity=0 の送信は不可。
ただし API を直接叩くと 500 になるため、バックエンドでも 400 を返すべき。

## 修正方針

`treatment_service.go` の Update バリデーションで `quantity <= 0` の場合に `ErrInvalidInput` を返す処理を追加。

## 優先度
低（UI 経由では発生しない。API 直接呼出しのエッジケース）

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-068（BE） | Backend | treatment PATCH で quantity ≤ 0 のバリデーション追加（500 → 400） |
