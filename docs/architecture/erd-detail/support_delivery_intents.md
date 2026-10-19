<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `support_delivery_intents`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10016](../../../backend/migrations/001_init.sql#L10016) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10016](../../../backend/migrations/001_init.sql#L10016) |
| `report_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10017](../../../backend/migrations/001_init.sql#L10017) |
| `report_version` | `bigint` | NOT NULL | — | — | [001_init.sql:10017](../../../backend/migrations/001_init.sql#L10017) |
| `operation_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10017](../../../backend/migrations/001_init.sql#L10017) |
| `confirmed_payload` | `jsonb` | NOT NULL | — | — | [001_init.sql:10018](../../../backend/migrations/001_init.sql#L10018) |
| `payload_digest` | `bytea` | NOT NULL | — | — | [001_init.sql:10018](../../../backend/migrations/001_init.sql#L10018) |
| `destination_count` | `integer` | NOT NULL | — | — | [001_init.sql:10019](../../../backend/migrations/001_init.sql#L10019) |
| `confirmed_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:10020](../../../backend/migrations/001_init.sql#L10020) |
| `confirmed_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:10020](../../../backend/migrations/001_init.sql#L10020) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10021](../../../backend/migrations/001_init.sql#L10021) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    support_delivery_intents {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint report_id FK
        bigint report_version
        bigint operation_id FK,UK
        jsonb confirmed_payload
        bytea payload_digest
        integer destination_count
        bigint confirmed_by FK
        timestamptz confirmed_at
        timestamptz created_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    support_bug_reports {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ support_delivery_intents : "clinic_id"
    staffs ||..o{ support_delivery_intents : "confirmed_by"
    support_bug_reports ||..o{ support_delivery_intents : "report_id, clinic_id"
    operation_bindings ||..o| support_delivery_intents : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `support_delivery_intents_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10016](../../../backend/migrations/001_init.sql#L10016) |
| `support_delivery_intents_confirmed_by_fkey` | `confirmed_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10020](../../../backend/migrations/001_init.sql#L10020) |
| `support_delivery_intents_report_id_clinic_id_fkey` | `report_id, clinic_id` | [support_bug_reports](support_bug_reports.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10022](../../../backend/migrations/001_init.sql#L10022) |
| `support_delivery_intents_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10023](../../../backend/migrations/001_init.sql#L10023) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `support_delivery_intents_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `support_delivery_intents_confirmed_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `support_delivery_intents_report_id_clinic_id_fkey`: `FOREIGN KEY(report_id,clinic_id) REFERENCES support_bug_reports(id,clinic_id) ON DELETE RESTRICT`
- `support_delivery_intents_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `support_delivery_intents_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10016](../../../backend/migrations/001_init.sql#L10016) |
| `support_delivery_intents_report_version_check` | check | `CHECK(report_version>0)` | [001_init.sql:10017](../../../backend/migrations/001_init.sql#L10017) |
| `support_delivery_intents_confirmed_payload_check` | check | `CHECK(jsonb_typeof(confirmed_payload)='object')` | [001_init.sql:10018](../../../backend/migrations/001_init.sql#L10018) |
| `support_delivery_intents_payload_digest_check` | check | `CHECK(octet_length(payload_digest)=32)` | [001_init.sql:10018](../../../backend/migrations/001_init.sql#L10018) |
| `support_delivery_intents_destination_count_check` | check | `CHECK(destination_count>0)` | [001_init.sql:10019](../../../backend/migrations/001_init.sql#L10019) |
| `support_delivery_intents_id_clinic_id_key` | unique | `UNIQUE(id,clinic_id)` | [001_init.sql:10021](../../../backend/migrations/001_init.sql#L10021) |
| `support_delivery_intents_clinic_id_operation_id_key` | unique | `UNIQUE(clinic_id,operation_id)` | [001_init.sql:10021](../../../backend/migrations/001_init.sql#L10021) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
