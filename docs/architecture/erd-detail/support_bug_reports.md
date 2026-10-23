<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `support_bug_reports`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:6578](../../../backend/migrations/001_init.sql#L6578) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:6579](../../../backend/migrations/001_init.sql#L6579) |
| `reporter_staff_id` | `bigint` | NOT NULL | — | FK | [001_init.sql:6580](../../../backend/migrations/001_init.sql#L6580) |
| `title` | `text` | NOT NULL | — | — | [001_init.sql:6581](../../../backend/migrations/001_init.sql#L6581) |
| `detail` | `text` | NOT NULL | `''` | — | [001_init.sql:6582](../../../backend/migrations/001_init.sql#L6582) |
| `page_url` | `text` | NOT NULL | `''` | — | [001_init.sql:6583](../../../backend/migrations/001_init.sql#L6583) |
| `route_path` | `text` | NOT NULL | `''` | — | [001_init.sql:6584](../../../backend/migrations/001_init.sql#L6584) |
| `user_agent` | `text` | NOT NULL | `''` | — | [001_init.sql:6585](../../../backend/migrations/001_init.sql#L6585) |
| `viewport` | `text` | NOT NULL | `''` | — | [001_init.sql:6586](../../../backend/migrations/001_init.sql#L6586) |
| `app_version` | `text` | NOT NULL | `''` | — | [001_init.sql:6587](../../../backend/migrations/001_init.sql#L6587) |
| `screenshot_key` | `text` | NULL | — | — | [001_init.sql:6588](../../../backend/migrations/001_init.sql#L6588) |
| `status` | `varchar(20)` | NOT NULL | `'open'` | — | [001_init.sql:6589](../../../backend/migrations/001_init.sql#L6589) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:6590](../../../backend/migrations/001_init.sql#L6590) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:6591](../../../backend/migrations/001_init.sql#L6591) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:6592](../../../backend/migrations/001_init.sql#L6592) |
| `version` | `bigint` | NOT NULL | `1` | — | [001_init.sql:10027](../../../backend/migrations/001_init.sql#L10027) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    support_bug_reports {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint reporter_staff_id FK
        text title
        text detail
        text page_url
        text route_path
        text user_agent
        text viewport
        text app_version
        text screenshot_key
        varchar_20 status
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
        bigint version
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ support_bug_reports : "clinic_id"
    staffs ||..o{ support_bug_reports : "reporter_staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `support_bug_reports_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:6579](../../../backend/migrations/001_init.sql#L6579) |
| `support_bug_reports_reporter_staff_id_fkey` | `reporter_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:6580](../../../backend/migrations/001_init.sql#L6580) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `support_bug_reports_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `support_bug_reports_reporter_staff_id_fkey`: `REFERENCES staffs(id)  ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `support_bug_reports_pkey` | primary | `PRIMARY KEY` | [001_init.sql:6578](../../../backend/migrations/001_init.sql#L6578) |
| `support_bug_reports_version_check` | check | `CHECK(version>0)` | [001_init.sql:10027](../../../backend/migrations/001_init.sql#L10027) |
| `uq_support_report_id_clinic` | unique | `CONSTRAINT uq_support_report_id_clinic UNIQUE(id,clinic_id)` | [001_init.sql:10028](../../../backend/migrations/001_init.sql#L10028) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_support_bug_reports_clinic_created` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_support_bug_reports_clinic_created     ON support_bug_reports (clinic_id, created_at DESC)     WHERE deleted_at IS NULL` | [001_init.sql:6595](../../../backend/migrations/001_init.sql#L6595) |
| `idx_support_bug_reports_clinic_status` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_support_bug_reports_clinic_status     ON support_bug_reports (clinic_id, status)     WHERE deleted_at IS NULL` | [001_init.sql:6598](../../../backend/migrations/001_init.sql#L6598) |
