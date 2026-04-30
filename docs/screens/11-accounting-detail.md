# 会計精算 仕様書

## 概要
- **画面の目的**: 診療費の計算、保険適用処理、支払い処理、領収書/診療明細書の発行
- **URLパターン**:
  - ペット選択: `/accounting/select-pet` (RequirePermission: ResourceAccounting create)
  - 新規: `/accounting/new?petId=xxx` (RequirePermission: ResourceAccounting create)
  - 既存: `/accounting/:id`
- **アクセス権限**: `ResourceAccounting` による RBAC 制御
  - `canView`: 会計詳細の閲覧
  - `canEdit`: 明細編集・会計確定・返金登録
  - `canCreate`: 新規会計作成
  - `canDelete`: キャンセル操作
- **遷移元**: `OwnerAccountingHistory` コンポーネントから各会計行をクリックして遷移

## 画面構成

### 閲覧専用バナー
- 表示条件: `id` が存在 かつ `canEdit === false`（既存会計を編集権限なしで参照）
- 実装: `role="status"` + `aria-label="閲覧専用モード"` 付き div を画面上部に表示
- EyeOff アイコンと「閲覧専用です。編集権限がありません。」テキスト

### メインレイアウト
- **ヘッダー**: 戻るボタン、帳票発行ボタン（`status === "completed"` のみ表示）
- **左カラム**: 明細一覧（`ItemListCard`）
  - カルテ連携明細（`/api/v1/billing-items/unbilled` から取得）+ 物販・その他手動追加ボタン
  - 合計金額サマリ（小計、税、合計）
- **右カラム**: 支払い・保険情報
  - ペット保険（窓口精算）: `InsuranceCard`（負担割合選択、自動計算）
  - 決済情報: `PaymentCard`（支払方法選択、お預かり・お釣り計算）

### 帳票プレビュー・印刷 Dialog
- 表示条件: `status === "completed"` の帳票発行ボタン押下
- `AccountingDocument` コンポーネントをダイアログ内で描画（インボイス対応）
- 「印刷」ボタン押下で `window.print()` を呼び出し
- 診療明細書と領収書を 1 枚に統合した「明細兼領収書」形式

### 確認ダイアログ (BUG-371)
- **編集確認**: `status === "completed"` の会計に変更を加えようとした場合 `ConfirmDialog` を表示
- **キャンセル確認**: 会計キャンセル操作時に `ConfirmDialog` (destructive variant) を表示

## 主な機能
- **集中計算ロジック**: 会計金額、消費税、保険負担額の計算はすべて `src/lib/calculations.ts` (`calculateBillingTotals`) に集約されています。これにより、画面表示と印刷用帳票で整合性を保証します。
- **React 19 アクション**: `useActionState` による会計確定。失敗時にはお預かり金額欄へ自動スクロール・フォーカスするアクセシビリティ対応。
- **物販・その他追加**: 2つのモードで品目を追加可能。
  - **マスタ選択モード** (`addMode: "master"`): 物販マスタから品目を検索・選択。
  - **手動入力モード** (`addMode: "manual"`): 未登録品目を直接入力。999,999,999円までの範囲チェックあり（BUG-072）。
- **保険窓口精算**: 明細ごとの保険適用フラグに基づき、負担割合（50%/70%/90%/100%）に応じた額を自動算出。
- **返金管理**: 会計完了済みのレコードに対して、理由を添えて部分返金を登録可能。残額計算もリアルタイムで行われる。
- **帳票発行**: インボイス対応の診療明細書および領収書のプレビュー表示。ブラウザ印刷機能と連動。

## API連携
| メソッド | エンドポイント | 用途 | 権限 |
|---------|--------------|------|------|
| GET | `/api/v1/accountings/:id` | 会計詳細取得 | view |
| POST | `/api/v1/accountings` | 会計作成 | create |
| PATCH | `/api/v1/accountings/:id` | 会計更新・確定 | edit |
| POST | `/api/v1/accountings/:id/cancel` | 会計キャンセル (BUG-371) | delete |
| GET | `/api/v1/accountings/:id/refunds` | 返金履歴取得 | view |
| POST | `/api/v1/accountings/:id/refunds` | 返金登録 | create |
| GET | `/api/v1/billing-items/unbilled` | カルテ連携未会計明細取得 | view |
| POST | `/api/v1/billing-items` | 明細追加 | create |
| PATCH | `/api/v1/billing-items/:id` | 明細更新 | edit |
| DELETE | `/api/v1/billing-items/:id` | 明細削除 | delete |
| GET | `/api/v1/masters/merchandise-items` | 物販マスタ取得 | view |
