<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `inventory_quantity_bases`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9734](../../../backend/migrations/001_init.sql#L9734) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9735](../../../backend/migrations/001_init.sql#L9735) |
| `inventory_item_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9736](../../../backend/migrations/001_init.sql#L9736) |
| `expected_item_version` | `bigint` | NOT NULL | — | — | [001_init.sql:9737](../../../backend/migrations/001_init.sql#L9737) |
| `confirmed_quantity` | `bigint` | NOT NULL | — | — | [001_init.sql:9738](../../../backend/migrations/001_init.sql#L9738) |
| `management_unit` | `text` | NOT NULL | — | — | [001_init.sql:9739](../../../backend/migrations/001_init.sql#L9739) |
| `legacy_reference` | `text` | NULL | — | — | [001_init.sql:9740](../../../backend/migrations/001_init.sql#L9740) |
| `confirmed_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9741](../../../backend/migrations/001_init.sql#L9741) |
| `confirmed_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:9742](../../../backend/migrations/001_init.sql#L9742) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9743](../../../backend/migrations/001_init.sql#L9743) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    inventory_quantity_bases {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint inventory_item_id FK,UK
        bigint expected_item_version
        bigint confirmed_quantity
        text management_unit
        text legacy_reference
        bigint confirmed_by FK
        timestamptz confirmed_at
        bigint operation_id FK
    }
    clinics {
        BIGSERIAL id PK
    }
    inventory_items {
        BIGSERIAL id PK,FK,UK
        bigint clinic_id FK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ inventory_quantity_bases : "clinic_id"
    staffs ||..o{ inventory_quantity_bases : "confirmed_by"
    inventory_items ||..o| inventory_quantity_bases : "inventory_item_id, clinic_id"
    operation_bindings ||..o{ inventory_quantity_bases : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `inventory_quantity_bases_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9735](../../../backend/migrations/001_init.sql#L9735) |
| `inventory_quantity_bases_confirmed_by_fkey` | `confirmed_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9741](../../../backend/migrations/001_init.sql#L9741) |
| `inventory_quantity_bases_inventory_item_id_clinic_id_fkey` | `inventory_item_id, clinic_id` | [inventory_items](inventory_items.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9746](../../../backend/migrations/001_init.sql#L9746) |
| `inventory_quantity_bases_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9747](../../../backend/migrations/001_init.sql#L9747) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `inventory_quantity_bases_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `inventory_quantity_bases_confirmed_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `inventory_quantity_bases_inventory_item_id_clinic_id_fkey`: `FOREIGN KEY (inventory_item_id,clinic_id) REFERENCES inventory_items(id,clinic_id) ON DELETE RESTRICT`
- `inventory_quantity_bases_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `inventory_quantity_bases_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9734](../../../backend/migrations/001_init.sql#L9734) |
| `inventory_quantity_bases_expected_item_version_check` | check | `CHECK (expected_item_version > 0)` | [001_init.sql:9737](../../../backend/migrations/001_init.sql#L9737) |
| `inventory_quantity_bases_confirmed_quantity_check` | check | `CHECK (confirmed_quantity >= 0)` | [001_init.sql:9738](../../../backend/migrations/001_init.sql#L9738) |
| `inventory_quantity_bases_management_unit_check` | check | `CHECK (btrim(management_unit)<>'')` | [001_init.sql:9739](../../../backend/migrations/001_init.sql#L9739) |
| `inventory_quantity_bases_id_clinic_id_inventory_item_id_key` | unique | `UNIQUE (id,clinic_id,inventory_item_id)` | [001_init.sql:9744](../../../backend/migrations/001_init.sql#L9744) |
| `inventory_quantity_bases_clinic_id_inventory_item_id_key` | unique | `UNIQUE (clinic_id,inventory_item_id)` | [001_init.sql:9745](../../../backend/migrations/001_init.sql#L9745) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
