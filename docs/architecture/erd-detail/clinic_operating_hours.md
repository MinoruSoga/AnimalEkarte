<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `clinic_operating_hours`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:11668](../../../backend/migrations/001_init.sql#L11668) |
| `schema_version` | `text` | NOT NULL | — | — | [001_init.sql:11669](../../../backend/migrations/001_init.sql#L11669) |
| `version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:11670](../../../backend/migrations/001_init.sql#L11670) |
| `weekly_hours` | `jsonb` | NOT NULL | — | — | [001_init.sql:11671](../../../backend/migrations/001_init.sql#L11671) |
| `closed_dates` | `jsonb` | NOT NULL | — | — | [001_init.sql:11672](../../../backend/migrations/001_init.sql#L11672) |
| `national_holiday_closed` | `boolean` | NOT NULL | — | — | [001_init.sql:11673](../../../backend/migrations/001_init.sql#L11673) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:11674](../../../backend/migrations/001_init.sql#L11674) |
| `updated_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:11675](../../../backend/migrations/001_init.sql#L11675) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    clinic_operating_hours {
        bigint clinic_id PK,FK
        text schema_version
        bigint version
        jsonb weekly_hours
        jsonb closed_dates
        boolean national_holiday_closed
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||--o| clinic_operating_hours : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_operating_hours_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:11668](../../../backend/migrations/001_init.sql#L11668) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `clinic_operating_hours_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `clinic_operating_hours_pkey` | primary | `PRIMARY KEY` | [001_init.sql:11668](../../../backend/migrations/001_init.sql#L11668) |
| `clinic_operating_hours_schema_version_check` | check | `CHECK (schema_version = 'clinic-operating-hours/v1')` | [001_init.sql:11669](../../../backend/migrations/001_init.sql#L11669) |
| `clinic_operating_hours_version_check` | check | `CHECK (version > 0)` | [001_init.sql:11670](../../../backend/migrations/001_init.sql#L11670) |
| `clinic_operating_hours_weekly_hours_check` | check | `CHECK (jsonb_typeof(weekly_hours) = 'array' AND jsonb_array_length(weekly_hours) = 7)` | [001_init.sql:11671](../../../backend/migrations/001_init.sql#L11671) |
| `clinic_operating_hours_closed_dates_check` | check | `CHECK (jsonb_typeof(closed_dates) = 'array')` | [001_init.sql:11672](../../../backend/migrations/001_init.sql#L11672) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
