<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medicine_dose_pattern_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9149](../../../backend/migrations/001_init.sql#L9149) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9150](../../../backend/migrations/001_init.sql#L9150) |
| `pattern_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9151](../../../backend/migrations/001_init.sql#L9151) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:9152](../../../backend/migrations/001_init.sql#L9152) |
| `pattern_name` | `text` | NULL | — | — | [001_init.sql:9153](../../../backend/migrations/001_init.sql#L9153) |
| `purpose` | `text` | NULL | — | — | [001_init.sql:9154](../../../backend/migrations/001_init.sql#L9154) |
| `administration_route` | `text` | NULL | — | — | [001_init.sql:9155](../../../backend/migrations/001_init.sql#L9155) |
| `parameter_snapshot` | `jsonb` | NOT NULL | — | — | [001_init.sql:9156](../../../backend/migrations/001_init.sql#L9156) |
| `source_reference` | `text` | NULL | — | — | [001_init.sql:9157](../../../backend/migrations/001_init.sql#L9157) |
| `approval_reference` | `text` | NULL | — | — | [001_init.sql:9158](../../../backend/migrations/001_init.sql#L9158) |
| `approved_by` | `bigint` | NULL | — | FK | [001_init.sql:9159](../../../backend/migrations/001_init.sql#L9159) |
| `approved_at` | `timestamptz` | NULL | — | — | [001_init.sql:9160](../../../backend/migrations/001_init.sql#L9160) |
| `is_active` | `boolean` | NOT NULL | — | — | [001_init.sql:9161](../../../backend/migrations/001_init.sql#L9161) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9162](../../../backend/migrations/001_init.sql#L9162) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9163](../../../backend/migrations/001_init.sql#L9163) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medicine_dose_pattern_revisions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint pattern_id FK,UK
        bigint revision UK
        text pattern_name
        text purpose
        text administration_route
        jsonb parameter_snapshot
        text source_reference
        text approval_reference
        bigint approved_by FK
        timestamptz approved_at
        boolean is_active
        bigint operation_id FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    medicine_dose_params {
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
    clinics ||..o{ medicine_dose_pattern_revisions : "clinic_id"
    staffs o|..o{ medicine_dose_pattern_revisions : "approved_by"
    medicine_dose_params ||..o{ medicine_dose_pattern_revisions : "pattern_id, clinic_id"
    operation_bindings ||..o{ medicine_dose_pattern_revisions : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medicine_dose_pattern_revisions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9150](../../../backend/migrations/001_init.sql#L9150) |
| `medicine_dose_pattern_revisions_approved_by_fkey` | `approved_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9159](../../../backend/migrations/001_init.sql#L9159) |
| `medicine_dose_pattern_revisions_pattern_id_clinic_id_fkey` | `pattern_id, clinic_id` | [medicine_dose_params](medicine_dose_params.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9167](../../../backend/migrations/001_init.sql#L9167) |
| `medicine_dose_pattern_revisions_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9168](../../../backend/migrations/001_init.sql#L9168) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medicine_dose_pattern_revisions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medicine_dose_pattern_revisions_approved_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `medicine_dose_pattern_revisions_pattern_id_clinic_id_fkey`: `FOREIGN KEY (pattern_id,clinic_id) REFERENCES medicine_dose_params(id,clinic_id) ON DELETE RESTRICT`
- `medicine_dose_pattern_revisions_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medicine_dose_pattern_revisions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9149](../../../backend/migrations/001_init.sql#L9149) |
| `medicine_dose_pattern_revisions_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:9152](../../../backend/migrations/001_init.sql#L9152) |
| `medicine_dose_pattern_revisions_parameter_snapshot_check` | check | `CHECK (jsonb_typeof(parameter_snapshot)='object')` | [001_init.sql:9156](../../../backend/migrations/001_init.sql#L9156) |
| `medicine_dose_pattern_revisions_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:9164](../../../backend/migrations/001_init.sql#L9164) |
| `medicine_dose_pattern_revisio_clinic_id_pattern_id_revision_key` | unique | `UNIQUE (clinic_id,pattern_id,revision)` | [001_init.sql:9165](../../../backend/migrations/001_init.sql#L9165) |
| `medicine_dose_pattern_revisions_check` | check | `CHECK ((approved_by IS NULL AND approved_at IS NULL) OR (approved_by IS NOT NULL AND approved_at IS NOT NULL AND approval_reference IS NOT NULL))` | [001_init.sql:9166](../../../backend/migrations/001_init.sql#L9166) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
