<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `support_chat_messages`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [013_support_chat_messages.sql:7](../../../backend/migrations/013_support_chat_messages.sql#L7) |
| `clinic_id` | `bigint` | NOT NULL | — | FK | [013_support_chat_messages.sql:8](../../../backend/migrations/013_support_chat_messages.sql#L8) |
| `staff_id` | `bigint` | NOT NULL | — | FK | [013_support_chat_messages.sql:9](../../../backend/migrations/013_support_chat_messages.sql#L9) |
| `role` | `varchar(16)` | NOT NULL | — | — | [013_support_chat_messages.sql:10](../../../backend/migrations/013_support_chat_messages.sql#L10) |
| `content` | `text` | NOT NULL | — | — | [013_support_chat_messages.sql:11](../../../backend/migrations/013_support_chat_messages.sql#L11) |
| `sources` | `jsonb` | NULL | — | — | [013_support_chat_messages.sql:12](../../../backend/migrations/013_support_chat_messages.sql#L12) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [013_support_chat_messages.sql:13](../../../backend/migrations/013_support_chat_messages.sql#L13) |
| `deleted_at` | `timestamptz` | NULL | — | — | [013_support_chat_messages.sql:14](../../../backend/migrations/013_support_chat_messages.sql#L14) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    support_chat_messages {
        BIGSERIAL id PK
        bigint clinic_id FK
        bigint staff_id FK
        varchar_16 role
        text content
        jsonb sources
        timestamptz created_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ support_chat_messages : "clinic_id"
    staffs ||..o{ support_chat_messages : "staff_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `support_chat_messages_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [013_support_chat_messages.sql:8](../../../backend/migrations/013_support_chat_messages.sql#L8) |
| `support_chat_messages_staff_id_fkey` | `staff_id` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [013_support_chat_messages.sql:9](../../../backend/migrations/013_support_chat_messages.sql#L9) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `support_chat_messages_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `support_chat_messages_staff_id_fkey`: `REFERENCES staffs(id)  ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `support_chat_messages_pkey` | primary | `PRIMARY KEY` | [013_support_chat_messages.sql:7](../../../backend/migrations/013_support_chat_messages.sql#L7) |
| `support_chat_messages_role_check` | check | `CHECK (role IN ('user', 'assistant'))` | [013_support_chat_messages.sql:10](../../../backend/migrations/013_support_chat_messages.sql#L10) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_support_chat_messages_clinic_staff_created` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_support_chat_messages_clinic_staff_created     ON support_chat_messages (clinic_id, staff_id, created_at, id)     WHERE deleted_at IS NULL` | [013_support_chat_messages.sql:17](../../../backend/migrations/013_support_chat_messages.sql#L17) |
