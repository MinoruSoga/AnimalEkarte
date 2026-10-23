<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `estimate_successor_links`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9556](../../../backend/migrations/001_init.sql#L9556) |
| `original_estimate_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9557](../../../backend/migrations/001_init.sql#L9557) |
| `presentation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9558](../../../backend/migrations/001_init.sql#L9558) |
| `successor_estimate_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9559](../../../backend/migrations/001_init.sql#L9559) |
| `reason` | `text` | NOT NULL | — | — | [001_init.sql:9560](../../../backend/migrations/001_init.sql#L9560) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:9561](../../../backend/migrations/001_init.sql#L9561) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    estimate_successor_links {
        bigint clinic_id PK,FK
        bigint original_estimate_id FK
        bigint presentation_id FK
        bigint successor_estimate_id PK,FK
        text reason
        bigint operation_id FK
    }
    clinics {
        BIGSERIAL id PK
    }
    estimate_presentations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint estimate_id FK,UK
    }
    estimates {
        BIGSERIAL id PK,FK,UK
        bigint clinic_id FK,UK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ estimate_successor_links : "clinic_id"
    estimate_presentations ||..o{ estimate_successor_links : "presentation_id, clinic_id, original_estimate_id"
    estimates ||--o| estimate_successor_links : "successor_estimate_id, clinic_id"
    operation_bindings ||..o{ estimate_successor_links : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `estimate_successor_links_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9556](../../../backend/migrations/001_init.sql#L9556) |
| `estimate_successor_links_presentation_id_clinic_id_origina_fkey` | `presentation_id, clinic_id, original_estimate_id` | [estimate_presentations](estimate_presentations.md) | `id, clinic_id, estimate_id` | `ON DELETE RESTRICT` | [001_init.sql:9564](../../../backend/migrations/001_init.sql#L9564) |
| `estimate_successor_links_successor_estimate_id_clinic_id_fkey` | `successor_estimate_id, clinic_id` | [estimates](estimates.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9565](../../../backend/migrations/001_init.sql#L9565) |
| `estimate_successor_links_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9566](../../../backend/migrations/001_init.sql#L9566) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `estimate_successor_links_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `estimate_successor_links_presentation_id_clinic_id_origina_fkey`: `FOREIGN KEY (presentation_id,clinic_id,original_estimate_id) REFERENCES estimate_presentations(id,clinic_id,estimate_id) ON DELETE RESTRICT`
- `estimate_successor_links_successor_estimate_id_clinic_id_fkey`: `FOREIGN KEY (successor_estimate_id,clinic_id) REFERENCES estimates(id,clinic_id) ON DELETE RESTRICT`
- `estimate_successor_links_operation_id_clinic_id_fkey`: `FOREIGN KEY (operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `estimate_successor_links_reason_check` | check | `CHECK (btrim(reason)<>'')` | [001_init.sql:9560](../../../backend/migrations/001_init.sql#L9560) |
| `estimate_successor_links_pkey` | primary | `PRIMARY KEY (clinic_id,successor_estimate_id)` | [001_init.sql:9562](../../../backend/migrations/001_init.sql#L9562) |
| `estimate_successor_links_check` | check | `CHECK (original_estimate_id<>successor_estimate_id)` | [001_init.sql:9563](../../../backend/migrations/001_init.sql#L9563) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
