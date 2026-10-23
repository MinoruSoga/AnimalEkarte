<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `trimming_course_included_services`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:7972](../../../backend/migrations/001_init.sql#L7972) |
| `course_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:7973](../../../backend/migrations/001_init.sql#L7973) |
| `service_identity_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:7974](../../../backend/migrations/001_init.sql#L7974) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:7975](../../../backend/migrations/001_init.sql#L7975) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:7976](../../../backend/migrations/001_init.sql#L7976) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    trimming_course_included_services {
        bigint clinic_id PK,FK
        bigint course_id PK,FK
        bigint service_identity_id PK,FK
        timestamptz created_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    trimming_courses {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    trimming_service_identities {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ trimming_course_included_services : "clinic_id"
    trimming_courses ||--o{ trimming_course_included_services : "course_id, clinic_id"
    trimming_service_identities ||--o{ trimming_course_included_services : "service_identity_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `trimming_course_included_services_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:7972](../../../backend/migrations/001_init.sql#L7972) |
| `trimming_course_included_services_course_id_clinic_id_fkey` | `course_id, clinic_id` | [trimming_courses](trimming_courses.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:7978](../../../backend/migrations/001_init.sql#L7978) |
| `trimming_course_included_serv_service_identity_id_clinic_i_fkey` | `service_identity_id, clinic_id` | [trimming_service_identities](trimming_service_identities.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:7979](../../../backend/migrations/001_init.sql#L7979) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `trimming_course_included_services_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `trimming_course_included_services_course_id_clinic_id_fkey`: `FOREIGN KEY (course_id, clinic_id) REFERENCES trimming_courses(id, clinic_id) ON DELETE RESTRICT`
- `trimming_course_included_serv_service_identity_id_clinic_i_fkey`: `FOREIGN KEY (service_identity_id, clinic_id) REFERENCES trimming_service_identities(id, clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `trimming_course_included_services_pkey` | primary | `PRIMARY KEY (clinic_id, course_id, service_identity_id)` | [001_init.sql:7977](../../../backend/migrations/001_init.sql#L7977) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
