<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `staff_reservation_exclusions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK | [001_init.sql:2685](../../../backend/migrations/001_init.sql#L2685) |
| `staff_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:2686](../../../backend/migrations/001_init.sql#L2686) |
| `reservation_type_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:2687](../../../backend/migrations/001_init.sql#L2687) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    staff_reservation_exclusions {
        BIGSERIAL id PK
        bigint staff_id FK,UK
        bigint reservation_type_id FK,UK
    }
    reservation_types {
        BIGSERIAL id PK,UK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    staffs ||..o{ staff_reservation_exclusions : "staff_id"
    reservation_types ||..o{ staff_reservation_exclusions : "reservation_type_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `staff_reservation_exclusions_staff_id_fkey` | `staff_id` | [staffs](staffs.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2686](../../../backend/migrations/001_init.sql#L2686) |
| `staff_reservation_exclusions_reservation_type_id_fkey` | `reservation_type_id` | [reservation_types](reservation_types.md) | `id` | `ON DELETE CASCADE` | [001_init.sql:2687](../../../backend/migrations/001_init.sql#L2687) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `staff_reservation_exclusions_staff_id_fkey`: `REFERENCES staffs(id) ON DELETE CASCADE`
- `staff_reservation_exclusions_reservation_type_id_fkey`: `REFERENCES reservation_types(id) ON DELETE CASCADE`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `staff_reservation_exclusions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:2685](../../../backend/migrations/001_init.sql#L2685) |
| `staff_reservation_exclusions_staff_id_reservation_type_id_key` | unique | `UNIQUE(staff_id, reservation_type_id)` | [001_init.sql:2688](../../../backend/migrations/001_init.sql#L2688) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
