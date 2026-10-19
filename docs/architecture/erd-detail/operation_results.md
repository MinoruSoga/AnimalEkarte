<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `operation_results`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8863](../../../backend/migrations/001_init.sql#L8863) |
| `binding_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8864](../../../backend/migrations/001_init.sql#L8864) |
| `result_refs` | `jsonb` | NOT NULL | — | — | [001_init.sql:8866](../../../backend/migrations/001_init.sql#L8866) |
| `committed_versions` | `jsonb` | NOT NULL | — | — | [001_init.sql:8867](../../../backend/migrations/001_init.sql#L8867) |
| `effect_summary` | `jsonb` | NOT NULL | — | — | [001_init.sql:8868](../../../backend/migrations/001_init.sql#L8868) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8869](../../../backend/migrations/001_init.sql#L8869) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    operation_results {
        bigint clinic_id PK,FK
        bigint binding_id PK,FK
        jsonb result_refs
        jsonb committed_versions
        jsonb effect_summary
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ operation_results : "clinic_id"
    operation_bindings ||--o| operation_results : "binding_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `operation_results_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8863](../../../backend/migrations/001_init.sql#L8863) |
| `operation_results_binding_id_clinic_id_fkey` | `binding_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8871](../../../backend/migrations/001_init.sql#L8871) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `operation_results_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `operation_results_binding_id_clinic_id_fkey`: `FOREIGN KEY (binding_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `operation_results_result_refs_check` | check | `CHECK (jsonb_typeof(result_refs)='object')` | [001_init.sql:8866](../../../backend/migrations/001_init.sql#L8866) |
| `operation_results_committed_versions_check` | check | `CHECK (jsonb_typeof(committed_versions)='object')` | [001_init.sql:8867](../../../backend/migrations/001_init.sql#L8867) |
| `operation_results_effect_summary_check` | check | `CHECK (jsonb_typeof(effect_summary)='object')` | [001_init.sql:8868](../../../backend/migrations/001_init.sql#L8868) |
| `operation_results_pkey` | primary | `PRIMARY KEY (clinic_id,binding_id)` | [001_init.sql:8870](../../../backend/migrations/001_init.sql#L8870) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
