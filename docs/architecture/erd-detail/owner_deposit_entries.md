<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `owner_deposit_entries`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:8437](../../../backend/migrations/001_init.sql#L8437) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8438](../../../backend/migrations/001_init.sql#L8438) |
| `owner_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8439](../../../backend/migrations/001_init.sql#L8439) |
| `operation_key` | `uuid` | NOT NULL | — | FK, UK | [001_init.sql:8440](../../../backend/migrations/001_init.sql#L8440) |
| `ordinal` | `integer` | NOT NULL | — | UK | [001_init.sql:8441](../../../backend/migrations/001_init.sql#L8441) |
| `account_revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:8442](../../../backend/migrations/001_init.sql#L8442) |
| `entry_kind` | `text` | NOT NULL | — | — | [001_init.sql:8443](../../../backend/migrations/001_init.sql#L8443) |
| `amount` | `bigint` | NOT NULL | — | — | [001_init.sql:8444](../../../backend/migrations/001_init.sql#L8444) |
| `source_entry_id` | `bigint` | NULL | — | FK | [001_init.sql:8445](../../../backend/migrations/001_init.sql#L8445) |
| `billing_id` | `bigint` | NULL | — | FK | [001_init.sql:8446](../../../backend/migrations/001_init.sql#L8446) |
| `appointment_id` | `bigint` | NULL | — | FK | [001_init.sql:8448](../../../backend/migrations/001_init.sql#L8448) |
| `occurred_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:8449](../../../backend/migrations/001_init.sql#L8449) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8450](../../../backend/migrations/001_init.sql#L8450) |
| `fund_record_correction_id` | `bigint` | NULL | — | FK | [001_init.sql:11165](../../../backend/migrations/001_init.sql#L11165) |
| `correction_reason` | `text` | NULL | — | — | [001_init.sql:11166](../../../backend/migrations/001_init.sql#L11166) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    owner_deposit_entries {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint owner_id FK,UK
        uuid operation_key FK,UK
        integer ordinal UK
        bigint account_revision UK
        text entry_kind
        bigint amount
        bigint source_entry_id FK
        bigint billing_id FK
        bigint appointment_id FK
        timestamptz occurred_at
        timestamptz recorded_at
        bigint fund_record_correction_id FK
        text correction_reason
    }
    appointments {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    billings {
        BIGSERIAL id PK,FK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    fund_record_corrections {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    owner_deposit_operations {
        bigint clinic_id PK,FK
        bigint owner_id PK,FK
        uuid operation_key PK
    }
    clinics ||..o{ owner_deposit_entries : "clinic_id"
    owner_deposit_operations ||..o{ owner_deposit_entries : "clinic_id, owner_id, operation_key"
    owner_deposit_entries o|..o{ owner_deposit_entries : "source_entry_id, clinic_id, owner_id"
    billings o|..o{ owner_deposit_entries : "billing_id, clinic_id"
    appointments o|..o{ owner_deposit_entries : "appointment_id, clinic_id"
    fund_record_corrections o|..o{ owner_deposit_entries : "fund_record_correction_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `owner_deposit_entries_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8438](../../../backend/migrations/001_init.sql#L8438) |
| `owner_deposit_entries_clinic_id_owner_id_operation_key_fkey` | `clinic_id, owner_id, operation_key` | [owner_deposit_operations](owner_deposit_operations.md) | `clinic_id, owner_id, operation_key` | `ON DELETE RESTRICT` | [001_init.sql:8458](../../../backend/migrations/001_init.sql#L8458) |
| `owner_deposit_entries_source_entry_id_clinic_id_owner_id_fkey` | `source_entry_id, clinic_id, owner_id` | [owner_deposit_entries](owner_deposit_entries.md) | `id, clinic_id, owner_id` | `ON DELETE RESTRICT` | [001_init.sql:8459](../../../backend/migrations/001_init.sql#L8459) |
| `owner_deposit_entries_billing_id_clinic_id_fkey` | `billing_id, clinic_id` | [billings](billings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8460](../../../backend/migrations/001_init.sql#L8460) |
| `owner_deposit_entries_appointment_id_clinic_id_fkey` | `appointment_id, clinic_id` | [appointments](appointments.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8462](../../../backend/migrations/001_init.sql#L8462) |
| `fk_deposit_record_correction` | `fund_record_correction_id, clinic_id` | [fund_record_corrections](fund_record_corrections.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:11167](../../../backend/migrations/001_init.sql#L11167) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `owner_deposit_entries_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `owner_deposit_entries_clinic_id_owner_id_operation_key_fkey`: `FOREIGN KEY (clinic_id,owner_id,operation_key) REFERENCES owner_deposit_operations(clinic_id,owner_id,operation_key) ON DELETE RESTRICT`
- `owner_deposit_entries_source_entry_id_clinic_id_owner_id_fkey`: `FOREIGN KEY (source_entry_id,clinic_id,owner_id) REFERENCES owner_deposit_entries(id,clinic_id,owner_id) ON DELETE RESTRICT`
- `owner_deposit_entries_billing_id_clinic_id_fkey`: `FOREIGN KEY (billing_id,clinic_id) REFERENCES billings(id,clinic_id) ON DELETE RESTRICT`
- `owner_deposit_entries_appointment_id_clinic_id_fkey`: `FOREIGN KEY (appointment_id,clinic_id) REFERENCES appointments(id,clinic_id) ON DELETE RESTRICT`
- `fk_deposit_record_correction`: `CONSTRAINT fk_deposit_record_correction FOREIGN KEY(fund_record_correction_id,clinic_id) REFERENCES fund_record_corrections(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `owner_deposit_entries_pkey` | primary | `PRIMARY KEY` | [001_init.sql:8437](../../../backend/migrations/001_init.sql#L8437) |
| `owner_deposit_entries_ordinal_check` | check | `CHECK (ordinal > 0)` | [001_init.sql:8441](../../../backend/migrations/001_init.sql#L8441) |
| `owner_deposit_entries_account_revision_check` | check | `CHECK (account_revision > 1)` | [001_init.sql:8442](../../../backend/migrations/001_init.sql#L8442) |
| `owner_deposit_entries_entry_kind_check` | check | `CHECK (entry_kind IN ('receive','apply','release','refund','correction_credit','correction_debit'))` | [001_init.sql:8443](../../../backend/migrations/001_init.sql#L8443) |
| `owner_deposit_entries_amount_check` | check | `CHECK (amount > 0)` | [001_init.sql:8444](../../../backend/migrations/001_init.sql#L8444) |
| `owner_deposit_entries_id_clinic_id_owner_id_key` | unique | `UNIQUE (id,clinic_id,owner_id)` | [001_init.sql:8451](../../../backend/migrations/001_init.sql#L8451) |
| `owner_deposit_entries_clinic_id_owner_id_operation_key_ordi_key` | unique | `UNIQUE (clinic_id,owner_id,operation_key,ordinal)` | [001_init.sql:8452](../../../backend/migrations/001_init.sql#L8452) |
| `owner_deposit_entries_clinic_id_owner_id_account_revision_key` | unique | `UNIQUE (clinic_id,owner_id,account_revision)` | [001_init.sql:8453](../../../backend/migrations/001_init.sql#L8453) |
| `uq_deposit_entry_id_clinic` | unique | `CONSTRAINT uq_deposit_entry_id_clinic UNIQUE(id,clinic_id)` | [001_init.sql:10120](../../../backend/migrations/001_init.sql#L10120) |
| `ck_deposit_target` | check | `CONSTRAINT ck_deposit_target CHECK(  (entry_kind='receive' AND source_entry_id IS NULL AND billing_id IS NULL)  OR(entry_kind IN('apply','release') AND source_entry_id IS NOT NULL AND billing_id IS NOT NULL AND appointment_id IS NULL)  OR(entry_kind='refund' AND source_entry_id IS NOT NULL AND billing_id IS NULL AND appointment_id IS NULL)  OR(entry_kind IN('correction_credit','correction_debit') AND source_entry_id IS NOT NULL AND billing_id IS NULL AND appointment_id IS NULL))` | [001_init.sql:10956](../../../backend/migrations/001_init.sql#L10956) |
| `ck_deposit_correction_source` | check | `CONSTRAINT ck_deposit_correction_source CHECK(  (entry_kind IN('correction_credit','correction_debit') AND fund_record_correction_id IS NOT NULL AND correction_reason IS NOT NULL AND btrim(correction_reason)<>'')  OR(entry_kind NOT IN('correction_credit','correction_debit') AND fund_record_correction_id IS NULL AND correction_reason IS NULL))` | [001_init.sql:11168](../../../backend/migrations/001_init.sql#L11168) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_deposit_entries_source` | NO | — | `CREATE INDEX idx_deposit_entries_source ON owner_deposit_entries(clinic_id,owner_id,source_entry_id)` | [001_init.sql:8464](../../../backend/migrations/001_init.sql#L8464) |
| `idx_deposit_entries_billing` | NO | `billing_id IS NOT NULL` | `CREATE INDEX idx_deposit_entries_billing ON owner_deposit_entries(clinic_id,billing_id,id) WHERE billing_id IS NOT NULL` | [001_init.sql:8465](../../../backend/migrations/001_init.sql#L8465) |
