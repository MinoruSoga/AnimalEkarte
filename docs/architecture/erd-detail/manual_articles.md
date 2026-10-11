<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `manual_articles`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [001_init.sql:2865](../../../backend/migrations/001_init.sql#L2865) |
| `category` | `text` | NOT NULL | — | UK | [001_init.sql:2866](../../../backend/migrations/001_init.sql#L2866) |
| `slug` | `text` | NOT NULL | — | UK | [001_init.sql:2867](../../../backend/migrations/001_init.sql#L2867) |
| `title` | `text` | NOT NULL | — | — | [001_init.sql:2868](../../../backend/migrations/001_init.sql#L2868) |
| `order_value` | `numeric(10, 2)` | NOT NULL | `9999` | — | [001_init.sql:2869](../../../backend/migrations/001_init.sql#L2869) |
| `section` | `text` | NOT NULL | — | — | [001_init.sql:2870](../../../backend/migrations/001_init.sql#L2870) |
| `body_markdown` | `text` | NOT NULL | — | — | [001_init.sql:2871](../../../backend/migrations/001_init.sql#L2871) |
| `updated_by_staff_id` | `bigint` | NULL | — | FK | [001_init.sql:2872](../../../backend/migrations/001_init.sql#L2872) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2873](../../../backend/migrations/001_init.sql#L2873) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2874](../../../backend/migrations/001_init.sql#L2874) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    manual_articles {
        bigserial id PK
        text category UK
        text slug UK
        text title
        numeric_10__2 order_value
        text section
        text body_markdown
        bigint updated_by_staff_id FK
        timestamptz created_at
        timestamptz updated_at
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    staffs o|..o{ manual_articles : "updated_by_staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `manual_articles_updated_by_staff_id_fkey` | `updated_by_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:2872](../../../backend/migrations/001_init.sql#L2872) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `manual_articles_updated_by_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `manual_articles_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2865](../../../backend/migrations/001_init.sql#L2865) |
| `manual_articles_category_check` | check | `CHECK (category IN ('screens', 'workflows'))` | [001_init.sql:2866](../../../backend/migrations/001_init.sql#L2866) |
| `manual_articles_category_slug_key` | unique | `UNIQUE (category, slug)` | [001_init.sql:2876](../../../backend/migrations/001_init.sql#L2876) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_manual_articles_category_slug` | NO | — | `CREATE INDEX idx_manual_articles_category_slug ON manual_articles(category, slug)` | [001_init.sql:2879](../../../backend/migrations/001_init.sql#L2879) |
| `idx_manual_articles_updated_at` | NO | — | `CREATE INDEX idx_manual_articles_updated_at    ON manual_articles(updated_at DESC)` | [001_init.sql:2880](../../../backend/migrations/001_init.sql#L2880) |
