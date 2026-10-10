<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `treatments`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1588](../../../backend/migrations/001_init.sql#L1588) |
| `medical_record_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1589](../../../backend/migrations/001_init.sql#L1589) |
| `item_type` | `treatment_item_type` | NOT NULL | `'other'` | — | [001_init.sql:1590](../../../backend/migrations/001_init.sql#L1590) |
| `consultation_id` | `bigint` | NULL | — | FK | [001_init.sql:1591](../../../backend/migrations/001_init.sql#L1591) |
| `procedure_id` | `bigint` | NULL | — | FK | [001_init.sql:1592](../../../backend/migrations/001_init.sql#L1592) |
| `medicine_id` | `bigint` | NULL | — | FK | [001_init.sql:1593](../../../backend/migrations/001_init.sql#L1593) |
| `is_selected` | `boolean` | NULL | `false` | — | [001_init.sql:1594](../../../backend/migrations/001_init.sql#L1594) |
| `status` | `treatment_status` | NULL | `'pending'` | — | [001_init.sql:1595](../../../backend/migrations/001_init.sql#L1595) |
| `content` | `text` | NOT NULL | `''` | — | [001_init.sql:1596](../../../backend/migrations/001_init.sql#L1596) |
| `memo` | `text` | NOT NULL | `''` | — | [001_init.sql:1597](../../../backend/migrations/001_init.sql#L1597) |
| `admin_route` | `varchar(50)` | NOT NULL | `''` | — | [001_init.sql:1598](../../../backend/migrations/001_init.sql#L1598) |
| `is_insurance` | `boolean` | NULL | `false` | — | [001_init.sql:1599](../../../backend/migrations/001_init.sql#L1599) |
| `unit_price` | `bigint` | NULL | `0` | — | [001_init.sql:1600](../../../backend/migrations/001_init.sql#L1600) |
| `quantity` | `numeric(10,2)` | NULL | `1` | — | [001_init.sql:3550](../../../backend/migrations/001_init.sql#L3550) |
| `discount_rate` | `numeric(5,2)` | NULL | `0` | — | [001_init.sql:1602](../../../backend/migrations/001_init.sql#L1602) |
| `discount_amount` | `bigint` | NULL | `0` | — | [001_init.sql:1603](../../../backend/migrations/001_init.sql#L1603) |
| `inventory_id` | `bigint` | NULL | — | FK | [001_init.sql:1604](../../../backend/migrations/001_init.sql#L1604) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1605](../../../backend/migrations/001_init.sql#L1605) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1606](../../../backend/migrations/001_init.sql#L1606) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1607](../../../backend/migrations/001_init.sql#L1607) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1608](../../../backend/migrations/001_init.sql#L1608) |
| `dose_weight_kg` | `numeric(6,2)` | NULL | — | — | [001_init.sql:3551](../../../backend/migrations/001_init.sql#L3551) |
| `dose_weight_source` | `varchar(255)` | NULL | — | — | [001_init.sql:3552](../../../backend/migrations/001_init.sql#L3552) |
| `dose_amount_mg` | `numeric(12,6)` | NULL | — | — | [001_init.sql:3553](../../../backend/migrations/001_init.sql#L3553) |
| `dose_amount_unit` | `text` | NULL | — | — | [001_init.sql:3554](../../../backend/migrations/001_init.sql#L3554) |
| `dose_param_snapshot` | `jsonb` | NULL | — | — | [001_init.sql:3555](../../../backend/migrations/001_init.sql#L3555) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:3561](../../../backend/migrations/001_init.sql#L3561) |
| `version` | `INTEGER` | NOT NULL | `1` | — | [005_child_records_version.sql:10](../../../backend/migrations/005_child_records_version.sql#L10) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    treatments {
        BIGSERIAL id PK
        bigint medical_record_id FK
        treatment_item_type item_type
        bigint consultation_id FK
        bigint procedure_id FK
        bigint medicine_id FK
        boolean is_selected
        treatment_status status
        text content
        text memo
        varchar_50 admin_route
        boolean is_insurance
        bigint unit_price
        numeric_10_2 quantity
        numeric_5_2 discount_rate
        bigint discount_amount
        bigint inventory_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        numeric_6_2 dose_weight_kg
        varchar_255 dose_weight_source
        numeric_12_6 dose_amount_mg
        text dose_amount_unit
        jsonb dose_param_snapshot
        bigint clinic_id FK
        INTEGER version
    }
    clinics {
        BIGSERIAL id PK
    }
    consultations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    inventory_items {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    medical_records {
        BIGSERIAL id PK,UK
    }
    medicines {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    procedures {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    medical_records ||..o{ treatments : "medical_record_id"
    consultations o|..o{ treatments : "consultation_id"
    procedures o|..o{ treatments : "procedure_id"
    medicines o|..o{ treatments : "medicine_id"
    inventory_items o|..o{ treatments : "inventory_id"
    clinics ||..o{ treatments : "clinic_id"
    medicines o|..o{ treatments : "medicine_id, clinic_id"
    consultations o|..o{ treatments : "consultation_id, clinic_id"
    procedures o|..o{ treatments : "procedure_id, clinic_id"
    inventory_items o|..o{ treatments : "inventory_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `treatments_medical_record_id_fkey` | `medical_record_id` | [medical_records](medical_records.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1589](../../../backend/migrations/001_init.sql#L1589) |
| `treatments_consultation_id_fkey` | `consultation_id` | [consultations](consultations.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1591](../../../backend/migrations/001_init.sql#L1591) |
| `treatments_procedure_id_fkey` | `procedure_id` | [procedures](procedures.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1592](../../../backend/migrations/001_init.sql#L1592) |
| `treatments_medicine_id_fkey` | `medicine_id` | [medicines](medicines.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1593](../../../backend/migrations/001_init.sql#L1593) |
| `treatments_inventory_id_fkey` | `inventory_id` | [inventory_items](inventory_items.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1604](../../../backend/migrations/001_init.sql#L1604) |
| `fk_treatments_clinic_id` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:3564](../../../backend/migrations/001_init.sql#L3564) |
| `fk_treatments_medicine_clinic` | `medicine_id, clinic_id` | [medicines](medicines.md) | `id, clinic_id` | `ON DELETE SET NULL (medicine_id)` | [001_init.sql:3573](../../../backend/migrations/001_init.sql#L3573) |
| `fk_treatments_consultation_clinic` | `consultation_id, clinic_id` | [consultations](consultations.md) | `id, clinic_id` | `ON DELETE SET NULL (consultation_id)` | [001_init.sql:3579](../../../backend/migrations/001_init.sql#L3579) |
| `fk_treatments_procedure_clinic` | `procedure_id, clinic_id` | [procedures](procedures.md) | `id, clinic_id` | `ON DELETE SET NULL (procedure_id)` | [001_init.sql:3585](../../../backend/migrations/001_init.sql#L3585) |
| `fk_treatments_inventory_clinic` | `inventory_id, clinic_id` | [inventory_items](inventory_items.md) | `id, clinic_id` | `ON DELETE SET NULL (inventory_id)` | [001_init.sql:3591](../../../backend/migrations/001_init.sql#L3591) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `treatments_medical_record_id_fkey`: `REFERENCES medical_records(id) ON DELETE CASCADE`
- `treatments_consultation_id_fkey`: `REFERENCES consultations(id) ON DELETE SET NULL`
- `treatments_procedure_id_fkey`: `REFERENCES procedures(id) ON DELETE SET NULL`
- `treatments_medicine_id_fkey`: `REFERENCES medicines(id) ON DELETE SET NULL`
- `treatments_inventory_id_fkey`: `REFERENCES inventory_items(id) ON DELETE SET NULL`
- `fk_treatments_clinic_id`: `CONSTRAINT fk_treatments_clinic_id     FOREIGN KEY (clinic_id)     REFERENCES clinics (id)     ON DELETE RESTRICT`
- `fk_treatments_medicine_clinic`: `CONSTRAINT fk_treatments_medicine_clinic     FOREIGN KEY (medicine_id, clinic_id)     REFERENCES medicines (id, clinic_id)     ON DELETE SET NULL (medicine_id)`
- `fk_treatments_consultation_clinic`: `CONSTRAINT fk_treatments_consultation_clinic     FOREIGN KEY (consultation_id, clinic_id)     REFERENCES consultations (id, clinic_id)     ON DELETE SET NULL (consultation_id)`
- `fk_treatments_procedure_clinic`: `CONSTRAINT fk_treatments_procedure_clinic     FOREIGN KEY (procedure_id, clinic_id)     REFERENCES procedures (id, clinic_id)     ON DELETE SET NULL (procedure_id)`
- `fk_treatments_inventory_clinic`: `CONSTRAINT fk_treatments_inventory_clinic     FOREIGN KEY (inventory_id, clinic_id)     REFERENCES inventory_items (id, clinic_id)     ON DELETE SET NULL (inventory_id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `treatments_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1588](../../../backend/migrations/001_init.sql#L1588) |
| `chk_treatment_item_ref` | check | `CONSTRAINT chk_treatment_item_ref CHECK (         (item_type = 'consultation' AND procedure_id IS NULL AND medicine_id IS NULL) OR         (item_type = 'procedure'    AND consultation_id IS NULL AND medicine_id IS NULL) OR         (item_type = 'medicine'     AND consultation_id IS NULL AND procedure_id IS NULL) OR         (item_type = 'other'        AND consultation_id IS NULL AND procedure_id IS NULL AND medicine_id IS NULL)     )` | [001_init.sql:1609](../../../backend/migrations/001_init.sql#L1609) |
| `chk_treatment_quantity` | check | `CONSTRAINT chk_treatment_quantity CHECK (quantity > 0)` | [001_init.sql:1615](../../../backend/migrations/001_init.sql#L1615) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_treatments_medical_record_id` | NO | — | `CREATE INDEX idx_treatments_medical_record_id ON treatments(medical_record_id)` | [001_init.sql:2277](../../../backend/migrations/001_init.sql#L2277) |
| `idx_treatments_consultation_id` | NO | `consultation_id IS NOT NULL` | `CREATE INDEX idx_treatments_consultation_id ON treatments(consultation_id) WHERE consultation_id IS NOT NULL` | [001_init.sql:2434](../../../backend/migrations/001_init.sql#L2434) |
| `idx_treatments_procedure_id` | NO | `procedure_id IS NOT NULL` | `CREATE INDEX idx_treatments_procedure_id ON treatments(procedure_id) WHERE procedure_id IS NOT NULL` | [001_init.sql:2435](../../../backend/migrations/001_init.sql#L2435) |
| `idx_treatments_medicine_id` | NO | `medicine_id IS NOT NULL` | `CREATE INDEX idx_treatments_medicine_id ON treatments(medicine_id) WHERE medicine_id IS NOT NULL` | [001_init.sql:2436](../../../backend/migrations/001_init.sql#L2436) |
| `idx_treatments_inventory_id` | NO | `inventory_id IS NOT NULL` | `CREATE INDEX idx_treatments_inventory_id ON treatments(inventory_id) WHERE inventory_id IS NOT NULL` | [001_init.sql:2437](../../../backend/migrations/001_init.sql#L2437) |
| `idx_treatments_clinic_id` | NO | — | `CREATE INDEX idx_treatments_clinic_id     ON treatments (clinic_id)` | [001_init.sql:3569](../../../backend/migrations/001_init.sql#L3569) |
| `idx_treatments_content_fold_trgm` | NO | — | `CREATE INDEX idx_treatments_content_fold_trgm   ON treatments USING gin ((translate(content, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [015_search_fold_indexes.sql:63](../../../backend/migrations/015_search_fold_indexes.sql#L63) |
| `idx_treatments_memo_fold_trgm` | NO | — | `CREATE INDEX idx_treatments_memo_fold_trgm   ON treatments USING gin ((translate(memo, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [015_search_fold_indexes.sql:66](../../../backend/migrations/015_search_fold_indexes.sql#L66) |
