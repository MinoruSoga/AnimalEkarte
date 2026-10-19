<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `companies`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:156](../../../backend/migrations/001_init.sql#L156) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:157](../../../backend/migrations/001_init.sql#L157) |
| `postal_code` | `text` | NOT NULL | `''` | — | [001_init.sql:158](../../../backend/migrations/001_init.sql#L158) |
| `address` | `text` | NOT NULL | `''` | — | [001_init.sql:159](../../../backend/migrations/001_init.sql#L159) |
| `phone_number` | `text` | NOT NULL | `''` | — | [001_init.sql:160](../../../backend/migrations/001_init.sql#L160) |
| `fax_number` | `text` | NOT NULL | `''` | — | [001_init.sql:161](../../../backend/migrations/001_init.sql#L161) |
| `registration_number` | `text` | NOT NULL | `''` | — | [001_init.sql:162](../../../backend/migrations/001_init.sql#L162) |
| `invoice_registration_number` | `text` | NOT NULL | `''` | — | [001_init.sql:163](../../../backend/migrations/001_init.sql#L163) |
| `director_name` | `text` | NOT NULL | `''` | — | [001_init.sql:164](../../../backend/migrations/001_init.sql#L164) |
| `email` | `text` | NOT NULL | `''` | — | [001_init.sql:165](../../../backend/migrations/001_init.sql#L165) |
| `website` | `text` | NOT NULL | `''` | — | [001_init.sql:166](../../../backend/migrations/001_init.sql#L166) |
| `logo_url` | `text` | NOT NULL | `''` | — | [001_init.sql:167](../../../backend/migrations/001_init.sql#L167) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:168](../../../backend/migrations/001_init.sql#L168) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:169](../../../backend/migrations/001_init.sql#L169) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    companies {
        BIGSERIAL id PK
        text name
        text postal_code
        text address
        text phone_number
        text fax_number
        text registration_number
        text invoice_registration_number
        text director_name
        text email
        text website
        text logo_url
        timestamptz created_at
        timestamptz updated_at
    }
```

## 外部キー（全件）

外部キーなし。

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `companies_pkey` | primary | `PRIMARY KEY` | [001_init.sql:156](../../../backend/migrations/001_init.sql#L156) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
