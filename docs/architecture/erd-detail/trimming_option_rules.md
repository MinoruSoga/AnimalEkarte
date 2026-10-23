<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `trimming_option_rules`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `clinic_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8001](../../../backend/migrations/001_init.sql#L8001) |
| `rule_set_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8002](../../../backend/migrations/001_init.sql#L8002) |
| `option_id` | `bigint` | NOT NULL | — | PK, FK | [001_init.sql:8003](../../../backend/migrations/001_init.sql#L8003) |
| `created_at` | `timestamptz` | NOT NULL | `clock_timestamp()` | — | [001_init.sql:8004](../../../backend/migrations/001_init.sql#L8004) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:8005](../../../backend/migrations/001_init.sql#L8005) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    trimming_option_rules {
        bigint clinic_id PK,FK
        bigint rule_set_id PK,FK
        bigint option_id PK,FK
        timestamptz created_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    trimming_option_rule_sets {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    trimming_options {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||--o{ trimming_option_rules : "clinic_id"
    trimming_option_rule_sets ||--o{ trimming_option_rules : "rule_set_id, clinic_id"
    trimming_options ||--o{ trimming_option_rules : "option_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `trimming_option_rules_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:8001](../../../backend/migrations/001_init.sql#L8001) |
| `trimming_option_rules_rule_set_id_clinic_id_fkey` | `rule_set_id, clinic_id` | [trimming_option_rule_sets](trimming_option_rule_sets.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8007](../../../backend/migrations/001_init.sql#L8007) |
| `trimming_option_rules_option_id_clinic_id_fkey` | `option_id, clinic_id` | [trimming_options](trimming_options.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:8008](../../../backend/migrations/001_init.sql#L8008) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `trimming_option_rules_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `trimming_option_rules_rule_set_id_clinic_id_fkey`: `FOREIGN KEY (rule_set_id, clinic_id) REFERENCES trimming_option_rule_sets(id, clinic_id) ON DELETE RESTRICT`
- `trimming_option_rules_option_id_clinic_id_fkey`: `FOREIGN KEY (option_id, clinic_id) REFERENCES trimming_options(id, clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `trimming_option_rules_pkey` | primary | `PRIMARY KEY (clinic_id, rule_set_id, option_id)` | [001_init.sql:8006](../../../backend/migrations/001_init.sql#L8006) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
