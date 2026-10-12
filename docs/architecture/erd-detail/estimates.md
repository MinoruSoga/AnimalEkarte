<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `estimates`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:1686](../../../backend/migrations/001_init.sql#L1686) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:1687](../../../backend/migrations/001_init.sql#L1687) |
| `estimate_no` | `text` | NOT NULL | — | UK | [001_init.sql:1688](../../../backend/migrations/001_init.sql#L1688) |
| `medical_record_id` | `bigint` | NULL | — | FK | [001_init.sql:1689](../../../backend/migrations/001_init.sql#L1689) |
| `title` | `text` | NOT NULL | `''` | — | [001_init.sql:1690](../../../backend/migrations/001_init.sql#L1690) |
| `owner_id` | `bigint` | NULL | — | FK | [001_init.sql:1691](../../../backend/migrations/001_init.sql#L1691) |
| `status` | `estimate_status` | NULL | `'draft'` | — | [001_init.sql:1692](../../../backend/migrations/001_init.sql#L1692) |
| `subtotal` | `bigint` | NOT NULL | `0` | — | [001_init.sql:1693](../../../backend/migrations/001_init.sql#L1693) |
| `tax_total` | `bigint` | NOT NULL | `0` | — | [001_init.sql:1694](../../../backend/migrations/001_init.sql#L1694) |
| `total_amount` | `bigint` | NOT NULL | `0` | — | [001_init.sql:1695](../../../backend/migrations/001_init.sql#L1695) |
| `insurance_amount` | `bigint` | NULL | `0` | — | [001_init.sql:1696](../../../backend/migrations/001_init.sql#L1696) |
| `discount_amount` | `bigint` | NULL | `0` | — | [001_init.sql:1697](../../../backend/migrations/001_init.sql#L1697) |
| `valid_until` | `date` | NULL | — | — | [001_init.sql:1698](../../../backend/migrations/001_init.sql#L1698) |
| `comment` | `text` | NOT NULL | `''` | — | [001_init.sql:1699](../../../backend/migrations/001_init.sql#L1699) |
| `notes` | `text` | NOT NULL | `''` | — | [001_init.sql:1700](../../../backend/migrations/001_init.sql#L1700) |
| `created_by` | `bigint` | NULL | — | FK | [001_init.sql:1701](../../../backend/migrations/001_init.sql#L1701) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1702](../../../backend/migrations/001_init.sql#L1702) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:1703](../../../backend/migrations/001_init.sql#L1703) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:1704](../../../backend/migrations/001_init.sql#L1704) |
| `supersedes_estimate_id` | `BIGINT` | NULL | — | FK | [001_init.sql:4831](../../../backend/migrations/001_init.sql#L4831) |
| `pet_id` | `bigint` | NULL | — | FK | [001_init.sql:5913](../../../backend/migrations/001_init.sql#L5913) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    estimates {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        text estimate_no UK
        bigint medical_record_id FK
        text title
        bigint owner_id FK
        estimate_status status
        bigint subtotal
        bigint tax_total
        bigint total_amount
        bigint insurance_amount
        bigint discount_amount
        date valid_until
        text comment
        text notes
        bigint created_by FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        BIGINT supersedes_estimate_id FK
        bigint pet_id FK
    }
    clinics {
        BIGSERIAL id PK
    }
    medical_records {
        BIGSERIAL id PK,UK
    }
    owners {
        BIGSERIAL id PK,UK
    }
    pets {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ estimates : "clinic_id"
    medical_records o|..o{ estimates : "medical_record_id"
    owners o|..o{ estimates : "owner_id"
    staffs o|..o{ estimates : "created_by"
    estimates o|..o{ estimates : "supersedes_estimate_id"
    pets o|..o{ estimates : "clinic_id, pet_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `estimates_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1687](../../../backend/migrations/001_init.sql#L1687) |
| `estimates_medical_record_id_fkey` | `medical_record_id` | [medical_records](medical_records.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:1689](../../../backend/migrations/001_init.sql#L1689) |
| `estimates_owner_id_fkey` | `owner_id` | [owners](owners.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1691](../../../backend/migrations/001_init.sql#L1691) |
| `estimates_created_by_fkey` | `created_by` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:1701](../../../backend/migrations/001_init.sql#L1701) |
| `estimates_supersedes_estimate_id_fkey` | `supersedes_estimate_id` | [estimates](estimates.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:4831](../../../backend/migrations/001_init.sql#L4831) |
| `fk_estimates_pet_clinic` | `clinic_id, pet_id` | [pets](pets.md) | `clinic_id, id` | `ON DELETE SET NULL (pet_id)` | [001_init.sql:5917](../../../backend/migrations/001_init.sql#L5917) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `estimates_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `estimates_medical_record_id_fkey`: `REFERENCES medical_records(id) ON DELETE RESTRICT`
- `estimates_owner_id_fkey`: `REFERENCES owners(id) ON DELETE SET NULL`
- `estimates_created_by_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`
- `estimates_supersedes_estimate_id_fkey`: `REFERENCES estimates(id)`
- `fk_estimates_pet_clinic`: `CONSTRAINT fk_estimates_pet_clinic   FOREIGN KEY (clinic_id, pet_id)   REFERENCES pets (clinic_id, id)   ON DELETE SET NULL (pet_id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `estimates_pkey` | primary | `PRIMARY KEY` | [001_init.sql:1686](../../../backend/migrations/001_init.sql#L1686) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_estimates_clinic_estimate_no` | YES | — | `CREATE UNIQUE INDEX idx_estimates_clinic_estimate_no ON estimates(clinic_id, estimate_no)` | [001_init.sql:2188](../../../backend/migrations/001_init.sql#L2188) |
| `idx_estimates_medical_record_id` | NO | — | `CREATE INDEX idx_estimates_medical_record_id ON estimates(medical_record_id)` | [001_init.sql:2326](../../../backend/migrations/001_init.sql#L2326) |
| `idx_estimates_status` | NO | — | `CREATE INDEX idx_estimates_status ON estimates(status)` | [001_init.sql:2327](../../../backend/migrations/001_init.sql#L2327) |
| `idx_estimates_owner_id` | NO | — | `CREATE INDEX idx_estimates_owner_id ON estimates(owner_id)` | [001_init.sql:2328](../../../backend/migrations/001_init.sql#L2328) |
| `idx_estimates_clinic_supersedes` | NO | `supersedes_estimate_id IS NOT NULL` | `CREATE INDEX IF NOT EXISTS idx_estimates_clinic_supersedes   ON estimates(clinic_id, supersedes_estimate_id)   WHERE supersedes_estimate_id IS NOT NULL` | [001_init.sql:4834](../../../backend/migrations/001_init.sql#L4834) |
| `idx_estimates_clinic_pet` | NO | `pet_id IS NOT NULL AND deleted_at IS NULL` | `CREATE INDEX IF NOT EXISTS idx_estimates_clinic_pet   ON estimates (clinic_id, pet_id)   WHERE pet_id IS NOT NULL AND deleted_at IS NULL` | [001_init.sql:5922](../../../backend/migrations/001_init.sql#L5922) |
