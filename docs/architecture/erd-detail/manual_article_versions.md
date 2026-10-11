<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `manual_article_versions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK | [001_init.sql:2890](../../../backend/migrations/001_init.sql#L2890) |
| `article_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2891](../../../backend/migrations/001_init.sql#L2891) |
| `title` | `text` | NOT NULL | — | — | [001_init.sql:2892](../../../backend/migrations/001_init.sql#L2892) |
| `order_value` | `numeric(10, 2)` | NOT NULL | — | — | [001_init.sql:2893](../../../backend/migrations/001_init.sql#L2893) |
| `section` | `text` | NOT NULL | — | — | [001_init.sql:2894](../../../backend/migrations/001_init.sql#L2894) |
| `body_markdown` | `text` | NOT NULL | — | — | [001_init.sql:2895](../../../backend/migrations/001_init.sql#L2895) |
| `edited_by_staff_id` | `bigint` | NULL | — | FK | [001_init.sql:2896](../../../backend/migrations/001_init.sql#L2896) |
| `edited_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2897](../../../backend/migrations/001_init.sql#L2897) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    manual_article_versions {
        bigserial id PK
        bigint article_id FK
        text title
        numeric_10__2 order_value
        text section
        text body_markdown
        bigint edited_by_staff_id FK
        timestamptz edited_at
    }
    manual_articles {
        bigserial id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    manual_articles ||..o{ manual_article_versions : "article_id"
    staffs o|..o{ manual_article_versions : "edited_by_staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `manual_article_versions_article_id_fkey` | `article_id` | [manual_articles](manual_articles.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2891](../../../backend/migrations/001_init.sql#L2891) |
| `manual_article_versions_edited_by_staff_id_fkey` | `edited_by_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:2896](../../../backend/migrations/001_init.sql#L2896) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `manual_article_versions_article_id_fkey`: `REFERENCES manual_articles(id) ON DELETE CASCADE`
- `manual_article_versions_edited_by_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `manual_article_versions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2890](../../../backend/migrations/001_init.sql#L2890) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_manual_article_versions_article` | NO | — | `CREATE INDEX idx_manual_article_versions_article ON manual_article_versions(article_id, edited_at DESC)` | [001_init.sql:2900](../../../backend/migrations/001_init.sql#L2900) |
