<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `cash_adjustment_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9822](../../../backend/migrations/001_init.sql#L9822) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9823](../../../backend/migrations/001_init.sql#L9823) |
| `event_kind` | `text` | NOT NULL | — | — | [001_init.sql:9824](../../../backend/migrations/001_init.sql#L9824) |
| `amount` | `bigint` | NOT NULL | — | — | [001_init.sql:9825](../../../backend/migrations/001_init.sql#L9825) |
| `occurred_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:9826](../../../backend/migrations/001_init.sql#L9826) |
| `reason` | `text` | NOT NULL | — | — | [001_init.sql:9827](../../../backend/migrations/001_init.sql#L9827) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9828](../../../backend/migrations/001_init.sql#L9828) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9829](../../../backend/migrations/001_init.sql#L9829) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9830](../../../backend/migrations/001_init.sql#L9830) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    cash_adjustment_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text event_kind
        bigint amount
        timestamptz occurred_at
        text reason
        bigint operation_id FK
        bigint recorded_by FK
        timestamptz recorded_at
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
    clinics ||..o{ cash_adjustment_events : "clinic_id"
    staffs ||..o{ cash_adjustment_events : "recorded_by"
    operation_bindings ||..o{ cash_adjustment_events : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `cash_adjustment_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9823](../../../backend/migrations/001_init.sql#L9823) |
| `cash_adjustment_events_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9829](../../../backend/migrations/001_init.sql#L9829) |
| `fk_cash_adjustment_operation` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9833](../../../backend/migrations/001_init.sql#L9833) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `cash_adjustment_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `cash_adjustment_events_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `fk_cash_adjustment_operation`: `CONSTRAINT fk_cash_adjustment_operation  FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `cash_adjustment_events_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9822](../../../backend/migrations/001_init.sql#L9822) |
| `cash_adjustment_events_event_kind_check` | check | `CHECK (event_kind IN ('replenish','collect'))` | [001_init.sql:9824](../../../backend/migrations/001_init.sql#L9824) |
| `cash_adjustment_events_amount_check` | check | `CHECK (amount > 0)` | [001_init.sql:9825](../../../backend/migrations/001_init.sql#L9825) |
| `cash_adjustment_events_reason_check` | check | `CHECK (btrim(reason)<>'')` | [001_init.sql:9827](../../../backend/migrations/001_init.sql#L9827) |
| `cash_adjustment_events_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:9831](../../../backend/migrations/001_init.sql#L9831) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
