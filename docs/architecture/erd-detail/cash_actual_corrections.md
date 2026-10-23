<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `cash_actual_corrections`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9920](../../../backend/migrations/001_init.sql#L9920) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9921](../../../backend/migrations/001_init.sql#L9921) |
| `close_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9922](../../../backend/migrations/001_init.sql#L9922) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:9923](../../../backend/migrations/001_init.sql#L9923) |
| `previous_correction_id` | `bigint` | NULL | — | FK | [001_init.sql:9924](../../../backend/migrations/001_init.sql#L9924) |
| `before_actual_amount` | `bigint` | NOT NULL | — | — | [001_init.sql:9925](../../../backend/migrations/001_init.sql#L9925) |
| `corrected_actual_amount` | `bigint` | NOT NULL | — | — | [001_init.sql:9926](../../../backend/migrations/001_init.sql#L9926) |
| `theoretical_basis_amount` | `bigint` | NOT NULL | — | — | [001_init.sql:9927](../../../backend/migrations/001_init.sql#L9927) |
| `source_generation` | `bigint` | NOT NULL | — | — | [001_init.sql:9928](../../../backend/migrations/001_init.sql#L9928) |
| `original_counted_at` | `timestamptz` | NULL | — | — | [001_init.sql:9929](../../../backend/migrations/001_init.sql#L9929) |
| `reason` | `text` | NOT NULL | — | — | [001_init.sql:9930](../../../backend/migrations/001_init.sql#L9930) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9931](../../../backend/migrations/001_init.sql#L9931) |
| `corrected_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9932](../../../backend/migrations/001_init.sql#L9932) |
| `corrected_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9933](../../../backend/migrations/001_init.sql#L9933) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    cash_actual_corrections {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint close_id FK,UK
        bigint revision UK
        bigint previous_correction_id FK
        bigint before_actual_amount
        bigint corrected_actual_amount
        bigint theoretical_basis_amount
        bigint source_generation
        timestamptz original_counted_at
        text reason
        bigint operation_id FK
        bigint corrected_by FK
        timestamptz corrected_at
    }
    cash_register_closes {
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
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ cash_actual_corrections : "clinic_id"
    staffs ||..o{ cash_actual_corrections : "corrected_by"
    cash_register_closes ||..o{ cash_actual_corrections : "close_id, clinic_id"
    cash_actual_corrections o|..o{ cash_actual_corrections : "previous_correction_id, clinic_id, close_id"
    operation_bindings ||..o{ cash_actual_corrections : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `cash_actual_corrections_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9921](../../../backend/migrations/001_init.sql#L9921) |
| `cash_actual_corrections_corrected_by_fkey` | `corrected_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9932](../../../backend/migrations/001_init.sql#L9932) |
| `cash_actual_corrections_close_id_clinic_id_fkey` | `close_id, clinic_id` | [cash_register_closes](cash_register_closes.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9936](../../../backend/migrations/001_init.sql#L9936) |
| `cash_actual_corrections_previous_correction_id_clinic_id_c_fkey` | `previous_correction_id, clinic_id, close_id` | [cash_actual_corrections](cash_actual_corrections.md) | `id, clinic_id, close_id` | `ON DELETE RESTRICT` | [001_init.sql:9937](../../../backend/migrations/001_init.sql#L9937) |
| `cash_actual_corrections_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9938](../../../backend/migrations/001_init.sql#L9938) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `cash_actual_corrections_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `cash_actual_corrections_corrected_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `cash_actual_corrections_close_id_clinic_id_fkey`: `FOREIGN KEY (close_id,clinic_id) REFERENCES cash_register_closes(id,clinic_id) ON DELETE RESTRICT`
- `cash_actual_corrections_previous_correction_id_clinic_id_c_fkey`: `FOREIGN KEY (previous_correction_id,clinic_id,close_id) REFERENCES cash_actual_corrections(id,clinic_id,close_id) ON DELETE RESTRICT`
- `cash_actual_corrections_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `cash_actual_corrections_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9920](../../../backend/migrations/001_init.sql#L9920) |
| `cash_actual_corrections_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:9923](../../../backend/migrations/001_init.sql#L9923) |
| `cash_actual_corrections_corrected_actual_amount_check` | check | `CHECK (corrected_actual_amount >= 0)` | [001_init.sql:9926](../../../backend/migrations/001_init.sql#L9926) |
| `cash_actual_corrections_source_generation_check` | check | `CHECK (source_generation > 0)` | [001_init.sql:9928](../../../backend/migrations/001_init.sql#L9928) |
| `cash_actual_corrections_reason_check` | check | `CHECK (btrim(reason)<>'')` | [001_init.sql:9930](../../../backend/migrations/001_init.sql#L9930) |
| `cash_actual_corrections_id_clinic_id_close_id_key` | unique | `UNIQUE (id,clinic_id,close_id)` | [001_init.sql:9934](../../../backend/migrations/001_init.sql#L9934) |
| `cash_actual_corrections_clinic_id_close_id_revision_key` | unique | `UNIQUE (clinic_id,close_id,revision)` | [001_init.sql:9935](../../../backend/migrations/001_init.sql#L9935) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
