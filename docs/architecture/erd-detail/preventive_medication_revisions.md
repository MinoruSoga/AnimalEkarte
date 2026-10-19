<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `preventive_medication_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9215](../../../backend/migrations/001_init.sql#L9215) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9216](../../../backend/migrations/001_init.sql#L9216) |
| `observation_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9217](../../../backend/migrations/001_init.sql#L9217) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:9218](../../../backend/migrations/001_init.sql#L9218) |
| `state` | `text` | NOT NULL | — | — | [001_init.sql:9219](../../../backend/migrations/001_init.sql#L9219) |
| `source_kind` | `text` | NOT NULL | — | — | [001_init.sql:9220](../../../backend/migrations/001_init.sql#L9220) |
| `source_reference` | `text` | NULL | — | — | [001_init.sql:9221](../../../backend/migrations/001_init.sql#L9221) |
| `confirmed_at` | `timestamptz` | NULL | — | — | [001_init.sql:9222](../../../backend/migrations/001_init.sql#L9222) |
| `time_precision` | `text` | NULL | — | — | [001_init.sql:9223](../../../backend/migrations/001_init.sql#L9223) |
| `event_year` | `integer` | NULL | — | — | [001_init.sql:9224](../../../backend/migrations/001_init.sql#L9224) |
| `event_month` | `integer` | NULL | — | — | [001_init.sql:9225](../../../backend/migrations/001_init.sql#L9225) |
| `event_date` | `date` | NULL | — | — | [001_init.sql:9226](../../../backend/migrations/001_init.sql#L9226) |
| `time_description` | `text` | NULL | — | — | [001_init.sql:9227](../../../backend/migrations/001_init.sql#L9227) |
| `notes` | `text` | NULL | — | — | [001_init.sql:9228](../../../backend/migrations/001_init.sql#L9228) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9229](../../../backend/migrations/001_init.sql#L9229) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9230](../../../backend/migrations/001_init.sql#L9230) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9231](../../../backend/migrations/001_init.sql#L9231) |
| `prescription_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10089](../../../backend/migrations/001_init.sql#L10089) |
| `prescription_revision_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10090](../../../backend/migrations/001_init.sql#L10090) |
| `change_reason` | `text` | NULL | — | — | [001_init.sql:10819](../../../backend/migrations/001_init.sql#L10819) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    preventive_medication_revisions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint observation_id FK,UK
        bigint revision UK
        text state
        text source_kind
        text source_reference
        timestamptz confirmed_at
        text time_precision
        integer event_year
        integer event_month
        date event_date
        text time_description
        text notes
        bigint recorded_by FK
        timestamptz recorded_at
        bigint operation_id FK
        bigint prescription_id FK
        bigint prescription_revision_id FK
        text change_reason
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    preventive_medication_observations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint prescription_id FK,UK
    }
    preventive_prescription_revisions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint prescription_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ preventive_medication_revisions : "clinic_id"
    staffs ||..o{ preventive_medication_revisions : "recorded_by"
    preventive_medication_observations ||..o{ preventive_medication_revisions : "observation_id, clinic_id"
    operation_bindings ||..o{ preventive_medication_revisions : "operation_id, clinic_id"
    preventive_medication_observations ||..o{ preventive_medication_revisions : "observation_id, clinic_id, prescription_id"
    preventive_prescription_revisions ||..o{ preventive_medication_revisions : "prescription_revision_id, clinic_id, prescription_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `preventive_medication_revisions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9216](../../../backend/migrations/001_init.sql#L9216) |
| `preventive_medication_revisions_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9229](../../../backend/migrations/001_init.sql#L9229) |
| `preventive_medication_revisions_observation_id_clinic_id_fkey` | `observation_id, clinic_id` | [preventive_medication_observations](preventive_medication_observations.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9235](../../../backend/migrations/001_init.sql#L9235) |
| `preventive_medication_revisions_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9236](../../../backend/migrations/001_init.sql#L9236) |
| `fk_medication_observation_parent` | `observation_id, clinic_id, prescription_id` | [preventive_medication_observations](preventive_medication_observations.md) | `id, clinic_id, prescription_id` | `ON DELETE RESTRICT` | [001_init.sql:10091](../../../backend/migrations/001_init.sql#L10091) |
| `fk_medication_prescription_version` | `prescription_revision_id, clinic_id, prescription_id` | [preventive_prescription_revisions](preventive_prescription_revisions.md) | `id, clinic_id, prescription_id` | `ON DELETE RESTRICT` | [001_init.sql:10093](../../../backend/migrations/001_init.sql#L10093) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `preventive_medication_revisions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `preventive_medication_revisions_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `preventive_medication_revisions_observation_id_clinic_id_fkey`: `FOREIGN KEY (observation_id,clinic_id) REFERENCES preventive_medication_observations(id,clinic_id) ON DELETE RESTRICT`
- `preventive_medication_revisions_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`
- `fk_medication_observation_parent`: `CONSTRAINT fk_medication_observation_parent  FOREIGN KEY(observation_id,clinic_id,prescription_id) REFERENCES preventive_medication_observations(id,clinic_id,prescription_id) ON DELETE RESTRICT`
- `fk_medication_prescription_version`: `CONSTRAINT fk_medication_prescription_version  FOREIGN KEY(prescription_revision_id,clinic_id,prescription_id) REFERENCES preventive_prescription_revisions(id,clinic_id,prescription_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `preventive_medication_revisions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9215](../../../backend/migrations/001_init.sql#L9215) |
| `preventive_medication_revisions_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:9218](../../../backend/migrations/001_init.sql#L9218) |
| `preventive_medication_revisions_state_check` | check | `CHECK (state IN ('planned','taken','not_taken','unknown'))` | [001_init.sql:9219](../../../backend/migrations/001_init.sql#L9219) |
| `preventive_medication_revisions_source_kind_check` | check | `CHECK (source_kind IN ('performed_record','owner_report','legacy_unknown','other'))` | [001_init.sql:9220](../../../backend/migrations/001_init.sql#L9220) |
| `preventive_medication_revisions_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:9232](../../../backend/migrations/001_init.sql#L9232) |
| `preventive_medication_revisio_clinic_id_observation_id_revi_key` | unique | `UNIQUE (clinic_id,observation_id,revision)` | [001_init.sql:9233](../../../backend/migrations/001_init.sql#L9233) |
| `preventive_medication_revisions_check` | check | `CHECK (app_private.valid_partial_clinical_date(time_precision,event_year,event_month,event_date,time_description) IS TRUE)` | [001_init.sql:9234](../../../backend/migrations/001_init.sql#L9234) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_preventive_medication_period` | NO | — | `CREATE INDEX idx_preventive_medication_period ON preventive_medication_revisions(clinic_id,event_year,event_month)` | [001_init.sql:9239](../../../backend/migrations/001_init.sql#L9239) |
