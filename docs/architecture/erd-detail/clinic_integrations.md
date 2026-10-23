<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `clinic_integrations`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:224](../../../backend/migrations/001_init.sql#L224) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:225](../../../backend/migrations/001_init.sql#L225) |
| `service` | `varchar(50)` | NOT NULL | — | UK | [001_init.sql:226](../../../backend/migrations/001_init.sql#L226) |
| `key_name` | `varchar(100)` | NOT NULL | — | UK | [001_init.sql:227](../../../backend/migrations/001_init.sql#L227) |
| `key_value` | `text` | NOT NULL | — | — | [001_init.sql:228](../../../backend/migrations/001_init.sql#L228) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:229](../../../backend/migrations/001_init.sql#L229) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:230](../../../backend/migrations/001_init.sql#L230) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    clinic_integrations {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        varchar_50 service UK
        varchar_100 key_name UK
        text key_value
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o{ clinic_integrations : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_integrations_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE NO ACTION` | [001_init.sql:225](../../../backend/migrations/001_init.sql#L225) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `clinic_integrations_clinic_id_fkey`: `REFERENCES clinics(id)`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `clinic_integrations_pkey` | primary | `PRIMARY KEY` | [001_init.sql:224](../../../backend/migrations/001_init.sql#L224) |
| `clinic_integrations_clinic_id_service_key_name_key` | unique | `UNIQUE (clinic_id, service, key_name)` | [001_init.sql:231](../../../backend/migrations/001_init.sql#L231) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_clinic_integrations_clinic_service` | NO | — | `CREATE INDEX idx_clinic_integrations_clinic_service     ON clinic_integrations (clinic_id, service)` | [001_init.sql:234](../../../backend/migrations/001_init.sql#L234) |
