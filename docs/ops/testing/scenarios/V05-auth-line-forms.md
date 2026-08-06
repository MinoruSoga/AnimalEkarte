# V05: 認証・LINE系フォーム検証（入力・更新・DB整合）

> **目的**: 認証（ログイン・パスワード管理）・LIFF/LINE予約（飼い主側）・LINE予約設定（病院側）・Lステップ連携の全 18 フォームについて、入力バリデーション・更新の永続化・DB 整合（FK 選択肢・一意制約・URL 直叩き）が実機ブラウザ経由で機能することを納品前に証明する。
> **所要目安**: 90分 / **深度**: フォーム検証
> **仕様正本**: [screens/21-login.md](../../../spec/screens/21-login.md)・[screens/28-line-reservation.md](../../../spec/screens/28-line-reservation.md)・[screens/31-lstep-integration.md](../../../spec/screens/31-lstep-integration.md)・[line/reservation-spec.md](../../../spec/line/reservation-spec.md)・[line/architecture.md](../../../spec/line/architecture.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: admin ロール。パスワード変更・再設定（V05-2/V05-4）は業務に影響しない試験用スタッフアカウントで行い、終了時に元の値へ戻す。
- `frontend/liff`（アカウント連携）と `frontend/line-reserve`（予約）は独立 SPA（Playwright E2E 対象外）。browser-test はローカルの各アプリ URL へ直接アクセスして実行する。LINE ID トークン（LIFF 認証）が必要なステップをローカルで再現できない場合は BLOCKED と記録する。
- Lステップ Write API は一時停止中（noop — [line/architecture.md](../../../spec/line/architecture.md) 注記 2026-07-10）。Lステップ側の実タグ変化は検証対象外とし、アプリ内 DB・画面・監査ログで観測できる範囲のみ確認する。
- パスワード再設定（V05-4）はリセットメールのトークン付きリンクを取得できることが前提。取得手段のない環境では該当行を BLOCKED と記録する。

## 共通チェック手順

各フォームに対し、特記のない限り以下 C1〜C3 を適用する。個別セクションはフォーム固有の差分・重点のみを記載する。

| ID | チェック | 手順 | 期待結果 |
|:--|:--|:--|:--|
| C1 | 入力チェック | (a) 必須欄をすべて空にして保存 (b) 代表的な形式違反（型・文字数・日付・金額）を 1 件入力して保存 (c) 境界値 1 件（最小長・最小値ちょうど）で保存 | (a)(b) エラーが表示され保存されない（再読込でレコード非作成/非更新を確認） (c) 保存成功 |
| C2 | 更新チェック | 編集 → 保存 → 詳細/一覧で反映確認 → ページ再読込 → 編集フォームを再オープン | 保存値が反映される・再読込後も永続する・再オープンで保存値が初期表示される |
| C3 | DB存在チェック | (a) FK 選択肢の元マスタへ新規レコード追加 → フォームの選択肢を再確認 (b) 一意制約対象の重複登録 (c) 存在しない ID での URL 直叩き | (a) 追加分が選択肢に反映される（ハードコードでない） (b) エラーが表示され保存されない (c) 404/エラー画面（白画面・無限ロードにならない） |

**適用境界（重複検証の排除）**:

- 「部分更新が他フィールドを消さないこと（PATCH 非破壊）」は **V05-9（ページ編集）のみ**で代表確認する（BE テストで網羅済み。基本設定と同一レコードのマージ更新であり統合点として最適）。
- 「削除済み（無効化済み）マスタ参照の挙動」は **V05-6（予約コース選択）のみ**で代表確認する（同上）。
- クロステナント隔離は本シナリオのスコープ外（BE isolation テスト群が正本）。

## 1. 認証系

### V05-1 ログイン（`auth-login` / `/login`）

セッションのみでドメインデータ非作成のため C2/C3 は対象外。手順 4 はレート制限がかかるため最後に実施する。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 両欄空・不正メール形式（`a@`）でそれぞれ送信 | エラーが表示されログインされない |
| 2 | 正しいメール + 7 文字パスワードで送信 | ログインされない。**FE/BE 乖離の重点確認**: FE input は minLength=6 のため通過するが、BE が 8 文字未満を拒否する（`LoginForm.tsx` vs `backend/internal/auth/http_response.go`） |
| 3 | 正しい資格情報で送信 → ページ再読込 → `/login` を直アクセス | ダッシュボードへ遷移。再読込後もセッション維持（httpOnly Cookie — [21-login.md §3.1](../../../spec/screens/21-login.md)）。ログイン済みの `/login` 直アクセスは即リダイレクト（`LoginForm.tsx` の認証済み Navigate） |
| 4 | 誤パスワードで 1 分内に 6 回連続送信 | レート制限（5 回/分）で拒否される（[21-login.md §1.2](../../../spec/screens/21-login.md)） |

### V05-2 パスワード変更（`auth-change-password` / 全画面共通 Sidebar アカウントメニュー）

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 3 項目すべて空で保存 | エラーが表示され保存されない（3 項目とも必須） |
| 2 | 新パスワード 7 文字 / 確認欄と不一致でそれぞれ保存 | それぞれエラーが表示され保存されない（FE: 8 文字以上・一致必須）。各欄の表示/非表示トグルが機能する |
| 3 | 英字のみ 8 文字（例: `abcdefgh`）で保存 | FE は通過するが BE パスワードポリシー（英字+数字混在必須 — [21-login.md §3.2](../../../spec/screens/21-login.md)）で拒否され、保存されない |
| 4 | 現在のパスワードを誤って入力し、妥当な新パスワードで保存 | エラーが表示され保存されない |
| 5 | 英数字混在 8 文字ちょうど（境界）で保存 → 再ログイン | 成功トースト「再度ログインしてください」→ 新パスワードでログイン可（C2 相当の永続確認）。確認後は元の値へ戻す |

### V05-3 パスワードリセット申請（`auth-forgot-password` / `/forgot-password`）

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | メール欄空で送信 | エラーが表示され送信されない |
| 2 | 登録済みメールアドレスで送信 | 送信案内（成功メッセージ）が表示される |
| 3 | 存在しないメールアドレスで送信 | 手順 2 と**同一の成功表示**（アカウント列挙防止 — [21-login.md §2](../../../spec/screens/21-login.md)。不存在でもエラーを出さないのが期待挙動） |

### V05-4 パスワード再設定（`auth-reset-password` / `/reset-password?token=`）

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | token なしの URL を直叩き | 「無効なリンクです」画面が表示され、再申請への導線がある（`ResetPasswordPage.tsx`・C3(c) 相当） |
| 2 | 必須空 / 7 文字 / 確認欄不一致でそれぞれ確定 | それぞれエラーが表示され保存されない（FE: 必須・8 文字以上・一致） |
| 3 | 英字のみ 8 文字で確定 | BE で拒否される（V05-2 手順 3 と同一のポリシー — [21-login.md §3.2](../../../spec/screens/21-login.md)） |
| 4 | 有効トークン + 英数字混在 8 文字で確定 | 成功し `/login` へ遷移。新パスワードでログイン可 |
| 5 | 手順 4 で使用済みのトークンで再度確定 | 拒否される（ワンタイムトークン・有効期限 1 時間 — [21-login.md §2](../../../spec/screens/21-login.md)。期限切れの実測は時間都合でスキップ可） |

## 2. LIFF・LINE予約（飼い主側）

### V05-5 LIFF LINEアカウント連携（`liff-account-link` / LIFF アプリ URL に `?token=`+`clinic_id` 付きアクセスで自動実行）

入力欄なしの自動実行フォーム（ページ表示＝連携実行）。エラー分岐の網羅が本体。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | clinicId または linkToken を欠いた URL で起動 | 無効 URL 表示となり、連携は実行されない |
| 2 | 有効な連携 URL で起動 | 連携成功表示。飼い主に LINE が紐づく（Identity Mapping — [line/architecture.md §2](../../../spec/line/architecture.md)）。院内 `/owners/:id` の LINE 連携セクション（V05-11）にも反映される |
| 3 | 連携済みの状態で再度同じ連携を実行 | 409「すでに連携済み」の専用表示となり、二重紐付けされない（C3(b) 相当） |
| 4 | 無効・期限切れの linkToken で起動 | 400 系のトークン無効/期限切れ表示となり、連携されない |

### V05-6 LINE予約 予約作成（`line-reserve-create` / line-reserve アプリ: 顧客情報 → ご要望 → 確定）

前段画面（コース・スタッフ・日時・顧客情報・ご要望）は画面内 state のみで、永続化は確定画面の予約 POST 1 箇所。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | コース選択肢を確認。院内の予約区分マスタで新規区分を公開 → 再表示 | マスタで公開設定された区分のみ表示され、追加分が反映される（[reservation-spec.md §2](../../../spec/line/reservation-spec.md)・C3(a)） |
| 2 | 【代表・無効化マスタ】予約作成後にそのコース区分を無効化 → 飼い主側と院内を再確認 | **一覧は inactive を除外**（GetCourses が !IsActive を skip）。確定 POST も inactive 拒否。既存予約表示は継続 |
| 3 | お名前・電話番号をスペースのみにして次へ | エラーが表示され進めない（FE: trim 後の非空必須） |
| 4 | 電話番号に数字以外（`abc`）を入力して進める | **形式拒否なし（2026-08-01 コード実測）**。FE `CustomerInfoPage.tsx` は `phone.trim()` 非空のみ。BE `liff_validation.go` は customer_fields 各 string ≤500 のみで phone 形式検証なし。`type=tel` に pattern なし。数字以外でも FE 次へ・BE 受理し得る（院内保存値の目視は LIFF 確定パス依存） |
| 5 | 新規ペット追加でペット名を空のまま追加 | エラーが表示され追加できない（名前非空必須）。既存紐付けペットが 1 頭なら自動選択される（`CustomerInfoPage.tsx`）。既存ペット選択肢は飼い主の実データ由来（C3(a)） |
| 6 | ご要望メモに 1001 文字を入力して確定（境界: 1000 文字は成功） | 拒否され保存されない（BE: request_text ≤1000 文字 — `backend/internal/reservation/liff_validation.go`） |
| 7 | 正常フローで確定 | 完了表示。院内 `/reservations` に source=line で自動反映（[reservation-spec.md §1](../../../spec/line/reservation-spec.md)・C2 相当）。マイ予約一覧にも表示される。LINE 完了通知の実配信はローカルでは観測対象外（同 §5） |
| 8 | 選択した時間枠へ院内側で先に予約を入れてから確定 | 409「選択された時間枠は既に予約が入っています」が表示され保存されず、再選択に誘導される（枠競合 — `backend/internal/reservation/reservation_validators.go`。C3(b) 相当） |

### V05-7 LINE予約 マイ予約キャンセル（`line-reserve-cancel` / line-reserve アプリ マイ予約一覧）

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | キャンセルボタン押下 | インライン確認が表示され、確認前は予約が変更されない |
| 2 | 確認して実行 → ページ再読込 | 即時キャンセル表示（楽観的更新 — `MyReservationsPage.tsx`）。再読込後も cancelled が永続（C2）。削除ではなく status 更新である |
| 3 | 院内 `/reservations` で同予約を確認 | キャンセルが反映されている。双方向の通知はローカルでは観測対象外（[reservation-spec.md §5](../../../spec/line/reservation-spec.md)） |
| 4 | 当日（直前）の予約でキャンセルを試行 | **キャンセル期限なし（2026-08-01 コード実測）**。FE `MyReservationsPage` は confirmed なら常にキャンセル UI。BE `CancelReservation` → `CancelByID` は non-cancelled 行の status 更新のみで、当日/リードタイム/直前判定は存在しない。当日直前もキャンセル成功し得る |

## 3. LINE予約 病院側設定

### V05-8 稼働・受付ルール設定（`line-reservation-settings` / `/line-reservation/settings`）

clinic 単位 1 レコードの全量 PUT（一意制約は UI 上到達不能のため C3(b) 対象外）。C2 は受付トグル・受付期間・表示月数・スロット間隔/モード・電話番号・通知メールの一式で実施。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | LINE予約受付を「停止中」で保存 → 飼い主側予約アプリを起動 | 保存され永続する（C2）。**2026-08-01 実測**: hospital PUT `status=stopped` 200 → public `GET /api/liff/1/settings` も `status:"stopped"`。コード意図は `App.tsx` で maintenance（`MaintenancePage`「メンテナンス中」・unit test あり）。**ただし runtime の owner SPA は Top（「新規予約」）のまま** — LIFF 初期化後の `setPage('top')` が maintenance を上書きし得る（**BUG-141**）。測定後 `running` へ復元済み |
| 2 | 受付期間・表示月数・スロット間隔へ 0/負値/範囲外を入力して保存 | FE の native min により拒否される: 最長受付 `min=1`・表示月数 `min=1`・スロット間隔 `min=5`。保存時 alert「値は 1 以上にする必要があります。」で API 未到達。最短受付は `min=0` で 0 入力可（2026-07-31 実測）。BE 到達時の境界は FE 通過後のみ別途確認 |
| 3 | 営業時間・休憩時間を編集して保存 | HHMM 形式で永続する（BE: break_hours は `[{start,end}]` HHMM 形式必須 — `backend/internal/reservation/line_reservation_setting_service.go`）。曜日別営業時間の有効/無効切替も再オープンで保持される |
| 4 | チャネル ID・LIFF ID を入力して保存 → 再読込 | 保存され永続する。チャネルシークレット・アクセストークンはこの画面では入力・再表示しない（2026-07-31 実測: 入力欄はチャネルID・LIFF ID のみ。secret/token ラベル・input なし） |

### V05-9 表示ページ編集（`line-reservation-page-editor` / `/line-reservation/page-editor`）

必須項目なしの 5 テキストエリア。基本設定（V05-8）と同一レコードへのマージ PUT。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | ヘッダーテキスト・予約時注意事項・キャンセル時注意事項・プライバシーポリシー・リクエスト例を編集して保存 | 保存成功。再読込・再オープンで永続（C2） |
| 2 | 飼い主側予約アプリで表示確認 | 編集した文言が反映される（[28-line-reservation.md §2](../../../spec/screens/28-line-reservation.md)） |
| 3 | 【代表 PATCH 非破壊】保存後に基本設定（V05-8）を再表示 | 受付期間・スロット間隔・クレデンシャル・曜日別営業時間が消えていない（同一エンティティのマージ更新） |
| 4 | 長文（1 万字程度）を保存 | **BE 上限あり・FE maxLength なし（2026-08-01 実測）**。`header_text`/`request_example` max=2000（1 万字 header → **400** `header_text は 2000 以下で入力してください`）。`reservation_notice`/`cancel_notice` max=10000（1 万字 notice → **200** 受理）。`privacy_policy` max=100000。FE textarea に client max なし |

### V05-10 LINE予約枠設定（`line-reservation-slots` / `/line-reservation/slots?typeId=`）

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 予約区分ツリーを確認 | 予約区分マスタの実データ由来（C3(a)）。無効区分は「（無効）」表記で選択可能（[28-line-reservation.md §4](../../../spec/screens/28-line-reservation.md)） |
| 2 | 日付セルをクリックし特定日枠（開始時刻）を追加 → 再読込 | 15 分刻みで追加でき、永続する（C2）。営業時間から自動生成される枠への「加算方式」の案内が画面上部に常時表示されている（同 §4） |
| 3 | 枠 1 件登録済みの区分を飼い主側で予約 | 登録した開始時刻が営業時間由来の枠へ追加される。営業時間内の他の時刻も予約可能なままで、枠のない日も営業時間から自動生成される（加算方式 — 同 §4） |
| 4 | 枠をすべて削除して飼い主側を再確認 | 追加分だけが消え、営業時間設定からの空き枠自動生成は継続する（同 §4） |
| 5 | 同一日×同一開始時刻を重複追加 | 拒否される（C3(b)）。既存時刻選択時に「この時刻は既に登録済みです」表示・追加ボタン disabled。画面上部案内も「重複する時刻は追加されません」（2026-07-31 実測。POST は初回のみ 201） |
| 6 | 毎週枠の表示・不正 typeId/親区分 ID で URL 直叩き | 毎週枠は読み取り専用（リピートアイコン付き。登録・削除は予約区分マスタ側 — 同 §4）。不正 typeId は有効な末端区分へ自動フォールバックし白画面にならない（`LineReservationSlotsSettings.tsx` の typeId 正規化・C3(c) 相当） |

### V05-11 飼い主⇄LINE顧客 紐付け/解除（`owner-line-customer-link` / `/owners/:id` 内 LINE 連携セクション）

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 紐付けダイアログの選択肢を確認 | 未紐付けの LINE 顧客のみ表示される（紐付け済みは除外 — 一意制約を選択肢除外で担保。C3(a)(b) を兼ねる）。未紐付け顧客 0 件時は紐付けボタン disabled（`LinkedLineCustomers.tsx`） |
| 2 | LINE 顧客を選択して紐付け → 再読込 | セクションに反映され永続する（C2） |
| 3 | 紐付けを解除 → 再読込 | 一覧から消え永続する（owner_id の null 更新）。解除した顧客が再び紐付け候補に現れる |

- BE 側の重複紐付け 409 ガード（Q22）は UI の選択肢除外により通常到達しない（BE テスト正本）。

## 4. Lステップ連携

### V05-12 Lステップ連携設定（`lstep-settings` / `/settings/integrations/lstep`）

clinic 単位 1 レコードの PATCH（C3(b) は UI 上到達不能）。シークレット 3（APIキー・channel access token・channel secret）+ テキスト 2（LIFF ID・ベース URL）+ 数値閾値 23 フィールド。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 閾値数値・LIFF ID・ベース URL を変更して C2 一式 | 永続し、再オープンで保存値が初期表示される |
| 2 | CPM バージョンを切り替えて保存 | V1/V2 の選択式（[31-lstep-integration.md §1.2](../../../spec/screens/31-lstep-integration.md)）。保存後に自動管理タグ体系が選択バージョンに対応する（同 §2.1） |
| 3 | 閾値へ 0 または負値を入力して保存 | FE の spinbutton `min=1` により 0 は invalid。保存時 alert「値は 1 以上にする必要があります。」で API 未到達（2026-07-31 実測）。読み出し側の 0 以下デフォルト補完（`lstep_settings_thresholds.go`）は FE 通過後の防御層 |
| 4 | シークレット 3 種を保存 → 再度開き空欄のまま別項目のみ変更して保存 | シークレットは「空欄=変更なし」として維持され、上書き消去されない（`lstep-settings-form-request.ts` の空値スキップ） |
| 5 | Lステップ/LINE の接続テストボタンを実行 | 結果（成功/失敗）が表示される。ローカルの疑似クレデンシャルでは失敗表示で可（導線と表示の確認が目的） |

### V05-13〜17 同ページ・薄いフォーム群（共通手順参照 + フォーム別差分表）

いずれも C1（必須・空）→ C2（保存 → 再読込永続）を共通手順どおり実施し、差分のみ以下に示す。

| # | フォーム（id） | ルート | 必須フィールド | 一意制約 | 特記チェック |
|:--|:--|:--|:--|:--|:--|
| V05-13 | 配信優先順位（`lstep-trigger-priority`） | `/settings/integrations/lstep` 内セクション | 各トリガーの優先順位（1 以上） | 同値可（同一優先階層） | 0 を入力 → 「優先順位は1以上を指定してください」で保存されない（C1 — `TriggerPrioritySection.tsx`）。変更時のみ保存ボタン活性。同値は許可（UI 文言「同値は同一優先階層として扱われます。」・seed も ノミダニ/フィラリア=4・ワクチン30/60=8 等を保持 — 2026-07-31 実測） |
| V05-14 | タグコードマッピング（`lstep-tag-code-mappings`） | 同上 | tagName 単位の entries（全量置換 PUT） | tagName 単位で置換 | 編集 → 保存 → 再読込で永続（C2）。**形式違反は BE 400（2026-08-01 実測）**: codes=[] → `codes must contain at least one entry`; codes=[''] → `codes must not contain empty values`; code_type=invalid_type → `invalid code_type: invalid_type`。空 entries PUT は 200（全削除・batch3 と同契約） |
| V05-15 | タグ設定（`lstep-tag-config`） | 同上（追加フォーム 2 種） | フォーム1: プレフィックス+カテゴリ / フォーム2: 疾患コード+タグ名 | プレフィックス重複 409 | 片方空で追加 → 「プレフィックスとカテゴリは必須です」「疾患コードとタグ名は必須です」（C1 — `LstepTagConfigSection.tsx`）。追加 → 再読込永続 → 行削除（C2）。同一プレフィックス再追加は POST 409（2026-07-31 実測）。**同一疾患コードも 409（2026-08-01 実測）**: 新規 POST 201 後、同一 `condition_code` 再 POST → **409** `lstep_condition_tag_mapping '' already exists`（DDL UNIQUE）。seed コード再追加も同様 409。テスト行は DELETE 204 で後始末 |
| V05-16 | 友だち属性 CSV 取込（`lstep-csv-import`） | `/lstep/analytics` 内セクション | CSV ファイル | — | 未選択・空ファイルで実行 → 「CSVファイルを選択してください」（C1 — `LstepCsvImportSection.tsx`）。取込後に履歴一覧へステータス行が追加される（C2 相当）。**列不正は 400（2026-08-01 実測）**: `POST .../lstep/csv-imports/friend-attributes` に `foo,bar` ヘッダのみ → **400** `required column not found: line_user_id (expected one of: LINE ID, line_user_id, userId)`（`lstep_csv_helpers.go`） |
| V05-17 | タグ一括解除（`lstep-bulk-tag-remove`） | `/settings/lstep/tags` の対象者ドロワーから起動 | 対象タグ + 対象飼い主（起動元で確定・ダイアログ内入力なし） | — | 「この操作は取り消せません」の確認ダイアログ経由でのみ実行可。進捗バー付き逐次実行・実行中キャンセル不可（`BulkTagRemoveDialog.tsx`）。**解除後の観測点（2026-08-01 実測）**: 手動タグ `優良顧客` で `DELETE /owners/:id/lstep/tags/:tag` が **204** でも、直後の `GET tag-summary` の `owner_count` と `GET .../lstep/owners?tag=` 件数は変化しない場合がある（Write API 停止/同期オフ時は外部タグ・キャッシュ非更新）。FE は `invalidateQueries(lstepTagSummary)` + toast「…名から解除しました」— 一覧再取得後も件数が同じなら「UI 上は成功だが件数不変」を記録。LSTEP 実タグは観測対象外 |

### V05-18 健診対象者一括タグ付与（`lstep-checkup-sync-create` / `/lstep/checkup-sync`）

前段の抽出条件フォームはプレビュー取得のみ（非永続・検索）だが境界検証が濃いため本エントリに統合。永続化は確認ダイアログからの POST。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 抽出条件の年齢・累計金額・年間来院回数に負値/小数を入力 | 0 以上の整数のみ受理され、違反はエラーで抽出できない。最小年齢 > 最大年齢もエラー |
| 2 | 検診種別未選択・対象者 0 件のまま実行 | 検診種別なしでは抽出できず、対象 0 件では実行ボタン disabled |
| 3 | 確認ダイアログでタグ名を空/スペースのみにする | 実行ボタンが無効（trim 後非空必須。検診種別由来の初期値あり） |
| 4 | タグ名に自動管理タグ名（例: `CPM_01_出会い` — [31-lstep-integration.md §2.1](../../../spec/screens/31-lstep-integration.md)）を指定して実行 | 拒否される（BE: 「tag_name は自動管理タグのため使用できません」 — `lstep/checkup_sync_service_create.go`） |
| 5 | Lステップ API 未設定の状態で実行 | 拒否される（BE: 「Lステップ API が設定されていません」 — 同ファイル） |
| 6 | API 設定済みで妥当なタグ名で実行 | 完了表示。実行が `audit_logs` に記録される（DB 参照は USER 実施 — S01 と同運用）。Write API 停止中のため Lステップ側実タグは変化しない |

## 確認観点

- 既存の機械テストが覆う範囲: FE component test（`ChangePasswordDialog` / `ForgotPasswordPage` / `use-liff-link` / `CustomerInfoPage` / `ConfirmPage` / `MyReservationsPage` / `LstepSettingsForm` / `TriggerPrioritySection` / `LstepTagCodeMappingsSection` / `LstepTagConfigSection` 等）が FE 単体のバリデーション分岐を、BE テスト（auth/password_reset・liff_validation・line_reservation_setting・lstep_settings/tag/csv/checkup_sync 各 service/handler test）がサーバ側検証・部分更新非破壊・テナント隔離を網羅する。E2E（auth-flows / line-reservation-flow / lstep-flow）は表示と主要導線のみ。
- 本シナリオは上記が個別レイヤで検証済みの挙動を**実ブラウザ + 実 DB の統合点（FE→BE→永続化→再表示）で通す受け入れ時の実機検証**であり、FE/BE バリデーション乖離（ログイン最小長 6 vs 8・パスワード英数字混在は BE のみ）・E2E 対象外の独立 SPA（liff / line-reserve）・予約可能枠の加算方式を重点とする。
- ログインのレート制限・リセットトークンのワンタイム性・リセット申請の列挙防止はセキュリティ境界（[21-login.md](../../../spec/screens/21-login.md) §1.2/§2）— 「拒否される/漏れない」ことを必ず確認する。
- NG 項目は [`STATUS.md` §3 受入バグ（正本）](../../../../STATUS.md) へ `## BUG-XXX:` 節として起票する（ローカル連番 最大+1・[README.md](README.md) のルールに従う）。
