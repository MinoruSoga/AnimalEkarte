<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `stay_plan_change_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9350](../../../backend/migrations/001_init.sql#L9350) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9351](../../../backend/migrations/001_init.sql#L9351) |
| `hospitalization_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9352](../../../backend/migrations/001_init.sql#L9352) |
| `care_plan_item_id` | `bigint` | NULL | — | FK | [001_init.sql:9353](../../../backend/migrations/001_init.sql#L9353) |
| `treatment_plan_id` | `bigint` | NULL | — | FK | [001_init.sql:9354](../../../backend/migrations/001_init.sql#L9354) |
| `revision` | `bigint` | NOT NULL | — | — | [001_init.sql:9355](../../../backend/migrations/001_init.sql#L9355) |
| `intent` | `text` | NOT NULL | — | — | [001_init.sql:9356](../../../backend/migrations/001_init.sql#L9356) |
| `before_values` | `jsonb` | NOT NULL | — | — | [001_init.sql:9357](../../../backend/migrations/001_init.sql#L9357) |
| `after_values` | `jsonb` | NOT NULL | — | — | [001_init.sql:9358](../../../backend/migrations/001_init.sql#L9358) |
| `reason` | `text` | NULL | — | — | [001_init.sql:9359](../../../backend/migrations/001_init.sql#L9359) |
| `confirmed_by` | `bigint` | NULL | — | FK | [001_init.sql:9360](../../../backend/migrations/001_init.sql#L9360) |
| `previous_event_id` | `bigint` | NULL | — | FK | [001_init.sql:9361](../../../backend/migrations/001_init.sql#L9361) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9362](../../../backend/migrations/001_init.sql#L9362) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9363](../../../backend/migrations/001_init.sql#L9363) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9364](../../../backend/migrations/001_init.sql#L9364) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    stay_plan_change_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint hospitalization_id FK
        bigint care_plan_item_id FK
        bigint treatment_plan_id FK
        bigint revision
        text intent
        jsonb before_values
        jsonb after_values
        text reason
        bigint confirmed_by FK
        bigint previous_event_id FK
        bigint operation_id FK
        bigint recorded_by FK
        timestamptz recorded_at
    }
    care_plan_items {
        BIGSERIAL id PK,UK
        bigint hospitalization_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    hospitalizations {
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
    treatment_plans {
        BIGSERIAL id PK,UK
        bigint hospitalization_id FK,UK
    }
    clinics ||..o{ stay_plan_change_events : "clinic_id"
    staffs o|..o{ stay_plan_change_events : "confirmed_by"
    staffs ||..o{ stay_plan_change_events : "recorded_by"
    hospitalizations ||..o{ stay_plan_change_events : "hospitalization_id, clinic_id"
    care_plan_items o|..o{ stay_plan_change_events : "care_plan_item_id, hospitalization_id"
    treatment_plans o|..o{ stay_plan_change_events : "treatment_plan_id, hospitalization_id"
    stay_plan_change_events o|..o{ stay_plan_change_events : "previous_event_id, clinic_id"
    operation_bindings ||..o{ stay_plan_change_events : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `stay_plan_change_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9351](../../../backend/migrations/001_init.sql#L9351) |
| `stay_plan_change_events_confirmed_by_fkey` | `confirmed_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9360](../../../backend/migrations/001_init.sql#L9360) |
| `stay_plan_change_events_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9363](../../../backend/migrations/001_init.sql#L9363) |
| `stay_plan_change_events_hospitalization_id_clinic_id_fkey` | `hospitalization_id, clinic_id` | [hospitalizations](hospitalizations.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9368](../../../backend/migrations/001_init.sql#L9368) |
| `stay_plan_change_events_care_plan_item_id_hospitalization__fkey` | `care_plan_item_id, hospitalization_id` | [care_plan_items](care_plan_items.md) | `id, hospitalization_id` | `ON DELETE RESTRICT` | [001_init.sql:9369](../../../backend/migrations/001_init.sql#L9369) |
| `stay_plan_change_events_treatment_plan_id_hospitalization__fkey` | `treatment_plan_id, hospitalization_id` | [treatment_plans](treatment_plans.md) | `id, hospitalization_id` | `ON DELETE RESTRICT` | [001_init.sql:9370](../../../backend/migrations/001_init.sql#L9370) |
| `stay_plan_change_events_previous_event_id_clinic_id_fkey` | `previous_event_id, clinic_id` | [stay_plan_change_events](stay_plan_change_events.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9371](../../../backend/migrations/001_init.sql#L9371) |
| `stay_plan_change_events_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9372](../../../backend/migrations/001_init.sql#L9372) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `stay_plan_change_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `stay_plan_change_events_confirmed_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `stay_plan_change_events_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `stay_plan_change_events_hospitalization_id_clinic_id_fkey`: `FOREIGN KEY (hospitalization_id,clinic_id) REFERENCES hospitalizations(id,clinic_id) ON DELETE RESTRICT`
- `stay_plan_change_events_care_plan_item_id_hospitalization__fkey`: `FOREIGN KEY (care_plan_item_id,hospitalization_id) REFERENCES care_plan_items(id,hospitalization_id) ON DELETE RESTRICT`
- `stay_plan_change_events_treatment_plan_id_hospitalization__fkey`: `FOREIGN KEY (treatment_plan_id,hospitalization_id) REFERENCES treatment_plans(id,hospitalization_id) ON DELETE RESTRICT`
- `stay_plan_change_events_previous_event_id_clinic_id_fkey`: `FOREIGN KEY (previous_event_id,clinic_id) REFERENCES stay_plan_change_events(id,clinic_id) ON DELETE RESTRICT`
- `stay_plan_change_events_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `stay_plan_change_events_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9350](../../../backend/migrations/001_init.sql#L9350) |
| `stay_plan_change_events_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:9355](../../../backend/migrations/001_init.sql#L9355) |
| `stay_plan_change_events_intent_check` | check | `CHECK (intent IN ('add','change','stop','post_completion_correction'))` | [001_init.sql:9356](../../../backend/migrations/001_init.sql#L9356) |
| `stay_plan_change_events_before_values_check` | check | `CHECK (jsonb_typeof(before_values)='object')` | [001_init.sql:9357](../../../backend/migrations/001_init.sql#L9357) |
| `stay_plan_change_events_after_values_check` | check | `CHECK (jsonb_typeof(after_values)='object')` | [001_init.sql:9358](../../../backend/migrations/001_init.sql#L9358) |
| `stay_plan_change_events_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:9365](../../../backend/migrations/001_init.sql#L9365) |
| `stay_plan_change_events_check` | check | `CHECK (num_nonnulls(care_plan_item_id,treatment_plan_id)=1)` | [001_init.sql:9366](../../../backend/migrations/001_init.sql#L9366) |
| `stay_plan_change_events_check1` | check | `CHECK (intent<>'post_completion_correction' OR (reason IS NOT NULL AND btrim(reason)<>''))` | [001_init.sql:9367](../../../backend/migrations/001_init.sql#L9367) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uq_stay_care_change_version` | YES | `care_plan_item_id IS NOT NULL` | `CREATE UNIQUE INDEX uq_stay_care_change_version ON stay_plan_change_events(clinic_id,care_plan_item_id,revision) WHERE care_plan_item_id IS NOT NULL` | [001_init.sql:9374](../../../backend/migrations/001_init.sql#L9374) |
| `uq_stay_treatment_change_version` | YES | `treatment_plan_id IS NOT NULL` | `CREATE UNIQUE INDEX uq_stay_treatment_change_version ON stay_plan_change_events(clinic_id,treatment_plan_id,revision) WHERE treatment_plan_id IS NOT NULL` | [001_init.sql:9375](../../../backend/migrations/001_init.sql#L9375) |
