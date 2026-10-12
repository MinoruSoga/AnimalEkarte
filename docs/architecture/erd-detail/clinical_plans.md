<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `clinical_plans`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1566](../../../backend/migrations/001_init.sql#L1566) |
| `medical_record_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1567](../../../backend/migrations/001_init.sql#L1567) |
| `physical_exam` | `text` | NOT NULL | `''` | — | [001_init.sql:1568](../../../backend/migrations/001_init.sql#L1568) |
| `diagnosis_type_id` | `bigint` | NULL | — | FK | [001_init.sql:1569](../../../backend/migrations/001_init.sql#L1569) |
| `diagnosis_name_id` | `bigint` | NULL | — | FK | [001_init.sql:1570](../../../backend/migrations/001_init.sql#L1570) |
| `diagnosis_2_type_id` | `bigint` | NULL | — | FK | [001_init.sql:1571](../../../backend/migrations/001_init.sql#L1571) |
| `diagnosis_2_name_id` | `bigint` | NULL | — | FK | [001_init.sql:1572](../../../backend/migrations/001_init.sql#L1572) |
| `diagnosis_details` | `text` | NOT NULL | `''` | — | [001_init.sql:1573](../../../backend/migrations/001_init.sql#L1573) |
| `treatment_policy` | `text` | NOT NULL | `''` | — | [001_init.sql:1574](../../../backend/migrations/001_init.sql#L1574) |
| `version` | `INTEGER` | NOT NULL | `1` | — | [001_init.sql:1575](../../../backend/migrations/001_init.sql#L1575) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1576](../../../backend/migrations/001_init.sql#L1576) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1577](../../../backend/migrations/001_init.sql#L1577) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1578](../../../backend/migrations/001_init.sql#L1578) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    clinical_plans {
        BIGSERIAL id PK
        bigint medical_record_id FK,UK
        text physical_exam
        bigint diagnosis_type_id FK
        bigint diagnosis_name_id FK
        bigint diagnosis_2_type_id FK
        bigint diagnosis_2_name_id FK
        text diagnosis_details
        text treatment_policy
        INTEGER version
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    diagnosis_names {
        BIGSERIAL id PK
    }
    diagnosis_types {
        BIGSERIAL id PK
    }
    medical_records {
        BIGSERIAL id PK,UK
    }
    medical_records ||..o| clinical_plans : "medical_record_id"
    diagnosis_types o|..o{ clinical_plans : "diagnosis_type_id"
    diagnosis_names o|..o{ clinical_plans : "diagnosis_name_id"
    diagnosis_types o|..o{ clinical_plans : "diagnosis_2_type_id"
    diagnosis_names o|..o{ clinical_plans : "diagnosis_2_name_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinical_plans_medical_record_id_fkey` | `medical_record_id` | [medical_records](medical_records.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1567](../../../backend/migrations/001_init.sql#L1567) |
| `clinical_plans_diagnosis_type_id_fkey` | `diagnosis_type_id` | [diagnosis_types](diagnosis_types.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1569](../../../backend/migrations/001_init.sql#L1569) |
| `clinical_plans_diagnosis_name_id_fkey` | `diagnosis_name_id` | [diagnosis_names](diagnosis_names.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1570](../../../backend/migrations/001_init.sql#L1570) |
| `clinical_plans_diagnosis_2_type_id_fkey` | `diagnosis_2_type_id` | [diagnosis_types](diagnosis_types.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1571](../../../backend/migrations/001_init.sql#L1571) |
| `clinical_plans_diagnosis_2_name_id_fkey` | `diagnosis_2_name_id` | [diagnosis_names](diagnosis_names.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1572](../../../backend/migrations/001_init.sql#L1572) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `clinical_plans_medical_record_id_fkey`: `REFERENCES medical_records(id) ON DELETE CASCADE`
- `clinical_plans_diagnosis_type_id_fkey`: `REFERENCES diagnosis_types(id) ON DELETE SET NULL`
- `clinical_plans_diagnosis_name_id_fkey`: `REFERENCES diagnosis_names(id) ON DELETE SET NULL`
- `clinical_plans_diagnosis_2_type_id_fkey`: `REFERENCES diagnosis_types(id) ON DELETE SET NULL`
- `clinical_plans_diagnosis_2_name_id_fkey`: `REFERENCES diagnosis_names(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `clinical_plans_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1566](../../../backend/migrations/001_init.sql#L1566) |
| `clinical_plans_medical_record_id_key` | unique | `UNIQUE` | [001_init.sql:1567](../../../backend/migrations/001_init.sql#L1567) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_clinical_plans_medical_record_id` | NO | — | `CREATE INDEX idx_clinical_plans_medical_record_id ON clinical_plans(medical_record_id)` | [001_init.sql:2292](../../../backend/migrations/001_init.sql#L2292) |
