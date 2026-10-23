<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `vaccination_certificate_editions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:8343](../../../backend/migrations/001_init.sql#L8343) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8344](../../../backend/migrations/001_init.sql#L8344) |
| `series_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:8345](../../../backend/migrations/001_init.sql#L8345) |
| `edition` | `bigint` | NOT NULL | — | UK | [001_init.sql:8346](../../../backend/migrations/001_init.sql#L8346) |
| `source_vaccination_version` | `bigint` | NOT NULL | — | — | [001_init.sql:8347](../../../backend/migrations/001_init.sql#L8347) |
| `snapshot` | `jsonb` | NOT NULL | — | — | [001_init.sql:8348](../../../backend/migrations/001_init.sql#L8348) |
| `previous_edition_id` | `bigint` | NULL | — | FK | [001_init.sql:8349](../../../backend/migrations/001_init.sql#L8349) |
| `issued_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:8350](../../../backend/migrations/001_init.sql#L8350) |
| `issued_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8351](../../../backend/migrations/001_init.sql#L8351) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    vaccination_certificate_editions {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint series_id FK,UK
        bigint edition UK
        bigint source_vaccination_version
        jsonb snapshot
        bigint previous_edition_id FK
        bigint issued_by FK
        timestamptz issued_at
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    vaccination_certificate_series {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ vaccination_certificate_editions : "clinic_id"
    staffs ||..o{ vaccination_certificate_editions : "issued_by"
    vaccination_certificate_series ||..o{ vaccination_certificate_editions : "series_id, clinic_id"
    vaccination_certificate_editions o|..o{ vaccination_certificate_editions : "previous_edition_id, clinic_id, series_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `vaccination_certificate_editions_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8344](../../../backend/migrations/001_init.sql#L8344) |
| `vaccination_certificate_editions_issued_by_fkey` | `issued_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8350](../../../backend/migrations/001_init.sql#L8350) |
| `vaccination_certificate_editions_series_id_clinic_id_fkey` | `series_id, clinic_id` | [vaccination_certificate_series](vaccination_certificate_series.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8355](../../../backend/migrations/001_init.sql#L8355) |
| `vaccination_certificate_editi_previous_edition_id_clinic_i_fkey` | `previous_edition_id, clinic_id, series_id` | [vaccination_certificate_editions](vaccination_certificate_editions.md) | `id, clinic_id, series_id` | `ON DELETE RESTRICT` | [001_init.sql:8356](../../../backend/migrations/001_init.sql#L8356) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `vaccination_certificate_editions_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `vaccination_certificate_editions_issued_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `vaccination_certificate_editions_series_id_clinic_id_fkey`: `FOREIGN KEY (series_id,clinic_id) REFERENCES vaccination_certificate_series(id,clinic_id) ON DELETE RESTRICT`
- `vaccination_certificate_editi_previous_edition_id_clinic_i_fkey`: `FOREIGN KEY (previous_edition_id,clinic_id,series_id) REFERENCES vaccination_certificate_editions(id,clinic_id,series_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `vaccination_certificate_editions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:8343](../../../backend/migrations/001_init.sql#L8343) |
| `vaccination_certificate_editions_edition_check` | check | `CHECK (edition > 0)` | [001_init.sql:8346](../../../backend/migrations/001_init.sql#L8346) |
| `vaccination_certificate_editio_source_vaccination_version_check` | check | `CHECK (source_vaccination_version > 0)` | [001_init.sql:8347](../../../backend/migrations/001_init.sql#L8347) |
| `vaccination_certificate_editions_snapshot_check` | check | `CHECK (jsonb_typeof(snapshot)='object')` | [001_init.sql:8348](../../../backend/migrations/001_init.sql#L8348) |
| `vaccination_certificate_editions_id_clinic_id_key` | unique | `UNIQUE (id,clinic_id)` | [001_init.sql:8352](../../../backend/migrations/001_init.sql#L8352) |
| `vaccination_certificate_editions_id_clinic_id_series_id_key` | unique | `UNIQUE (id,clinic_id,series_id)` | [001_init.sql:8353](../../../backend/migrations/001_init.sql#L8353) |
| `vaccination_certificate_edition_clinic_id_series_id_edition_key` | unique | `UNIQUE (clinic_id,series_id,edition)` | [001_init.sql:8354](../../../backend/migrations/001_init.sql#L8354) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
