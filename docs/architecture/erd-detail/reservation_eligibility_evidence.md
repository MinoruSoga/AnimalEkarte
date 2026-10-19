<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `reservation_eligibility_evidence`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:11008](../../../backend/migrations/001_init.sql#L11008) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:11008](../../../backend/migrations/001_init.sql#L11008) |
| `source_kind` | `text` | NOT NULL | — | UK | [001_init.sql:11009](../../../backend/migrations/001_init.sql#L11009) |
| `source_reference` | `text` | NOT NULL | — | UK | [001_init.sql:11010](../../../backend/migrations/001_init.sql#L11010) |
| `original_values` | `jsonb` | NOT NULL | — | — | [001_init.sql:11010](../../../backend/migrations/001_init.sql#L11010) |
| `captured_at` | `timestamptz` | NOT NULL | — | — | [001_init.sql:11011](../../../backend/migrations/001_init.sql#L11011) |
| `captured_by` | `bigint` | NULL | — | FK | [001_init.sql:11011](../../../backend/migrations/001_init.sql#L11011) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    reservation_eligibility_evidence {
        BIGSERIAL id PK
        bigint clinic_id FK,UK
        text source_kind UK
        text source_reference UK
        jsonb original_values
        timestamptz captured_at
        bigint captured_by FK
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ reservation_eligibility_evidence : "clinic_id"
    staffs o|..o{ reservation_eligibility_evidence : "captured_by"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `reservation_eligibility_evidence_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:11008](../../../backend/migrations/001_init.sql#L11008) |
| `reservation_eligibility_evidence_captured_by_fkey` | `captured_by` | [staffs](staffs.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:11011](../../../backend/migrations/001_init.sql#L11011) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `reservation_eligibility_evidence_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `reservation_eligibility_evidence_captured_by_fkey`: `REFERENCES staffs(id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `reservation_eligibility_evidence_pkey` | primary | `PRIMARY KEY` | [001_init.sql:11008](../../../backend/migrations/001_init.sql#L11008) |
| `reservation_eligibility_evidence_source_kind_check` | check | `CHECK(source_kind IN('legacy_negative_capability','legacy_occupation_gate'))` | [001_init.sql:11009](../../../backend/migrations/001_init.sql#L11009) |
| `reservation_eligibility_evidence_original_values_check` | check | `CHECK(jsonb_typeof(original_values)='object')` | [001_init.sql:11010](../../../backend/migrations/001_init.sql#L11010) |
| `reservation_eligibility_evide_clinic_id_source_kind_source__key` | unique | `UNIQUE(clinic_id,source_kind,source_reference)` | [001_init.sql:11012](../../../backend/migrations/001_init.sql#L11012) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
