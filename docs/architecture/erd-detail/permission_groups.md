<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `permission_groups`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1206](../../../backend/migrations/001_init.sql#L1206) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1207](../../../backend/migrations/001_init.sql#L1207) |
| `name` | `varchar(100)` | NOT NULL | — | — | [001_init.sql:1208](../../../backend/migrations/001_init.sql#L1208) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:1209](../../../backend/migrations/001_init.sql#L1209) |
| `color` | `varchar(7)` | NOT NULL | `'#6B7280'` | — | [001_init.sql:1210](../../../backend/migrations/001_init.sql#L1210) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1211](../../../backend/migrations/001_init.sql#L1211) |
| `sort_order` | `int` | NOT NULL | `0` | — | [001_init.sql:1212](../../../backend/migrations/001_init.sql#L1212) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1213](../../../backend/migrations/001_init.sql#L1213) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1214](../../../backend/migrations/001_init.sql#L1214) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1215](../../../backend/migrations/001_init.sql#L1215) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    permission_groups {
        BIGSERIAL id PK
        bigint clinic_id FK
        varchar_100 name
        text description
        varchar_7 color
        boolean is_active
        int sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ permission_groups : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `permission_groups_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1207](../../../backend/migrations/001_init.sql#L1207) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `permission_groups_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `permission_groups_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1206](../../../backend/migrations/001_init.sql#L1206) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uk_permission_groups` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX uk_permission_groups ON permission_groups(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:1218](../../../backend/migrations/001_init.sql#L1218) |
