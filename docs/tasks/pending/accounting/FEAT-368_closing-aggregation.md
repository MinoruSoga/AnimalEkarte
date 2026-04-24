# FEAT-368: レジ締め・日次集計・月間売上集計・締め時間設定

**作成日**: 2026-04-14  
**更新日**: 2026-04-20  
**Status**: Pending（全確定。実装着手待ち）  
**Priority**: HIGH（業務運用必須機能）  
**Affects**: `features/accounting`, `billing_items.category`, `payments.method`（マスタ化）, 新規テーブル `cash_register_closes`, `closing_special_periods`, `clinic_holidays`, `payment_methods`

---

## 要求定義（原文）

### 要求1 — レジ締め・部門別集計・レジ金突合・印刷
> 現金・クレジットでの締めごとで集計  
> 診療・外科・RV・フード・トリミング・ホテル・用品・トレセンの締めごとに集計  
> レジの締めを行う際に、レジ金と突合する必要があるため、必ずできるようにしてほしい  
> 退院も一覧に載せてほしい  
> その一覧は印刷できるようにしてほしい

### 要求2 — AM/PM中間締め・月間売上集計・締め時間設定
> 日次集計と中間計  
> 平日土 AM締め（診療終了18:30 / 〜13:59）、PM締め（14:00〜18:29）  
> 日曜 AM締め（診療終了17:30 / 〜13:59）、PM締め（14:00〜17:29）  
> 年末年始などで時間短縮があるので時間を変更できる仕様  
> 受付では不要だが経理で月間売上集計が必要

---

## 現状評価

| 要望 | 現状 |
|------|------|
| 支払方法別集計 | `payments.method` enum は存在するが集計 API／画面は**未実装**。マスタ化が必要 |
| 部門別集計（8カテゴリ） | `billing_items.category` が要望8部門と**不一致**（後述） |
| レジ金突合 | テーブル・API・画面**すべて未実装** |
| 退院会計の一覧掲載 | `billings.hospitalization_id` で紐付け可能（データ基盤あり） |
| 印刷 | 個別領収書のみ。集計一覧の印刷は**未実装** |
| AM/PM 中間締め | **未実装** |
| 締め時間設定画面 | **未実装** |
| 月間売上集計 | **未実装** |
| 休診日設定 | **未実装** |
| 支払方法マスタ | **未実装**（enum 固定のまま） |

---

## 1. 締め区分の定義

### 1.1 AM/PM 境界時刻

診療日を AM / PM の2区分に分割し、各時間帯の会計実績を集計する。  
**祝日は関係なし（祝日も通常営業）。休診日は 1.3 の設定で管理する。**

| 曜日区分 | 区分 | 集計対象時間帯 | 診療終了 |
|---------|------|--------------|---------|
| 平日・土曜 | AM締め | 診療開始 〜 13:59 | 18:30 |
| 平日・土曜 | PM締め | 14:00 〜 18:29 | 18:30 |
| 日曜 | AM締め | 診療開始 〜 13:59 | 17:30 |
| 日曜 | PM締め | 14:00 〜 17:29 | 17:30 |

境界時刻・診療終了時刻は設定画面から変更可能（Section 5）。

### 1.2 特別期間（年末年始・お盆等）

特定日付範囲で標準設定を上書き可能。特別期間設定が存在する日は**標準設定より優先**。

| 設定項目 | 例 |
|---------|-----|
| 対象期間 | 2025-12-29 〜 2026-01-03 |
| AM/PM 境界時刻 | 13:00 |
| PM締め終了時刻 | 16:30 |
| 備考 | 年末年始短縮診療 |

### 1.3 休診日設定

祝日判定は行わない。代わりに設定画面（Section 5.3）で「休診日」を管理する。

| 設定種別 | 内容 | 例 |
|---------|------|-----|
| 週次休診曜日 | 毎週特定曜日を休診に設定 | 毎週日曜（is_closed = true） |
| 個別休診日 | 特定日を個別に休診登録 | 2026-01-01 |

休診日はレジ締め画面で「この日は休診日です」と表示し、締め操作をスキップできる。

---

## 2. 部門カテゴリ定義

**確定済み（Q1a: RV = rabies vaccine / 狂犬病ワクチン）**

| 要求部門 | DB カテゴリ値 | 備考 |
|---------|-----------|------|
| 診療 | `examination`, `procedure`, `test`, `medicine` | 既存 enum 流用・統合 |
| 外科 | `surgery` | 既存 enum 流用 |
| **RV** | **`vaccine`** | 狂犬病ワクチン（rabies vaccine）。**enum 追加必要** |
| フード | `food` | 既存 enum 流用 |
| トリミング | `trimming` | **enum 追加必要** |
| ホテル | `hotel` | **enum 追加必要** |
| 用品 | `goods` | 既存 enum 流用 |
| トレセン | `training` | **enum 追加必要** |

集計画面での表示名（label）は別途マスタまたは定数で管理する。  
`vaccine` → "RV"、`examination` 等 → "診療" のようにグルーピングして表示する。

---

## 3. レジ締め機能

### 3.1 締めの単位

**1回の締め = 1診療日の AM または PM 区分**。AM・PM を別々に実行（PM締め後が「日締め」相当）。

### 3.2 集計対象レコード

| 対象 | 条件 |
|------|------|
| 通常会計 | `billings.status = 'paid'` かつ `paid_at` が締め時間帯内 |
| 退院会計 | `billings.hospitalization_id IS NOT NULL` も同様に含める |
| 返金 | `billing_refunds` の返金額を差し引いた**実収額**で集計 |

### 3.3 集計マトリクス（画面表示・印刷）

支払方法はマスタ化後に動的列として表示する（現金・クレジットカード・電子マネー 等）。

| 部門 ↓ / 支払方法 → | 現金 | クレジットカード | 電子マネー | 合計 |
|--------------------|------|----------------|----------|------|
| 診療 | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| 外科 | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| RV | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| フード | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| トリミング | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| ホテル | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| 用品 | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| トレセン | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx | ¥xx,xxx |
| **合計** | **¥xx,xxx** | **¥xx,xxx** | **¥xx,xxx** | **¥xx,xxx** |

### 3.4 レジ金突合

| 項目 | 説明 |
|-----|------|
| 理論値（現金） | 集計マトリクスの現金合計（自動計算） |
| 実際のレジ金 | 担当者が手入力 |
| 差額 | 実際のレジ金 − 理論値（± 表示） |

突合結果は **DB に永続保存**（監査・不正検知のため）。

### 3.5 個別会計一覧（締め対象）

| 列 | 内容 |
|----|------|
| 時刻 | `paid_at`（HH:MM） |
| 飼主名 / ペット名 | |
| 種別 | 通常 / 退院 |
| 部門 | 主要部門 |
| 支払方法 | 現金 / クレジットカード / 電子マネー 等 |
| 請求金額 | 税込 |
| 返金額 | 返金があれば表示 |
| 実収額 | 請求金額 − 返金額 |

### 3.6 締め実行フロー

```
1. 対象日・区分（AM/PM）を選択
2. 集計マトリクスをプレビュー
3. 実際のレジ金を入力（現金のみ）
4. 差額を確認
5. メモ入力（任意）
6. 「締める」ボタン → 確認モーダル
7. POST → cash_register_closes テーブルに保存
8. 締め完了 → 印刷（A4縦）
```

締め済み区分の `billings` / `payments` を後から編集した場合は**警告を表示**（ハードブロックなし）。

---

## 4. 月間売上集計（経理向け）

### 4.1 アクセス制御

権限は `/settings/permission-groups`（権限グループ設定画面）で各グループごとに設定する。

| リソース | 定数 | 対象機能 |
|---------|------|---------|
| `cash-register-close` | `ResourceCashRegisterClose` | レジ締め実行・締め履歴閲覧 |
| `accounting-reports` | `ResourceAccountingReports` | 月次売上集計・CSV エクスポート |
| `closing-settings` | `ResourceClosingSettings` | 締め時間・特別期間・休診日の設定変更 |

各リソースに対して view / create / edit / delete を権限グループごとに付与できる。

### 4.2 月次サマリ

| 項目 | 説明 |
|-----|------|
| 診療日数 | 会計実績が1件以上ある日の数 |
| 総会計件数 | 月間の会計済み件数 |
| 総売上（税込） | 月間の合計請求金額 |
| 総返金額 | 月間の返金合計 |
| 実収金額 | 総売上 − 総返金額 |
| 支払方法別合計 | payment_methods マスタに基づく動的集計 |
| 部門別合計 | 8部門それぞれの月間合計 |

### 4.3 日別明細テーブル

| 列 | 内容 |
|----|------|
| 日付 | YYYY-MM-DD（曜日） |
| AM 件数 / AM 実収 | AM締め実績 |
| PM 件数 / PM 実収 | PM締め実績 |
| 日計実収 | AM + PM |
| 返金 | 当日の返金合計 |
| 締め状態 | AM締め済 / PM締め済 / 未締め / 休診日 |

### 4.4 CSV エクスポート

```
日付, 曜日区分, AM件数, AM合計, AM現金, AMクレジット, AM電子マネー, AM返金, AM実収,
PM件数, PM合計, PM現金, PMクレジット, PM電子マネー, PM返金, PM実収,
日計件数, 日計合計, 日計返金, 日計実収
```

---

## 5. 締め時間設定画面（`/settings/closing-time`）

- **アクセス権限**: `closing-settings` リソースの edit 権限を持つグループ
- **構成**: タブなし、1ページで完結する設定画面（3セクション）

```
集計・締め時間設定
────────────────────────────────────────
[セクション1] 標準締め時間
[セクション2] 特別期間（複数登録可）
[セクション3] 休診日設定
────────────────────────────────────────
```

### 5.1 標準締め時間

| フィールド | 項目ID | 入力部品 | 必須 | デフォルト | 説明 |
|-----------|--------|---------|------|-----------|------|
| AM / PM 境界時刻 | `am_pm_boundary` | `TimePicker` | ✅ | `14:00` | この時刻を境に AM締め / PM締めに分割する |
| 平日・土曜の診療終了時刻 | `closing_weekday_end` | `TimePicker` | ✅ | `18:30` | PM締めの終了時刻（平日・土） |
| 日曜の診療終了時刻 | `closing_sunday_end` | `TimePicker` | ✅ | `17:30` | PM締めの終了時刻（日曜）。祝日は平日設定を使用 |

設定値を変更すると、適用される集計区分プレビューがリアルタイムで更新される。

```
標準締め時間
┌────────────────────────────────────┐
│  AM / PM 境界時刻      [14:00]     │
│  平日・土曜 診療終了    [18:30]     │
│  日曜 診療終了          [17:30]     │
└────────────────────────────────────┘

この設定が適用される集計区分:
  平日・土曜 │ AM締め: 診療開始 〜 13:59
             │ PM締め: 14:00   〜 18:29
  日曜       │ AM締め: 診療開始 〜 13:59
             │ PM締め: 14:00   〜 17:29
```

- **保存**: 「保存」ボタンで一括 PATCH
- **UI**: `PropertyRow` + `PropInput` / `TimePicker` を使用したNotionスタイル（ボーダーレス）

### 5.2 特別期間設定（CRUD）

年末年始・お盆等の時間短縮診療に対応。標準設定より**優先して適用**される。

#### フォーム項目

| フィールド | 項目ID | 入力部品 | 必須 | 説明 |
|-----------|--------|---------|------|------|
| 開始日 | `start_date` | `DatePicker` | ✅ | |
| 終了日 | `end_date` | `DatePicker` | ✅ | 開始日以降のみ選択可 |
| AM/PM 境界時刻 | `am_pm_boundary` | `TimePicker` | ✅ | |
| PM締め終了時刻 | `pm_end` | `TimePicker` | ✅ | |
| 備考 | `note` | `Input` | | 最大100文字 |

#### バリデーション

| ルール | エラーメッセージ |
|--------|-----------------|
| 終了日 ≥ 開始日 | 「終了日は開始日以降に設定してください」 |
| 既存期間と日付が重複しない | 「期間が他の特別期間と重複しています」 |
| AM/PM境界時刻 < PM締め終了時刻 | 「PM締め終了時刻は境界時刻より後に設定してください」 |

#### 画面イメージ

```
特別期間                              [+ 追加]
┌──────────────────────────────────────────────┐
│ 2025-12-29 〜 2026-01-03                      │
│ AM/PM境界: 13:00  PM終了: 16:30              │
│ 年末年始短縮診療                    [編集][削除] │
├──────────────────────────────────────────────┤
│ 2025-08-13 〜 2025-08-15                      │
│ AM/PM境界: 13:00  PM終了: 16:00              │
│ お盆短縮診療                        [編集][削除] │
└──────────────────────────────────────────────┘
```

- 追加・削除は即時 POST / DELETE（「保存」ボタン不要）
- 削除時は確認ダイアログを表示
- 保存後、翌日以降の集計・締め画面に即座に反映

### 5.3 休診日設定

祝日判定は行わない。休診日を明示的に設定する。

#### 5.3.1 週次休診曜日

チェックボックスで週次の休診曜日を設定する。

```
休診曜日
┌────────────────────────────────────────┐
│  □ 月  □ 火  □ 水  □ 木  □ 金  □ 土  ☑ 日 │
└────────────────────────────────────────┘
```

- 複数曜日選択可
- `clinic_settings` の `closed_weekdays` カラムで管理

#### 5.3.2 個別休診日（臨時休診）

| フィールド | 入力部品 | 説明 |
|-----------|---------|------|
| 休診日 | `DatePicker` | 個別の休診日を追加 |
| 備考 | `Input` | 最大50文字（任意） |

```
個別休診日                            [+ 追加]
┌──────────────────────────────────────────────┐
│ 2026-01-01  元日                    [削除]    │
│ 2026-04-30  院長不在               [削除]    │
└──────────────────────────────────────────────┘
```

- 追加・削除は即時 POST / DELETE

---

## 6. 画面構成

### 6.1 レジ締め画面（`/accounting/close`）

```
対象日: [2025-04-20]  区分: [AM] [PM]  [集計プレビュー]
──────────────────────────────────────────────────────
部門別 × 支払方法別 集計マトリクス
  診療     | ¥xx,xxx | ¥xx,xxx | ... | ¥xx,xxx
  外科     | ...
  合計     | ¥xx,xxx | ¥xx,xxx | ... | ¥xx,xxx
──────────────────────────────────────────────────────
レジ金突合
  理論値（現金）: ¥ xx,xxx
  実際のレジ金:  [      ] 円  ← 手入力
  差額:         ± ¥ xxx
──────────────────────────────────────────────────────
個別会計一覧（退院含む）
  時刻 | 飼主 | ペット | 種別 | 部門 | 支払 | 実収
  ...
[印刷プレビュー]  [締める]
```

### 6.2 締め履歴画面（`/accounting/close/history`）

過去の締めレコードを一覧表示。差額が出た履歴はハイライト。

### 6.3 集計レポート画面（`/accounting/reports`）

```
[月次レポート] [締め履歴]
  期間選択: [2025年] [4月] [表示]
  月次サマリカード（総売上・実収・件数・部門別）
  [CSVエクスポート]
  日別明細テーブル（締め状態列あり）
```

---

## 7. API 設計

| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| GET | `/api/v1/cash-register/preview?date=YYYY-MM-DD&period=am` | 締め前プレビュー |
| POST | `/api/v1/cash-register/closes` | 締め実行・保存 |
| GET | `/api/v1/cash-register/closes` | 締め履歴一覧 |
| GET | `/api/v1/cash-register/closes/:id` | 締め詳細 |
| GET | `/api/v1/reports/monthly?year=2025&month=4` | 月次集計取得 |
| GET | `/api/v1/reports/monthly/csv?year=2025&month=4` | 月次 CSV エクスポート |
| GET | `/api/v1/closing-settings` | 締め時間設定取得 |
| PATCH | `/api/v1/closing-settings` | 標準締め時間更新 |
| GET | `/api/v1/closing-settings/special-periods` | 特別期間一覧 |
| POST | `/api/v1/closing-settings/special-periods` | 特別期間追加 |
| PATCH | `/api/v1/closing-settings/special-periods/:id` | 特別期間更新 |
| DELETE | `/api/v1/closing-settings/special-periods/:id` | 特別期間削除 |
| GET | `/api/v1/closing-settings/holidays` | 個別休診日一覧 |
| POST | `/api/v1/closing-settings/holidays` | 個別休診日追加 |
| DELETE | `/api/v1/closing-settings/holidays/:id` | 個別休診日削除 |
| GET | `/api/v1/payment-methods` | 支払方法マスタ一覧 |
| POST | `/api/v1/payment-methods` | 支払方法追加 |
| PATCH | `/api/v1/payment-methods/:id` | 支払方法更新 |
| DELETE | `/api/v1/payment-methods/:id` | 支払方法削除 |

---

## 8. DB 設計

### 8.1 item_category enum 追加

```sql
ALTER TYPE item_category ADD VALUE 'vaccine';    -- RV（狂犬病ワクチン）
ALTER TYPE item_category ADD VALUE 'trimming';
ALTER TYPE item_category ADD VALUE 'hotel';
ALTER TYPE item_category ADD VALUE 'training';
```

### 8.2 締め時間設定（clinic_settings へのカラム追加）

```sql
ALTER TABLE clinic_settings
  ADD COLUMN closing_am_pm_boundary TIME NOT NULL DEFAULT '14:00',
  ADD COLUMN closing_weekday_end    TIME NOT NULL DEFAULT '18:30',
  ADD COLUMN closing_sunday_end     TIME NOT NULL DEFAULT '17:30',
  ADD COLUMN closed_weekdays        INTEGER[] NOT NULL DEFAULT '{}';
  -- 例: {0} = 日曜のみ休診, {0,6} = 日曜・土曜休診（0=日, 1=月...6=土）
```

### 8.3 特別期間テーブル

```sql
CREATE TABLE closing_special_periods (
  id             BIGSERIAL PRIMARY KEY,
  clinic_id      BIGINT NOT NULL REFERENCES clinics(id),
  start_date     DATE   NOT NULL,
  end_date       DATE   NOT NULL,
  am_pm_boundary TIME   NOT NULL,
  pm_end         TIME   NOT NULL,
  note           VARCHAR(100),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_date_range CHECK (start_date <= end_date)
);
CREATE INDEX idx_closing_special_periods_clinic
  ON closing_special_periods(clinic_id, start_date, end_date);
```

### 8.4 個別休診日テーブル

```sql
CREATE TABLE clinic_holidays (
  id         BIGSERIAL PRIMARY KEY,
  clinic_id  BIGINT NOT NULL REFERENCES clinics(id),
  date       DATE   NOT NULL,
  note       VARCHAR(50),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (clinic_id, date)
);
CREATE INDEX idx_clinic_holidays_clinic
  ON clinic_holidays(clinic_id, date);
```

### 8.5 レジ締めテーブル

```sql
CREATE TABLE cash_register_closes (
  id                      BIGSERIAL PRIMARY KEY,
  clinic_id               BIGINT NOT NULL REFERENCES clinics(id),
  close_date              DATE   NOT NULL,
  period                  VARCHAR(2) NOT NULL CHECK (period IN ('am', 'pm')),
  theoretical_cash        BIGINT NOT NULL DEFAULT 0,
  actual_cash             BIGINT NOT NULL DEFAULT 0,
  cash_difference         BIGINT NOT NULL DEFAULT 0,
  category_breakdown      JSONB  NOT NULL DEFAULT '{}',
  -- {"vaccine": {"cash": 1000, "credit_card": 2000, ...}, ...}
  memo                    TEXT DEFAULT '',
  closed_by               BIGINT REFERENCES staffs(id),
  closed_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (clinic_id, close_date, period)
);
CREATE INDEX idx_cash_register_closes_clinic
  ON cash_register_closes(clinic_id, close_date DESC);
```

### 8.6 支払方法マスタテーブル（payment_methods）

`payments.method` の enum 固定をやめ、マスタテーブルで管理する。

```sql
CREATE TABLE payment_methods (
  id             BIGSERIAL PRIMARY KEY,
  clinic_id      BIGINT NOT NULL REFERENCES clinics(id),
  name           VARCHAR(50) NOT NULL,        -- 例: "現金", "クレジットカード", "電子マネー"
  display_order  INTEGER NOT NULL DEFAULT 0,
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at     TIMESTAMPTZ,
  UNIQUE (clinic_id, name) WHERE deleted_at IS NULL
);
CREATE INDEX idx_payment_methods_clinic
  ON payment_methods(clinic_id, display_order) WHERE deleted_at IS NULL;
```

#### シードデータ（初期データ）

```sql
-- 既存の enum 値 (cash / credit_card / electronic_money) に対応
INSERT INTO payment_methods (clinic_id, name, display_order)
SELECT id, '現金', 1 FROM clinics;

INSERT INTO payment_methods (clinic_id, name, display_order)
SELECT id, 'クレジットカード', 2 FROM clinics;

INSERT INTO payment_methods (clinic_id, name, display_order)
SELECT id, '電子マネー', 3 FROM clinics;
```

#### payments テーブルの変更

```sql
-- payments.method (VARCHAR) を payment_method_id (FK) に移行
ALTER TABLE payments
  ADD COLUMN payment_method_id BIGINT REFERENCES payment_methods(id);

-- データ移行後に method カラムを削除
-- UPDATE payments SET payment_method_id = ... WHERE method = 'cash' ...
-- ALTER TABLE payments DROP COLUMN method;
```

---

## 9. 調査済みの実装状況

### DB
- `payments.method`: `cash` / `credit_card` / `electronic_money`（enum 固定。マスタ化対象）
- `billing_items.category`: 8種類（`examination`/`test`/`procedure`/`surgery`/`medicine`/`food`/`goods`/`other`）。`vaccine`/`trimming`/`hotel`/`training` の追加が必要
- `billings.hospitalization_id`: 退院紐付け可能
- レジ締めテーブル・締め時間設定カラム・休診日テーブル・支払方法マスタ: **なし**

### Backend
- `backend/internal/model/accounting.go`: Billing / BillingItem / Payment モデル
- `backend/internal/handler/accounting_handler.go`: 個別会計 CRUD のみ
- 集計エンドポイント: **0件**

### Frontend
- `frontend/src/features/accounting/routes/AccountingList.tsx`: 個別会計リスト
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`: 個別詳細
- 締め画面・集計レポート・設定画面: **なし**

---

## 10. リスク・懸念事項

| リスク | 影響 | 対策 |
|--------|------|------|
| `item_category` enum 追加で既存データ `other` が未整理になる | 中 | マイグレーション時に既存データを新カテゴリへ再分類。運用者の合意必要 |
| 締め後のデータ変更による整合性崩れ | 高 | `category_breakdown` JSONB にスナップショット保存。編集時に警告表示 |
| JST タイムゾーン境界（跨日・深夜会計） | 中 | `paid_at` は `TIMESTAMPTZ`。JST 変換を明示して範囲クエリ |
| `payments.method` → `payment_method_id` 移行コスト | 高 | 既存の会計データすべてに影響。段階的移行（旧カラム維持→移行→削除）で対応 |
| 締め時刻以降に発生した支払いの扱い | 高 | 翌区分に繰越？警告のみ？→ 要確認 |

---

## 11. 未確認事項

なし。全質問確定済み。実装着手可能。  
→ `docs/tasks/open/accounting/` に移動してイシュー分解・実装着手。

---

## 12. 確定済み事項

| 質問 | 回答 |
|------|------|
| Q1a: **RV** の意味 | **rabies vaccine（狂犬病ワクチン）**。DB カテゴリ: `vaccine` |
| Q2: 電子マネーを集計・突合に含めるか | **Yes**。支払方法は `payment_methods` テーブルでマスタ化して動的管理 |
| Q3: 祝日判定が必要か | **不要**（祝日も通常営業）。代わりに休診日を設定画面（5.3）で管理する |
| カテゴリ（トリミング・ホテル・トレセン） | `trimming` / `hotel` / `training` enum 追加 |
| 締め単位 | AM / PM の2区分 |
| レジ金突合の永続化 | DB 保存（`cash_register_closes`） |
| 拠点単位 | クリニック単位（`clinic_id`） |
| 印刷サイズ | A4 縦 |
| 担当者記録 | `closed_by` (staff_id) |
| 締め後ロック | 警告のみ（ハードブロックなし） |
| 締め時間の変更 | 設定画面（`/settings/closing-time`）から可能 |
| アクセス制御 | RBAC（`ResourceCashRegisterClose` / `ResourceAccountingReports` / `ResourceClosingSettings`） |
| 曜日区分 | 平日・土曜（18:30終了）/ 日曜（17:30終了）。祝日は平日扱い |
| Q10: 消費税区分（8%/10%）の按分表示 | **必要**。`billing_items.tax_rate` で管理済み。実装漏れとして BUG-404 に起票 |
