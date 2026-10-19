<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_mapping_draft_items`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [001_init.sql:7006](../../../backend/migrations/001_init.sql#L7006) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:7007](../../../backend/migrations/001_init.sql#L7007) |
| `draft_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:7008](../../../backend/migrations/001_init.sql#L7008) |
| `device_item_code` | `varchar(64)` | NOT NULL | `''` | — | [001_init.sql:7009](../../../backend/migrations/001_init.sql#L7009) |
| `value_raw` | `text` | NOT NULL | `''` | — | [001_init.sql:7010](../../../backend/migrations/001_init.sql#L7010) |
| `device_unit` | `varchar(64)` | NOT NULL | `''` | — | [001_init.sql:7011](../../../backend/migrations/001_init.sql#L7011) |
| `chart_unit` | `varchar(32)` | NOT NULL | `''` | — | [001_init.sql:7012](../../../backend/migrations/001_init.sql#L7012) |
| `unit_mismatch` | `boolean` | NOT NULL | `false` | — | [001_init.sql:7013](../../../backend/migrations/001_init.sql#L7013) |
| `exam_type_field_id` | `bigint` | NULL | — | FK | [001_init.sql:7014](../../../backend/migrations/001_init.sql#L7014) |
| `proposed_field_name` | `varchar(255)` | NOT NULL | `''` | — | [001_init.sql:7015](../../../backend/migrations/001_init.sql#L7015) |
| `item_source` | `varchar(32)` | NOT NULL | — | — | [001_init.sql:7016](../../../backend/migrations/001_init.sql#L7016) |
| `llm_request` | `jsonb` | NULL | — | — | [001_init.sql:7017](../../../backend/migrations/001_init.sql#L7017) |
| `llm_response` | `jsonb` | NULL | — | — | [001_init.sql:7018](../../../backend/migrations/001_init.sql#L7018) |
| `llm_error` | `varchar(64)` | NOT NULL | `''` | — | [001_init.sql:7019](../../../backend/migrations/001_init.sql#L7019) |
| `sort_order` | `integer` | NOT NULL | `0` | — | [001_init.sql:7020](../../../backend/migrations/001_init.sql#L7020) |
| `confirmed` | `boolean` | NOT NULL | `false` | — | [001_init.sql:7021](../../../backend/migrations/001_init.sql#L7021) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:7022](../../../backend/migrations/001_init.sql#L7022) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:7023](../../../backend/migrations/001_init.sql#L7023) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lab_device_mapping_draft_items {
        bigserial id PK
        bigint clinic_id FK
        bigint draft_id FK
        varchar_64 device_item_code
        text value_raw
        varchar_64 device_unit
        varchar_32 chart_unit
        boolean unit_mismatch
        bigint exam_type_field_id FK
        varchar_255 proposed_field_name
        varchar_32 item_source
        jsonb llm_request
        jsonb llm_response
        varchar_64 llm_error
        integer sort_order
        boolean confirmed
        timestamptz created_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    exam_type_fields {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    lab_device_mapping_drafts {
        bigserial id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ lab_device_mapping_draft_items : "clinic_id"
    exam_type_fields o|..o{ lab_device_mapping_draft_items : "exam_type_field_id, clinic_id"
    lab_device_mapping_drafts ||..o{ lab_device_mapping_draft_items : "clinic_id, draft_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lab_device_mapping_draft_items_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:7007](../../../backend/migrations/001_init.sql#L7007) |
| `fk_lab_device_mapping_draft_items_field_clinic` | `exam_type_field_id, clinic_id` | [exam_type_fields](exam_type_fields.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:7033](../../../backend/migrations/001_init.sql#L7033) |
| `fk_lab_device_mapping_draft_items_draft_clinic` | `clinic_id, draft_id` | [lab_device_mapping_drafts](lab_device_mapping_drafts.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:7137](../../../backend/migrations/001_init.sql#L7137) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_mapping_draft_items_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_lab_device_mapping_draft_items_field_clinic`: `CONSTRAINT fk_lab_device_mapping_draft_items_field_clinic     FOREIGN KEY (exam_type_field_id, clinic_id)     REFERENCES exam_type_fields (id, clinic_id)     ON DELETE RESTRICT`
- `fk_lab_device_mapping_draft_items_draft_clinic`: `CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic     FOREIGN KEY (clinic_id, draft_id)     REFERENCES lab_device_mapping_drafts (clinic_id, id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_mapping_draft_items_pkey` | primary | `PRIMARY KEY` | [001_init.sql:7006](../../../backend/migrations/001_init.sql#L7006) |
| `chk_lab_device_mapping_draft_items_source` | check | `CONSTRAINT chk_lab_device_mapping_draft_items_source         CHECK (item_source IN ('dictionary', 'llm', 'blank', 'comment'))` | [001_init.sql:7028](../../../backend/migrations/001_init.sql#L7028) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lab_device_mapping_draft_items_clinic_draft` | NO | — | `CREATE INDEX idx_lab_device_mapping_draft_items_clinic_draft     ON lab_device_mapping_draft_items (clinic_id, draft_id, sort_order)` | [001_init.sql:7038](../../../backend/migrations/001_init.sql#L7038) |
