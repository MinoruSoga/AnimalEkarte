# LTV/顧客集計機能 仕様書

> **ステータス**: Draft  
> **対象範囲**: 飼い主単位の売上・来院回数・最終来院日集計  
> **最終更新**: 2026-04-26  
> **関連実装**: `/owners/ltv` LTV/CPM ダッシュボード、`GET /api/v1/clinics/:clinic_id/owners/ltv`

---

## 1. 目的

飼い主単位で診療費・来院状況を集計し、以下の業務に使える一覧を提供する。

| # | 集計 | 目的 |
|---|------|------|
| 1 | 顧客ごとの年間売上ランキング | 年間の累計診療費が高い飼い主を把握し、優良顧客施策や経営分析に使う |
| 2 | 顧客ごとの来院回数 | 任意期間内の来院頻度を把握し、継続来院・離脱兆候を確認する |
| 3 | 最終来院日での分類 | 最終来院からの経過期間別に飼い主を分類し、3ヶ月・6ヶ月・1年以上未来院のフォロー対象を抽出する |

この機能は既存の LTV/CPM ダッシュボードを基盤にする。現行実装で取得できる値と、追加実装が必要な値を分けて定義する。

---

## 2. 現行実装サマリー

### 2.1 画面

| 項目 | 内容 |
|------|------|
| 画面名 | LTV/顧客集計ダッシュボード |
| URL | `/owners/ltv` |
| フロント実装 | `frontend/src/features/owners/ltv/LtvDashboardPage.tsx` |
| APIクライアント | `frontend/src/features/owners/api/get-ltv-owners.ts` |
| 表示カラム | 飼い主名、LINE連携、CPMステージ、累計診療費、年間来院、累計来院、最終来院日、初診日 |
| CSV出力 | 表示中または選択中の飼い主を `ltv-owners-YYYY-MM-DD.csv` として出力 |

### 2.2 API

| 項目 | 内容 |
|------|------|
| 現行エンドポイント | `GET /api/v1/clinics/:clinic_id/owners/ltv` |
| 旧互換エンドポイント | `GET /api/v1/owners/ltv` |
| ハンドラ | `backend/internal/handler/ltv_handler.go` |
| サービス | `backend/internal/service/ltv_service.go` |
| リポジトリ | `backend/internal/repository/ltv_repository.go` |
| 権限 | フロントは `/owners` 配下のため `owners:view` が必要。バックエンド側は protected route 上で認証済み clinic context を使う |

`/v1/clinics/:clinic_id/...` の `:clinic_id` はフロント都合のURLパラメータであり、バックエンド集計では JWT/認証コンテキストから取得した clinic ID を正とする。

### 2.3 現行レスポンス

```json
{
  "total": 142,
  "page": 1,
  "per_page": 50,
  "owners": [
    {
      "owner_id": "1",
      "owner_name": "山田 太郎",
      "line_user_id_masked": "U****abcd",
      "has_line": true,
      "total_amount": 156000,
      "total_fee": 156000,
      "total_visit_count": 12,
      "annual_visit_count": 4,
      "last_visit_date": "2026-03-10",
      "first_visit_date": "2022-05-01",
      "cpm_stage": "noah"
    }
  ]
}
```

`total_fee` はフロント互換のための `total_amount` エイリアスである。画面側では `total_fee` を表示している。

### 2.4 現行クエリパラメータ

| パラメータ | 型 | 現行挙動 |
|------------|----|----------|
| `page` | int | 1以上。デフォルト `1` |
| `per_page` | int | 1〜200。デフォルト `50` |
| `sort` | string | `total_fee` / `total_amount` / `annual_visit_count` / `visit_count` / `last_visit_date` / `last_visit` |
| `order` | string | `asc` / `desc`。未指定または不正値は `desc` |
| `min_total_amount` | int | 累計診療費下限。`min_total_fee` も互換対応 |
| `max_total_amount` | int | 累計診療費上限。`max_total_fee` も互換対応 |
| `min_visit_count` | int | 累計来院回数下限 |
| `cpm_stage` | string | `encounter` / `growing` / `core` / `noah` / `spot` / `dormant` |
| `line_linked` | bool | `true` の場合 LINE連携済みのみ。`has_line=true` も互換対応 |

注意点:

- フロント型には `search` があるが、現行バックエンドは飼い主名検索を未実装。
- フロント型には `sort` があるが、現行画面に明示的なソートUIはない。
- `sort=annual_visit_count` は現行ハンドラで `visit_count_*` に変換され、リポジトリでは `total_visit_count` で並ぶ。年間来院回数そのもののソートではない。
- `cpm_stage` フィルタは DB ではなく Go サービス層で計算後に適用する。そのためページングも Go 側で行う。

---

## 3. データソースと集計単位

### 3.1 集計単位

集計単位は常に `owners.id`、つまり飼い主単位とする。同一飼い主に複数ペットがいる場合、全ペットの診療記録・会計を合算する。

| エンティティ | 用途 | 主な参照カラム |
|--------------|------|----------------|
| `owners` | 集計対象の飼い主 | `id`, `clinic_id`, `name`, `line_user_id`, `lstep_opt_out`, `deleted_at` |
| `medical_records` | 来院回数、初診日、最終来院日 | `owner_id`, `clinic_id`, `date`, `deleted_at` |
| `billings` | 診療費・売上 | `medical_record_id`, `owner_id`, `clinic_id`, `total_amount`, `status`, `completed_at`, `deleted_at` |
| `payments` | 実支払額を使う場合の支払情報 | `billing_id`, `billing_amount`, `deleted_at` |
| `billing_refunds` | 純売上を使う場合の返金額 | `billing_id`, `amount` |

### 3.2 現行LTV APIの集計定義

現行 `FindOwnerLTV` は以下の考え方で集計している。

| 項目 | 現行定義 |
|------|----------|
| 対象飼い主 | `owners.clinic_id = clinicID` かつ `owners.deleted_at IS NULL` |
| 来院元 | `medical_records`。`mr.owner_id = owners.id`、`mr.clinic_id = owners.clinic_id`、`mr.deleted_at IS NULL` |
| 累計診療費 | `COALESCE(SUM(billings.total_amount), 0)` |
| 会計JOIN | `billings.medical_record_id = medical_records.id`、`billings.clinic_id = owners.clinic_id`、`billings.deleted_at IS NULL` |
| 会計ステータス | 現行LTV APIでは `status = completed` 条件なし |
| 累計来院回数 | `COUNT(DISTINCT medical_records.date)` |
| 年間来院回数 | `COUNT(DISTINCT CASE WHEN medical_records.date >= NOW() - INTERVAL '365 days' THEN medical_records.date END)` |
| 最終来院日 | `MAX(medical_records.date)` |
| 初診日 | `MIN(medical_records.date)` |

現在の LTV 表示値は「会計確定済みの実入金」ではなく、カルテに紐づく `billings.total_amount` の合計である。レジ締め・月次売上レポートと同じ実売上に合わせる場合は、`status = completed`、`completed_at`、`payments.billing_amount`、`billing_refunds.amount` を使う集計へ切り替える。

---

## 4. 共通仕様

### 4.1 日付・期間

| 項目 | 仕様 |
|------|------|
| タイムゾーン | 集計条件・表示日は Asia/Tokyo 基準 |
| 日付形式 | APIクエリ、レスポンスとも `YYYY-MM-DD` |
| 期間の扱い | `from` は含む、`to` も日付として含む。SQLでは `>= from` かつ `< to + 1 day` に正規化する |
| デフォルト対象年 | 年間売上ランキングは指定がなければ当年の `01-01` から `12-31` |
| ローリング年間 | LTV/CPM の `annual_visit_count` は現行通り過去365日。カレンダー年の年間集計とは別指標 |

### 4.2 金額

| 指標 | 定義 | 用途 |
|------|------|------|
| `gross_total_amount` | `billings.total_amount` の合計 | 現行LTV互換。請求総額ベース |
| `paid_amount` | `payments.billing_amount` の合計 | 会計完了後の実請求額ベース |
| `net_paid_amount` | `payments.billing_amount - billing_refunds.amount` | 返金控除後の売上。レジ締め・月次レポートに近い |

顧客ごとの年間売上ランキングの**既定の集計基準は `gross_total_amount`** とする。これは現行LTV/CPMダッシュボードと同じ定義で、飼い主の年間累計診療費として最も自然に読めるためである。

ただし、経理向けの売上分析や返金控除後の実売上を見たい場合は `net_paid_amount` を別モードとして選べるようにする。

### 4.3 来院回数

来院回数は飼い主単位の `medical_records.date` で数える。現行LTV APIに合わせ、同一飼い主が同じ日に複数カルテを持つ場合は1回とする。

例:

| ケース | カウント |
|--------|----------|
| 同じ飼い主の同じペットで同日2カルテ | 1回 |
| 同じ飼い主の別ペットで同日2カルテ | 1回 |
| 同じ飼い主で別日2カルテ | 2回 |

Lステップタグ同期用の `FindOwnerVisitSummary` は `COUNT(*)` を使う箇所があるため、現行実装内で「同日複数カルテ」の扱いに差がある。顧客集計画面では `COUNT(DISTINCT date)` に統一する。

### 4.4 ゼロ件の扱い

| 集計 | デフォルト | 理由 |
|------|------------|------|
| 年間売上ランキング | 売上0円は除外 | ランキング用途のため |
| 来院回数一覧 | 0回来院はオプションで含める | 未来院者抽出に使う場合があるため |
| 最終来院日分類 | 来院なしを `no_visit` として含める | 新規登録のみ・未受診の飼い主を把握するため |

---

## 5. 機能仕様

### 5.1 顧客ごとの年間売上ランキング

#### 5.1.1 概要

指定年または指定期間内の診療費を飼い主単位で合計し、売上降順に並べる一覧。

#### 5.1.2 入力条件

| 条件 | 型 | デフォルト | 説明 |
|------|----|------------|------|
| `year` | int | 当年 | カレンダー年指定。`from` / `to` がある場合は使わない |
| `from` | date | `year-01-01` | 集計開始日 |
| `to` | date | `year-12-31` | 集計終了日 |
| `amount_basis` | string | `gross_total_amount` | `gross_total_amount` / `paid_amount` / `net_paid_amount` |
| `min_amount` | int | なし | 金額下限 |
| `max_amount` | int | なし | 金額上限 |
| `include_zero` | bool | `false` | 0円の飼い主も含めるか |
| `sort` | string | `annual_amount` | `annual_amount` / `visit_count` / `last_visit_date` |
| `order` | string | `desc` | `asc` / `desc` |
| `page` | int | `1` | ページ番号 |
| `per_page` | int | `50` | 1ページ件数。最大200 |

#### 5.1.3 表示項目

| カラム | 内容 |
|--------|------|
| 順位 | 売上順の連番。同額の場合は同順位または `owner_id` 昇順で安定ソート |
| 飼い主名 | `owners.name`。クリックで飼い主詳細へ遷移 |
| 年間診療費 | 指定期間内の診療費合計 |
| 会計件数 | 指定期間内の対象会計件数 |
| 期間内来院回数 | 指定期間内の `COUNT(DISTINCT medical_records.date)` |
| 最終来院日 | 全期間の `MAX(medical_records.date)` |
| 初診日 | 全期間の `MIN(medical_records.date)` |
| LINE連携 | `owners.line_user_id` の有無 |
| CPMステージ | 既存 LTV/CPM と同じ計算値 |

#### 5.1.4 推奨SQL

経理上の売上ランキングとして扱う場合は、会計完了日時と実支払額を基準にする。

```sql
WITH billing_agg AS (
  SELECT
    b.owner_id,
    COALESCE(SUM(p.billing_amount), 0)
      - COALESCE(SUM(r.refund_amount), 0) AS annual_amount,
    COUNT(DISTINCT b.id) AS billing_count
  FROM billings b
  LEFT JOIN payments p
    ON p.billing_id = b.id
   AND p.deleted_at IS NULL
  LEFT JOIN (
    SELECT billing_id, SUM(amount) AS refund_amount
    FROM billing_refunds
    GROUP BY billing_id
  ) r
    ON r.billing_id = b.id
  WHERE b.clinic_id = @clinic_id
    AND b.deleted_at IS NULL
    AND b.status = 'completed'
    AND b.completed_at AT TIME ZONE 'Asia/Tokyo' >= @from
    AND b.completed_at AT TIME ZONE 'Asia/Tokyo' < @to_plus_1_day
  GROUP BY b.owner_id
),
period_visits AS (
  SELECT
    owner_id,
    COUNT(DISTINCT date) AS period_visit_count
  FROM medical_records
  WHERE clinic_id = @clinic_id
    AND deleted_at IS NULL
    AND date >= @from::date
    AND date <= @to::date
  GROUP BY owner_id
),
all_visits AS (
  SELECT
    owner_id,
    MAX(date) AS last_visit_date,
    MIN(date) AS first_visit_date
  FROM medical_records
  WHERE clinic_id = @clinic_id
    AND deleted_at IS NULL
  GROUP BY owner_id
)
SELECT
  o.id AS owner_id,
  o.name AS owner_name,
  COALESCE(ba.annual_amount, 0) AS annual_amount,
  COALESCE(ba.billing_count, 0) AS billing_count,
  COALESCE(pv.period_visit_count, 0) AS period_visit_count,
  av.last_visit_date,
  av.first_visit_date
FROM owners o
LEFT JOIN billing_agg ba ON ba.owner_id = o.id
LEFT JOIN period_visits pv ON pv.owner_id = o.id
LEFT JOIN all_visits av ON av.owner_id = o.id
WHERE o.clinic_id = @clinic_id
  AND o.deleted_at IS NULL
  AND COALESCE(ba.annual_amount, 0) > 0
ORDER BY annual_amount DESC, o.id ASC;
```

現行LTV API互換の診療費ランキングにする場合は、`payments` / `billing_refunds` ではなく `billings.total_amount` を合計する。ただし `status` 未指定だと未完了・キャンセル予定の金額が混入する可能性があるため、年間売上ランキングでは `status = completed` を必須にする。

---

### 5.2 顧客ごとの来院回数

#### 5.2.1 概要

指定期間内の来院回数を飼い主単位で集計し、来院回数順または最終来院日順で一覧化する。

#### 5.2.2 入力条件

| 条件 | 型 | デフォルト | 説明 |
|------|----|------------|------|
| `from` | date | なし | 集計開始日。未指定なら `period_preset` か UI デフォルトに従う |
| `to` | date | なし | 集計終了日。未指定なら当日 |
| `period_preset` | string | `last_12_months` | `last_3_months` / `last_6_months` / `last_12_months` / `calendar_year` |
| `min_visit_count` | int | なし | 来院回数下限 |
| `max_visit_count` | int | なし | 来院回数上限 |
| `include_zero` | bool | `false` | 0回来院の飼い主を含めるか |
| `sort` | string | `visit_count` | `visit_count` / `last_visit_date` / `owner_name` |
| `order` | string | `desc` | `asc` / `desc` |

#### 5.2.3 表示項目

| カラム | 内容 |
|--------|------|
| 飼い主名 | `owners.name` |
| 期間内来院回数 | 指定期間内の `COUNT(DISTINCT medical_records.date)` |
| 累計来院回数 | 全期間の `COUNT(DISTINCT medical_records.date)` |
| 最終来院日 | 全期間の `MAX(medical_records.date)` |
| 期間内最終来院日 | 指定期間内の `MAX(medical_records.date)` |
| 初診日 | 全期間の `MIN(medical_records.date)` |
| 年間来院回数 | 過去365日。既存 `annual_visit_count` と同じ |

#### 5.2.4 推奨SQL

```sql
SELECT
  o.id AS owner_id,
  o.name AS owner_name,
  COUNT(DISTINCT CASE
    WHEN mr.date >= @from AND mr.date <= @to THEN mr.date
  END) AS period_visit_count,
  COUNT(DISTINCT mr.date) AS total_visit_count,
  COUNT(DISTINCT CASE
    WHEN mr.date >= CURRENT_DATE - INTERVAL '365 days' THEN mr.date
  END) AS rolling_annual_visit_count,
  MAX(mr.date) AS last_visit_date,
  MAX(CASE WHEN mr.date >= @from AND mr.date <= @to THEN mr.date END) AS period_last_visit_date,
  MIN(mr.date) AS first_visit_date
FROM owners o
LEFT JOIN medical_records mr
  ON mr.owner_id = o.id
 AND mr.clinic_id = o.clinic_id
 AND mr.deleted_at IS NULL
WHERE o.clinic_id = @clinic_id
  AND o.deleted_at IS NULL
GROUP BY o.id, o.name
HAVING COUNT(DISTINCT CASE
  WHEN mr.date >= @from AND mr.date <= @to THEN mr.date
END) > 0
ORDER BY period_visit_count DESC, last_visit_date DESC NULLS LAST, o.id ASC;
```

---

### 5.3 最終来院日での分類

#### 5.3.1 概要

飼い主ごとの最終来院日を表示し、最終来院からの経過日数で分類する。

#### 5.3.2 分類

| 分類キー | 条件 | 表示ラベル | 想定用途 |
|----------|------|------------|----------|
| `within_3m` | 0〜89日 | 3ヶ月未満 | 通常フォロー |
| `over_3m` | 90〜179日 | 3ヶ月以上 | 軽い再来院促進 |
| `over_6m` | 180〜364日 | 6ヶ月以上 | 未来院リカバリー |
| `over_1y` | 365日以上 | 1年以上 | 長期休眠向け案内 |
| `no_visit` | 来院記録なし | 来院なし | 登録のみ・未受診 |

上記は相互排他的な分類である。フィルタとして「3ヶ月以上」を指定する場合は `days_since_last_visit >= 90` とし、6ヶ月以上・1年以上も含む。

#### 5.3.3 既存Lステップ休眠タグとの関係

既存の Lステップ休眠判定は以下の閾値を使う。

| 閾値 | 自動タグ | 備考 |
|------|----------|------|
| 180日 | `dormant_180d` | 6ヶ月未来院リカバリー |
| 210日 | `dormant_210d` | 240日前の追加接触 |
| 240日 | `dormant_240d`, `cpm_dormant` | CPM上の休眠扱い |
| 365日 | `dormant_365d` | 1年以上休眠 |

顧客集計画面の分類は、業務上わかりやすい `3ヶ月 / 6ヶ月 / 1年以上` を表示軸にする。Lステップ自動タグとは閾値が完全一致しないため、画面分類とタグ状態を混同しない。

#### 5.3.4 入力条件

| 条件 | 型 | デフォルト | 説明 |
|------|----|------------|------|
| `last_visit_bucket` | string | なし | `within_3m` / `over_3m` / `over_6m` / `over_1y` / `no_visit` |
| `min_days_since_last_visit` | int | なし | 経過日数下限。例: 90 |
| `max_days_since_last_visit` | int | なし | 経過日数上限 |
| `include_no_visit` | bool | `true` | 来院なしを含めるか |
| `sort` | string | `last_visit_date` | `last_visit_date` / `days_since_last_visit` / `owner_name` |
| `order` | string | `asc` | 最終来院が古い順をデフォルトにする |

#### 5.3.5 表示項目

| カラム | 内容 |
|--------|------|
| 飼い主名 | `owners.name` |
| 最終来院日 | `MAX(medical_records.date)` |
| 経過日数 | `CURRENT_DATE - last_visit_date` |
| 分類 | `within_3m` / `over_3m` / `over_6m` / `over_1y` / `no_visit` |
| 累計来院回数 | 全期間の `COUNT(DISTINCT medical_records.date)` |
| 年間来院回数 | 過去365日の `COUNT(DISTINCT medical_records.date)` |
| 累計診療費 | 既存LTVと同じ `total_fee` |
| CPMステージ | 既存ロジックで計算 |

#### 5.3.6 推奨SQL

```sql
WITH owner_visits AS (
  SELECT
    o.id AS owner_id,
    o.name AS owner_name,
    MAX(mr.date) AS last_visit_date,
    MIN(mr.date) AS first_visit_date,
    COUNT(DISTINCT mr.date) AS total_visit_count,
    COUNT(DISTINCT CASE
      WHEN mr.date >= CURRENT_DATE - INTERVAL '365 days' THEN mr.date
    END) AS annual_visit_count
  FROM owners o
  LEFT JOIN medical_records mr
    ON mr.owner_id = o.id
   AND mr.clinic_id = o.clinic_id
   AND mr.deleted_at IS NULL
  WHERE o.clinic_id = @clinic_id
    AND o.deleted_at IS NULL
  GROUP BY o.id, o.name
)
SELECT
  owner_id,
  owner_name,
  last_visit_date,
  CASE
    WHEN last_visit_date IS NULL THEN NULL
    ELSE CURRENT_DATE - last_visit_date
  END AS days_since_last_visit,
  CASE
    WHEN last_visit_date IS NULL THEN 'no_visit'
    WHEN CURRENT_DATE - last_visit_date >= 365 THEN 'over_1y'
    WHEN CURRENT_DATE - last_visit_date >= 180 THEN 'over_6m'
    WHEN CURRENT_DATE - last_visit_date >= 90 THEN 'over_3m'
    ELSE 'within_3m'
  END AS last_visit_bucket,
  total_visit_count,
  annual_visit_count,
  first_visit_date
FROM owner_visits
ORDER BY last_visit_date ASC NULLS FIRST, owner_id ASC;
```

---

## 6. API設計方針

### 6.1 短期方針: 既存LTV APIの互換拡張

既存画面・既存クライアントを壊さないため、まずは `GET /api/v1/clinics/:clinic_id/owners/ltv` を拡張する。

追加候補:

| パラメータ | 説明 |
|------------|------|
| `search` | 飼い主名部分一致。`owners.name` または `owners.name_kana` |
| `from` / `to` | 来院回数・売上の任意期間 |
| `year` | 年間売上ランキング用の対象年 |
| `amount_basis` | `gross_total_amount` / `paid_amount` / `net_paid_amount` |
| `last_visit_bucket` | 最終来院分類 |
| `min_days_since_last_visit` / `max_days_since_last_visit` | 最終来院からの経過日数フィルタ |
| `include_zero` | 売上0円・来院0回を含める |

互換維持:

- 既存の `total_amount` / `total_fee` は残す。
- 既存の `annual_visit_count` は過去365日のままにする。
- 新しくカレンダー年の来院回数を返す場合は `period_visit_count` のように別名にする。
- `sort=annual_visit_count` の挙動は、仕様通り年間来院回数で並ぶよう修正する。累計来院回数ソートは `sort=total_visit_count` を追加する。

### 6.2 本設計の採用方針

このプロジェクトでは、集計機能を既存の `owners` / `ltv` 系ルートに寄せて拡張する。

理由:

- 既存の `owners` 一覧・詳細・LTV 画面が feature 内で完結しており、UI も API も同じ責務に揃っている
- Handler → Service → Repository の薄い分割を崩さずに実装できる
- `frontend/src/features/owners/ltv/*` に既存の参照実装があり、そこへ集約するのが最も自然

したがって、売上ランキング・来院回数・最終来院分類はすべて `GET /api/v1/clinics/:clinic_id/owners/ltv` の拡張として提供する。

拡張時は、集計の種類が増えてもレスポンス構造を分岐させず、`owners[]` の各行に必要な派生値を追加する。画面側はタブまたはフィルタで表示軸を切り替える。

追加の一覧 API を設ける場合は、用途が完全に異なるレポートのみを対象にする。今回の3要件はその対象にしない。

---

## 7. 画面仕様

### 7.1 配置

既存 `/owners/ltv` を「LTV/顧客集計」画面として拡張する。メニュー名は現行の `LTV/CPM` から、業務用途に応じて `LTV/顧客集計` へ変更する。

画面の既定タブは `売上ランキング` にする。既定フィルタは以下。

- 売上ランキング: `year=当年`, `amount_basis=gross_total_amount`, `sort=annual_amount`, `order=desc`
- 来院回数: `period_preset=last_12_months`, `sort=visit_count`, `order=desc`
- 最終来院: `last_visit_bucket=over_3m`, `sort=last_visit_date`, `order=asc`

### 7.2 タブ構成

| タブ | 内容 |
|------|------|
| 売上ランキング | 年間・任意期間の売上順一覧 |
| 来院回数 | 任意期間の来院回数順一覧 |
| 最終来院 | 最終来院日の古い順・分類別一覧 |
| LTV/CPM | 既存のLTV/CPM抽出、Lステップ連携 |

初期実装を小さくする場合はタブを作らず、既存テーブルにフィルタと分類カラムを追加する。

### 7.3 共通UI

| UI | 内容 |
|----|------|
| 期間指定 | `今年` / `昨年` / `過去3ヶ月` / `過去6ヶ月` / `過去12ヶ月` / `カスタム` |
| 飼い主名検索 | `owners.name` / `owners.name_kana` の部分一致 |
| 並び替え | 売上、来院回数、最終来院日、飼い主名 |
| 金額フィルタ | 下限・上限 |
| 来院回数フィルタ | 下限・上限 |
| 最終来院分類 | 3ヶ月未満、3ヶ月以上、6ヶ月以上、1年以上、来院なし |
| CSV出力 | 表示中または選択中の一覧を出力 |

### 7.4 CSV

CSVはUTF-8 BOM付きとし、現行 `LtvDashboardPage` と同じ方式を使う。

推奨ヘッダ:

```text
owner_id,owner_name,annual_amount,billing_count,period_visit_count,total_visit_count,annual_visit_count,last_visit_date,days_since_last_visit,last_visit_bucket,first_visit_date,has_line,cpm_stage
```

---

## 8. 権限・セキュリティ

| 操作 | 権限 |
|------|------|
| LTV/顧客集計一覧の閲覧 | `owners:view` |
| CSV出力 | `owners:view` |
| Lステップ一括タグ付与 | `owners:edit` |

データ分離:

- すべての集計で `clinic_id` を必須条件にする。
- URL上の `:clinic_id` は信頼せず、認証コンテキストの clinic ID を使う。
- `owners.deleted_at IS NULL`、`medical_records.deleted_at IS NULL`、`billings.deleted_at IS NULL` を必ず条件に含める。

個人情報:

- `line_user_id` はレスポンスに生値を返さず、現行通り `line_user_id_masked` のみ返す。
- CSVにも `line_user_id` 生値は含めない。

---

## 9. パフォーマンス

### 9.1 現行インデックス

関連する主な既存インデックス:

| インデックス | 用途 |
|--------------|------|
| `idx_owners_clinic_id` | clinic内の飼い主抽出 |
| `idx_owners_name_trgm` | 飼い主名検索 |
| `idx_medical_records_clinic_owner` | 飼い主単位のカルテ集計 |
| `idx_medical_records_clinic_date` | 日付範囲でのカルテ集計 |
| `idx_billings_owner_id` | 飼い主単位の会計集計 |
| `idx_billings_clinic_date` | `scheduled_date` ベースの会計一覧 |
| `idx_billings_clinic_status` | 会計ステータス絞り込み |

### 9.2 追加検討インデックス

年間売上ランキングを `completed_at` 基準で集計する場合、現行の `scheduled_date` インデックスだけでは不足する可能性がある。

候補:

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

### 9.3 ページング

現行LTV APIは CPMステージフィルタを Go 側で適用するため、DBから全件取得後にページングしている。飼い主数が増える場合は以下を検討する。

- CPMステージを定期計算してキャッシュテーブル化する。
- 売上・来院回数・最終来院日の集計結果を materialized view か日次スナップショットに保存する。
- CPMフィルタなしの場合はDB側で `LIMIT/OFFSET` する。

---

## 10. 実装差分チェックリスト

現行実装から、ユーザー要望の3機能を完成させるための差分は以下。

| 項目 | 現状 | 必要対応 |
|------|------|----------|
| 年間売上ランキング | 累計LTVのみ。期間指定なし | `year` / `from` / `to` と期間内売上集計を追加 |
| 売上の会計ステータス | LTV APIは `status` 条件なし | 売上ランキングでは `completed` のみにする |
| 返金控除 | LTV APIは未控除 | 経理用途では `billing_refunds` を控除する |
| 任意期間の来院回数 | 過去365日の `annual_visit_count` と累計のみ | `period_visit_count` を追加 |
| 最終来院分類 | `last_visit_date` は表示済み | `days_since_last_visit` と `last_visit_bucket` を追加 |
| 飼い主名検索 | FE型のみ存在 | BEで `search` を実装 |
| 年間来院回数ソート | `annual_visit_count` 指定でも累計来院順 | 年間来院回数でのDB/Goソートに修正 |
| ソートUI | 現行画面に未露出 | テーブルヘッダまたはセレクトで追加 |
| 0件表示 | LTV APIは全飼い主を返し得る | 集計種別ごとに `include_zero` を制御 |

---

## 11. テスト観点

### 11.1 バックエンド

- 年間売上ランキングで、期間外の会計が含まれない。
- `status != completed` の会計が売上ランキングに含まれない。
- 返金がある場合、`net_paid_amount` が返金額分だけ減る。
- 同じ飼い主の同日複数カルテが来院1回として数えられる。
- `from` / `to` の境界日が含まれる。
- `last_visit_bucket` が 89/90/179/180/364/365 日境界で正しい。
- 来院なしの飼い主が `no_visit` になる。
- 他 clinic の飼い主・カルテ・会計が混入しない。
- `search` が `name` / `name_kana` に効く。
- `page` / `per_page` の境界値と最大200件制限が効く。

### 11.2 フロントエンド

- 各タブで期間・フィルタ変更時に `page` が1へ戻る。
- 売上ランキングが金額降順で表示される。
- 来院回数が指定期間の値として表示される。
- 最終来院分類の表示ラベルが経過日数と一致する。
- CSV出力が表示中/選択中の対象を正しく出す。
- LINE User ID の生値が画面・CSVに出ない。

---

## 12. 関連ドキュメント

- `docs/line/lstep-integration.md` — LTV、年間来院回数、CPM、休眠判定の全体仕様
- `docs/tasks/closed/lstep/LSTEP-BE-010-ltv-aggregation-api.md` — LTV集計APIの実装タスク
- `docs/tasks/closed/lstep/LSTEP-FE-005-ltv-cpm-dashboard.md` — LTV/CPMダッシュボード仕様
- `docs/CASH_REGISTER_SPEC.md` — レジ締め・売上集計仕様
- `docs/ERD.md` — 関連テーブル定義
