<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `companies`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:170](../../../backend/migrations/001_init.sql#L170) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:171](../../../backend/migrations/001_init.sql#L171) |
| `postal_code` | `text` | NOT NULL | `''` | — | [001_init.sql:172](../../../backend/migrations/001_init.sql#L172) |
| `address` | `text` | NOT NULL | `''` | — | [001_init.sql:173](../../../backend/migrations/001_init.sql#L173) |
| `phone_number` | `text` | NOT NULL | `''` | — | [001_init.sql:174](../../../backend/migrations/001_init.sql#L174) |
| `fax_number` | `text` | NOT NULL | `''` | — | [001_init.sql:175](../../../backend/migrations/001_init.sql#L175) |
| `registration_number` | `text` | NOT NULL | `''` | — | [001_init.sql:176](../../../backend/migrations/001_init.sql#L176) |
| `invoice_registration_number` | `text` | NOT NULL | `''` | — | [001_init.sql:177](../../../backend/migrations/001_init.sql#L177) |
| `director_name` | `text` | NOT NULL | `''` | — | [001_init.sql:178](../../../backend/migrations/001_init.sql#L178) |
| `email` | `text` | NOT NULL | `''` | — | [001_init.sql:179](../../../backend/migrations/001_init.sql#L179) |
| `website` | `text` | NOT NULL | `''` | — | [001_init.sql:180](../../../backend/migrations/001_init.sql#L180) |
| `logo_url` | `text` | NOT NULL | `''` | — | [001_init.sql:181](../../../backend/migrations/001_init.sql#L181) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:182](../../../backend/migrations/001_init.sql#L182) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:183](../../../backend/migrations/001_init.sql#L183) |

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
| `companies_pkey` | primary | `PRIMARY KEY` | [001_init.sql:170](../../../backend/migrations/001_init.sql#L170) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
