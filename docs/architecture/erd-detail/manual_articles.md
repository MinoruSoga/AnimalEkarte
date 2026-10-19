<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `manual_articles`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK, FK | [001_init.sql:2873](../../../backend/migrations/001_init.sql#L2873) |
| `category` | `text` | NOT NULL | — | UK | [001_init.sql:2874](../../../backend/migrations/001_init.sql#L2874) |
| `slug` | `text` | NOT NULL | — | UK | [001_init.sql:2875](../../../backend/migrations/001_init.sql#L2875) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2881](../../../backend/migrations/001_init.sql#L2881) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2882](../../../backend/migrations/001_init.sql#L2882) |
| `version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:9978](../../../backend/migrations/001_init.sql#L9978) |
| `current_version_id` | `bigint` | NULL | — | FK | [001_init.sql:9979](../../../backend/migrations/001_init.sql#L9979) |
| `published_version_id` | `bigint` | NULL | — | FK | [001_init.sql:9980](../../../backend/migrations/001_init.sql#L9980) |
| `hidden_at` | `timestamptz` | NULL | — | — | [001_init.sql:9981](../../../backend/migrations/001_init.sql#L9981) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    manual_articles {
        bigserial id PK,FK
        text category UK
        text slug UK
        timestamptz created_at
        timestamptz updated_at
        bigint version
        bigint current_version_id FK
        bigint published_version_id FK
        timestamptz hidden_at
    }
    manual_article_versions {
        bigserial id PK,UK
        bigint article_id FK,UK
    }
    manual_article_versions o|..o| manual_articles : "current_version_id, id"
    manual_article_versions o|..o| manual_articles : "published_version_id, id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `fk_manual_current` | `current_version_id, id` | [manual_article_versions](manual_article_versions.md) | `id, article_id` | `ON DELETE RESTRICT` | [001_init.sql:9991](../../../backend/migrations/001_init.sql#L9991) |
| `fk_manual_published` | `published_version_id, id` | [manual_article_versions](manual_article_versions.md) | `id, article_id` | `ON DELETE RESTRICT` | [001_init.sql:9992](../../../backend/migrations/001_init.sql#L9992) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `fk_manual_current`: `CONSTRAINT fk_manual_current FOREIGN KEY(current_version_id,id) REFERENCES manual_article_versions(id,article_id) ON DELETE RESTRICT`
- `fk_manual_published`: `CONSTRAINT fk_manual_published FOREIGN KEY(published_version_id,id) REFERENCES manual_article_versions(id,article_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `manual_articles_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2873](../../../backend/migrations/001_init.sql#L2873) |
| `manual_articles_category_check` | check | `CHECK (category IN ('screens', 'workflows'))` | [001_init.sql:2874](../../../backend/migrations/001_init.sql#L2874) |
| `manual_articles_category_slug_key` | unique | `UNIQUE (category, slug)` | [001_init.sql:2884](../../../backend/migrations/001_init.sql#L2884) |
| `manual_articles_version_check` | check | `CHECK(version>0)` | [001_init.sql:9978](../../../backend/migrations/001_init.sql#L9978) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_manual_articles_category_slug` | NO | — | `CREATE INDEX idx_manual_articles_category_slug ON manual_articles(category, slug)` | [001_init.sql:2887](../../../backend/migrations/001_init.sql#L2887) |
| `idx_manual_articles_updated_at` | NO | — | `CREATE INDEX idx_manual_articles_updated_at    ON manual_articles(updated_at DESC)` | [001_init.sql:2888](../../../backend/migrations/001_init.sql#L2888) |
