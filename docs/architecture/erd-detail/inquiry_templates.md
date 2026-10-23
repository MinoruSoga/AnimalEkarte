<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `inquiry_templates`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1111](../../../backend/migrations/001_init.sql#L1111) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1112](../../../backend/migrations/001_init.sql#L1112) |
| `category` | `text` | NOT NULL | `''` | — | [001_init.sql:1113](../../../backend/migrations/001_init.sql#L1113) |
| `title` | `text` | NOT NULL | — | — | [001_init.sql:1114](../../../backend/migrations/001_init.sql#L1114) |
| `content` | `text` | NOT NULL | `''` | — | [001_init.sql:1115](../../../backend/migrations/001_init.sql#L1115) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1116](../../../backend/migrations/001_init.sql#L1116) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1117](../../../backend/migrations/001_init.sql#L1117) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1118](../../../backend/migrations/001_init.sql#L1118) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1119](../../../backend/migrations/001_init.sql#L1119) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1120](../../../backend/migrations/001_init.sql#L1120) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    inquiry_templates {
        BIGSERIAL id PK
        bigint clinic_id FK
        text category
        text title
        text content
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ inquiry_templates : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `inquiry_templates_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1112](../../../backend/migrations/001_init.sql#L1112) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `inquiry_templates_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `inquiry_templates_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1111](../../../backend/migrations/001_init.sql#L1111) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_inquiry_templates_clinic_id` | NO | — | `CREATE INDEX idx_inquiry_templates_clinic_id ON inquiry_templates(clinic_id)` | [001_init.sql:2254](../../../backend/migrations/001_init.sql#L2254) |
| `idx_inquiry_templates_clinic_category` | NO | — | `CREATE INDEX idx_inquiry_templates_clinic_category ON inquiry_templates(clinic_id, category)` | [001_init.sql:2255](../../../backend/migrations/001_init.sql#L2255) |
| `idx_inquiry_templates_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_inquiry_templates_active ON inquiry_templates(id) WHERE deleted_at IS NULL` | [001_init.sql:2503](../../../backend/migrations/001_init.sql#L2503) |
