<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `reservation_type_groups`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:803](../../../backend/migrations/001_init.sql#L803) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:804](../../../backend/migrations/001_init.sql#L804) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:805](../../../backend/migrations/001_init.sql#L805) |
| `color` | `text` | NOT NULL | `'#3B82F6'` | — | [001_init.sql:806](../../../backend/migrations/001_init.sql#L806) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:807](../../../backend/migrations/001_init.sql#L807) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:808](../../../backend/migrations/001_init.sql#L808) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:809](../../../backend/migrations/001_init.sql#L809) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:810](../../../backend/migrations/001_init.sql#L810) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:811](../../../backend/migrations/001_init.sql#L811) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    reservation_type_groups {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        text color
        integer sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ reservation_type_groups : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `reservation_type_groups_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:804](../../../backend/migrations/001_init.sql#L804) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `reservation_type_groups_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `reservation_type_groups_pkey` | primary | `PRIMARY KEY` | [001_init.sql:803](../../../backend/migrations/001_init.sql#L803) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_rtg_clinic` | NO | — | `CREATE INDEX idx_rtg_clinic ON reservation_type_groups(clinic_id)` | [001_init.sql:814](../../../backend/migrations/001_init.sql#L814) |
| `idx_reservation_type_groups_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_reservation_type_groups_sort  ON reservation_type_groups(clinic_id, sort_order) WHERE deleted_at IS NULL` | [001_init.sql:2267](../../../backend/migrations/001_init.sql#L2267) |
| `idx_reservation_type_groups_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_reservation_type_groups_active ON reservation_type_groups(id) WHERE deleted_at IS NULL` | [001_init.sql:2475](../../../backend/migrations/001_init.sql#L2475) |
