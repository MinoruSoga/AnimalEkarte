<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `reservation_type_unavailable_times`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2764](../../../backend/migrations/001_init.sql#L2764) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2765](../../../backend/migrations/001_init.sql#L2765) |
| `reservation_type_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2766](../../../backend/migrations/001_init.sql#L2766) |
| `unavailable_type` | `text` | NOT NULL | — | — | [001_init.sql:2767](../../../backend/migrations/001_init.sql#L2767) |
| `day_of_week` | `smallint` | NULL | — | — | [001_init.sql:2769](../../../backend/migrations/001_init.sql#L2769) |
| `specific_date` | `date` | NULL | — | — | [001_init.sql:2773](../../../backend/migrations/001_init.sql#L2773) |
| `start_time` | `varchar(5)` | NOT NULL | — | — | [001_init.sql:2780](../../../backend/migrations/001_init.sql#L2780) |
| `end_time` | `varchar(5)` | NOT NULL | — | — | [001_init.sql:2781](../../../backend/migrations/001_init.sql#L2781) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2782](../../../backend/migrations/001_init.sql#L2782) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2783](../../../backend/migrations/001_init.sql#L2783) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    reservation_type_unavailable_times {
        BIGSERIAL id PK
        bigint clinic_id FK
        bigint reservation_type_id FK
        text unavailable_type
        smallint day_of_week
        date specific_date
        varchar_5 start_time
        varchar_5 end_time
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    reservation_types {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ reservation_type_unavailable_times : "clinic_id"
    reservation_types ||..o{ reservation_type_unavailable_times : "reservation_type_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `reservation_type_unavailable_times_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:2765](../../../backend/migrations/001_init.sql#L2765) |
| `reservation_type_unavailable_times_reservation_type_id_fkey` | `reservation_type_id` | [reservation_types](reservation_types.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2766](../../../backend/migrations/001_init.sql#L2766) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `reservation_type_unavailable_times_clinic_id_fkey`: `REFERENCES clinics(id)`
- `reservation_type_unavailable_times_reservation_type_id_fkey`: `REFERENCES reservation_types(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `reservation_type_unavailable_times_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2764](../../../backend/migrations/001_init.sql#L2764) |
| `reservation_type_unavailable_times_unavailable_type_check` | check | `CHECK (unavailable_type IN ('weekly', 'specific'))` | [001_init.sql:2767](../../../backend/migrations/001_init.sql#L2767) |
| `reservation_type_unavailable_times_check` | check | `CHECK (                             (unavailable_type = 'weekly' AND day_of_week BETWEEN 0 AND 6)                             OR (unavailable_type = 'specific' AND day_of_week IS NULL)                         )` | [001_init.sql:2769](../../../backend/migrations/001_init.sql#L2769) |
| `reservation_type_unavailable_times_check1` | check | `CHECK (                             (unavailable_type = 'specific' AND specific_date IS NOT NULL)                             OR (unavailable_type = 'weekly' AND specific_date IS NULL)                         )` | [001_init.sql:2773](../../../backend/migrations/001_init.sql#L2773) |
| `reservation_type_unavailable_times_start_time_check` | check | `CHECK (start_time ~ '^\d{2}:\d{2}$')` | [001_init.sql:2780](../../../backend/migrations/001_init.sql#L2780) |
| `reservation_type_unavailable_times_end_time_check` | check | `CHECK (end_time ~ '^\d{2}:\d{2}$')` | [001_init.sql:2781](../../../backend/migrations/001_init.sql#L2781) |
| `chk_unavailable_time_range` | check | `CONSTRAINT chk_unavailable_time_range CHECK (end_time > start_time)` | [001_init.sql:2784](../../../backend/migrations/001_init.sql#L2784) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_rtype_unavailable_clinic_type` | NO | — | `CREATE INDEX idx_rtype_unavailable_clinic_type     ON reservation_type_unavailable_times(clinic_id, reservation_type_id)` | [001_init.sql:2788](../../../backend/migrations/001_init.sql#L2788) |
| `idx_rtype_unavailable_weekly` | NO | `unavailable_type = 'weekly'` | `CREATE INDEX idx_rtype_unavailable_weekly     ON reservation_type_unavailable_times(reservation_type_id, day_of_week)     WHERE unavailable_type = 'weekly'` | [001_init.sql:2790](../../../backend/migrations/001_init.sql#L2790) |
| `idx_rtype_unavailable_specific` | NO | `unavailable_type = 'specific'` | `CREATE INDEX idx_rtype_unavailable_specific     ON reservation_type_unavailable_times(reservation_type_id, specific_date)     WHERE unavailable_type = 'specific'` | [001_init.sql:2793](../../../backend/migrations/001_init.sql#L2793) |
