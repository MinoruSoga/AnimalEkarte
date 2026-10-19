<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `record_edit_right_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:9002](../../../backend/migrations/001_init.sql#L9002) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9003](../../../backend/migrations/001_init.sql#L9003) |
| `medical_record_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9004](../../../backend/migrations/001_init.sql#L9004) |
| `previous_epoch` | `bigint` | NOT NULL | — | — | [001_init.sql:9005](../../../backend/migrations/001_init.sql#L9005) |
| `resulting_epoch` | `bigint` | NOT NULL | — | UK | [001_init.sql:9006](../../../backend/migrations/001_init.sql#L9006) |
| `intent` | `text` | NOT NULL | — | — | [001_init.sql:9007](../../../backend/migrations/001_init.sql#L9007) |
| `prior_holder_staff_id` | `bigint` | NULL | — | FK | [001_init.sql:9008](../../../backend/migrations/001_init.sql#L9008) |
| `new_holder_staff_id` | `bigint` | NULL | — | FK | [001_init.sql:9009](../../../backend/migrations/001_init.sql#L9009) |
| `session_ref` | `uuid` | NULL | — | — | [001_init.sql:9010](../../../backend/migrations/001_init.sql#L9010) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9011](../../../backend/migrations/001_init.sql#L9011) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9012](../../../backend/migrations/001_init.sql#L9012) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    record_edit_right_events {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        bigint medical_record_id FK,UK
        bigint previous_epoch
        bigint resulting_epoch UK
        text intent
        bigint prior_holder_staff_id FK
        bigint new_holder_staff_id FK
        uuid session_ref
        bigint operation_id FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    record_edit_rights {
        bigint clinic_id PK,FK
        bigint medical_record_id PK,FK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ record_edit_right_events : "clinic_id"
    staffs o|..o{ record_edit_right_events : "prior_holder_staff_id"
    staffs o|..o{ record_edit_right_events : "new_holder_staff_id"
    record_edit_rights ||..o{ record_edit_right_events : "clinic_id, medical_record_id"
    operation_bindings ||..o{ record_edit_right_events : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `record_edit_right_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9003](../../../backend/migrations/001_init.sql#L9003) |
| `record_edit_right_events_prior_holder_staff_id_fkey` | `prior_holder_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9008](../../../backend/migrations/001_init.sql#L9008) |
| `record_edit_right_events_new_holder_staff_id_fkey` | `new_holder_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9009](../../../backend/migrations/001_init.sql#L9009) |
| `record_edit_right_events_clinic_id_medical_record_id_fkey` | `clinic_id, medical_record_id` | [record_edit_rights](record_edit_rights.md) | `clinic_id, medical_record_id` | `ON DELETE RESTRICT` | [001_init.sql:9014](../../../backend/migrations/001_init.sql#L9014) |
| `record_edit_right_events_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9015](../../../backend/migrations/001_init.sql#L9015) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `record_edit_right_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `record_edit_right_events_prior_holder_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `record_edit_right_events_new_holder_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `record_edit_right_events_clinic_id_medical_record_id_fkey`: `FOREIGN KEY (clinic_id,medical_record_id) REFERENCES record_edit_rights(clinic_id,medical_record_id) ON DELETE RESTRICT`
- `record_edit_right_events_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `record_edit_right_events_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9002](../../../backend/migrations/001_init.sql#L9002) |
| `record_edit_right_events_previous_epoch_check` | check | `CHECK (previous_epoch > 0)` | [001_init.sql:9005](../../../backend/migrations/001_init.sql#L9005) |
| `record_edit_right_events_check` | check | `CHECK (resulting_epoch > previous_epoch)` | [001_init.sql:9006](../../../backend/migrations/001_init.sql#L9006) |
| `record_edit_right_events_intent_check` | check | `CHECK (intent IN ('acquire','release','force_acquire','disconnect'))` | [001_init.sql:9007](../../../backend/migrations/001_init.sql#L9007) |
| `record_edit_right_events_clinic_id_medical_record_id_result_key` | unique | `UNIQUE (clinic_id,medical_record_id,resulting_epoch)` | [001_init.sql:9013](../../../backend/migrations/001_init.sql#L9013) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
