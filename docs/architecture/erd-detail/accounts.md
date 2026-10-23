<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `accounts`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:272](../../../backend/migrations/001_init.sql#L272) |
| `email` | `text` | NOT NULL | — | UK | [001_init.sql:273](../../../backend/migrations/001_init.sql#L273) |
| `password_hash` | `text` | NOT NULL | — | — | [001_init.sql:274](../../../backend/migrations/001_init.sql#L274) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:275](../../../backend/migrations/001_init.sql#L275) |
| `is_system_admin` | `boolean` | NOT NULL | `false` | — | [001_init.sql:276](../../../backend/migrations/001_init.sql#L276) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:277](../../../backend/migrations/001_init.sql#L277) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:278](../../../backend/migrations/001_init.sql#L278) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:279](../../../backend/migrations/001_init.sql#L279) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    accounts {
        BIGSERIAL id PK
        text email UK
        text password_hash
        boolean is_active
        boolean is_system_admin
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
```

## 外部キー（全件）

外部キーなし。

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `accounts_pkey` | primary | `PRIMARY KEY` | [001_init.sql:272](../../../backend/migrations/001_init.sql#L272) |
| `accounts_email_key` | unique | `UNIQUE` | [001_init.sql:273](../../../backend/migrations/001_init.sql#L273) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_accounts_email` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_accounts_email ON accounts(email) WHERE deleted_at IS NULL` | [001_init.sql:282](../../../backend/migrations/001_init.sql#L282) |
| `idx_accounts_system_admin` | NO | `is_system_admin = true AND deleted_at IS NULL` | `CREATE INDEX idx_accounts_system_admin ON accounts(is_system_admin) WHERE is_system_admin = true AND deleted_at IS NULL` | [001_init.sql:283](../../../backend/migrations/001_init.sql#L283) |
