<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `cash_close_source_events`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK, UK | [001_init.sql:9912](../../../backend/migrations/001_init.sql#L9912) |
| `close_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9913](../../../backend/migrations/001_init.sql#L9913) |
| `source_event_id` | `bigint` | NOT NULL | — | PK, FK, UK | [001_init.sql:9914](../../../backend/migrations/001_init.sql#L9914) |
| `source_version` | `bigint` | NOT NULL | — | — | [001_init.sql:9915](../../../backend/migrations/001_init.sql#L9915) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    cash_close_source_events {
        bigint clinic_id PK,FK,UK
        bigint close_id PK,FK
        bigint source_event_id PK,FK,UK
        bigint source_version
    }
    cash_close_contexts {
        bigint clinic_id PK,FK
        bigint close_id PK,FK
    }
    clinics {
        BIGSERIAL id PK
    }
    finance_source_events {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ cash_close_source_events : "clinic_id"
    cash_close_contexts ||--o{ cash_close_source_events : "clinic_id, close_id"
    finance_source_events ||--o| cash_close_source_events : "source_event_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `cash_close_source_events_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9912](../../../backend/migrations/001_init.sql#L9912) |
| `cash_close_source_events_clinic_id_close_id_fkey` | `clinic_id, close_id` | [cash_close_contexts](cash_close_contexts.md) | `clinic_id, close_id` | `ON DELETE RESTRICT` | [001_init.sql:9918](../../../backend/migrations/001_init.sql#L9918) |
| `cash_close_source_events_source_event_id_clinic_id_fkey` | `source_event_id, clinic_id` | [finance_source_events](finance_source_events.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9919](../../../backend/migrations/001_init.sql#L9919) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `cash_close_source_events_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `cash_close_source_events_clinic_id_close_id_fkey`: `FOREIGN KEY (clinic_id,close_id) REFERENCES cash_close_contexts(clinic_id,close_id) ON DELETE RESTRICT`
- `cash_close_source_events_source_event_id_clinic_id_fkey`: `FOREIGN KEY (source_event_id,clinic_id) REFERENCES finance_source_events(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `cash_close_source_events_source_version_check` | check | `CHECK (source_version > 0)` | [001_init.sql:9915](../../../backend/migrations/001_init.sql#L9915) |
| `cash_close_source_events_pkey` | primary | `PRIMARY KEY (clinic_id,close_id,source_event_id)` | [001_init.sql:9916](../../../backend/migrations/001_init.sql#L9916) |
| `cash_close_source_events_clinic_id_source_event_id_key` | unique | `UNIQUE (clinic_id,source_event_id)` | [001_init.sql:9917](../../../backend/migrations/001_init.sql#L9917) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
