<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `owners`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:302](../../../backend/migrations/001_init.sql#L302) |
| `clinic_id` | `bigint` | NOT NULL | — | FK, UK | [001_init.sql:303](../../../backend/migrations/001_init.sql#L303) |
| `name` | `text` | NOT NULL | — | — | [001_init.sql:304](../../../backend/migrations/001_init.sql#L304) |
| `name_kana` | `text` | NOT NULL | `''` | — | [001_init.sql:305](../../../backend/migrations/001_init.sql#L305) |
| `birth_date` | `date` | NULL | — | — | [001_init.sql:306](../../../backend/migrations/001_init.sql#L306) |
| `company` | `text` | NOT NULL | `''` | — | [001_init.sql:307](../../../backend/migrations/001_init.sql#L307) |
| `postal_code` | `text` | NOT NULL | `''` | — | [001_init.sql:308](../../../backend/migrations/001_init.sql#L308) |
| `address1` | `text` | NOT NULL | `''` | — | [001_init.sql:309](../../../backend/migrations/001_init.sql#L309) |
| `address2` | `text` | NOT NULL | `''` | — | [001_init.sql:310](../../../backend/migrations/001_init.sql#L310) |
| `home_postal_code` | `text` | NOT NULL | `''` | — | [001_init.sql:311](../../../backend/migrations/001_init.sql#L311) |
| `home_address1` | `text` | NOT NULL | `''` | — | [001_init.sql:312](../../../backend/migrations/001_init.sql#L312) |
| `home_address2` | `text` | NOT NULL | `''` | — | [001_init.sql:313](../../../backend/migrations/001_init.sql#L313) |
| `phone` | `text` | NOT NULL | `''` | — | [001_init.sql:314](../../../backend/migrations/001_init.sql#L314) |
| `company_phone` | `text` | NOT NULL | `''` | — | [001_init.sql:315](../../../backend/migrations/001_init.sql#L315) |
| `email` | `text` | NOT NULL | `''` | — | [001_init.sql:316](../../../backend/migrations/001_init.sql#L316) |
| `remarks` | `text` | NOT NULL | `''` | — | [001_init.sql:317](../../../backend/migrations/001_init.sql#L317) |
| `is_dangerous` | `boolean` | NOT NULL | `false` | — | [001_init.sql:318](../../../backend/migrations/001_init.sql#L318) |
| `discount_rate` | `numeric(5,2)` | NOT NULL | `0` | — | [001_init.sql:319](../../../backend/migrations/001_init.sql#L319) |
| `membership_type` | `membership_type` | NOT NULL | `'non_member'` | — | [001_init.sql:320](../../../backend/migrations/001_init.sql#L320) |
| `line_user_id` | `text` | NULL | — | — | [001_init.sql:322](../../../backend/migrations/001_init.sql#L322) |
| `lstep_opt_out` | `boolean` | NOT NULL | `false` | — | [001_init.sql:324](../../../backend/migrations/001_init.sql#L324) |
| `lstep_opt_out_at` | `timestamptz` | NULL | — | — | [001_init.sql:325](../../../backend/migrations/001_init.sql#L325) |
| `lstep_opt_out_reason` | `text` | NULL | — | — | [001_init.sql:326](../../../backend/migrations/001_init.sql#L326) |
| `line_followed_at` | `timestamptz` | NULL | — | — | [001_init.sql:328](../../../backend/migrations/001_init.sql#L328) |
| `line_blocked_at` | `timestamptz` | NULL | — | — | [001_init.sql:329](../../../backend/migrations/001_init.sql#L329) |
| `line_id_confirmed_by` | `bigint` | NULL | — | FK | [001_init.sql:331](../../../backend/migrations/001_init.sql#L331) |
| `line_id_confirmed_at` | `timestamptz` | NULL | — | — | [001_init.sql:332](../../../backend/migrations/001_init.sql#L332) |
| `delivery_excluded` | `boolean` | NOT NULL | `false` | — | [001_init.sql:333](../../../backend/migrations/001_init.sql#L333) |
| `delivery_excluded_reason` | `varchar(100)` | NULL | — | — | [001_init.sql:334](../../../backend/migrations/001_init.sql#L334) |
| `is_transferred` | `boolean` | NOT NULL | `false` | — | [001_init.sql:335](../../../backend/migrations/001_init.sql#L335) |
| `transfer_at` | `timestamptz` | NULL | — | — | [001_init.sql:336](../../../backend/migrations/001_init.sql#L336) |
| `delivery_caution` | `boolean` | NOT NULL | `false` | — | [001_init.sql:338](../../../backend/migrations/001_init.sql#L338) |
| `delivery_caution_reason` | `varchar(100)` | NULL | — | — | [001_init.sql:339](../../../backend/migrations/001_init.sql#L339) |
| `dm_preference` | `boolean` | NULL | — | — | [001_init.sql:341](../../../backend/migrations/001_init.sql#L341) |
| `created_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:342](../../../backend/migrations/001_init.sql#L342) |
| `updated_at` | `timestamptz` | NOT NULL | `now()` | — | [001_init.sql:343](../../../backend/migrations/001_init.sql#L343) |
| `deleted_at` | `timestamptz` | NULL | — | — | [001_init.sql:344](../../../backend/migrations/001_init.sql#L344) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    owners {
        BIGSERIAL id PK,UK
        bigint clinic_id FK,UK
        text name
        text name_kana
        date birth_date
        text company
        text postal_code
        text address1
        text address2
        text home_postal_code
        text home_address1
        text home_address2
        text phone
        text company_phone
        text email
        text remarks
        boolean is_dangerous
        numeric_5_2 discount_rate
        membership_type membership_type
        text line_user_id
        boolean lstep_opt_out
        timestamptz lstep_opt_out_at
        text lstep_opt_out_reason
        timestamptz line_followed_at
        timestamptz line_blocked_at
        bigint line_id_confirmed_by FK
        timestamptz line_id_confirmed_at
        boolean delivery_excluded
        varchar_100 delivery_excluded_reason
        boolean is_transferred
        timestamptz transfer_at
        boolean delivery_caution
        varchar_100 delivery_caution_reason
        boolean dm_preference
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    clinics {
        BIGSERIAL id PK
    }
    staffs {
        BIGSERIAL id PK,UK
    }
    clinics ||..o{ owners : "clinic_id"
    staffs o|..o{ owners : "line_id_confirmed_by"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `owners_clinic_id_fkey` | `clinic_id` | [clinics](clinics.md) | `id` | `ON DELETE RESTRICT` | [001_init.sql:303](../../../backend/migrations/001_init.sql#L303) |
| `owners_line_id_confirmed_by_fkey` | `line_id_confirmed_by` | [staffs](staffs.md) | `id` | `ON DELETE SET NULL` | [001_init.sql:331](../../../backend/migrations/001_init.sql#L331) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `owners_clinic_id_fkey`: `REFERENCES clinics(id) ON DELETE RESTRICT`
- `owners_line_id_confirmed_by_fkey`: `REFERENCES staffs(id) ON DELETE SET NULL`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `owners_pkey` | primary | `PRIMARY KEY` | [001_init.sql:302](../../../backend/migrations/001_init.sql#L302) |
| `owners_name_kana_check` | check | `CHECK (name_kana !~ '[ァ-ヶ]')` | [001_init.sql:305](../../../backend/migrations/001_init.sql#L305) |
| `uq_owners_clinic_id_id` | unique | `CONSTRAINT uq_owners_clinic_id_id     UNIQUE (clinic_id, id)` | [001_init.sql:4102](../../../backend/migrations/001_init.sql#L4102) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `uk_owners_clinic_line_user_id` | YES | `line_user_id IS NOT NULL AND deleted_at IS NULL` | `CREATE UNIQUE INDEX uk_owners_clinic_line_user_id     ON owners(clinic_id, line_user_id)     WHERE line_user_id IS NOT NULL AND deleted_at IS NULL` | [001_init.sql:349](../../../backend/migrations/001_init.sql#L349) |
| `idx_owners_line_user_id` | NO | `line_user_id IS NOT NULL AND deleted_at IS NULL` | `CREATE INDEX idx_owners_line_user_id     ON owners(line_user_id)     WHERE line_user_id IS NOT NULL AND deleted_at IS NULL` | [001_init.sql:354](../../../backend/migrations/001_init.sql#L354) |
| `idx_owners_delivery_excluded` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_owners_delivery_excluded     ON owners (clinic_id, delivery_excluded)     WHERE deleted_at IS NULL` | [001_init.sql:368](../../../backend/migrations/001_init.sql#L368) |
| `idx_owners_is_transferred` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_owners_is_transferred     ON owners (clinic_id, is_transferred)     WHERE deleted_at IS NULL` | [001_init.sql:372](../../../backend/migrations/001_init.sql#L372) |
| `idx_owners_line_id_confirmed` | NO | `deleted_at IS NULL` | `CREATE INDEX idx_owners_line_id_confirmed     ON owners (clinic_id, line_id_confirmed_at)     WHERE deleted_at IS NULL` | [001_init.sql:376](../../../backend/migrations/001_init.sql#L376) |
| `idx_owners_line_id_confirmed_by` | NO | `line_id_confirmed_by IS NOT NULL` | `CREATE INDEX idx_owners_line_id_confirmed_by     ON owners (line_id_confirmed_by)     WHERE line_id_confirmed_by IS NOT NULL` | [001_init.sql:380](../../../backend/migrations/001_init.sql#L380) |
| `uk_owners_clinic_email` | YES | `deleted_at IS NULL AND email IS NOT NULL AND email != ''` | `CREATE UNIQUE INDEX uk_owners_clinic_email ON owners(clinic_id, email) WHERE deleted_at IS NULL AND email IS NOT NULL AND email != ''` | [001_init.sql:2202](../../../backend/migrations/001_init.sql#L2202) |
| `idx_owners_clinic_id` | NO | — | `CREATE INDEX idx_owners_clinic_id ON owners(clinic_id)` | [001_init.sql:2246](../../../backend/migrations/001_init.sql#L2246) |
| `uk_owners_clinic_phone` | YES | `deleted_at IS NULL AND phone <> ''` | `CREATE UNIQUE INDEX IF NOT EXISTS uk_owners_clinic_phone     ON owners (clinic_id, phone)     WHERE deleted_at IS NULL AND phone <> ''` | [001_init.sql:4540](../../../backend/migrations/001_init.sql#L4540) |
| `idx_owners_name_fold_trgm` | NO | — | `CREATE INDEX idx_owners_name_fold_trgm   ON owners USING gin ((translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6664](../../../backend/migrations/001_init.sql#L6664) |
| `idx_owners_name_trgm` | NO | — | `CREATE INDEX idx_owners_name_trgm   ON owners USING gin (name gin_trgm_ops)` | [001_init.sql:6671](../../../backend/migrations/001_init.sql#L6671) |
| `idx_owners_name_kana_fold_trgm` | NO | — | `CREATE INDEX idx_owners_name_kana_fold_trgm   ON owners USING gin ((translate(name_kana, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ ')) gin_trgm_ops)` | [001_init.sql:6674](../../../backend/migrations/001_init.sql#L6674) |
| `idx_owners_phone_trgm` | NO | — | `CREATE INDEX idx_owners_phone_trgm   ON owners USING gin (phone gin_trgm_ops)` | [001_init.sql:6680](../../../backend/migrations/001_init.sql#L6680) |
| `idx_owners_email_trgm` | NO | — | `CREATE INDEX idx_owners_email_trgm   ON owners USING gin (email gin_trgm_ops)` | [001_init.sql:6683](../../../backend/migrations/001_init.sql#L6683) |
| `idx_owners_name_compact_fold_trgm` | NO | — | `CREATE INDEX idx_owners_name_compact_fold_trgm   ON owners USING gin ((regexp_replace(translate(name, 'ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ　', 'ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ '), '[[:space:]　]+', '', 'g')) gin_trgm_ops)` | [001_init.sql:6687](../../../backend/migrations/001_init.sql#L6687) |
| `idx_owners_id_text` | NO | — | `CREATE INDEX idx_owners_id_text   ON owners ((id::text))` | [001_init.sql:6693](../../../backend/migrations/001_init.sql#L6693) |
