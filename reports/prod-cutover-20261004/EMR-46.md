# EMR-46 — 締め時刻設定 現行値調査・本番投入案

> **Ticket**: Plane `EMR-46`（本番 締め時刻設定の投入値確認 = GitHub #252 / BRT-43 / GOLIVE_RUNBOOK §1 item 10）
> **調査日**: 2026-10-04 ／ **実施者**: delegated agent（worktree `AnimalEkarte-emr46` / branch `emr-46-cutover`）
> **範囲**: read-only。STG PlanetScale への SELECT のみ（RLS は session GUC `app.bypass_rls` / `app.current_clinic_ids` で読み取りスコープを制御）。PATCH・seed 書き換え・Plane 書き込みは未実施。

## 1. 設定の保管場所（schema / 書き込み経路）

- `clinic_settings`（`clinic_id` PK・1:1、`001_init.sql` L2057–2093）: `closing_am_start`（default 09:00、**画面・PATCH とも編集不可**）、`closing_am_pm_boundary`（default 14:00）、`closing_weekday_end`（default 18:30）、`closing_sunday_end`（default 17:30）、`closed_weekdays` `smallint[]`（**0=日 … 6=土**。BE: `closing_settings_service.go:426` が `int64(date.Weekday())` と比較、FE: `StandardClosingTimeSection.tsx` の checkbox `closed_weekday_0..6`）。
- `closing_special_periods`（L2124–2136）: 日付範囲の境界時刻上書き。STG 現行 **0 件**。
- `clinic_holidays`（L2043–2051）: 日付単位の休診日。STG 現行 **0 件**。
- 書き込み経路: 画面 `/settings/closing-time` → `PATCH /api/v1/closing-settings`（audit_logs 記録あり）。`closed_weekdays` は集計専用であり **LINE 予約の空き枠計算とは非連動**（`line_reservation_settings.closed_weekdays` は別モデル）。
- 集計区分: AM `[am_start, boundary)` / PM `[boundary, pm_end)` / EMG `[pm_end, 翌 am_start)`（越日、半開区間）。

## 2. STG 現行値（PlanetScale `animalekarte-stg`/`main`、2026-10-04 実測）

| clinic_id | 医院 | am_start | boundary | weekday_end | sunday_end | closed_weekdays | updated_at |
|---|---|---|---|---|---|---|---|
| 1 | 八王子病院 | 09:00 | **14:00** | **18:30** | **17:30** | **{0} = 日** | 2026-07-06 05:59:40 UTC（seed 投入時刻） |
| 2 | 城東センター病院 | 09:00 | **13:30** | **19:00** | **18:00** | **{4} = 木** | 同上（seed） |
| 3 | ノア動物病院 敷島 | 09:00 | **13:30** | **19:00** | **18:00** | **{4} = 木** | 同上（seed） |
| 4 | ノア動物病院 Hako bu neco | 09:00 | **13:30** | **19:00** | **18:00** | **{4} = 木** | 同上（seed） |

全院 `updated_at` が seed 投入時刻のまま = **UI/PATCH で一度も更新されていない**。`closing_special_periods`・`clinic_holidays` も全院 0 件。

### 現行値の出自（重要）

STG 値は実測値ではなく **デモ seed 由来のプレースホルダ**：

- clinic 1: demo seed `003_seed_demo.sql`（`76405f44d`、2026-05-24「expand demo seed data for STG readiness」）`(1,'14:00','18:30','17:30',ARRAY[0])` — ほぼ製品 default。
- clinics 2–4: 同コミットの demo INSERT `(2,'13:30','19:00','18:00',ARRAY[4]) -- 水曜定休`。**コメントは「水曜定休」だが `4` はアプリ規約では木曜**（Wed=3）— デモ値の曜日ミスラベルの可能性が高い。医院 3/4 の行は `74b9c3d69`（2026-09-04「skeleton rows for local reset」）で clinic 2 の行をそのまま複写したもの。旧デモ clinic 3 の値 `(14:00/18:00/17:00/{0,6} 土日定休)` は破棄済み。
- 旧システム側に締め時刻マスタは存在しない（`old_db` の `data-migration-field-map.json`・`manual-evidence-decisions.md` に該当マッピング無し）→ 実値は移行データからは復元不可。

## 3. 実値の証拠（移行予約データ・LINE 設定）

| 証拠 | 内容 | 示唆 |
|---|---|---|
| 予約曜日活動（STG `appointments`、JST・2015+） | clinic 1: 全曜日が全期間アクティブ。**日曜は 2020–2026 毎年 最多級**（300–522 件/年） | {0}=日休 は実運用と矛盾（日曜稼働） |
| 同上 | clinic 2: 2020–21 は木曜も稼働（275/272 件）。2022+ は全曜日とも疎 | {4}=木休 は少なくとも当時の運用と矛盾 |
| 同上 | clinic 3: 木曜に活動（2020:48, 2021:21）。clinic 4: 同（2020:33, 2021:37） | clinic 3/4 の {4} も複写由来で実根拠なし |
| LINE 予約設定 `line_reservation_settings` | **clinic 1 のみ登録**: business 09:00–19:00・`closed_weekdays: []`・`break_hours: []`・15 分枠 | clinic 1 は LINE 側でも定休無し → {0} と矛盾。clinic 2–4 は未設定 |
| 予約時刻の分布 | `end_time = start_time`（duration 未移行・中央値 0）・start は 09 時台に集中 | **時刻分布は証拠として使えない**。閉院時刻の導出不可 |
| shift_templates / shift_entries / reservation_type_available_slots | ほぼ空（entries 21/14/1 件・2026-09 以降のテスト投入） | 証拠不足 |
| `line_reservation_settings.business_hours` | clinic 1 = 09:00–19:00 | 八王子の実営業は 19:00 までの可能性（#252 の 18:30 締めより後） |

**結論（証拠の評価）**: 現行 STG 値はどのフィールドも「医院の実値」ではない。曜日休診フラグは clinic 1 の日曜休・clinics 2–4 の木曜休いずれも移行データ上の活動と矛盾するため、**実値として信用できないデモ残留**と判定する。

## 4. 本番投入案

裁定の正本は **#252 PO 裁定（2026-07-15、先方共有「締め・集計要望」資料2）: 全院を城東同値で投入**。これが唯一のオーナー承認済みの投入値であり、証拠調査の結果これを覆す確証は取れなかった。

| フィールド | 投入案（全院同一） | 根拠 |
|---|---|---|
| `closing_am_start` | **09:00** | 資料2「AM 9:00:00〜」。編集不可のため**確認のみ**（現行値 09:00 で一致済み） |
| `closing_am_pm_boundary` | **12:00** | 資料2「PM 12:00:00〜18:29:59」 |
| `closing_weekday_end` | **18:30** | 資料2「EMG 18:30:00〜翌 8:59:59」 |
| `closing_sunday_end` | **18:30** | 資料2 に曜日区別なし（#252 記載どおり平日同値） |
| `closed_weekdays` | **要オーナー確認（下記 Q1）** | #252 は「変更しない」だが現行値はデモ残留のため、空 `{}` への明示確認を推奨 |

投入後の派生レンジ（半開）: AM `[09:00, 12:00)` / PM `[12:00, 18:30)` / EMG `[18:30, 翌 09:00)`。`closed_weekdays` に該当曜日は締め対象外。

投入経路（#252 の手順どおり）: 各院で画面 `/settings/closing-time` から PATCH → 時間帯プレビュー目視 → 非機密 receipt。**過去締めの再計算は行われない**（#215）。`.env.staging` の RDS 接続先は退役済みで使用不可 — 本番投入は本番環境確定後、同じ画面経路で行う（直接 DB 更新は `closing_am_start`≠09:00 が発覚した場合のみ・事前承認必須）。

## 5. オーナーへの確認事項（open questions）

1. **`closed_weekdays`（最重要）**: 現行 {0}=日(八王子)・{4}=木(城東/敷島/Hako) はデモ値で、予約実績と矛盾（八王子は日曜が最多曜日、全院が木曜稼働あり）。#252 の「変更しない」は実値が正しい前提の指示であり、現値はその前提を満たさない。**各院の実定休を確認し、誤りなら投入時に正しい値（定休なしなら `{}`）へ直す**かを裁定してほしい。特に clinic 2–4 は seed コメント「水曜定休」vs 値 `4`=木曜の食い違いがあり、実定休が水曜なら `{3}` が正しい。
2. **日曜終了時刻**: 資料2 に曜日区別の記載なし。日曜が短縮営業（午前のみ等）の医院があれば `closing_sunday_end` を個別値にする必要（#252 は一律 18:30 を採用済み。個別修正は裁定後の運用として許容）。
3. **特別期間・休診日の初期登録**: 年末年始等で境界時刻が変わる期間は `closing_special_periods` へ、単日休診は `clinic_holidays` へ本番投入時に登録するか。現行両方空。
4. **19:00 以降の会計取り扱い**: clinic 1 の LINE 設定（09:00–19:00）・clinic 2 の seed 値（19:00）が示唆する「実際は 19:00 頃まで診療」運用では、18:30 締めにすると 18:30–19:00 の会計が PM でなく EMG 帰属になる。#252 裁定どおりでよいか（資料2 が「18:30 以降=EMG」を明示しているため通常は YES、運用差異が判明した院のみ後から個別修正、の裁定範囲内）。
5. **対象環境**: 本番 DB/環境が未構築のままなら、投入タイミング = 本番構築後の初回設定フェーズ（GOLIVE §1 item 10）であることの確認。

## 6. 実行した作業・制約

- 読み取り: STG PlanetScale（`ap-northeast-2.pg.psdb.cloud`、gitignored `scripts/stg-uat-old-db-handoff.local.env` の credential を SELECT のみに使用、`sslmode=verify-full`、`default_transaction_read_only=on`）。
- `.env.staging`（RDS 旧値・gitignored）は接続不可を確認したうえで未使用（AWS 退役 2026-07-20 どおり陳腐化）。
- 参照: `001_init.sql` L2043–2140、`clinic_settings.csv`（002_master）、`closing_settings_service.go`、`StandardClosingTimeSection.tsx`、GitHub #252 本文（`gh issue view`）、`LINMIG-233.md`、`76405f44d`/`74b9c3d69`/`09d2c9e2b` の seed 履歴、`old_db` フィールドマップ群。
- 未実施: 本番環境への接続（未構築・資格情報なし）、PATCH/DB 書き込み、Plane 更新、値の新規発明。
- docs-only 変更のためランタイム検証不要（規約どおり報告）。
