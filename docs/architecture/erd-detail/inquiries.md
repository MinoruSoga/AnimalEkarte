<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `inquiries`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1551](../../../backend/migrations/001_init.sql#L1551) |
| `medical_record_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1552](../../../backend/migrations/001_init.sql#L1552) |
| `chief_complaint_type_id` | `bigint` | NULL | — | FK | [001_init.sql:1553](../../../backend/migrations/001_init.sql#L1553) |
| `chief_complaint` | `text` | NOT NULL | `''` | — | [001_init.sql:1554](../../../backend/migrations/001_init.sql#L1554) |
| `history` | `text` | NOT NULL | `''` | — | [001_init.sql:1555](../../../backend/migrations/001_init.sql#L1555) |
| `current_medications` | `text` | NOT NULL | `''` | — | [001_init.sql:1556](../../../backend/migrations/001_init.sql#L1556) |
| `allergy_info` | `text` | NOT NULL | `''` | — | [001_init.sql:1557](../../../backend/migrations/001_init.sql#L1557) |
| `last_meal` | `text` | NOT NULL | `''` | — | [001_init.sql:1558](../../../backend/migrations/001_init.sql#L1558) |
| `last_defecation` | `text` | NOT NULL | `''` | — | [001_init.sql:1559](../../../backend/migrations/001_init.sql#L1559) |
| `last_urination` | `text` | NOT NULL | `''` | — | [001_init.sql:1560](../../../backend/migrations/001_init.sql#L1560) |
| `appetite` | `appetite_level` | NULL | — | — | [001_init.sql:1561](../../../backend/migrations/001_init.sql#L1561) |
| `water_intake` | `water_intake_level` | NULL | — | — | [001_init.sql:1562](../../../backend/migrations/001_init.sql#L1562) |
| `owner_observations` | `text` | NOT NULL | `''` | — | [001_init.sql:1563](../../../backend/migrations/001_init.sql#L1563) |
| `notes` | `text` | NOT NULL | `''` | — | [001_init.sql:1564](../../../backend/migrations/001_init.sql#L1564) |
| `staff_id` | `bigint` | NULL | — | FK | [001_init.sql:1565](../../../backend/migrations/001_init.sql#L1565) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1566](../../../backend/migrations/001_init.sql#L1566) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1567](../../../backend/migrations/001_init.sql#L1567) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    inquiries {
        BIGSERIAL id PK
        bigint medical_record_id FK,UK
        bigint chief_complaint_type_id FK
        text chief_complaint
        text history
        text current_medications
        text allergy_info
        text last_meal
        text last_defecation
        text last_urination
        appetite_level appetite
        water_intake_level water_intake
        text owner_observations
        text notes
        bigint staff_id FK
        timestamptz created_at
        timestamptz updated_at
    }
    chief_complaint_types {
        BIGSERIAL id PK
    }
    medical_records {
        BIGSERIAL id PK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    medical_records ||..o| inquiries : "medical_record_id"
    chief_complaint_types o|..o{ inquiries : "chief_complaint_type_id"
    staffs o|..o{ inquiries : "staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `inquiries_medical_record_id_fkey` | `medical_record_id` | [medical_records](medical_records.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:1552](../../../backend/migrations/001_init.sql#L1552) |
| `inquiries_chief_complaint_type_id_fkey` | `chief_complaint_type_id` | [chief_complaint_types](chief_complaint_types.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1553](../../../backend/migrations/001_init.sql#L1553) |
| `inquiries_staff_id_fkey` | `staff_id` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1565](../../../backend/migrations/001_init.sql#L1565) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `inquiries_medical_record_id_fkey`: `REFERENCES medical_records(id) ON DELETE CASCADE`
- `inquiries_chief_complaint_type_id_fkey`: `REFERENCES chief_complaint_types(id) ON DELETE SET NULL`
- `inquiries_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `inquiries_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1551](../../../backend/migrations/001_init.sql#L1551) |
| `inquiries_medical_record_id_key` | unique | `UNIQUE` | [001_init.sql:1552](../../../backend/migrations/001_init.sql#L1552) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_inquiries_medical_record_id` | NO | — | `CREATE INDEX idx_inquiries_medical_record_id ON inquiries(medical_record_id)` | [001_init.sql:2301](../../../backend/migrations/001_init.sql#L2301) |
| `idx_inquiries_chief_complaint_fold_trgm` | NO | — | `CREATE INDEX idx_inquiries_chief_complaint_fold_trgm   ON inquiries USING gin ((translate(chief_complaint, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6708](../../../backend/migrations/001_init.sql#L6708) |
