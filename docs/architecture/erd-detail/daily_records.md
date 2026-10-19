<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `daily_records`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1741](../../../backend/migrations/001_init.sql#L1741) |
| `hospitalization_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1742](../../../backend/migrations/001_init.sql#L1742) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1743](../../../backend/migrations/001_init.sql#L1743) |
| `date` | `date` | NOT NULL | — | UK | [001_init.sql:1744](../../../backend/migrations/001_init.sql#L1744) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1745](../../../backend/migrations/001_init.sql#L1745) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1746](../../../backend/migrations/001_init.sql#L1746) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    daily_records {
        BIGSERIAL id PK
        bigint hospitalization_id FK,UK
        bigint clinic_id FK
        date date UK
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    hospitalizations {
        BIGSERIAL id PK
    }
    hospitalizations ||..o{ daily_records : "hospitalization_id"
    clinics ||..o{ daily_records : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `daily_records_hospitalization_id_fkey` | `hospitalization_id` | [hospitalizations](hospitalizations.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1742](../../../backend/migrations/001_init.sql#L1742) |
| `daily_records_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1743](../../../backend/migrations/001_init.sql#L1743) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `daily_records_hospitalization_id_fkey`: `REFERENCES hospitalizations(id) ON DELETE CASCADE`
- `daily_records_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `daily_records_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1741](../../../backend/migrations/001_init.sql#L1741) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_daily_records_hosp_date` | YES | — | `CREATE UNIQUE INDEX idx_daily_records_hosp_date ON daily_records(hospitalization_id, date)` | [001_init.sql:2199](../../../backend/migrations/001_init.sql#L2199) |
| `idx_daily_records_clinic_id` | NO | — | `CREATE INDEX idx_daily_records_clinic_id ON daily_records(clinic_id)` | [001_init.sql:2455](../../../backend/migrations/001_init.sql#L2455) |
