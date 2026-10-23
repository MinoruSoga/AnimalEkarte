<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `record_edit_rights`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9003](../../../backend/migrations/001_init.sql#L9003) |
| `medical_record_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9004](../../../backend/migrations/001_init.sql#L9004) |
| `epoch` | `bigint` | NOT NULL | `1` | — | [001_init.sql:9005](../../../backend/migrations/001_init.sql#L9005) |
| `state` | `text` | NOT NULL | — | — | [001_init.sql:9006](../../../backend/migrations/001_init.sql#L9006) |
| `holder_staff_id` | `bigint` | NULL | — | FK | [001_init.sql:9007](../../../backend/migrations/001_init.sql#L9007) |
| `session_ref` | `uuid` | NULL | — | — | [001_init.sql:9008](../../../backend/migrations/001_init.sql#L9008) |
| `acquired_at` | `timestamptz` | NULL | — | — | [001_init.sql:9009](../../../backend/migrations/001_init.sql#L9009) |
| `released_at` | `timestamptz` | NULL | — | — | [001_init.sql:9010](../../../backend/migrations/001_init.sql#L9010) |
| `last_seen_at` | `timestamptz` | NULL | — | — | [001_init.sql:9011](../../../backend/migrations/001_init.sql#L9011) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    record_edit_rights {
        bigint clinic_id PK,FK
        bigint medical_record_id PK,FK
        bigint epoch
        text state
        bigint holder_staff_id FK
        uuid session_ref
        timestamptz acquired_at
        timestamptz released_at
        timestamptz last_seen_at
    }
    clinics {
        BIGSERIAL id PK
    }
    medical_records {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||--o{ record_edit_rights : "clinic_id"
    staffs o|..o{ record_edit_rights : "holder_staff_id"
    medical_records ||--o| record_edit_rights : "medical_record_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `record_edit_rights_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9003](../../../backend/migrations/001_init.sql#L9003) |
| `record_edit_rights_holder_staff_id_fkey` | `holder_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9007](../../../backend/migrations/001_init.sql#L9007) |
| `record_edit_rights_medical_record_id_clinic_id_fkey` | `medical_record_id, clinic_id` | [medical_records](medical_records.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9015](../../../backend/migrations/001_init.sql#L9015) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `record_edit_rights_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `record_edit_rights_holder_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `record_edit_rights_medical_record_id_clinic_id_fkey`: `FOREIGN KEY (medical_record_id,clinic_id) REFERENCES medical_records(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `record_edit_rights_epoch_check` | check | `CHECK (epoch > 0)` | [001_init.sql:9005](../../../backend/migrations/001_init.sql#L9005) |
| `record_edit_rights_state_check` | check | `CHECK (state IN ('free','held'))` | [001_init.sql:9006](../../../backend/migrations/001_init.sql#L9006) |
| `record_edit_rights_pkey` | primary | `PRIMARY KEY (clinic_id,medical_record_id)` | [001_init.sql:9012](../../../backend/migrations/001_init.sql#L9012) |
| `record_edit_rights_check` | check | `CHECK ((state='free' AND holder_staff_id IS NULL AND session_ref IS NULL)       OR (state='held' AND holder_staff_id IS NOT NULL AND session_ref IS NOT NULL AND acquired_at IS NOT NULL))` | [001_init.sql:9013](../../../backend/migrations/001_init.sql#L9013) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
