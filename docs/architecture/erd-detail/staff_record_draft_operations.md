<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `staff_record_draft_operations`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9035](../../../backend/migrations/001_init.sql#L9035) |
| `draft_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9036](../../../backend/migrations/001_init.sql#L9036) |
| `operation_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9037](../../../backend/migrations/001_init.sql#L9037) |
| `input_generation` | `bigint` | NOT NULL | — | — | [001_init.sql:9038](../../../backend/migrations/001_init.sql#L9038) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    staff_record_draft_operations {
        bigint clinic_id PK,FK
        bigint draft_id PK,FK
        bigint operation_id PK,FK
        bigint input_generation
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staff_record_drafts {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ staff_record_draft_operations : "clinic_id"
    staff_record_drafts ||--o{ staff_record_draft_operations : "draft_id, clinic_id"
    operation_bindings ||--o{ staff_record_draft_operations : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `staff_record_draft_operations_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9035](../../../backend/migrations/001_init.sql#L9035) |
| `staff_record_draft_operations_draft_id_clinic_id_fkey` | `draft_id, clinic_id` | [staff_record_drafts](staff_record_drafts.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9040](../../../backend/migrations/001_init.sql#L9040) |
| `staff_record_draft_operations_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9041](../../../backend/migrations/001_init.sql#L9041) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `staff_record_draft_operations_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `staff_record_draft_operations_draft_id_clinic_id_fkey`: `FOREIGN KEY (draft_id,clinic_id) REFERENCES staff_record_drafts(id,clinic_id) ON DELETE RESTRICT`
- `staff_record_draft_operations_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `staff_record_draft_operations_input_generation_check` | check | `CHECK (input_generation > 0)` | [001_init.sql:9038](../../../backend/migrations/001_init.sql#L9038) |
| `staff_record_draft_operations_pkey` | primary | `PRIMARY KEY (clinic_id,draft_id,operation_id)` | [001_init.sql:9039](../../../backend/migrations/001_init.sql#L9039) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
