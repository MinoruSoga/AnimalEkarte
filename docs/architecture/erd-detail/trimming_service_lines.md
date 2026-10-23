<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `trimming_service_lines`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:8102](../../../backend/migrations/001_init.sql#L8102) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8103](../../../backend/migrations/001_init.sql#L8103) |
| `detail_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8104](../../../backend/migrations/001_init.sql#L8104) |
| `course_id` | `bigint` | NULL | — | FK | [001_init.sql:8105](../../../backend/migrations/001_init.sql#L8105) |
| `option_id` | `bigint` | NULL | — | FK | [001_init.sql:8106](../../../backend/migrations/001_init.sql#L8106) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8107](../../../backend/migrations/001_init.sql#L8107) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    trimming_service_lines {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint detail_id FK,UK
        bigint course_id FK
        bigint option_id FK
        timestamptz created_at
    }
    appointment_trimming_details {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    trimming_courses {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    trimming_options {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ trimming_service_lines : "clinic_id"
    appointment_trimming_details ||..o{ trimming_service_lines : "detail_id, clinic_id"
    trimming_courses o|..o{ trimming_service_lines : "course_id, clinic_id"
    trimming_options o|..o{ trimming_service_lines : "option_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `trimming_service_lines_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8103](../../../backend/migrations/001_init.sql#L8103) |
| `trimming_service_lines_detail_id_clinic_id_fkey` | `detail_id, clinic_id` | [appointment_trimming_details](appointment_trimming_details.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8111](../../../backend/migrations/001_init.sql#L8111) |
| `trimming_service_lines_course_id_clinic_id_fkey` | `course_id, clinic_id` | [trimming_courses](trimming_courses.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8112](../../../backend/migrations/001_init.sql#L8112) |
| `trimming_service_lines_option_id_clinic_id_fkey` | `option_id, clinic_id` | [trimming_options](trimming_options.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8113](../../../backend/migrations/001_init.sql#L8113) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `trimming_service_lines_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `trimming_service_lines_detail_id_clinic_id_fkey`: `FOREIGN KEY (detail_id,clinic_id) REFERENCES appointment_trimming_details(id,clinic_id) ON DELETE RESTRICT`
- `trimming_service_lines_course_id_clinic_id_fkey`: `FOREIGN KEY (course_id,clinic_id) REFERENCES trimming_courses(id,clinic_id) ON DELETE RESTRICT`
- `trimming_service_lines_option_id_clinic_id_fkey`: `FOREIGN KEY (option_id,clinic_id) REFERENCES trimming_options(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `trimming_service_lines_pkey` | primary | `PRIMARY KEY` | [001_init.sql:8102](../../../backend/migrations/001_init.sql#L8102) |
| `trimming_service_lines_id_clinic_id_key` | unique | `UNIQUE (id, clinic_id)` | [001_init.sql:8108](../../../backend/migrations/001_init.sql#L8108) |
| `trimming_service_lines_id_clinic_id_detail_id_key` | unique | `UNIQUE (id, clinic_id, detail_id)` | [001_init.sql:8109](../../../backend/migrations/001_init.sql#L8109) |
| `trimming_service_lines_check` | check | `CHECK (num_nonnulls(course_id,option_id)=1)` | [001_init.sql:8110](../../../backend/migrations/001_init.sql#L8110) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uq_trimming_source_course` | YES | `course_id IS NOT NULL` | `CREATE UNIQUE INDEX uq_trimming_source_course ON trimming_service_lines(clinic_id,detail_id,course_id) WHERE course_id IS NOT NULL` | [001_init.sql:8115](../../../backend/migrations/001_init.sql#L8115) |
| `uq_trimming_source_option` | YES | `option_id IS NOT NULL` | `CREATE UNIQUE INDEX uq_trimming_source_option ON trimming_service_lines(clinic_id,detail_id,option_id) WHERE option_id IS NOT NULL` | [001_init.sql:8116](../../../backend/migrations/001_init.sql#L8116) |
