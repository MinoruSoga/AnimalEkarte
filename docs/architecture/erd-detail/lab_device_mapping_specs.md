<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_mapping_specs`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [023_lab_device_mapping_drafts.sql:87](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L87) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [023_lab_device_mapping_drafts.sql:88](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L88) |
| `source_type` | `varchar(32)` | NOT NULL | — | UK | [023_lab_device_mapping_drafts.sql:89](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L89) |
| `species` | `varchar(32)` | NOT NULL | — | UK | [023_lab_device_mapping_drafts.sql:90](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L90) |
| `device_item_code` | `varchar(64)` | NOT NULL | — | UK | [023_lab_device_mapping_drafts.sql:91](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L91) |
| `unit` | `varchar(32)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:92](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L92) |
| `value_shape` | `varchar(32)` | NOT NULL | — | — | [023_lab_device_mapping_drafts.sql:93](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L93) |
| `exam_type_field_id` | `bigint` | NOT NULL | — | FK | [023_lab_device_mapping_drafts.sql:94](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L94) |
| `confirmed_draft_id` | `bigint` | NOT NULL | — | FK | [023_lab_device_mapping_drafts.sql:95](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L95) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [023_lab_device_mapping_drafts.sql:96](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L96) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [023_lab_device_mapping_drafts.sql:97](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L97) |
| `deleted_at` | `timestamptz` | NULL | — | — | [023_lab_device_mapping_drafts.sql:98](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L98) |

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
| `lab_device_mapping_specs_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [023_lab_device_mapping_drafts.sql:88](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L88) |
| `fk_lab_device_mapping_specs_field_clinic` | `exam_type_field_id, clinic_id` | [exam_type_fields](exam_type_fields.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [023_lab_device_mapping_drafts.sql:114](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L114) |
| `fk_lab_device_mapping_specs_draft_clinic` | `clinic_id, confirmed_draft_id` | [lab_device_mapping_drafts](lab_device_mapping_drafts.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [024_lab_device_mapping_drafts_clinic_unique.sql:21](../../../backend/migrations/024_lab_device_mapping_drafts_clinic_unique.sql#L21) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_mapping_specs_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `fk_lab_device_mapping_specs_field_clinic`: `CONSTRAINT fk_lab_device_mapping_specs_field_clinic     FOREIGN KEY (exam_type_field_id, clinic_id)     REFERENCES exam_type_fields (id, clinic_id)     ON DELETE RESTRICT`
- `fk_lab_device_mapping_specs_draft_clinic`: `CONSTRAINT fk_lab_device_mapping_specs_draft_clinic     FOREIGN KEY (clinic_id, confirmed_draft_id)     REFERENCES lab_device_mapping_drafts (clinic_id, id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_mapping_specs_pkey` | primary | `PRIMARY KEY` | [023_lab_device_mapping_drafts.sql:87](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L87) |
| `uq_lab_device_mapping_specs` | unique | `CONSTRAINT uq_lab_device_mapping_specs         UNIQUE (clinic_id, source_type, species, device_item_code)` | [023_lab_device_mapping_drafts.sql:99](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L99) |
| `chk_lab_device_mapping_specs_species` | check | `CONSTRAINT chk_lab_device_mapping_specs_species         CHECK (species IN ('canine', 'feline'))` | [023_lab_device_mapping_drafts.sql:101](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L101) |
| `chk_lab_device_mapping_specs_source_type` | check | `CONSTRAINT chk_lab_device_mapping_specs_source_type         CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab'))` | [023_lab_device_mapping_drafts.sql:103](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L103) |
| `chk_lab_device_mapping_specs_value_shape` | check | `CONSTRAINT chk_lab_device_mapping_specs_value_shape         CHECK (value_shape IN ('numeric', 'inequality', 'qual_and_num', 'dash', 'text'))` | [023_lab_device_mapping_drafts.sql:105](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L105) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
