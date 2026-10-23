<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_item_masters`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [001_init.sql:5953](../../../backend/migrations/001_init.sql#L5953) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:5954](../../../backend/migrations/001_init.sql#L5954) |
| `source_type` | `varchar(32)` | NOT NULL | — | UK | [001_init.sql:5955](../../../backend/migrations/001_init.sql#L5955) |
| `device_item_code` | `varchar(64)` | NOT NULL | — | UK | [001_init.sql:5956](../../../backend/migrations/001_init.sql#L5956) |
| `unit` | `varchar(32)` | NOT NULL | `''` | — | [001_init.sql:5957](../../../backend/migrations/001_init.sql#L5957) |
| `value_shape` | `varchar(32)` | NOT NULL | — | — | [001_init.sql:5958](../../../backend/migrations/001_init.sql#L5958) |
| `exam_type_field_id` | `bigint` | NULL | — | FK | [001_init.sql:5959](../../../backend/migrations/001_init.sql#L5959) |
| `sort_order` | `integer` | NOT NULL | `0` | — | [001_init.sql:5960](../../../backend/migrations/001_init.sql#L5960) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:5961](../../../backend/migrations/001_init.sql#L5961) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:5962](../../../backend/migrations/001_init.sql#L5962) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:5963](../../../backend/migrations/001_init.sql#L5963) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lab_device_item_masters {
        bigserial id PK
        bigint clinic_id FK,UK
        varchar_32 source_type UK
        varchar_64 device_item_code UK
        varchar_32 unit
        varchar_32 value_shape
        bigint exam_type_field_id FK
        integer sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    exam_type_fields {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ lab_device_item_masters : "clinic_id"
    exam_type_fields o|..o{ lab_device_item_masters : "exam_type_field_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lab_device_item_masters_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:5954](../../../backend/migrations/001_init.sql#L5954) |
| `fk_lab_device_item_masters_field_clinic` | `exam_type_field_id, clinic_id` | [exam_type_fields](exam_type_fields.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:5973](../../../backend/migrations/001_init.sql#L5973) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_item_masters_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_lab_device_item_masters_field_clinic`: `CONSTRAINT fk_lab_device_item_masters_field_clinic     FOREIGN KEY (exam_type_field_id, clinic_id)     REFERENCES exam_type_fields (id, clinic_id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_item_masters_pkey` | primary | `PRIMARY KEY` | [001_init.sql:5953](../../../backend/migrations/001_init.sql#L5953) |
| `chk_lab_device_item_masters_source_type` | check | `CONSTRAINT chk_lab_device_item_masters_source_type         CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab'))` | [001_init.sql:5964](../../../backend/migrations/001_init.sql#L5964) |
| `chk_lab_device_item_masters_value_shape` | check | `CONSTRAINT chk_lab_device_item_masters_value_shape         CHECK (value_shape IN ('numeric', 'inequality', 'qual_and_num', 'dash', 'text'))` | [001_init.sql:5966](../../../backend/migrations/001_init.sql#L5966) |
| `uq_lab_device_item_masters_clinic_source_code` | unique | `CONSTRAINT uq_lab_device_item_masters_clinic_source_code         UNIQUE (clinic_id, source_type, device_item_code)` | [001_init.sql:5968](../../../backend/migrations/001_init.sql#L5968) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lab_device_item_masters_clinic_source` | NO | — | `CREATE INDEX idx_lab_device_item_masters_clinic_source     ON lab_device_item_masters (clinic_id, source_type, sort_order)` | [001_init.sql:5978](../../../backend/migrations/001_init.sql#L5978) |
