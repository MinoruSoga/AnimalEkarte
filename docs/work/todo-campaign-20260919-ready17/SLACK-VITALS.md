# SLACK-VITALS: 常時表示バイタルの配置案（VitalsTab vs ReadyPanels ヘッダー）

状態: **表示配置の比較 READY／製品実装・実機受入 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-VITALS`（L124–128、索引 L421）。保持する現場条件:

- 「名前のすぐ近く」「クリックなしで表示」「時間不要」（出典 929–937。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **表示/入力操作の削減**と**診療記録の時刻保持**を分離する
- 既存保存経路を再利用し、ヘッダーへ別の永続値を作らない
- 「時間不要」を **保存時刻削除**や **測定時刻捏造**へ読み替えない
- 臨床上の時刻扱いが決まるまで persist 変更は停止（本票では **UNKNOWN**）

本票は [VitalsTab](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx) の既存 save path と、[MedicalRecordFormReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) ヘッダー配置を比較する。製品コード・テストは変更しない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-VITALS` の owned path および人間が読む調査票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-COMPLAINT.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。臨床時刻の意味は本票で決めない。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「名前のすぐ近く」が飼主名かペット名か | ヘッダーは飼主ボタン + ペット名 span が並ぶ（後述） | **UNKNOWN**。PO が対象ラベルを確定するまで片方へ実装しない |
| 常時表示する項目（T / HR / RR / 体重 / メモ） | 保存フィールドは 4 数値 + notes | **UNKNOWN**。PO 確認（todo-issue L128） |
| 最新値 vs 今回測定の臨床上の識別 | コード上は別ソース（後述）。現場の呼び方は未採取 | **UNKNOWN** |
| 空欄 / 未保存 / 複数測定の現場運用 | コードは空行・複数行・未保存フォームを区別する | **UNKNOWN**。PO 確認まで採用値を置かない |
| 「時間不要」が入力省略か表示省略か | UI は `datetime-local` 必須。BE create は `recorded_at` required | **UNKNOWN**。表示省略だけは設計可。persist は停止 |
| 臨床上の記録時刻の意味（測定時刻 / 入力時刻 / 来院時刻） | GORM `default:now()` と FE 必須入力が共存 | **UNKNOWN**。now() を測定時刻として採用しない |
| 1366×625 実機 CSS viewport | ローカル基準は CHART-FIT と同じ 1366×625 CSS px | **UNKNOWN**。実機 `innerWidth`/`innerHeight` は採取前 |
| フロント/API revision | 本票作成時 worktree HEAD `aac697645` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

現場の「バイタルが見えない / 時間が要らない」は次のどれでも同じ言葉になる。表示配置と persist 契約を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| H — ヘッダー体重 | [MedicalRecordStickyHeader](../../../frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx) L194–198 が `weight={selectedPet.weight}` を [PatientContextHeader](../../../frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx) L181–188 へ渡す。ペットマスタの体重 | **今回測定のバイタル体重**と同一視しない。ヘッダー体重は `vital_records` を読まない |
| V — 今回カルテのバイタル一覧 | `GET /v1/medical-records/:id/vitals`。[useGetVitals](../../../frontend/src/features/medical-records/api/vitals.ts) L22–37。BE は `medical_record_id` で絞り `recorded_at ASC`（[vital_repository.go](../../../backend/internal/medicalrecord/vital_repository.go) L46–50） | ペット生涯の「最新バイタル」API はこの経路に無い。一覧末尾をペット最新と書かない |
| C — クリック導線 | 右下 FAB「バイタル記録」→ modal。値はヘッダーに出ない | 「名前の近くに出ている」と FAB 到達を同一視しない |
| T — 時刻 | 追加/編集 UI は記録日時必須。create JSON も `recorded_at` required | 「時間不要」= 列削除 / default now() 捏造 / 既存行の時刻消去、と読まない |
| D — 入院日次バイタル | [DailyVitalsSection](../../../frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx) は入院タブ。別 `daily_record_id` | 一般診療カルテの SLACK-VITALS 再現対象にしない |

## 現行経路（保存は 1 系統）

ヘッダーへ **第二ストアを作らない**。入力・保存・再読込の正本は次だけ。

1. **入口（クリック必須）:** [MedicalRecordFormReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) L345–358 の [MedicalRecordFloatingActions](../../../frontend/src/features/medical-records/components/MedicalRecordFormActions.tsx) `onVitalsClick` → `setIsVitalsOpen(true)`。ボタンは L68–85。会計タブでは FAB 全体が `return null`（L53）。見積書タブではバイタルボタンを出さない（L68）。新規カルテ / 確定済 / 死亡ペットは disabled または非表示。
2. **モーダル:** [MedicalRecordFormModals](../../../frontend/src/features/medical-records/components/MedicalRecordFormModals.tsx) L69–78。`!isNewRecord && recordId` のときだけ [VitalsModal](../../../frontend/src/features/medical-records/components/VitalsModal.tsx) をマウント。`DialogContent` は `max-w-4xl max-h-[85vh] overflow-y-auto`（L31）。中身は **同じ** [VitalsTab](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx)。
3. **一覧:** `useGetVitals(medicalRecordId)`。FE は `recorded_at` で再ソート ASC（VitalsTab L88–96）。0 件はテーブル空行（[VitalsTabTable.tsx](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTabTable.tsx) L75–79）。複数件は 1 行/記録。
4. **追加:** 空フォーム `EMPTY_VITALS_ADD_FORM.recorded_at === ""`（[vitals-tab-table-model.ts](../../../frontend/src/features/medical-records/lib/vitals-tab-table-model.ts) L18–26）。未入力は「記録日時は必須です」（VitalsTab L125–126）。未来日時拒否（L128–130）。T/HR/RR/体重のいずれか必須（L142–150）。`CreateVitalInput.recorded_at` は `jstDateTimeLocalToISOString`（L158–159）。`POST /v1/medical-records/:id/vitals`（vitals.ts L46–52）。
5. **BE create:** [createVitalRequest](../../../backend/internal/medicalrecord/vital_request.go) L12–13 `RecordedAt time.Time \`json:"recorded_at" binding:"required"\``。[VitalHandler.CreateVital](../../../backend/internal/medicalrecord/vital_handler.go) L83–97。サービスは `RecordedAt: input.RecordedAt` をそのまま書く（[vital_service.go](../../../backend/internal/medicalrecord/vital_service.go) L120–124）。モデル列は NOT NULL + `default:now()`（[vital.go](../../../backend/internal/model/vital.go) L22）だが、HTTP create はクライアント値を要求する。**欠落 JSON を now() で埋めて成功扱いにする経路は handler に無い。**
6. **更新:** PATCH の `recorded_at` は optional（vital_request.go L43–44）。UI 編集行は空なら「記録日時は必須です」（[VitalsTabRows.tsx](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx) L113–114）。表示列は `formatRecordedAt`（L42）。
7. **権限・死亡:** 作成/編集/削除は `medical-records` permission と `isPetDeceased`。死亡時は保存メッセージで拒否（VitalsTab L108–113, L191–193）。

invalidate は `queryKeys.medicalRecords.vitals(medicalRecordId)` のみ（vitals.ts L54–55）。ペットマスタ `selectedPet.weight` は更新しない。

## 最新値 vs 今回測定（コード上の区別）

PO が臨床ラベルを決めるまで、UI 案では次の **3 ソースを別名のまま**出す。混ぜた「最新バイタル」ラベルをヘッダーに置かない。

| ソース | 何の値か | 取得 | 空のとき |
| --- | --- | --- | --- |
| **ペットマスタ体重** | 患者ヘッダーの Weight アイコン。名前の近くに既にある | `selectedPet.weight`（StickyHeader L198） | `weight` falsy ならアイコンごと出さない（PatientContextHeader L181） |
| **今回カルテの測定列** | この `medical_record_id` に紐づく `vital_records` | GET vitals。ASC なので **配列末尾が今回カルテ内の recorded_at 最大** | 0 件は空状態。未保存の add form は一覧に入らない |
| **ペット生涯の最新測定** | 他日カルテや入院バイタルを含む可能性 | **このカルテ経路に query が無い** | **UNKNOWN**。推測で GET を足さない |

「最新」をヘッダーに出す案は、PO が **今回カルテ末尾**なのか **ペット生涯**なのかを決めるまで停止。本票の候補は今回カルテ GET の再利用に限る。

## 空欄 / 未保存 / 複数測定（コード事実。PO は UNKNOWN）

| 状態 | コード | ヘッダー常時表示へ出すときの未決事項（PO） |
| --- | --- | --- |
| **空欄（0 件）** | テーブル「バイタル記録がありません」（VitalsTabTable L77–78）。FAB は新規以外なら押せる | プレースホルダを出すか、行を出さないか。空をマスタ体重で埋めない |
| **未保存（add form 中）** | `isAdding` のローカル state。POST 成功まで GET に出ない。`recorded_at === ""` では POST しない | 入力中の値をヘッダーに出すか（出したら第二 state）。推奨: **出さない**。既存 GET のみ |
| **未保存（新規カルテ）** | FAB disabled「カルテを保存してから利用できます」。VitalsModal 未マウント | 新規中は常時表示なし。ペット体重だけが名前の近くにある |
| **複数測定** | 同一 MR に N 行。時刻順。グラフトグルあり | 末尾 1 件か全件か。全件は 1366 幅で折れる。PO 確認まで全件ヘッダー固定を選ばない |
| **部分空欄（T だけ等）** | 1 項目以上あれば作成可。`displayNum` は null を `"-"` | ヘッダーで `-` を出すか項目ごと隠すか |

これらの採用値は **UNKNOWN**。コードが区別できることと、現場がどう呼びたいかは別である。

## 配置オプション比較（実装しない。ヘッダーは表示専用）

制約: 既存 9 タブを保つ（[medical-record-form-model.ts](../../../frontend/src/features/medical-records/routes/medical-record-form-model.ts) L4–14: 問診 / 診察/治療プラン / 治療 / 予防接種 / 定期健診 / 検査 / 画像 / 見積書 / 会計(医師確認)）。タブ list は StickyHeader L214–216 で `overflow-x-auto`。狭い画面の基準は [UAT-R2-CHART-FIT](../todo-campaign-20260918/UAT-R2-CHART-FIT.md)（1366×625 CSS px ローカル基準。実機 UNKNOWN）。

全案とも **書き込みは既存 VitalsTab POST/PATCH/DELETE**。ヘッダーは `useGetVitals`（同一 query key）の read か、現状維持。

| ID | 配置 | 名前の近く | クリックなし | 時刻 UI | 9 タブ / 1366×625 | recorded_at persist | 備考 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **O0 現状** | 右下 FAB → VitalsModal → VitalsTab | いいえ。名前列は飼主/ペット/マスタ体重 | いいえ。値は modal 内 | テーブル列「記録日時」+ 入力 `datetime-local` | FAB はタブを圧迫しない。modal `max-w-4xl` は 1366 で横スクロールし得る | 変更なし | 現場条件のうち「近く」「クリックなし」を満たさない |
| **O1 ヘッダー read-only チップ（推奨候補）** | PatientContextHeader のペット名隣、またはその直下行。`useGetVitals` の今回カルテ末尾（または PO 指定の 1 件）を T/HR/RR/体重だけ描画。入力は現行 modal | はい（ペット名隣）。飼主名隣にするかは PO | 表示はクリック不要。編集は現行 FAB/modal | **表示省略**（チップに時刻を出さない）。modal 側の必須入力は残す | チップ 1 行。既存 `flex-wrap`（PatientContextHeader L103）に乗る。9 タブ行は触らない。保険カード・contextControls と幅競争 → 実機で折り返し確認 | **触らない**。GET の `recorded_at` はソートキーとして残し、画面に出さないだけ | 第二ストアなし。マスタ体重（H）と測定体重を併記するならラベルを分ける |
| **O2 sticky とタブの間の 1 行ストリップ** | StickyHeader で PatientContextHeader と `UnifiedTabsList` の間 | 名前の下。厳密な「すぐ近く」かは PO | 表示はクリック不要 | 表示省略可 | **タブ行の上に 1 行追加**。625 高さでは CHART-FIT の sticky 圧迫を増やす。9 タブ overflow は別問題のまま | 触らない | O1 より縦コストが大きい。幅はタブと独立 |
| **O3 ヘッダーで直接入力** | 名前隣に number input | はい | 表示はクリック不要。フォーカスは必要 | 入力 UI から時刻を消す案が出やすい | 9 タブは維持。入力欄は 1366 で折れる | create は `recorded_at` required のまま。**UI 省略時に now() や visit_date を埋めるのは捏造**。臨床時刻 UNKNOWN のため **この案の persist は停止** | 未保存中のヘッダー state が第二ストアになりやすい。採用しない |
| **O4 第 10 タブ** | `MEDICAL_RECORD_TABS` に「バイタル」 | いいえ | タブクリックが要る | 現状のテーブル | **9 タブ制約違反** | 触らない | 棄却 |
| **O5 本文常時 VitalsTab** | タブ領域の上または横にテーブル常設 | いいえ | テーブルは見える | 時刻列あり | 9 タブ + テーブルで 625 が不足。modal の `max-h-[85vh]` を本文に持ち込む | 触らない | 操作削減に反する。棄却 |

**設計として先に残すのは O1**（必要なら縦位置だけ O2）。O0 は現状。O3–O5 は制約または捏造リスクで落とす。

O1 でも入力は現行 VitalsTab。ヘッダーは **表示だけ**。保存成功後の invalidate でチップが更新される（同一 query key）。

## 時刻省略は display-only

todo-issue L128: 時刻入力/表示の省略案は記録時刻の根拠を保つ。本票の解釈:

| してよい（設計） | してはいけない（停止） |
| --- | --- |
| O1/O2 チップから時刻文字列を出さない | `vital_records.recorded_at` 列の削除・NULL 化 |
| VitalsTab の時刻列を後段で畳む UI（PO 後） | create から `recorded_at` を外す |
| ソート・監査・用量体重解決は `recorded_at` を使い続ける | 欠落時に `time.Now()` / 来院日時 / カルテ日付を **測定時刻として書く** |
| | 既存行の時刻を表示省略と同時に書き換える |

GORM `default:now()`（vital.go L22）は DB デフォルトであり、臨床測定時刻の合意ではない。HTTP create は required のまま。**臨床上の時刻扱いが決まるまで persist 変更は停止。**

## 9 タブ / 狭い画面

- タブ数は 9。StickyHeader の tab list は横スクロール。常時表示案は **タブ数を増やさない**（O4 棄却）。
- 1366×625 は CHART-FIT のローカル基準。実機 inner は UNKNOWN。O1 は `flex-wrap` で折返す。合格は「チップと 9 タブと保存 FAB が到達できる」ことであり、文字縮小やタブ削除ではない。
- VitalsModal `max-w-4xl`（896px 想定）+ テーブル `overflow-x-auto`（VitalsTabTable L71）。入力経路の狭画面は modal 内横スクロールで既に吸収。ヘッダー案がこれを壊さないこと。
- 会計タブでは FAB が消える（FormActions L53）。O1 チップを会計でも出すかは PO。出しても **入力導線が会計タブに無い**ことを案内に書く。

## 完了 / PO・停止

todo-issue L128 を本票に落とす:

- 常時表示する項目、最新値と今回測定の識別、空欄/未保存/複数測定の扱いを臨床 PO が確認する
- 全 9 タブ / 狭い画面でも入力・保存できる受入条件（入力は現行 modal + VitalsTab を再利用）
- 時刻の表示省略は可。保存時刻削除・測定時刻捏造はしない
- 臨床上の時刻扱いが決まるまでその変更は停止

本票は配置比較まで。O1 の実装・実機 1366×625 採取・PO 項目確定は後段。製品コードは変更していない。
