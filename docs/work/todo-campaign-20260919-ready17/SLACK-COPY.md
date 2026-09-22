# SLACK-COPY: 同一医院/ペットの前回記録を選択コピー（preview / 取消、請求済み ID は流用しない）

状態: **転記削減の設計 READY／複写の保存意味は PO 未確定／製品実装・実機受入 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-COPY`（L154–158、索引 L430）。保持する現場条件:

- 前回治療 / 主訴をコピーしたい（出典 987–994。現行 `todo-issue.md` は要約のみ。原文行は本票では再掲しない。同一出典ブロックは [SLACK-PLAN-MANUAL](./SLACK-PLAN-MANUAL.md) と共有し、プラン手入力と複写を混ぜない）
- **転記削減**が目的。紙カルテの「前回を全部写す」をデジタル化しない
- [問診履歴](../../../frontend/src/features/medical-records/components/InterviewHistory.tsx) は過去詳細へのリンクであり、コピー実装ではない
- 主訴の定型文挿入や接種履歴の複製は、前回治療/主訴コピーが実装済みの根拠にしない
- 保存前 preview と取消で元記録・現入力を失わない
- 旧 `treatment.id` / 請求済み `billing_items.treatment_id` を新しい請求として流用しない（無音の再請求禁止）
- **追記 vs 置換は UNKNOWN**。本票で採用しない

本票は InterviewHistory → 治療 API → 主訴保存をトレースし、同一医院・同一ペットの前回記録を**選択的に**取り込む案だけを置く。製品コード・テストは変更しない。

呼び出し行: **無い。** 本ファイルは製品コードから import されない。キャンペーン unit `SLACK-COPY` の owned path および人間が読む調査票である（sibling `SLACK-COMPLAINT.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` に本 unit の票は無く、`todo-issue.md` L154–158 は出典要約であり本票の代替ではない。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「前回」が最新確定か、任意の過去か、下書きも含むか | 履歴 GET は status 未指定・最大 50 件（後述）。「前回」ラベルは無い | **UNKNOWN**。曖昧な「前回」自動選択は停止 |
| コピー対象（主訴本文 / 主訴区分 / 治療行 / 方針 notes） | 経路は別（後述） | **UNKNOWN**。臨床 PO が対象項目を決めるまで一括コピーしない |
| 追記か置換か | 定型文挿入は **置換**。inquiry PATCH は送った文字列で上書き。治療 Create は常に新規行 | 複写の保存意味は **UNKNOWN**（todo-issue L156, L158） |
| 数量・単価・日付・薬剤の再確認手順 | 治療 Create は qty/price/medicine を新規検証する。日付列は treatments に無い | **UNKNOWN**。preview なしの自動再実施は停止 |
| 請求済みを再実施したいか | 未請求抽出は `billing_items.treatment_id` の NOT EXISTS（後述） | **UNKNOWN**。請求済み ID 流用は禁止。再実施するなら新規行 |
| フロント/API revision | 本票作成時 worktree HEAD `aac697645`（`docs: TODOを回答済み要件から着手可能に整理`） | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

現場の「前回をコピー」は次のどれでも同じ言葉になる。詳細リンク・定型文・会計未請求取込・複写を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| H — 問診抜粋リンク | [InterviewHistory.tsx](../../../frontend/src/features/medical-records/components/InterviewHistory.tsx) L76–80 は各行を `Link` `paths.medicalRecords.detail.getHref(item.id)`。見出しは「問診抜粋」「全文は詳細で確認できます」（L51–52） | コピー済みと読まない。テストは引用ボタン不在を固定（[InterviewHistory.test.tsx](../../../frontend/src/features/medical-records/components/InterviewHistory.test.tsx) L133–145, L173） |
| M — 空履歴のモック | [MedicalRecordInterview.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordInterview.tsx) L78–79: `historyItems` が空なら `DEFAULT_HISTORY_ITEMS`（L34–58、架空日付 2022 と id `"1"`/`"2"`/`"3"`） | モック行をコピー元にしない。空なら空と出す |
| T — 定型文挿入 | InterviewChiefComplaint L106–124「定型文挿入」。親 L71–76 `setChiefComplaint(text)` で **全文置換**。文言はモジュール定数 `INTERVIEW_TEMPLATES`（L24–32） | 前回カルテの主訴コピーと同一視しない |
| C — 主訴保存 | 問診タブ PATCH [use-medical-record-save-action.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) L236–246 → [inquiry_service.Save](../../../backend/internal/medicalrecord/inquiry_service.go) L37–65。`chief_complaint` はポインタがあれば文字列代入 | 保存＝前回複写ではない。現行は現カルテ 1 行の upsert |
| R — 治療新規 POST | [treatments.ts](../../../frontend/src/features/medical-records/api/treatments.ts) L47–68 `POST /v1/medical-records/:id/treatments`。[CreateTreatmentInput](../../../frontend/src/features/medical-records/types/index.ts) L52–73 に `id` は無い。[createTreatmentRequest](../../../backend/internal/medicalrecord/treatment_request.go) L7–26 も `id` を bind しない | 旧行 PATCH / 旧 ID の使い回しではない。Create は常に新規 PK |
| U — 未請求取込 | [FindUnbilledByPetID](../../../backend/internal/medicalrecord/treatment_repository.go) L107–114 は同一 clinic/pet の確定確認済みカルテ治療のうち、非キャンセル `billing_items.treatment_id` が無い行。[treatmentToUnbilledBillingItem](../../../backend/internal/billing/billing_item_unbilled.go) L124–142 は候補の `TreatmentID` に **既存 treatment.id** を載せる | 会計の未請求取込と「前回を今日のカルテへ写す」を同一視しない。未請求取込は ID 参照が仕事。コピーは新規行が仕事 |
| P — 飼主レポート履歴 | [ListPetTreatmentHistory](../../../backend/internal/medicalrecord/treatment_handler.go) L78–114 `GET /pets/:id/treatment-history`。clinic+pet JOIN。#158/#159 読み取り | レポート閲覧をコピー API と読まない。参照ソース候補ではある |
| V — 接種複製 | todo-issue L157 が明示。本票は治療/主訴。vaccination の `uq_billing_items_vaccination_lifetime` は別契約 | 接種 unique を治療コピーの根拠にしない |

## 現行経路

### 1. InterviewHistory は過去詳細リンク（コピーではない）

1. [MedicalRecordFormReadyPanels.tsx](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) L44 が `useGetPetMedicalHistory(selectedPet.id, recordId)`。L253 で問診パネルへ渡す。
2. [useGetPetMedicalHistory](../../../frontend/src/hooks/use-medical-records.ts) L113–137: `GET /v1/medical-records?limit=50&page=1&pet_id=`。`clinic_ids` は付けない（axios の選択医院ヘッダに依存）。現 `recordId` を除外。`status` フィルタ無し（確定も下書きも混ざる）。
3. 変換 L100–110: `inquiry.chief_complaint` だけを title/content にする。treatments は載せない。feature 側 [transforms.ts](../../../frontend/src/features/medical-records/api/transforms.ts) L12–24 も同形。
4. BE 一覧は [ListMedicalRecords](../../../backend/internal/medicalrecord/medical_record_handler.go) L25–63。`ResolveListClinicIDsForPermission`（所属かつ `medical-records:view`）。#86 横断一覧になり得る。**コピー元は現カルテの `clinic_id` と同一医院に限定する**（治療 POST は [ExtractClinicID](../../../backend/internal/medicalrecord/treatment_handler.go) L139。treatments.ts L18–22 は record 自身の clinicId を `X-Clinic-ID` にする）。
5. UI: InterviewHistory L76–99 は `Link` のみ。クリックは `/medical-records/:id` へ遷移。inquiry / treatments を現カルテへ書き込まない。
6. **ナビは現入力を失い得る。** sibling SLACK-DETAILS の U1（同一タブ navigate）と同じ。コピー UI をこの Link の上に載せてはならない。

### 2. 治療 API（新規行。クライアントは ID を送れない）

1. タブ追加は [use-treatments-tab.ts](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) L45 が `useCreateTreatment`。マスタ選択ペイロード [treatments-tab-model.ts](../../../frontend/src/features/medical-records/lib/treatments-tab-model.ts) L21–38: `item_type` / `content` / `unit_price` / `quantity` / `medicine_id`（number）。`id` も `inventory_id` も付けない。
2. POST body に `id` が無い。[createTreatmentInTx](../../../backend/internal/medicalrecord/treatment_service_tx.go) L25–73 が struct を組み `treatmentRepo.Create`。PK は DB 採番。Location は新規 ID（handler L171）。
3. 所有は親カルテ経由。`lockDraftMedicalRecord`（tx.go L20–22）。確定済みへは追加できない。
4. マスタ FK は caller clinic 所属を検証（[treatment_service.go](../../../backend/internal/medicalrecord/treatment_service.go) L195–200）。数量 `<=0` 拒否（L179–181）。投与量は Create 時に再評価（tx.go L47–67）。**過去 qty を無検証で再実施しない。**
5. `InventoryID` があるときだけ在庫減算（tx.go L75–80）。前回行の `inventory_id` をコピーすると在庫が二重減算される。コピー案は **inventory_id を送らない**（現行マスタ追加と同じ）。
6. PATCH/DELETE は既存 ID を現カルテ配下で触る経路。コピー適用に使わない。
7. treatments に実施日列は無い。日付の正本は親 `medical_records.date`。コピーで「前回の日付」を治療行 ID に載せない。

### 3. 請求との境界（無音再請求をここで止める）

1. `billing_items.treatment_id` は FK + 部分 INDEX（[001_init.sql](../../../backend/migrations/001_init.sql) L1925, L2169）。**UNIQUE ではない。** 生涯 UNIQUE があるのは vaccination / exam（同ファイル L4061, L6199）だけ。
2. 会計 Create は `treatment_id` を受け取る（[billing_item_request.go](../../../backend/internal/billing/billing_item_request.go) L49）。未請求候補は既存 ID を `TreatmentID` に載せる（unbilled.go L138）。候補表示では `BillingItem.ID = t.ID`（L128）——表示用であり、コピーでこの ID を新しい billing 行の PK/FK にしてはならない。
3. コピー適用後の会計は、**今日の新規 treatment.id** だけを未請求抽出に乗せる。旧 ID を `createBillingItemRequest.treatment_id` に入れない。
4. 旧行が既に非キャンセル請求に載っていても INDEX だけでは二重請求を止めない。禁止は設計契約であり、DB unique を根拠にしない。

### 4. 主訴保存（定型文 ≠ 前回コピー）

1. 区分 + 本文 + 定型ボタン: [InterviewChiefComplaint.tsx](../../../frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx)。preview/取消のコピー UI は無い。
2. 定型挿入は **置換**（MedicalRecordInterview L71–76）。追記 API は無い。
3. 問診保存は PATCH inquiries。DEFAULT と同一なら `chief_complaint` キー省略（save-action L237–240）。違うときだけ文字列。区分は常に `number | null`。
4. BE Save は nil でなければ `inquiry.ChiefComplaint = *input.ChiefComplaint`（inquiry_service.go L53–55）。送った値がそのカルテの本文になる。前回レコードを読まない。
5. 確定済み / 権限なし / 死亡ペットは保存拒否（save-action L220–232、InterviewChiefComplaint L43–44）。コピー適用も同じゲート。

## 選択コピー案（実装しない。契約だけ）

目的は転記工程の削除（product philosophy ②）。新しい治療グラフや「前回ボタンで全件確定」は作らない。

### 参照（読み取り）

| 対象 | 使う既存読み | 使わない |
| --- | --- | --- |
| コピー元カルテの列挙 | 現カルテと同じ `clinic_id` + 同じ `pet_id`。日付・record_no・status を表示。ユーザーが **1 件を選ぶ** | 「前回」自動。空履歴時の DEFAULT_HISTORY_ITEMS。他院一覧（#86 横断） |
| 主訴本文/区分 | 選んだレコードの GET 詳細の inquiry | InterviewHistory の 2 行抜粋だけを正本にしない |
| 治療行 | `GET /v1/medical-records/:sourceId/treatments`（record clinic の `X-Clinic-ID`）または pet history を **候補表示**に限る | 未請求取込 API の `TreatmentID` を今日の POST に載せる |

ソース行には `source_record_id` / `source_date` / `source_treatment_id`（追跡用表示）を出す。適用ペイロードには `source_treatment_id` を **書かない**。

### Preview / 取消（必須）

確認ダイアログだけを安全性の根拠にしない。

| 操作 | 契約 |
| --- | --- |
| 開く | 現カルテルート上の overlay（SLACK-DETAILS U0）。`InterviewHistory` の `Link` を発火しない。現入力は unmount しない |
| Preview | 取り込む項目のチェックリスト。主訴本文、区分、治療ごと content / item_type / quantity / unit_price / medicine_id / procedure_id / consultation_id。請求済みフラグは表示のみ（未請求 NOT EXISTS を参照してよい）。数量・単価・薬剤はここで再確認する |
| 既定 | 全選択しない。請求済み行は **未チェック**。投薬行は未チェックを推奨（todo-issue: 過去の投薬判断を未確認で再実施しない） |
| 取消 | overlay を閉じる。inquiry state・treatments query・元記録は無変更 |
| 適用前 | 現カルテが draft・同一 clinic/pet・`medical-records:create`（治療）/ 問診 `canEdit`。確定済み・死亡は拒否 |

### 適用（PO が対象を決めた後）

| 対象 | ワイヤ | 禁止 |
| --- | --- | --- |
| 治療（選択行） | 現カルテへ **1 行ずつ** `POST .../treatments`。body は CreateTreatmentInput の値コピー。`status` は未指定（service 既定 `pending`）。`inventory_id` 省略。`id` なし | ソース `treatment.id` を URL や body に載せる。ソース行 PATCH。billing `treatment_id` に旧 ID。`status=completed` の黙従。在庫 ID のコピー |
| 主訴 | PO が置換/追記を決めるまで **persist しない**。Preview のテキスト比較だけ | 定型文置換を「前回コピー」と呼ぶ。inquiry をソース record へ PATCH |
| 会計 | 新規 treatment PK が未請求抽出に載るのを待つ | コピーと同時に billing_items を旧 ID で Create |

置換 vs 追記（主訴本文、既存治療行とのマージ）は **UNKNOWN**。競合（今日すでに手入力がある）の勝者も PO。本票は「preview で両方見せ、取消可能」まで。黙って上書きしない。

## 完了 / PO・停止

todo-issue L158 を本票に落とす:

- コピー元の記録 ID と日付が preview で追える
- 対象項目（主訴本文/区分/治療行のどれか）を臨床 PO が決める
- 追記 vs 置換、現入力との競合を臨床 PO が決める。決まるまで persist しない
- 数量/単価/薬剤（日付は親カルテ）を preview で再確認する。投薬の自動再実施は停止
- 保存前 preview と取消で元記録・現入力を失わない（overlay。詳細 Link を使わない）
- 旧 ID / 請求済み状態を新しい charges に流用しない。再実施は新規 POST
- 「前回」の曖昧選択、他院記録、モック履歴、接種 unique の流用は停止
- 該当しなければ操作案内（問診抜粋から詳細を開いて手転記）へ。詳細ナビは未保存喪失リスクを案内する

臨床判断（何を今日やるか）をコピーが代行しない。医院ごとの「いつも前回全部」運用はコードに無く **UNKNOWN**。本票で作らない。
