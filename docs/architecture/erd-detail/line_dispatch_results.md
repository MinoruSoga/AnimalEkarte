<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `line_dispatch_results`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10204](../../../backend/migrations/001_init.sql#L10204) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10204](../../../backend/migrations/001_init.sql#L10204) |
| `attempt_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10204](../../../backend/migrations/001_init.sql#L10204) |
| `intent_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10204](../../../backend/migrations/001_init.sql#L10204) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10205](../../../backend/migrations/001_init.sql#L10205) |
| `state` | `text` | NOT NULL | — | — | [001_init.sql:10205](../../../backend/migrations/001_init.sql#L10205) |
| `external_id` | `text` | NULL | — | — | [001_init.sql:10206](../../../backend/migrations/001_init.sql#L10206) |
| `evidence` | `jsonb` | NOT NULL | — | — | [001_init.sql:10206](../../../backend/migrations/001_init.sql#L10206) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10206](../../../backend/migrations/001_init.sql#L10206) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    line_dispatch_results {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint attempt_id FK,UK
        bigint intent_id FK,UK
        bigint revision UK
        text state
        text external_id
        jsonb evidence
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    line_dispatch_attempts {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint intent_id FK,UK
    }
    clinics ||..o{ line_dispatch_results : "clinic_id"
    line_dispatch_attempts ||..o{ line_dispatch_results : "attempt_id, clinic_id, intent_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `line_dispatch_results_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10204](../../../backend/migrations/001_init.sql#L10204) |
| `line_dispatch_results_attempt_id_clinic_id_intent_id_fkey` | `attempt_id, clinic_id, intent_id` | [line_dispatch_attempts](line_dispatch_attempts.md) | `id, clinic_id, intent_id` | `ON DELETE RESTRICT` | [001_init.sql:10209](../../../backend/migrations/001_init.sql#L10209) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `line_dispatch_results_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `line_dispatch_results_attempt_id_clinic_id_intent_id_fkey`: `FOREIGN KEY(attempt_id,clinic_id,intent_id) REFERENCES line_dispatch_attempts(id,clinic_id,intent_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `line_dispatch_results_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10204](../../../backend/migrations/001_init.sql#L10204) |
| `line_dispatch_results_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10205](../../../backend/migrations/001_init.sql#L10205) |
| `line_dispatch_results_state_check` | check | `CHECK(state IN('succeeded','definite_failure','unknown','verified_absent'))` | [001_init.sql:10205](../../../backend/migrations/001_init.sql#L10205) |
| `line_dispatch_results_evidence_check` | check | `CHECK(jsonb_typeof(evidence)='object')` | [001_init.sql:10206](../../../backend/migrations/001_init.sql#L10206) |
| `line_dispatch_results_id_clinic_id_attempt_id_intent_id_key` | unique | `UNIQUE(id,clinic_id,attempt_id,intent_id)` | [001_init.sql:10207](../../../backend/migrations/001_init.sql#L10207) |
| `line_dispatch_results_clinic_id_attempt_id_revision_key` | unique | `UNIQUE(clinic_id,attempt_id,revision)` | [001_init.sql:10207](../../../backend/migrations/001_init.sql#L10207) |
| `line_dispatch_results_check` | check | `CHECK(state<>'succeeded' OR external_id IS NOT NULL)` | [001_init.sql:10208](../../../backend/migrations/001_init.sql#L10208) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
