<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `insurances`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:765](../../../backend/migrations/001_init.sql#L765) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:766](../../../backend/migrations/001_init.sql#L766) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:767](../../../backend/migrations/001_init.sql#L767) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:768](../../../backend/migrations/001_init.sql#L768) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:769](../../../backend/migrations/001_init.sql#L769) |
| `coverage_rate` | `integer` | NOT NULL | `0` | — | [001_init.sql:770](../../../backend/migrations/001_init.sql#L770) |
| `contact_phone` | `text` | NOT NULL | `''` | — | [001_init.sql:771](../../../backend/migrations/001_init.sql#L771) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:772](../../../backend/migrations/001_init.sql#L772) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:773](../../../backend/migrations/001_init.sql#L773) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:774](../../../backend/migrations/001_init.sql#L774) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:775](../../../backend/migrations/001_init.sql#L775) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    insurances {
        BIGSERIAL id PK
        bigint clinic_id FK
        text name
        boolean is_active
        text description
        integer coverage_rate
        text contact_phone
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ insurances : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `insurances_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:766](../../../backend/migrations/001_init.sql#L766) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `insurances_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `insurances_pkey` | primary | `PRIMARY KEY` | [001_init.sql:765](../../../backend/migrations/001_init.sql#L765) |
| `insurances_coverage_rate_check` | check | `CHECK (coverage_rate >= 0 AND coverage_rate <= 100)` | [001_init.sql:770](../../../backend/migrations/001_init.sql#L770) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_insurances_clinic_id` | NO | — | `CREATE INDEX idx_insurances_clinic_id ON insurances(clinic_id)` | [001_init.sql:2218](../../../backend/migrations/001_init.sql#L2218) |
| `idx_insurances_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_insurances_clinic_sort        ON insurances(clinic_id, sort_order)              WHERE deleted_at IS NULL` | [001_init.sql:2264](../../../backend/migrations/001_init.sql#L2264) |
| `idx_insurance_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_insurance_clinic_name ON insurances(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2460](../../../backend/migrations/001_init.sql#L2460) |
| `idx_insurances_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_insurances_active ON insurances(id) WHERE deleted_at IS NULL` | [001_init.sql:2473](../../../backend/migrations/001_init.sql#L2473) |
