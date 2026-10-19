<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `closing_boundary_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9868](../../../backend/migrations/001_init.sql#L9868) |
| `revision` | `bigint` | NOT NULL | — | PK | [001_init.sql:9869](../../../backend/migrations/001_init.sql#L9869) |
| `boundary_snapshot` | `jsonb` | NOT NULL | — | — | [001_init.sql:9870](../../../backend/migrations/001_init.sql#L9870) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9871](../../../backend/migrations/001_init.sql#L9871) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9872](../../../backend/migrations/001_init.sql#L9872) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    closing_boundary_revisions {
        bigint clinic_id PK,FK
        bigint revision PK
        jsonb boundary_snapshot
        bigint operation_id FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ closing_boundary_revisions : "clinic_id"
    operation_bindings ||..o{ closing_boundary_revisions : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `closing_boundary_revisions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9868](../../../backend/migrations/001_init.sql#L9868) |
| `closing_boundary_revisions_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9874](../../../backend/migrations/001_init.sql#L9874) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `closing_boundary_revisions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `closing_boundary_revisions_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `closing_boundary_revisions_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:9869](../../../backend/migrations/001_init.sql#L9869) |
| `closing_boundary_revisions_boundary_snapshot_check` | check | `CHECK (jsonb_typeof(boundary_snapshot)='object')` | [001_init.sql:9870](../../../backend/migrations/001_init.sql#L9870) |
| `closing_boundary_revisions_pkey` | primary | `PRIMARY KEY (clinic_id,revision)` | [001_init.sql:9873](../../../backend/migrations/001_init.sql#L9873) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
