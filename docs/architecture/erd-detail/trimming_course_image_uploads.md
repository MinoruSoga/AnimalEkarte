<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `trimming_course_image_uploads`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `BIGINT` | NOT NULL | — | PK, FK | [003_trimming_course_image_uploads.sql:4](../../../backend/migrations/003_trimming_course_image_uploads.sql#L4) |
| `object_key` | `TEXT` | NOT NULL | — | PK | [003_trimming_course_image_uploads.sql:5](../../../backend/migrations/003_trimming_course_image_uploads.sql#L5) |
| `uploader_staff_id` | `BIGINT` | NOT NULL | — | FK | [003_trimming_course_image_uploads.sql:6](../../../backend/migrations/003_trimming_course_image_uploads.sql#L6) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `NOW()` | — | [003_trimming_course_image_uploads.sql:7](../../../backend/migrations/003_trimming_course_image_uploads.sql#L7) |
| `uploaded_at` | `TIMESTAMPTZ` | NULL | — | — | [003_trimming_course_image_uploads.sql:8](../../../backend/migrations/003_trimming_course_image_uploads.sql#L8) |
| `submitted_at` | `TIMESTAMPTZ` | NULL | — | — | [003_trimming_course_image_uploads.sql:9](../../../backend/migrations/003_trimming_course_image_uploads.sql#L9) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    trimming_course_image_uploads {
        BIGINT clinic_id PK,FK
        TEXT object_key PK
        BIGINT uploader_staff_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ uploaded_at
        TIMESTAMPTZ submitted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||--o{ trimming_course_image_uploads : "clinic_id"
    staffs ||..o{ trimming_course_image_uploads : "uploader_staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `trimming_course_image_uploads_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [003_trimming_course_image_uploads.sql:4](../../../backend/migrations/003_trimming_course_image_uploads.sql#L4) |
| `trimming_course_image_uploads_uploader_staff_id_fkey` | `uploader_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [003_trimming_course_image_uploads.sql:6](../../../backend/migrations/003_trimming_course_image_uploads.sql#L6) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `trimming_course_image_uploads_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `trimming_course_image_uploads_uploader_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `trimming_course_image_uploads_pkey` | primary | `PRIMARY KEY (clinic_id, object_key)` | [003_trimming_course_image_uploads.sql:10](../../../backend/migrations/003_trimming_course_image_uploads.sql#L10) |
| `trimming_course_image_uploads_check` | check | `CHECK (clinic_id > 0 AND uploader_staff_id > 0)` | [003_trimming_course_image_uploads.sql:11](../../../backend/migrations/003_trimming_course_image_uploads.sql#L11) |
| `trimming_course_image_uploads_check1` | check | <code>CHECK (object_key LIKE &#x27;trimming-images/&#x27; &#124;&#124; clinic_id::text &#124;&#124; &#x27;/%&#x27;            AND object_key ~ &#x27;^trimming-images/[1-9][0-9]*/[0-9]+_[0-9a-f]{32}\.(jpg&#124;jpeg&#124;png&#124;gif)$&#x27;)</code> | [003_trimming_course_image_uploads.sql:12](../../../backend/migrations/003_trimming_course_image_uploads.sql#L12) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
