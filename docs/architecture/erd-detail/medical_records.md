<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medical_records`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:1413](../../../backend/migrations/001_init.sql#L1413) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1414](../../../backend/migrations/001_init.sql#L1414) |
| `record_no` | `text` | NOT NULL | — | UK | [001_init.sql:1415](../../../backend/migrations/001_init.sql#L1415) |
| `date` | `date` | NOT NULL | — | — | [001_init.sql:1416](../../../backend/migrations/001_init.sql#L1416) |
| `owner_id` | `bigint` | NULL | — | FK | [001_init.sql:1417](../../../backend/migrations/001_init.sql#L1417) |
| `pet_id` | `bigint` | NULL | — | FK | [001_init.sql:1418](../../../backend/migrations/001_init.sql#L1418) |
| `doctor_id` | `bigint` | NULL | — | FK | [001_init.sql:1419](../../../backend/migrations/001_init.sql#L1419) |
| `appointment_id` | `bigint` | NULL | — | FK | [001_init.sql:1420](../../../backend/migrations/001_init.sql#L1420) |
| `status` | `medical_record_status` | NULL | `'draft'` | — | [001_init.sql:1421](../../../backend/migrations/001_init.sql#L1421) |
| `version` | `INTEGER` | NOT NULL | `1` | — | [001_init.sql:1422](../../../backend/migrations/001_init.sql#L1422) |
| `entered_by` | `bigint` | NULL | — | FK | [001_init.sql:1423](../../../backend/migrations/001_init.sql#L1423) |
| `recommendation_reason` | `varchar(100)` | NULL | — | — | [001_init.sql:1426](../../../backend/migrations/001_init.sql#L1426) |
| `visit_type` | `visit_type` | NULL | — | — | [001_init.sql:1427](../../../backend/migrations/001_init.sql#L1427) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1428](../../../backend/migrations/001_init.sql#L1428) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1429](../../../backend/migrations/001_init.sql#L1429) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1430](../../../backend/migrations/001_init.sql#L1430) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medical_records {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text record_no UK
        date date
        bigint owner_id FK
        bigint pet_id FK
        bigint doctor_id FK
        bigint appointment_id FK
        medical_record_status status
        INTEGER version
        bigint entered_by FK
        varchar_100 recommendation_reason
        visit_type visit_type
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    appointments {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    pets {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ medical_records : "clinic_id"
    owners o|..o{ medical_records : "owner_id"
    pets o|..o{ medical_records : "pet_id"
    pets o|..o{ medical_records : "clinic_id, pet_id"
    owners o|..o{ medical_records : "clinic_id, owner_id"
    appointments o|..o{ medical_records : "appointment_id, clinic_id"
    staffs o|..o{ medical_records : "entered_by"
    staffs o|..o{ medical_records : "doctor_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medical_records_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1414](../../../backend/migrations/001_init.sql#L1414) |
| `medical_records_owner_id_fkey` | `owner_id` | [owners](owners.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1417](../../../backend/migrations/001_init.sql#L1417) |
| `medical_records_pet_id_fkey` | `pet_id` | [pets](pets.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1418](../../../backend/migrations/001_init.sql#L1418) |
| `fk_medical_records_clinic_pet` | `clinic_id, pet_id` | [pets](pets.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:4321](../../../backend/migrations/001_init.sql#L4321) |
| `fk_medical_records_clinic_owner` | `clinic_id, owner_id` | [owners](owners.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:4327](../../../backend/migrations/001_init.sql#L4327) |
| `fk_medical_records_appointment_clinic` | `appointment_id, clinic_id` | [appointments](appointments.md) | `id, clinic_id` | `ON DELETE SET NULL (appointment_id)` | [001_init.sql:5822](../../../backend/migrations/001_init.sql#L5822) |
| `fk_medical_records_entered_by` | `entered_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [002_medical_records_entered_by_staff_fk.sql:10](../../../backend/migrations/002_medical_records_entered_by_staff_fk.sql#L10) |
| `fk_medical_records_doctor` | `doctor_id` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [017_multiclinic_staff_fk_fix.sql:28](../../../backend/migrations/017_multiclinic_staff_fk_fix.sql#L28) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medical_records_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medical_records_owner_id_fkey`: `REFERENCES owners(id) ON DELETE RESTRICT`
- `medical_records_pet_id_fkey`: `REFERENCES pets(id) ON DELETE RESTRICT`
- `fk_medical_records_clinic_pet`: `CONSTRAINT fk_medical_records_clinic_pet     FOREIGN KEY (clinic_id, pet_id)     REFERENCES pets (clinic_id, id)     ON DELETE RESTRICT`
- `fk_medical_records_clinic_owner`: `CONSTRAINT fk_medical_records_clinic_owner     FOREIGN KEY (clinic_id, owner_id)     REFERENCES owners (clinic_id, id)     ON DELETE RESTRICT`
- `fk_medical_records_appointment_clinic`: `CONSTRAINT fk_medical_records_appointment_clinic         FOREIGN KEY (appointment_id, clinic_id)         REFERENCES appointments (id, clinic_id)         ON DELETE SET NULL (appointment_id)`
- `fk_medical_records_entered_by`: `CONSTRAINT fk_medical_records_entered_by         FOREIGN KEY (entered_by)         REFERENCES staffs (id)         ON DELETE RESTRICT`
- `fk_medical_records_doctor`: `CONSTRAINT fk_medical_records_doctor         FOREIGN KEY (doctor_id) REFERENCES staffs (id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medical_records_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1413](../../../backend/migrations/001_init.sql#L1413) |
| `uq_medical_records_id_clinic` | unique | `CONSTRAINT uq_medical_records_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:4286](../../../backend/migrations/001_init.sql#L4286) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_medical_records_clinic_record_no` | YES | — | `CREATE UNIQUE INDEX idx_medical_records_clinic_record_no ON medical_records(clinic_id, record_no)` | [001_init.sql:2185](../../../backend/migrations/001_init.sql#L2185) |
| `idx_medical_records_clinic_date` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medical_records_clinic_date   ON medical_records(clinic_id, date DESC)   WHERE deleted_at IS NULL` | [001_init.sql:2357](../../../backend/migrations/001_init.sql#L2357) |
| `idx_medical_records_clinic_pet` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medical_records_clinic_pet   ON medical_records(clinic_id, pet_id)   WHERE deleted_at IS NULL` | [001_init.sql:2361](../../../backend/migrations/001_init.sql#L2361) |
| `idx_medical_records_clinic_owner` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medical_records_clinic_owner   ON medical_records(clinic_id, owner_id)   WHERE deleted_at IS NULL` | [001_init.sql:2365](../../../backend/migrations/001_init.sql#L2365) |
| `idx_medical_records_clinic_status` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medical_records_clinic_status   ON medical_records(clinic_id, status)   WHERE deleted_at IS NULL` | [001_init.sql:2369](../../../backend/migrations/001_init.sql#L2369) |
| `idx_medical_records_doctor_id` | NO | `doctor_id IS NOT NULL` | `CREATE INDEX idx_medical_records_doctor_id ON medical_records(doctor_id) WHERE doctor_id IS NOT NULL` | [001_init.sql:2432](../../../backend/migrations/001_init.sql#L2432) |
| `idx_medical_records_entered_by` | NO | `entered_by IS NOT NULL` | `CREATE INDEX idx_medical_records_entered_by ON medical_records(entered_by) WHERE entered_by IS NOT NULL` | [001_init.sql:2433](../../../backend/migrations/001_init.sql#L2433) |
| `idx_medical_records_appointment_id` | NO | `appointment_id IS NOT NULL` | `CREATE INDEX IF NOT EXISTS idx_medical_records_appointment_id     ON medical_records (appointment_id)     WHERE appointment_id IS NOT NULL` | [001_init.sql:3878](../../../backend/migrations/001_init.sql#L3878) |
| `uq_medical_records_clinic_appointment_active` | YES | `appointment_id IS NOT NULL AND deleted_at IS NULL` | `CREATE UNIQUE INDEX uq_medical_records_clinic_appointment_active     ON medical_records (clinic_id, appointment_id)     WHERE appointment_id IS NOT NULL AND deleted_at IS NULL` | [001_init.sql:3882](../../../backend/migrations/001_init.sql#L3882) |
| `idx_medical_records_record_no_trgm` | NO | — | `CREATE INDEX idx_medical_records_record_no_trgm   ON medical_records USING gin (record_no gin_trgm_ops)` | [015_search_fold_indexes.sql:57](../../../backend/migrations/015_search_fold_indexes.sql#L57) |
