<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `checkup_result_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9966](../../../backend/migrations/001_init.sql#L9966) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9966](../../../backend/migrations/001_init.sql#L9966) |
| `checkup_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9967](../../../backend/migrations/001_init.sql#L9967) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:9967](../../../backend/migrations/001_init.sql#L9967) |
| `package_version` | `bigint` | NOT NULL | — | — | [001_init.sql:9968](../../../backend/migrations/001_init.sql#L9968) |
| `package_snapshot` | `jsonb` | NOT NULL | — | — | [001_init.sql:9968](../../../backend/migrations/001_init.sql#L9968) |
| `fields_snapshot` | `jsonb` | NOT NULL | — | — | [001_init.sql:9969](../../../backend/migrations/001_init.sql#L9969) |
| `restored_from_id` | `bigint` | NULL | — | FK | [001_init.sql:9970](../../../backend/migrations/001_init.sql#L9970) |
| `reason` | `text` | NULL | — | — | [001_init.sql:9970](../../../backend/migrations/001_init.sql#L9970) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9970](../../../backend/migrations/001_init.sql#L9970) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9971](../../../backend/migrations/001_init.sql#L9971) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9971](../../../backend/migrations/001_init.sql#L9971) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    checkup_result_revisions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint checkup_id FK,UK
        bigint revision UK
        bigint package_version
        jsonb package_snapshot
        jsonb fields_snapshot
        bigint restored_from_id FK
        text reason
        bigint operation_id FK
        bigint recorded_by FK
        timestamptz recorded_at
    }
    checkups {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ checkup_result_revisions : "clinic_id"
    staffs ||..o{ checkup_result_revisions : "recorded_by"
    checkups ||..o{ checkup_result_revisions : "checkup_id, clinic_id"
    checkup_result_revisions o|..o{ checkup_result_revisions : "restored_from_id, clinic_id, checkup_id"
    operation_bindings ||..o{ checkup_result_revisions : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `checkup_result_revisions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9966](../../../backend/migrations/001_init.sql#L9966) |
| `checkup_result_revisions_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9971](../../../backend/migrations/001_init.sql#L9971) |
| `checkup_result_revisions_checkup_id_clinic_id_fkey` | `checkup_id, clinic_id` | [checkups](checkups.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9973](../../../backend/migrations/001_init.sql#L9973) |
| `checkup_result_revisions_restored_from_id_clinic_id_checku_fkey` | `restored_from_id, clinic_id, checkup_id` | [checkup_result_revisions](checkup_result_revisions.md) | `id, clinic_id, checkup_id` | `ON DELETE RESTRICT` | [001_init.sql:9974](../../../backend/migrations/001_init.sql#L9974) |
| `checkup_result_revisions_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9975](../../../backend/migrations/001_init.sql#L9975) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `checkup_result_revisions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `checkup_result_revisions_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `checkup_result_revisions_checkup_id_clinic_id_fkey`: `FOREIGN KEY(checkup_id,clinic_id) REFERENCES checkups(id,clinic_id) ON DELETE RESTRICT`
- `checkup_result_revisions_restored_from_id_clinic_id_checku_fkey`: `FOREIGN KEY(restored_from_id,clinic_id,checkup_id) REFERENCES checkup_result_revisions(id,clinic_id,checkup_id) ON DELETE RESTRICT`
- `checkup_result_revisions_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `checkup_result_revisions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9966](../../../backend/migrations/001_init.sql#L9966) |
| `checkup_result_revisions_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:9967](../../../backend/migrations/001_init.sql#L9967) |
| `checkup_result_revisions_package_version_check` | check | `CHECK(package_version>0)` | [001_init.sql:9968](../../../backend/migrations/001_init.sql#L9968) |
| `checkup_result_revisions_package_snapshot_check` | check | `CHECK(jsonb_typeof(package_snapshot)='object')` | [001_init.sql:9968](../../../backend/migrations/001_init.sql#L9968) |
| `checkup_result_revisions_fields_snapshot_check` | check | `CHECK(jsonb_typeof(fields_snapshot)='array')` | [001_init.sql:9969](../../../backend/migrations/001_init.sql#L9969) |
| `checkup_result_revisions_id_clinic_id_checkup_id_key` | unique | `UNIQUE(id,clinic_id,checkup_id)` | [001_init.sql:9972](../../../backend/migrations/001_init.sql#L9972) |
| `checkup_result_revisions_clinic_id_checkup_id_revision_key` | unique | `UNIQUE(clinic_id,checkup_id,revision)` | [001_init.sql:9972](../../../backend/migrations/001_init.sql#L9972) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
