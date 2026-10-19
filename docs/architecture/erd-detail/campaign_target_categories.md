<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `campaign_target_categories`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2932](../../../backend/migrations/001_init.sql#L2932) |
| `campaign_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:2933](../../../backend/migrations/001_init.sql#L2933) |
| `category` | `item_category` | NOT NULL | — | UK | [001_init.sql:2934](../../../backend/migrations/001_init.sql#L2934) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    campaign_target_categories {
        BIGSERIAL id PK
        bigint campaign_id FK,UK
        item_category category UK
    }
    campaigns {
        BIGSERIAL id PK
    }
    campaigns ||..o{ campaign_target_categories : "campaign_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `campaign_target_categories_campaign_id_fkey` | `campaign_id` | [campaigns](campaigns.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2933](../../../backend/migrations/001_init.sql#L2933) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `campaign_target_categories_campaign_id_fkey`: `REFERENCES campaigns(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `campaign_target_categories_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2932](../../../backend/migrations/001_init.sql#L2932) |
| `uq_campaign_target_categories` | unique | `CONSTRAINT uq_campaign_target_categories UNIQUE (campaign_id, category)` | [001_init.sql:2935](../../../backend/migrations/001_init.sql#L2935) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_campaign_target_categories_campaign` | NO | — | `CREATE INDEX idx_campaign_target_categories_campaign ON campaign_target_categories(campaign_id)` | [001_init.sql:2947](../../../backend/migrations/001_init.sql#L2947) |
