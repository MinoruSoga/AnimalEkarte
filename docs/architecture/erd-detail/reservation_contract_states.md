<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `reservation_contract_states`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8914](../../../backend/migrations/001_init.sql#L8914) |
| `contract_kind` | `text` | NOT NULL | — | PK | [001_init.sql:8915](../../../backend/migrations/001_init.sql#L8915) |
| `revision` | `bigint` | NOT NULL | — | — | [001_init.sql:8916](../../../backend/migrations/001_init.sql#L8916) |
| `cutover_state` | `text` | NOT NULL | — | — | [001_init.sql:8917](../../../backend/migrations/001_init.sql#L8917) |
| `applicable_kinds` | `text[]` | NOT NULL | — | — | [001_init.sql:8918](../../../backend/migrations/001_init.sql#L8918) |
| `updated_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8919](../../../backend/migrations/001_init.sql#L8919) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    reservation_contract_states {
        bigint clinic_id PK,FK
        text contract_kind PK
        bigint revision
        text cutover_state
        text[] applicable_kinds
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||--o{ reservation_contract_states : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `reservation_contract_states_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8914](../../../backend/migrations/001_init.sql#L8914) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `reservation_contract_states_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `reservation_contract_states_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:8916](../../../backend/migrations/001_init.sql#L8916) |
| `reservation_contract_states_cutover_state_check` | check | `CHECK (cutover_state IN ('blocked','enabled'))` | [001_init.sql:8917](../../../backend/migrations/001_init.sql#L8917) |
| `reservation_contract_states_pkey` | primary | `PRIMARY KEY (clinic_id,contract_kind)` | [001_init.sql:8920](../../../backend/migrations/001_init.sql#L8920) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
