<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `staff_clinic_assignments`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1180](../../../backend/migrations/001_init.sql#L1180) |
| `staff_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1181](../../../backend/migrations/001_init.sql#L1181) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1182](../../../backend/migrations/001_init.sql#L1182) |
| `is_main` | `boolean` | NOT NULL | `false` | — | [001_init.sql:1183](../../../backend/migrations/001_init.sql#L1183) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1184](../../../backend/migrations/001_init.sql#L1184) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1185](../../../backend/migrations/001_init.sql#L1185) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1186](../../../backend/migrations/001_init.sql#L1186) |
| `capabilities_configured_at` | `TIMESTAMPTZ` | NULL | — | — | [022_staff_clinic_assignments_capability_configured_at.sql:18](../../../backend/migrations/022_staff_clinic_assignments_capability_configured_at.sql#L18) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    staff_clinic_assignments {
        BIGSERIAL id PK
        bigint staff_id FK,UK
        bigint clinic_id FK,UK
        boolean is_main
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        TIMESTAMPTZ capabilities_configured_at
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    staffs ||..o{ staff_clinic_assignments : "staff_id"
    clinics ||..o{ staff_clinic_assignments : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `staff_clinic_assignments_staff_id_fkey` | `staff_id` | [staffs](staffs.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1181](../../../backend/migrations/001_init.sql#L1181) |
| `staff_clinic_assignments_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1182](../../../backend/migrations/001_init.sql#L1182) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `staff_clinic_assignments_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE CASCADE`
- `staff_clinic_assignments_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `staff_clinic_assignments_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1180](../../../backend/migrations/001_init.sql#L1180) |
| `uk_staff_clinic` | unique | `CONSTRAINT uk_staff_clinic UNIQUE (staff_id, clinic_id)` | [001_init.sql:1187](../../../backend/migrations/001_init.sql#L1187) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_staff_clinic_staff` | NO | — | `CREATE INDEX idx_staff_clinic_staff ON staff_clinic_assignments(staff_id)` | [001_init.sql:1190](../../../backend/migrations/001_init.sql#L1190) |
| `idx_staff_clinic_clinic` | NO | — | `CREATE INDEX idx_staff_clinic_clinic ON staff_clinic_assignments(clinic_id)` | [001_init.sql:1191](../../../backend/migrations/001_init.sql#L1191) |
| `idx_staff_clinic_main` | NO | — | `CREATE INDEX idx_staff_clinic_main ON staff_clinic_assignments(staff_id, is_main)` | [001_init.sql:1192](../../../backend/migrations/001_init.sql#L1192) |
