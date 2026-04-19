# BUG-LINE-008: LINE予約作成で日時・営業時間・予約窓の検証が一切行われていない

## 概要

`POST /api/liff/:clinicId/reservations` が **過去日・将来遠過ぎ・営業時間外** のリクエストも全て `201 Created` で受け入れる。業務ロジック的に無効な予約が大量に DB に作成される。

## 再現（全て 201 Created になる）

```javascript
// 1. 過去日付（6年前）
await postReservation({ date: '2020-01-01', ... });   // → 201 R-20200101-0001

// 2. 予約窓超過（booking_window_max_days=30 なのに 90日先）
await postReservation({ date: '2026-07-13', ... });   // → 201 R-20260713-0001

// 3. 営業時間外（営業 9:00-19:00 なのに 5:00）
await postReservation({ date: '2026-04-25', start_time: '0500', ... }); // → 201
```

## 影響（CRITICAL）

- **過去日の予約が作成できる** — 業務ロジック破壊
- **予約窓設定が意味をなさない** — クリニック側の制御不能
- **営業時間外に予約が入る** — スタッフが対応不能（19:00 閉店後の 20:00 でも成功、朝 6:00 でも成功）
- **休憩時間中（12:00-13:00）の予約も成功**
- **DB に業務上無効なデータが蓄積**

## 追加実測（クリーンな日での検証）

他に予約が無い日を選んでも以下が全て `201 Created`:

| テスト | 日付 | 時刻 | 結果 |
|---|---|---|---|
| 休憩時間中 | 2026-05-10 | 12:15-12:30 | ❌ 201 作成 |
| 営業終了後 | 2026-05-11 | 20:00-20:15 | ❌ 201 作成 |
| 営業開始前 | 2026-05-12 | 06:00-06:15 | ❌ 201 作成 |

一方 **SLOT_TAKEN** (同一時刻重複) と **DAILY_LIMIT** (同日上限超過) は正しく検出される。
つまり業務時間関連の検証だけが抜けている。

## 原因推定

`liff_service.go::CreateReservation` が Timeslot Engine の結果を検証せず、フォームから来たリクエストをそのまま受け入れている。

**重要**: `GET /available-times` は営業時間・休憩を正しく除外している（06:00, 12:00-12:59, 19:00, 20:00 が返らない）。
つまり **表示ロジックは正常だが、サーバーサイド検証が欠落**している。

これは典型的な「クライアント側は防いでいるがサーバー側が守っていない」セキュリティホールパターン。
悪意あるユーザーが curl で直接 POST すれば任意の無効時刻で予約を作成可能。

以下の検証が必須:
1. `date >= today`
2. `date <= today + booking_window_max_days`
3. `date >= today + booking_window_min_days`
4. `start_time >= business_hours.start` かつ `end_time <= business_hours.end`
5. `start_time` が break_hours 外
6. 曜日が closed_weekdays に含まれない
7. 祝日チェック（national_holiday_closed）
8. `(date, staff_id, start_time)` が既存予約と重複しない

## 優先度

**CRITICAL** — 業務ロジック完全無効。本番デプロイ不可。

## 確認環境

- staging: `https://api.stg.noah-karte.com/api/liff/3/reservations`
- 影響: 既に id=16,17,18 の無効予約が staging DB に作成済み（クリーンアップ必要）
