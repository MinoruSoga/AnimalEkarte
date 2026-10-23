<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `support_delivery_attempts`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10045](../../../backend/migrations/001_init.sql#L10045) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10045](../../../backend/migrations/001_init.sql#L10045) |
| `destination_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10046](../../../backend/migrations/001_init.sql#L10046) |
| `attempt_key` | `uuid` | NOT NULL | — | UK | [001_init.sql:10046](../../../backend/migrations/001_init.sql#L10046) |
| `attempt_number` | `bigint` | NOT NULL | — | UK | [001_init.sql:10046](../../../backend/migrations/001_init.sql#L10046) |
| `previous_attempt_id` | `bigint` | NULL | — | FK | [001_init.sql:10047](../../../backend/migrations/001_init.sql#L10047) |
| `authorized_retry_result_id` | `bigint` | NULL | — | FK | [001_init.sql:10047](../../../backend/migrations/001_init.sql#L10047) |
| `claimed_by` | `text` | NOT NULL | — | — | [001_init.sql:10048](../../../backend/migrations/001_init.sql#L10048) |
| `claimed_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:10048](../../../backend/migrations/001_init.sql#L10048) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    support_delivery_attempts {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint destination_id FK,UK
        uuid attempt_key UK
        bigint attempt_number UK
        bigint previous_attempt_id FK
        bigint authorized_retry_result_id FK
        text claimed_by
        timestamptz claimed_at
    }
    clinics {
        BIGSERIAL id PK
    }
    support_delivery_destinations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    support_delivery_results {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint attempt_id FK,UK
        bigint destination_id FK,UK
    }
    clinics ||..o{ support_delivery_attempts : "clinic_id"
    support_delivery_destinations ||..o{ support_delivery_attempts : "destination_id, clinic_id"
    support_delivery_attempts o|..o{ support_delivery_attempts : "previous_attempt_id, clinic_id, destination_id"
    support_delivery_results o|..o{ support_delivery_attempts : "authorized_retry_result_id, clinic_id, previous_attempt_id, destination_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `support_delivery_attempts_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10045](../../../backend/migrations/001_init.sql#L10045) |
| `support_delivery_attempts_destination_id_clinic_id_fkey` | `destination_id, clinic_id` | [support_delivery_destinations](support_delivery_destinations.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10052](../../../backend/migrations/001_init.sql#L10052) |
| `support_delivery_attempts_previous_attempt_id_clinic_id_de_fkey` | `previous_attempt_id, clinic_id, destination_id` | [support_delivery_attempts](support_delivery_attempts.md) | `id, clinic_id, destination_id` | `ON DELETE RESTRICT` | [001_init.sql:10053](../../../backend/migrations/001_init.sql#L10053) |
| `fk_support_retry_evidence` | `authorized_retry_result_id, clinic_id, previous_attempt_id, destination_id` | [support_delivery_results](support_delivery_results.md) | `id, clinic_id, attempt_id, destination_id` | `ON DELETE RESTRICT` | [001_init.sql:10065](../../../backend/migrations/001_init.sql#L10065) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `support_delivery_attempts_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `support_delivery_attempts_destination_id_clinic_id_fkey`: `FOREIGN KEY(destination_id,clinic_id) REFERENCES support_delivery_destinations(id,clinic_id) ON DELETE RESTRICT`
- `support_delivery_attempts_previous_attempt_id_clinic_id_de_fkey`: `FOREIGN KEY(previous_attempt_id,clinic_id,destination_id) REFERENCES support_delivery_attempts(id,clinic_id,destination_id) ON DELETE RESTRICT`
- `fk_support_retry_evidence`: `CONSTRAINT fk_support_retry_evidence  FOREIGN KEY(authorized_retry_result_id,clinic_id,previous_attempt_id,destination_id)  REFERENCES support_delivery_results(id,clinic_id,attempt_id,destination_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `support_delivery_attempts_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10045](../../../backend/migrations/001_init.sql#L10045) |
| `support_delivery_attempts_attempt_number_check` | check | `CHECK(attempt_number>0)` | [001_init.sql:10046](../../../backend/migrations/001_init.sql#L10046) |
| `support_delivery_attempts_claimed_by_check` | check | `CHECK(btrim(claimed_by)<>'')` | [001_init.sql:10048](../../../backend/migrations/001_init.sql#L10048) |
| `support_delivery_attempts_id_clinic_id_destination_id_key` | unique | `UNIQUE(id,clinic_id,destination_id)` | [001_init.sql:10049](../../../backend/migrations/001_init.sql#L10049) |
| `support_delivery_attempts_clinic_id_attempt_key_key` | unique | `UNIQUE(clinic_id,attempt_key)` | [001_init.sql:10049](../../../backend/migrations/001_init.sql#L10049) |
| `support_delivery_attempts_clinic_id_destination_id_attempt__key` | unique | `UNIQUE(clinic_id,destination_id,attempt_number)` | [001_init.sql:10049](../../../backend/migrations/001_init.sql#L10049) |
| `support_delivery_attempts_check` | check | `CHECK((attempt_number=1 AND previous_attempt_id IS NULL AND authorized_retry_result_id IS NULL)  OR(attempt_number>1 AND previous_attempt_id IS NOT NULL AND authorized_retry_result_id IS NOT NULL))` | [001_init.sql:10050](../../../backend/migrations/001_init.sql#L10050) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
