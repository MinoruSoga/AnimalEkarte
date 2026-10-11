<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `support_bug_reports`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [011_support_bug_reports.sql:7](../../../backend/migrations/011_support_bug_reports.sql#L7) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [011_support_bug_reports.sql:8](../../../backend/migrations/011_support_bug_reports.sql#L8) |
| `reporter_staff_id` | `bigint` | NOT NULL | — | FK | [011_support_bug_reports.sql:9](../../../backend/migrations/011_support_bug_reports.sql#L9) |
| `title` | `text` | NOT NULL | — | — | [011_support_bug_reports.sql:10](../../../backend/migrations/011_support_bug_reports.sql#L10) |
| `detail` | `text` | NOT NULL | `''` | — | [011_support_bug_reports.sql:11](../../../backend/migrations/011_support_bug_reports.sql#L11) |
| `page_url` | `text` | NOT NULL | `''` | — | [011_support_bug_reports.sql:12](../../../backend/migrations/011_support_bug_reports.sql#L12) |
| `route_path` | `text` | NOT NULL | `''` | — | [011_support_bug_reports.sql:13](../../../backend/migrations/011_support_bug_reports.sql#L13) |
| `user_agent` | `text` | NOT NULL | `''` | — | [011_support_bug_reports.sql:14](../../../backend/migrations/011_support_bug_reports.sql#L14) |
| `viewport` | `text` | NOT NULL | `''` | — | [011_support_bug_reports.sql:15](../../../backend/migrations/011_support_bug_reports.sql#L15) |
| `app_version` | `text` | NOT NULL | `''` | — | [011_support_bug_reports.sql:16](../../../backend/migrations/011_support_bug_reports.sql#L16) |
| `screenshot_key` | `text` | NULL | — | — | [011_support_bug_reports.sql:17](../../../backend/migrations/011_support_bug_reports.sql#L17) |
| `status` | `varchar(20)` | NOT NULL | `'open'` | — | [011_support_bug_reports.sql:18](../../../backend/migrations/011_support_bug_reports.sql#L18) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [011_support_bug_reports.sql:19](../../../backend/migrations/011_support_bug_reports.sql#L19) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [011_support_bug_reports.sql:20](../../../backend/migrations/011_support_bug_reports.sql#L20) |
| `deleted_at` | `timestamptz` | NULL | — | — | [011_support_bug_reports.sql:21](../../../backend/migrations/011_support_bug_reports.sql#L21) |
| `plane_issue_id` | `text` | NULL | — | — | [016_support_bug_reports_plane.sql:8](../../../backend/migrations/016_support_bug_reports_plane.sql#L8) |
| `plane_issue_url` | `text` | NULL | — | — | [016_support_bug_reports_plane.sql:9](../../../backend/migrations/016_support_bug_reports_plane.sql#L9) |
| `plane_sync_error` | `text` | NULL | — | — | [016_support_bug_reports_plane.sql:10](../../../backend/migrations/016_support_bug_reports_plane.sql#L10) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    support_bug_reports {
        BIGSERIAL id PK
        bigint clinic_id FK
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
        text plane_issue_id
        text plane_issue_url
        text plane_sync_error
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
| `support_bug_reports_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [011_support_bug_reports.sql:8](../../../backend/migrations/011_support_bug_reports.sql#L8) |
| `support_bug_reports_reporter_staff_id_fkey` | `reporter_staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [011_support_bug_reports.sql:9](../../../backend/migrations/011_support_bug_reports.sql#L9) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `support_bug_reports_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `support_bug_reports_reporter_staff_id_fkey`: `REFERENCES staffs(id)  ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `support_bug_reports_pkey` | primary | `PRIMARY KEY` | [011_support_bug_reports.sql:7](../../../backend/migrations/011_support_bug_reports.sql#L7) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_support_bug_reports_clinic_created` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_support_bug_reports_clinic_created     ON support_bug_reports (clinic_id, created_at DESC)     WHERE deleted_at IS NULL` | [011_support_bug_reports.sql:24](../../../backend/migrations/011_support_bug_reports.sql#L24) |
| `idx_support_bug_reports_clinic_status` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_support_bug_reports_clinic_status     ON support_bug_reports (clinic_id, status)     WHERE deleted_at IS NULL` | [011_support_bug_reports.sql:27](../../../backend/migrations/011_support_bug_reports.sql#L27) |
