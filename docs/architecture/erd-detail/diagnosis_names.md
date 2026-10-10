<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `diagnosis_names`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1041](../../../backend/migrations/001_init.sql#L1041) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1042](../../../backend/migrations/001_init.sql#L1042) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:1043](../../../backend/migrations/001_init.sql#L1043) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:1044](../../../backend/migrations/001_init.sql#L1044) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:1045](../../../backend/migrations/001_init.sql#L1045) |
| `diagnosis_type_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1046](../../../backend/migrations/001_init.sql#L1046) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1047](../../../backend/migrations/001_init.sql#L1047) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1048](../../../backend/migrations/001_init.sql#L1048) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1049](../../../backend/migrations/001_init.sql#L1049) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1050](../../../backend/migrations/001_init.sql#L1050) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    diagnosis_names {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        boolean is_active
        text description
        bigint diagnosis_type_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    diagnosis_types {
        BIGSERIAL id PK
    }
    clinics ||..o{ diagnosis_names : "clinic_id"
    diagnosis_types ||..o{ diagnosis_names : "diagnosis_type_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `diagnosis_names_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1042](../../../backend/migrations/001_init.sql#L1042) |
| `diagnosis_names_diagnosis_type_id_fkey` | `diagnosis_type_id` | [diagnosis_types](diagnosis_types.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1046](../../../backend/migrations/001_init.sql#L1046) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `diagnosis_names_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `diagnosis_names_diagnosis_type_id_fkey`: `REFERENCES diagnosis_types(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `diagnosis_names_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1041](../../../backend/migrations/001_init.sql#L1041) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_diagnosis_names_clinic_id` | NO | — | `CREATE INDEX idx_diagnosis_names_clinic_id ON diagnosis_names(clinic_id)` | [001_init.sql:2229](../../../backend/migrations/001_init.sql#L2229) |
| `idx_diagnosis_names_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_diagnosis_names_clinic_sort   ON diagnosis_names(clinic_id, sort_order)         WHERE deleted_at IS NULL` | [001_init.sql:2261](../../../backend/migrations/001_init.sql#L2261) |
| `idx_diagnosis_names_type_id` | NO | — | `CREATE INDEX idx_diagnosis_names_type_id ON diagnosis_names(diagnosis_type_id)` | [001_init.sql:2431](../../../backend/migrations/001_init.sql#L2431) |
| `idx_diagnosis_names_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_diagnosis_names_active ON diagnosis_names(id) WHERE deleted_at IS NULL` | [001_init.sql:2481](../../../backend/migrations/001_init.sql#L2481) |
