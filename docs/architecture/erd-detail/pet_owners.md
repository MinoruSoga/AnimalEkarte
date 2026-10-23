<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `pet_owners`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:4122](../../../backend/migrations/001_init.sql#L4122) |
| `clinic_id` | `BIGINT` | NOT NULL | — | FK | [001_init.sql:4123](../../../backend/migrations/001_init.sql#L4123) |
| `pet_id` | `BIGINT` | NOT NULL | — | FK, UK | [001_init.sql:4124](../../../backend/migrations/001_init.sql#L4124) |
| `owner_id` | `BIGINT` | NOT NULL | — | FK, UK | [001_init.sql:4125](../../../backend/migrations/001_init.sql#L4125) |
| `relationship` | `TEXT` | NOT NULL | `''` | — | [001_init.sql:4126](../../../backend/migrations/001_init.sql#L4126) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `now()` | — | [001_init.sql:4127](../../../backend/migrations/001_init.sql#L4127) |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `now()` | — | [001_init.sql:4128](../../../backend/migrations/001_init.sql#L4128) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    pet_owners {
        BIGSERIAL id PK
        BIGINT clinic_id FK
        BIGINT pet_id FK,UK
        BIGINT owner_id FK,UK
        TEXT relationship
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    pets {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ pet_owners : "clinic_id"
    pets ||..o{ pet_owners : "clinic_id, pet_id"
    owners ||..o{ pet_owners : "clinic_id, owner_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `pet_owners_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:4123](../../../backend/migrations/001_init.sql#L4123) |
| `pet_owners_clinic_id_pet_id_fkey` | `clinic_id, pet_id` | [pets](pets.md) | `clinic_id, id` | `ON DELETE NO ACTION` | [001_init.sql:4130](../../../backend/migrations/001_init.sql#L4130) |
| `pet_owners_clinic_id_owner_id_fkey` | `clinic_id, owner_id` | [owners](owners.md) | `clinic_id, id` | `ON DELETE NO ACTION` | [001_init.sql:4131](../../../backend/migrations/001_init.sql#L4131) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `pet_owners_clinic_id_fkey`: `REFERENCES clinics(id)`
- `pet_owners_clinic_id_pet_id_fkey`: `FOREIGN KEY (clinic_id, pet_id) REFERENCES pets (clinic_id, id)`
- `pet_owners_clinic_id_owner_id_fkey`: `FOREIGN KEY (clinic_id, owner_id) REFERENCES owners (clinic_id, id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `pet_owners_pkey` | primary | `PRIMARY KEY` | [001_init.sql:4122](../../../backend/migrations/001_init.sql#L4122) |
| `pet_owners_pet_id_owner_id_key` | unique | `UNIQUE (pet_id, owner_id)` | [001_init.sql:4129](../../../backend/migrations/001_init.sql#L4129) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_pet_owners_clinic_pet` | NO | — | `CREATE INDEX idx_pet_owners_clinic_pet ON pet_owners (clinic_id, pet_id)` | [001_init.sql:4134](../../../backend/migrations/001_init.sql#L4134) |
| `idx_pet_owners_clinic_owner` | NO | — | `CREATE INDEX idx_pet_owners_clinic_owner ON pet_owners (clinic_id, owner_id)` | [001_init.sql:4135](../../../backend/migrations/001_init.sql#L4135) |
