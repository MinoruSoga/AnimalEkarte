# 008: ServiceType カラーをDBの color フィールドと同期

**ステータス:** open
**優先度:** high
**関連API:** `GET /v1/masters/service-types`（レスポンスに `color` フィールドが追加済み）

## 背景

バックエンドの `service_type` モデル・ハンドラに `color` フィールドが追加された。
現状の `useServiceTypeColorMap` は、既知の種別名に固定色を割り当て・未知の種別はパレット回転で対応している。
仕様書では「予約種別カラーはマスタ（`service_types.color`）と連動し、動的に凡例を生成する」と明記されている。

## 問題

1. `useServiceTypeColorMap` が `MasterItem` 型（`color` フィールドなし）を使っているため、DBの `color` を参照できない
2. 予約カレンダーの凡例色がDBの設定と一致しない
3. マスタ設定のカラーピッカーで変更しても反映されない

## 実装方針

### 方針A（推奨）: serviceType専用APIレイヤーを作成
- `features/master/api/service-types.ts` を作成（staffs.ts と同様のパターン）
- `BackendServiceType.color` を含む型定義
- `useListServiceTypes()` hook
- `useServiceTypeColorMap` を `color` フィールドを持つ型から色を読む実装に変更

### 方針B（簡易）: transformGenericMasterItem を拡張
- `GenericMasterBackendItem` に `color?: string` を追加
- `transformGenericMasterItem` で `color` を pass-through
- `MasterItem` 型に `color?: string` を追加

## 影響範囲

- `frontend/src/features/master/hooks/useServiceTypeColorMap.ts`
- `frontend/src/features/reservations/` （予約カレンダーの凡例）
- `frontend/src/types/` または `features/master/api/service-types.ts`（型定義）

## 注意

方針Aを選択した場合、`useServiceTypeColorMap` の破壊的変更になるため、
予約画面（`ReservationManagement`, `MonthView`, `WeekView`）での使用箇所を確認して修正する。
