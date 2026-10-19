<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `reservation_type_name_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8888](../../../backend/migrations/001_init.sql#L8888) |
| `reservation_type_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8889](../../../backend/migrations/001_init.sql#L8889) |
| `revision` | `bigint` | NOT NULL | — | PK | [001_init.sql:8890](../../../backend/migrations/001_init.sql#L8890) |
| `previous_name` | `text` | NULL | — | — | [001_init.sql:8891](../../../backend/migrations/001_init.sql#L8891) |
| `display_name` | `text` | NOT NULL | — | — | [001_init.sql:8892](../../../backend/migrations/001_init.sql#L8892) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:8893](../../../backend/migrations/001_init.sql#L8893) |
| `changed_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:8894](../../../backend/migrations/001_init.sql#L8894) |
| `changed_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8895](../../../backend/migrations/001_init.sql#L8895) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    reservation_type_name_revisions {
        bigint clinic_id PK,FK
        bigint reservation_type_id PK,FK
        bigint revision PK
        text previous_name
        text display_name
        bigint operation_id FK
        bigint changed_by FK
        timestamptz changed_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    reservation_types {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||--o{ reservation_type_name_revisions : "clinic_id"
    staffs ||..o{ reservation_type_name_revisions : "changed_by"
    reservation_types ||--o{ reservation_type_name_revisions : "reservation_type_id, clinic_id"
    operation_bindings ||..o{ reservation_type_name_revisions : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `reservation_type_name_revisions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8888](../../../backend/migrations/001_init.sql#L8888) |
| `reservation_type_name_revisions_changed_by_fkey` | `changed_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8894](../../../backend/migrations/001_init.sql#L8894) |
| `reservation_type_name_revisio_reservation_type_id_clinic_i_fkey` | `reservation_type_id, clinic_id` | [reservation_types](reservation_types.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8897](../../../backend/migrations/001_init.sql#L8897) |
| `reservation_type_name_revisions_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8898](../../../backend/migrations/001_init.sql#L8898) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `reservation_type_name_revisions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `reservation_type_name_revisions_changed_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `reservation_type_name_revisio_reservation_type_id_clinic_i_fkey`: `FOREIGN KEY (reservation_type_id,clinic_id) REFERENCES reservation_types(id,clinic_id) ON DELETE RESTRICT`
- `reservation_type_name_revisions_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `reservation_type_name_revisions_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:8890](../../../backend/migrations/001_init.sql#L8890) |
| `reservation_type_name_revisions_pkey` | primary | `PRIMARY KEY (clinic_id,reservation_type_id,revision)` | [001_init.sql:8896](../../../backend/migrations/001_init.sql#L8896) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
