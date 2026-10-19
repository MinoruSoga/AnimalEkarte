<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `line_customers`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:1257](../../../backend/migrations/001_init.sql#L1257) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1258](../../../backend/migrations/001_init.sql#L1258) |
| `line_user_id` | `text` | NOT NULL | — | UK | [001_init.sql:1259](../../../backend/migrations/001_init.sql#L1259) |
| `display_name` | `text` | NOT NULL | `''` | — | [001_init.sql:1260](../../../backend/migrations/001_init.sql#L1260) |
| `real_name` | `text` | NOT NULL | `''` | — | [001_init.sql:1261](../../../backend/migrations/001_init.sql#L1261) |
| `additional_fields` | `jsonb` | NOT NULL | `'{}'` | — | [001_init.sql:1262](../../../backend/migrations/001_init.sql#L1262) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1264](../../../backend/migrations/001_init.sql#L1264) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1265](../../../backend/migrations/001_init.sql#L1265) |
| `followed_at` | `timestamptz` | NULL | — | — | [001_init.sql:10889](../../../backend/migrations/001_init.sql#L10889) |
| `blocked_at` | `timestamptz` | NULL | — | — | [001_init.sql:10889](../../../backend/migrations/001_init.sql#L10889) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    line_customers {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text line_user_id UK
        text display_name
        text real_name
        jsonb additional_fields
        timestamptz created_at
        timestamptz updated_at
        timestamptz followed_at
        timestamptz blocked_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ line_customers : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `line_customers_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:1258](../../../backend/migrations/001_init.sql#L1258) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `line_customers_clinic_id_fkey`: `REFERENCES clinics(id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `line_customers_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1257](../../../backend/migrations/001_init.sql#L1257) |
| `line_customers_clinic_id_line_user_id_key` | unique | `UNIQUE(clinic_id, line_user_id)` | [001_init.sql:1266](../../../backend/migrations/001_init.sql#L1266) |
| `uq_line_customers_id_clinic` | unique | `CONSTRAINT uq_line_customers_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:1272](../../../backend/migrations/001_init.sql#L1272) |
| `uq_line_customer_id_clinic` | unique | `CONSTRAINT uq_line_customer_id_clinic UNIQUE(id,clinic_id)` | [001_init.sql:10890](../../../backend/migrations/001_init.sql#L10890) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
