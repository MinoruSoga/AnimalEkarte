# SLACK-MANUAL-URINE: 手入力尿試験紙の保存・再読込・表示設計

状態: **手入力経路の受入設計 READY／合成検証 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-MANUAL-URINE`（9月9日返信 212–294）。敷島/猫は尿試験紙を**目視**、城東/八王子は**機器測定**と明示されている。本票は手動結果の UI → request → 永続化 → 再読込 → カルテ/検歴表示を現行コードへ対応づける。機器受信へ統合しない。臨床的な陽性/陰性の意味・院別カットオフは推定しない。

機器受信そのものは todo-issue の `### SLACK-LAB`（同出典ブロック）へ分離する。実装済み受信コードを手入力の完了根拠にしない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-MANUAL-URINE` の owned path および人間が読む調査票である。

## 医院事実（コード外・UNKNOWN）

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 敷島/猫の試験紙銘柄・項目名・目視凡例（+/- の臨床意味） | なし | **UNKNOWN**。写真未確認。本票で凡例を作らない |
| 城東の尿機器 | 城東 COM3 アークレイ PU-4010 は decoder-only。agent 既定 slot / 運用 support 対象外（[LAB_DEVICE_CONNECTIVITY](../../ops/deploy/LAB_DEVICE_CONNECTIVITY.md)） | 現場 7E1/8E1 差は未 review。実機受信状態は SLACK-LAB の UNKNOWN |
| 八王子の尿機器型番 | なし（写真未確認、SLACK-LAB と同旨） | **UNKNOWN** |
| 院ごとの `exam_types` 名称が「尿検査」か、手入力用マスタが既にあるか | テスト/シードに「尿検査」文字列はあるが、対象医院の本番マスタではない | **UNKNOWN**。医院承認の期待表が来るまで項目を固定しない |
| 目視結果を機器受信 exam へ上書きしてよいか | 現行は別 write 経路。同一 exam 行へのマージ API はない | 混同・上書き禁止（要件）。統合実装を本票で提案しない |

## 経路を混ぜない（origin / method）

測定方法と原本由来の区別は、新しい `origin` / `method` 列を発明せず、現行フィールドで保持する。

| 経路 | 入口 | 永続化の印 | 値の編集 |
| --- | --- | --- | --- |
| **手入力（本票）** | [ExaminationForm](../../../frontend/src/features/examinations/routes/ExaminationForm.tsx) + [ExaminationFormFields](../../../frontend/src/features/examinations/components/ExaminationFormFields.tsx) + [ExamItemsTable](../../../frontend/src/features/examinations/components/ExamItemsTable.tsx) | `exams.job_id` は NULL。コメント: 「手動作成の exam は NULL」（[examination_record.go](../../../backend/internal/model/examination_record.go) L35–37）。Create は JobID をセットしない（[examination_service.go](../../../backend/internal/medicalrecord/examination_service.go) L235–245） | 結果値はスタッフが `inspection_value` に入力。`status` / `is_abnormal` は request で受け付けない |
| **機器受信（LAB、本票対象外）** | `/lab-device` [LabDeviceBoard](../../../frontend/src/features/lab-device/routes/LabDeviceBoard.tsx)、[lab_device_receive_service.go](../../../backend/internal/medicalrecord/lab_device_receive_service.go)、ADR-007 | `LabExamPersistInput.JobID` を `exams.job_id` に保持（[lab_import_examination_service.go](../../../backend/internal/medicalrecord/lab_import_examination_service.go) L17–26, L65）。定性値は InspectionValue に**そのまま**格納（L30–31） | 受信値エディタは ADR-007 でやらない。attach は `POST /lab-imports/:job_id/attach`。値は編集しない（[13-examinations-form.md](../../spec/screens/13-examinations-form.md) §1.3） |

親 GET の [examinationResponse](../../../backend/internal/medicalrecord/examination_response.go) L18–37 / `toExaminationResponse` L39–64 は `job_id` も items も載せない。FE `transformExamination` も `job_id` を持たない（[examination.ts](../../../frontend/src/lib/transforms/examination.ts) L38–55）。カルテ側グループ表示の識別は `machine` バッジのみ（[ExaminationGroup](../../../frontend/src/features/medical-records/components/ExaminationGroup.tsx) L57）。手入力フォームに機器名欄は無い（後述）。したがって **画面上の origin 表示は現行ギャップ**であり、HTTP 親レスポンスだけでは機器結果と手入力を区別できない。区別の正本は DB の `job_id` と write owner（手動 Create vs LabImport persist）とする。

`machine` は create/update JSON にあるが、ExaminationFormFields は入力しない。新規手入力では空文字が既定（model `default:''`）。機器経路は機器ヒントを入れうる。空の `machine` を「目視」と解釈しない。逆に `machine` が入っていても `job_id` 無しなら手入力の可能性を残す。

## UI → request → 永続化 → 再読込 → 表示

### 1. 入力 UI

[ExaminationFormFields.tsx](../../../frontend/src/features/examinations/components/ExaminationFormFields.tsx) はヘッダのみ。尿試験紙専用ウィジェットはない。

| UI | フィールド | 備考 |
| --- | --- | --- |
| 検査種別 | `testTypeId` / `testType` | マスタ `examination`。尿専用の固定 ID はコードに無い |
| 担当医 | `doctorId` | 候補は [ExaminationForm.tsx](../../../frontend/src/features/examinations/routes/ExaminationForm.tsx) が `staffType === "doctor" && isActive` で絞る。FormFields は渡された `staffList` を出すだけ |
| 検査日 | `date` | JST 日付 |
| ステータス | `status` 日本語ラベル | 依頼中 / 検査中 / 結果入力済み / 完了 / 確定 |
| 備考・所見 | `resultSummary` | 自由文。試験紙凡例の置き場ではない |
| 機器名 | **UI なし** | `formData.machine` は hook が PATCH/POST に載せるが、本フォームは変更しない |

項目行は同画面の [ExamItemsTable](../../../frontend/src/features/examinations/components/ExamItemsTable.tsx):

| 列 | 編集 | 送信 |
| --- | --- | --- |
| 項目名 | テンプレ行（`examTypeFieldId` あり）は固定。手動追加行のみ編集 | `name` required |
| 結果値 | text、`inputMode="decimal"` | `inspection_value`。定性記号も文字列として入る |
| 単位 | 表示のみ | `unit`（テンプレはマスタ、手動追加は空） |
| 基準値 | 表示のみ | `reference_value`（テンプレは `normalValue` をコピー） |
| 判定 | 表示のみ | request に `status` / `is_abnormal` を送らない（ExamItemsTable コメント、examination_request.go L161–162） |

テンプレ展開: 検査種別変更 → `GET /v1/masters/examination-types/:id` → [buildRowsFromTemplate](../../../frontend/src/features/examinations/hooks/use-examination-form-model.ts) L39–49。単位・表示用基準はフィールドマスタ。判定用 min/max は FE で再計算しない。

手動追加: [addManualItem](../../../frontend/src/features/examinations/hooks/use-examination-form-items.ts) L181–199。`examTypeFieldId` なし、`unit` / `normalValue` / `referenceValue` は空。項目名空かつ結果値ありは保存拒否（[validateExaminationSave](../../../frontend/src/features/examinations/hooks/use-examination-form-model.ts) L135–139）。

### 2. Request（examination_request.go）

[upsertExamItemRequest](../../../backend/internal/medicalrecord/examination_request.go) L161–171:

`exam_type_field_id`, `name` (required), `inspection_value`, `normal_value`, `unit`, `reference_value`, `sort_order`

Create 親 bind（L91–102）: `medical_record_id`, `pet_id`, `exam_type_id`, `doctor_id`, `date`, `result_summary`, `machine`, `status`, `items`。`job_id` は bind しない。

FE 新規 [buildCreateExaminationRequest](../../../frontend/src/features/examinations/hooks/use-examination-form-model.ts) L201–217 は `status` を送らない。UI で「結果入力済み」等を選んでも Create は BE 既定 `pending`（examination_service.go L225–228）。ステータス変更は編集 PATCH 側（同ファイル L180–198 の `status`）。新規手入力の受入では「画面のステータス＝保存後ステータス」と見なさない。

Handler: `POST /api/v1/examinations` [CreateExamination](../../../backend/internal/medicalrecord/examination_handler.go) L119–144。Update は PATCH 親 + items。項目単独は `PUT /examinations/:id/items`。

FE 変換: [rowsToRequest](../../../frontend/src/features/examinations/hooks/use-examination-form-model.ts) L52–64。空名行は送らない。Create: [buildCreateExaminationRequest](../../../frontend/src/features/examinations/hooks/use-examination-form-model.ts) L201–217。`POST /v1/examinations`（[create-examination.ts](../../../frontend/src/features/examinations/api/create-examination.ts)）。

### 3. 永続化と判定

Create は `exams` 行を作り、items があれば ReplaceItems。基準値はペット種別に `exam_type_field_id` がある行だけ解決する（[buildExamResultsFromInputs](../../../backend/internal/medicalrecord/examination_items.go) L230–272）。手動行（field ID nil）は numeric/qualitative bounds なし → `assessExamResult` は **未判定**（`status=normal`, `is_abnormal=false`, `isAssessed=false`）。

定性比較の**コード上のトークン**（臨床意味ではない）は [exam_result_assessment.go](../../../backend/internal/medicalrecord/exam_result_assessment.go) L12–18:

`(-)` / `(±)` / `(+)` / `(++)` / `(+++)`

正規化は全角括弧と空白のみ（L131–144）。`"陰性"` / `"＋"` / `"(++++)"` は **認識しない**（[exam_result_assessment_test.go](../../../backend/internal/medicalrecord/exam_result_assessment_test.go) L64–67）。認識できない値は未判定。数値と定性 bounds の共存は fail-closed 未判定。

**本票はこれらのトークンを医院の試験紙凡例として採用しない。** マスタに qualitative min/max が医院承認で載った場合の比較実装の事実だけを記す。敷島目視の「+」がコードの `(+)` かは **UNKNOWN**。

機器 persist は同じ `exam_results` テーブルへ書くが JobID 付き。定性は ParseFloat 不能でも InspectionValue に保持（lab_import_examination_service.go L30–31, L78）。手入力経路の ReplaceItems と同一テーブルでも、**同一ジョブへの手入力マージはしない**。

### 4. 再読込

保存成功後、フォームは一覧へ遷移する（[ExaminationForm.tsx](../../../frontend/src/features/examinations/routes/ExaminationForm.tsx) L145–149）。同一画面に留まって再 GET しない。

再表示の読み込み:

| 画面 | API | items |
| --- | --- | --- |
| 検査編集を開き直す | `GET /v1/examinations/:id` 親（[get-examination.ts](../../../frontend/src/features/examinations/api/get-examination.ts)）+ `GET /v1/examinations/:id/items` | 項目は items エンドポイント。親 GET の JSON に items を期待しない |
| カルテ検査タブ | `GET /v1/examinations?pet_id=&medical_record_id=&include_items=true`（[get-record-examinations.ts](../../../frontend/src/features/medical-records/api/get-record-examinations.ts) L46–64） | include_items で結果行。limit は HISTORY_FETCH_LIMIT。超過は切り詰め表示 |
| 検歴ピボット | 同一ペット履歴 + `GET .../items`（[ExamPivotTable](../../../frontend/src/features/examinations/components/ExamPivotTable.tsx)） | 空の inspectionValue は行に出さない（L73–75） |

FE 再マップ: [mapExamResultsToFormRows](../../../frontend/src/features/examinations/hooks/use-examination-form-model.ts) L67–83。`inspectionValue` / `unit` / `referenceValue` / サーバ判定を復元。手動行は `examTypeFieldId` undefined のまま。

### 5. カルテ / 検歴表示（「chart」）

製品に検査値の折れ線グラフ専用コンポーネントは無い。要件の表示先は次の 2 面:

1. **カルテ検査タブ** [MedicalRecordExamination](../../../frontend/src/features/medical-records/components/MedicalRecordExamination.tsx) → [ExaminationGroup](../../../frontend/src/features/medical-records/components/ExaminationGroup.tsx): 日付、`machine` バッジ、項目名 / 結果値 / 単位 / 基準値 / HIGH・LOW・未判定。
2. **検歴ピボット** `?historyView=pivot` [ExamPivotTable](../../../frontend/src/features/examinations/components/ExamPivotTable.tsx): 項目×日付の表。基準表示は保存済み `referenceValue`、なければ `refMin-refMax`、なければ定性 bounds の snapshot。トークンを補完しない（L47–61）。

未紐付け機器バナー [LabDeviceUnlinkedBanner](../../../frontend/src/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner.tsx) はカルテ検査タブに出る。手入力フォームとは別。バナーから値を手入力フォームへコピーする導線は本票に含めない。

印刷は保存済み print-snapshot のみ。未保存 formItems を使わない（ExaminationForm.tsx L104–108）。

## 合成で確認するケース（actual は未実行）

対象: 専用合成 clinic / 猫ペット / 尿検査に相当する exam_type。実患者・実試験紙写真・本番機器は使わない。医院承認の項目表が無い間、項目名は fixture ラベルに留め、臨床意味を書かない。

| ID | 操作 | 期待 | 混同してはいけないこと |
| --- | --- | --- | --- |
| M1 | テンプレ行に文字列結果を入れて保存→一覧→再オープン | `inspection_value` / `unit` / `reference_value` が一致。`job_id` NULL | 機器 persist と同一 API にしない |
| M2 | 「検査項目を追加」で名前+結果を保存→再読込 | 手動行が残る。`exam_type_field_id` null。判定は未判定 | マスタ field に勝手に紐付けない |
| M3 | 結果あり・名前空の手動行を保存 | FE が拒否。silent drop しない（既存テスト） | |
| M4 | カルテ検査タブで同一レコードを表示 | 結果値・単位・基準の文字列一致。`machine` 空でも欠落扱いにしない | 空 machine を目視と決めない |
| M5 | `historyView=pivot` で同ペットの手入力行 | 値がある項目がセルに出る。空値は出ない | 機器ジョブの列と結合しない |
| M6 | 同一ペットに手入力 exam と `job_id` 付き受信 exam を共存 | 2 行のまま。互いに上書きしない | 同日同種別を 1 行に畳まない |
| M7 | コードが知る定性トークンと、知らない文字列（例: テストが拒否する `陰性`） | 文字列としては保存され得る。判定はマスタ bounds とトークン一致時のみ。臨床意味は UNKNOWN | `(+)` を「陽性」と翻訳して医院凡例にしない |
| M8 | 確定後の結果編集 | ロック。解除は unconfirm 権限 | 機器 Undo と混ぜない |

成果物列（実行時）: `case / fixture / input string / request JSON keys / DB job_id / reread inspection_value,unit,reference_value / chart or pivot cell / expected / actual / evidence`。actual は全行 **未実行 / UNKNOWN**。

## 既存テストと GAP（ファイルは追加しない）

| 対象 | 既存 | GAP |
| --- | --- | --- |
| 手動行 add/rename/delete と PATCH items | [use-examination-form.items-part1.test.ts](../../../frontend/src/features/examinations/hooks/use-examination-form.items-part1.test.ts) `手動行をimmutableに追加・改名・削除できる`, `追加した手動行の名前と結果値をPATCH itemsへ送る`, `結果値がある手動行の空名を拒否し、silent drop しない` | 尿試験紙の定性文字列専用ケースなし |
| 定性トークン順と比較、非正規表記拒否 | [exam_result_assessment_test.go](../../../backend/internal/medicalrecord/exam_result_assessment_test.go) `TestQualitativeValueOrder` | 医院凡例との一致は対象外。`陰性` はコード上 rejected notation |
| 手動 Create が JobID 無し | Create 構造体に JobID なし（examination_service.go L235–245）。model コメント | 手入力 vs 受信の共存・非上書きの統合テストなし |
| 機器 persist の job_id | [lab_device_exam_persist_test.go](../../../backend/internal/medicalrecord/lab_device_exam_persist_test.go) | 手入力経路へ merge しないことの明示テストなし |
| カルテ表示 | ExaminationGroup / MedicalRecordExamination テスト | 手入力・空 machine の表示契約は薄い |
| `inputMode="decimal"` × 定性記号 | なし | 一部ブラウザで `+` 入力が不便になり得る。実機は UNKNOWN。専用キーパッドを本票で追加しない |
| FE で job_id / origin 表示 | transform が job_id を落とす | 見た目の由来区別は **GAP**。実装は医院が混同した場合の後続単位。本票は DB `job_id` を正本とする |

## 完了 / 停止

完了（後続の実装・UAT）: 医院が承認した項目/定性表現/単位/基準の期待表と、保存・再読込・カルテ/ピボット表示が一致する。機器結果と手入力が別 exam として残り、上書きされない。

本キャンペーン単位の完了: 本票が現行経路を引用し、UNKNOWN を残し、カットオフを発明せず、機器 ingest にマージしない。合成実行の actual は未実行のまま。

停止:

- 院別試験紙凡例・カットオフ・「+ の臨床意味」が無い → 判定ロジックやマスタ qualitative bounds を医院値で埋めない。
- 八王子/猫の機器型番が写真未確認 → 手入力を機器デコーダに接続しない。
- PU-4010 は decoder-only / 運用未承認 → 城東尿を「機器で完了」としない。
- `job_id` を FE が表示しないことだけで手入力実装を失敗としない。混同事故が出たら表示 GAP を別単位にする。
- 自動受信へ無理に統合しない。Lab import の duplicate 判定（同日同種別だけでは同一視しない、lab_import_examination_service.go L62–64）を手入力のマージ根拠にしない。

## 実行しないこと

- 製品コード / テスト / `todo-issue.md` / campaign ledger の変更。
- `make migrate`、実 DB、STG、実患者写真。
- 全マスタを会計へ混ぜる、機器受信 UI の改修、新しい origin 列の先行実装。
