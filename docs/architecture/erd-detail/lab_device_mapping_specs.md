<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_mapping_specs`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [001_init.sql:7059](../../../backend/migrations/001_init.sql#L7059) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:7060](../../../backend/migrations/001_init.sql#L7060) |
| `source_type` | `varchar(32)` | NOT NULL | — | UK | [001_init.sql:7061](../../../backend/migrations/001_init.sql#L7061) |
| `species` | `varchar(32)` | NOT NULL | — | UK | [001_init.sql:7062](../../../backend/migrations/001_init.sql#L7062) |
| `device_item_code` | `varchar(64)` | NOT NULL | — | UK | [001_init.sql:7063](../../../backend/migrations/001_init.sql#L7063) |
| `unit` | `varchar(32)` | NOT NULL | `''` | — | [001_init.sql:7064](../../../backend/migrations/001_init.sql#L7064) |
| `value_shape` | `varchar(32)` | NOT NULL | — | — | [001_init.sql:7065](../../../backend/migrations/001_init.sql#L7065) |
| `exam_type_field_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:7066](../../../backend/migrations/001_init.sql#L7066) |
| `confirmed_draft_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:7067](../../../backend/migrations/001_init.sql#L7067) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:7068](../../../backend/migrations/001_init.sql#L7068) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:7069](../../../backend/migrations/001_init.sql#L7069) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:7070](../../../backend/migrations/001_init.sql#L7070) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lab_device_mapping_specs {
        bigserial id PK
        bigint clinic_id FK,UK
        varchar_32 source_type UK
        varchar_32 species UK
        varchar_64 device_item_code UK
        varchar_32 unit
        varchar_32 value_shape
        bigint exam_type_field_id FK
        bigint confirmed_draft_id FK
        timestamptz created_at
        timestamptz updated_at
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
    clinics ||..o{ lab_device_mapping_specs : "clinic_id"
    exam_type_fields ||..o{ lab_device_mapping_specs : "exam_type_field_id, clinic_id"
    lab_device_mapping_drafts ||..o{ lab_device_mapping_specs : "clinic_id, confirmed_draft_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lab_device_mapping_specs_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:7060](../../../backend/migrations/001_init.sql#L7060) |
| `fk_lab_device_mapping_specs_field_clinic` | `exam_type_field_id, clinic_id` | [exam_type_fields](exam_type_fields.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:7086](../../../backend/migrations/001_init.sql#L7086) |
| `fk_lab_device_mapping_specs_draft_clinic` | `clinic_id, confirmed_draft_id` | [lab_device_mapping_drafts](lab_device_mapping_drafts.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:7155](../../../backend/migrations/001_init.sql#L7155) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_mapping_specs_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_lab_device_mapping_specs_field_clinic`: `CONSTRAINT fk_lab_device_mapping_specs_field_clinic     FOREIGN KEY (exam_type_field_id, clinic_id)     REFERENCES exam_type_fields (id, clinic_id)     ON DELETE RESTRICT`
- `fk_lab_device_mapping_specs_draft_clinic`: `CONSTRAINT fk_lab_device_mapping_specs_draft_clinic     FOREIGN KEY (clinic_id, confirmed_draft_id)     REFERENCES lab_device_mapping_drafts (clinic_id, id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_mapping_specs_pkey` | primary | `PRIMARY KEY` | [001_init.sql:7059](../../../backend/migrations/001_init.sql#L7059) |
| `uq_lab_device_mapping_specs` | unique | `CONSTRAINT uq_lab_device_mapping_specs         UNIQUE (clinic_id, source_type, species, device_item_code)` | [001_init.sql:7071](../../../backend/migrations/001_init.sql#L7071) |
| `chk_lab_device_mapping_specs_species` | check | `CONSTRAINT chk_lab_device_mapping_specs_species         CHECK (species IN ('canine', 'feline'))` | [001_init.sql:7073](../../../backend/migrations/001_init.sql#L7073) |
| `chk_lab_device_mapping_specs_source_type` | check | `CONSTRAINT chk_lab_device_mapping_specs_source_type         CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab'))` | [001_init.sql:7075](../../../backend/migrations/001_init.sql#L7075) |
| `chk_lab_device_mapping_specs_value_shape` | check | `CONSTRAINT chk_lab_device_mapping_specs_value_shape         CHECK (value_shape IN ('numeric', 'inequality', 'qual_and_num', 'dash', 'text'))` | [001_init.sql:7077](../../../backend/migrations/001_init.sql#L7077) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
