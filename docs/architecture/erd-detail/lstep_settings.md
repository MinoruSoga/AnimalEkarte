<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lstep_settings`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:472](../../../backend/migrations/001_init.sql#L472) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:473](../../../backend/migrations/001_init.sql#L473) |
| `is_sync_enabled` | `boolean` | NOT NULL | `false` | — | [001_init.sql:474](../../../backend/migrations/001_init.sql#L474) |
| `sync_enabled_at` | `timestamptz` | NULL | — | — | [001_init.sql:475](../../../backend/migrations/001_init.sql#L475) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:476](../../../backend/migrations/001_init.sql#L476) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:477](../../../backend/migrations/001_init.sql#L477) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lstep_settings {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        boolean is_sync_enabled
        timestamptz sync_enabled_at
        timestamptz created_at
        timestamptz updated_at
    }
    clinics {
        BIGSERIAL id PK
    }
    clinics ||..o| lstep_settings : "clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `lstep_settings_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:473](../../../backend/migrations/001_init.sql#L473) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `lstep_settings_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lstep_settings_pkey` | primary | `PRIMARY KEY` | [001_init.sql:472](../../../backend/migrations/001_init.sql#L472) |
| `lstep_settings_clinic_id_key` | unique | `UNIQUE` | [001_init.sql:473](../../../backend/migrations/001_init.sql#L473) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lstep_settings_clinic_id` | NO | — | `CREATE INDEX idx_lstep_settings_clinic_id ON lstep_settings (clinic_id)` | [001_init.sql:480](../../../backend/migrations/001_init.sql#L480) |
