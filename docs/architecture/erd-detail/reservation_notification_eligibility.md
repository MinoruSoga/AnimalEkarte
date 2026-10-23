<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `reservation_notification_eligibility`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:10172](../../../backend/migrations/001_init.sql#L10172) |
| `appointment_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:10172](../../../backend/migrations/001_init.sql#L10172) |
| `owner_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10172](../../../backend/migrations/001_init.sql#L10172) |
| `binding_revision` | `bigint` | NOT NULL | — | — | [001_init.sql:10173](../../../backend/migrations/001_init.sql#L10173) |
| `appointment_version` | `bigint` | NOT NULL | — | — | [001_init.sql:10173](../../../backend/migrations/001_init.sql#L10173) |
| `state` | `text` | NOT NULL | — | — | [001_init.sql:10174](../../../backend/migrations/001_init.sql#L10174) |
| `confirmed_by` | `bigint` | NULL | — | FK | [001_init.sql:10175](../../../backend/migrations/001_init.sql#L10175) |
| `confirmed_at` | `timestamptz` | NULL | — | — | [001_init.sql:10175](../../../backend/migrations/001_init.sql#L10175) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10175](../../../backend/migrations/001_init.sql#L10175) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    reservation_notification_eligibility {
        bigint clinic_id PK,FK
        bigint appointment_id PK,FK
        bigint owner_id FK
        bigint binding_revision
        bigint appointment_version
        text state
        bigint confirmed_by FK
        timestamptz confirmed_at
        bigint operation_id FK
    }
    appointments {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    owner_line_binding_states {
        bigint clinic_id PK,FK
        bigint owner_id PK,FK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||--o{ reservation_notification_eligibility : "clinic_id"
    staffs o|..o{ reservation_notification_eligibility : "confirmed_by"
    appointments ||--o| reservation_notification_eligibility : "appointment_id, clinic_id"
    owner_line_binding_states ||..o{ reservation_notification_eligibility : "clinic_id, owner_id"
    operation_bindings ||..o{ reservation_notification_eligibility : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `reservation_notification_eligibility_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10172](../../../backend/migrations/001_init.sql#L10172) |
| `reservation_notification_eligibility_confirmed_by_fkey` | `confirmed_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10175](../../../backend/migrations/001_init.sql#L10175) |
| `reservation_notification_eligibil_appointment_id_clinic_id_fkey` | `appointment_id, clinic_id` | [appointments](appointments.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10177](../../../backend/migrations/001_init.sql#L10177) |
| `reservation_notification_eligibility_clinic_id_owner_id_fkey` | `clinic_id, owner_id` | [owner_line_binding_states](owner_line_binding_states.md) | `clinic_id, owner_id` | `ON DELETE RESTRICT` | [001_init.sql:10178](../../../backend/migrations/001_init.sql#L10178) |
| `reservation_notification_eligibilit_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10179](../../../backend/migrations/001_init.sql#L10179) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `reservation_notification_eligibility_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `reservation_notification_eligibility_confirmed_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `reservation_notification_eligibil_appointment_id_clinic_id_fkey`: `FOREIGN KEY(appointment_id,clinic_id) REFERENCES appointments(id,clinic_id) ON DELETE RESTRICT`
- `reservation_notification_eligibility_clinic_id_owner_id_fkey`: `FOREIGN KEY(clinic_id,owner_id) REFERENCES owner_line_binding_states(clinic_id,owner_id) ON DELETE RESTRICT`
- `reservation_notification_eligibilit_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `reservation_notification_eligibility_binding_revision_check` | check | `CHECK(binding_revision>0)` | [001_init.sql:10173](../../../backend/migrations/001_init.sql#L10173) |
| `reservation_notification_eligibility_appointment_version_check` | check | `CHECK(appointment_version>0)` | [001_init.sql:10173](../../../backend/migrations/001_init.sql#L10173) |
| `reservation_notification_eligibility_state_check` | check | `CHECK(state IN('stopped','resume_pending','confirmed'))` | [001_init.sql:10174](../../../backend/migrations/001_init.sql#L10174) |
| `reservation_notification_eligibility_pkey` | primary | `PRIMARY KEY(clinic_id,appointment_id)` | [001_init.sql:10176](../../../backend/migrations/001_init.sql#L10176) |
| `reservation_notification_eligibility_check` | check | `CHECK(state<>'confirmed' OR(confirmed_by IS NOT NULL AND confirmed_at IS NOT NULL))` | [001_init.sql:10176](../../../backend/migrations/001_init.sql#L10176) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
