# SLACK-COMPLAINT: 主訴区分の新規未選択 vs 既存クリア

状態: **保存経路調査 READY／実機保存・再読込 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-COMPLAINT`（L118–122、索引 L419）。保持する現場条件:

- 「主訴区分は空欄で入力」（出典 929–937。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない）
- **空欄許可は依頼済み**。可否を PO へ再質問しない。新しい仕様待ちに戻さない
- 納品区分・受入者の確認は別途。本票は保存経路の事実トレースだけにする

本票は [InterviewChiefComplaint](../../../frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx) → 問診 PATCH / 作成 request → inquiries 永続化 → GET 再読込 を、**新規の未選択（N）**と**既存選択の意図的解除（C）**に分離する。製品コード・テストは変更しない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-COMPLAINT` の owned path および人間が読む調査票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-STAFF-SELECT.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「空欄で入力」が新規カルテか、既存区分の解除か | UI は両方あり得る（後述）。報告一文では区別できない | **UNKNOWN**。採取前に片方へ決めない |
| 空欄時に残したい主訴本文の実例 | 定型テンプレ初期値はある（後述）。現場文面は別 | **UNKNOWN** |
| 既存区分を消す操作をしたか（クリア UI の有無はコードで分かる） | SearchableSelect に空 option / クリアボタンは無い | 現場がどう空欄にしたかは **UNKNOWN** |
| フロント/API revision | 本票作成時 worktree HEAD `aac697645`（`docs: TODOを回答済み要件から着手可能に整理`） | 再現セッションの SHA は **UNKNOWN**。採取時に固定する |

## 混ぜてはいけないケース

現場の「空欄で入力」は次のどれでも同じ言葉になる。空欄**許可の再質問**と、**保存経路の欠陥**と、**操作案内**を混ぜない。

### N — 新規未選択（一度も区分を選んでいない）

`chiefComplaintTypeId` の初期値は `null`（[use-medical-record-diagnosis-state.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-diagnosis-state.ts) L8）。空白許可は既にある。ここを仕様待ちに戻さない。

| ID | 条件（コード） | いまの経路 | 期待する分離 |
| --- | --- | --- | --- |
| N0 画面初期 | 問診タブ [InterviewChiefComplaint](../../../frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx) L94–103。`value={chiefComplaintTypeId ? String(chiefComplaintTypeId) : ""}`。placeholder「選択してください」（読込中は「読み込み中...」）。required / 空禁止の validation はコンポーネント内に無い | 区分はプレースホルダ。主訴詳細は独立 textarea（L128–139） | 区分未選択でも主訴本文を書ける。空欄可否を再確認しない |
| N1 カルテ自動作成 | [use-medical-record-auto-create.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-auto-create.ts) L182–190 の POST は `pet_id` / `owner_id` / `visit_date` / `visit_type` / `appointment_id` / `status` / `recommendation_reason` のみ。`chief_complaint` も `chief_complaint_type_id` も送らない | [CreateMedicalRecordRequest](../../../frontend/src/features/medical-records/api/types.ts) L28–29 は `chief_complaint_type_id?: number`（`null` を型に含まない）。省略はフィールド欠落 | 新規作成時点では区分は **omit**。JSON `null` ではない |
| N2 作成時 subrecord | BE [createMedicalRecordRequest](../../../backend/internal/medicalrecord/medical_record_request.go) L136 は `ChiefComplaintTypeID *uint64`（`omitempty` なし）。[toSubRecordsInput](../../../backend/internal/medicalrecord/medical_record_request.go) L253–265 がそのまま渡す。[hasInquirySubRecordInput](../../../backend/internal/medicalrecord/medical_record_subrecords.go) L128–131 は type / complaint / notes のどれかが non-nil のときだけ inquiry upsert。N1 の POST は三つとも無い | 自動作成では inquiry 行を作らない（入力が無いため） | 未選択を「作成失敗」と読まない。inquiry 未作成と type NULL 行を混同しない |
| N3 問診タブ保存 | 既存 `recordId` の「問診」保存は [use-medical-record-save-action.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) L236–246。`chief_complaint_type_id: snapshot.chiefComplaintTypeId` を **常に**渡す。未選択なら JS `null` | [UpdateInquiryRequest](../../../frontend/src/features/medical-records/api/inquiries.ts) L6–8 は `chief_complaint_type_id?: number \| null`。axios は `null` を JSON `null` にする（`undefined` だけ省略） | 新規未選択の **保存**は omit ではなく **JSON null**。N1 の作成 omit と混ぜない |
| N4 主訴本文 | 本文が `DEFAULT_CHIEF_COMPLAINT` と同一なら `chief_complaint` は `undefined`（省略）。違うときだけ文字列（save-action L237–240）。既定文は [use-medical-record-form-model.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-form-model.ts) L3–5 | 区分と本文は別フィールド。区分 null でも本文 PATCH は走る | 区分空欄で本文が消えるなら N ではなく保存ペイロード/再 hydrate を疑う。空欄禁止へ戻さない |
| N5 FK | [inquiry_service.Save](../../../backend/internal/medicalrecord/inquiry_service.go) L39–44 は `validateOwnedMasterFK(..., input.ChiefComplaintTypeID, ...)`。[validators.go](../../../backend/internal/medicalrecord/validators.go) L29–32 は sharedkernel へ委譲。nil ID は所有マスタ照合を要求しない（クロス医院 FK テストは **non-nil の他人 ID** を拒否する） | null は FK エラーにしない | 空欄保存の 4xx を「空欄は禁止」と読まない。実ステータス未採取なら N5 は BLOCKED |

### C — 既存選択の意図的解除

一度選んだ区分を空に戻す経路。空欄**許可**は既にある。クリア UI が無ければ操作案内、PATCH が残存するなら経路修正。どちらも「空欄にしてよいか」の仕様待ちではない。

| ID | 条件（コード） | いまの経路 | 期待する分離 |
| --- | --- | --- | --- |
| C0 クリア操作 | [SearchableSelect](../../../frontend/src/components/ui/searchable-select.tsx) は option の `onSelect` → `handleSelect(opt.value)`（L105–108, L118）。空 value の option もクリアボタンも無い。再選択も同じ ID を渡す | 画面操作だけでは `onValueChange("")` が起きない | **意図的クリアの UI が無い**。現場が空に見えても C 未実施のことがある。再現前に C を実装前提にしない |
| C1 空文字→null の受け口 | InterviewChiefComplaint L97: `onValueChange={(value) => setChiefComplaintTypeId(value ? Number(value) : null)}`。props は `number \| null`（L24–25） | Select が `""` を出せば state は null。現行 SearchableSelect は出さない | 受け口があることと、ユーザーが空にできることを同一視しない |
| C2 問診 PATCH | 保存は常に `chief_complaint_type_id: snapshot.chiefComplaintTypeId`（save-action L241）。state が null なら JSON `null` | 既存 ID が state に残っていればその数値が送られる | UI で消せないなら C2 の null は到達しない。到達したら C3 以降 |
| C3 hydrate の片方向 | [use-apply-medical-record.ts](../../../frontend/src/features/medical-records/hooks/use-apply-medical-record.ts) L47–49: `chiefComplaintTypeId != null` のときだけ setter。null / 省略は **既存 state を触らない**。BUG-406 follow-up は「値がある場合に載る」（[use-medical-record-form.test.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-form.test.ts) L282–305）。null へ戻すテストは無い | サーバが空でも、画面に残った選択値は消え得る | 再読込後も区分が見えることを「空欄禁止」と読まない。C3 は hydrate ギャップ |
| C4 確定済み | 問診フィールドは `!canEdit \|\| isFinalized` で disabled（InterviewChiefComplaint L43–44, L99）。PATCH も確定済みで拒否（inquiry_repository Conflict）。save-action は finalized を mutation 前に拒否 | 確定後クリアは対象外 | 確定カルテの空欄要望を draft 経路に混ぜない |

## omit 対 JSON null

層が違うと「送っていない」と「明示 null」が同じ Go 値になる。医院ポリシーではなくワイヤの事実。

| 層 | omit（キーなし） | JSON `null` | 数値 ID |
| --- | --- | --- | --- |
| 新規 POST（N1） | auto-create がキーを付けない。Create TS 型に `null` が無い | この経路では送っていない | この経路では送っていない |
| 問診 PATCH（N3/C2） | `undefined` だけ axios が落とす。save-action は type を常に `number \| null` で渡すので **省略しない** | 未選択なら `chief_complaint_type_id: null` | 選択中なら数値 |
| BE bind | [updateInquiryRequest](../../../backend/internal/medicalrecord/inquiry_request.go) L3–5 / create L136 は `*uint64` で `omitempty` なし。キー省略も JSON null も **nil ポインタ** | 同左 | non-nil |
| サービスコメント | [UpsertInquiryInput](../../../backend/internal/medicalrecord/inquiry_service.go) L11–17 は「nil = 未送信フィールド」。Save L47–52 は nil なら struct の type を触らない（ゼロ値の nil のまま） | bind 後は omit と null を区別できない | 代入する |
| 永続化 | [SaveByMedicalRecordID](../../../backend/internal/medicalrecord/inquiry_repository.go) L74–84 は `map[string]any` に `"chief_complaint_type_id": inquiry.ChiefComplaintTypeID` を **毎回含める**。既存行を読んでマージしない。ゼロ値 struct の nil ポインタがマップに入る | omit と null は同じ nil ポインタ | ポインタ付き ID |
| 作成 subrecord | type が nil なら inquiry に type を書かない（[medical_record_subrecords.go](../../../backend/internal/medicalrecord/medical_record_subrecords.go) L36–38）。inquiry upsert 自体は N2 のとおり入力が無いとスキップ | 作成 JSON null も bind 後 nil → 同じ | 入力ありなら upsert |
| GET 応答 | [InquirySummaryResponse](../../../backend/internal/medicalrecord/medical_record_response.go) L45 と [inquiryResponse](../../../backend/internal/medicalrecord/inquiry_response.go) L13 は `json:"chief_complaint_type_id,omitempty"`。nil は **キー省略**。JSON null は出さない | 応答に null は出ない | 数値 |
| FE 再読込 | [transformMedicalRecord](../../../frontend/src/lib/transforms/medical-record.ts) L52–53: `record.inquiry?.chief_complaint_type_id ?? null`。キー省略は JS `undefined` → `null` | 応答に現れない | 数値。BUG-013 テスト（[transforms.test.ts](../../../frontend/src/features/medical-records/api/transforms.test.ts) L144–149） |

要点:

1. **新規作成（N1）の未選択は omit。問診保存（N3）の未選択は JSON null。** 同じ「空」でも HTTP 形が違う。
2. **Go の PATCH bind は omit と null を区別しない**（どちらも `*uint64` nil）。
3. サービスは nil を「未送信」と書くが、repository は既存 type を読まずマップへ type 列を載せる。GORM `Updates(map)` が SQL NULL になるかは **本票では実 DB 未実行**（inquiry_repository_test は type 列の NULL 化を断言していない）。persist の NULL 化は **UNKNOWN**。失敗したら C の最小修正対象であり、空欄許可の再質問ではない。
4. GET は nil を omit する。FE は omit を `null` に正規化する。C3 の hydrate は `null` を state に書き戻さない。

## 現行経路（UI → request → 保存 → 再読込）

1. **入口:** [MedicalRecordFormReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) L251 → 問診タブ [MedicalRecordInterview](../../../frontend/src/features/medical-records/components/MedicalRecordInterview.tsx) L88 → InterviewChiefComplaint。区分は SearchableSelect。マスタは [useGetChiefComplaintTypes](../../../frontend/src/features/medical-records/api/get-chief-complaint-types.ts)。
2. **新規:** ペット選択後 auto-create が POST `/v1/medical-records`（N1）。navigate で detail。問診 state は診断 state の初期 `null` のまま（inquiry が無ければ transform も `?? null`）。
3. **問診保存:** タブ「問診」の formAction → PATCH `/v1/medical-records/:id/inquiries`（[inquiry_handler.go](../../../backend/internal/medicalrecord/inquiry_handler.go) L22–43）→ `updateInquiryRequest.toServiceInput` → `inquiryService.Save` → `SaveByMedicalRecordID`（親カルテ FOR UPDATE、draft のみ）。
4. **成功後:** `useUpdateInquiry` が `queryKeys.medicalRecords.detail(recordId)` を invalidate（inquiries.ts L17–19）。再 GET の embed は InquirySummary。transform → `useApplyMedicalRecord`。
5. **主訴本文:** 区分と独立。空欄区分でも本文を失わないことが受入（todo-issue L122）。本文が DEFAULT のままだと PATCH から `chief_complaint` キーが落ちる（省略）。type キーは落ちない。

## 再現プロトコル（実機前に固定する軸）

同一医院・同一権限・draft カルテで、N と C を別レコード（または C の前に一度区分を保存した同一レコード）で取る。空欄許可を質問しない。

| 手順 | 固定すること | PASS の見方 | やってはいけないこと |
| --- | --- | --- | --- |
| N 新規未選択 | 区分はプレースホルダのまま。主訴本文を DEFAULT から変更して保存 → 再読込 | 本文が残り、区分は空のまま。4xx にしない | 空欄禁止バリデーションを足す。PO に可否を聞く |
| C 既存解除 | 一度区分を選んで保存・再読込で ID が戻ることを確認してから、空に戻して保存 | **C0 で空に戻せない**なら、クリア操作なしと記録し、操作案内候補。無理に仕様追加しない。戻せたのに再読込で残るなら C3/persist を疑う | クリア UI が無いのに「空欄が禁止されている」と結論する |
| omit vs null | 開発者ツールで N1 POST と N3 PATCH の JSON を保存 | N1 に type キー無し、N3 未選択は `null` | 片方だけ見て「常に omit」または「常に null」と書く |

対象環境不足は該当行 BLOCKED。根拠がない一括「必須化」や「全医院で空にできない」断定を先行実装しない。

## 完了 / PO・停止

todo-issue L122 の完了条件を本票に落とす:

- 区分未設定で主訴本文を失わず保存できる（N）。失敗した画面/経路だけ最小修正
- 既存区分の意図的解除も受入に含める（C）。クリア UI が無ければ案内。PATCH/hydrate が残存するならその経路だけ
- 該当しなければ操作案内へ
- 納品区分・受入者確認は別途
- **空欄許可そのものを新しい仕様待ちに戻さない**

臨床的な「空欄にしてよいか」は依頼済みとして閉じる。医院ごとの必須運用ルールはコードに無く **UNKNOWN**。本票で作らない。
