# S04: LIFF 飼い主予約ジャーニー通し

> **目的**: 飼い主が LINE（LIFF）から予約を完結でき、空き枠計算がシフト・休憩・既存予約を正しく反映し、予約・キャンセルが病院側の画面と通知に即時反映されることを納品前に証明する。
> **所要目安**: 25分 / **深度**: 薄い+境界
> **仕様正本**: [line/reservation-spec.md §1〜§5](../../../spec/line/reservation-spec.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。**LIFF モックはローカル専用**: フロント（line-reserve アプリ）は `VITE_LIFF_MOCK=true`、バックエンドは `LIFF_MOCK=true` で認証をバイパスする。バックエンドは release モードでは `LIFF_MOCK` 設定を拒否して起動しない（release guard）ため、STG/本番では実 LINE アカウントで実施する。
- モック時は実 LINE への通知受信確認ができない（通知の期待結果は送信ログまたは STG で確認）。
- LINE 予約受付が「受付中」であること（LINE 予約設定 — [28-line-reservation.md](../../../spec/screens/28-line-reservation.md)）。
- 予約区分マスタに LINE 公開設定された診察系コースとトリミング系コースが存在すること（§3）。
- 指名対象スタッフに当日〜翌日の勤務シフトがあり、シフト内に休憩時間が 1 件以上登録されていること（§4 の境界確認に使用）。
- 病院側確認用ログイン: reception ロール。
- 依存シナリオ: なし（S01 で死亡登録に使ったペットの飼主は使用しない）。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | LIFF 予約アプリ（`frontend/line-reserve`）を開き、新規予約を開始する | フローは **顧客情報(ペット含む)→コース→[トリミング時: コース/オプション]→スタッフ→日付→時間→ご要望→確認→完了**（`App.tsx` step1〜。仕様§2表記と順不同） |
| 2 | コース選択で診察系コースを選択する | マスタで公開設定された予約区分のみが表示される（`GET /api/liff/:clinicId/courses` — §2-1） |
| 3 | スタッフ選択で前提条件のスタッフを指名する | 担当医の指名、または「指名なし」（`settings.show_no_staff_option` 時）が選択できる（§2-2） |
| 4 | 日付→時間選択で当該スタッフの日を開く | 空きスロットが医院設定の間隔（既定 15 分）で表示され、勤務時間内のみ提示される（§2-3・§4 Positive。API: `available-dates` / `available-times`） |
| 5 | **境界**: スタッフの休憩時間帯の前後を確認する | 休憩時間と重なる・またがる枠が表示されない（§4 Negative）。既存予約と重なる枠も表示されない（§4 Conflict） |
| 6 | ご要望 → 予約確定（ペットは先頭 CustomerInfo で選択済み） | 予約が完了する（§2-4/§2-5）。飼い主 LINE へ完了メッセージ、病院へ受付通知が送信される（§5。モック時は送信確認のみ・受信確認は STG。枠競合時はインライン notice で再選択） |
| 7 | 病院側: 予約管理のソースフィルタで「LINE予約」（value=`line`）を選ぶ | 当該予約が抽出できる（UI ラベルは「LINE予約」。§3「予約の俯瞰」） |
| 8 | 病院側: 受付カンバン（ホーム）の当日ボードを確認する | 当該予約が source=line として「受付予約」列に自動反映されている（§1・[01-reception.md](../../../spec/screens/01-reception.md)） |
| 9 | LIFF のマイ予約（予約確認）を開く | 手順 6 の予約が表示される（`GET /api/liff/:clinicId/my-reservations`） |
| 10 | マイ予約から当該予約をキャンセルする | キャンセルが成立し、飼い主・病院双方へ即時ステータス通知が送信される（§5） |
| 11 | 再度日時選択を開き、手順 6 で確保していた枠を確認する | キャンセルにより枠が解放され、同じ枠が空きとして再表示される（§4 Conflict の除外要素が消えるため） |
| 12 | 新規予約でトリミング系コースを選択する | トリミング時は `TrimmingCourseSelectPage` → `TrimmingOptionSelectPage` が挿入され、コース・オプション後にスタッフへ進む（`GET …/trimming-courses` / `trimming-options`）。**runtime 2026-08-01 BLOCKED**: `/line-reserve/1/` で `GET /api/liff/1/courses` **401** / UI『ログイン情報の有効期限が切れました。LINEアプリを再起動して開き直してください。』— ローカルは `VITE_LIFF_MOCK=true` と BE `LIFF_MOCK=true` の両方が必須 |

## 確認観点

- 予約の「変更」機能と前日リマインドは未実装（§5）— 変更導線が無いことは欠陥ではない。
- 空き枠計算はバックエンド `timeslot_engine.go`（§4 の 3 要素合算）。LIFF 認証モックは `backend/internal/middleware/liff_auth.go`。release モードでは `LIFF_MOCK=true` だと起動拒否（`backend/internal/config/config.go`: `LIFF_MOCK must not be set in release mode`）。
- FE モックフラグは `VITE_LIFF_MOCK=true`（`frontend/line-reserve/src/lib/liff-config.ts` の `LIFF_MOCK`）。ローカル `.env` と compose で BE/FE 両方を揃える。
- URL パスの clinicId と、病院側で予約が現れるクリニックが一致すること（clinic_id 隔離）。
- 予約確定時に枠が直前で埋まっていた場合は競合 notice が表示され再選択に戻ること（alert ではなくインライン — 異常時の無音失敗がないこと）。
- 予約可能枠は加算方式: 登録した開始時刻は営業時間から自動生成した枠へ追加され、他の営業時間内枠を無効化しない（`backend/internal/reservation/availability_slot_merge.go`・[28-line-reservation.md §4](../../../spec/screens/28-line-reservation.md)）。

---

## 実装突合

- 突合日: 2026-08-07
- HEAD: `844e43f69`
- 変更サマリ:
  - フローを日付/時間分割・トリミング step2b/2c を含む現行 `App.tsx` 順に更新
  - 病院側ソースフィルタ UI ラベルを「LINE予約」（value=`line`）に修正（「のみ」は無し）
  - 401 メッセージ全文・`VITE_LIFF_MOCK`/`LIFF_MOCK` release guard を明記
