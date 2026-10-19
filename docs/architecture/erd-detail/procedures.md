<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `procedures`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:928](../../../backend/migrations/001_init.sql#L928) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:929](../../../backend/migrations/001_init.sql#L929) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:930](../../../backend/migrations/001_init.sql#L930) |
| `price` | `bigint` | NULL | — | — | [001_init.sql:931](../../../backend/migrations/001_init.sql#L931) |
| `is_active` | `boolean` | NOT NULL | `true` | — | [001_init.sql:932](../../../backend/migrations/001_init.sql#L932) |
| `description` | `text` | NOT NULL | `''` | — | [001_init.sql:933](../../../backend/migrations/001_init.sql#L933) |
| `duration` | `integer` | NULL | — | — | [001_init.sql:934](../../../backend/migrations/001_init.sql#L934) |
| `anesthesia` | `anesthesia_type` | NULL | `'none'` | — | [001_init.sql:935](../../../backend/migrations/001_init.sql#L935) |
| `parent_id` | `bigint` | NULL | — | FK | [001_init.sql:936](../../../backend/migrations/001_init.sql#L936) |
| `tax_type` | `tax_type` | NOT NULL | `'excluded'` | — | [001_init.sql:937](../../../backend/migrations/001_init.sql#L937) |
| `tax_rate` | `numeric` | NOT NULL | `0.10` | — | [001_init.sql:938](../../../backend/migrations/001_init.sql#L938) |
| `sort_order` | `integer` | NULL | `0` | — | [001_init.sql:939](../../../backend/migrations/001_init.sql#L939) |
| `is_surgery` | `boolean` | NOT NULL | `false` | — | [001_init.sql:941](../../../backend/migrations/001_init.sql#L941) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:942](../../../backend/migrations/001_init.sql#L942) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:943](../../../backend/migrations/001_init.sql#L943) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:944](../../../backend/migrations/001_init.sql#L944) |
| `is_non_insurance` | `BOOLEAN` | NULL | — | — | [001_init.sql:7605](../../../backend/migrations/001_init.sql#L7605) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    procedures {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        bigint price
        boolean is_active
        text description
        integer duration
        anesthesia_type anesthesia
        bigint parent_id FK
        tax_type tax_type
        numeric tax_rate
        integer sort_order
        boolean is_surgery
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        BOOLEAN is_non_insurance
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ procedures : "clinic_id"
    procedures o|..o{ procedures : "parent_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `procedures_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:929](../../../backend/migrations/001_init.sql#L929) |
| `procedures_parent_id_fkey` | `parent_id` | [procedures](procedures.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:936](../../../backend/migrations/001_init.sql#L936) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `procedures_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `procedures_parent_id_fkey`: `REFERENCES procedures(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `procedures_pkey` | primary | `PRIMARY KEY` | [001_init.sql:928](../../../backend/migrations/001_init.sql#L928) |
| `uq_procedures_id_clinic` | unique | `CONSTRAINT uq_procedures_id_clinic UNIQUE (id, clinic_id)` | [001_init.sql:952](../../../backend/migrations/001_init.sql#L952) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_procedures_clinic_is_surgery` | NO | `is_surgery = true` | `CREATE INDEX IF NOT EXISTS idx_procedures_clinic_is_surgery     ON procedures(clinic_id, is_surgery)     WHERE is_surgery = true` | [001_init.sql:947](../../../backend/migrations/001_init.sql#L947) |
| `idx_procedures_clinic_id` | NO | — | `CREATE INDEX idx_procedures_clinic_id ON procedures(clinic_id)` | [001_init.sql:2231](../../../backend/migrations/001_init.sql#L2231) |
| `idx_procedures_parent_id` | NO | — | `CREATE INDEX idx_procedures_parent_id ON procedures(parent_id)` | [001_init.sql:2232](../../../backend/migrations/001_init.sql#L2232) |
| `idx_procedures_clinic_sort` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_procedures_clinic_sort        ON procedures(clinic_id, sort_order)              WHERE deleted_at IS NULL` | [001_init.sql:2264](../../../backend/migrations/001_init.sql#L2264) |
| `idx_procedures_clinic_name` | YES | `deleted_at IS NULL` | `CREATE UNIQUE INDEX idx_procedures_clinic_name ON procedures(clinic_id, name) WHERE deleted_at IS NULL` | [001_init.sql:2462](../../../backend/migrations/001_init.sql#L2462) |
| `idx_procedures_active` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_procedures_active ON procedures(id) WHERE deleted_at IS NULL` | [001_init.sql:2485](../../../backend/migrations/001_init.sql#L2485) |
| `idx_procedures_name_fold_trgm` | NO | — | `CREATE INDEX idx_procedures_name_fold_trgm   ON procedures USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6717](../../../backend/migrations/001_init.sql#L6717) |
