# LINE予約ページ設定 仕様書 (Reservation UI Customization)

## 概要
- **画面の目的**: 飼い主が操作する LINE 予約アプリ（LIFF）内の各画面に表示される文言、ポリシー、およびデザイン補助情報の管理。
- **URLパターン**: `/line-reservation/page-editor`
- **アクセス権限**: フロントエンドのルートガードは予約管理権限（`ResourceReservations` = `reservations`、`frontend/src/app/routes/operations-routes.tsx:53-56`）。一方、保存先API（`PUT /clinics/:clinic_id/line-reservation-settings`）は `ResourceHospitalSettings`（`hospital-settings`）の `edit` 権限を要求する（`backend/internal/handler/reservation_line_routes.go:21-22`）。画面には到達できても医院設定権限がなければ保存に失敗し得る点に注意。

---

## 画面構成

### 1. コンテンツ編集フォーム（単一ページ）
`LineReservationPageEditor` は LIFF フェーズ別のタブ切り替えではなく、5つのテキストエリアを縦に並べた単一フォーム。セクション選択UI・`PageSectionTabs` のようなコンポーネントは存在しない（`frontend/src/features/line-reservation/routes/LineReservationPageEditor.tsx:34-65`）。
- **ヘッダーテキスト (`header_text`)**: LINE予約ページのヘッダーに表示するテキスト。
- **予約時の注意事項 (`reservation_notice`)**: 予約時に顧客に表示する注意事項。
- **キャンセル時の注意事項 (`cancel_notice`)**: キャンセル時に表示する注意事項。
- **プライバシーポリシー (`privacy_policy`)**: 個人情報の取り扱いに関する説明。
- **リクエスト例 (`request_example`)**: 予約リクエストの記入例。

`body_text` / `footer_text` という単一フィールドは存在しない（`line_reservation_settings` 実装基準）。

---

## 主要な機能

### 1. ダイナミック・コンテンツ
ここで保存された内容は即座に LINE プラットフォーム上の LIFF アプリに反映されます。季節ごとのキャンペーン案内や、緊急時の休診告知などを病院スタッフ自身で柔軟に行えます。

### 2. ポリシー管理
キャンセルポリシーや個人情報の取り扱いに関する文言を一元管理し、法的・運営上のリスクをコントロールします。

---

## 技術仕様

### 使用コンポーネント
- **`LineReservationPageEditor`**: メインページ。`Textarea`（shadcn/ui）5個を縦並びで配置した単一フォーム。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/clinics/:clinic_id/line-reservation-settings` | LINE予約設定（文言含む）を取得 | `hospital-settings` | `view` |
| PUT | `/api/v1/clinics/:clinic_id/line-reservation-settings` | LINE予約ページ文言を含む設定更新 | `hospital-settings` | `edit` |

---
