<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `hotel_capacity_settings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9271](../../../backend/migrations/001_init.sql#L9271) |
| `max_concurrent_pets` | `integer` | NOT NULL | — | — | [001_init.sql:9272](../../../backend/migrations/001_init.sql#L9272) |
| `version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:9273](../../../backend/migrations/001_init.sql#L9273) |
| `updated_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:9274](../../../backend/migrations/001_init.sql#L9274) |
| `updated_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9275](../../../backend/migrations/001_init.sql#L9275) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    hotel_capacity_settings {
        bigint clinic_id PK,FK
        integer max_concurrent_pets
        bigint version
        bigint updated_by FK
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||--o| hotel_capacity_settings : "clinic_id"
    staffs ||..o{ hotel_capacity_settings : "updated_by"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `hotel_capacity_settings_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9271](../../../backend/migrations/001_init.sql#L9271) |
| `hotel_capacity_settings_updated_by_fkey` | `updated_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9274](../../../backend/migrations/001_init.sql#L9274) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `hotel_capacity_settings_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `hotel_capacity_settings_updated_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `hotel_capacity_settings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9271](../../../backend/migrations/001_init.sql#L9271) |
| `hotel_capacity_settings_max_concurrent_pets_check` | check | `CHECK (max_concurrent_pets > 0)` | [001_init.sql:9272](../../../backend/migrations/001_init.sql#L9272) |
| `hotel_capacity_settings_version_check` | check | `CHECK (version > 0)` | [001_init.sql:9273](../../../backend/migrations/001_init.sql#L9273) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
