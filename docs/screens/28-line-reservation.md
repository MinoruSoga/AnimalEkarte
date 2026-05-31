# LINE予約設定 仕様書 (LINE Reservation Settings)

<!-- TODO: Add screenshot at docs/screens/images/28-line-reservation-settings.png -->

## 概要
- **画面の目的**: 飼い主向け LINE 予約システムの稼働状態、受付ルール、および表示内容を一元管理する。
- **URLパターン**: 
  - 基本設定: `/line-reservation/settings`
  - ページ編集: `/line-reservation/page-editor`
- **アクセス権限**: 医院管理者権限が必要（`ResourceHospitalSettings`）

---

## 画面構成

### 1. 稼働・受付ルール設定 (`/settings`)
- **システム稼働**: 「稼働中」または「メンテナンス中」の切り替え。
- **受付期間**: 当日から何日先までの予約を許可するか（例：3日前から60日先まで）。
- **予約制限**: 同日の多重予約や、月間の予約回数上限の設定。
- **時間枠モード**: 
    - **隙間なく詰める**: 予約枠を前から順番に埋める設定。
    - **自由選択**: スタッフのシフト時間内で自由に選択させる設定。

### 2. 表示ページ編集 (`/page-editor`)
LINE アプリ内で飼い主が見る画面の文言を編集します。
- **ヘッダーメッセージ**: トップページの告知文。
- **利用規約・ポリシー**: 予約時の注意事項やキャンセルポリシーの定義。

### 3. 連携クレデンシャル
- **LINE連携**: Messaging API の Channel ID, Secret, および LIFF ID の登録。
- **接続テスト**: API が正しく疎通しているか、ボタン一つで確認可能。

---

## 主要な機能

### 1. リアルタイムな空き枠計算
ここでの設定と、スタッフの勤務シフト (`/shifts`)、および既に確定している予約 (`/reservations`) を掛け合わせ、飼い主側には常に「最新の空き状況」が表示されます。

### 2. 自動通知連携
予約の完了、キャンセル、および前日のリマインドが、Messaging API を通じて飼い主の LINE へ自動送信されます。

---

## 技術仕様

### 使用コンポーネント
- **`LineReservationSettings`**: 管理者向け設定ページ。
- **`LiffApp` (別プロジェクト)**: 飼い主が操作する LINE 内アプリ。
- **`TimeslotEngine`**: 空き時間を算出するバックエンドロジック。

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/line-reservation-settings` | 稼働状態やルールの取得。 |
| PATCH | `/api/v1/line-reservation-settings` | 設定の更新。 |
| POST | `/api/v1/line-reservation/test-connection` | API 接続疎通確認。 |

---
