<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `support_delivery_destinations`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:10024](../../../backend/migrations/001_init.sql#L10024) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10024](../../../backend/migrations/001_init.sql#L10024) |
| `intent_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:10025](../../../backend/migrations/001_init.sql#L10025) |
| `provider` | `text` | NOT NULL | — | UK | [001_init.sql:10025](../../../backend/migrations/001_init.sql#L10025) |
| `destination_ref` | `text` | NOT NULL | — | UK | [001_init.sql:10026](../../../backend/migrations/001_init.sql#L10026) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    support_delivery_destinations {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        bigint intent_id FK,UK
        text provider UK
        text destination_ref UK
    }
    clinics {
        BIGSERIAL id PK
    }
    support_delivery_intents {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
    }
    clinics ||..o{ support_delivery_destinations : "clinic_id"
    support_delivery_intents ||..o{ support_delivery_destinations : "intent_id, clinic_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `support_delivery_destinations_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:10024](../../../backend/migrations/001_init.sql#L10024) |
| `support_delivery_destinations_intent_id_clinic_id_fkey` | `intent_id, clinic_id` | [support_delivery_intents](support_delivery_intents.md) | `id, clinic_id` | `ON DELETE RESTRICT` | [001_init.sql:10028](../../../backend/migrations/001_init.sql#L10028) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `support_delivery_destinations_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `support_delivery_destinations_intent_id_clinic_id_fkey`: `FOREIGN KEY(intent_id,clinic_id) REFERENCES support_delivery_intents(id,clinic_id) ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `support_delivery_destinations_pkey` | primary | `PRIMARY KEY` | [001_init.sql:10024](../../../backend/migrations/001_init.sql#L10024) |
| `support_delivery_destinations_provider_check` | check | `CHECK(provider IN('discord','plane'))` | [001_init.sql:10025](../../../backend/migrations/001_init.sql#L10025) |
| `support_delivery_destinations_destination_ref_check` | check | `CHECK(btrim(destination_ref)<>'')` | [001_init.sql:10026](../../../backend/migrations/001_init.sql#L10026) |
| `support_delivery_destinations_id_clinic_id_key` | unique | `UNIQUE(id,clinic_id)` | [001_init.sql:10027](../../../backend/migrations/001_init.sql#L10027) |
| `support_delivery_destinations_clinic_id_intent_id_provider__key` | unique | `UNIQUE(clinic_id,intent_id,provider,destination_ref)` | [001_init.sql:10027](../../../backend/migrations/001_init.sql#L10027) |

## インデックス

明示的なCREATE INDEXなし。PK/UNIQUEに付随する暗黙のindexは制約欄を参照してください。
