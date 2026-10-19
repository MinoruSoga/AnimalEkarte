<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `trimming_courses`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:995](../../../backend/migrations/001_init.sql#L995) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:996](../../../backend/migrations/001_init.sql#L996) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:997](../../../backend/migrations/001_init.sql#L997) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:998](../../../backend/migrations/001_init.sql#L998) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:999](../../../backend/migrations/001_init.sql#L999) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:1000](../../../backend/migrations/001_init.sql#L1000) |
| `target_size` | `target_size` | NULL | — | — | [001_init.sql:1001](../../../backend/migrations/001_init.sql#L1001) |
| `duration` | `integer` | NULL | — | — | [001_init.sql:1002](../../../backend/migrations/001_init.sql#L1002) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1003](../../../backend/migrations/001_init.sql#L1003) |
| `course_type_id` | `bigint` | NULL | — | FK | [001_init.sql:1004](../../../backend/migrations/001_init.sql#L1004) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1005](../../../backend/migrations/001_init.sql#L1005) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1006](../../../backend/migrations/001_init.sql#L1006) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1007](../../../backend/migrations/001_init.sql#L1007) |
| `short_name` | `text` | NULL | — | — | [001_init.sql:7914](../../../backend/migrations/001_init.sql#L7914) |
| `image_object_key` | `text` | NULL | — | — | [001_init.sql:7915](../../../backend/migrations/001_init.sql#L7915) |
| `allowed_weekdays` | `smallint[]` | NULL | — | — | [001_init.sql:7916](../../../backend/migrations/001_init.sql#L7916) |
| `contract_version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:7917](../../../backend/migrations/001_init.sql#L7917) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    trimming_courses {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        bigint price
        boolean is_active
        text description
        target_size target_size
        integer duration
        integer sort_order
        bigint course_type_id FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        text short_name
        text image_object_key
        smallint[] allowed_weekdays
        bigint contract_version
    }
    clinics {
        BIGSERIAL id PK
    }
    trimming_course_types {
        BIGSERIAL id PK
    }
    clinics ||..o{ trimming_courses : "clinic_id"
    trimming_course_types o|..o{ trimming_courses : "course_type_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `trimming_courses_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:996](../../../backend/migrations/001_init.sql#L996) |
| `trimming_courses_course_type_id_fkey` | `course_type_id` | [trimming_course_types](trimming_course_types.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1004](../../../backend/migrations/001_init.sql#L1004) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `trimming_courses_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `trimming_courses_course_type_id_fkey`: `REFERENCES trimming_course_types(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `trimming_courses_pkey` | primary | `PRIMARY KEY` | [001_init.sql:995](../../../backend/migrations/001_init.sql#L995) |
| `trimming_courses_contract_version_check` | check | `CHECK (contract_version > 0)` | [001_init.sql:7917](../../../backend/migrations/001_init.sql#L7917) |
| `ck_trimming_course_short_name` | check | `CONSTRAINT ck_trimming_course_short_name CHECK (short_name IS NULL OR btrim(short_name) <> '')` | [001_init.sql:7918](../../../backend/migrations/001_init.sql#L7918) |
| `ck_trimming_course_image_key` | check | `CONSTRAINT ck_trimming_course_image_key CHECK (image_object_key IS NULL OR btrim(image_object_key) <> '')` | [001_init.sql:7919](../../../backend/migrations/001_init.sql#L7919) |
| `ck_trimming_course_weekdays` | check | `CONSTRAINT ck_trimming_course_weekdays CHECK (app_private.valid_trimming_weekdays(allowed_weekdays))` | [001_init.sql:7920](../../../backend/migrations/001_init.sql#L7920) |
| `uq_trimming_courses_id_clinic` | unique | `CONSTRAINT uq_trimming_courses_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:7921](../../../backend/migrations/001_init.sql#L7921) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_trimming_courses_course_type_id` | NO | `course_type_id IS NOT NULL` | `CREATE INDEX idx_trimming_courses_course_type_id ON trimming_courses(course_type_id) WHERE course_type_id IS NOT NULL` | [001_init.sql:1010](../../../backend/migrations/001_init.sql#L1010) |
| `idx_trimming_courses_clinic_id` | NO | — | `CREATE INDEX idx_trimming_courses_clinic_id ON trimming_courses(clinic_id)` | [001_init.sql:2234](../../../backend/migrations/001_init.sql#L2234) |
| `idx_trimming_courses_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_trimming_courses_clinic_sort  ON trimming_courses(clinic_id, sort_order)        WHERE deleted_at IS NULL` | [001_init.sql:2270](../../../backend/migrations/001_init.sql#L2270) |
| `idx_trimming_courses_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_trimming_courses_clinic_name ON trimming_courses(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2466](../../../backend/migrations/001_init.sql#L2466) |
| `idx_trimming_courses_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_trimming_courses_active ON trimming_courses(id) WHERE deleted_at IS NULL` | [001_init.sql:2486](../../../backend/migrations/001_init.sql#L2486) |
