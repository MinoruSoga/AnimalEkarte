<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `pets`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:1111](../../../backend/migrations/001_init.sql#L1111) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1112](../../../backend/migrations/001_init.sql#L1112) |
| `owner_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1113](../../../backend/migrations/001_init.sql#L1113) |
| `pet_number` | `text` | NOT NULL | `''` | — | [001_init.sql:1114](../../../backend/migrations/001_init.sql#L1114) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:1115](../../../backend/migrations/001_init.sql#L1115) |
| `name_kana` | `text` | NOT NULL | `''` | — | [001_init.sql:1116](../../../backend/migrations/001_init.sql#L1116) |
| `animal_species_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1117](../../../backend/migrations/001_init.sql#L1117) |
| `gender` | `pet_gender` | NOT NULL | `'unknown'` | — | [001_init.sql:1118](../../../backend/migrations/001_init.sql#L1118) |
| `status` | `pet_status` | NOT NULL | `'alive'` | — | [001_init.sql:1119](../../../backend/migrations/001_init.sql#L1119) |
| `birth_date` | `date` | NULL | — | — | [001_init.sql:1120](../../../backend/migrations/001_init.sql#L1120) |
| `breed` | `text` | NOT NULL | `''` | — | [001_init.sql:1121](../../../backend/migrations/001_init.sql#L1121) |
| `color` | `text` | NOT NULL | `''` | — | [001_init.sql:1122](../../../backend/migrations/001_init.sql#L1122) |
| `weight` | `numeric(6,2)` | NULL | — | — | [001_init.sql:1123](../../../backend/migrations/001_init.sql#L1123) |
| `neutered_date` | `date` | NULL | — | — | [001_init.sql:1124](../../../backend/migrations/001_init.sql#L1124) |
| `acquisition_type` | `acquisition_type` | NULL | — | — | [001_init.sql:1125](../../../backend/migrations/001_init.sql#L1125) |
| `danger_level` | `danger_level` | NOT NULL | `'low'` | — | [001_init.sql:1126](../../../backend/migrations/001_init.sql#L1126) |
| `food` | `text` | NOT NULL | `''` | — | [001_init.sql:1127](../../../backend/migrations/001_init.sql#L1127) |
| `environment` | `text` | NOT NULL | `''` | — | [001_init.sql:1128](../../../backend/migrations/001_init.sql#L1128) |
| `phone` | `text` | NOT NULL | `''` | — | [001_init.sql:1129](../../../backend/migrations/001_init.sql#L1129) |
| `last_visit` | `date` | NULL | — | — | [001_init.sql:1130](../../../backend/migrations/001_init.sql#L1130) |
| `insurance_id` | `bigint` | NULL | — | FK | [001_init.sql:1131](../../../backend/migrations/001_init.sql#L1131) |
| `remarks` | `text` | NOT NULL | `''` | — | [001_init.sql:1132](../../../backend/migrations/001_init.sql#L1132) |
| `deceased_at` | `timestamptz` | NULL | — | — | [001_init.sql:1134](../../../backend/migrations/001_init.sql#L1134) |
| `deceased_reason` | `text` | NULL | — | — | [001_init.sql:1135](../../../backend/migrations/001_init.sql#L1135) |
| `blood_type` | `text` | NULL | — | — | [001_init.sql:1137](../../../backend/migrations/001_init.sql#L1137) |
| `microchip_number` | `text` | NULL | — | — | [001_init.sql:1138](../../../backend/migrations/001_init.sql#L1138) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1139](../../../backend/migrations/001_init.sql#L1139) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1140](../../../backend/migrations/001_init.sql#L1140) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1141](../../../backend/migrations/001_init.sql#L1141) |
| `danger_reason` | `text` | NULL | — | — | [001_init.sql:4027](../../../backend/migrations/001_init.sql#L4027) |
| `version` | `INTEGER` | NOT NULL | `1` | — | [001_init.sql:4145](../../../backend/migrations/001_init.sql#L4145) |
| `name_origin` | `text` | NULL | — | — | [012_pets_name_origin_meeting_story.sql:8](../../../backend/migrations/012_pets_name_origin_meeting_story.sql#L8) |
| `meeting_story` | `text` | NULL | — | — | [012_pets_name_origin_meeting_story.sql:9](../../../backend/migrations/012_pets_name_origin_meeting_story.sql#L9) |
| `neutered_status` | `pet_neutered_status` | NOT NULL | `'unknown'` | — | [018_pet_neutered_status.sql:10](../../../backend/migrations/018_pet_neutered_status.sql#L10) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    pets {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint owner_id FK
        text pet_number
        text name
        text name_kana
        bigint animal_species_id FK
        pet_gender gender
        pet_status status
        date birth_date
        text breed
        text color
        numeric_6_2 weight
        date neutered_date
        acquisition_type acquisition_type
        danger_level danger_level
        text food
        text environment
        text phone
        date last_visit
        bigint insurance_id FK
        text remarks
        timestamptz deceased_at
        text deceased_reason
        text blood_type
        text microchip_number
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        text danger_reason
        INTEGER version
        text name_origin
        text meeting_story
        pet_neutered_status neutered_status
    }
    animal_species {
        BIGSERIAL id PK
    }
    clinics {
        BIGSERIAL id PK
    }
    insurances {
        BIGSERIAL id PK
    }
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ pets : "clinic_id"
    owners ||..o{ pets : "owner_id"
    animal_species ||..o{ pets : "animal_species_id"
    insurances o|..o{ pets : "insurance_id"
    owners ||..o{ pets : "clinic_id, owner_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `pets_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1112](../../../backend/migrations/001_init.sql#L1112) |
| `pets_owner_id_fkey` | `owner_id` | [owners](owners.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1113](../../../backend/migrations/001_init.sql#L1113) |
| `pets_animal_species_id_fkey` | `animal_species_id` | [animal_species](animal_species.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1117](../../../backend/migrations/001_init.sql#L1117) |
| `pets_insurance_id_fkey` | `insurance_id` | [insurances](insurances.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1131](../../../backend/migrations/001_init.sql#L1131) |
| `fk_pets_clinic_owner` | `clinic_id, owner_id` | [owners](owners.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:4276](../../../backend/migrations/001_init.sql#L4276) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `pets_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `pets_owner_id_fkey`: `REFERENCES owners(id) ON DELETE RESTRICT`
- `pets_animal_species_id_fkey`: `REFERENCES animal_species(id) ON DELETE RESTRICT`
- `pets_insurance_id_fkey`: `REFERENCES insurances(id) ON DELETE SET NULL`
- `fk_pets_clinic_owner`: `CONSTRAINT fk_pets_clinic_owner     FOREIGN KEY (clinic_id, owner_id)     REFERENCES owners (clinic_id, id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `pets_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1111](../../../backend/migrations/001_init.sql#L1111) |
| `pets_name_kana_check` | check | `CHECK (name_kana !~ '[ァ-ヶ]')` | [001_init.sql:1116](../../../backend/migrations/001_init.sql#L1116) |
| `uq_pets_clinic_id_id` | unique | `CONSTRAINT uq_pets_clinic_id_id     UNIQUE (clinic_id, id)` | [001_init.sql:4090](../../../backend/migrations/001_init.sql#L4090) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_pets_deceased` | NO | `deceased_at IS NOT NULL` | `CREATE INDEX idx_pets_deceased ON pets (clinic_id, deceased_at)     WHERE deceased_at IS NOT NULL` | [001_init.sql:1145](../../../backend/migrations/001_init.sql#L1145) |
| `idx_pets_clinic_id` | NO | — | `CREATE INDEX idx_pets_clinic_id ON pets(clinic_id)` | [001_init.sql:2239](../../../backend/migrations/001_init.sql#L2239) |
| `idx_pets_name_trgm` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_pets_name_trgm ON pets USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL` | [001_init.sql:2337](../../../backend/migrations/001_init.sql#L2337) |
| `idx_pets_owner_id` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_pets_owner_id   ON pets(owner_id)   WHERE deleted_at IS NULL` | [001_init.sql:2387](../../../backend/migrations/001_init.sql#L2387) |
| `idx_pets_clinic_owner_living` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_pets_clinic_owner_living   ON pets(clinic_id, owner_id, deceased_at)   WHERE deleted_at IS NULL` | [001_init.sql:2392](../../../backend/migrations/001_init.sql#L2392) |
| `idx_pets_animal_species_id` | NO | — | `CREATE INDEX idx_pets_animal_species_id ON pets(animal_species_id)` | [001_init.sql:2429](../../../backend/migrations/001_init.sql#L2429) |
| `idx_pets_insurance_id` | NO | `insurance_id IS NOT NULL` | `CREATE INDEX idx_pets_insurance_id ON pets(insurance_id) WHERE insurance_id IS NOT NULL` | [001_init.sql:2430](../../../backend/migrations/001_init.sql#L2430) |
| `idx_pets_name_fold_trgm` | NO | — | `CREATE INDEX idx_pets_name_fold_trgm   ON pets USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [015_search_fold_indexes.sql:48](../../../backend/migrations/015_search_fold_indexes.sql#L48) |
| `idx_pets_name_kana_fold_trgm` | NO | — | `CREATE INDEX idx_pets_name_kana_fold_trgm   ON pets USING gin ((translate(name_kana, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [015_search_fold_indexes.sql:51](../../../backend/migrations/015_search_fold_indexes.sql#L51) |
| `idx_pets_pet_number_trgm` | NO | — | `CREATE INDEX idx_pets_pet_number_trgm   ON pets USING gin (pet_number gin_trgm_ops)` | [015_search_fold_indexes.sql:54](../../../backend/migrations/015_search_fold_indexes.sql#L54) |
