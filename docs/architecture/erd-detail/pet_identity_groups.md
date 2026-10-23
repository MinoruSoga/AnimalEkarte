<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `pet_identity_groups`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:4670](../../../backend/migrations/001_init.sql#L4670) |
| `created_clinic_id` | `BIGINT` | NOT NULL | — | FK, UK | [001_init.sql:4671](../../../backend/migrations/001_init.sql#L4671) |
| `owner_group_created_clinic_id` | `BIGINT` | NOT NULL | — | FK | [001_init.sql:4672](../../../backend/migrations/001_init.sql#L4672) |
| `owner_group_id` | `BIGINT` | NOT NULL | — | FK | [001_init.sql:4673](../../../backend/migrations/001_init.sql#L4673) |
| `version` | `BIGINT` | NOT NULL | `1` | — | [001_init.sql:4674](../../../backend/migrations/001_init.sql#L4674) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `now()` | — | [001_init.sql:4675](../../../backend/migrations/001_init.sql#L4675) |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `now()` | — | [001_init.sql:4676](../../../backend/migrations/001_init.sql#L4676) |
| `deleted_at` | `TIMESTAMPTZ` | NULL | — | — | [001_init.sql:4677](../../../backend/migrations/001_init.sql#L4677) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    pet_identity_groups {
        BIGSERIAL id PK,UK
        BIGINT created_clinic_id FK,UK
        BIGINT owner_group_created_clinic_id FK
        BIGINT owner_group_id FK
        BIGINT version
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    owner_identity_groups {
        BIGSERIAL id PK,UK
        BIGINT created_clinic_id FK,UK
    }
    clinics ||..o{ pet_identity_groups : "created_clinic_id"
    owner_identity_groups ||..o{ pet_identity_groups : "owner_group_created_clinic_id, owner_group_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `pet_identity_groups_created_clinic_id_fkey` | `created_clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:4671](../../../backend/migrations/001_init.sql#L4671) |
| `pet_identity_groups_owner_group_created_clinic_id_owner_gr_fkey` | `owner_group_created_clinic_id, owner_group_id` | [owner_identity_groups](owner_identity_groups.md) | `created_clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:4679](../../../backend/migrations/001_init.sql#L4679) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `pet_identity_groups_created_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `pet_identity_groups_owner_group_created_clinic_id_owner_gr_fkey`: `FOREIGN KEY (owner_group_created_clinic_id, owner_group_id)         REFERENCES owner_identity_groups(created_clinic_id, id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `pet_identity_groups_pkey` | primary | `PRIMARY KEY` | [001_init.sql:4670](../../../backend/migrations/001_init.sql#L4670) |
| `pet_identity_groups_version_check` | check | `CHECK (version >= 1)` | [001_init.sql:4674](../../../backend/migrations/001_init.sql#L4674) |
| `pet_identity_groups_created_clinic_id_id_key` | unique | `UNIQUE (created_clinic_id, id)` | [001_init.sql:4678](../../../backend/migrations/001_init.sql#L4678) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
