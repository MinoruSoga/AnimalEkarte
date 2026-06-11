# 予約区分マスタ 仕様書 (Reservation Types)

## 概要
- **画面 of Purpose**: 診察、ワクチン、手術、トリミング等の予約枠（スロット）の定義と、LINE 予約への公開設定。
- **URLパターン**: `/settings/reservation-type`
- **アクセス権限**: 予約管理権限が必要（`ResourceMasterReservationType`）

---

## 1. 画面構成

### 1.1 予約区分一覧
- **区分グループ**: 「一般診療」「トリミング」「特殊枠」等の大分類。
- **表示項目**: 名称、標準所要時間、カレンダー色、LINE 公開状況。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **基本属性**: 名称、略称、標準所要時間（15 分単位）。
- **ビジュアル設定**: カレンダー上で表示されるカラーバッジの選択。
- **LINE 予約連携**:
    - **ページに表示**: オンにすると飼い主向け LIFF アプリの選択肢に出現。
    - **所要時間（LINE）**: 院内用とは別に、LINE 予約時専用の所要時間を設定可能。
- **予約可能枠**: 予約可能な開始時刻（毎週／特定日）をリスト形式で追加・削除。「カレンダーで編集」リンクから [LINE予約枠カレンダーページ](../28-line-reservation.md)（`/line-reservation/slots?typeId=:id`）へ遷移し、月カレンダーで日別に編集できる。
- **予約不可時間**: 予約を受け付けない時間帯（毎週／特定日）を時間範囲で登録。
- **対応職種**: この予約区分を担当できる職種を紐付け（1 件以上紐付けると、担当可能スタッフが勤務する日のみ予約可能になる）。

> ⚠️ **予約可能枠のホワイトリスト挙動**: 予約可能枠を 1 件でも登録すると、その予約区分は登録された開始時刻のみ予約可能になり、枠のない日は予約不可になる。詳細は [LINE予約設定 §4](../28-line-reservation.md) を参照。

---

## 主要な機能

### 1. カレンダー・スケジューリング
ここで設定された「標準所要時間」は、予約作成時のデフォルト枠サイズとして適用され、スムーズな予約入力を支援します。

### 2. カラーコーディング
診療内容に応じた色分け（例：手術は赤、ワクチンは緑）により、カレンダーを俯瞰した際の院内の忙しさやリソース配置を一目で把握可能にします。

---

## 技術仕様

### 3.1 構成コンポーネント
- **`ReservationTypeSettings`**: メインコンテナ。
- **`ColorPicker`**: 意味的カラーパレットからの選択。
- **`TimeDurationInput`**: 時間枠の数値入力（ステップ制御）。

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/masters/reservation-types` | 有効な予約区分一覧の取得。 |
| PATCH | `/api/v1/masters/reservation-types/:id` | 名称、時間、色の更新。 |
| GET / POST / DELETE | `/api/v1/masters/reservation-types/:id/available-slots[/:slotId]` | 予約可能枠の取得・追加・削除。 |
| GET / POST / DELETE | `/api/v1/masters/reservation-types/:id/unavailable-times[/:timeId]` | 予約不可時間の取得・追加・削除。 |
| GET / POST / DELETE | `/api/v1/masters/reservation-types/:id/occupations[/:occupationId]` | 対応職種の取得・紐付け・解除。 |

---
