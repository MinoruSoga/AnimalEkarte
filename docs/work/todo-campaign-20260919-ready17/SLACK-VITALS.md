# SLACK-VITALS: 常時表示バイタルの配置案（VitalsTab vs ReadyPanels ヘッダー）

> Task migrated to Plane `EMR-190`. This file remains supporting acceptance/evidence material; use Plane for current status.

状態: **ヘッダー常時表示（今回カルテ末尾・時刻なし read-only チップ）は `873685b0b` で実装済み／実機受入・PO 項目確定・persist 変更は未実行**。初版は `aac697645` 時点で「比較 READY／未実装」として起筆されたが、同キャンペーンの `873685b0b`（"show microchip and latest vitals on the chart header"）で O1 相当が製品コードに入った。本票は `923bb99ba` 時点のコードへ再検証済み。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-VITALS`（L124–128、索引 L421）。保持する現場条件:

- 「名前のすぐ近く」「クリックなしで表示」「時間不要」（出典 929–937。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **表示/入力操作の削減**と**診療記録の時刻保持**を分離する
- 既存保存経路を再利用し、ヘッダーへ別の永続値を作らない（実装も GET 再利用で第二ストアなし）
- 「時間不要」を **保存時刻削除**や **測定時刻捏造**へ読み替えない
- 臨床上の時刻扱いが決まるまで persist 変更は停止（本票では **UNKNOWN**）

本票は [VitalsTab](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx) の既存 save path と、[MedicalRecordFormReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) ヘッダー配置を比較する。製品コード・テストは変更しない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-VITALS` の owned path および人間が読む調査票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-COMPLAINT.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。臨床時刻の意味は本票で決めない。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「名前のすぐ近く」が飼主名かペット名か | ヘッダーは飼主ボタン + ペット名 span が並び、その下の属性行にチップが出る（[PatientContextHeader](../../../frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx) L211–232、[MedicalRecordStickyHeader](../../../frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx) L102–103, L217） | **UNKNOWN**。実装はペット属性行を選んだ。飼主名側が正しいなら PO 判断で移動 |
| 常時表示する項目（T / HR / RR / 体重 / メモ） | 保存フィールドは 4 数値 + notes。チップは T/HR/RR/測定体重の 4 項目（メモなし） | **UNKNOWN**。PO 確認（todo-issue L128）。メモを出すかは未確認 |
| 最新値 vs 今回測定の臨床上の識別 | チップは**今回カルテ内 `recorded_at` 最大**の 1 件（[visit-vital-chips.ts](../../../frontend/src/features/medical-records/lib/visit-vital-chips.ts) L21–22）。ペット生涯の最新はこの経路に無い | **UNKNOWN**。コードは今回カルテ末尾を選択済み。臨床ラベルが「生涯最新」を要求するなら別 API が必要 |
| 空欄 / 未保存 / 複数測定の現場運用 | コードは空行・複数行・未保存フォームを区別。チップは 0 件・全項目 null で非表示、複数測定は末尾 1 件、入力中フォームは出さない | **UNKNOWN**。コード側の採用値は既に入っている。現場運用と不合なら PO 確認で見直し |
| 「時間不要」が入力省略か表示省略か | チップは時刻を出さない（表示省略は実装済み）。入力 UI は `datetime-local` 必須。BE create は `recorded_at` required | **UNKNOWN**。persist は停止のまま |
| 臨床上の記録時刻の意味（測定時刻 / 入力時刻 / 来院時刻） | GORM `default:now()` と FE 必須入力が共存 | **UNKNOWN**。now() を測定時刻として採用しない |
| 1366×625 実機 CSS viewport | ローカル基準は CHART-FIT と同じ 1366×625 CSS px。チップは既存 `flex-wrap` 行に乗る | **UNKNOWN**。実機 `innerWidth`/`innerHeight` と折返し確認は採取前 |
| フロント/API revision | 本票更新時 worktree HEAD `923bb99ba`（チップ実装 `873685b0b`） | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

現場の「バイタルが見えない / 時間が要らない」は次のどれでも同じ言葉になる。表示配置と persist 契約を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| H — ヘッダー体重 | [MedicalRecordStickyHeader](../../../frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx) L206 が `weight={selectedPet.weight}` を [PatientContextHeader](../../../frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx) L203–210 へ渡す。ペットマスタの体重 | **今回測定のバイタル体重**と同一視しない。ヘッダー体重は `vital_records` を読まない。チップ側の体重は「測定体重」ラベルで分離済み（L225–229） |
| V — 今回カルテのバイタル一覧 | `GET /v1/medical-records/:id/vitals`。[useGetVitals](../../../frontend/src/features/medical-records/api/vitals.ts) L30–38。BE は `medical_record_id` で絞り `recorded_at ASC`（[vital_repository.go](../../../backend/internal/medicalrecord/vital_repository.go) L46–50）。同一 GET の consumer は VitalsTab・StickyHeader チップ・治療タブ用量体重（[use-treatments-tab.ts](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) L56–61）の 3 箇所 | ペット生涯の「最新バイタル」API はこの経路に無い。一覧末尾・チップ値をペット最新と書かない |
| C — クリック導線 | 右下 FAB「バイタル記録」→ modal が**入力**の唯一経路。値はヘッダーの read-only チップにも出る（同 GET・時刻なし） | 「名前の近くに出ている」と FAB 到達を同一視しない。チップは編集不可で、正本は modal 内 VitalsTab |
| T — 時刻 | 追加/編集 UI は記録日時必須。create JSON も `recorded_at` required。チップは時刻を描画しない（display-only 省略） | 「時間不要」= 列削除 / default now() 捏造 / 既存行の時刻消去、と読まない |
| D — 入院日次バイタル | [DailyVitalsSection](../../../frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx) は入院タブ。別 `daily_record_id` | 一般診療カルテの SLACK-VITALS 再現対象にしない |

## 現行経路（保存は 1 系統・表示は 2 面）

ヘッダーへ **第二ストアを作らない**。入力・保存・再読込の正本は次だけ。表示はチップと modal の 2 面が同じ GET を読む。

0. **ヘッダー表示（クリック不要・read-only）:** StickyHeader L102 `useGetVitals(medicalRecordId ?? "", recordClinicId)` → `latestVisitVitalChips`（`recorded_at` 最大の 1 件を T/HR/RR/測定体重へ。0 件・全項目 null は null）→ PatientContextHeader `vitalsSummary`（L211–232）。時刻は出さない。保存 invalidate が同一 query key のためチップは保存後に自動更新される。
1. **入口（クリック必須・入力）:** [MedicalRecordFormReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) L347–364 の [MedicalRecordFloatingActions](../../../frontend/src/features/medical-records/components/MedicalRecordFormActions.tsx) `onVitalsClick` → `setIsVitalsOpen(true)`。ボタンは L68–86。会計タブでは FAB 全体が `return null`（L53）。見積書タブではバイタルボタンを出さない（L68）。新規カルテ / 確定済 / 死亡ペットは disabled または非表示。
2. **モーダル:** [MedicalRecordFormModals](../../../frontend/src/features/medical-records/components/MedicalRecordFormModals.tsx) L70–78。`!isNewRecord && recordId` のときだけ [VitalsModal](../../../frontend/src/features/medical-records/components/VitalsModal.tsx) をマウント。`DialogContent` は `max-w-4xl max-h-[85vh] overflow-y-auto`（L31）。中身は **同じ** [VitalsTab](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx)。
3. **一覧:** `useGetVitals(medicalRecordId)`。FE は `recorded_at` で再ソート ASC（VitalsTab L88–96）。0 件はテーブル空行（[VitalsTabTable.tsx](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTabTable.tsx) L75–80）。複数件は 1 行/記録。
4. **追加:** 空フォーム `EMPTY_VITALS_ADD_FORM.recorded_at === ""`（[vitals-tab-table-model.ts](../../../frontend/src/features/medical-records/lib/vitals-tab-table-model.ts) L18–26）。未入力は「記録日時は必須です」（VitalsTab L125–126）。未来日時拒否（L128–130）。T/HR/RR/体重のいずれか必須（L142–150）。`CreateVitalInput.recorded_at` は `jstDateTimeLocalToISOString`（L158–159）。`POST /v1/medical-records/:id/vitals`（vitals.ts L46–53）。
5. **BE create:** [createVitalRequest](../../../backend/internal/medicalrecord/vital_request.go) L13 `RecordedAt time.Time \`json:"recorded_at" binding:"required"\``。[VitalHandler.CreateVital](../../../backend/internal/medicalrecord/vital_handler.go) L83–97。サービスは `RecordedAt: input.RecordedAt` をそのまま書く（[vital_service.go](../../../backend/internal/medicalrecord/vital_service.go) L128）。モデル列は NOT NULL + `default:now()`（[vital.go](../../../backend/internal/model/vital.go) L22）だが、HTTP create はクライアント値を要求する。**欠落 JSON を now() で埋めて成功扱いにする経路は handler に無い。**
6. **更新:** PATCH の `recorded_at` は optional（vital_request.go L44）。UI 編集行は空なら「記録日時は必須です」（[VitalsTabRows.tsx](../../../frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx) L113–114）。表示列は `formatRecordedAt`（L42）。
7. **権限・死亡:** 作成/編集/削除は `medical-records` permission と `isPetDeceased`。死亡時は保存メッセージで拒否（VitalsTab L108–113, L191–193）。チップは死亡ペットでも表示される（読み取り専用・死後の記録参照は正常業務）。

invalidate は `queryKeys.medicalRecords.vitals(medicalRecordId)` prefix のみ（vitals.ts L55, L78, L98）。clinicId 付きキーを包含する prefix match のため mutation 側は clinicId なしで足りる（query-keys.ts L352–357 のコメント）。ペットマスタ `selectedPet.weight` は更新しない。

## 患者・カルテ切替（patient-switch）と空状態

`["vitals", medicalRecordId, clinicId?]` の query key（[query-keys.ts](../../../frontend/src/lib/query-keys.ts) L358–361）でカルテ単位に分離される。

| 状態 | コード | 観測される挙動 |
| --- | --- | --- |
| **カルテ切替**（同一フォームで recordId 変化） | `useParams` の `id` → `useGetVitals(newId)` が別 key で fetch | チップ・modal 一覧とも新カルテの値へ自動更新。旧カルテの値は残らない |
| **ペット切替** | 別ペット = 別ルート遷移。フォームは unmount | チップは新ペットの当該カルテ末尾を出す（ペット生涯最新ではない） |
| **新規カルテ（recordId なし）** | `useGetVitals("")` は `enabled: false`（vitals.ts L34）+ StickyHeader L103 の `medicalRecordId` guard | チップ非表示。VitalsModal 未マウント（FormModals L70）。FAB disabled |
| **空欄（0 件）** | `latestVisitVitalChips` は null（visit-vital-chips.ts L18–20） | チップは出さない（行を出さない採用）。modal 内は「バイタル記録がありません」行 |
| **全項目 null の 1 件** | chips 構築後に全 null なら null（L37–44） | 同上。メモのみの記録はチップに出ない |
| **部分空欄** | 項目ごと `!= null` で描画（PatientContextHeader L216–230） | ある項目だけ出す。`-` は出さない |
| **複数測定** | `recorded_at` 最大の 1 件のみ（visit-vital-chips.ts L21–22） | 末尾 1 件。全件ヘッダー固定は採っていない |
| **未保存（add form 中）** | `isAdding` は VitalsTab の local state。POST 成功まで GET に出ない | チップは入力中の値を出さない（第二 state なし・推奨どおり） |
| **modal を開いたままカルテ切替** | VitalsModal は `!isNewRecord && recordId` で mount され、VitalsTab の local state（`isAdding`/`addForm`/`editingId`/`showGraph`）は `medicalRecordId` に key 付けされていない | 一覧は新カルテへ追従するが、入力中ドラフトは新カルテ文脈へ持ち越され得る。通常のカルテ遷移はフォーム route 自体が変わり unmount するため edge case。必要なら `<VitalsTab key={medicalRecordId}>` で閉じられる（要 PO/実装判断） |

## 最新値 vs 今回測定（コード上の区別）

PO が臨床ラベルを決めるまで、UI 案では次の **3 ソースを別名のまま**出す。混ぜた「最新バイタル」ラベルをヘッダーに置かない。

| ソース | 何の値か | 取得 | 空のとき |
| --- | --- | --- | --- |
| **ペットマスタ体重** | 患者ヘッダーの Weight アイコン。名前の近くに既にある | `selectedPet.weight`（StickyHeader L206） | `weight` falsy ならアイコンごと出さない（PatientContextHeader L203） |
| **今回カルテの測定列** | この `medical_record_id` に紐づく `vital_records` | GET vitals。ASC なので **配列末尾が今回カルテ内の recorded_at 最大**。チップはこの末尾を描画 | 0 件は空状態・チップ非表示。未保存の add form は一覧に入らない |
| **ペット生涯の最新測定** | 他日カルテや入院バイタルを含む可能性 | **このカルテ経路に query が無い** | **UNKNOWN**。推測で GET を足さない |

チップは「今回カルテ末尾」を既に採用しているが、「今回のバイタル」の aria-label どまりで「最新」を名乗らない（PatientContextHeader L214）。現場がペット生涯の最新を期待するなら **別 API 設計が必要**で、本経路の再利用では実現できない。PO 確認まで生涯最新の表示を足さない。

## 配置オプション比較（ヘッダーは表示専用）

制約: 既存 9 タブを保つ（[medical-record-form-model.ts](../../../frontend/src/features/medical-records/routes/medical-record-form-model.ts) L4–14: 問診 / 診察/治療プラン / 治療 / 予防接種 / 定期健診 / 検査 / 画像 / 見積書 / 会計(医師確認)）。タブ list は StickyHeader L224–226 で `overflow-x-auto`。狭い画面の基準は [UAT-R2-CHART-FIT](../todo-campaign-20260918/UAT-R2-CHART-FIT.md)（1366×625 CSS px ローカル基準。実機 UNKNOWN）。

全案とも **書き込みは既存 VitalsTab POST/PATCH/DELETE**。ヘッダーは `useGetVitals`（同一 query key）の read か、現状維持。

| ID | 配置 | 名前の近く | クリックなし | 時刻 UI | 9 タブ / 1366×625 | recorded_at persist | 備考 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **O0 現状** | **ヘッダー read-only チップ（ペット属性行）+ 右下 FAB → VitalsModal → VitalsTab** | はい。チップはペット名直下の属性行（マスタ体重とは別ラベル「測定体重」） | 表示はクリック不要。編集は現行 FAB/modal | チップは時刻を出さない。modal はテーブル列「記録日時」+ 入力 `datetime-local` | チップは既存 `flex-wrap`（PatientContextHeader L169）に乗る。modal `max-w-4xl` は 1366 で横スクロールし得る | 変更なし | `873685b0b` で実装済み。残る実機確認は折返しとタブ到達性 |
| **O1 初版の推奨案** | PatientContextHeader のペット名隣または直下行に `useGetVitals` 末尾を描画 | はい | 同上 | 表示省略 | 同上 | 触らない | **O0 として実装済み**。以降の比較は「現状維持 vs 位置変更」に読み替える |
| **O2 sticky とタブの間の 1 行ストリップ** | StickyHeader で PatientContextHeader と `UnifiedTabsList` の間 | 名前の下。厳密な「すぐ近く」かは PO | 表示はクリック不要 | 表示省略可 | **タブ行の上に 1 行追加**。625 高さでは CHART-FIT の sticky 圧迫を増やす | 触らない | 現行チップより縦コストが大きい。採用理由が薄い |
| **O3 ヘッダーで直接入力** | 名前隣に number input | はい | 表示はクリック不要。フォーカスは必要 | 入力 UI から時刻を消す案が出やすい | 9 タブは維持。入力欄は 1366 で折れる | create は `recorded_at` required のまま。**UI 省略時に now() や visit_date を埋めるのは捏造**。臨床時刻 UNKNOWN のため **この案の persist は停止** | 未保存中のヘッダー state が第二ストアになりやすい。採用しない |
| **O4 第 10 タブ** | `MEDICAL_RECORD_TABS` に「バイタル」 | いいえ | タブクリックが要る | 現状のテーブル | **9 タブ制約違反** | 触らない | 棄却 |
| **O5 本文常時 VitalsTab** | タブ領域の上または横にテーブル常設 | いいえ | テーブルは見える | 時刻列あり | 9 タブ + テーブルで 625 が不足。modal の `max-h-[85vh]` を本文に持ち込む | 触らない | 操作削減に反する。棄却 |

**結論: O1 相当は実装済み（現行 O0）。** 残るのは配置の微調整（O2 への移動・飼主名側への移動）と PO 項目確定だけで、新規の表示実装は不要。O3–O5 は引き続き制約または捏造リスクで棄却。

チップの更新は保存成功後の invalidate（同一 query key）で行われる。ヘッダーは **表示だけ**。

## 時刻省略は display-only

todo-issue L128: 時刻入力/表示の省略案は記録時刻の根拠を保つ。本票の解釈:

| してよい（設計） | してはいけない（停止） |
| --- | --- |
| チップから時刻文字列を出さない（**実装済み**: vitalsSummary に時刻フィールドが無い） | `vital_records.recorded_at` 列の削除・NULL 化 |
| VitalsTab の時刻列を後段で畳む UI（PO 後） | create から `recorded_at` を外す |
| ソート・監査・用量体重解決（use-treatments-tab L56–61 の `resolveLatestVitalWeight`）は `recorded_at` を使い続ける | 欠落時に `time.Now()` / 来院日時 / カルテ日付を **測定時刻として書く** |
| | 既存行の時刻を表示省略と同時に書き換える |

GORM `default:now()`（vital.go L22）は DB デフォルトであり、臨床測定時刻の合意ではない。HTTP create は required のまま。**臨床上の時刻扱いが決まるまで persist 変更は停止。**

## 9 タブ / 狭い画面

- タブ数は 9。StickyHeader の tab list は横スクロール。常時表示案は **タブ数を増やさない**（O4 棄却）。
- 1366×625 は CHART-FIT のローカル基準。実機 inner は UNKNOWN。チップは `flex-wrap` の属性行に実装済みで、折返しは既存機構に従う。合格は「チップと 9 タブと保存 FAB が到達できる」ことであり、文字縮小やタブ削除ではない。
- VitalsModal `max-w-4xl`（896px 想定）+ テーブル `overflow-x-auto`（VitalsTabTable L71）。入力経路の狭画面は modal 内横スクロールで既に吸収。ヘッダー案がこれを壊さないこと。
- 会計タブでは FAB が消える（FormActions L53）が、チップは StickyHeader にあるため**会計タブでも表示される**。入力導線が会計タブに無い点は実装上そのまま。PO が会計でチップを消したいなら別判断。

## 完了 / PO・停止

todo-issue L128 を本票に落とす:

- 常時表示する項目、最新値と今回測定の識別、空欄/未保存/複数測定の扱いを臨床 PO が確認する（コードは 4 項目・今回カルテ末尾・空欄非表示・末尾 1 件を採用済み。現場と不合なら変更）
- 全 9 タブ / 狭い画面でも入力・保存できる受入条件（入力は現行 modal + VitalsTab を再利用。実機 1366×625 の折返し確認が残る）
- 時刻の表示省略は可（チップでは実施済み）。保存時刻削除・測定時刻捏造はしない
- 臨床上の時刻扱いが決まるまでその変更は停止

本票は配置比較と現行挙動の記録まで。実機 1366×625 採取・PO 項目確定・persist 変更は後段。本票の更新では製品コードを変更していない。
