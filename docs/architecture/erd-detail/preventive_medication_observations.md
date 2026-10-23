<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `preventive_medication_observations`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:9223](../../../backend/migrations/001_init.sql#L9223) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9224](../../../backend/migrations/001_init.sql#L9224) |
| `prescription_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:9225](../../../backend/migrations/001_init.sql#L9225) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9226](../../../backend/migrations/001_init.sql#L9226) |
| `display_year` | `integer` | NOT NULL | — | — | [001_init.sql:10841](../../../backend/migrations/001_init.sql#L10841) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    preventive_medication_observations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint prescription_id FK,UK
        timestamptz created_at
        integer display_year
    }
    clinics {
        BIGSERIAL id PK
    }
    preventive_prescriptions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ preventive_medication_observations : "clinic_id"
    preventive_prescriptions ||..o{ preventive_medication_observations : "prescription_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `preventive_medication_observations_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9224](../../../backend/migrations/001_init.sql#L9224) |
| `preventive_medication_observatio_prescription_id_clinic_id_fkey` | `prescription_id, clinic_id` | [preventive_prescriptions](preventive_prescriptions.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9228](../../../backend/migrations/001_init.sql#L9228) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `preventive_medication_observations_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `preventive_medication_observatio_prescription_id_clinic_id_fkey`: `FOREIGN KEY (prescription_id,clinic_id) REFERENCES preventive_prescriptions(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `preventive_medication_observations_pkey` | primary | `PRIMARY KEY` | [001_init.sql:9223](../../../backend/migrations/001_init.sql#L9223) |
| `preventive_medication_observations_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:9227](../../../backend/migrations/001_init.sql#L9227) |
| `uq_medication_observation_prescription` | unique | `CONSTRAINT uq_medication_observation_prescription UNIQUE(id,clinic_id,prescription_id)` | [001_init.sql:10105](../../../backend/migrations/001_init.sql#L10105) |
| `preventive_medication_observations_display_year_check` | check | `CHECK(display_year BETWEEN 1 AND 9999)` | [001_init.sql:10841](../../../backend/migrations/001_init.sql#L10841) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
