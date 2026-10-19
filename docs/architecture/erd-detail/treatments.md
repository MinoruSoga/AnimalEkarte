<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `treatments`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:1596](../../../backend/migrations/001_init.sql#L1596) |
| `medical_record_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:1597](../../../backend/migrations/001_init.sql#L1597) |
| `item_type` | `treatment_item_type` | NOT NULL | `'other'` | — | [001_init.sql:1598](../../../backend/migrations/001_init.sql#L1598) |
| `consultation_id` | `bigint` | NULL | — | FK | [001_init.sql:1599](../../../backend/migrations/001_init.sql#L1599) |
| `procedure_id` | `bigint` | NULL | — | FK | [001_init.sql:1600](../../../backend/migrations/001_init.sql#L1600) |
| `medicine_id` | `bigint` | NULL | — | FK | [001_init.sql:1601](../../../backend/migrations/001_init.sql#L1601) |
| `is_selected` | `boolean` | NULL | `false` | — | [001_init.sql:1602](../../../backend/migrations/001_init.sql#L1602) |
| `status` | `treatment_status` | NULL | `'pending'` | — | [001_init.sql:1603](../../../backend/migrations/001_init.sql#L1603) |
| `content` | `text` | NOT NULL | `''` | — | [001_init.sql:1604](../../../backend/migrations/001_init.sql#L1604) |
| `memo` | `text` | NOT NULL | `''` | — | [001_init.sql:1605](../../../backend/migrations/001_init.sql#L1605) |
| `admin_route` | `varchar(50)` | NOT NULL | `''` | — | [001_init.sql:1606](../../../backend/migrations/001_init.sql#L1606) |
| `is_insurance` | `boolean` | NULL | `false` | — | [001_init.sql:1607](../../../backend/migrations/001_init.sql#L1607) |
| `unit_price` | `bigint` | NULL | `0` | — | [001_init.sql:1608](../../../backend/migrations/001_init.sql#L1608) |
| `quantity` | `numeric(10,2)` | NULL | `1` | — | [001_init.sql:3558](../../../backend/migrations/001_init.sql#L3558) |
| `discount_rate` | `numeric(5,2)` | NULL | `0` | — | [001_init.sql:1610](../../../backend/migrations/001_init.sql#L1610) |
| `discount_amount` | `bigint` | NULL | `0` | — | [001_init.sql:1611](../../../backend/migrations/001_init.sql#L1611) |
| `inventory_id` | `bigint` | NULL | — | FK | [001_init.sql:1612](../../../backend/migrations/001_init.sql#L1612) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:1613](../../../backend/migrations/001_init.sql#L1613) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1614](../../../backend/migrations/001_init.sql#L1614) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1615](../../../backend/migrations/001_init.sql#L1615) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1616](../../../backend/migrations/001_init.sql#L1616) |
| `dose_weight_kg` | `numeric(6,2)` | NULL | — | — | [001_init.sql:3559](../../../backend/migrations/001_init.sql#L3559) |
| `dose_weight_source` | `varchar(255)` | NULL | — | — | [001_init.sql:3560](../../../backend/migrations/001_init.sql#L3560) |
| `dose_amount_mg` | `numeric(12,6)` | NULL | — | — | [001_init.sql:3561](../../../backend/migrations/001_init.sql#L3561) |
| `dose_amount_unit` | `text` | NULL | — | — | [001_init.sql:3562](../../../backend/migrations/001_init.sql#L3562) |
| `dose_param_snapshot` | `jsonb` | NULL | — | — | [001_init.sql:3563](../../../backend/migrations/001_init.sql#L3563) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:3569](../../../backend/migrations/001_init.sql#L3569) |
| `version` | `INTEGER` | NOT NULL | `1` | — | [001_init.sql:6333](../../../backend/migrations/001_init.sql#L6333) |
| `contract_version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:8284](../../../backend/migrations/001_init.sql#L8284) |
| `fulfillment_generation` | `bigint` | NOT NULL | `1` | — | [001_init.sql:8285](../../../backend/migrations/001_init.sql#L8285) |
| `administration_dose_amount` | `numeric(12,6)` | NULL | — | — | [001_init.sql:8829](../../../backend/migrations/001_init.sql#L8829) |
| `administration_dose_unit` | `text` | NULL | — | — | [001_init.sql:8830](../../../backend/migrations/001_init.sql#L8830) |
| `administration_instructions` | `text` | NULL | — | — | [001_init.sql:8831](../../../backend/migrations/001_init.sql#L8831) |
| `prescription_total_amount` | `numeric(12,6)` | NULL | — | — | [001_init.sql:8832](../../../backend/migrations/001_init.sql#L8832) |
| `prescription_total_unit` | `text` | NULL | — | — | [001_init.sql:8833](../../../backend/migrations/001_init.sql#L8833) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    treatments {
        BIGSERIAL id PK,UK
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
        bigint clinic_id FK,UK
        INTEGER version
        bigint contract_version
        bigint fulfillment_generation
        numeric_12_6 administration_dose_amount
        text administration_dose_unit
        text administration_instructions
        numeric_12_6 prescription_total_amount
        text prescription_total_unit
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
| `treatments_medical_record_id_fkey` | `medical_record_id` | [medical_records](medical_records.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1597](../../../backend/migrations/001_init.sql#L1597) |
| `treatments_consultation_id_fkey` | `consultation_id` | [consultations](consultations.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1599](../../../backend/migrations/001_init.sql#L1599) |
| `treatments_procedure_id_fkey` | `procedure_id` | [procedures](procedures.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1600](../../../backend/migrations/001_init.sql#L1600) |
| `treatments_medicine_id_fkey` | `medicine_id` | [medicines](medicines.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1601](../../../backend/migrations/001_init.sql#L1601) |
| `treatments_inventory_id_fkey` | `inventory_id` | [inventory_items](inventory_items.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1612](../../../backend/migrations/001_init.sql#L1612) |
| `fk_treatments_clinic_id` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:3572](../../../backend/migrations/001_init.sql#L3572) |
| `fk_treatments_medicine_clinic` | `medicine_id, clinic_id` | [medicines](medicines.md) | `id, clinic_id` | `ON DELETE SET NULL (medicine_id)` | [001_init.sql:3581](../../../backend/migrations/001_init.sql#L3581) |
| `fk_treatments_consultation_clinic` | `consultation_id, clinic_id` | [consultations](consultations.md) | `id, clinic_id` | `ON DELETE SET NULL (consultation_id)` | [001_init.sql:3587](../../../backend/migrations/001_init.sql#L3587) |
| `fk_treatments_procedure_clinic` | `procedure_id, clinic_id` | [procedures](procedures.md) | `id, clinic_id` | `ON DELETE SET NULL (procedure_id)` | [001_init.sql:3593](../../../backend/migrations/001_init.sql#L3593) |
| `fk_treatments_inventory_clinic` | `inventory_id, clinic_id` | [inventory_items](inventory_items.md) | `id, clinic_id` | `ON DELETE SET NULL (inventory_id)` | [001_init.sql:3599](../../../backend/migrations/001_init.sql#L3599) |

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
| `treatments_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1596](../../../backend/migrations/001_init.sql#L1596) |
| `chk_treatment_item_ref` | check | `CONSTRAINT chk_treatment_item_ref CHECK (         (item_type = 'consultation' AND procedure_id IS NULL AND medicine_id IS NULL) OR         (item_type = 'procedure'    AND consultation_id IS NULL AND medicine_id IS NULL) OR         (item_type = 'medicine'     AND consultation_id IS NULL AND procedure_id IS NULL) OR         (item_type = 'other'        AND consultation_id IS NULL AND procedure_id IS NULL AND medicine_id IS NULL)     )` | [001_init.sql:1617](../../../backend/migrations/001_init.sql#L1617) |
| `chk_treatment_quantity` | check | `CONSTRAINT chk_treatment_quantity CHECK (quantity > 0)` | [001_init.sql:1623](../../../backend/migrations/001_init.sql#L1623) |
| `uq_treatments_id_clinic` | unique | `CONSTRAINT uq_treatments_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:7927](../../../backend/migrations/001_init.sql#L7927) |
| `treatments_contract_version_check` | check | `CHECK (contract_version > 0)` | [001_init.sql:8284](../../../backend/migrations/001_init.sql#L8284) |
| `treatments_fulfillment_generation_check` | check | `CHECK (fulfillment_generation > 0)` | [001_init.sql:8285](../../../backend/migrations/001_init.sql#L8285) |
| `ck_administration_dose_pair` | check | `CONSTRAINT ck_administration_dose_pair CHECK (       (administration_dose_amount IS NULL AND administration_dose_unit IS NULL)       OR (administration_dose_amount IS NOT NULL AND administration_dose_unit IS NOT NULL AND btrim(administration_dose_unit)<>''))` | [001_init.sql:8834](../../../backend/migrations/001_init.sql#L8834) |
| `ck_prescription_total_pair` | check | `CONSTRAINT ck_prescription_total_pair CHECK (       (prescription_total_amount IS NULL AND prescription_total_unit IS NULL)       OR (prescription_total_amount IS NOT NULL AND prescription_total_unit IS NOT NULL AND btrim(prescription_total_unit)<>''))` | [001_init.sql:8837](../../../backend/migrations/001_init.sql#L8837) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_treatments_medical_record_id` | NO | — | `CREATE INDEX idx_treatments_medical_record_id ON treatments(medical_record_id)` | [001_init.sql:2285](../../../backend/migrations/001_init.sql#L2285) |
| `idx_treatments_consultation_id` | NO | `consultation_id IS NOT NULL` | `CREATE INDEX idx_treatments_consultation_id ON treatments(consultation_id) WHERE consultation_id IS NOT NULL` | [001_init.sql:2442](../../../backend/migrations/001_init.sql#L2442) |
| `idx_treatments_procedure_id` | NO | `procedure_id IS NOT NULL` | `CREATE INDEX idx_treatments_procedure_id ON treatments(procedure_id) WHERE procedure_id IS NOT NULL` | [001_init.sql:2443](../../../backend/migrations/001_init.sql#L2443) |
| `idx_treatments_medicine_id` | NO | `medicine_id IS NOT NULL` | `CREATE INDEX idx_treatments_medicine_id ON treatments(medicine_id) WHERE medicine_id IS NOT NULL` | [001_init.sql:2444](../../../backend/migrations/001_init.sql#L2444) |
| `idx_treatments_inventory_id` | NO | `inventory_id IS NOT NULL` | `CREATE INDEX idx_treatments_inventory_id ON treatments(inventory_id) WHERE inventory_id IS NOT NULL` | [001_init.sql:2445](../../../backend/migrations/001_init.sql#L2445) |
| `idx_treatments_clinic_id` | NO | — | `CREATE INDEX idx_treatments_clinic_id     ON treatments (clinic_id)` | [001_init.sql:3577](../../../backend/migrations/001_init.sql#L3577) |
| `idx_treatments_content_fold_trgm` | NO | — | `CREATE INDEX idx_treatments_content_fold_trgm   ON treatments USING gin ((translate(content, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6711](../../../backend/migrations/001_init.sql#L6711) |
| `idx_treatments_memo_fold_trgm` | NO | — | `CREATE INDEX idx_treatments_memo_fold_trgm   ON treatments USING gin ((translate(memo, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6714](../../../backend/migrations/001_init.sql#L6714) |
