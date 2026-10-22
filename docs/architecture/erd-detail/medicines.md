<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medicines`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:749](../../../backend/migrations/001_init.sql#L749) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:750](../../../backend/migrations/001_init.sql#L750) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:751](../../../backend/migrations/001_init.sql#L751) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:752](../../../backend/migrations/001_init.sql#L752) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:753](../../../backend/migrations/001_init.sql#L753) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:754](../../../backend/migrations/001_init.sql#L754) |
| `parent_id` | `bigint` | NULL | — | FK | [001_init.sql:755](../../../backend/migrations/001_init.sql#L755) |
| `dosage_form` | `dosage_form` | NULL | — | — | [001_init.sql:756](../../../backend/migrations/001_init.sql#L756) |
| `medicine_unit` | `medicine_unit` | NULL | — | — | [001_init.sql:757](../../../backend/migrations/001_init.sql#L757) |
| `inventory_id` | `bigint` | NULL | — | FK | [001_init.sql:758](../../../backend/migrations/001_init.sql#L758) |
| `default_quantity` | `numeric(10,2)` | NULL | `1` | — | [001_init.sql:3461](../../../backend/migrations/001_init.sql#L3461) |
| `tax_type` | `tax_type` | NOT NULL | `'excluded'` | — | [001_init.sql:760](../../../backend/migrations/001_init.sql#L760) |
| `tax_rate` | `numeric` | NOT NULL | `0.10` | — | [001_init.sql:761](../../../backend/migrations/001_init.sql#L761) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:762](../../../backend/migrations/001_init.sql#L762) |
| `is_non_insurance` | `boolean` | NOT NULL | `false` | — | [001_init.sql:763](../../../backend/migrations/001_init.sql#L763) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:764](../../../backend/migrations/001_init.sql#L764) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:765](../../../backend/migrations/001_init.sql#L765) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:766](../../../backend/migrations/001_init.sql#L766) |
| `calculation_type` | `medicine_calculation_type` | NOT NULL | `'none'` | — | [001_init.sql:3457](../../../backend/migrations/001_init.sql#L3457) |
| `strength` | `numeric(10,4)` | NULL | — | — | [001_init.sql:3458](../../../backend/migrations/001_init.sql#L3458) |
| `frequency_per_day` | `integer` | NULL | — | — | [001_init.sql:3459](../../../backend/migrations/001_init.sql#L3459) |
| `default_duration_days` | `integer` | NULL | — | — | [001_init.sql:3460](../../../backend/migrations/001_init.sql#L3460) |
| `is_bill_discount_excluded` | `boolean` | NULL | — | — | [007_bill_discount_excluded_and_master_versions.sql:24](../../../backend/migrations/007_bill_discount_excluded_and_master_versions.sql#L24) |
| `version` | `bigint` | NOT NULL | `1` | — | [007_bill_discount_excluded_and_master_versions.sql:40](../../../backend/migrations/007_bill_discount_excluded_and_master_versions.sql#L40) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medicines {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        bigint price
        boolean is_active
        text description
        bigint parent_id FK
        dosage_form dosage_form
        medicine_unit medicine_unit
        bigint inventory_id FK
        numeric_10_2 default_quantity
        tax_type tax_type
        numeric tax_rate
        integer sort_order
        boolean is_non_insurance
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        medicine_calculation_type calculation_type
        numeric_10_4 strength
        integer frequency_per_day
        integer default_duration_days
        boolean is_bill_discount_excluded
        bigint version
    }
    clinics {
        BIGSERIAL id PK
    }
    inventory_items {
        BIGSERIAL id PK,FK,UK
    }
    clinics ||..o{ medicines : "clinic_id"
    medicines o|..o{ medicines : "parent_id"
    inventory_items o|..o{ medicines : "inventory_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medicines_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:750](../../../backend/migrations/001_init.sql#L750) |
| `medicines_parent_id_fkey` | `parent_id` | [medicines](medicines.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:755](../../../backend/migrations/001_init.sql#L755) |
| `medicines_inventory_id_fkey` | `inventory_id` | [inventory_items](inventory_items.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:758](../../../backend/migrations/001_init.sql#L758) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medicines_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medicines_parent_id_fkey`: `REFERENCES medicines(id) ON DELETE SET NULL`
- `medicines_inventory_id_fkey`: `REFERENCES inventory_items(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medicines_pkey` | primary | `PRIMARY KEY` | [001_init.sql:749](../../../backend/migrations/001_init.sql#L749) |
| `ck_medicines_per_weight_strength` | check | `CONSTRAINT ck_medicines_per_weight_strength     CHECK (calculation_type = 'none' OR strength IS NOT NULL)` | [001_init.sql:3465](../../../backend/migrations/001_init.sql#L3465) |
| `ck_medicines_strength_positive` | check | `CONSTRAINT ck_medicines_strength_positive     CHECK (strength IS NULL OR strength > 0)` | [001_init.sql:3470](../../../backend/migrations/001_init.sql#L3470) |
| `ck_medicines_frequency_positive` | check | `CONSTRAINT ck_medicines_frequency_positive     CHECK (frequency_per_day IS NULL OR frequency_per_day > 0)` | [001_init.sql:3474](../../../backend/migrations/001_init.sql#L3474) |
| `ck_medicines_duration_positive` | check | `CONSTRAINT ck_medicines_duration_positive     CHECK (default_duration_days IS NULL OR default_duration_days > 0)` | [001_init.sql:3478](../../../backend/migrations/001_init.sql#L3478) |
| `uq_medicines_id_clinic` | unique | `CONSTRAINT uq_medicines_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:3484](../../../backend/migrations/001_init.sql#L3484) |
| `ck_medicines_positive_version` | check | `CONSTRAINT ck_medicines_positive_version CHECK (version > 0)` | [007_bill_discount_excluded_and_master_versions.sql:41](../../../backend/migrations/007_bill_discount_excluded_and_master_versions.sql#L41) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_medicines_clinic_id` | NO | — | `CREATE INDEX idx_medicines_clinic_id ON medicines(clinic_id)` | [001_init.sql:2224](../../../backend/migrations/001_init.sql#L2224) |
| `idx_medicines_parent_id` | NO | — | `CREATE INDEX idx_medicines_parent_id ON medicines(parent_id)` | [001_init.sql:2225](../../../backend/migrations/001_init.sql#L2225) |
| `idx_medicines_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medicines_clinic_sort         ON medicines(clinic_id, sort_order)               WHERE deleted_at IS NULL` | [001_init.sql:2262](../../../backend/migrations/001_init.sql#L2262) |
| `idx_medicines_inventory_id` | NO | `inventory_id IS NOT NULL` | `CREATE INDEX idx_medicines_inventory_id ON medicines(inventory_id) WHERE inventory_id IS NOT NULL` | [001_init.sql:2450](../../../backend/migrations/001_init.sql#L2450) |
| `idx_medicines_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_medicines_clinic_name ON medicines(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2460](../../../backend/migrations/001_init.sql#L2460) |
| `idx_medicines_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medicines_active ON medicines(id) WHERE deleted_at IS NULL` | [001_init.sql:2480](../../../backend/migrations/001_init.sql#L2480) |
| `idx_medicines_name_fold_trgm` | NO | — | `CREATE INDEX idx_medicines_name_fold_trgm   ON medicines USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6720](../../../backend/migrations/001_init.sql#L6720) |
