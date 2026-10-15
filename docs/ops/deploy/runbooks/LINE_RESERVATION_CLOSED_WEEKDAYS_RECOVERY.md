# Runbook: line_reservation_settings.closed_weekdays 破損の運用復旧

> **目的**: `line_reservation_settings.closed_weekdays`（JSON）が破損し、管理 UI（設定フォーム・ページエディタ）が保存できなくなり、LIFF 予約が fail-closed で拒否されている clinic を、承認済みの認証付き設定 API で復旧する手順を定義する。
> **読者**: 運用担当（対象 clinic に対する `hospital-settings:view`（GET）と `hospital-settings:edit`（PUT）の両権限を持つ承認済みオペレーター）。
> **タイミング**: 破損検知後、対象 clinic の合意済み休診曜日が明確になっているとき。

> **現状の検証（2026-10-15, origin/main 4fa382b83 時点・コード静的確認のみ）**:
> 復旧はこの時点の API 契約に基づく。以降の変更で契約が変わった場合は本書を更新してから実行する。

## 背景

- 破損した `closed_weekdays` はフロントエンドでは `parseClosedWeekdays` が `null` を返し、`LineReservationSettingsForm` と `LineReservationPageEditor` の両方で保存がブロックされる（UI からの自己修復が不可能）。
- バックエンドは fail-closed 設計のため、破損行があると LIFF の空き日付計算と予約検証がエラーになり、**LINE 予約が全件拒否され続ける**（`reservation/available_dates.go`・`reservation/reservation_validators.go` A-3）。
- 復旧では DB 直操作（SQL・管理コンソール経由の UPDATE を含む）を**禁止**する。本書内で SQL を実行してはならない。必ず認証付き API を使い、clinic が合意した休診曜日のみを是正する。

## 制約（必読）

- **実行者と事前確認**: 対象 clinic に対する `hospital-settings:view` と `hospital-settings:edit` の両権限を持つ認証済みオペレーターのみ。実行前に、対象環境（STG/本番）と対象 clinic を確定し、他のオペレーターと設定変更の同時作業がないことを確認・調整する。read→modify→PUT の一連の間は同時編集を停止しておく。API には楽観ロック（version 照合）が**存在しないため**、途中で他人の変更が観測されたら直ちに中止する。
- **partial PATCH は存在しない**: このリソースに PATCH はなく、`PUT` は**フル upsert** である。実装は repository で `Select(lineReservationSettingUpdatableColumns()).Updates(&intended)` を行い、**allowlist 内の全フィールド（false・0・空文字を含む）が書き込まれる**。したがって省略フィールドはゼロ値/既定値で上書きされ、「closed_weekdays だけ直す」場合でも必ず直前の `GET` 本文から組み立てた**全 allowlist フィールドを含むペイロード**を返送する。
- **書き込み対象は allowlist のみ**: `id`・`clinic_id`・`created_at`・`updated_at` は応答メタデータであり送信しない。`line_access_token`（資格情報）は GET 応答に含まれず、service 層が既存値を保持する契約のため送信も確認もしない。他の資格情報の取得・閲覧・送信も行わない。
- **`[]` への初期化は承認が必要**: 「正しい曜日が不明だから一旦空配列」という対応は、休診曜日の無効化を意味するため、clinic の明示承認がない限り行わない。
- **実際の運用実行は未検証**: 本書は静的コード確認に基づく手順書であり、実環境での実行記録は存在しない。実行時は結果（GET/PUT の応答・UI 再確認）を運用記録に残すこと。

## 1. 検出

症状のいずれかに該当したら破損を疑う:

1. 管理 UI: LINE 予約設定フォーム／ページエディタで `closed_weekdays` 不正エラーが表示され、保存できない。
2. LIFF: `GET /api/liff/{clinicId}/available-dates` がエラー応答（invalid input の実装上 400 系。例: 「休診設定の形式が正しくありません」「closed_weekdays は 0〜6 の整数配列である必要があります」）となり、予約カレンダーが出ない。
3. 予約作成が fail-closed で拒否され続ける。

確定は、自 clinic 権限内の認証付き `GET /api/v1/clinics/{clinic_id}/line-reservation-settings`（手順 2.1 と同じ）の応答 `closed_weekdays` を手元で判定する:

- 有効: 各要素が整数 `0`〜`6`、またはレガシー文字列 `"0"`〜`"6"` の配列。数値と該当文字列の混在も有効。**重複は有効として受け付けられ保持される**（不正ではない）。未設定・JSON `null`・空配列は休診曜日なしとして扱われ、破損とは区別する。
- 破損: 上記の未設定・JSON `null` を除く非配列、または範囲外・小数・boolean・`null` 要素などを含む配列。
- 判定基準は `parseClosedWeekdaysJSON`（`backend/internal/reservation/available_dates.go`）に一致させる。複数 clinic の一括監査や SQL による検出は本書の対象外とする。

## 2. 復旧（API のみ）

対象 clinic の合意済み休診曜日（0=日 … 6=土 の整数）を確定してから実行する。手順全体を通して同時編集がないことを維持する（制約参照）。

1. 現状を取得して全文を保存する（ロールバック用の原本）:

   ```
   GET /api/v1/clinics/{clinic_id}/line-reservation-settings
   ```

   - 200: 本文をそのまま保存（`closed_weekdays` に破損値が入っている）。
   - 204: レコード不在。破損ではなく未設定 → 本書の対象外、手順中止。
   - 401/403: 権限不足。この時点で中止し権限の付与可否を確認する。
   - 5xx・タイムアウト: 結果は不明。盲目的に再送せず、読み戻しで状態を確認してから再開を判断する。

   本文はアクセス制限された作業場所だけに保管する。応答の `clinic_id` が対象医院と一致することを確認する。他の設定に不正値がある、または保存時の既定値補完によって曜日以外も変わる場合は、この手順では保存せず別途是正を合意する。

2. 取得した本文からペイロードを組み立て、**`closed_weekdays` のみ**を合意済みの整数配列に置き換えて送信する:

   ```
   PUT /api/v1/clinics/{clinic_id}/line-reservation-settings
   ```

   ペイロード構築例（ローカルでの値組み立てのみを示す。通信コマンド・fetch・資格情報の入出力は含めない）:

   ```javascript
   // current: 手順 2.1 の GET 応答をパース済みのオブジェクトとして扱う（その時点の値が「現在値」）。
   // approvedWeekdays: clinic が合意した休診曜日の整数配列（0=日 … 6=土）。
   //   例として [1, 2] を想像しているが、実際は必ず対象 clinic の合意値を使う。
   const WRITABLE_FIELDS = [
     "status", "header_text", "reservation_notice", "cancel_notice",
     "privacy_policy", "closed_weekdays", "closed_dates",
     "national_holiday_closed", "business_hours", "business_hours_by_weekday",
     "break_hours", "daily_limit", "monthly_limit",
     "booking_window_max_days", "booking_window_min_days", "calendar_months",
     "phone_number", "notification_email", "request_example",
     "time_slot_mode", "time_slot_interval_minutes", "no_staff_mode",
     "show_no_staff_option", "additional_fields", "line_channel_id", "liff_id",
   ];

   // 全 allowlist フィールドを現在値からコピーする。
   // 型変換をせず、false・0・null の明示値もそのまま保持する。
   // business_hours_by_weekday が GET 応答に存在しない場合は、そのまま欠落のままにする。
   const payload = {
     ...Object.fromEntries(
       WRITABLE_FIELDS
         .filter((field) => Object.hasOwn(current, field))
         .map((field) => [field, current[field]])
     ),
     closed_weekdays: [...approvedWeekdays],
   };
   ```

   - `id`・`clinic_id`・`created_at`・`updated_at`・`line_access_token` は**送信しない**。`line_access_token` は GET 応答に含まれず、service 層が既存値を保持する契約。ダミー値も含め資格情報を送らないこと。
   - レスポンスは既存行の復旧では **200** を期待する。**201 が返った場合は、手順 2.1 以降にレコードが消えた可能性（競合）**があるため、再送せず停止してエスカレーションする。

3. 読み戻し確認:

   ```
   GET /api/v1/clinics/{clinic_id}/line-reservation-settings
   ```

   **書き込み対象の非資格情報設定フィールドのみ**を突き合わせ確認する（`updated_at` は通常変わるため比較対象外。`id`・`clinic_id`・`created_at`・`updated_at`・`line_access_token` は比較しない）。`closed_weekdays` が送信した整数配列と一致し、他の allowlist フィールドが手順 2.1 の値から変化していないこと。一致しない場合は即時停止。

4. UI 確認: 管理画面で LINE 予約設定フォームとページエディタの両方を再読込し、エラーが消えて保存・閲覧が可能になっていることを目視確認する。

5. 動作確認: 対象 clinic の LIFF 予約カレンダーで、(a) 合意済み休診曜日に予約枠が出ないこと、(b) 営業曜日には枠が出て予約フローが進むことを確認する（テスト予約の作成可否は運用判断に従う。患者データを作成しないこと）。

## 3. 停止条件・ロールバック

- 停止条件: GET が 401/403、PUT が 400 や 5xx・タイムアウト（結果不明のため**盲目的な再送は禁止**。必ず先に読み戻して状態を確認する）、PUT の応答 201（レコード消失競合の疑い）、読み戻しで不整合、同時変更の観測、UI 再読込後もエラーが残る、LIFF カレンダーが想定と異なる。いずれかで**即時中止**し、エスカレーションする。SQL による直接修正や `[]` 初期化で回避しない。
- エスカレーション記録: 患者・飼主・連絡先・資格情報を含まないこと。GET 本文全体や生データを共有せず、最小限のステータス（各要求の結果コード・時刻）と、機微でない設定値の差分のみを添付する。
- ロールバック: 保存済みの手順 2.1 の本文（`closed_weekdays` は破損値のまま）を**送信しない**。代わりに、復旧前の直近の正しかった `closed_weekdays` 値が判明している場合は、同じ GET→置換→PUT 手順でその値へ戻す。判明しない場合は clinic の合意待ちとし、`[]` も破損値のまま運用も行わない。

## 関連

- 類似の形状監査手順（別手順書）: [BREAK-HOURS-SHAPE-AUDIT.md](BREAK-HOURS-SHAPE-AUDIT.md)
- 本手順の API 契約の実装: `backend/internal/reservation/line_reservation_setting_*.go`、`backend/internal/reservation/available_dates.go`
