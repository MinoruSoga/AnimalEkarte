<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `shift_entry_breaks`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2708](../../../backend/migrations/001_init.sql#L2708) |
| `shift_entry_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2709](../../../backend/migrations/001_init.sql#L2709) |
| `starts_at` | `timestamptz` | NULL | — | — | [001_init.sql:8964](../../../backend/migrations/001_init.sql#L8964) |
| `ends_at` | `timestamptz` | NULL | — | — | [001_init.sql:8965](../../../backend/migrations/001_init.sql#L8965) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    shift_entry_breaks {
        BIGSERIAL id PK
        bigint shift_entry_id FK
        timestamptz starts_at
        timestamptz ends_at
    }
    shift_entries {
        BIGSERIAL id PK,UK
    }
    shift_entries ||..o{ shift_entry_breaks : "shift_entry_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `shift_entry_breaks_shift_entry_id_fkey` | `shift_entry_id` | [shift_entries](shift_entries.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2709](../../../backend/migrations/001_init.sql#L2709) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `shift_entry_breaks_shift_entry_id_fkey`: `REFERENCES shift_entries(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `shift_entry_breaks_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2708](../../../backend/migrations/001_init.sql#L2708) |
| `ck_shift_break_explicit_interval` | check | `CONSTRAINT ck_shift_break_explicit_interval CHECK ((starts_at IS NULL AND ends_at IS NULL)       OR (starts_at IS NOT NULL AND ends_at IS NOT NULL AND ends_at>starts_at))` | [001_init.sql:8966](../../../backend/migrations/001_init.sql#L8966) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_shift_entry_breaks_entry` | NO | — | `CREATE INDEX idx_shift_entry_breaks_entry ON shift_entry_breaks(shift_entry_id)` | [001_init.sql:2713](../../../backend/migrations/001_init.sql#L2713) |
