<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `closing_special_periods`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2145](../../../backend/migrations/001_init.sql#L2145) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2146](../../../backend/migrations/001_init.sql#L2146) |
| `start_date` | `date` | NOT NULL | — | — | [001_init.sql:2147](../../../backend/migrations/001_init.sql#L2147) |
| `end_date` | `date` | NOT NULL | — | — | [001_init.sql:2148](../../../backend/migrations/001_init.sql#L2148) |
| `am_pm_boundary` | `time` | NOT NULL | — | — | [001_init.sql:2149](../../../backend/migrations/001_init.sql#L2149) |
| `pm_end` | `time` | NOT NULL | — | — | [001_init.sql:2150](../../../backend/migrations/001_init.sql#L2150) |
| `note` | `varchar(100)` | NOT NULL | `''` | — | [001_init.sql:2151](../../../backend/migrations/001_init.sql#L2151) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2152](../../../backend/migrations/001_init.sql#L2152) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2153](../../../backend/migrations/001_init.sql#L2153) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    closing_special_periods {
        BIGSERIAL id PK
        bigint clinic_id FK
        date start_date
        date end_date
        time am_pm_boundary
        time pm_end
        varchar_100 note
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ closing_special_periods : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `closing_special_periods_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2146](../../../backend/migrations/001_init.sql#L2146) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `closing_special_periods_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `closing_special_periods_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2145](../../../backend/migrations/001_init.sql#L2145) |
| `chk_closing_special_periods_date_range` | check | `CONSTRAINT chk_closing_special_periods_date_range CHECK (start_date <= end_date)` | [001_init.sql:2154](../../../backend/migrations/001_init.sql#L2154) |
| `chk_closing_special_periods_time_order` | check | `CONSTRAINT chk_closing_special_periods_time_order CHECK (am_pm_boundary < pm_end)` | [001_init.sql:2155](../../../backend/migrations/001_init.sql#L2155) |
| `excl_closing_special_periods_clinic_daterange` | exclude | `CONSTRAINT excl_closing_special_periods_clinic_daterange     EXCLUDE USING gist (         clinic_id WITH =,         daterange(start_date, end_date, '[]') WITH &&     )` | [001_init.sql:4594](../../../backend/migrations/001_init.sql#L4594) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_closing_special_periods_clinic` | NO | — | `CREATE INDEX idx_closing_special_periods_clinic ON closing_special_periods(clinic_id, start_date, end_date)` | [001_init.sql:2158](../../../backend/migrations/001_init.sql#L2158) |
