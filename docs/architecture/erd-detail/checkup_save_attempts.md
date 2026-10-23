<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `checkup_save_attempts`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9958](../../../backend/migrations/001_init.sql#L9958) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9958](../../../backend/migrations/001_init.sql#L9958) |
| `pet_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9959](../../../backend/migrations/001_init.sql#L9959) |
| `performed_date` | `date` | NOT NULL | — | — | [001_init.sql:9959](../../../backend/migrations/001_init.sql#L9959) |
| `operation_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9960](../../../backend/migrations/001_init.sql#L9960) |
| `request_fingerprint` | `bytea` | NOT NULL | — | — | [001_init.sql:9960](../../../backend/migrations/001_init.sql#L9960) |
| `actor_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9961](../../../backend/migrations/001_init.sql#L9961) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9961](../../../backend/migrations/001_init.sql#L9961) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    checkup_save_attempts {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint pet_id FK
        date performed_date
        bigint operation_id FK,UK
        bytea request_fingerprint
        bigint actor_id FK
        timestamptz created_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
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
    clinics ||..o{ checkup_save_attempts : "clinic_id"
    staffs ||..o{ checkup_save_attempts : "actor_id"
    pets ||..o{ checkup_save_attempts : "pet_id, clinic_id"
    operation_bindings ||..o| checkup_save_attempts : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `checkup_save_attempts_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9958](../../../backend/migrations/001_init.sql#L9958) |
| `checkup_save_attempts_actor_id_fkey` | `actor_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9961](../../../backend/migrations/001_init.sql#L9961) |
| `checkup_save_attempts_pet_id_clinic_id_fkey` | `pet_id, clinic_id` | [pets](pets.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9963](../../../backend/migrations/001_init.sql#L9963) |
| `checkup_save_attempts_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9964](../../../backend/migrations/001_init.sql#L9964) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `checkup_save_attempts_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `checkup_save_attempts_actor_id_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `checkup_save_attempts_pet_id_clinic_id_fkey`: `FOREIGN KEY(pet_id,clinic_id) REFERENCES pets(id,clinic_id) ON DELETE RESTRICT`
- `checkup_save_attempts_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `checkup_save_attempts_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9958](../../../backend/migrations/001_init.sql#L9958) |
| `checkup_save_attempts_request_fingerprint_check` | check | `CHECK(octet_length(request_fingerprint)=32)` | [001_init.sql:9960](../../../backend/migrations/001_init.sql#L9960) |
| `checkup_save_attempts_id_clinic_id_key` | unique | `UNIQUE(id,clinic_id)` | [001_init.sql:9962](../../../backend/migrations/001_init.sql#L9962) |
| `checkup_save_attempts_clinic_id_operation_id_key` | unique | `UNIQUE(clinic_id,operation_id)` | [001_init.sql:9962](../../../backend/migrations/001_init.sql#L9962) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
