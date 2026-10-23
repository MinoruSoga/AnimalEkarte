<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medicines`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:763](../../../backend/migrations/001_init.sql#L763) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:764](../../../backend/migrations/001_init.sql#L764) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:765](../../../backend/migrations/001_init.sql#L765) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:766](../../../backend/migrations/001_init.sql#L766) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:767](../../../backend/migrations/001_init.sql#L767) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:768](../../../backend/migrations/001_init.sql#L768) |
| `parent_id` | `bigint` | NULL | — | FK | [001_init.sql:769](../../../backend/migrations/001_init.sql#L769) |
| `dosage_form` | `dosage_form` | NULL | — | — | [001_init.sql:770](../../../backend/migrations/001_init.sql#L770) |
| `medicine_unit` | `medicine_unit` | NULL | — | — | [001_init.sql:771](../../../backend/migrations/001_init.sql#L771) |
| `inventory_id` | `bigint` | NULL | — | FK | [001_init.sql:772](../../../backend/migrations/001_init.sql#L772) |
| `default_quantity` | `numeric(10,2)` | NULL | `1` | — | [001_init.sql:3475](../../../backend/migrations/001_init.sql#L3475) |
| `tax_type` | `tax_type` | NOT NULL | `'excluded'` | — | [001_init.sql:774](../../../backend/migrations/001_init.sql#L774) |
| `tax_rate` | `numeric` | NOT NULL | `0.10` | — | [001_init.sql:775](../../../backend/migrations/001_init.sql#L775) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:776](../../../backend/migrations/001_init.sql#L776) |
| `is_non_insurance` | `boolean` | NOT NULL | `false` | — | [001_init.sql:777](../../../backend/migrations/001_init.sql#L777) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:778](../../../backend/migrations/001_init.sql#L778) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:779](../../../backend/migrations/001_init.sql#L779) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:780](../../../backend/migrations/001_init.sql#L780) |
| `calculation_type` | `medicine_calculation_type` | NOT NULL | `'none'` | — | [001_init.sql:3471](../../../backend/migrations/001_init.sql#L3471) |
| `strength` | `numeric(10,4)` | NULL | — | — | [001_init.sql:3472](../../../backend/migrations/001_init.sql#L3472) |
| `frequency_per_day` | `integer` | NULL | — | — | [001_init.sql:3473](../../../backend/migrations/001_init.sql#L3473) |
| `default_duration_days` | `integer` | NULL | — | — | [001_init.sql:3474](../../../backend/migrations/001_init.sql#L3474) |
| `is_bill_discount_excluded` | `boolean` | NULL | — | — | [001_init.sql:11625](../../../backend/migrations/001_init.sql#L11625) |
| `version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:11641](../../../backend/migrations/001_init.sql#L11641) |

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
| `medicines_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:764](../../../backend/migrations/001_init.sql#L764) |
| `medicines_parent_id_fkey` | `parent_id` | [medicines](medicines.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:769](../../../backend/migrations/001_init.sql#L769) |
| `medicines_inventory_id_fkey` | `inventory_id` | [inventory_items](inventory_items.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:772](../../../backend/migrations/001_init.sql#L772) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medicines_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medicines_parent_id_fkey`: `REFERENCES medicines(id) ON DELETE SET NULL`
- `medicines_inventory_id_fkey`: `REFERENCES inventory_items(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medicines_pkey` | primary | `PRIMARY KEY` | [001_init.sql:763](../../../backend/migrations/001_init.sql#L763) |
| `ck_medicines_per_weight_strength` | check | `CONSTRAINT ck_medicines_per_weight_strength     CHECK (calculation_type = 'none' OR strength IS NOT NULL)` | [001_init.sql:3479](../../../backend/migrations/001_init.sql#L3479) |
| `ck_medicines_strength_positive` | check | `CONSTRAINT ck_medicines_strength_positive     CHECK (strength IS NULL OR strength > 0)` | [001_init.sql:3484](../../../backend/migrations/001_init.sql#L3484) |
| `ck_medicines_frequency_positive` | check | `CONSTRAINT ck_medicines_frequency_positive     CHECK (frequency_per_day IS NULL OR frequency_per_day > 0)` | [001_init.sql:3488](../../../backend/migrations/001_init.sql#L3488) |
| `ck_medicines_duration_positive` | check | `CONSTRAINT ck_medicines_duration_positive     CHECK (default_duration_days IS NULL OR default_duration_days > 0)` | [001_init.sql:3492](../../../backend/migrations/001_init.sql#L3492) |
| `uq_medicines_id_clinic` | unique | `CONSTRAINT uq_medicines_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:3498](../../../backend/migrations/001_init.sql#L3498) |
| `ck_medicines_positive_version` | check | `CONSTRAINT ck_medicines_positive_version CHECK (version > 0)` | [001_init.sql:11642](../../../backend/migrations/001_init.sql#L11642) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_medicines_clinic_id` | NO | — | `CREATE INDEX idx_medicines_clinic_id ON medicines(clinic_id)` | [001_init.sql:2238](../../../backend/migrations/001_init.sql#L2238) |
| `idx_medicines_parent_id` | NO | — | `CREATE INDEX idx_medicines_parent_id ON medicines(parent_id)` | [001_init.sql:2239](../../../backend/migrations/001_init.sql#L2239) |
| `idx_medicines_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medicines_clinic_sort         ON medicines(clinic_id, sort_order)               WHERE deleted_at IS NULL` | [001_init.sql:2276](../../../backend/migrations/001_init.sql#L2276) |
| `idx_medicines_inventory_id` | NO | `inventory_id IS NOT NULL` | `CREATE INDEX idx_medicines_inventory_id ON medicines(inventory_id) WHERE inventory_id IS NOT NULL` | [001_init.sql:2464](../../../backend/migrations/001_init.sql#L2464) |
| `idx_medicines_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_medicines_clinic_name ON medicines(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2474](../../../backend/migrations/001_init.sql#L2474) |
| `idx_medicines_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_medicines_active ON medicines(id) WHERE deleted_at IS NULL` | [001_init.sql:2494](../../../backend/migrations/001_init.sql#L2494) |
| `idx_medicines_name_fold_trgm` | NO | — | `CREATE INDEX idx_medicines_name_fold_trgm   ON medicines USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6734](../../../backend/migrations/001_init.sql#L6734) |
