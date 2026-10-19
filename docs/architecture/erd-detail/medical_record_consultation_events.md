<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medical_record_consultation_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10995](../../../backend/migrations/001_init.sql#L10995) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10995](../../../backend/migrations/001_init.sql#L10995) |
| `medical_record_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10996](../../../backend/migrations/001_init.sql#L10996) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10996](../../../backend/migrations/001_init.sql#L10996) |
| `event_kind` | `text` | NOT NULL | — | — | [001_init.sql:10997](../../../backend/migrations/001_init.sql#L10997) |
| `previous_event_id` | `bigint` | NULL | — | FK | [001_init.sql:10997](../../../backend/migrations/001_init.sql#L10997) |
| `source_kind` | `text` | NOT NULL | — | — | [001_init.sql:10998](../../../backend/migrations/001_init.sql#L10998) |
| `occurred_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:10999](../../../backend/migrations/001_init.sql#L10999) |
| `reason` | `text` | NULL | — | — | [001_init.sql:10999](../../../backend/migrations/001_init.sql#L10999) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10999](../../../backend/migrations/001_init.sql#L10999) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:11000](../../../backend/migrations/001_init.sql#L11000) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:11000](../../../backend/migrations/001_init.sql#L11000) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medical_record_consultation_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint medical_record_id FK,UK
        bigint revision UK
        text event_kind
        bigint previous_event_id FK
        text source_kind
        timestamptz occurred_at
        text reason
        bigint operation_id FK
        bigint recorded_by FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    medical_records {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ medical_record_consultation_events : "clinic_id"
    staffs ||..o{ medical_record_consultation_events : "recorded_by"
    medical_records ||..o{ medical_record_consultation_events : "medical_record_id, clinic_id"
    medical_record_consultation_events o|..o{ medical_record_consultation_events : "previous_event_id, clinic_id, medical_record_id"
    operation_bindings ||..o{ medical_record_consultation_events : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medical_record_consultation_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10995](../../../backend/migrations/001_init.sql#L10995) |
| `medical_record_consultation_events_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:11000](../../../backend/migrations/001_init.sql#L11000) |
| `medical_record_consultation_ev_medical_record_id_clinic_id_fkey` | `medical_record_id, clinic_id` | [medical_records](medical_records.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:11002](../../../backend/migrations/001_init.sql#L11002) |
| `medical_record_consultation_e_previous_event_id_clinic_id__fkey` | `previous_event_id, clinic_id, medical_record_id` | [medical_record_consultation_events](medical_record_consultation_events.md) | `id, clinic_id, medical_record_id` | `ON DELETE RESTRICT` | [001_init.sql:11003](../../../backend/migrations/001_init.sql#L11003) |
| `medical_record_consultation_events_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:11004](../../../backend/migrations/001_init.sql#L11004) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medical_record_consultation_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medical_record_consultation_events_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `medical_record_consultation_ev_medical_record_id_clinic_id_fkey`: `FOREIGN KEY(medical_record_id,clinic_id) REFERENCES medical_records(id,clinic_id) ON DELETE RESTRICT`
- `medical_record_consultation_e_previous_event_id_clinic_id__fkey`: `FOREIGN KEY(previous_event_id,clinic_id,medical_record_id) REFERENCES medical_record_consultation_events(id,clinic_id,medical_record_id) ON DELETE RESTRICT`
- `medical_record_consultation_events_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medical_record_consultation_events_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10995](../../../backend/migrations/001_init.sql#L10995) |
| `medical_record_consultation_events_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10996](../../../backend/migrations/001_init.sql#L10996) |
| `medical_record_consultation_events_event_kind_check` | check | `CHECK(event_kind IN('ended','end_undone'))` | [001_init.sql:10997](../../../backend/migrations/001_init.sql#L10997) |
| `medical_record_consultation_events_source_kind_check` | check | `CHECK(source_kind IN('consultation','retrospective'))` | [001_init.sql:10998](../../../backend/migrations/001_init.sql#L10998) |
| `medical_record_consultation_e_id_clinic_id_medical_record_i_key` | unique | `UNIQUE(id,clinic_id,medical_record_id)` | [001_init.sql:11001](../../../backend/migrations/001_init.sql#L11001) |
| `medical_record_consultation_e_clinic_id_medical_record_id_r_key` | unique | `UNIQUE(clinic_id,medical_record_id,revision)` | [001_init.sql:11001](../../../backend/migrations/001_init.sql#L11001) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
