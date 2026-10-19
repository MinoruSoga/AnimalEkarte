<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `shift_templates`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2708](../../../backend/migrations/001_init.sql#L2708) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2709](../../../backend/migrations/001_init.sql#L2709) |
| `name` | `varchar(100)` | NOT NULL | — | — | [001_init.sql:2710](../../../backend/migrations/001_init.sql#L2710) |
| `shift_type` | `shift_type` | NOT NULL | `'full'` | — | [001_init.sql:2711](../../../backend/migrations/001_init.sql#L2711) |
| `start_time` | `time` | NULL | — | — | [001_init.sql:2712](../../../backend/migrations/001_init.sql#L2712) |
| `end_time` | `time` | NULL | — | — | [001_init.sql:2713](../../../backend/migrations/001_init.sql#L2713) |
| `notes` | `text` | NOT NULL | `''` | — | [001_init.sql:2714](../../../backend/migrations/001_init.sql#L2714) |
| `sort_order` | `integer` | NOT NULL | `0` | — | [001_init.sql:2715](../../../backend/migrations/001_init.sql#L2715) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:2716](../../../backend/migrations/001_init.sql#L2716) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2717](../../../backend/migrations/001_init.sql#L2717) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2718](../../../backend/migrations/001_init.sql#L2718) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:2719](../../../backend/migrations/001_init.sql#L2719) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    shift_templates {
        BIGSERIAL id PK
        bigint clinic_id FK
        varchar_100 name
        shift_type shift_type
        time start_time
        time end_time
        text notes
        integer sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ shift_templates : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `shift_templates_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:2709](../../../backend/migrations/001_init.sql#L2709) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `shift_templates_clinic_id_fkey`: `REFERENCES clinics(id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `shift_templates_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2708](../../../backend/migrations/001_init.sql#L2708) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uk_shift_templates_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX uk_shift_templates_clinic_name ON shift_templates(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2723](../../../backend/migrations/001_init.sql#L2723) |
