<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `common_pet_basic_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10774](../../../backend/migrations/001_init.sql#L10774) |
| `common_pet_id` | `uuid` | NOT NULL | — | FK, UK | [001_init.sql:10774](../../../backend/migrations/001_init.sql#L10774) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10775](../../../backend/migrations/001_init.sql#L10775) |
| `source_clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10775](../../../backend/migrations/001_init.sql#L10775) |
| `source_local_pet_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10776](../../../backend/migrations/001_init.sql#L10776) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10776](../../../backend/migrations/001_init.sql#L10776) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:10777](../../../backend/migrations/001_init.sql#L10777) |
| `name_kana` | `text` | NULL | — | — | [001_init.sql:10777](../../../backend/migrations/001_init.sql#L10777) |
| `animal_species_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10777](../../../backend/migrations/001_init.sql#L10777) |
| `breed` | `text` | NULL | — | — | [001_init.sql:10778](../../../backend/migrations/001_init.sql#L10778) |
| `gender` | `pet_gender` | NULL | — | — | [001_init.sql:10778](../../../backend/migrations/001_init.sql#L10778) |
| `status` | `pet_status` | NULL | — | — | [001_init.sql:10778](../../../backend/migrations/001_init.sql#L10778) |
| `birth_date` | `date` | NULL | — | — | [001_init.sql:10778](../../../backend/migrations/001_init.sql#L10778) |
| `color` | `text` | NULL | — | — | [001_init.sql:10778](../../../backend/migrations/001_init.sql#L10778) |
| `registration_weight` | `numeric` | NULL | — | — | [001_init.sql:10779](../../../backend/migrations/001_init.sql#L10779) |
| `registration_weight_unit` | `text` | NULL | — | — | [001_init.sql:10779](../../../backend/migrations/001_init.sql#L10779) |
| `neutered_status` | `pet_neutered_status` | NOT NULL | `'unknown'` | — | [001_init.sql:10779](../../../backend/migrations/001_init.sql#L10779) |
| `neutered_date` | `date` | NULL | — | — | [001_init.sql:10779](../../../backend/migrations/001_init.sql#L10779) |
| `acquisition_type` | `acquisition_type` | NULL | — | — | [001_init.sql:10780](../../../backend/migrations/001_init.sql#L10780) |
| `danger_level` | `danger_level` | NULL | — | — | [001_init.sql:10780](../../../backend/migrations/001_init.sql#L10780) |
| `danger_reason` | `text` | NULL | — | — | [001_init.sql:10780](../../../backend/migrations/001_init.sql#L10780) |
| `food` | `text` | NULL | — | — | [001_init.sql:10780](../../../backend/migrations/001_init.sql#L10780) |
| `environment` | `text` | NULL | — | — | [001_init.sql:10780](../../../backend/migrations/001_init.sql#L10780) |
| `phone` | `text` | NULL | — | — | [001_init.sql:10780](../../../backend/migrations/001_init.sql#L10780) |
| `remarks` | `text` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `deceased_at` | `timestamptz` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `deceased_reason` | `text` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `blood_type` | `text` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `microchip_number` | `text` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `name_origin` | `text` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `meeting_story` | `text` | NULL | — | — | [001_init.sql:10781](../../../backend/migrations/001_init.sql#L10781) |
| `provenance` | `jsonb` | NOT NULL | — | — | [001_init.sql:10782](../../../backend/migrations/001_init.sql#L10782) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:10782](../../../backend/migrations/001_init.sql#L10782) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10783](../../../backend/migrations/001_init.sql#L10783) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    common_pet_basic_revisions {
        BIGSERIAL id PK,UK
        uuid common_pet_id FK,UK
        bigint revision UK
        bigint source_clinic_id FK
        bigint source_local_pet_id FK
        bigint operation_id FK
        text name
        text name_kana
        bigint animal_species_id FK
        text breed
        pet_gender gender
        pet_status status
        date birth_date
        text color
        numeric registration_weight
        text registration_weight_unit
        pet_neutered_status neutered_status
        date neutered_date
        acquisition_type acquisition_type
        danger_level danger_level
        text danger_reason
        text food
        text environment
        text phone
        text remarks
        timestamptz deceased_at
        text deceased_reason
        text blood_type
        text microchip_number
        text name_origin
        text meeting_story
        jsonb provenance
        bigint recorded_by FK
        timestamptz recorded_at
    }
    animal_species {
        BIGSERIAL id PK
    }
    clinics {
        BIGSERIAL id PK
    }
    common_pet_subjects {
        uuid id PK,FK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    pets {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    common_pet_subjects ||..o{ common_pet_basic_revisions : "common_pet_id"
    clinics ||..o{ common_pet_basic_revisions : "source_clinic_id"
    animal_species ||..o{ common_pet_basic_revisions : "animal_species_id"
    staffs ||..o{ common_pet_basic_revisions : "recorded_by"
    pets ||..o{ common_pet_basic_revisions : "source_local_pet_id, source_clinic_id"
    operation_bindings ||..o{ common_pet_basic_revisions : "operation_id, source_clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `common_pet_basic_revisions_common_pet_id_fkey` | `common_pet_id` | [common_pet_subjects](common_pet_subjects.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10774](../../../backend/migrations/001_init.sql#L10774) |
| `common_pet_basic_revisions_source_clinic_id_fkey` | `source_clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10775](../../../backend/migrations/001_init.sql#L10775) |
| `common_pet_basic_revisions_animal_species_id_fkey` | `animal_species_id` | [animal_species](animal_species.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10777](../../../backend/migrations/001_init.sql#L10777) |
| `common_pet_basic_revisions_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10782](../../../backend/migrations/001_init.sql#L10782) |
| `common_pet_basic_revisions_source_local_pet_id_source_clin_fkey` | `source_local_pet_id, source_clinic_id` | [pets](pets.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10784](../../../backend/migrations/001_init.sql#L10784) |
| `common_pet_basic_revisions_operation_id_source_clinic_id_fkey` | `operation_id, source_clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10785](../../../backend/migrations/001_init.sql#L10785) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `common_pet_basic_revisions_common_pet_id_fkey`: `REFERENCES common_pet_subjects(id) ON DELETE RESTRICT`
- `common_pet_basic_revisions_source_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `common_pet_basic_revisions_animal_species_id_fkey`: `REFERENCES animal_species(id) ON DELETE RESTRICT`
- `common_pet_basic_revisions_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `common_pet_basic_revisions_source_local_pet_id_source_clin_fkey`: `FOREIGN KEY(source_local_pet_id,source_clinic_id) REFERENCES pets(id,clinic_id) ON DELETE RESTRICT`
- `common_pet_basic_revisions_operation_id_source_clinic_id_fkey`: `FOREIGN KEY(operation_id,source_clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `common_pet_basic_revisions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10774](../../../backend/migrations/001_init.sql#L10774) |
| `common_pet_basic_revisions_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10775](../../../backend/migrations/001_init.sql#L10775) |
| `common_pet_basic_revisions_name_check` | check | `CHECK(btrim(name)<>'')` | [001_init.sql:10777](../../../backend/migrations/001_init.sql#L10777) |
| `common_pet_basic_revisions_name_kana_check` | check | `CHECK(name_kana IS NULL OR name_kana !~ '[ァ-ヶ]')` | [001_init.sql:10777](../../../backend/migrations/001_init.sql#L10777) |
| `common_pet_basic_revisions_registration_weight_check` | check | `CHECK(registration_weight IS NULL OR registration_weight>0)` | [001_init.sql:10779](../../../backend/migrations/001_init.sql#L10779) |
| `common_pet_basic_revisions_provenance_check` | check | `CHECK(jsonb_typeof(provenance)='object')` | [001_init.sql:10782](../../../backend/migrations/001_init.sql#L10782) |
| `common_pet_basic_revisions_id_common_pet_id_key` | unique | `UNIQUE(id,common_pet_id)` | [001_init.sql:10783](../../../backend/migrations/001_init.sql#L10783) |
| `common_pet_basic_revisions_common_pet_id_revision_key` | unique | `UNIQUE(common_pet_id,revision)` | [001_init.sql:10783](../../../backend/migrations/001_init.sql#L10783) |
| `ck_common_pet_danger_reason` | check | `CONSTRAINT ck_common_pet_danger_reason CHECK(danger_level IS DISTINCT FROM 'high' OR(danger_reason IS NOT NULL AND btrim(danger_reason)<>''))` | [001_init.sql:11204](../../../backend/migrations/001_init.sql#L11204) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
