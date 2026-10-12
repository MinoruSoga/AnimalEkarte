<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lstep_tag_cache`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:380](../../../backend/migrations/001_init.sql#L380) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:381](../../../backend/migrations/001_init.sql#L381) |
| `owner_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:382](../../../backend/migrations/001_init.sql#L382) |
| `tag_name` | `varchar(100)` | NOT NULL | — | UK | [001_init.sql:383](../../../backend/migrations/001_init.sql#L383) |
| `category` | `varchar(20)` | NOT NULL | `'auto'` | — | [001_init.sql:384](../../../backend/migrations/001_init.sql#L384) |
| `synced_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:386](../../../backend/migrations/001_init.sql#L386) |
| `reason` | `text` | NULL | — | — | [001_init.sql:387](../../../backend/migrations/001_init.sql#L387) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lstep_tag_cache {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        bigint owner_id FK,UK
        varchar_100 tag_name UK
        varchar_20 category
        timestamptz synced_at
        text reason
    }
    clinics {
        BIGSERIAL id PK
    }
    owners {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ lstep_tag_cache : "clinic_id"
    owners ||..o{ lstep_tag_cache : "owner_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lstep_tag_cache_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:381](../../../backend/migrations/001_init.sql#L381) |
| `lstep_tag_cache_owner_id_fkey` | `owner_id` | [owners](owners.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:382](../../../backend/migrations/001_init.sql#L382) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lstep_tag_cache_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `lstep_tag_cache_owner_id_fkey`: `REFERENCES owners(id)  ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lstep_tag_cache_pkey` | primary | `PRIMARY KEY` | [001_init.sql:380](../../../backend/migrations/001_init.sql#L380) |
| `lstep_tag_cache_category_check` | check | `CHECK (category IN ('auto', 'manual'))` | [001_init.sql:384](../../../backend/migrations/001_init.sql#L384) |
| `lstep_tag_cache_clinic_id_owner_id_tag_name_key` | unique | `UNIQUE (clinic_id, owner_id, tag_name)` | [001_init.sql:388](../../../backend/migrations/001_init.sql#L388) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lstep_tag_cache_clinic_tag` | NO | — | `CREATE INDEX idx_lstep_tag_cache_clinic_tag ON lstep_tag_cache (clinic_id, tag_name)` | [001_init.sql:392](../../../backend/migrations/001_init.sql#L392) |
| `idx_lstep_tag_cache_owner` | NO | — | `CREATE INDEX idx_lstep_tag_cache_owner ON lstep_tag_cache (clinic_id, owner_id)` | [001_init.sql:394](../../../backend/migrations/001_init.sql#L394) |
