<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `admission_records`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9316](../../../backend/migrations/001_init.sql#L9316) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9317](../../../backend/migrations/001_init.sql#L9317) |
| `hospitalization_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9318](../../../backend/migrations/001_init.sql#L9318) |
| `stay_case_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9319](../../../backend/migrations/001_init.sql#L9319) |
| `admission_event_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9320](../../../backend/migrations/001_init.sql#L9320) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9321](../../../backend/migrations/001_init.sql#L9321) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    admission_records {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint hospitalization_id FK
        bigint stay_case_id FK,UK
        bigint admission_event_id FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    hospitalizations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    stay_movement_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint stay_case_id FK,UK
    }
    clinics ||..o{ admission_records : "clinic_id"
    hospitalizations ||..o{ admission_records : "hospitalization_id, clinic_id"
    stay_movement_events ||..o| admission_records : "admission_event_id, clinic_id, stay_case_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `admission_records_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9317](../../../backend/migrations/001_init.sql#L9317) |
| `admission_records_hospitalization_id_clinic_id_fkey` | `hospitalization_id, clinic_id` | [hospitalizations](hospitalizations.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9324](../../../backend/migrations/001_init.sql#L9324) |
| `admission_records_admission_event_id_clinic_id_stay_case_i_fkey` | `admission_event_id, clinic_id, stay_case_id` | [stay_movement_events](stay_movement_events.md) | `id, clinic_id, stay_case_id` | `ON DELETE RESTRICT` | [001_init.sql:9325](../../../backend/migrations/001_init.sql#L9325) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `admission_records_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `admission_records_hospitalization_id_clinic_id_fkey`: `FOREIGN KEY (hospitalization_id,clinic_id) REFERENCES hospitalizations(id,clinic_id) ON DELETE RESTRICT`
- `admission_records_admission_event_id_clinic_id_stay_case_i_fkey`: `FOREIGN KEY (admission_event_id,clinic_id,stay_case_id) REFERENCES stay_movement_events(id,clinic_id,stay_case_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `admission_records_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9316](../../../backend/migrations/001_init.sql#L9316) |
| `admission_records_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:9322](../../../backend/migrations/001_init.sql#L9322) |
| `admission_records_clinic_id_stay_case_id_key` | unique | `UNIQUE (clinic_id,stay_case_id)` | [001_init.sql:9323](../../../backend/migrations/001_init.sql#L9323) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
