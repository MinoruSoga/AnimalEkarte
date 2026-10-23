<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `examination_field_semantic_mappings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10142](../../../backend/migrations/001_init.sql#L10142) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10142](../../../backend/migrations/001_init.sql#L10142) |
| `exam_type_field_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10143](../../../backend/migrations/001_init.sql#L10143) |
| `animal_species_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10144](../../../backend/migrations/001_init.sql#L10144) |
| `meaning` | `text` | NOT NULL | — | — | [001_init.sql:10145](../../../backend/migrations/001_init.sql#L10145) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10145](../../../backend/migrations/001_init.sql#L10145) |
| `verified_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:10146](../../../backend/migrations/001_init.sql#L10146) |
| `verified_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:10146](../../../backend/migrations/001_init.sql#L10146) |
| `evidence` | `text` | NOT NULL | — | — | [001_init.sql:10147](../../../backend/migrations/001_init.sql#L10147) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10147](../../../backend/migrations/001_init.sql#L10147) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    examination_field_semantic_mappings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint exam_type_field_id FK,UK
        bigint animal_species_id FK,UK
        text meaning
        bigint revision UK
        bigint verified_by FK
        timestamptz verified_at
        text evidence
        bigint operation_id FK
    }
    animal_species {
        BIGSERIAL id PK
    }
    clinics {
        BIGSERIAL id PK
    }
    exam_type_fields {
        BIGSERIAL id PK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ examination_field_semantic_mappings : "clinic_id"
    exam_type_fields ||..o{ examination_field_semantic_mappings : "exam_type_field_id"
    animal_species ||..o{ examination_field_semantic_mappings : "animal_species_id"
    staffs ||..o{ examination_field_semantic_mappings : "verified_by"
    operation_bindings ||..o{ examination_field_semantic_mappings : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `examination_field_semantic_mappings_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10142](../../../backend/migrations/001_init.sql#L10142) |
| `examination_field_semantic_mappings_exam_type_field_id_fkey` | `exam_type_field_id` | [exam_type_fields](exam_type_fields.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10143](../../../backend/migrations/001_init.sql#L10143) |
| `examination_field_semantic_mappings_animal_species_id_fkey` | `animal_species_id` | [animal_species](animal_species.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10144](../../../backend/migrations/001_init.sql#L10144) |
| `examination_field_semantic_mappings_verified_by_fkey` | `verified_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10146](../../../backend/migrations/001_init.sql#L10146) |
| `examination_field_semantic_mappings_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10149](../../../backend/migrations/001_init.sql#L10149) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `examination_field_semantic_mappings_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `examination_field_semantic_mappings_exam_type_field_id_fkey`: `REFERENCES exam_type_fields(id) ON DELETE RESTRICT`
- `examination_field_semantic_mappings_animal_species_id_fkey`: `REFERENCES animal_species(id) ON DELETE RESTRICT`
- `examination_field_semantic_mappings_verified_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `examination_field_semantic_mappings_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `examination_field_semantic_mappings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10142](../../../backend/migrations/001_init.sql#L10142) |
| `examination_field_semantic_mappings_meaning_check` | check | `CHECK(meaning IN('FIV','FeLV','mf','HW'))` | [001_init.sql:10145](../../../backend/migrations/001_init.sql#L10145) |
| `examination_field_semantic_mappings_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10145](../../../backend/migrations/001_init.sql#L10145) |
| `examination_field_semantic_mappings_evidence_check` | check | `CHECK(btrim(evidence)<>'')` | [001_init.sql:10147](../../../backend/migrations/001_init.sql#L10147) |
| `examination_field_semantic_mappings_id_clinic_id_key` | unique | `UNIQUE(id,clinic_id)` | [001_init.sql:10148](../../../backend/migrations/001_init.sql#L10148) |
| `examination_field_semantic_ma_clinic_id_exam_type_field_id__key` | unique | `UNIQUE(clinic_id,exam_type_field_id,animal_species_id,revision)` | [001_init.sql:10148](../../../backend/migrations/001_init.sql#L10148) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
