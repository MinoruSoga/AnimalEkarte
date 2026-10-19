<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `animal_species`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:233](../../../backend/migrations/001_init.sql#L233) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:234](../../../backend/migrations/001_init.sql#L234) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:235](../../../backend/migrations/001_init.sql#L235) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:236](../../../backend/migrations/001_init.sql#L236) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:237](../../../backend/migrations/001_init.sql#L237) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:238](../../../backend/migrations/001_init.sql#L238) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    animal_species {
        BIGSERIAL id PK
        text name
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }
```

## 外部キー（全件）

外部キーなし。

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `animal_species_pkey` | primary | `PRIMARY KEY` | [001_init.sql:233](../../../backend/migrations/001_init.sql#L233) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_animal_species_name` | YES | `is_active = true` | `CREATE UNIQUE INDEX idx_animal_species_name ON animal_species(name) WHERE is_active = true` | [001_init.sql:2473](../../../backend/migrations/001_init.sql#L2473) |
