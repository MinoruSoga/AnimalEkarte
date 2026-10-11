<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_mapping_draft_items`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [023_lab_device_mapping_drafts.sql:46](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L46) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [023_lab_device_mapping_drafts.sql:47](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L47) |
| `draft_id` | `bigint` | NOT NULL | — | FK | [023_lab_device_mapping_drafts.sql:48](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L48) |
| `device_item_code` | `varchar(64)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:49](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L49) |
| `value_raw` | `text` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:50](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L50) |
| `device_unit` | `varchar(64)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:51](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L51) |
| `chart_unit` | `varchar(32)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:52](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L52) |
| `unit_mismatch` | `boolean` | NOT NULL | `false` | — | [023_lab_device_mapping_drafts.sql:53](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L53) |
| `exam_type_field_id` | `bigint` | NULL | — | FK | [023_lab_device_mapping_drafts.sql:54](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L54) |
| `proposed_field_name` | `varchar(255)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:55](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L55) |
| `item_source` | `varchar(32)` | NOT NULL | — | — | [023_lab_device_mapping_drafts.sql:56](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L56) |
| `llm_request` | `jsonb` | NULL | — | — | [023_lab_device_mapping_drafts.sql:57](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L57) |
| `llm_response` | `jsonb` | NULL | — | — | [023_lab_device_mapping_drafts.sql:58](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L58) |
| `llm_error` | `varchar(64)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:59](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L59) |
| `sort_order` | `integer` | NOT NULL | `0` | — | [023_lab_device_mapping_drafts.sql:60](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L60) |
| `confirmed` | `boolean` | NOT NULL | `false` | — | [023_lab_device_mapping_drafts.sql:61](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L61) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [023_lab_device_mapping_drafts.sql:62](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L62) |
| `deleted_at` | `timestamptz` | NULL | — | — | [023_lab_device_mapping_drafts.sql:63](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L63) |

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
| `lab_device_mapping_draft_items_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [023_lab_device_mapping_drafts.sql:47](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L47) |
| `fk_lab_device_mapping_draft_items_field_clinic` | `exam_type_field_id, clinic_id` | [exam_type_fields](exam_type_fields.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [023_lab_device_mapping_drafts.sql:73](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L73) |
| `fk_lab_device_mapping_draft_items_draft_clinic` | `clinic_id, draft_id` | [lab_device_mapping_drafts](lab_device_mapping_drafts.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [024_lab_device_mapping_drafts_clinic_unique.sql:15](../../../backend/migrations/024_lab_device_mapping_drafts_clinic_unique.sql#L15) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_mapping_draft_items_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_lab_device_mapping_draft_items_field_clinic`: `CONSTRAINT fk_lab_device_mapping_draft_items_field_clinic     FOREIGN KEY (exam_type_field_id, clinic_id)     REFERENCES exam_type_fields (id, clinic_id)     ON DELETE RESTRICT`
- `fk_lab_device_mapping_draft_items_draft_clinic`: `CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic     FOREIGN KEY (clinic_id, draft_id)     REFERENCES lab_device_mapping_drafts (clinic_id, id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_mapping_draft_items_pkey` | primary | `PRIMARY KEY` | [023_lab_device_mapping_drafts.sql:46](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L46) |
| `chk_lab_device_mapping_draft_items_source` | check | `CONSTRAINT chk_lab_device_mapping_draft_items_source         CHECK (item_source IN ('dictionary', 'llm', 'blank', 'comment'))` | [023_lab_device_mapping_drafts.sql:68](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L68) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lab_device_mapping_draft_items_clinic_draft` | NO | — | `CREATE INDEX idx_lab_device_mapping_draft_items_clinic_draft     ON lab_device_mapping_draft_items (clinic_id, draft_id, sort_order)` | [023_lab_device_mapping_drafts.sql:78](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L78) |
