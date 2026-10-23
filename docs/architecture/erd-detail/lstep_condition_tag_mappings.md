<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lstep_condition_tag_mappings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:658](../../../backend/migrations/001_init.sql#L658) |
| `condition_code` | `VARCHAR(50)` | NOT NULL | — | UK | [001_init.sql:659](../../../backend/migrations/001_init.sql#L659) |
| `tag_name` | `VARCHAR(100)` | NOT NULL | — | — | [001_init.sql:660](../../../backend/migrations/001_init.sql#L660) |
| `description` | `TEXT` | NULL | — | — | [001_init.sql:661](../../../backend/migrations/001_init.sql#L661) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | — | [001_init.sql:662](../../../backend/migrations/001_init.sql#L662) |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | — | [001_init.sql:663](../../../backend/migrations/001_init.sql#L663) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lstep_condition_tag_mappings {
        BIGSERIAL id PK
        VARCHAR_50 condition_code UK
        VARCHAR_100 tag_name
        TEXT description
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

## 外部キー（全件）

外部キーなし。

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lstep_condition_tag_mappings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:658](../../../backend/migrations/001_init.sql#L658) |
| `lstep_condition_tag_mappings_condition_code_key` | unique | `UNIQUE` | [001_init.sql:659](../../../backend/migrations/001_init.sql#L659) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
