<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `line_reservation_settings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2632](../../../backend/migrations/001_init.sql#L2632) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:2633](../../../backend/migrations/001_init.sql#L2633) |
| `status` | `text` | NOT NULL | `'stopped'` | — | [001_init.sql:2634](../../../backend/migrations/001_init.sql#L2634) |
| `header_text` | `text` | NOT NULL | `''` | — | [001_init.sql:2637](../../../backend/migrations/001_init.sql#L2637) |
| `reservation_notice` | `text` | NOT NULL | `''` | — | [001_init.sql:2638](../../../backend/migrations/001_init.sql#L2638) |
| `cancel_notice` | `text` | NOT NULL | `''` | — | [001_init.sql:2639](../../../backend/migrations/001_init.sql#L2639) |
| `privacy_policy` | `text` | NOT NULL | `''` | — | [001_init.sql:2640](../../../backend/migrations/001_init.sql#L2640) |
| `closed_weekdays` | `jsonb` | NOT NULL | `'[]'` | — | [001_init.sql:2643](../../../backend/migrations/001_init.sql#L2643) |
| `closed_dates` | `jsonb` | NOT NULL | `'[]'` | — | [001_init.sql:2644](../../../backend/migrations/001_init.sql#L2644) |
| `national_holiday_closed` | `boolean` | NOT NULL | `false` | — | [001_init.sql:2645](../../../backend/migrations/001_init.sql#L2645) |
| `business_hours` | `jsonb` | NOT NULL | `'{"start":"0900","end":"1900"}'` | — | [001_init.sql:2646](../../../backend/migrations/001_init.sql#L2646) |
| `business_hours_by_weekday` | `jsonb` | NULL | — | — | [001_init.sql:2647](../../../backend/migrations/001_init.sql#L2647) |
| `break_hours` | `jsonb` | NOT NULL | `'[{"start":"1200","end":"1300"}]'` | — | [001_init.sql:2648](../../../backend/migrations/001_init.sql#L2648) |
| `daily_limit` | `int` | NULL | `1` | — | [001_init.sql:2649](../../../backend/migrations/001_init.sql#L2649) |
| `monthly_limit` | `int` | NULL | — | — | [001_init.sql:2650](../../../backend/migrations/001_init.sql#L2650) |
| `booking_window_max_days` | `int` | NOT NULL | `30` | — | [001_init.sql:2651](../../../backend/migrations/001_init.sql#L2651) |
| `booking_window_min_days` | `int` | NOT NULL | `2` | — | [001_init.sql:2652](../../../backend/migrations/001_init.sql#L2652) |
| `calendar_months` | `int` | NOT NULL | `2` | — | [001_init.sql:2653](../../../backend/migrations/001_init.sql#L2653) |
| `phone_number` | `text` | NOT NULL | `''` | — | [001_init.sql:2654](../../../backend/migrations/001_init.sql#L2654) |
| `notification_email` | `text` | NOT NULL | `''` | — | [001_init.sql:2655](../../../backend/migrations/001_init.sql#L2655) |
| `request_example` | `text` | NOT NULL | `''` | — | [001_init.sql:2656](../../../backend/migrations/001_init.sql#L2656) |
| `time_slot_mode` | `text` | NOT NULL | `'minimize_gaps'` | — | [001_init.sql:2657](../../../backend/migrations/001_init.sql#L2657) |
| `time_slot_interval_minutes` | `int` | NOT NULL | `15` | — | [001_init.sql:2658](../../../backend/migrations/001_init.sql#L2658) |
| `no_staff_mode` | `text` | NOT NULL | `'first_available'` | — | [001_init.sql:2659](../../../backend/migrations/001_init.sql#L2659) |
| `show_no_staff_option` | `boolean` | NOT NULL | `true` | — | [001_init.sql:2660](../../../backend/migrations/001_init.sql#L2660) |
| `additional_fields` | `jsonb` | NOT NULL | `'[         {"key":"phone","label":"電話番号","required":true,"placeholder":"例) 090-1234-5678"},         {"key":"owner_name","label":"飼い主名","required":true,"placeholder":""},         {"key":"pet_info","label":"ペットの名前と種類","required":true,"placeholder":"例) ポチ（柴犬）"},         {"key":"symptoms","label":"診察内容","required":false,"placeholder":""}     ]'` | — | [001_init.sql:2663](../../../backend/migrations/001_init.sql#L2663) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2676](../../../backend/migrations/001_init.sql#L2676) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2677](../../../backend/migrations/001_init.sql#L2677) |
| `acceptance_version` | `bigint` | NOT NULL | `1` | — | [009_line_acceptance_version.sql:14](../../../backend/migrations/009_line_acceptance_version.sql#L14) |
| `reviewed_acceptance_version` | `bigint` | NULL | — | — | [009_line_acceptance_version.sql:15](../../../backend/migrations/009_line_acceptance_version.sql#L15) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    line_reservation_settings {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        text status
        text header_text
        text reservation_notice
        text cancel_notice
        text privacy_policy
        jsonb closed_weekdays
        jsonb closed_dates
        boolean national_holiday_closed
        jsonb business_hours
        jsonb business_hours_by_weekday
        jsonb break_hours
        int daily_limit
        int monthly_limit
        int booking_window_max_days
        int booking_window_min_days
        int calendar_months
        text phone_number
        text notification_email
        text request_example
        text time_slot_mode
        int time_slot_interval_minutes
        text no_staff_mode
        boolean show_no_staff_option
        jsonb additional_fields
        timestamptz created_at
        timestamptz updated_at
        bigint acceptance_version
        bigint reviewed_acceptance_version
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o| line_reservation_settings : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `line_reservation_settings_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:2633](../../../backend/migrations/001_init.sql#L2633) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `line_reservation_settings_clinic_id_fkey`: `REFERENCES clinics(id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `line_reservation_settings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2632](../../../backend/migrations/001_init.sql#L2632) |
| `line_reservation_settings_clinic_id_key` | unique | `UNIQUE` | [001_init.sql:2633](../../../backend/migrations/001_init.sql#L2633) |
| `line_reservation_settings_acceptance_version_check` | check | `CHECK (acceptance_version > 0)` | [009_line_acceptance_version.sql:14](../../../backend/migrations/009_line_acceptance_version.sql#L14) |
| `ck_line_reservation_reviewed_acceptance_version` | check | `CONSTRAINT ck_line_reservation_reviewed_acceptance_version         CHECK (reviewed_acceptance_version IS NULL OR (reviewed_acceptance_version > 0 AND reviewed_acceptance_version <= acceptance_version))` | [009_line_acceptance_version.sql:16](../../../backend/migrations/009_line_acceptance_version.sql#L16) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
