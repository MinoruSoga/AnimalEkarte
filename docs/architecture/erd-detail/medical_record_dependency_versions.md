<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `medical_record_dependency_versions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9060](../../../backend/migrations/001_init.sql#L9060) |
| `medical_record_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9061](../../../backend/migrations/001_init.sql#L9061) |
| `owner_kind` | `text` | NOT NULL | — | PK | [001_init.sql:9062](../../../backend/migrations/001_init.sql#L9062) |
| `revision` | `bigint` | NOT NULL | — | — | [001_init.sql:9063](../../../backend/migrations/001_init.sql#L9063) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    medical_record_dependency_versions {
        bigint clinic_id PK,FK
        bigint medical_record_id PK,FK
        text owner_kind PK
        bigint revision
    }
    clinics {
        BIGSERIAL id PK
    }
    medical_records {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ medical_record_dependency_versions : "clinic_id"
    medical_records ||--o{ medical_record_dependency_versions : "medical_record_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `medical_record_dependency_versions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9060](../../../backend/migrations/001_init.sql#L9060) |
| `medical_record_dependency_vers_medical_record_id_clinic_id_fkey` | `medical_record_id, clinic_id` | [medical_records](medical_records.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9065](../../../backend/migrations/001_init.sql#L9065) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `medical_record_dependency_versions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `medical_record_dependency_vers_medical_record_id_clinic_id_fkey`: `FOREIGN KEY (medical_record_id,clinic_id) REFERENCES medical_records(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `medical_record_dependency_versions_revision_check` | check | `CHECK (revision > 0)` | [001_init.sql:9063](../../../backend/migrations/001_init.sql#L9063) |
| `medical_record_dependency_versions_pkey` | primary | `PRIMARY KEY (clinic_id,medical_record_id,owner_kind)` | [001_init.sql:9064](../../../backend/migrations/001_init.sql#L9064) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
