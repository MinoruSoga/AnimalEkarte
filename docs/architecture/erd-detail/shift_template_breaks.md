<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `shift_template_breaks`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2730](../../../backend/migrations/001_init.sql#L2730) |
| `shift_template_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2731](../../../backend/migrations/001_init.sql#L2731) |
| `break_start` | `time` | NOT NULL | — | — | [001_init.sql:2732](../../../backend/migrations/001_init.sql#L2732) |
| `break_end` | `time` | NOT NULL | — | — | [001_init.sql:2733](../../../backend/migrations/001_init.sql#L2733) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    shift_template_breaks {
        BIGSERIAL id PK
        bigint shift_template_id FK
        time break_start
        time break_end
    }
    shift_templates {
        BIGSERIAL id PK
    }
    shift_templates ||..o{ shift_template_breaks : "shift_template_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `shift_template_breaks_shift_template_id_fkey` | `shift_template_id` | [shift_templates](shift_templates.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2731](../../../backend/migrations/001_init.sql#L2731) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `shift_template_breaks_shift_template_id_fkey`: `REFERENCES shift_templates(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `shift_template_breaks_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2730](../../../backend/migrations/001_init.sql#L2730) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_shift_template_breaks_template` | NO | — | `CREATE INDEX idx_shift_template_breaks_template ON shift_template_breaks(shift_template_id)` | [001_init.sql:2736](../../../backend/migrations/001_init.sql#L2736) |
