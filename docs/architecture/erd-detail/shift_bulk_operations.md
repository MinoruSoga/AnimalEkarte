<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `shift_bulk_operations`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8990](../../../backend/migrations/001_init.sql#L8990) |
| `operation_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8991](../../../backend/migrations/001_init.sql#L8991) |
| `operation_kind` | `text` | NOT NULL | — | — | [001_init.sql:8992](../../../backend/migrations/001_init.sql#L8992) |
| `source_snapshot` | `jsonb` | NOT NULL | — | — | [001_init.sql:8993](../../../backend/migrations/001_init.sql#L8993) |
| `confirmed_rows` | `jsonb` | NOT NULL | — | — | [001_init.sql:8994](../../../backend/migrations/001_init.sql#L8994) |
| `preview_digest` | `bytea` | NOT NULL | — | — | [001_init.sql:8995](../../../backend/migrations/001_init.sql#L8995) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    shift_bulk_operations {
        bigint clinic_id PK,FK
        bigint operation_id PK,FK
        text operation_kind
        jsonb source_snapshot
        jsonb confirmed_rows
        bytea preview_digest
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ shift_bulk_operations : "clinic_id"
    operation_bindings ||--o| shift_bulk_operations : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `shift_bulk_operations_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8990](../../../backend/migrations/001_init.sql#L8990) |
| `shift_bulk_operations_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8997](../../../backend/migrations/001_init.sql#L8997) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `shift_bulk_operations_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `shift_bulk_operations_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `shift_bulk_operations_operation_kind_check` | check | `CHECK (operation_kind IN ('copy','csv'))` | [001_init.sql:8992](../../../backend/migrations/001_init.sql#L8992) |
| `shift_bulk_operations_source_snapshot_check` | check | `CHECK (jsonb_typeof(source_snapshot)='object')` | [001_init.sql:8993](../../../backend/migrations/001_init.sql#L8993) |
| `shift_bulk_operations_confirmed_rows_check` | check | `CHECK (jsonb_typeof(confirmed_rows)='array')` | [001_init.sql:8994](../../../backend/migrations/001_init.sql#L8994) |
| `shift_bulk_operations_preview_digest_check` | check | `CHECK (octet_length(preview_digest)=32)` | [001_init.sql:8995](../../../backend/migrations/001_init.sql#L8995) |
| `shift_bulk_operations_pkey` | primary | `PRIMARY KEY (clinic_id,operation_id)` | [001_init.sql:8996](../../../backend/migrations/001_init.sql#L8996) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
