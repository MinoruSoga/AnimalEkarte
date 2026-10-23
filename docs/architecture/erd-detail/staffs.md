<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `staffs`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:289](../../../backend/migrations/001_init.sql#L289) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:290](../../../backend/migrations/001_init.sql#L290) |
| `account_id` | `bigint` | NULL | — | FK | [001_init.sql:291](../../../backend/migrations/001_init.sql#L291) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:292](../../../backend/migrations/001_init.sql#L292) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:293](../../../backend/migrations/001_init.sql#L293) |
| `license_number` | `text` | NOT NULL | `''` | — | [001_init.sql:294](../../../backend/migrations/001_init.sql#L294) |
| `occupation_id` | `bigint` | NULL | — | FK | [001_init.sql:295](../../../backend/migrations/001_init.sql#L295) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:296](../../../backend/migrations/001_init.sql#L296) |
| `staff_type` | `staff_type` | NOT NULL | `'doctor'` | — | [001_init.sql:297](../../../backend/migrations/001_init.sql#L297) |
| `reservation_display_name` | `text` | NOT NULL | `''` | — | [001_init.sql:298](../../../backend/migrations/001_init.sql#L298) |
| `reservation_visible` | `boolean` | NOT NULL | `true` | — | [001_init.sql:299](../../../backend/migrations/001_init.sql#L299) |
| `reservation_comment` | `text` | NOT NULL | `''` | — | [001_init.sql:300](../../../backend/migrations/001_init.sql#L300) |
| `reservation_image_url` | `text` | NOT NULL | `''` | — | [001_init.sql:301](../../../backend/migrations/001_init.sql#L301) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:302](../../../backend/migrations/001_init.sql#L302) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:303](../../../backend/migrations/001_init.sql#L303) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:304](../../../backend/migrations/001_init.sql#L304) |
| `position_name` | `text` | NULL | — | — | [001_init.sql:8954](../../../backend/migrations/001_init.sql#L8954) |
| `employment_category` | `text` | NULL | — | — | [001_init.sql:8954](../../../backend/migrations/001_init.sql#L8954) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    staffs {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint account_id FK
        text name
        boolean is_active
        text license_number
        bigint occupation_id FK
        integer sort_order
        staff_type staff_type
        text reservation_display_name
        boolean reservation_visible
        text reservation_comment
        text reservation_image_url
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        text position_name
        text employment_category
    }
    accounts {
        BIGSERIAL id PK
    }
    clinics {
        BIGSERIAL id PK
    }
    occupations {
        BIGSERIAL id PK
    }
    clinics ||..o{ staffs : "clinic_id"
    accounts o|..o{ staffs : "account_id"
    occupations o|..o{ staffs : "occupation_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `staffs_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:290](../../../backend/migrations/001_init.sql#L290) |
| `staffs_account_id_fkey` | `account_id` | [accounts](accounts.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:291](../../../backend/migrations/001_init.sql#L291) |
| `staffs_occupation_id_fkey` | `occupation_id` | [occupations](occupations.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:295](../../../backend/migrations/001_init.sql#L295) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `staffs_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `staffs_account_id_fkey`: `REFERENCES accounts(id) ON DELETE SET NULL`
- `staffs_occupation_id_fkey`: `REFERENCES occupations(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `staffs_pkey` | primary | `PRIMARY KEY` | [001_init.sql:289](../../../backend/migrations/001_init.sql#L289) |
| `uq_staffs_id_clinic` | unique | `CONSTRAINT uq_staffs_id_clinic         UNIQUE (id, clinic_id)` | [001_init.sql:5716](../../../backend/migrations/001_init.sql#L5716) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_staffs_account` | NO | — | `CREATE INDEX idx_staffs_account ON staffs(account_id)` | [001_init.sql:307](../../../backend/migrations/001_init.sql#L307) |
| `idx_staffs_clinic` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_staffs_clinic ON staffs(clinic_id) WHERE deleted_at IS NULL` | [001_init.sql:308](../../../backend/migrations/001_init.sql#L308) |
| `idx_staffs_occupation_id` | NO | — | `CREATE INDEX idx_staffs_occupation_id ON staffs(occupation_id)` | [001_init.sql:2448](../../../backend/migrations/001_init.sql#L2448) |
