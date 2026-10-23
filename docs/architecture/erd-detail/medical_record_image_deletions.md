<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medical_record_image_deletions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:6916](../../../backend/migrations/001_init.sql#L6916) |
| `clinic_id` | `BIGINT` | NOT NULL | — | FK, UK | [001_init.sql:6917](../../../backend/migrations/001_init.sql#L6917) |
| `medical_record_id` | `BIGINT` | NOT NULL | — | — | [001_init.sql:6918](../../../backend/migrations/001_init.sql#L6918) |
| `image_id` | `BIGINT` | NOT NULL | — | UK | [001_init.sql:6919](../../../backend/migrations/001_init.sql#L6919) |
| `image_object_key` | `TEXT` | NOT NULL | `''` | — | [001_init.sql:6920](../../../backend/migrations/001_init.sql#L6920) |
| `thumbnail_object_key` | `TEXT` | NOT NULL | `''` | — | [001_init.sql:6921](../../../backend/migrations/001_init.sql#L6921) |
| `storage_status` | `TEXT` | NOT NULL | — | — | [001_init.sql:6922](../../../backend/migrations/001_init.sql#L6922) |
| `attempt_count` | `INTEGER` | NOT NULL | `0` | — | [001_init.sql:6923](../../../backend/migrations/001_init.sql#L6923) |
| `last_error` | `TEXT` | NOT NULL | `''` | — | [001_init.sql:6924](../../../backend/migrations/001_init.sql#L6924) |
| `actor_staff_id` | `BIGINT` | NOT NULL | — | — | [001_init.sql:6925](../../../backend/migrations/001_init.sql#L6925) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | — | [001_init.sql:6926](../../../backend/migrations/001_init.sql#L6926) |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | — | [001_init.sql:6927](../../../backend/migrations/001_init.sql#L6927) |
| `deleted_at` | `TIMESTAMPTZ` | NULL | — | — | [001_init.sql:6928](../../../backend/migrations/001_init.sql#L6928) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medical_record_image_deletions {
        BIGSERIAL id PK
        BIGINT clinic_id FK,UK
        BIGINT medical_record_id
        BIGINT image_id UK
        TEXT image_object_key
        TEXT thumbnail_object_key
        TEXT storage_status
        INTEGER attempt_count
        TEXT last_error
        BIGINT actor_staff_id
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ medical_record_image_deletions : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medical_record_image_deletions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:6917](../../../backend/migrations/001_init.sql#L6917) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medical_record_image_deletions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medical_record_image_deletions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:6916](../../../backend/migrations/001_init.sql#L6916) |
| `medical_record_image_deletions_status_check` | check | `CONSTRAINT medical_record_image_deletions_status_check         CHECK (storage_status IN ('pending_storage', 'storage_deleted', 'retained_shared'))` | [001_init.sql:6929](../../../backend/migrations/001_init.sql#L6929) |
| `medical_record_image_deletions_clinic_image_uidx` | unique | `CONSTRAINT medical_record_image_deletions_clinic_image_uidx         UNIQUE (clinic_id, image_id)` | [001_init.sql:6931](../../../backend/migrations/001_init.sql#L6931) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_medical_record_image_deletions_clinic_record` | NO | — | `CREATE INDEX idx_medical_record_image_deletions_clinic_record     ON medical_record_image_deletions (clinic_id, medical_record_id)` | [001_init.sql:6935](../../../backend/migrations/001_init.sql#L6935) |
