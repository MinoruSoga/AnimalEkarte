<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `owner_line_binding_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10151](../../../backend/migrations/001_init.sql#L10151) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10151](../../../backend/migrations/001_init.sql#L10151) |
| `owner_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10151](../../../backend/migrations/001_init.sql#L10151) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10152](../../../backend/migrations/001_init.sql#L10152) |
| `intent` | `text` | NOT NULL | — | — | [001_init.sql:10152](../../../backend/migrations/001_init.sql#L10152) |
| `before_state` | `jsonb` | NOT NULL | — | — | [001_init.sql:10153](../../../backend/migrations/001_init.sql#L10153) |
| `after_state` | `jsonb` | NOT NULL | — | — | [001_init.sql:10153](../../../backend/migrations/001_init.sql#L10153) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10154](../../../backend/migrations/001_init.sql#L10154) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10154](../../../backend/migrations/001_init.sql#L10154) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    owner_line_binding_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint owner_id FK,UK
        bigint revision UK
        text intent
        jsonb before_state
        jsonb after_state
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
    owner_line_binding_states {
        bigint clinic_id PK,FK
        bigint owner_id PK,FK
    }
    clinics ||..o{ owner_line_binding_events : "clinic_id"
    owner_line_binding_states ||..o{ owner_line_binding_events : "clinic_id, owner_id"
    operation_bindings ||..o{ owner_line_binding_events : "operation_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `owner_line_binding_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10151](../../../backend/migrations/001_init.sql#L10151) |
| `owner_line_binding_events_clinic_id_owner_id_fkey` | `clinic_id, owner_id` | [owner_line_binding_states](owner_line_binding_states.md) | `clinic_id, owner_id` | `ON DELETE RESTRICT` | [001_init.sql:10156](../../../backend/migrations/001_init.sql#L10156) |
| `owner_line_binding_events_operation_id_clinic_id_fkey` | `operation_id, clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10157](../../../backend/migrations/001_init.sql#L10157) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `owner_line_binding_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `owner_line_binding_events_clinic_id_owner_id_fkey`: `FOREIGN KEY(clinic_id,owner_id) REFERENCES owner_line_binding_states(clinic_id,owner_id) ON DELETE RESTRICT`
- `owner_line_binding_events_operation_id_clinic_id_fkey`: `FOREIGN KEY(operation_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `owner_line_binding_events_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10151](../../../backend/migrations/001_init.sql#L10151) |
| `owner_line_binding_events_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10152](../../../backend/migrations/001_init.sql#L10152) |
| `owner_line_binding_events_intent_check` | check | `CHECK(intent IN('link','unlink_request','unlink_complete','verify'))` | [001_init.sql:10152](../../../backend/migrations/001_init.sql#L10152) |
| `owner_line_binding_events_before_state_check` | check | `CHECK(jsonb_typeof(before_state)='object')` | [001_init.sql:10153](../../../backend/migrations/001_init.sql#L10153) |
| `owner_line_binding_events_after_state_check` | check | `CHECK(jsonb_typeof(after_state)='object')` | [001_init.sql:10153](../../../backend/migrations/001_init.sql#L10153) |
| `owner_line_binding_events_id_clinic_id_key` | unique | `UNIQUE(id,clinic_id)` | [001_init.sql:10155](../../../backend/migrations/001_init.sql#L10155) |
| `owner_line_binding_events_clinic_id_owner_id_revision_key` | unique | `UNIQUE(clinic_id,owner_id,revision)` | [001_init.sql:10155](../../../backend/migrations/001_init.sql#L10155) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
