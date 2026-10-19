<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `financial_document_artifacts`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9718](../../../backend/migrations/001_init.sql#L9718) |
| `edition_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:9719](../../../backend/migrations/001_init.sql#L9719) |
| `renderer_version` | `text` | NOT NULL | — | — | [001_init.sql:9720](../../../backend/migrations/001_init.sql#L9720) |
| `object_key` | `text` | NOT NULL | — | — | [001_init.sql:9721](../../../backend/migrations/001_init.sql#L9721) |
| `content_sha256` | `bytea` | NOT NULL | — | — | [001_init.sql:9722](../../../backend/migrations/001_init.sql#L9722) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:9723](../../../backend/migrations/001_init.sql#L9723) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    financial_document_artifacts {
        bigint clinic_id PK,FK
        bigint edition_id PK,FK
        text renderer_version
        text object_key
        bytea content_sha256
        timestamptz created_at
    }
    clinics {
        BIGSERIAL id PK
    }
    financial_document_editions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ financial_document_artifacts : "clinic_id"
    financial_document_editions ||--o| financial_document_artifacts : "edition_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `financial_document_artifacts_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:9718](../../../backend/migrations/001_init.sql#L9718) |
| `financial_document_artifacts_edition_id_clinic_id_fkey` | `edition_id, clinic_id` | [financial_document_editions](financial_document_editions.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:9725](../../../backend/migrations/001_init.sql#L9725) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `financial_document_artifacts_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `financial_document_artifacts_edition_id_clinic_id_fkey`: `FOREIGN KEY (edition_id,clinic_id) REFERENCES financial_document_editions(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `financial_document_artifacts_object_key_check` | check | `CHECK (btrim(object_key)<>'')` | [001_init.sql:9721](../../../backend/migrations/001_init.sql#L9721) |
| `financial_document_artifacts_content_sha256_check` | check | `CHECK (octet_length(content_sha256)=32)` | [001_init.sql:9722](../../../backend/migrations/001_init.sql#L9722) |
| `financial_document_artifacts_pkey` | primary | `PRIMARY KEY (clinic_id,edition_id)` | [001_init.sql:9724](../../../backend/migrations/001_init.sql#L9724) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
