<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `appointment_change_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:8937](../../../backend/migrations/001_init.sql#L8937) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8938](../../../backend/migrations/001_init.sql#L8938) |
| `appointment_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8939](../../../backend/migrations/001_init.sql#L8939) |
| `resulting_version` | `bigint` | NOT NULL | — | UK | [001_init.sql:8940](../../../backend/migrations/001_init.sql#L8940) |
| `intent` | `text` | NOT NULL | — | — | [001_init.sql:8941](../../../backend/migrations/001_init.sql#L8941) |
| `before_values` | `jsonb` | NOT NULL | — | — | [001_init.sql:8942](../../../backend/migrations/001_init.sql#L8942) |
| `after_values` | `jsonb` | NOT NULL | — | — | [001_init.sql:8943](../../../backend/migrations/001_init.sql#L8943) |
| `reason` | `text` | NULL | — | — | [001_init.sql:8944](../../../backend/migrations/001_init.sql#L8944) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:8945](../../../backend/migrations/001_init.sql#L8945) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:8946](../../../backend/migrations/001_init.sql#L8946) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8947](../../../backend/migrations/001_init.sql#L8947) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    appointment_change_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint appointment_id FK,UK
        bigint resulting_version UK
        text intent
        jsonb before_values
        jsonb after_values
        text reason
        bigint operation_id FK
        bigint recorded_by FK
        timestamptz recorded_at
    }
    appointments {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ appointment_change_events : "clinic_id"
    staffs ||..o{ appointment_change_events : "recorded_by"
    appointments ||..o{ appointment_change_events : "appointment_id, clinic_id"
    operation_bindings ||..o{ appointment_change_events : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `appointment_change_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8938](../../../backend/migrations/001_init.sql#L8938) |
| `appointment_change_events_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8946](../../../backend/migrations/001_init.sql#L8946) |
| `appointment_change_events_appointment_id_clinic_id_fkey` | `appointment_id, clinic_id` | [appointments](appointments.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8950](../../../backend/migrations/001_init.sql#L8950) |
| `appointment_change_events_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8951](../../../backend/migrations/001_init.sql#L8951) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `appointment_change_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `appointment_change_events_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `appointment_change_events_appointment_id_clinic_id_fkey`: `FOREIGN KEY (appointment_id,clinic_id) REFERENCES appointments(id,clinic_id) ON DELETE RESTRICT`
- `appointment_change_events_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `appointment_change_events_pkey` | primary | `PRIMARY KEY` | [001_init.sql:8937](../../../backend/migrations/001_init.sql#L8937) |
| `appointment_change_events_resulting_version_check` | check | `CHECK (resulting_version > 0)` | [001_init.sql:8940](../../../backend/migrations/001_init.sql#L8940) |
| `appointment_change_events_before_values_check` | check | `CHECK (jsonb_typeof(before_values)='object')` | [001_init.sql:8942](../../../backend/migrations/001_init.sql#L8942) |
| `appointment_change_events_after_values_check` | check | `CHECK (jsonb_typeof(after_values)='object')` | [001_init.sql:8943](../../../backend/migrations/001_init.sql#L8943) |
| `appointment_change_events_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:8948](../../../backend/migrations/001_init.sql#L8948) |
| `appointment_change_events_clinic_id_appointment_id_resultin_key` | unique | `UNIQUE (clinic_id,appointment_id,resulting_version)` | [001_init.sql:8949](../../../backend/migrations/001_init.sql#L8949) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
