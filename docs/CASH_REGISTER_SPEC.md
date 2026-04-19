# レジ締め・集計業務 仕様書

> **ステータス**: Draft — 実装前確認用  
> **対象機能**: BUG-368（日次集計）拡張 + 新規 月次集計 + 設定ページ  
> **最終更新**: 2026-04-19

---

## 1. 用語定義

| 用語 | 一般名称 | 意味 |
|------|---------|------|
| AM締め | 午前シフト締め（Shift Closing - AM） | 午前診療分の売上・現金照合を確定する操作 |
| PM締め | 午後シフト締め（Shift Closing - PM） | 午後診療分の売上・現金照合を確定する操作 |
| 中間計 | シフト区分集計 / 区分別小計 | AM・PM 各シフトの売上合計（日計の途中集計） |
| 日計 | 日次売上集計 | 1日（AM＋PM）の売上合計 |
| 月計 | 月次売上集計 | 1ヶ月間の売上合計 |
| 過不足 | 現金差異 | 「システム上の現金売上」と「実際の手持ち現金」の差額 |
| 締め確定 | 締め実行 / ロック | 集計を確定し以後の変更を不可にする操作 |

---

## 2. 設定ページ（営業時間・シフト設定）

### 2-1. 設置場所

**`/hospital-settings`（病院設定）** に「レジ締め設定」タブを追加する。

```
/hospital-settings
  ├── クリニック情報（既存）
  └── レジ締め設定（新規追加）  ← ここに営業時間・シフト設定を配置
```

### 2-2. 設定項目

設定ページで管理する項目は以下の通り。ここで登録した値が締め処理・集計の全ての基準となる。

#### (A) 曜日区分ごとのデフォルト設定

| 設定項目 | 内容 | 例 |
|---------|------|-----|
| 曜日区分 | `通常営業日`（月〜土）または `日曜` | — |
| AM/PM 切替時刻 | 午前を締めて午後に切り替える時刻 | `14:00` |
| 閉院時刻 | PM締め終了・その日の診療終了時刻 | `18:30`（通常）/ `17:30`（日曜） |

**設定例（初期値）:**

| 曜日区分 | AM/PM切替時刻 | 閉院時刻 |
|---------|-------------|---------|
| 通常営業日（月〜土） | 14:00 | 18:30 |
| 日曜 | 14:00 | 17:30 |

#### (B) 臨時設定（特定日付オーバーライド）

年末年始・院長不在日・短縮営業日など、通常の曜日設定とは異なる日を個別登録する。

| 設定項目 | 内容 |
|---------|------|
| 対象日付 | 特定の日付（YYYY-MM-DD） |
| AM/PM切替時刻 | その日限りの切替時刻 |
| 閉院時刻 | その日限りの閉院時刻 |
| 備考 | 理由メモ（例: 年末短縮、院長休暇等） |

**適用優先順位**: `臨時設定（特定日付）> 日曜設定 > 通常設定`

### 2-3. 設定データの利用フロー

```
[設定ページ]
  ↓ 執務担当者が営業時間を登録・変更
  ↓ POST/PATCH /v1/register-shift-settings

[DB: register_shift_settings]
  ↓ レジ締め処理・集計のたびに参照

[締め実行時]
  1. 当日の設定を取得（臨時 → 日曜 → 通常の優先順で解決）
  2. AM期間 = 前日の閉院時刻 〜 当日の切替時刻-1秒
  3. PM期間 = 当日の切替時刻 〜 当日の閉院時刻-1秒
  4. 各期間内の completed_at を持つ billings を集計

[日次・月次集計画面]
  設定値を参照してグラフ・表の時間軸を決定
```

---

## 3. 権限設定

権限グループ設定ページ（`/master/permission-groups`）から以下のリソースを付与する。

| 操作 | 権限リソース | アクション | 付与対象の想定 |
|------|------------|----------|-------------|
| 営業時間・シフト設定の変更 | `register_settings`（新規） | `edit` | 執務担当者 |
| AM/PM 締め実行（確定） | `register_closing`（新規） | `create` | 執務担当者 |
| 日次集計の閲覧 | `accounting` | `view` | 受付・執務 |
| 月次集計の閲覧 | `accounting` | `view` | 経理・執務 |

---

## 4. 締め処理フロー（確定操作）

### 4-1. AM締めフロー

```
[14:00 頃 — 設定値に基づく切替時刻]
執務担当者がレジ締め画面を開く
    ↓
「AM締め実行」ボタンを押す
    ↓
システムが当日の設定を取得し AM 期間を確定
  → 期間: 前日の閉院時刻(設定値) 〜 当日の切替時刻-1秒(設定値)
    ↓
AM 期間の集計を表示
  - 現金売上金額（自動計算）
  - カード・電子マネー売上（自動計算）
  - 件数
    ↓
担当者が「実際の手持ち現金」を入力
    ↓
過不足 = 手持ち現金 − システム上の現金売上（自動表示）
    ↓
備考欄（任意）を入力
    ↓
「確定」ボタン → 確認モーダル
    ↓
締め確定 → 該当期間をロック（以後変更不可）
締め確定記録を register_closes テーブルに保存
```

### 4-2. PM締めフロー

AM締めと同様。PM 期間 = `切替時刻(設定値) 〜 閉院時刻-1秒(設定値)`。

### 4-3. 締め後の再オープン

- 締め確定後は原則変更不可
- 特別事由がある場合：`register_settings` の `admin` 権限で再オープン可能
- 再オープンは監査ログに記録する

### 4-4. 未実行締めの扱い

- 当日 AM 締めが未実行のまま翌日になった場合 → 警告を表示し、前日分として遡及締めを実行できる
- 遡及締めは「遡及」フラグ付きで保存し、月次集計では通常分と区別できる

---

## 5. 日次集計項目（レジ締め画面）

### 5-1. AM/PM 各シフト表示

| 項目 | 内容 |
|------|------|
| 集計期間 | 設定値から算出した開始〜終了時刻 |
| 会計件数 | 支払完了した会計の件数 |
| 現金売上 | 支払方法=現金 の billing_amount 合計 |
| カード売上 | 支払方法=クレジットカード の合計 |
| 電子マネー売上 | 支払方法=電子マネー の合計 |
| **売上合計** | 全支払方法の合計 |
| 手持ち現金（入力） | 担当者が実際に数えた現金金額 |
| **過不足** | 手持ち現金 − 現金売上（自動計算） |
| 確定者 | 締め実行した担当者名 |
| 確定日時 | 締め確定のタイムスタンプ |

### 5-2. 日計サマリー（AM＋PM合計）

| 項目 | 内容 |
|------|------|
| 件数合計 | AM件数 + PM件数 |
| 現金売上合計 | |
| カード売上合計 | |
| 電子マネー売上合計 | |
| **日計売上合計** | 全支払方法・全シフト合計 |
| 診療区分別内訳 | examination / test / procedure / surgery / medicine / food / goods / other |
| 返金件数・金額 | 当日発生した返金の合計 |
| 純売上 | 売上合計 − 返金合計 |

---

## 6. 月次集計項目（経理向け）

受付は不要。経理権限を持つスタッフのみ閲覧可能。

### 6-1. サマリー指標

| 項目 | 内容 |
|------|------|
| 月間売上合計 | 支払完了会計の billing_amount 合計 |
| 前月比 | 金額・増減率 |
| 前年同月比 | 金額・増減率 |
| 月間件数 | 会計完了件数 |
| 1件あたり平均 | 売上合計 ÷ 件数 |
| 月間純売上 | 売上 − 返金合計 |
| 未収金繰越 | 月末時点の status=waiting 合計 |

### 6-2. 支払方法別集計

| 支払方法 | 件数 | 金額 | 構成比 |
|---------|------|------|-------|
| 現金 | | | |
| クレジットカード | | | |
| 電子マネー | | | |

### 6-3. 診療区分別集計

| 区分 | 件数 | 金額 | 構成比 |
|------|------|------|-------|
| 診察 | | | |
| 検査 | | | |
| 処置 | | | |
| 手術 | | | |
| 薬品 | | | |
| フード | | | |
| グッズ | | | |
| その他 | | | |

### 6-4. 日別売上推移

- 日付 × 売上金額の一覧表（グラフ表示も可）
- AM/PM 別内訳も表示

### 6-5. 担当医別集計

| 担当医 | 件数 | 売上合計 | 平均単価 |
|--------|------|---------|---------|

### 6-6. 新規/再診別集計

| 区分 | 件数 | 売上合計 | 構成比 |
|------|------|---------|-------|
| 初診 | | | |
| 再診 | | | |

### 6-7. 返金・キャンセル集計

| 項目 | 件数 | 金額 |
|------|------|------|
| 返金 | | |
| キャンセル（status=cancelled） | | |

### 6-8. 過不足累計（経理チェック用）

- 月間の AM/PM 各締めの過不足額一覧
- 月間過不足合計

---

## 7. データ設計

### 7-1. 新規テーブル

```sql
-- シフト時間帯設定（設定ページで管理）
CREATE TABLE register_shift_settings (
  id           BIGSERIAL PRIMARY KEY,
  clinic_id    BIGINT NOT NULL REFERENCES clinics(id),
  -- 曜日区分: 'weekday_sat'（月〜土）| 'sunday'（日曜） | 'specific'（特定日）
  day_type     VARCHAR(20) NOT NULL,
  -- specific の場合のみ使用。他は NULL
  target_date  DATE,
  -- AM/PM切替時刻（例: 14:00:00）
  am_end_time  TIME NOT NULL,
  -- PM終了=閉院時刻（例: 18:30:00）
  pm_end_time  TIME NOT NULL,
  memo         TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at   TIMESTAMPTZ,
  CONSTRAINT uk_register_shift_settings
    UNIQUE (clinic_id, day_type, target_date)
);

-- 締め確定記録
CREATE TABLE register_closes (
  id                BIGSERIAL PRIMARY KEY,
  clinic_id         BIGINT NOT NULL REFERENCES clinics(id),
  close_date        DATE NOT NULL,
  shift             VARCHAR(10) NOT NULL,  -- 'am' | 'pm'
  -- 集計期間（設定値から算出して保存）
  period_start      TIMESTAMPTZ NOT NULL,
  period_end        TIMESTAMPTZ NOT NULL,
  -- 集計結果
  billing_count     INT NOT NULL DEFAULT 0,
  cash_amount       BIGINT NOT NULL DEFAULT 0,
  card_amount       BIGINT NOT NULL DEFAULT 0,
  electronic_amount BIGINT NOT NULL DEFAULT 0,
  total_amount      BIGINT NOT NULL DEFAULT 0,
  -- 担当者入力
  counted_cash      BIGINT,               -- 実際の手持ち現金
  cash_diff         BIGINT,               -- 過不足（counted_cash - cash_amount）
  memo              TEXT,
  is_retroactive    BOOLEAN NOT NULL DEFAULT FALSE,
  -- 確定者
  closed_by         BIGINT NOT NULL REFERENCES staffs(id),
  closed_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- 再オープン
  reopened_by       BIGINT REFERENCES staffs(id),
  reopened_at       TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 1日1クリニック1シフトのみ確定可能
CREATE UNIQUE INDEX uk_register_closes_shift
  ON register_closes(clinic_id, close_date, shift)
  WHERE reopened_at IS NULL;
```

### 7-2. APIエンドポイント設計

```
# 設定ページ用
GET    /v1/register-shift-settings          設定一覧取得
PUT    /v1/register-shift-settings          設定一括更新（曜日区分ごと）
POST   /v1/register-shift-settings/specific 臨時設定追加
DELETE /v1/register-shift-settings/specific/:id  臨時設定削除

# 締め処理用
GET    /v1/accountings/daily-summary        日次集計（設定値を参照して期間算出）
POST   /v1/register-closes                  締め確定
GET    /v1/register-closes?date=YYYY-MM     月の締め状況一覧

# 月次集計用
GET    /v1/accountings/monthly-summary?year=YYYY&month=MM
```

### 7-3. 既存テーブルへのインデックス追加

```sql
CREATE INDEX idx_billings_clinic_completed
  ON billings(clinic_id, completed_at)
  WHERE deleted_at IS NULL AND status = 'completed';
```

---

## 8. 画面構成

```
/hospital-settings
  └── [レジ締め設定タブ]（新規）
        曜日区分ごとの AM/PM切替時刻・閉院時刻
        臨時設定（特定日付の追加・削除）

/accounting/daily-summary（BUG-368 拡張）
  └── 設定値を参照して AM/PM 期間を算出・表示
        AM 集計表（未確定 → 確定ボタン）
        PM 集計表（未確定 → 確定ボタン）
        日計サマリー

/accounting/monthly-summary（新規・経理向け）
  └── 月選択
        6-1〜6-8 の集計表
        CSVエクスポート
```

---

## 9. 実装優先順位

| フェーズ | 内容 | 優先度 |
|---------|------|-------|
| Phase 1 | `register_shift_settings` テーブル + API | HIGH |
| Phase 1 | `/hospital-settings` にレジ締め設定タブ追加 | HIGH |
| Phase 1 | 日次集計（BUG-368）を設定値参照に切り替え + AM/PM分割 | HIGH |
| Phase 2 | 締め確定操作（`register_closes` 記録・ロック） | HIGH |
| Phase 2 | 権限リソース追加（`register_settings` / `register_closing`） | HIGH |
| Phase 3 | 月次集計画面 `/accounting/monthly-summary` | MEDIUM |
| Phase 3 | 過不足入力・差異管理 | MEDIUM |
| Phase 4 | 月次 CSV エクスポート | LOW |
| Phase 4 | 前月比・前年同月比グラフ | LOW |

---

## 10. 未決事項

| # | 質問 | 回答・現在の仮定 |
|---|------|---------------|
| 1 | 祝日カレンダーはシステム自動判定か手動か？ | **解決**: 祝日も通常営業のため不要。短縮は臨時設定で手動対応。 |
| 2 | 月次CSVのフォーマット（会計ソフト連携の有無）？ | 汎用CSV想定 |
| 3 | 複数クリニックの場合、締めは院ごと独立か？ | 独立（clinic_id 単位） |
| 4 | 担当医別集計は `medical_records.doctor_id` を参照するか？ | Yes |

---

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-19 | 初版作成（BUG-368 仕様整理・拡張設計） |
| 2026-04-19 | 祝日も通常営業のため「日祝」区分廃止。短縮営業は臨時設定で対応。 |
| 2026-04-19 | 設定ページ（/hospital-settings）で営業時間を管理し、全締め処理がその設定値を参照する設計に変更。 |
