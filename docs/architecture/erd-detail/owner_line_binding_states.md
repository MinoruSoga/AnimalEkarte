<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `owner_line_binding_states`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:10155](../../../backend/migrations/001_init.sql#L10155) |
| `owner_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:10155](../../../backend/migrations/001_init.sql#L10155) |
| `revision` | `bigint` | NOT NULL | `1` | — | [001_init.sql:10156](../../../backend/migrations/001_init.sql#L10156) |
| `state` | `text` | NOT NULL | — | — | [001_init.sql:10156](../../../backend/migrations/001_init.sql#L10156) |
| `line_customer_id` | `bigint` | NULL | — | FK | [001_init.sql:10157](../../../backend/migrations/001_init.sql#L10157) |
| `verified_by` | `bigint` | NULL | — | FK | [001_init.sql:10158](../../../backend/migrations/001_init.sql#L10158) |
| `verified_at` | `timestamptz` | NULL | — | — | [001_init.sql:10158](../../../backend/migrations/001_init.sql#L10158) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10159](../../../backend/migrations/001_init.sql#L10159) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    owner_line_binding_states {
        bigint clinic_id PK,FK
        bigint owner_id PK,FK
        bigint revision
        text state
        bigint line_customer_id FK
        bigint verified_by FK
        timestamptz verified_at
        bigint operation_id FK
    }
    clinics {
        BIGSERIAL id PK
    }
    line_customers {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||--o{ owner_line_binding_states : "clinic_id"
    line_customers o|..o{ owner_line_binding_states : "line_customer_id"
    staffs o|..o{ owner_line_binding_states : "verified_by"
    owners ||--o| owner_line_binding_states : "owner_id, clinic_id"
    operation_bindings ||..o{ owner_line_binding_states : "operation_id, clinic_id"
    line_customers o|..o{ owner_line_binding_states : "line_customer_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `owner_line_binding_states_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10155](../../../backend/migrations/001_init.sql#L10155) |
| `owner_line_binding_states_line_customer_id_fkey` | `line_customer_id` | [line_customers](line_customers.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10157](../../../backend/migrations/001_init.sql#L10157) |
| `owner_line_binding_states_verified_by_fkey` | `verified_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10158](../../../backend/migrations/001_init.sql#L10158) |
| `owner_line_binding_states_owner_id_clinic_id_fkey` | `owner_id, clinic_id` | [owners](owners.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10161](../../../backend/migrations/001_init.sql#L10161) |
| `owner_line_binding_states_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10162](../../../backend/migrations/001_init.sql#L10162) |
| `fk_line_binding_customer_clinic` | `line_customer_id, clinic_id` | [line_customers](line_customers.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10907](../../../backend/migrations/001_init.sql#L10907) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `owner_line_binding_states_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `owner_line_binding_states_line_customer_id_fkey`: `REFERENCES line_customers(id) ON DELETE RESTRICT`
- `owner_line_binding_states_verified_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `owner_line_binding_states_owner_id_clinic_id_fkey`: `FOREIGN KEY(owner_id,clinic_id) REFERENCES owners(id,clinic_id) ON DELETE RESTRICT`
- `owner_line_binding_states_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`
- `fk_line_binding_customer_clinic`: `CONSTRAINT fk_line_binding_customer_clinic FOREIGN KEY(line_customer_id,clinic_id) REFERENCES line_customers(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `owner_line_binding_states_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10156](../../../backend/migrations/001_init.sql#L10156) |
| `owner_line_binding_states_state_check` | check | `CHECK(state IN('linked','unlinked','unlink_pending','legacy_unknown'))` | [001_init.sql:10156](../../../backend/migrations/001_init.sql#L10156) |
| `owner_line_binding_states_pkey` | primary | `PRIMARY KEY(clinic_id,owner_id)` | [001_init.sql:10159](../../../backend/migrations/001_init.sql#L10159) |
| `owner_line_binding_states_check` | check | `CHECK(num_nonnulls(verified_by,verified_at) IN(0,2))` | [001_init.sql:10160](../../../backend/migrations/001_init.sql#L10160) |
| `owner_line_binding_states_check1` | check | `CHECK(state<>'linked' OR line_customer_id IS NOT NULL)` | [001_init.sql:10160](../../../backend/migrations/001_init.sql#L10160) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uq_linked_line_customer_owner` | YES | `state IN('linked','unlink_pending')` | `CREATE UNIQUE INDEX uq_linked_line_customer_owner ON owner_line_binding_states(clinic_id,line_customer_id) WHERE state IN('linked','unlink_pending')` | [001_init.sql:10908](../../../backend/migrations/001_init.sql#L10908) |
