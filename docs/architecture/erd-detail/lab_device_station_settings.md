<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_device_station_settings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:6052](../../../backend/migrations/001_init.sql#L6052) |
| `wait_ttl_seconds` | `integer` | NOT NULL | `1800` | — | [001_init.sql:6053](../../../backend/migrations/001_init.sql#L6053) |
| `slots_json` | `jsonb` | NOT NULL | `'[]'::jsonb` | — | [001_init.sql:6054](../../../backend/migrations/001_init.sql#L6054) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:6055](../../../backend/migrations/001_init.sql#L6055) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:6056](../../../backend/migrations/001_init.sql#L6056) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lab_device_station_settings {
        bigint clinic_id PK,FK
        integer wait_ttl_seconds
        jsonb slots_json
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||--o| lab_device_station_settings : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lab_device_station_settings_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:6052](../../../backend/migrations/001_init.sql#L6052) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lab_device_station_settings_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_device_station_settings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:6052](../../../backend/migrations/001_init.sql#L6052) |
| `chk_lab_device_station_settings_ttl` | check | `CONSTRAINT chk_lab_device_station_settings_ttl         CHECK (wait_ttl_seconds >= 60 AND wait_ttl_seconds <= 86400)` | [001_init.sql:6057](../../../backend/migrations/001_init.sql#L6057) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
