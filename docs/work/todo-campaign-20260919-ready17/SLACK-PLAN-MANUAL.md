# SLACK-PLAN-MANUAL: 診察/治療プラン vs 治療タブと既存手入力

状態: **画面/API 対応 READY／製品実装・自動実施・自動請求 未実行**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-PLAN-MANUAL`（L148–152、索引 L429）。保持する現場条件:

- 「治療プランと治療の違い」と「プランへの手入力」は **別質問**。前者は画面名の対応表、後者は既存手入力の場所
- カルテ見出し「治療プラン」は **`treatments` 行**を編集する。同名の backend `treatment_plans` / [treatment_plan_request.go](../../../backend/internal/medicalrecord/treatment_plan_request.go) の存在を、この UI が予定専用である根拠にしない
- 予定を自動で実施・請求済みにしない。会計は `billing_confirmations` 確認後の未請求集計であり、行追加 POST では走らない
- マスタ単価の欠落・0円・再読込不一致は本票で直さず [UAT-R2-MASTER-PATH](../../../todo-issue.md#uat-r2-master-path) / [全経路票](../todo-campaign-20260918/UAT-R2-MASTER-PATH.md) へ統合する

本票は [MedicalRecordDiagnosisPlan](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) と [use-treatments-tab](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) を照合する。planned-vs-billed / 自動実施・自動請求の意味は変更しない。製品 UI の意味変更はしない。到達性の回帰として [TreatmentTable.test.tsx](../../../frontend/src/features/medical-records/components/TreatmentTable.test.tsx) のみ許可（検索優先 `onOpenSearch || onAddRow`）。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-PLAN-MANUAL` の owned path および人間が読む調査票である。attempt `att-plan-manual-20260921-001` で treatments-tab 手入力と plan-table 検索優先 gap を再確認し、欠落していた TreatmentTable 到達性テストを追加した。入院 `treatment_plans` の仕様変更は本票の範囲外。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「治療プラン」と呼んでいる画面 | タブ名は「診察/治療プラン」、見出しは「治療プラン」、隣タブは「治療」 | **UNKNOWN**。現場がどちらを指すか採取する |
| 「手入力」がマスタ無し行か、マスタ行の単価上書きか | 両方の経路がある（後述） | **UNKNOWN** |
| 0円行を請求したいか | Create は `unit_price >= 0` を許容。0 は未請求候補から除外されない | **UNKNOWN**。デモだから 0 とは断定しない |
| 「予定だけで未実施/非請求」が必要か | 現行のプラン表は `treatments` を即保存し、確認後は未請求に載り得る | **UNKNOWN**。必要と分かった場合だけ臨床 PO が移行条件を裁定するまで仕様変更を停止 |
| フロント/API revision | 本票作成時 worktree HEAD `aac697645`。attempt `att-plan-manual-20260921-001` 照合 HEAD `37cc028b730745574b23db2f238c3525e4e2fd20`（citations L236–239 / L211–214 / L247–248 一致） | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけない名前

現場の「プラン」「治療」「予定」は次のどれでも同じ言葉になる。表・API・入院プランを同一視しない。

| ID | 画面 / 見出し | 保存先 | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| P — 診察/治療プランタブ | [MedicalRecordClinicalTabs](../../../frontend/src/features/medical-records/components/MedicalRecordClinicalTabs.tsx) L66–117。タブ値 `"診察/治療プラン"`（[medical-record-form-model.ts](../../../frontend/src/features/medical-records/routes/medical-record-form-model.ts) L6, L43–44） | **2系統:** 所見3欄は `clinical_plans`。表は `treatments` | 「プラン」だから `treatment_plans` に書く、と読まない |
| T — 治療タブ | 同ファイル L118–125。タブ値 `"治療"`。ラッパ [MedicalRecordTreatment](../../../frontend/src/features/medical-records/components/MedicalRecordTreatment.tsx) → [TreatmentsTab](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx) | **同じ** `treatments`（[use-treatments-tab.ts](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) L10–14, L44–45） | 別テーブル。プラン表と治療タブは **同一レコード集合の別 UI** |
| C — 所見テキスト | DiagnosisPlan の `physicalExam` / `plan` / `assessment`。カルテ保存が PATCH clinical-plan（[use-medical-record-save-action.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) L249–283） | `clinical_plans` | 治療行の予定ステータスではない。保存しても未請求明細は増えない |
| H — 入院治療プラン | [treatment_plan_request.go](../../../backend/internal/medicalrecord/treatment_plan_request.go) L3–14。route は `/hospitalizations/:id/treatment-plans` と `/medical-records/:id/treatment-plans`（[routes_hospitalization.go](../../../backend/internal/medicalrecord/routes_hospitalization.go) L29–36）。FE 利用は入院フォーム（`frontend/src/features/hospitalization/api/`） | `treatment_plans` | この request 型があることを、カルテ「治療プラン」見出しが予定専用である根拠にしない |
| B — 会計確認 | タブ `"会計(医師確認)"` [MedicalRecordBillCheck](../../../frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx)。未請求 SQL は `bc.status = 'confirmed'`（後述） | `billing_confirmations` → 未請求 `billing_items` | プラン表への行追加を請求確定と読まない |

## 画面名 → 対象データ → 保存時点 → 未請求/会計

| 画面名（UI） | コンポーネント | 対象データ | 保存時点 | 未請求 / 会計 |
| --- | --- | --- | --- | --- |
| カルテタブ **診察/治療プラン** の所見・診断 | [MedicalRecordDiagnosisPlan](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) + [ClinicalPlanSection](../../../frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx) | `clinical_plans`（身体所見・治療方針・診断詳細・診断名） | カルテのタブ保存。行追加では送らない | 会計自動連携 **N/A**。所見保存を治療行保存と混ぜない |
| 同タブ見出し **治療プラン** の表 | 同上 L228–254。`useGetTreatments` / `useCreateTreatment` / `useUpdateTreatment` / `useDeleteTreatment`（L78–84） | `treatments` via `POST/PATCH/DELETE /v1/medical-records/:id/treatments`（[treatments.ts](../../../frontend/src/features/medical-records/api/treatments.ts) L27–54、[routes_records.go](../../../backend/internal/medicalrecord/routes_records.go) L48–52） | **即時。** 新規カルテはプレースホルダ「カルテを保存してから治療プランを作成できます」（DiagnosisPlan L235–240） | 行追加だけでは未請求に出ない。親カルテの `billing_confirmations.status='confirmed'` かつ未紐付け `billing_items` のとき [FindUnbilledByPetID](../../../backend/internal/medicalrecord/treatment_repository.go) L107–123 → [treatmentToUnbilledBillingItem](../../../backend/internal/billing/billing_item_unbilled.go) L124–142（`source=medical_record`, `treatment_id`） |
| カルテタブ **治療** | [MedicalRecordTreatment](../../../frontend/src/features/medical-records/components/MedicalRecordTreatment.tsx) L24–28 は新規時「カルテを保存してから治療明細を追加できます」。保存後は [useTreatmentsTab](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) | **同じ** `treatments` | **即時** create/update/delete/reorder | 上と同じ未請求条件。プラン表で作った行は治療タブに再読込され、逆も同じ |
| カルテタブ **会計(医師確認)** | [MedicalRecordBillCheck](../../../frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx) L84–94 | 同じ `treatments` を再取得し、確認 API を別途呼ぶ | 確認ボタンは confirmation write。治療行の create を呼ばない経路と、マスタから治療を足す経路が共存 | 確認後に会計側が未請求候補を読む。complete は別会計フロー。プラン保存の副作用ではない |
| 入院フォームの治療プラン | hospitalization `treatment-plans` API | `treatment_plans` | 入院 create 同梱または nested POST | 退院会計は `care_plan_items` 側（MASTER-PATH 票）。本カルテ UI の未請求 SQL は `treatments.id` を見る |

未請求 SQL は `is_selected` も `treatments.status` も見ない（[treatment_repository.go](../../../backend/internal/medicalrecord/treatment_repository.go) L109–116）。billing パッケージに `IsSelected` 参照は無い。選択解除や `pending` 表示を「非請求」と案内しない。除外条件は削除済み、確認未了、既に非キャンセル billing に `treatment_id` がある、または親カルテに非キャンセル billing があること。

## 既存手入力（新規行を足す場所）

要望「プランへ手入力したい」は、現行配信版に **手入力 create がある**。無い機能として新 API を足す前に、操作位置を案内する。

| 入口 | 操作 | Create body | UI が handler を呼ぶか |
| --- | --- | --- | --- |
| **治療タブ（案内の第一候補）** | [TreatmentAddControls](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx) L211–214 の「手入力で追加」→ 種別・内容 → 「追加」 | [handleAddSubmit](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts) L184–214。`item_type` は選択値、`content` 必須 trim、**`unit_price: 0`**, `quantity: 1`, `is_selected: true`, `is_insurance: false` | **呼ぶ。** 内容空はクライアントで return（L185） |
| **診察/治療プランの表** | [handleAddRow](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) L135–147。`item_type: "other"`, `content: ""`, **`unit_price: 0`**, `quantity: 1`, `is_selected: true` | `onAddRow={canCreate ? handleAddRow : undefined}`（L248） | **現行ボタンは検索優先。** [TreatmentTable](../../../frontend/src/features/medical-records/components/TreatmentTable.tsx) L236–239 は `onClick={onOpenSearch \|\| onAddRow}` かつラベル「行を追加（検索）」。`canCreate` 時は `onOpenSearch` が truthy のため **handleAddRow はクリック経路から到達しない**。コード上の空行 create は残っているが、プラン表の見えるボタンはマスタ検索 |
| 両タブのマスタ選択 | DiagnosisPlan [handleSelectTreatment](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) L149–165。治療タブ [buildMasterSelectionPayload](../../../frontend/src/features/medical-records/lib/treatments-tab-model.ts) L21–38 | マスタ `unitPrice` をコピー。手入力ではない | 単価欠落はマスタ側。本票でマスタフォームを直さない |
| 追加後のセル編集 | DiagnosisPlan [handleUpdateItem](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) L116–132。治療タブ `handleUpdate` | `unit_price` / `content` 等を PATCH | マスタ無し行の単価入力、およびマスタ行の上書き。保存済み行の編集であり、新規手入力ボタンの代替 |

治療タブだけがラベル付き「手入力で追加」を出す。プラン表でマスタ無し行が必要なら、現行操作は **治療タブで手入力**（同一 `treatments`）か、検索で近いマスタを足してから単価/内容を PATCH。プラン表の空行 handler を「画面に出ている手入力」と案内しない。

マスタ選択の差分（バグ修正対象ではない。混ぜない）:

| | 診察/治療プラン | 治療タブ |
| --- | --- | --- |
| 単価 | `item.unitPrice` | 同じ `item.unitPrice` |
| `medicine_id` | **送らない**（L152–162） | `Number(item.medicineId)`（treatments-tab-model L32） |
| `is_insurance` | **true**（L158） | **false**（L34） |
| 用量ゲート | なし | 薬剤は体重・上限で create を止め得る（use-treatments-tab L235–297） |
| 並び替え | なし | `PUT /treatments` reorder |

## 価格欠落は MASTER-PATH（本票で直さない）

9月16日 Slack の「治療プランから入力したマスタの金額」「マスタがない場合と0円」は [todo-issue.md UAT-R2-MASTER-PATH](../../../todo-issue.md#uat-r2-master-path) L56 が既に同じコンポーネントを引用している。

| 症状 | 本票の扱い | 統合先 |
| --- | --- | --- |
| マスタ登録画面の単価が保存/再読込で消える | 12フォーム検証。本票は触らない | [UAT-R2-MASTER-PATH.md](../todo-campaign-20260918/UAT-R2-MASTER-PATH.md) 表「現行 route / form 対応」 |
| 検索で選んだ行の `unit_price` がマスタと違う | 下流ケース。選択 payload は `item.unitPrice` | 同票「診察・処置・薬剤などの治療」行。DiagnosisPlan と use-treatments-tab の両入口を追う |
| マスタが無く手入力したい | **既存機能。** 治療タブ「手入力で追加」。0円 create は欠落ではない | MASTER の価格欠落と混同しない（同票「12フォーム外」の手入力治療行） |
| 手入力後に 0 のまま請求された | 0 は Create 許容（[treatment_service.go](../../../backend/internal/medicalrecord/treatment_service.go) L176–178）。未請求変換は `t.UnitPrice` をそのまま載せる（billing_item_unbilled.go L132） | 0 と NULL と未保存を同値にしない（MASTER-PATH 「0/空欄/欠損/未保存」）。医院が 0 を請求したいかは UNKNOWN |
| 入院 `treatment_plans` の単価 | 別テーブル | MASTER-PATH の入院プラン/ケージ。本 UI の根拠にしない |

本票はマスタフォーム・会計選択欄・負数バリデーションを変更しない。価格バグの修正 receipt は MASTER-PATH の経路別行に書く。

## 自動実施・自動請求をしない（停止条件）

現行コードに「プラン行を実施済み/請求済みへ昇格する」ジョブは、この UI 経路に無い。次を実装しない。

- 診察/治療プランの表保存を `status=completed` へ自動 PATCH する
- 行追加 POST の成功で `billing_confirmations` を confirmed にする
- 行追加 POST の成功で `billings` / `billing_items` を作る
- `treatment_plans` 行を `treatments` へ自動コピーして請求する
- 予定専用フラグを推測して未請求 SQL から除外する（PO 裁定前）

真に「予定だけで未実施/非請求」が業務要件なら、実施への移行条件を臨床 PO が裁定するまで仕様変更を停止する（todo-issue L152）。既存の手入力で質問が足りるなら操作案内で閉じる。

## 期待ケース（受入。実行はしない）

| ID | 操作 | 期待 |
| --- | --- | --- |
| E1 | 保存済みカルテの治療タブで「手入力で追加」→ 内容入力 → 追加 | `POST .../treatments` 1件。再読込で治療タブとプラン表の両方に同じ行。会計確認前は未請求に出ない |
| E2 | プラン表「行を追加（検索）」 | 検索ダイアログ。空行 `handleAddRow` は走らない。回帰: [TreatmentTable.test.tsx](../../../frontend/src/features/medical-records/components/TreatmentTable.test.tsx) が `onOpenSearch` 優先を固定 |
| E3 | マスタ選択 | `unit_price` はマスタ値。欠落/0/再読込不一致は MASTER-PATH のケース |
| E4 | 会計(医師確認) で confirmed | その後の未請求に `treatment_id` 付き候補。プラン POST 単体では confirmed にならない |
| E5 | 入院画面の治療プラン | `treatment_plan_request` / `treatment_plans`。カルテ見出し「治療プラン」の保存先ではない |

## PO・停止

- 既存手入力の場所・保存・再読込・会計との関係は上表。足りれば操作質問として閉じる
- 価格欠落は MASTER へ統合。本 unit でマスタ/会計コードを直さない
- 「予定専用」が必要かは UNKNOWN。裁定前に自動実施・自動請求・別テーブルへの載せ替えをしない
- 入院 `treatment_plans` とカルテ `treatments` を一つの「プラン」仕様にまとめない
