<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `manual_article_versions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `bigserial` | NOT NULL | — | PK, UK | [001_init.sql:2898](../../../backend/migrations/001_init.sql#L2898) |
| `article_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:2899](../../../backend/migrations/001_init.sql#L2899) |
| `title` | `text` | NOT NULL | — | — | [001_init.sql:2900](../../../backend/migrations/001_init.sql#L2900) |
| `order_value` | `numeric(10, 2)` | NOT NULL | — | — | [001_init.sql:2901](../../../backend/migrations/001_init.sql#L2901) |
| `section` | `text` | NOT NULL | — | — | [001_init.sql:2902](../../../backend/migrations/001_init.sql#L2902) |
| `body_markdown` | `text` | NOT NULL | — | — | [001_init.sql:2903](../../../backend/migrations/001_init.sql#L2903) |
| `edited_by_staff_id` | `bigint` | NULL | — | FK | [001_init.sql:2904](../../../backend/migrations/001_init.sql#L2904) |
| `edited_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2905](../../../backend/migrations/001_init.sql#L2905) |
| `content_digest` | `bytea` | NULL | — | — | [001_init.sql:9986](../../../backend/migrations/001_init.sql#L9986) |
| `content_revision` | `bigint` | NULL | — | UK | [001_init.sql:9987](../../../backend/migrations/001_init.sql#L9987) |
| `restored_from_id` | `bigint` | NULL | — | FK | [001_init.sql:9988](../../../backend/migrations/001_init.sql#L9988) |
| `source_kind` | `text` | NULL | — | — | [001_init.sql:9989](../../../backend/migrations/001_init.sql#L9989) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    manual_article_versions {
        bigserial id PK,UK
        bigint article_id FK,UK
        text title
        numeric_10__2 order_value
        text section
        text body_markdown
        bigint edited_by_staff_id FK
        timestamptz edited_at
        bytea content_digest
        bigint content_revision UK
        bigint restored_from_id FK
        text source_kind
    }
    manual_articles {
        bigserial id PK,FK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    staffs o|..o{ manual_article_versions : "edited_by_staff_id"
    manual_articles ||..o{ manual_article_versions : "article_id"
    manual_article_versions o|..o{ manual_article_versions : "restored_from_id, article_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `manual_article_versions_edited_by_staff_id_fkey` | `edited_by_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:2904](../../../backend/migrations/001_init.sql#L2904) |
| `fk_manual_version_article` | `article_id` | [manual_articles](manual_articles.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9985](../../../backend/migrations/001_init.sql#L9985) |
| `fk_manual_restore` | `restored_from_id, article_id` | [manual_article_versions](manual_article_versions.md) | `id, article_id` | `ON DELETE RESTRICT` | [001_init.sql:9992](../../../backend/migrations/001_init.sql#L9992) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `manual_article_versions_edited_by_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`
- `fk_manual_version_article`: `CONSTRAINT fk_manual_version_article FOREIGN KEY(article_id) REFERENCES manual_articles(id) ON DELETE RESTRICT`
- `fk_manual_restore`: `CONSTRAINT fk_manual_restore FOREIGN KEY(restored_from_id,article_id) REFERENCES manual_article_versions(id,article_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `manual_article_versions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2898](../../../backend/migrations/001_init.sql#L2898) |
| `manual_article_versions_content_digest_check` | check | `CHECK(octet_length(content_digest)=32)` | [001_init.sql:9986](../../../backend/migrations/001_init.sql#L9986) |
| `manual_article_versions_content_revision_check` | check | `CHECK(content_revision>0)` | [001_init.sql:9987](../../../backend/migrations/001_init.sql#L9987) |
| `manual_article_versions_source_kind_check` | check | `CHECK(source_kind IN('static','override','edited','restored','legacy_unknown'))` | [001_init.sql:9989](../../../backend/migrations/001_init.sql#L9989) |
| `uq_manual_version_article` | unique | `CONSTRAINT uq_manual_version_article UNIQUE(id,article_id)` | [001_init.sql:9990](../../../backend/migrations/001_init.sql#L9990) |
| `uq_manual_content_revision` | unique | `CONSTRAINT uq_manual_content_revision UNIQUE(article_id,content_revision)` | [001_init.sql:9991](../../../backend/migrations/001_init.sql#L9991) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_manual_article_versions_article` | NO | — | `CREATE INDEX idx_manual_article_versions_article ON manual_article_versions(article_id, edited_at DESC)` | [001_init.sql:2908](../../../backend/migrations/001_init.sql#L2908) |
