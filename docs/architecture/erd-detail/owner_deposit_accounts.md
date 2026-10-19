<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `owner_deposit_accounts`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8402](../../../backend/migrations/001_init.sql#L8402) |
| `owner_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8403](../../../backend/migrations/001_init.sql#L8403) |
| `revision` | `bigint` | NOT NULL | `1` | — | [001_init.sql:8404](../../../backend/migrations/001_init.sql#L8404) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8405](../../../backend/migrations/001_init.sql#L8405) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    owner_deposit_accounts {
        bigint clinic_id PK,FK
        bigint owner_id PK,FK
        bigint revision
        timestamptz created_at
    }
    clinics {
        BIGSERIAL id PK
    }
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ owner_deposit_accounts : "clinic_id"
    owners ||--o| owner_deposit_accounts : "clinic_id, owner_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `owner_deposit_accounts_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8402](../../../backend/migrations/001_init.sql#L8402) |
| `owner_deposit_accounts_clinic_id_owner_id_fkey` | `clinic_id, owner_id` | [owners](owners.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:8407](../../../backend/migrations/001_init.sql#L8407) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `owner_deposit_accounts_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `owner_deposit_accounts_clinic_id_owner_id_fkey`: `FOREIGN KEY (clinic_id,owner_id) REFERENCES owners(clinic_id,id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `owner_deposit_accounts_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:8404](../../../backend/migrations/001_init.sql#L8404) |
| `owner_deposit_accounts_pkey` | primary | `PRIMARY KEY (clinic_id,owner_id)` | [001_init.sql:8406](../../../backend/migrations/001_init.sql#L8406) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
