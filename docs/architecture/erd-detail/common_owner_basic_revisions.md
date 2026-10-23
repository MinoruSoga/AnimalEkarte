<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `common_owner_basic_revisions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10758](../../../backend/migrations/001_init.sql#L10758) |
| `common_owner_id` | `uuid` | NOT NULL | — | FK, UK | [001_init.sql:10758](../../../backend/migrations/001_init.sql#L10758) |
| `revision` | `bigint` | NOT NULL | — | UK | [001_init.sql:10759](../../../backend/migrations/001_init.sql#L10759) |
| `source_clinic_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10759](../../../backend/migrations/001_init.sql#L10759) |
| `source_local_owner_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10760](../../../backend/migrations/001_init.sql#L10760) |
| `operation_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:10760](../../../backend/migrations/001_init.sql#L10760) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:10761](../../../backend/migrations/001_init.sql#L10761) |
| `name_kana` | `text` | NULL | — | — | [001_init.sql:10761](../../../backend/migrations/001_init.sql#L10761) |
| `birth_date` | `date` | NULL | — | — | [001_init.sql:10761](../../../backend/migrations/001_init.sql#L10761) |
| `company` | `text` | NULL | — | — | [001_init.sql:10761](../../../backend/migrations/001_init.sql#L10761) |
| `postal_code` | `text` | NULL | — | — | [001_init.sql:10762](../../../backend/migrations/001_init.sql#L10762) |
| `address1` | `text` | NULL | — | — | [001_init.sql:10762](../../../backend/migrations/001_init.sql#L10762) |
| `address2` | `text` | NULL | — | — | [001_init.sql:10762](../../../backend/migrations/001_init.sql#L10762) |
| `home_postal_code` | `text` | NULL | — | — | [001_init.sql:10762](../../../backend/migrations/001_init.sql#L10762) |
| `home_address1` | `text` | NULL | — | — | [001_init.sql:10762](../../../backend/migrations/001_init.sql#L10762) |
| `home_address2` | `text` | NULL | — | — | [001_init.sql:10762](../../../backend/migrations/001_init.sql#L10762) |
| `phone` | `text` | NULL | — | — | [001_init.sql:10763](../../../backend/migrations/001_init.sql#L10763) |
| `company_phone` | `text` | NULL | — | — | [001_init.sql:10763](../../../backend/migrations/001_init.sql#L10763) |
| `email` | `text` | NULL | — | — | [001_init.sql:10763](../../../backend/migrations/001_init.sql#L10763) |
| `cygram_result` | `jsonb` | NULL | — | — | [001_init.sql:10764](../../../backend/migrations/001_init.sql#L10764) |
| `cygram_rule_version` | `text` | NULL | — | — | [001_init.sql:10764](../../../backend/migrations/001_init.sql#L10764) |
| `provenance` | `jsonb` | NOT NULL | — | — | [001_init.sql:10766](../../../backend/migrations/001_init.sql#L10766) |
| `recorded_by` | `bigint` | NOT NULL | — | FK | [001_init.sql:10766](../../../backend/migrations/001_init.sql#L10766) |
| `recorded_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:10767](../../../backend/migrations/001_init.sql#L10767) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    common_owner_basic_revisions {
        BIGSERIAL id PK,UK
        uuid common_owner_id FK,UK
        bigint revision UK
        bigint source_clinic_id FK
        bigint source_local_owner_id FK
        bigint operation_id FK
        text name
        text name_kana
        date birth_date
        text company
        text postal_code
        text address1
        text address2
        text home_postal_code
        text home_address1
        text home_address2
        text phone
        text company_phone
        text email
        jsonb cygram_result
        text cygram_rule_version
        jsonb provenance
        bigint recorded_by FK
        timestamptz recorded_at
    }
    clinics {
        BIGSERIAL id PK
    }
    common_owner_subjects {
        uuid id PK,FK
    }
    operation_bindings {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    common_owner_subjects ||..o{ common_owner_basic_revisions : "common_owner_id"
    clinics ||..o{ common_owner_basic_revisions : "source_clinic_id"
    staffs ||..o{ common_owner_basic_revisions : "recorded_by"
    owners ||..o{ common_owner_basic_revisions : "source_local_owner_id, source_clinic_id"
    operation_bindings ||..o{ common_owner_basic_revisions : "operation_id, source_clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `common_owner_basic_revisions_common_owner_id_fkey` | `common_owner_id` | [common_owner_subjects](common_owner_subjects.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10758](../../../backend/migrations/001_init.sql#L10758) |
| `common_owner_basic_revisions_source_clinic_id_fkey` | `source_clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10759](../../../backend/migrations/001_init.sql#L10759) |
| `common_owner_basic_revisions_recorded_by_fkey` | `recorded_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10766](../../../backend/migrations/001_init.sql#L10766) |
| `common_owner_basic_revisions_source_local_owner_id_source__fkey` | `source_local_owner_id, source_clinic_id` | [owners](owners.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10768](../../../backend/migrations/001_init.sql#L10768) |
| `common_owner_basic_revisions_operation_id_source_clinic_id_fkey` | `operation_id, source_clinic_id` | [operation_bindings](operation_bindings.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10769](../../../backend/migrations/001_init.sql#L10769) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `common_owner_basic_revisions_common_owner_id_fkey`: `REFERENCES common_owner_subjects(id) ON DELETE RESTRICT`
- `common_owner_basic_revisions_source_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `common_owner_basic_revisions_recorded_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`
- `common_owner_basic_revisions_source_local_owner_id_source__fkey`: `FOREIGN KEY(source_local_owner_id,source_clinic_id) REFERENCES owners(id,clinic_id) ON DELETE RESTRICT`
- `common_owner_basic_revisions_operation_id_source_clinic_id_fkey`: `FOREIGN KEY(operation_id,source_clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `common_owner_basic_revisions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10758](../../../backend/migrations/001_init.sql#L10758) |
| `common_owner_basic_revisions_revision_check` | check | `CHECK(revision>0)` | [001_init.sql:10759](../../../backend/migrations/001_init.sql#L10759) |
| `common_owner_basic_revisions_provenance_check` | check | `CHECK(jsonb_typeof(provenance)='object')` | [001_init.sql:10766](../../../backend/migrations/001_init.sql#L10766) |
| `common_owner_basic_revisions_check` | check | `CHECK((cygram_result IS NULL AND cygram_rule_version IS NULL) OR(birth_date IS NOT NULL AND cygram_result IS NOT NULL AND cygram_rule_version IS NOT NULL AND btrim(cygram_rule_version)<>''))` | [001_init.sql:10765](../../../backend/migrations/001_init.sql#L10765) |
| `common_owner_basic_revisions_id_common_owner_id_key` | unique | `UNIQUE(id,common_owner_id)` | [001_init.sql:10767](../../../backend/migrations/001_init.sql#L10767) |
| `common_owner_basic_revisions_common_owner_id_revision_key` | unique | `UNIQUE(common_owner_id,revision)` | [001_init.sql:10767](../../../backend/migrations/001_init.sql#L10767) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
