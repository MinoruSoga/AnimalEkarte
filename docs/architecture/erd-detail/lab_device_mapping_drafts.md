<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_mapping_drafts`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK, UK | [023_lab_device_mapping_drafts.sql:17](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L17) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [023_lab_device_mapping_drafts.sql:18](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L18) |
| `source_type` | `varchar(32)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:19](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L19) |
| `species` | `varchar(32)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:20](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L20) |
| `status` | `varchar(32)` | NOT NULL | — | — | [023_lab_device_mapping_drafts.sql:21](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L21) |
| `raw_payload` | `bytea` | NOT NULL | — | — | [023_lab_device_mapping_drafts.sql:22](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L22) |
| `raw_fingerprint` | `varchar(64)` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:23](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L23) |
| `error_message` | `text` | NOT NULL | `''` | — | [023_lab_device_mapping_drafts.sql:24](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L24) |
| `confirmed_at` | `timestamptz` | NULL | — | — | [023_lab_device_mapping_drafts.sql:25](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L25) |
| `confirmed_by_staff_id` | `bigint` | NULL | — | — | [023_lab_device_mapping_drafts.sql:26](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L26) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [023_lab_device_mapping_drafts.sql:27](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L27) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [023_lab_device_mapping_drafts.sql:28](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L28) |
| `deleted_at` | `timestamptz` | NULL | — | — | [023_lab_device_mapping_drafts.sql:29](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L29) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lab_device_mapping_drafts {
        bigserial id PK,UK
        bigint clinic_id FK,UK
        varchar_32 source_type
        varchar_32 species
        varchar_32 status
        bytea raw_payload
        varchar_64 raw_fingerprint
        text error_message
        timestamptz confirmed_at
        bigint confirmed_by_staff_id
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ lab_device_mapping_drafts : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lab_device_mapping_drafts_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [023_lab_device_mapping_drafts.sql:18](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L18) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_mapping_drafts_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_mapping_drafts_pkey` | primary | `PRIMARY KEY` | [023_lab_device_mapping_drafts.sql:17](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L17) |
| `chk_lab_device_mapping_drafts_source_type` | check | `CONSTRAINT chk_lab_device_mapping_drafts_source_type         CHECK (source_type IN ('', 'fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab'))` | [023_lab_device_mapping_drafts.sql:31](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L31) |
| `chk_lab_device_mapping_drafts_species` | check | `CONSTRAINT chk_lab_device_mapping_drafts_species         CHECK (species IN ('', 'canine', 'feline'))` | [023_lab_device_mapping_drafts.sql:33](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L33) |
| `chk_lab_device_mapping_drafts_status` | check | `CONSTRAINT chk_lab_device_mapping_drafts_status         CHECK (status IN ('unconfirmed', 'confirmed', 'failed'))` | [023_lab_device_mapping_drafts.sql:35](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L35) |
| `uq_lab_device_mapping_drafts_id_clinic` | unique | `CONSTRAINT uq_lab_device_mapping_drafts_id_clinic     UNIQUE (clinic_id, id)` | [024_lab_device_mapping_drafts_clinic_unique.sql:11](../../../backend/migrations/024_lab_device_mapping_drafts_clinic_unique.sql#L11) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lab_device_mapping_drafts_clinic_source` | NO | — | `CREATE INDEX idx_lab_device_mapping_drafts_clinic_source     ON lab_device_mapping_drafts (clinic_id, source_type, species, created_at)` | [023_lab_device_mapping_drafts.sql:39](../../../backend/migrations/023_lab_device_mapping_drafts.sql#L39) |
