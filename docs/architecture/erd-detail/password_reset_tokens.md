<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `password_reset_tokens`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2822](../../../backend/migrations/001_init.sql#L2822) |
| `account_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:2823](../../../backend/migrations/001_init.sql#L2823) |
| `token_hash` | `text` | NOT NULL | — | UK | [001_init.sql:2824](../../../backend/migrations/001_init.sql#L2824) |
| `expires_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:2825](../../../backend/migrations/001_init.sql#L2825) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:2826](../../../backend/migrations/001_init.sql#L2826) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    password_reset_tokens {
        BIGSERIAL id PK
        bigint account_id FK
        text token_hash UK
        timestamptz expires_at
        timestamptz created_at
    }
    accounts {
        BIGSERIAL id PK
    }
    accounts ||..o{ password_reset_tokens : "account_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `password_reset_tokens_account_id_fkey` | `account_id` | [accounts](accounts.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2823](../../../backend/migrations/001_init.sql#L2823) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `password_reset_tokens_account_id_fkey`: `REFERENCES accounts(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `password_reset_tokens_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2822](../../../backend/migrations/001_init.sql#L2822) |
| `password_reset_tokens_token_hash_key` | unique | `UNIQUE` | [001_init.sql:2824](../../../backend/migrations/001_init.sql#L2824) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_prt_account` | NO | — | `CREATE INDEX idx_prt_account ON password_reset_tokens(account_id)` | [001_init.sql:2829](../../../backend/migrations/001_init.sql#L2829) |
