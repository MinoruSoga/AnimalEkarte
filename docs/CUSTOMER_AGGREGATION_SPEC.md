# 顧客集計機能 仕様書

> **ステータス**: Draft  
> **対象範囲**: 飼い主単位の年間売上ランキング、来院回数、最終来院日分類
> **最終更新**: 2026-04-26  
> **関連API**: `GET /api/v1/clinics/:clinic_id/owners/aggregations`

---

## 1. 目的

飼い主単位で診療費と来院状況を集計し、次の3つの業務に使える一覧を提供する。

| # | 集計 | 目的 |
|---|------|------|
| 1 | 顧客ごとの年間売上ランキング | 年間の累計診療費が高い飼い主を把握する |
| 2 | 顧客ごとの来院回数 | 任意期間内の来院頻度を把握する |
| 3 | 最終来院日での分類 | 最終来院からの経過期間で未来院対象を抽出する |

この機能は独立した顧客集計ダッシュボードとして実装し、売上・来院・最終来院の3軸を同じ一覧基盤で切り替えられるようにする。

---

## 2. 用語と命名規則

### 2.1 命名規則

| 層 | 命名方針 | 例 |
|----|----------|----|
| 画面表示名 | 日本語の業務用語を使う | `売上ランキング`、`来院回数`、`最終来院` |
| API パラメータ | `snake_case` で統一する | `period_preset`、`amount_basis`、`last_visit_bucket` |
| レスポンス項目 | `snake_case` で統一する | `annual_amount`、`period_visit_count`、`days_since_last_visit` |
| 列挙値 | 仕様書と同じ英字キーを使う | `gross_total_amount`、`over_3m`、`no_visit` |

### 2.2 正規名称

- 年間売上は `annual_amount`
- 期間内来院回数は `period_visit_count`
- 累計来院回数は `total_visit_count`
- 最終来院分類は `last_visit_bucket`
- 経過日数は `days_since_last_visit`

既存互換の別名を追加する場合でも、正規名称は上記に固定する。

---

## 3. データソース

### 3.1 集計単位

集計単位は `owners.id` の飼い主単位とする。同一飼い主に複数ペットがいる場合は、全ペットの診療記録・会計を合算する。

| エンティティ | 用途 | 参照カラム |
|--------------|------|------------|
| `owners` | 集計対象の飼い主 | `id`, `clinic_id`, `name`, `deleted_at` |
| `medical_records` | 来院回数、初診日、最終来院日 | `owner_id`, `clinic_id`, `date`, `deleted_at` |
| `billings` | 診療費・売上 | `medical_record_id`, `owner_id`, `clinic_id`, `total_amount`, `status`, `completed_at`, `deleted_at` |
| `payments` | 支払ベース集計を使う場合の支払額 | `billing_id`, `billing_amount`, `deleted_at` |
| `billing_refunds` | 返金控除を使う場合の返金額 | `billing_id`, `amount` |

### 3.2 集計基準

| 項目 | 定義 |
|------|------|
| 年間売上 | `billings.total_amount` の合計、または `amount_basis` に応じた支払ベース集計 |
| 期間内来院回数 | `medical_records.date` の distinct 件数 |
| 累計来院回数 | 全期間の `medical_records.date` の distinct 件数 |
| 最終来院日 | `MAX(medical_records.date)` |
| 初診日 | `MIN(medical_records.date)` |
| 経過日数 | `CURRENT_DATE - last_visit_date` |

### 3.3 来院回数の数え方

来院回数は飼い主単位の `medical_records.date` で数える。同一飼い主が同じ日に複数カルテを持つ場合は1回とする。

---

## 4. 機能仕様

### 4.1 顧客ごとの年間売上ランキング

#### 概要

指定年または指定期間内の診療費を飼い主単位で合計し、売上降順に並べる一覧。

#### 入力条件

| 条件 | 型 | デフォルト | 説明 |
|------|----|------------|------|
| `year` | int | 当年 | 年指定。`from` / `to` がある場合は補助条件として扱う |
| `from` | date | `year-01-01` | 集計開始日 |
| `to` | date | `year-12-31` | 集計終了日 |
| `amount_basis` | string | `gross_total_amount` | `gross_total_amount` / `paid_amount` / `net_paid_amount` |
| `min_amount` | int | なし | 年間売上の下限 |
| `max_amount` | int | なし | 年間売上の上限 |
| `include_zero` | bool | `false` | 0円の飼い主を含めるか |
| `search` | string | なし | 飼い主名の部分一致検索 |
| `sort` | string | `annual_amount` | `annual_amount` / `period_visit_count` / `last_visit_date` / `days_since_last_visit` |
| `order` | string | `desc` | `asc` / `desc` |
| `page` | int | `1` | ページ番号 |
| `per_page` | int | `50` | 1ページ件数。最大200 |

#### 表示項目

| カラム | 内容 |
|--------|------|
| 順位 | 売上順の連番 |
| 飼い主名 | `owners.name` |
| 年間診療費 | 指定期間内の診療費合計 |
| 会計件数 | 指定期間内の対象会計件数 |
| 期間内来院回数 | 指定期間内の `COUNT(DISTINCT medical_records.date)` |
| 最終来院日 | 全期間の `MAX(medical_records.date)` |
| 初診日 | 全期間の `MIN(medical_records.date)` |

#### 集計ルール

- 既定の集計基準は `gross_total_amount` とする。
- `paid_amount` は支払額ベース、`net_paid_amount` は返金控除後の売上として使う。
- 期間外の会計は集計対象外とする。
- `status = completed` の会計を対象にする。

---

### 4.2 顧客ごとの来院回数

#### 概要

指定期間内の来院回数を飼い主単位で集計し、来院回数順または最終来院日順で一覧化する。

#### 入力条件

| 条件 | 型 | デフォルト | 説明 |
|------|----|------------|------|
| `from` | date | なし | 集計開始日 |
| `to` | date | なし | 集計終了日 |
| `period_preset` | string | `last_12_months` | `last_3_months` / `last_6_months` / `last_12_months` / `calendar_year` |
| `min_visit_count` | int | なし | 期間内来院回数の下限 |
| `max_visit_count` | int | なし | 期間内来院回数の上限 |
| `include_zero` | bool | `false` | 0回来院を含めるか |
| `search` | string | なし | 飼い主名の部分一致検索 |
| `sort` | string | `period_visit_count` | `period_visit_count` / `last_visit_date` / `owner_name` |
| `order` | string | `desc` | `asc` / `desc` |

#### 表示項目

| カラム | 内容 |
|--------|------|
| 飼い主名 | `owners.name` |
| 期間内来院回数 | 指定期間内の `COUNT(DISTINCT medical_records.date)` |
| 累計来院回数 | 全期間の `COUNT(DISTINCT medical_records.date)` |
| 年間来院回数 | 過去365日の `COUNT(DISTINCT medical_records.date)` |
| 最終来院日 | 全期間の `MAX(medical_records.date)` |
| 初診日 | 全期間の `MIN(medical_records.date)` |

---

### 4.3 最終来院日での分類

#### 概要

飼い主ごとの最終来院日を表示し、最終来院からの経過日数で分類する。

#### 分類

| 分類キー | 条件 | 表示ラベル |
|----------|------|------------|
| `within_3m` | 0〜89日 | 3ヶ月未満 |
| `over_3m` | 90〜179日 | 3ヶ月以上 |
| `over_6m` | 180〜364日 | 6ヶ月以上 |
| `over_1y` | 365日以上 | 1年以上 |
| `no_visit` | 来院記録なし | 来院なし |

#### 入力条件

| 条件 | 型 | デフォルト | 説明 |
|------|----|------------|------|
| `last_visit_bucket` | string | なし | `within_3m` / `over_3m` / `over_6m` / `over_1y` / `no_visit` |
| `min_days_since_last_visit` | int | なし | 経過日数下限 |
| `max_days_since_last_visit` | int | なし | 経過日数上限 |
| `include_no_visit` | bool | `true` | 来院なしを含めるか |
| `search` | string | なし | 飼い主名の部分一致検索 |
| `sort` | string | `last_visit_date` | `last_visit_date` / `days_since_last_visit` / `owner_name` |
| `order` | string | `asc` | 最終来院が古い順をデフォルトにする |

#### 表示項目

| カラム | 内容 |
|--------|------|
| 飼い主名 | `owners.name` |
| 最終来院日 | `MAX(medical_records.date)` |
| 経過日数 | `CURRENT_DATE - last_visit_date` |
| 分類 | `within_3m` / `over_3m` / `over_6m` / `over_1y` / `no_visit` |
| 累計来院回数 | 全期間の `COUNT(DISTINCT medical_records.date)` |
| 年間来院回数 | 過去365日の `COUNT(DISTINCT medical_records.date)` |
| 累計診療費 | `billings.total_amount` の合計 |

---

## 5. API設計

### 5.1 エンドポイント

```http
GET /api/v1/clinics/:clinic_id/owners/aggregations
Authorization: Bearer {jwt_token}
```

### 5.2 クエリパラメータ

| パラメータ | 型 | 説明 |
|------------|----|------|
| `metric` | string | `annual_sales` / `visit_count` / `last_visit` |
| `year` | int | 年間売上ランキングの対象年 |
| `from` | date | 集計開始日 |
| `to` | date | 集計終了日 |
| `period_preset` | string | 期間プリセット |
| `amount_basis` | string | 売上基準 |
| `last_visit_bucket` | string | 最終来院分類 |
| `search` | string | 飼い主名検索 |
| `include_zero` | bool | 0円・0回の含有制御 |
| `include_no_visit` | bool | 来院なしの含有制御 |
| `sort` | string | 並び順 |
| `order` | string | `asc` / `desc` |
| `page` | integer | ページ番号 |
| `per_page` | integer | 1ページ件数 |

### 5.3 レスポンス

```json
{
  "total": 142,
  "page": 1,
  "per_page": 50,
  "metric": "annual_sales",
  "owners": [
    {
      "owner_id": "1",
      "owner_name": "山田 太郎",
      "annual_amount": 156000,
      "billing_count": 12,
      "period_visit_count": 4,
      "total_visit_count": 12,
      "last_visit_date": "2026-03-10",
      "days_since_last_visit": 47,
      "last_visit_bucket": "over_3m",
      "first_visit_date": "2022-05-01"
    }
  ]
}
```

### 5.4 エラー

- 400: パラメータ不正
- 401: 認証エラー
- 403: 権限不足
- 404: クリニックまたは対象データなし

---

## 6. 画面仕様

### 6.1 画面構成

- 画面名: `顧客集計`
- URL: `/owners/aggregations`
- 既定表示: 売上ランキング

### 6.2 タブ構成

| タブ | 内容 |
|------|------|
| 売上ランキング | 年間・任意期間の売上順一覧 |
| 来院回数 | 任意期間の来院回数順一覧 |
| 最終来院 | 最終来院日の古い順・分類別一覧 |

### 6.3 共通UI

| UI | 内容 |
|----|------|
| 期間指定 | `今年` / `昨年` / `過去3ヶ月` / `過去6ヶ月` / `過去12ヶ月` / `カスタム` |
| 飼い主名検索 | 部分一致 |
| 並び替え | 売上、来院回数、最終来院日、飼い主名 |
| 金額フィルタ | 下限・上限 |
| 来院回数フィルタ | 下限・上限 |
| 最終来院分類 | 3ヶ月未満、3ヶ月以上、6ヶ月以上、1年以上、来院なし |
| CSV出力 | 表示中または選択中の一覧を出力 |

### 6.4 CSV

CSV は UTF-8 BOM 付きとし、表示中の集計軸に応じた列順で出力する。

推奨ヘッダ:

```text
owner_id,owner_name,annual_amount,billing_count,period_visit_count,total_visit_count,last_visit_date,days_since_last_visit,last_visit_bucket,first_visit_date
```

---

## 7. 権限・セキュリティ

| 操作 | 権限 |
|------|------|
| 顧客集計一覧の閲覧 | `owners:view` |
| CSV出力 | `owners:view` |

データ分離:

- すべての集計で `clinic_id` を必須条件にする。
- URL 上の `:clinic_id` は信頼せず、認証コンテキストの clinic ID を使う。
- `owners.deleted_at IS NULL`、`medical_records.deleted_at IS NULL`、`billings.deleted_at IS NULL` を必ず条件に含める。

---

## 8. パフォーマンス

### 8.1 現行インデックス

| インデックス | 用途 |
|--------------|------|
| `idx_owners_clinic_id` | clinic 内の飼い主抽出 |
| `idx_owners_name_trgm` | 飼い主名検索 |
| `idx_medical_records_clinic_owner` | 飼い主単位のカルテ集計 |
| `idx_medical_records_clinic_date` | 日付範囲でのカルテ集計 |
| `idx_billings_owner_id` | 飼い主単位の会計集計 |
| `idx_billings_clinic_date` | 会計一覧 |
| `idx_billings_clinic_status` | 会計ステータス絞り込み |

### 8.2 追加検討インデックス

年間売上ランキングを `completed_at` 基準で集計する場合、現行の `scheduled_date` インデックスだけでは不足する可能性がある。

```sql
CREATE INDEX idx_billings_clinic_completed_owner
  ON billings(clinic_id, completed_at, owner_id)
  WHERE deleted_at IS NULL AND status = 'completed';
```

来院分類を頻繁に使う場合:

```sql
CREATE INDEX idx_medical_records_clinic_owner_date
  ON medical_records(clinic_id, owner_id, date DESC)
  WHERE deleted_at IS NULL;
```

### 8.3 ページング

件数が増える場合は以下を検討する。

- 売上・来院回数・最終来院日の集計結果を materialized view か日次スナップショットに保存する
- フィルタが少ない場合は DB 側で `LIMIT/OFFSET` する

---

## 9. 実装差分チェックリスト

| 項目 | 現状 | 必要対応 |
|------|------|----------|
| 年間売上ランキング | 累計診療費のみ。期間指定なし | `year` / `from` / `to` と期間内売上集計を追加 |
| 売上基準 | `total_amount` 固定 | `amount_basis` を追加 |
| 任意期間の来院回数 | 過去365日の指標のみ | `period_visit_count` を追加 |
| 最終来院分類 | `last_visit_date` は表示済み | `days_since_last_visit` と `last_visit_bucket` を追加 |
| 飼い主名検索 | FE 型のみ存在 | BE で `search` を実装 |
| ソートUI | 現行画面に未露出 | テーブルヘッダまたはセレクトで追加 |
| 0件表示 | 全件を返し得る | 集計種別ごとに `include_zero` を制御 |

---

## 10. テスト観点

### 10.1 バックエンド

- 年間売上ランキングで、期間外の会計が含まれない
- `status != completed` の会計が売上ランキングに含まれない
- 返金がある場合、`net_paid_amount` が返金額分だけ減る
- 同じ飼い主の同日複数カルテが来院1回として数えられる
- `from` / `to` の境界日が含まれる
- `last_visit_bucket` が 89/90/179/180/364/365 日境界で正しい
- 来院なしの飼い主が `no_visit` になる
- 他 clinic の飼い主・カルテ・会計が混入しない
- `search` が `name` に効く
- `page` / `per_page` の境界値と最大200件制限が効く

### 10.2 フロントエンド

- 各タブで期間・フィルタ変更時に `page` が1へ戻る
- 売上ランキングが金額降順で表示される
- 来院回数が指定期間の値として表示される
- 最終来院分類の表示ラベルが経過日数と一致する
- CSV出力が表示中/選択中の対象を正しく出す
