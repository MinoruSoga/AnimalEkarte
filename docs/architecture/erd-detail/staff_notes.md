<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `staff_notes`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1892](../../../backend/migrations/001_init.sql#L1892) |
| `daily_record_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1893](../../../backend/migrations/001_init.sql#L1893) |
| `time` | `time` | NOT NULL | — | — | [001_init.sql:1894](../../../backend/migrations/001_init.sql#L1894) |
| `content` | `text` | NOT NULL | `''` | — | [001_init.sql:1895](../../../backend/migrations/001_init.sql#L1895) |
| `staff_id` | `bigint` | NULL | — | FK | [001_init.sql:1896](../../../backend/migrations/001_init.sql#L1896) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1897](../../../backend/migrations/001_init.sql#L1897) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1898](../../../backend/migrations/001_init.sql#L1898) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    staff_notes {
        BIGSERIAL id PK
        bigint daily_record_id FK
        time time
        text content
        bigint staff_id FK
        timestamptz created_at
        timestamptz updated_at
    }
    daily_records {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    daily_records ||..o{ staff_notes : "daily_record_id"
    staffs o|..o{ staff_notes : "staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `staff_notes_daily_record_id_fkey` | `daily_record_id` | [daily_records](daily_records.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1893](../../../backend/migrations/001_init.sql#L1893) |
| `staff_notes_staff_id_fkey` | `staff_id` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1896](../../../backend/migrations/001_init.sql#L1896) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `staff_notes_daily_record_id_fkey`: `REFERENCES daily_records(id) ON DELETE CASCADE`
- `staff_notes_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `staff_notes_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1892](../../../backend/migrations/001_init.sql#L1892) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
