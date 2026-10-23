<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `estimate_billing_handoffs`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK, UK | [001_init.sql:9571](../../../backend/migrations/001_init.sql#L9571) |
| `presentation_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9572](../../../backend/migrations/001_init.sql#L9572) |
| `estimate_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9573](../../../backend/migrations/001_init.sql#L9573) |
| `estimate_item_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9574](../../../backend/migrations/001_init.sql#L9574) |
| `billing_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9575](../../../backend/migrations/001_init.sql#L9575) |
| `billing_item_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9576](../../../backend/migrations/001_init.sql#L9576) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9577](../../../backend/migrations/001_init.sql#L9577) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    estimate_billing_handoffs {
        bigint clinic_id PK,FK,UK
        bigint presentation_id PK,FK
        bigint estimate_id FK
        bigint estimate_item_id PK,FK
        bigint billing_id FK
        bigint billing_item_id FK,UK
        bigint operation_id FK
    }
    billing_items {
        BIGSERIAL id PK,UK
        bigint billing_id FK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    estimate_items {
        BIGSERIAL id PK,UK
        bigint estimate_id FK,UK
    }
    estimate_presentations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint estimate_id FK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ estimate_billing_handoffs : "clinic_id"
    estimate_presentations ||..o{ estimate_billing_handoffs : "presentation_id, clinic_id, estimate_id"
    estimate_items ||..o{ estimate_billing_handoffs : "estimate_item_id, estimate_id"
    billing_items ||..o| estimate_billing_handoffs : "billing_item_id, clinic_id, billing_id"
    operation_bindings ||..o{ estimate_billing_handoffs : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `estimate_billing_handoffs_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9571](../../../backend/migrations/001_init.sql#L9571) |
| `estimate_billing_handoffs_presentation_id_clinic_id_estima_fkey` | `presentation_id, clinic_id, estimate_id` | [estimate_presentations](estimate_presentations.md) | `id, clinic_id, estimate_id` | `ON DELETE RESTRICT` | [001_init.sql:9580](../../../backend/migrations/001_init.sql#L9580) |
| `estimate_billing_handoffs_estimate_item_id_estimate_id_fkey` | `estimate_item_id, estimate_id` | [estimate_items](estimate_items.md) | `id, estimate_id` | `ON DELETE RESTRICT` | [001_init.sql:9581](../../../backend/migrations/001_init.sql#L9581) |
| `estimate_billing_handoffs_billing_item_id_clinic_id_billin_fkey` | `billing_item_id, clinic_id, billing_id` | [billing_items](billing_items.md) | `id, clinic_id, billing_id` | `ON DELETE RESTRICT` | [001_init.sql:9582](../../../backend/migrations/001_init.sql#L9582) |
| `estimate_billing_handoffs_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9583](../../../backend/migrations/001_init.sql#L9583) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `estimate_billing_handoffs_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `estimate_billing_handoffs_presentation_id_clinic_id_estima_fkey`: `FOREIGN KEY (presentation_id,clinic_id,estimate_id) REFERENCES estimate_presentations(id,clinic_id,estimate_id) ON DELETE RESTRICT`
- `estimate_billing_handoffs_estimate_item_id_estimate_id_fkey`: `FOREIGN KEY (estimate_item_id,estimate_id) REFERENCES estimate_items(id,estimate_id) ON DELETE RESTRICT`
- `estimate_billing_handoffs_billing_item_id_clinic_id_billin_fkey`: `FOREIGN KEY (billing_item_id,clinic_id,billing_id) REFERENCES billing_items(id,clinic_id,billing_id) ON DELETE RESTRICT`
- `estimate_billing_handoffs_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `estimate_billing_handoffs_pkey` | primary | `PRIMARY KEY (clinic_id,presentation_id,estimate_item_id)` | [001_init.sql:9578](../../../backend/migrations/001_init.sql#L9578) |
| `estimate_billing_handoffs_clinic_id_billing_item_id_key` | unique | `UNIQUE (clinic_id,billing_item_id)` | [001_init.sql:9579](../../../backend/migrations/001_init.sql#L9579) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
