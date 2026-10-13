# 予約管理 仕様書 (Reservations)

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | page.reservations.reservation-management／画面仕様 |
| 決定状態 | 既存記載を継承（決定根拠は未確認）。個別の合意・未決記述は本文のまま |
| 実装状態 | 既存記載を継承（再監査なし） |
| 実機検証状態 | 今回未実施。過去の記録と今回の文書確認を分ける |
| 文書確認日 | 2026-10-13（命名・形式・参照・本文保全のみ） |

<a id="purpose"></a>
## 目的・対象範囲

| シナリオ | 業務上の価値 | 優先度・理由 |
|:---|:---|:---|
| 患者を選び予約する | 死亡ペットの新規予約を防ぐ | 未決／要確認（既存文書に優先度の根拠なし） |

### 概要

- **画面の目的**: 診療、トリミング、手術等の全予約をカレンダー形式で統合管理する。

<a id="access-preconditions"></a>
## 利用者・権限・前提

- **アクセス権限**: `reservations:view`（`RequirePermission` によるルートガード。デフォルト action は `view`）。作成/編集/削除は `usePermission` によりさらに `create`/`edit`/`delete` 権限で出し分け。

利用者の職種別の範囲は未記載／要確認。操作前提の詳細は本書の各操作・保存契約と関連正本を参照する。

<a id="navigation"></a>
## 入口・出口

- **URLパターン**: `/reservations`

| 入口／出口 | 契機 | 引継ぎ情報 | 根拠 |
|:---|:---|:---|:---|
| URLパターン | 画面を開く | 上記URL／既存導線。呼出元の網羅は要確認 | 本節の元記載 |
| 戻り先・他画面 | 本節の既存操作 | 記載が無い引継ぎ・戻り先は要確認 | 本節／関連フロー |

<a id="display-inputs"></a>
## 表示情報・入力項目

| 項目 | 意味 | 型／形式 | 必須／任意 | 初期値 | 候補元 | 制約／根拠 |
|:---|:---|:---|:---|:---|:---|:---|
| 患者 | 予約対象の特定 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 飼主・ペット検索 | 下記「患者の特定」 |
| 来院区分 | 初診／再診の区分 | 未記載／要確認 | 未記載／要確認 | 作成モード別（下記詳細） | 初診／再診 | 下記「予約詳細の入力」 |
| 予約内容 | 予約に付随する内容（notes） | 未記載／要確認 | 未記載／要確認 | 既存値を表示 | 未記載／要確認 | 下記「予約詳細の入力」 |
| 開始・終了時刻 | 予約時間 | HH:mm（fallbackは15分候補） | 未記載／要確認 | 終了時刻は開始選択で自動設定 | 空き枠API、fallbackはTIME_OPTIONS | 下記「予約詳細の入力」 |
| 予約区分・担当医／指名 | 枠と担当者の指定 | 未記載／要確認 | 未記載／要確認 | 未記載／要確認 | 予約区分マスタ・スタッフ候補 | [横断ルール](../../../workflows/reservation-to-record-flow.md#staff-capability) |
| 予約経路 | 入口の記録 | 未記載／要確認 | 未記載／要確認（編集モードのみ表示） | 未記載／要確認 | LINE／電話／受付／診察室／記録入力 | 下記「予約詳細の入力」 |

### 画面構成とビュー



#### 1. カレンダービュー

用途に応じて2つの表示形式を切り替え可能です。
- **月表示 (`MonthView`)**: 日付ごとの予約件数と概要を把握。セルクリックで該当日の週表示へ遷移。
- **週表示 (`WeekView`)**: 日付ごとの時間軸（15分単位）で空き枠と重複を確認。表示日数は 5日/7日 を切替可能（`DaysRangeToggle`）。URL `?days=`（未指定は 5）。休診日はカレンダー選択不可。

#### 2. インテリジェント・ツールバー

- **動的カラー凡例**: マスタ設定（`ReservationTypeGroup.color`）と連動し、診療区分ごとの配色を動的に表示。
- **マルチソース・フィルタ**: 「手動予約」と「LINE予約（source=line）」を切り替えて表示可能。
- **担当医フィルタ**: 特定の医師のスケジュールのみを抽出。
- **キャンセル済みの非表示**: `filterCalendarAppointments` が `cancelled` のみを描画前に除外する（#116）。LINE・院内のどちらからのキャンセルも同じ扱い。`no_show` は表示対象のまま。予約 API の既定取得や保存データは変更せず、履歴用途の取得には影響しない。

### 予約登録フロー (`ReservationFormModal`)

デスクトップでは「患者選択」パネルと「予約情報」パネルを左右分割で同時表示します（モバイルのみ、`StepIndicator` による Step1/Step2 タブ切替。両タブは自由に行き来可能で、順序を強制するゲートではありません）。

<a id="requirement-fr-reservations-deceased-pet-exclusion"></a>
#### 患者の特定（臨床安全ガード）

- **高速検索**: 飼主名、ペット名によるリアルタイム検索。
- **死亡ペットの保護**: ステータスが「死亡」のペットは一覧でグレーアウトされ、選択が物理的に無効化されます（`PatientSelectionTable.tsx`）。
- **結果スクロール**（EMR-318）: 患者結果パネルを高さ制約内のスクロール領域とし、キーボードでフォーカスできる。Windows の実際のスクロールは別途受入で確認する。
- **飼主詳細からの受付追加**（EMR-318）: 予約作成権限がある場合、保存済みの生存ペットの「受付」から当日フォームを開く。保存直前に選択医院の患者状態と当日予約を再取得する。当日 confirmed が1件なら予約の日時・内容を保持して checked_in へ変更し、読取時の updated_at をサーバーで照合する。別操作による版変更は409として再読込を求める。受付済／診療中／会計待ちは再刻印せず既存予約を案内し、複数候補や既存予約を含む複数ペットは受付画面で個別に選ぶ。完了・未来院・キャンセル・次回予定は再利用候補から除外する。当日予約がなければ通常の予約作成経路で受付を追加する。候補0件からの受付POST／batchには `prevent_duplicate_reception=true` を指定する。サーバーは患者単位ロック後に当日未完了予約を再検索し、既存候補があれば409、新規／次回予定昇格は0件の場合だけ許可する。409ではフォームを保持し、次の保存で再取得して既存予約を再利用する。同日完了後の別来院は許可する。実DBでの並行要求の確認は別途受入が必要。

#### 予約詳細の入力

- **来院区分の入力支援**（EMR-274）: 通常の新規作成では、新規飼主モードを「初診」、既存ペット選択中を「再診」に固定し、表示値と送信値を一致させる。同一飼主の複数ペット選択・別飼主への切替も再診、新規飼主への切替は初診となる。患者未選択時は選択制限なし。予約編集・次回予定の確定では既存値を維持し、来院区分を変更可能。新規飼主の登録成功後、予約保存のみ失敗した場合も新規飼主モードを維持して初診を送信する。
  - この入力支援は受診履歴の判定やAPI制約を追加しない。未受診でも登録済みペットを選べば再診となる。予約の来院区分は予約連携カルテ作成時の初診ウェルカム通知判定へ伝わる。一方、3日後・7日後フォローは飼主の最初のカルテ日による別判定（EMR-276の通知条件整備は別票）。訂正は予約編集で行う。
  - 予約フォーム内に既存飼主への新規ペット追加機能はない。飼主詳細等で登録後、既存ペットとして選択する場合も再診扱い。
- **予約内容**（EMR-275）: 作成・編集フォームと予約詳細は同じ表示名称を使う。既存値を表示・保存し、DB/APIの項目名は `notes` のまま。受付側（EMR-278）も表示名称は「予約内容」。
- **スマート・スナップ**: 追記（2026-10-07 監査）: 時刻候補の契約は [reservation-to-record-flow.md §2.2](../../../workflows/reservation-to-record-flow.md) が正本で、通常の新規入力では `GET /v1/reservations/available-times` の返却 slot key を時刻候補に使うのが primary。`TIME_OPTIONS`（15分刻み・ゼロ埋め HH:mm）の Select とカレンダー空き枠クリック位置の15分スナップは **fallback / edit behavior** として残る（通常の新規入力で空き枠 API を置き換えるものではない）。
- **終了時刻の自動計算**（EMR-191）: 開始時刻を選択すると終了時刻が自動設定される。LINE 空き枠スロットが存在する時刻ではスロットの終了時刻を採用し、無い場合は選択中の予約区分の `duration_minutes`（未設定・0 は既定15分）を開始時刻へ加算する（`resolveEndTimeOnStartChange`）。自動設定後も終了時刻は手動で変更可能。
- **空き状況の〇△✕表示**（EMR-170）: 予約区分＋日付を選択済みの場合、開始時刻 Select の各候補に空き状況を `〇 空きあり` / `△ 残り1枠` / `✕ 満員` の記号＋テキストで併記する（色だけに依存しない）。✕（満員）枠は選択不可。`GET /v1/reservations/available-times` に `include_unavailable=true` を付けて満員枠も取得する。判定は `remaining = min(空きスタッフ数, 予約区分定員残)` で、remaining>1 が 〇、=1 が △、=0 が ✕。スタッフ制約も定員も無い枠は remaining=null で常に 〇。LINE 空き枠設定未登録・通信エラー・区分未選択時の手動時刻入力パスではバッジは表示しない。
- **重複ブッキング警告**: 同一時間帯に同一担当者の予約が重複した場合、保存前に警告を表示（新規作成・編集保存時、およびドラッグ&ドロップでの時間変更時の双方でチェック）。
- **担当医指名**: 患者の希望に応じた「指名」設定が可能。
- **予約経路** (`ReservationRouteSelect`、編集モードのみ): LINE/電話/受付/診察室/記録入力のいずれかを選択・記録。メインの保存操作とは独立して即時保存される。

#### ユーザーインタラクション

- **ドラッグ&ドロップ**: `motion/react` を採用。週表示において、カードの上下ドラッグにより15分単位で時間を直感的に変更可能。
- **レスポンシブ**: デスクトップ環境を優先しつつ、タブレット端末での視認性も確保。

<a id="operations"></a>
## 操作・業務ルール・状態遷移

| 要件ID | 定義の所在 | 決定根拠 |
|:---|:---|:---|
| FR-reservations-deceased-pet-exclusion | [既存の要件本文](#requirement-fr-reservations-deceased-pet-exclusion) | 既存記載を継承。ID付与で合意済みにしない |

| 操作 | 実行条件 | 成功後状態／遷移 | 失敗挙動 | 根拠 |
|:---|:---|:---|:---|:---|
| 作成・編集・取消 | 操作別create／edit／delete | 主経路は下記「ステータス管理と連携」 | 重複警告・409。詳細は入力節 | 入力節・状態節 |
| 次回予定を確定 | reservations:edit、生存ペット | 実予約へ昇格 | 未記載／要確認 | 状態節・[次回予定契約](../../../workflows/reservation-to-record-flow.md#next-visit-intent) |

### ステータス管理と連携

予約詳細（`ReservationDetailModal`）から、来院後のステータスをシームレスに更新できます。

- **ステータス**: 詳細モーダルは 仮予約 / 予約確定 / 受付済 / 診療中 / 会計待ち / 完了 / キャンセル / 未来院 を扱う。来院後の主経路は `予約確定` → `受付済` → `診療中` → `会計待ち` → `完了`。
- **tentative（EMR-308）**: `status='tentative'` の appointment は「次回予定・未予約」— 日付のみ確定（`intended_date`）で時間・枠・担当医は未定。カレンダー（`MonthView`/`WeekViewAppointmentCard`）は色ではなく明示テキスト「次回予定・未予約」（`TENTATIVE_APPOINTMENT_LABEL`）で識別し、空き枠・定員を消費しない。カルテヘッダーの「次回予定」ボタン（`PUT/DELETE /v1/reservations/pets/:petId/next-visit-intent`）で作成・取消され、詳細モーダルではステータス変更でなく「予約を確定」で実予約へ昇格する。当日受付への昇格は [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) を参照。
- **期限超過の次回予定一覧（EMR-308）**: `use-reservation-management` が `GET /v1/reservations` に `status=tentative` の等値フィルタを付けて tentative を取得し、`ExpiredTentativePanel` が `intended_date` が JST 当日より前の行を actionable list として表示する。各行の「確定」は `POST /v1/reservations/:id/confirm-tentative`（`reservations:edit`）で実予約へ昇格し、死亡ペットでは確定ボタンを出さない。次回予定の管理面は LINE reminder ではなくこの院内 API/画面が担う。
- **受付への反映**: プッシュではない。受付側の 30 秒ポーリングとフォーカス再取得。
- **レコード作成**: 予約詳細からワンクリックで「カルテ作成」「入院・ホテル登録」「トリミング記録作成」を開始可能。

**来院後の主経路:**

```mermaid
stateDiagram-v2
    direction LR
    state "予約確定" as confirmed
    state "受付済" as checkedIn
    state "診療中" as consulting
    state "会計待ち" as billing
    state "完了" as done

    confirmed --> checkedIn
    checkedIn --> consulting
    consulting --> billing
    billing --> done

    note right of confirmed
        詳細モーダルでは他に
        仮予約・キャンセル・未来院 も扱う。
        キャンセル (cancelled) はカレンダー描画前に除外、
        未来院 (no_show) は表示対象のまま。
    end note
```

<a id="persistence"></a>
## 保存・取消・再試行・同時編集

予約経路の即時保存・受付追加失敗の入力保持は表示入力節、取消と次回予定確定は操作節を参照する。通常予約保存の再試行・結果不明時の手順は未記載／要確認。

<a id="errors-empty"></a>
## エラー・空状態

| 状況 | 検知／表示 | 利用者の次の操作 | 根拠 |
|:---|:---|:---|:---|
| 満員・空き枠取得失敗 | 満員は選択不可、取得失敗時は手動入力でバッジなし | 未記載／要確認 | 入力節の既存契約 |
| 受付追加の409 | フォームを保持 | 再取得して既存予約を再利用 | 入力節「患者の特定」 |
| 検索0件・その他通信失敗 | 未記載／要確認 | 未記載／要確認 | 個別仕様は要確認 |

<a id="acceptance"></a>
## 受入条件

この表は既存仕様から抽出した代表条件。すべての入力・操作・異常系への要件IDと受入条件の網羅的対応は未記載／要確認。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-reservations-deceased-pet-exclusion | 検索結果に死亡ペットが表示されている | そのペットを選ぼうとする | 選択が無効で予約対象にできない | 患者選択のこの条件を単独で確認 | 死亡ペットを選択できない | [要件](#requirement-fr-reservations-deceased-pet-exclusion)・[既存UATの所在](../README.md#受入条件の所在)。実機未確認 |

<a id="decisions-evidence"></a>
## 未決事項・関連仕様・根拠

未記載の型・必須条件・初期値・エラー・受入結果は要確認。既存の未実装・判断待ち・臨床承認待ちは以下の元記載を維持する。今回の形式整理を合意や実装完了の根拠にしない。

[文書規約](../../../authoring-guide.md)・[ドメイン入口](../README.md)。根拠は整理前の既存仕様本文と、その本文が示す日付付き追記・関連資料。文書確認日は上記の構造確認日であり、過去の承認日を置き換えない。API表・コード名は参照情報で、wire型とDB定義の正本は文書規約に従う。

### 技術仕様



#### API連携

| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/reservations` | カレンダー期間内の予約一括取得（`status=tentative` で次回予定の一覧取得にも利用、EMR-308） | `reservations` | `view` |
| POST | `/api/v1/reservations` | 新規予約の作成（重複チェック含む） | `reservations` | `create` |
| PATCH | `/api/v1/reservations/:id` | 時間変更・ステータス更新 | `reservations` | `edit` |
| PATCH | `/api/v1/reservations/:id/reservation-route` | 予約経路の更新 | `reservations` | `edit` |
| DELETE | `/api/v1/reservations/:id` | 予約のキャンセル・削除 | `reservations` | `delete` |
| GET | `/api/v1/reservations/available-times` | 予約可能時間枠＋空き状況（`include_unavailable=true` で満員枠も返却、EMR-170） | `reservations` | `view` |
| PUT | `/api/v1/reservations/pets/:petId/next-visit-intent` | ペットの次回予定（tentative）登録/更新（EMR-308。権限は操作別 — [reservation-to-record-flow.md](../../../workflows/reservation-to-record-flow.md) §2.3） | `reservations` | 新規=`create`・既存 tentative 上書き=`create`+`edit` |
| DELETE | `/api/v1/reservations/pets/:petId/next-visit-intent` | 次回予定の取消（対象不在でも冪等204、EMR-308） | `reservations` | `edit` |
| POST | `/api/v1/reservations/:id/confirm-tentative` | 次回予定（tentative）の実予約確定（EMR-308） | `reservations` | `edit` |

