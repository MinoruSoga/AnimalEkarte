<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `operation_attempt_outcomes`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:10079](../../../backend/migrations/001_init.sql#L10079) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10079](../../../backend/migrations/001_init.sql#L10079) |
| `binding_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10080](../../../backend/migrations/001_init.sql#L10080) |
| `attempt_key` | `uuid` | NOT NULL | — | UK | [001_init.sql:10080](../../../backend/migrations/001_init.sql#L10080) |
| `outcome` | `text` | NOT NULL | — | — | [001_init.sql:10081](../../../backend/migrations/001_init.sql#L10081) |
| `details` | `jsonb` | NOT NULL | — | — | [001_init.sql:10082](../../../backend/migrations/001_init.sql#L10082) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10082](../../../backend/migrations/001_init.sql#L10082) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    operation_attempt_outcomes {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        bigint binding_id FK,UK
        uuid attempt_key UK
        text outcome
        jsonb details
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ operation_attempt_outcomes : "clinic_id"
    operation_bindings ||..o{ operation_attempt_outcomes : "binding_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `operation_attempt_outcomes_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10079](../../../backend/migrations/001_init.sql#L10079) |
| `operation_attempt_outcomes_binding_id_clinic_id_fkey` | `binding_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10084](../../../backend/migrations/001_init.sql#L10084) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `operation_attempt_outcomes_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `operation_attempt_outcomes_binding_id_clinic_id_fkey`: `FOREIGN KEY(binding_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `operation_attempt_outcomes_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10079](../../../backend/migrations/001_init.sql#L10079) |
| `operation_attempt_outcomes_outcome_check` | check | `CHECK(outcome IN('rejected','unknown','retryable_failure'))` | [001_init.sql:10081](../../../backend/migrations/001_init.sql#L10081) |
| `operation_attempt_outcomes_details_check` | check | `CHECK(jsonb_typeof(details)='object')` | [001_init.sql:10082](../../../backend/migrations/001_init.sql#L10082) |
| `operation_attempt_outcomes_clinic_id_binding_id_attempt_key_key` | unique | `UNIQUE(clinic_id,binding_id,attempt_key)` | [001_init.sql:10083](../../../backend/migrations/001_init.sql#L10083) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
