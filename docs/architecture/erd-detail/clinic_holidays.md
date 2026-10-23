<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `clinic_holidays`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2064](../../../backend/migrations/001_init.sql#L2064) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:2065](../../../backend/migrations/001_init.sql#L2065) |
| `date` | `date` | NOT NULL | — | UK | [001_init.sql:2066](../../../backend/migrations/001_init.sql#L2066) |
| `reason` | `text` | NOT NULL | `''` | — | [001_init.sql:2067](../../../backend/migrations/001_init.sql#L2067) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2068](../../../backend/migrations/001_init.sql#L2068) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2069](../../../backend/migrations/001_init.sql#L2069) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    clinic_holidays {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        date date UK
        text reason
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ clinic_holidays : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_holidays_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:2065](../../../backend/migrations/001_init.sql#L2065) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `clinic_holidays_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `clinic_holidays_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2064](../../../backend/migrations/001_init.sql#L2064) |
| `uk_clinic_holidays_clinic_date` | unique | `CONSTRAINT uk_clinic_holidays_clinic_date UNIQUE (clinic_id, date)` | [001_init.sql:2070](../../../backend/migrations/001_init.sql#L2070) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_clinic_holidays_clinic_date` | NO | — | `CREATE INDEX idx_clinic_holidays_clinic_date ON clinic_holidays(clinic_id, date)` | [001_init.sql:2072](../../../backend/migrations/001_init.sql#L2072) |
