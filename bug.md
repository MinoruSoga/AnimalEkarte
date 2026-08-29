# Login demo accounts

### BUG-500（DONE）ログインページのデモアカウント欄を復元し、実在アカウントを以前UIで約10件表示する

## 状態
- DONE / main 統合済み（`ab6a78f16 fix(auth): restore DEV staff-attach demo accounts (BUG-500)`）
- DEV-only で `DEMO_ACCOUNTS`（stg-staff 約10件）表示を確認済み。再実装不要。

## 現象（履歴）
- ログインページのデモアカウントエリアが表示されない / 中身が空。
- 当時 `frontend/src/features/auth/components/LoginForm.tsx` の `DEMO_ACCOUNTS` が空配列 `[]`。

## 非対象
- merge / push / Done / 本番・STG への demo UI 再公開（本 BUG の追加作業なし）

---

# UAT 記録ルール（2026-08-28 確定）

- **製品 FAIL のみ**本ファイルへ追記する（`docs/ops/testing/TEST_ARCHITECTURE.md`）。
- env / seed / 権限不足の **BLOCKED は書かない**。PARTIAL は製品欠陥と断定しない限り書かない。
- 同一現象の **見出し重複禁止**。既存節があれば追記・更新する。
- シナリオ md に結果を書かない。証跡は `reports/uat-YYYY-MM-DD/`。
- agent-loop 用: 未完了は `### BUG-xxx…` 見出し。完了は見出しに `DONE` を含める。

---

# UAT 2026-08-28 製品 FAIL

### BUG-501（DONE）S03: 一覧からの新規接種で実施日当日デフォルトの独立フォームに到達できない

- [x] DONE — main 統合 `d7f80b4a1`（2026-08-29）

## 現象
- 予防接種一覧→新規登録→ペット選択は `/medical-records/new?petId=&tab=予防接種` へ遷移し、タブは「カルテを保存してから使用できます」。独立新規 `/vaccinations/new` は select-pet へリダイレクト。カルテ経路の追加フォームは実施日を空で開始する（`MedicalRecordVaccination` `useState("")`）。仕様 15 §1.1 / S03#1 の当日デフォルトを満たす新規画面が無い。
- Evidence: 06-form-open.png / 29-new-path-locked.png. href after PACO select = /medical-records/new?petId=11010476&tab=予防接種. Tab body: 「カルテを保存してから使用できます」. /vaccinations/new?petId=11010476 → /vaccinations/select-pet.

## 期待
- 一覧から新規接種へ進むと、実施日が当日デフォルトの独立（または等価）フォームに到達できる。
- 仕様 15 §1.1 / S03#1 を満たす。

## 受け入れ
1. 予防接種一覧→新規で、実施日当日デフォルトの登録フォームに到達できる。
2. `/vaccinations/new` 経路が select-pet 無限リダイレクトにならない（または代替経路が同等に使える）。
3. カルテ未保存でも接種新規がブロックされない、または仕様どおりの到達手段が1つ以上ある。
4. 関連ユニット/lint が通る（変更範囲）。source/merge/Done はしない。

## 非対象
- merge / push / Done / 本番変更
- seed 年の一括書き換え（BUG-502 側）

---

### BUG-502（DONE）S03: 予防接種一覧が 2029 seed の 100 件窓で当日登録の次回予定を見せない

- [x] DONE — main 統合 `ad15b6c7b`（2026-08-29）

## 現象
- API has S03 record but list search PACO did not show it (100-window of 2029 seed?). rows=[['2029-12-01', '宮坂\u3000哲夫', '大豆', 'ｲﾍﾞﾙﾒｯｸ\u3000ＰＩ-34', '2029-12-01', ''], ['2029-12-01', '宮坂\u3000哲夫', 'ひな子', 'ｲﾍﾞﾙﾒｯｸ\u3000ＰＩ-34', '2029-12-01', ''], ['2029-12-01', '東山\u3000克己', '絵瑠', '10%Ｐｒｏ-Ｈｅａｒｔ12(～20kg)', '2029-12-01', '']] api=[{'id': 1000000001, 'next_date': '2027-08-28T09:00:00+09:00', 'lot1': 'S03-LOT-FIL'}, {'id': 1000000000, 'next_date': '2026-09-10T09:00:00+09:00', 'lot1': 'S03-LOT-1'}]

## 期待
- 当日登録した次回予定が、検索（PACO 等）で一覧に見える。
- 既定の窓/ソートが未来 seed 年に支配されて直近登録を隠さない。

## 受け入れ
1. S03 で作った record が API と一覧の両方で同一ペット検索に出る。
2. 100 件窓や日付ソートが直近登録を不可視にしない（ページング/既定ソート/フィルタの是正）。
3. 変更範囲のテストが通る。source/merge/Done はしない。

## 非対象
- merge / push / Done
- 臨床 seed の全件再生成（必要な最小修正に限る）

---

### BUG-503（DONE）S06: 新規カルテ表示時の auto-draft（URL が /medical-records/:id に昇格）が起きない

- [x] DONE — main 統合 `a16fb2325`（2026-08-29）

## 現象
- `/medical-records/new?petId=11010476` (PACO / 006366-001 / エフ.ア.トイ / living) を開いても draft の自動 POST が走らず、URL は `/new` のまま。失敗バナーも「再試行する」も出ない。フローティング「保存」も新規作成しない。
- 原因: auto-create は general 予約区分が無いと effect を silent return する。clinic 2 の `GET /api/v1/masters/reservation-types` は trimming「トリミング」1件のみ（general/再診なし）。同日予約も 0 件。
- Evidence: reports/uat-2026-08-28/s06/02-after-autocreate.png / f-02-new-wait.png / f-02b-after-wait.png. hooks=[].
- **Related (S11#3)**: トリミング併用時も `/medical-records/new?petId=…` で draft 昇格せず同一根（reservation-types に general 無し）。別見出しは立てない。

## 期待
- 新規カルテを開くと draft が自動作成され、URL が `/medical-records/:id` に昇格する。
- general 予約区分が無い場合でも、silent return ではなく作成できるか、失敗がバナー/再試行で見える。

## 受け入れ
1. clinic に general が無く trimming のみでも、`/medical-records/new?petId=` から draft 昇格する（または明示エラー+再試行）。
2. 失敗時に hooks 空のまま無言で終わらない。
3. S11#3 同一根（トリミング併用）も回帰しない。
4. 変更範囲のテストが通る。source/merge/Done はしない。

## 非対象
- merge / push / Done
- 予約マスタの運用データ全投入（製品側フォールバック/エラー表示の是正が主）

---

### BUG-504（DONE）S12: 飼主編集の「連携用URLを発行」が 404 で失敗する

- [x] DONE — main 統合 `21ba41e58`（2026-08-29）

## 現象
- 未連携・生存ペットありの飼主編集で発行ボタンは出るが、POST /api/v1/owners/:id/line/link-token が 404。トースト「LINE連携用URL発行対象が見つかりません。」。URL欄は出ない。BUG-017 rewrite は 200 で非該当。
- Evidence: POST /owners/:id/line/link-token 404 for all tried owners=['10435272', '10411622', '10410201', '10411950', '10414457', '10415584'] errors=['not found', 'not found', 'not found', 'not found', 'not found', 'not found'] uiToast=['LINE連携用URL発行対象が見つかりません。'] settings={'status': 404, 'keys': ['error'], 'hasToken': False, 'hasLiff': False, 'liffHost': '', 'error': 'not found', 'text': '{"error":"not found"}'} (owner edit loads; issue API 404)

## 期待
- 未連携・生存ペットあり飼主で link-token 発行が成功し、連携用 URL 欄が出る。
- 対象外条件なら 404 ではなく、理由が分かる 4xx + UI メッセージ。

## 受け入れ
1. 条件を満たす飼主で POST link-token が 2xx かつ token/URL が UI に出る。
2. 対象なし時のメッセージが実態と一致する。
3. BUG-017 rewrite 経路を壊さない。
4. 変更範囲のテストが通る。source/merge/Done はしない。

## 非対象
- merge / push / Done
- 実 LINE 本番送信・外部 LIFF 契約変更

---

### BUG-505（DONE）S04: LINE予約設定欠落時の ErrorPage 文言が「設定の取得に失敗しました」にならない

- [x] DONE — main 統合 `f5d1d2a31`（2026-08-29）

## 現象
- `/line-reserve/2/` を開くと ErrorPage（タイトル「エラーが発生しました」・再読み込みあり）は出るが、本文が「初期化に失敗しました」。
- `GET /api/liff/2/settings` は 404。病院側 `GET /api/v1/clinics/2/line-reservation-settings` は 204。設定欠落自体は env。
- 欠落時 catch は「設定の取得に失敗しました」+ page=error をセットする実装なのに、VITE_LIFF_MOCK では useLiff が即 isReady + mock-token のため profile effect が setPage('top') で上書きし、その後 !settings ガードの汎用「初期化に失敗しました」になる（BUG-141 と同系）。
- BUG-402 は非該当（/line-reserve/2/ 200、/line-reserve/2/src/main.tsx 200 javascript）。

## 期待
- 設定欠落時は ErrorPage 本文が「設定の取得に失敗しました」になる。
- mock LIFF ready が settings エラー表示を上書きしない。

## 受け入れ
1. settings 404/欠落時に「設定の取得に失敗しました」が出る（VITE_LIFF_MOCK でも）。
2. 「初期化に失敗しました」への誤上書きが起きない。
3. BUG-402 経路（main.tsx 配信）を壊さない。
4. 変更範囲のテストが通る。source/merge/Done はしない。

## 非対象
- merge / push / Done
- 実 LIFF 資格情報の投入（env BLOCKED 自体の解消は別）

---

### BUG-506（DONE）S11: 統合会計の精算 POST が 400 で通らず未請求が残る

- [x] DONE — main 統合 `5fe8fba0a`（2026-08-29）

## 現象
- BRT-174 · pet チョコ (11065393) · clinic2 · fixture あり（course/option 価格あり · 執行相当権限）
- 未請求に medical_record + trimming が載る（steps 5–6 PASS）が、精算の確定 `POST` が **400** `参照先の組み合わせが正しくありません`
- UI は `/accounting/new?petId=11065393` のまま step 7 FAIL
- Evidence: `reports/uat-2026-08-28/s11/notes.md` · `run_brt174c.log` · complete http=400
- **Related step 8**: 上記 complete 失敗後、未請求再取得 `n=3` http=200（精算後クリア期待に対して FAIL）。complete 失敗の結果でも未請求が残ることは確認。別 BUG は立てない。

## 期待
- 未請求の medical_record + trimming 組み合わせで精算の確定が成功する。
- 成功後は未請求がクリアされる。

## 受け入れ
1. 同 fixture で complete が 2xx になり、会計が確定状態になる。
2. 精算後の未請求再取得が 0 件（または仕様どおりクリア）。
3. 400 メッセージ「参照先の組み合わせが正しくありません」が正当ケースで出ない。
4. 変更範囲のテストが通る。source/merge/Done はしない。

## 非対象
- merge / push / Done
- BRT-174 の Linear Done 遷移

---

### BUG-507（DONE）V02: 不在ID /inventory/:id が空の在庫編集フォームになる

- [x] DONE — main 統合 `c8c1e6b33`（2026-08-29）

## 現象
- `/inventory/999999001` 直叩きで 404/エラー画面にならず、タイトル「在庫編集」の空フォームが開く（品名・単位空、カテゴリ既定「医薬品」、在庫 0/0、「在庫切れ」警告、閲覧専用バナー）。
- C3-3 期待: 存在しない ID はエラー画面（白画面・空フォーム禁止）。`/inventory/new` は一般ロールで権限エラー（env）だが、不在 ID の edit 相当表示は製品欠陥。
- Evidence: reports/uat-2026-08-28/v02/c33-inventory.png · V02.inventory-form.route.C3-3

## 期待
- 存在しない inventory ID はエラー画面（空の編集フォームを出さない）。

## 受け入れ
1. `/inventory/999999001` でエラー画面（または not found）になり、空フォームにならない。
2. `/inventory/new` の権限挙動を意図せず壊さない。
3. 変更範囲のテストが通る。source/merge/Done はしない。

## 非対象
- merge / push / Done
- 権限 env の整備そのもの
