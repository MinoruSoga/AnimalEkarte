# レジ締め・集計業務 仕様書

> **ステータス**: Draft — 実装前確認用  
> **対象機能**: BUG-368（日次集計）拡張 + 新規 月次集計  
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

## 2. シフト時間帯設定

### 2-1. デフォルト設定

| 曜日区分 | AM締め（期間） | 切替時刻 | PM締め（期間） | 閉院時刻 |
|---------|--------------|---------|--------------|---------|
| 平日・土曜 | 前日PM終了〜13:59 | 14:00 | 14:00〜18:29 | 18:30 |
| 日曜・祝日 | 前日PM終了〜13:59 | 14:00 | 14:00〜17:29 | 17:30 |

> **AM締め期間の補足**: AM締めは「前日のPM締め閉院時刻」から「当日13:59」をカバーする。  
> これは当日の早朝会計（前日の外泊入院精算等）も含むため、日付をまたいで集計する。

### 2-2. 曜日区分の定義

```
平日    : 月〜金（祝日を除く）
土曜    : 土曜日（祝日を除く）
日祝    : 日曜日 + 祝日
年末年始: 12/30〜1/3（デフォルト。設定で変更可）
```

### 2-3. 時間帯変更の仕様

- 変更単位：**曜日区分ごと**に「AM/PM切替時刻」「閉院時刻」を設定
- 臨時変更：特定日付を指定して当日限りのオーバーライドが可能
- 適用優先順位：`特定日付設定 > 日祝設定 > 平日土設定`
- 変更可能権限：**執務権限**（後述「権限設定」参照）

---

## 3. 権限設定

### 3-1. 必要な権限

| 操作 | 権限リソース | アクション | 付与対象の想定 |
|------|------------|----------|-------------|
| シフト時間帯の設定・変更 | `register_settings`（新規） | `edit` | 執務担当者 |
| AM/PM 締め実行（確定） | `register_closing`（新規） | `create` | 執務担当者 |
| 日次集計の閲覧 | `accounting` | `view` | 受付・執務 |
| 月次集計の閲覧 | `accounting` | `view` | 経理・執務 |

### 3-2. 実装方針

権限ページ（`/master/permission-groups`）の `ResourceAccounting` とは**別リソース**として `register_settings` / `register_closing` を追加し、権限グループ設定から付与する。

---

## 4. 締め処理フロー（確定操作）

### 4-1. AM締めフロー

```
[14:00 頃]
執務担当者がレジ締め画面を開く
    ↓
「AM締め実行」ボタンを押す
    ↓
システムが AM 期間の集計を表示
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
締め確定 → 該当期間はロック（以後変更不可）
締め確定記録を DB に保存
```

### 4-2. PM締めフロー

AM締めと同様。「PM締め実行」で PM 期間を集計・確定。

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
| 期間 | 集計対象の開始〜終了時刻 |
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

| 項目 | 内容 | 備考 |
|------|------|------|
| 月間売上合計 | 支払完了会計の billing_amount 合計 | |
| 前月比 | 金額・増減率 | |
| 前年同月比 | 金額・増減率 | |
| 月間件数 | 会計完了件数 | |
| 1件あたり平均 | 売上合計 ÷ 件数 | |
| 月間純売上 | 売上 − 返金合計 | |
| 未収金繰越 | 月末時点の status=waiting 合計 | |

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
| 医師A | | | |
| 医師B | | | |

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

## 7. データ設計案

### 7-1. 新規テーブル

```sql
-- シフト時間帯設定
CREATE TABLE register_shift_settings (
  id           BIGSERIAL PRIMARY KEY,
  clinic_id    BIGINT NOT NULL REFERENCES clinics(id),
  day_type     VARCHAR(20) NOT NULL,  -- 'weekday_sat' | 'holiday' | 'special'
  target_date  DATE,                   -- special の場合のみ、特定日付
  am_end_time  TIME NOT NULL,          -- AM締め終了時刻（例: 13:59）
  pm_end_time  TIME NOT NULL,          -- PM締め終了時刻＝閉院時刻（例: 18:29）
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at   TIMESTAMPTZ
);

-- 締め確定記録
CREATE TABLE register_closes (
  id             BIGSERIAL PRIMARY KEY,
  clinic_id      BIGINT NOT NULL REFERENCES clinics(id),
  close_date     DATE NOT NULL,                    -- 締め対象日
  shift          VARCHAR(10) NOT NULL,             -- 'am' | 'pm'
  period_start   TIMESTAMPTZ NOT NULL,             -- 集計開始時刻
  period_end     TIMESTAMPTZ NOT NULL,             -- 集計終了時刻
  billing_count  INT NOT NULL DEFAULT 0,
  cash_amount    BIGINT NOT NULL DEFAULT 0,        -- 現金売上
  card_amount    BIGINT NOT NULL DEFAULT 0,        -- カード売上
  electronic_amount BIGINT NOT NULL DEFAULT 0,    -- 電子マネー売上
  total_amount   BIGINT NOT NULL DEFAULT 0,        -- 売上合計
  counted_cash   BIGINT,                           -- 実際の手持ち現金（入力値）
  cash_diff      BIGINT,                           -- 過不足（counted_cash - cash_amount）
  memo           TEXT,
  is_retroactive BOOLEAN NOT NULL DEFAULT FALSE,   -- 遡及締めフラグ
  closed_by      BIGINT NOT NULL REFERENCES staffs(id),
  closed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reopened_by    BIGINT REFERENCES staffs(id),
  reopened_at    TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 複合ユニーク: 1日1クリニック1シフト
CREATE UNIQUE INDEX uk_register_closes_shift
  ON register_closes(clinic_id, close_date, shift)
  WHERE reopened_at IS NULL;
```

### 7-2. 既存テーブルへの影響

`billings.completed_at` を締め期間フィルタに使用するため、インデックスを追加推奨:

```sql
CREATE INDEX idx_billings_clinic_completed
  ON billings(clinic_id, completed_at)
  WHERE deleted_at IS NULL AND status = 'completed';
```

---

## 8. 画面構成案

```
/accounting
  └── /accounting/daily-summary     ← BUG-368（既存・要拡張）
        AM締め実行ボタン
        PM締め実行ボタン
        AM/PM 各集計表 + 日計サマリー
        診療区分別内訳

/accounting/monthly-summary         ← 新規（経理向け）
        月選択（年月ピッカー）
        エクスポート（CSV）
        6-1〜6-8 の集計表
```

---

## 9. 実装優先順位

| フェーズ | 内容 | 優先度 |
|---------|------|-------|
| Phase 1 | シフト時間帯設定テーブル + 設定画面 | HIGH |
| Phase 1 | AM/PM 分割での日次集計表示（BUG-368 拡張） | HIGH |
| Phase 2 | 締め確定操作（register_closes 記録） | HIGH |
| Phase 2 | 権限リソース追加（register_settings / register_closing） | HIGH |
| Phase 3 | 月次集計画面（/accounting/monthly-summary） | MEDIUM |
| Phase 3 | 過不足入力・差異管理 | MEDIUM |
| Phase 4 | 月次 CSV エクスポート | LOW |
| Phase 4 | 前月比・前年同月比グラフ | LOW |

---

## 10. 未決事項（実装前に確認が必要）

| # | 質問 | 現在の仮定 |
|---|------|----------|
| 1 | 日祝カレンダーはシステム自動判定か手動入力か？ | 手動で祝日フラグを立てる方式を想定 |
| 2 | 月次CSVのフォーマット（会計ソフト連携の有無）？ | 汎用CSV想定 |
| 3 | 複数クリニックの場合、締めは院ごと独立か？ | 独立（clinic_id 単位） |
| 4 | 担当医別集計は、medical_records.doctor_id を参照するか？ | Yes |

---

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-04-19 | 初版作成（BUG-368 仕様整理・拡張設計） |
