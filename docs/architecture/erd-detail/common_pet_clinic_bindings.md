<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `common_pet_clinic_bindings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10748](../../../backend/migrations/001_init.sql#L10748) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10748](../../../backend/migrations/001_init.sql#L10748) |
| `common_pet_id` | `uuid` | NOT NULL | — | FK, UK | [001_init.sql:10749](../../../backend/migrations/001_init.sql#L10749) |
| `local_pet_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10749](../../../backend/migrations/001_init.sql#L10749) |
| `confirmed_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:10751](../../../backend/migrations/001_init.sql#L10751) |
| `confirmed_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:10751](../../../backend/migrations/001_init.sql#L10751) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10752](../../../backend/migrations/001_init.sql#L10752) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    common_pet_clinic_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        uuid common_pet_id FK,UK
        bigint local_pet_id FK,UK
        bigint confirmed_by FK
        timestamptz confirmed_at
        bigint operation_id FK
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
    clinics ||..o{ common_pet_clinic_bindings : "clinic_id"
    staffs ||..o{ common_pet_clinic_bindings : "confirmed_by"
    common_pet_subjects ||..o{ common_pet_clinic_bindings : "common_pet_id"
    pets ||..o| common_pet_clinic_bindings : "local_pet_id, clinic_id"
    operation_bindings ||..o{ common_pet_clinic_bindings : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `common_pet_clinic_bindings_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10748](../../../backend/migrations/001_init.sql#L10748) |
| `common_pet_clinic_bindings_confirmed_by_fkey` | `confirmed_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10751](../../../backend/migrations/001_init.sql#L10751) |
| `common_pet_clinic_bindings_common_pet_id_fkey` | `common_pet_id` | [common_pet_subjects](common_pet_subjects.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10753](../../../backend/migrations/001_init.sql#L10753) |
| `common_pet_clinic_bindings_local_pet_id_clinic_id_fkey` | `local_pet_id, clinic_id` | [pets](pets.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10754](../../../backend/migrations/001_init.sql#L10754) |
| `common_pet_clinic_bindings_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10755](../../../backend/migrations/001_init.sql#L10755) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `common_pet_clinic_bindings_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `common_pet_clinic_bindings_confirmed_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `common_pet_clinic_bindings_common_pet_id_fkey`: `FOREIGN KEY(common_pet_id) REFERENCES common_pet_subjects(id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED`
- `common_pet_clinic_bindings_local_pet_id_clinic_id_fkey`: `FOREIGN KEY(local_pet_id,clinic_id) REFERENCES pets(id,clinic_id) ON DELETE RESTRICT`
- `common_pet_clinic_bindings_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `common_pet_clinic_bindings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10748](../../../backend/migrations/001_init.sql#L10748) |
| `common_pet_clinic_bindings_id_clinic_id_key` | unique | `UNIQUE(id,clinic_id)` | [001_init.sql:10752](../../../backend/migrations/001_init.sql#L10752) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uq_common_pet_clinic` | YES | — | `CREATE UNIQUE INDEX uq_common_pet_clinic ON common_pet_clinic_bindings(clinic_id,common_pet_id)` | [001_init.sql:10911](../../../backend/migrations/001_init.sql#L10911) |
| `uq_common_pet_local` | YES | — | `CREATE UNIQUE INDEX uq_common_pet_local ON common_pet_clinic_bindings(clinic_id,local_pet_id)` | [001_init.sql:10912](../../../backend/migrations/001_init.sql#L10912) |
