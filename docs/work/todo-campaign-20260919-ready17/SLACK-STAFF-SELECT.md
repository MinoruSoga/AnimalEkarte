# SLACK-STAFF-SELECT: 担当者 0件/取得中/失敗 vs 候補あり・選択不能

状態: **再現調査 READY／実機採取 未実行（端末 ID UNKNOWN）**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-STAFF-SELECT`（643–697、重複 725–735）。保持する現場条件:

- iPad と一部 PC で獣医師（担当者）を選べない
- iPad で Chrome に変えても不可
- 治療マスタ選択は可能

本票は院内予約フォームの担当者 combobox を、**候補が無い（E）**と**候補はあるがタッチ/popover で選べない（P）**に分離する。capability の肯定フィルタは維持する。PO へ全 staff 開放を提案して回避しない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-STAFF-SELECT` の owned path および人間が読む調査票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-MANUAL-URINE.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 対象 iPad 機種 / iPadOS / Safari / Chrome 版 | なし | **UNKNOWN**。未入手なら実機分は BLOCKED |
| 動かない PC の OS / ブラウザ / ポインタ（mouse vs coarse） | なし | **UNKNOWN** |
| 動く PC との差分（同一医院・日付・区分・権限・staff 所属） | なし | **UNKNOWN**。採取前にブラウザの古さを原因にしない |
| 報告時の予約区分・日付・出勤・`capable_courses` | フィルタはコード上 fail-closed | 現場の実データは **UNKNOWN** |
| 「獣医師を選べない」が空リストか、リストが出てタップ無効か | UI は両方あり得る（後述） | **UNKNOWN**。報告一文では区別できない |

## 混ぜてはいけないケース

現場の「選べない」は次のどれでも同じ言葉になる。実装・再現・PO 提案を混ぜない。

### E — 候補リストが空（または未確定）

候補 `staffSelectOptions.length === 0`。SearchableSelect は開いても option が無い。原因はフィルタまたは取得状態。**capability を外して全 staff を出さない。**

| ID | 条件（コード） | いまの画面 | 期待する分離 |
| --- | --- | --- | --- |
| E0 取得中 | `candidatesSettled === false`。type 選択後は `reservationStaffs === undefined` で [filterStaffCandidatesByCapability](../../../frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts) L31–32 が `[]`。日付ありなら on-duty 未到着も同様 | `emptyMessage` だけ。type ありなら「この条件で対応可能なスタッフがいません」（[ReservationFormFields.tsx](../../../frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx) L225–230）。loading 専用 UI は無い | 「取得中」と 0件を分ける。orphan 確定もしない（L256–267, [resolveStaffSelectionEligibility](../../../frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts) L77–78） |
| E1 取得失敗 | `hasQueryError`: 日付ありかつ `isOnDutyError`、または type ありかつ `isReservationStaffError`（ReservationFormFields.tsx L253–255） | 時刻側は `availableTimesErrorMessage` があるが、担当者側に同等の error 文は無い。空メッセージと衝突し得る | 「スタッフ候補の取得に失敗」を空 0件と分ける。失敗を capability 0件と読まない |
| E2 出勤 0 | 日付あり、on-duty 成功、active ∩ on-duty が空。type 未選択なら empty は「この日に出勤しているスタッフがいません」 | 空リスト | シフト事実。capability を外しても埋まらないことがある |
| E3 capability 0 | type あり、metadata 到着、active ∩ on-duty の誰も `capable_courses` にその type id を持たない（filter-staff-candidates.ts L34–42）。map 欠落 staff も除外（L36–38）。`capable_courses` 欠落は `[]`（[use-reservation-types.ts](../../../frontend/src/hooks/use-reservation-types.ts) L99–106） | 「この条件で対応可能なスタッフがいません」 | **正しい空**。マスタの対応コースを直す。全 staff 開放は禁止 |

E0 と E3 が同じ empty 文なのが現行ギャップ。再現で「一瞬/ずっと空」を分けないと、P やマスタ不備を取り違える。

### P — 候補はあるが選択できない

`staffSelectOptions.length > 0`（DOM に `role="option"`）。トリガーは開けるが、行タップ/クリックで `formData.doctor` が変わらない、または popover が即閉じる。

| ID | 疑い（コード上の既知面） | 治療マスタとの差 | やってはいけないこと |
| --- | --- | --- | --- |
| P1 Dialog×Popover ヒット | 予約モーダルは [Dialog](../../../frontend/src/components/ui/dialog.tsx)。担当者は [SearchableSelect](../../../frontend/src/components/ui/searchable-select.tsx) = Radix Popover + cmdk。modal Dialog は body `pointer-events: none`。対策済み: Popover `modal`（searchable-select.tsx L133–136）、[PopoverContent](../../../frontend/src/components/ui/popover.tsx) `pointer-events-auto`（L27–29）、[shouldPreventDialogOutsideInteraction](../../../frontend/src/components/ui/dialog-portaled-overlay.ts)（dialog.tsx L65–78） | 治療マスタは [TreatmentSearchDialog](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx)（カード/リストの **入れ子 Dialog**）。SearchableSelect ではない。予約区分も popover-in-dialog の wheel 問題で [ReservationTypePickerDialog](../../../frontend/src/components/shared/ReservationFormModal/ReservationTypePickerDialog.tsx) に置換済み（ReservationTypeAndStaffFields.tsx L63–64） | 治療マスタが動くことを「SearchableSelect は全端末で健全」の根拠にしない |
| P2 タッチ / cmdk `onSelect` | CommandItem は `onPointerDown` で `preventDefault`（searchable-select.tsx L119–124）。coarse pointer では検索欄 autofocus を止める（L172–178）。`onCloseAutoFocus` も prevent（L180–183） | 治療マスタは cmdk 行ではない | 再現なしで「ブラウザが古い」を原因にしない |
| P3 テストが本物の combobox を外している | [ReservationFormModal.staff-candidates.test.tsx](../../../frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-candidates.test.tsx) L9–14: jsdom + カバレッジで Dialog FocusScope が Popover を即閉じるため **SearchableSelect を stub**。capability テストは option 集合だけ。入れ子ヒットは CI で証明されない | [searchable-select.test.tsx](../../../frontend/src/components/ui/searchable-select.test.tsx) の「担当者候補を開き選択できる（modal Popover）」は **Dialog 無し** | stub 成功を実機 PASS にしない |

P のとき候補フィルタを疑って E3 用の全 staff 開放を出さない。

## 現行経路（院内予約）

1. **入口:** [ReservationFormModal](../../../frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx) → [ReservationFormFields](../../../frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx) → [ReservationTypeAndStaffFields](../../../frontend/src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx) L153–164 の `SearchableSelect`（`triggerTestId="res-staff-trigger"`）。
2. **候補の積上げ（この順。後段が空なら前段の人数は出ない）:**
   - `useGetMasterItems("staff")` の `status === "active"`（ReservationFormFields.tsx L83–85）
   - 日付あり: `useGetOnDutyStaffs` の ID 積集合（L87–91, L175–178）
   - type あり: `filterStaffCandidatesByCapability`（L180–185）。pending map は空配列（fail-closed）
3. **表示:** option は `{ value: staff.id, label: name }`（L221–223）。disabled は付けない。
4. **選択値の orphan:** フィルタで落ちた既存 `doctor` は `fallbackLabel` で名前を残す。確定 orphan は settled かつ非 error かつ非 eligible のときだけ（filter-staff-candidates.ts L77–83）。確定時メッセージ: `STAFF_ORPHAN_REASON_MESSAGE`。submit は `isConfirmedOrphan` で拒否（ReservationFormModal.tsx L264–266, L304–306）。**loading/error を orphan にしない。**
5. **保存→再読込:** 本票の受入は「候補表示 → 選択 → 保存 → 再読込で同じ担当者」。サーバ契約の変更は対象外。無資格者（capability 無し）が option に出ないことを E3 で確認する。

肯定フィルタの正本は `capable_courses` のみ。legacy 除外フィールドは [use-reservation-types.ts](../../../frontend/src/hooks/use-reservation-types.ts) L99–106 で投影しない。欠落は「全員対応可」ではない。

## 治療マスタが「選べる」ことの読み方

報告の対照はカルテ治療タブのマスタ検索と読む（予約担当者と同じ SearchableSelect ではない）。

| 面 | ウィジェット | Dialog 内か |
| --- | --- | --- |
| 予約・担当者 | SearchableSelect（Popover + cmdk） | 親が ReservationFormModal Dialog |
| 予約・予約区分 | ReservationTypePickerDialog | 入れ子 Dialog（popover から意図的に置換） |
| 治療マスタ | TreatmentSearchDialog | 入れ子 Dialog。行は Dialog 内ボタン/カード |
| 治療タブの種別 | ネイティブ `<select>`（[TreatmentsTabParts.tsx](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx) L146–156） | ページ上 |

治療マスタ成功は「iPad であらゆる combobox が壊れている」を否定する。予約担当者特有の **Dialog 内 portaled Popover** を残す。マスタ SidePeek の親カテゴリ SearchableSelect（[TreatmentItemSidePanel.tsx](../../../frontend/src/features/master/components/TreatmentItemSidePanel.tsx) L218）は報告の対照に使わない（画面が違う）。

## 再現プロトコル（実機前に固定する軸）

同一医院・同一ログイン権限・同一日付・同一予約区分・同一 staff 所属で、動く PC と動かない PC/iPad を並べる。端末 ID は記録する。未入手なら実機セルは BLOCKED。

採取（推測で埋めない）:

1. Network: `GET /v1/masters/staffs`、`GET /v1/shifts/on-duty-staffs?date=`、`GET /v1/clinics/:id/reservation-staffs` の status と、対象 type id を含む `capable_courses` 件数
2. 担当者トリガーを開いたとき: option 数、empty 文、aria-expanded、popover の有無
3. option が見えるか（E vs P の分岐）。見えるならタップ/クリック後の `res-staff-trigger` ラベルと保存 payload の doctor
4. 同じ端末で治療マスタ検索が選べるか（対照維持）
5. `pointer: coarse` かどうか、画面幅、ブラウザ名と版。版は記録のみ。再現なしに原因へ昇格しない

## 最小提案（実装は本 unit 外）

[docs/product-philosophy.md](../../product-philosophy.md) の順: 要件は「資格のある担当者を予約に結び保存できること」。全 staff 一覧は工程を増やし無資格者を予約に載せるので①で捨てる。新しい担当者画面を足さない。

| 観測 | 最小変更 | 禁止 |
| --- | --- | --- |
| E0 | 取得中は empty 文を出さず loading。option 0 を E3 と共有しない | pending 中に全 staff を仮表示 |
| E1 | on-duty / reservation-staffs 失敗を担当者欄の error にする（時刻の `availableTimesErrorMessage` と同型） | 失敗時にフィルタ無し fallback |
| E2 | 出勤 0 の現行文を維持。シフトマスタへ | capability 緩和 |
| E3 | 空は正しい。reservation-staffs の `capable_courses` 整備。無資格者が option に出ない回帰は既存 [staff-candidates テスト](../../../frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-candidates.test.tsx) / [filter-staff-candidates.test.ts](../../../frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.test.ts) | 全 staff 開放、legacy 除外フィールド復活 |
| P1/P2 | 実機でヒット不能が残るなら、予約区分と同じ **入れ子 Dialog ピッカー**へ担当者だけ置換するか、Dialog×Popover の残ギャップを実機で直す。capability の入力（`staffSelectOptions`）は変えない | occupancy 的な別ロック、確認ダイアログ増設、capability 削除 |
| P3 | 入れ子 Dialog 実機またはブラウザ再現を受入にする。jsdom stub を P の PASS にしない | 「CI が緑」だけで iPad 完了 |

確認ダイアログは安全性の根拠にしない。担当者の資格はサーバ側の reservation staff / capability 契約が正本。FE の空リストは表示の fail-closed であり、全件表示で「選べたことにする」補償ではない。

## 停止条件

- 対象端末未入手 → 実機受入 BLOCKED。コード追跡とケース分離は本票で完了扱いにできる
- 再現が E3 のみ → 製品バグではなくマスタ。PO にコース割当を返す。フィルタ削除を提案しない
- 再現が P のみ → フィルタを触らない
- ブラウザの古さを、差分ログ無しの原因にしない
- 本票を製品 import しない。campaign ledger / `todo-issue.md` は編集しない

## 合成・実機ケース（actual は未実行）

対象: 合成 clinic。実患者名は使わない。デバイス列は UNKNOWN。

| ID | 操作 | 期待 | 混同してはいけないこと |
| --- | --- | --- | --- |
| S1 | type 選択直後（reservation-staffs 未到着） | E0。orphan 確定しない。empty を E3 文のまま出さない（現行は出る → ギャップ） | 対応者ゼロ |
| S2 | reservation-staffs または on-duty が 5xx | E1。担当者 error。全 staff を出さない | ネットワーク失敗を capability 0 |
| S3 | 出勤 0、type なし | E2 の出勤 empty | ピッカー故障 |
| S4 | 出勤あり、type の capable が 1 名 | その 1 名だけ option。非対応は出ない（既存テスト） | 全員表示 |
| S5 | option が 1 件以上見える端末で行を選ぶ | doctor が変わり、保存→再 GET で一致 | 空リスト問題 |
| S6 | 同じ端末で治療マスタ | 選択できる（報告条件） | 予約 combobox と同じ実装とみなす |
| S7 | 資格なし staff を URL/残留 id で保持 | 確定 orphan 文 + 解除。submit 拒否 | フィルタを外して保存可能にする |

実機 ID / OS / ブラウザ: **UNKNOWN**。S5 が失敗し S4/S6 が成功なら P。S4 が空で API 上 capable が 0 なら E3。S4 が空で API 上 capable > 0 なら E0/E1 かクライアント積上げバグであり、それでも全 staff 開放はしない。
