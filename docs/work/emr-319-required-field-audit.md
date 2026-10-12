# EMR-319 必須フィールド表示棚卸し（11区分和集合）

監査日: 2026-10-09 / 対象 worktree: `emr-cross-ui` / 実施: UNIT-EMR-319 att-319-001

> **ステータス（是正実施後）**: 本ドキュメントの各表は是正前の棚卸しスナップショット（point-in-time）として残す。⚠️/🟡 の全項目は att-319-001 で表示配線済み。最終状態は文末の「是正後の最終状態」を参照。

## 監査基準

チケット要件: 「入力必須項目が未入力で保存/次へ進む時に、その項目が赤枠+エラーメッセージになること」。
各必須項目について以下3点を記録する:

- **aria-invalid**: コントロールに `aria-invalid`（または同等 prop）がエラー連動で付与されるか
- **赤枠**: エラー時に destructive 系の赤枠（`aria-invalid:border-destructive` / `C.borderDanger` / `STYLE.formInputError` 相当）が出るか
- **エラー文**: `FormFieldError` / `role="alert"` 等でエラーメッセージが描画されるか

判定列: ✅ = 配線済み / ⚠️ = エラー文のみ・赤枠なし（要配線）/ 🟡 = ネイティブ `required` のみ（ブラウザ標準メッセージ、アプリ層配線なし）/ ➖ = 対象外（必須項目なし・監査対象外）

**検証ルール非変更の前提**: 必須判定ロジック（validationErrors / fieldErrors の生成・required フラグ・文案）は一切変更しない。以下の「是正」は表示配線のみ。

## 共通部品の現状

| 部品 | aria-invalid 対応 | 備考 |
|---|---|---|
| `components/ui/input.tsx` | ✅ `aria-invalid:border-destructive aria-invalid:bg-destructive/5` | prop を渡すだけで赤枠+薄赤背景 |
| `components/ui/textarea.tsx` | ✅ 同上 | |
| `components/ui/select.tsx` (`SelectTrigger`) | ✅ 同上 | |
| `components/ui/checkbox.tsx` | ✅ `aria-invalid:border-destructive` + ring | |
| `components/ui/radio-group.tsx` | ✅ 同上 | |
| `components/ui/searchable-select.tsx` | ✅ `ariaInvalid` prop → trigger | 呼び出し側が未指定の箇所あり |
| `components/shared/SidePeek/SidePeekTitleInput.tsx` | ✅ `error` prop → `aria-invalid` + FormFieldError | マスタ側パネルのタイトル=名称は全件配線済み |
| `components/shared/SidePeek/MoneyInput.tsx` | ✅ `error` prop → `aria-invalid` | |
| `components/shared/DatePicker` (SinglePicker) | ❌ `ariaInvalid` prop なし | 本単位で additive prop を追加 |
| `components/shared/MasterSelectModal/MasterSelectTrigger.tsx` | ❌ prop なし | 本単位で additive prop を追加 |
| `components/shared/NextScheduleField` | ⚠️ `error` prop で文のみ、内部 DatePicker 未配線 | 内部で DatePicker へ `ariaInvalid` 伝播を追加 |

---

## 1. 予約（`components/shared/ReservationFormModal` / `features/reservations`）

### ReservationFormModal（新規/編集・tentative 確定共通）

必須判定: `ReservationFormModal.tsx` `errors.patient/date/time/type/doctor`（doctor は orphan 確定時のみ）。

| 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|
| 予約対象（患者選択） | SelectedPatientSummary コンテナ | ➖（非入力コントロール） | ❌ コンテナ枠線が通常色のまま | ✅ `FormFieldError` | ⚠️ 赤枠不足 |
| 日付 | Popover トリガー button | ❌ | ❌ | ✅ `FormFieldError` | ⚠️ |
| 時間（開始/終了） | SelectTrigger ×2 | ❌ | ❌ | ✅ `FormFieldError` | ⚠️ |
| 予約区分 | 区分ボタン | ✅ | ✅ `C.borderDanger` | ✅ `FormFieldError` | ✅ |
| 担当医（orphan 確定時） | staff select 表示 | ➖（エラーは状態由来） | ➖ | ✅ `FormFieldError` | ✅（必須ルールなし・状態エラーのみ） |

### NewOwnerInlineForm（新規飼主インライン）

| 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|
| 飼主名 | Input | ❌ | ❌ | ✅ | ⚠️ |
| 電話番号 | Input | ✅ `aria-invalid` + describedby | ✅ | ✅ | ✅ |
| ペット名 | Input | ❌ | ❌ | ✅ | ⚠️ |
| 動物種 | SearchableSelect | ❌（`ariaInvalid` prop 未指定） | ❌ | ✅ | ⚠️ |
| 主訴 | Textarea | ❌ | ❌ | ✅ | ⚠️ |

### その他

- `PatientSelectionResults` / `PatientSelectionTable`: 選択リスト（入力項目なし）→ ➖
- `NextVisitResetDialog` / `ReservationTypePickerDialog` / `ReservationNotesField`: 必須項目なし → ➖
- `features/reservations` 配下に独自フォームなし（フォーム本体は shared/ReservationFormModal）→ ➖

## 2. カルテ（`features/medical-records`）

| フォーム | 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|---|
| CheckupsTab 追加行 | 実施日 | DatePicker | ❌（prop なし） | ❌ | ✅ | ⚠️ |
| 〃 | 健診種別 | CheckupTypeSelect(SelectTrigger) | ❌ | ❌ | ✅ | ⚠️ |
| VitalsTab 追加行 | 記録日時 | raw `<input type="datetime-local">` | ❌ | ❌ | ✅ | ⚠️ |
| 〃 | 体温他 | raw `<input type="number">` | ❌ | ❌ | ✅ | ⚠️ |
| VitalsTab 編集行 | 記録日時・体温 | raw input ×2 | ❌ | ❌ | ✅ | ⚠️ |
| VaccinationForm（カルテ内） | ワクチン種別 | SearchableSelect | ✅ `ariaInvalid` | ✅ | ✅ | ✅ |
| 〃 | 接種日 | DatePicker | ⚠️ prop 不在 | ❌ | ✅ | ⚠️ |
| DiagnosisHeaderDiagnosis | 診断1 | SearchableSelect | ✅ | ✅ | ✅ | ✅ |
| NextVisitDateField | 次回日 | input | ✅ | ✅ | ✅ | ✅ |
| TreatmentQuantityCell | 用量 | input | ✅ | ✅ | ✅ | ✅ |
| MedicalRecordForm 本体 | — | — | — | — | — | ➖ 必須項目なし（各タブ内フォームに集約） |

## 3. 入院（`features/hospitalization`）

| フォーム | 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|---|
| HospitalizationForm | ペット | （前画面で選択・フォーム内に入力コントロール無し） | ➖ | ➖ | ✅ `FormFieldError` | ✅（表示のみ対象・対話コントロール非存在のためテキスト表示が妥当） |
| 〃 | ケージ/個室 | — | — | — | — | ➖ EMR-316 で必須撤回済み |
| CarePlanTab AddForm | 名称 | MasterSelect 経路/名称入力 | ❌ | ❌ | ✅ `role="alert"`（統合文） | ⚠️ |
| 〃 | その他理由（手入力項目時） | Input | ❌ | ❌ | ✅ 同上 | ⚠️ |
| 〃 | 単価（手入力項目時） | Input | ❌ | ❌ | ✅ 同上 | ⚠️ |

## 4. 会計（`features/accounting` / `cash-register` / `estimates`）

| フォーム | 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|---|
| PaymentCard | 釣銭上書き split 金額 | Input | ✅ | ✅ | ✅ | ✅ |
| ItemListCard | 項目名（EMR-229） | Input | ❌（blur 空値はフックで toast のみ） | ❌ | ⚠️ toast のみ | ⚠️ |
| ItemListCard | 手入力単価 | Input | ➖ | ➖ | ✅ `FormFieldError`（manualPriceError） | ⚠️（要 aria-invalid 確認） |
| CreditCorrectionDialog | 訂正理由（必須表記） | Textarea | ❌ | ❌ | ⚠️ toast のみ | ⚠️ |
| RefundSection | 返金額 | Input | ❌（超過時 toast、空は無言 return） | ❌ | ⚠️ toast/無言 | ⚠️ |
| CashRegisterClosePanels | 実際の現金額 | Input | ✅ `aria-invalid`（<0 で）+ `required` | ✅ | ✅ | ✅ |
| EstimateFormSections | タイトル | Input | ❌ | ❌ | ✅ `FormFieldError` | ⚠️ |
| 〃 | ステータス | SelectTrigger | ❌（describedby のみ） | ❌ | ✅ | ⚠️ |

## 5. 飼主・ペット（`features/owners` / `features/pets`）

| フォーム | 項目 | 判定 |
|---|---|---|
| OwnerForm（OwnerBasicFields） | 飼主名・よみ・電話・メール | ✅（aria-invalid+describedby+formInputError+FormFieldError の参照実装） |
| OwnerForm（OwnerAddressFields） | 郵便番号・自宅郵便番号 | ✅ |
| OwnerForm（OwnerMembershipFields） | 値引率 | ✅ |
| PetEditModal（Identity/Physical/SubOwners） | ペット名・動物種・性別・特記理由・体重等 | ✅ |
| PetSubOwnersSection | 続柄重複 | ✅ |
| PetCareSection / PetChronicConditionsSection / OwnerInfoSection | — | ➖ 必須入力なし |
| `features/pets` | — | ➖ フォームなし（api のみ） |

## 6. トリミング（`features/trimming`）

| 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|
| 担当者（staffId） | PatientInfoCard 内 staff Button | ❌ | ❌ | ✅ `FormFieldError` | ⚠️ |
| コース（courseId） | MasterSelectTrigger(block) | ❌（prop なし） | ❌ | ✅ | ⚠️ |
| 予約区分（reservationTypeId） | —（設定欠落エラー） | ➖ | ➖ | ✅ | ✅（フィールド入力でなく設定エラー） |

## 7. 予防接種（`features/vaccinations`）

| 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|
| 接種日 | DatePicker | ❌ | ❌ | ✅ | ⚠️ |
| ワクチン | SelectTrigger | ❌ | ❌ | ✅ | ⚠️ |
| 次回予定日（条件付き必須） | NextScheduleField→DatePicker | ❌ | ❌ | ✅ `error` prop → `<p role="alert">` | ⚠️ |

## 8. 健診（`features/checkups`）

| 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|
| 実施日 | DatePicker | ❌ | ❌ | ✅ | ⚠️ |
| 健診種別 | SearchableSelect | ❌（prop 未指定） | ❌ | ✅ | ⚠️ |

## 9. 検査（`features/examinations` / `features/lab-device`）

| フォーム | 項目 | コントロール | aria-invalid | 赤枠 | エラー文 | 判定 |
|---|---|---|---|---|---|---|
| ExaminationFormFields | 検査種別 | SearchableSelect | ✅ `ariaInvalid` | ✅ | ✅ | ✅ |
| 〃 | 担当医 | SearchableSelect | ✅ | ✅ | ✅ | ✅ |
| ExaminationForm | 手動検査項目の項目名 | ExamItemsTable 内 input | ❌ | ❌ | ✅ `role="alert"` セクション文 | ⚠️ |
| ExaminationUnconfirmDialog | 確定解除理由 | Textarea | ✅ | ✅ | ✅ | ✅ |
| `features/lab-device` | — | — | — | — | — | ➖ 必須入力フォームなし（デバイスボード/マスタ登録は master 区分側） |

## 10. 在庫（`features/inventory`）

| 項目 | 判定 |
|---|---|
| 名称・単位・数量・下限在庫（InventoryFormSections） | ✅（aria-invalid + aria-required + FormFieldError 全項目配線済み） |

## 11. 設定系（`settings` / `clinic-settings` / `closing-settings` / `master` / `shifts`）

### settings

| フォーム | 項目 | 判定 |
|---|---|---|
| LstepTagConfigSection 追加行 ×3（プレフィックス+カテゴリ / 疾患コード+タグ名 / 送信目的+タグプレフィックス） | raw input ×2 ずつ | ⚠️ フォーム共通の `state.error` 文のみ。input に aria-invalid/個別赤枠なし（form は `noValidate` 済み） |
| LstepSettingsForm | — | ➖ 必須項目なし（接続設定） |

### clinic-settings

| フォーム | 項目 | 判定 |
|---|---|---|
| ClinicMasterSidePanel | 医院名 | ⚠️ FormFieldError あり・スタイル付き input に aria-invalid/赤枠なし |

### closing-settings（ネイティブ required 方式）

| フォーム | 項目 | 判定 |
|---|---|---|
| HolidaySection | 日付 | 🟡 `required` のみ（ブラウザ標準メッセージで送信阻止。アプリ層の aria-invalid/赤枠/FormFieldError なし） |
| SpecialPeriodSection | 開始日・終了日・午前午後区切り・午後終了 | 🟡 同上 |
| StandardClosingTimeSection | 標準締め時間（3項目） | 🟡 同上 |

### master（サイドパネル群）

| フォーム | 項目 | 判定 |
|---|---|---|
| 全パネル共通タイトル=名称（AnimalSpecies/Cage/Complaint/DiagnosisName/DiagnosisType/Hospitalization/Insurance/InterviewTemplate/LabDeviceItem/Medicine/Merchandise/Occupation/PaymentMethod/PermissionGroup/ReservationType(Group)/Staff/TreatmentItem/TrimmingCourse(Type)/TrimmingOption） | 名称 | ✅ `titleError` → SidePeekTitleInput が `aria-invalid`+FormFieldError |
| TreatmentItemSidePanel | 単価 | ✅ `priceError` → MoneyInput `aria-invalid` |
| InsuranceSidePanel | カバー率 | ✅ `aria-invalid` |
| DiagnosisNameSidePanel | カテゴリ | ⚠️ FormFieldError あり・SelectTrigger に aria-invalid なし |
| CampaignSidePanel | 開始日・終了日 | ⚠️ `periodError` は素の `<p>`（role なし）・両 date Input に aria-invalid なし |
| DiagnosisSettingsSidePanels / StaffSidePanelSections | — | ➖ レイアウト部品（必須入力なし） |

### shifts

| フォーム | 項目 | 判定 |
|---|---|---|
| ShiftFormDialog | 開始時刻・終了時刻（勤務種別が RequiresTimeSlot 時必須） | ⚠️ `state.timeError` → FormFieldError あり・time Input に aria-invalid なし |
| ShiftTemplateSettings | 開始/終了時刻・休憩時刻 | ⚠️ `timeError` → FormFieldError あり・PropInput(time) に aria-invalid なし |

---

## 集計と是正見積

- **配線済み（✅）**: 約 40 項目（owners・inventory・検査種別/担当医・cash-register・master 名称群・PaymentCard・VaccinationForm(MR)・DiagnosisHeader 等）
- **要配線（⚠️）**: 約 36 項目
- **ネイティブ required のみ（🟡）**: closing-settings 3 フォーム計 8 項目
- **対象外（➖）**: lab-device・pets・reception・必須項目なしの各部品

### 是正アプローチ（表示配線のみ・検証ルール非変更）

1. `DatePicker`（SinglePicker）に `ariaInvalid?: boolean` を additive 追加（input に `aria-invalid`、外枠 div に danger border/ring）→ DatePicker 系 7 箇所を一括解消。
2. `MasterSelectTrigger` に `ariaInvalid` prop 追加 → trimming コース等。
3. `NextScheduleField` 内部で DatePicker へ `ariaInvalid={Boolean(error)}` 伝播（prop 追加なし）。
4. `CheckupTypeSelect`（CheckupsTabRows 内ローカル関数）に `ariaInvalid` → SelectTrigger。
5. `SearchableSelect` は prop 既存 → 呼び出し側に `ariaInvalid` を渡すのみ（NewOwnerInlineForm 動物種・CheckupForm 健診種別）。
6. raw `input`（VitalsTab・LstepTag・closing-settings）は `aria-invalid` 属性 + 条件付き `STYLE.formInputError`/`C.borderDanger` クラス。
7. `state.error` 統合文のフォーム（CarePlanAddForm・LstepTag）は **アクションの返り値に `field` 識別子を additive 追加**（文案・判定は不変）して対象 input にだけ `aria-invalid` を付与。
8. toast のみの会計系（CreditCorrection 理由・Refund 金額）はローカル `*Error` state + `aria-invalid` + FormFieldError を追加（toast は残す・additive）。
9. 予約モーダル: 日付トリガー button・時間 SelectTrigger ×2・患者コンテナ枠線。
10. closing-settings: `required` は残し `onInvalid` でアプリ側の aria-invalid+FormFieldError を駆動（ネイティブバブルは preventDefault で抑制し、設計言語の表示に統一）。必須ルールの定義は `required` 属性のまま不変。

### 見積と分岐判定

上記は全て機械的な表示配線（prop 追加 + 属性付与 + 既存 FormFieldError/token 利用）。想定修正ファイル ~24、追加テスト assertion は既存 seed テストへの追記中心。**見積 ≈1 日以内で是正分岐（fix）を選択**。closing-settings の onInvalid 配線と CarePlanAddForm の field 識別子追加だけが設計判断を要し、合計しても 2 日を大きく下回る。

### 範囲外 finding（受付・その他）

- 受付（`features/reception`）: チケット未列挙のため監査対象外。参考として ReceptionDetailModalParts 等に必須系入力は見当たらず（予約操作は shared/ReservationFormModal に集約）、大きな抜けは確認されなかった。詳細監査は別チケット推奨。

---

## 是正後の最終状態（att-319-001 完了時点）

上記の監査表は是正前の棚卸し（是正アプローチの根拠）であり、以下が実施済みの最終状態。検証ルール（必須判定・文案・`required`/`isRequired`・zod schema・`validationErrors`/`fieldErrors` 生成）は一切変更していない。

### 是正済み項目（監査表の ⚠️/🟡 全件）

| 是正手段 | 対象 | 結果 |
|---|---|---|
| `DatePicker`（SinglePicker）へ `ariaInvalid?: boolean` を additive 追加 | 予約日・入退院・予防接種日・健診日・検査日等の DatePicker 系 7 箇所 | ✅ `aria-invalid` + 赤枠 |
| `MasterSelectTrigger` へ `ariaInvalid` prop 追加 | トリミング コース等 | ✅ |
| `NextScheduleField` → 内部 DatePicker へ `ariaInvalid` 伝播 | 予防接種・健診の次回予定日 | ✅ |
| `CheckupTypeSelect`（CheckupsTabRows 内ローカル）へ `ariaInvalid` | 健診 健診種別 | ✅ |
| `SearchableSelect` 既存 `ariaInvalid` prop の呼び出し側配線 | NewOwnerInlineForm 動物種・CheckupForm 健診種別 | ✅ |
| raw `input` へ `aria-invalid` + `STYLE.formInputError`/`C.borderDanger` | VitalsTab・LstepTag・closing-settings・会計インライン編集 | ✅ |
| アクション返り値への `field` 識別子 additive 追加（統合エラー文の分解） | CarePlanAddForm・LstepTag | ✅ |
| toast のみ → ローカル `*Error` state + `aria-invalid` + FormFieldError 追加 | 会計 CreditCorrection 理由・Refund 金額 | ✅ |
| 予約モーダル: 日付トリガー button・時間 SelectTrigger・患者コンテナ | 予約フォーム 3 箇所 | ✅ |
| closing-settings: `required` 維持 + `onInvalid` で `aria-invalid`+FormFieldError 駆動 | 3 フォーム 8 項目 | ✅（ネイティブバブル抑制＋アプリ表示） |

### 検証済み事項

- `aria-invalid` は実コントロール（input / SelectTrigger / DatePicker input / コンテナ）に付与され、`aria-describedby` の id は対応する `FormFieldError`/`role="alert"` 要素と一致する。
- `aria-invalid` は有効時 `undefined`（属性なし）とし、誤ってリテラル `"false"` を出力しない（ReservationDateTimeFields・EstimateForm・CheckupsTabRows で修正済み）。
- スコープ検証（A2/A4）: 98 テストファイル seed に対し `scripts/checks/verify-agent-task.py` 実行 — vitest 901 テスト PASS・ESLint（`--max-warnings 0`）exit 0・Prettier exit 0。

### 残件（Phase 2 推奨・本チケット範囲外）

- 受付（`features/reception`）の詳細監査 — 上記「範囲外 finding」参照。
- `aria-describedby` の未付与箇所の統一（aria-invalid + エラー文は満たすが describedby 未接続の少数例 — 表示要件は充足）。

---

## レビュー是正（PR #614 ラウンド2・レビュー指摘対応）

PR #614 レビュー（HEAD `1d76912e2` 時点）で指摘された是正漏れを追加対応した。「是正後の最終状態」節の記述は att-319-001 時点のものであり、本節が最新の正本となる。検証ルール（必須判定・文案・zod schema）はラウンド2でも一切変更していない。

### ラウンド2 是正項目

| 指摘 | 是正 | 検証ルール不変の根拠 |
|---|---|---|
| CreditCorrection 金額が toast のみ | `amountError` state + `aria-invalid` + `aria-describedby` + `FormFieldError` + `STYLE.formInputError` を NumberInput へ配線（再オープン時のリセット含む） | 判定 `!Number.isFinite || <=0` と文案・toast は既存のまま |
| 死亡日時あり（`deceased_at` 非null・status非死亡）でもバッジ表示 | `isPetDeceasedForClinicalWrite`（status=死亡 OR deceasedAt非null）へ全surface統一。共有ヘッダー status・選択ゲート・臨床書込ゲートも同契約に揃えた | BE 死亡契約（`pet/repository_lifecycle_or_contract_test.go`）への準拠。バッジ抑止は新規ルール追加ではなく契約適用 |
| CarePlanAddForm / ItemListCard / RefundSection の必須エラーに到達不能 | ボタン `disabled` から検証派生条件のみ除去（pending 中の `isRefunding`/`addPending` は維持）。action/handler 内の既存判定はそのままフィールドエラーへ接続 | ボタンの到達性のみ変更。必須判定・早期return・文案は不変 |
| LstepTag 片側未入力で両欄 invalid | `requiredInvalid` → `missingFields`（未入力フィールド名配列）で per-field invalid 化。共有エラー要素はペア専用 ID（`fieldA__fieldB-error`）を両欄から参照 | 必須判定（両方必須）と文案は不変。invalid 対象の絞り込みのみ |
| `ring-[#C0392B]/30` 直書き逸脱 | `STYLE.formInputError` へ統一（DatePickerSingle・MasterSelectTrigger・PatientInfoCard・予約モーダル2箇所） | 視覚等価のトークン化のみ |
| HospitalizationFormFields.test fixture TS2353 | `satisfies` を実 prop 型から導出（`NonNullable<Parameters<typeof HospitalizationFormFields>[0]["selectedPet"]>`） | テスト型のみ |
| reservation petSummaryResponse に birth_date 未到達 | `birth_date`/`deceased_at` を additive シリアライズ。transform が `petBirthDate`/`petDeceasedAt` を伝播、`ReservationDetailModal`/`HospitalizationPatientHeader` は `useGetPet`（owners:view 依存）を廃止し wire 値を使用 | ワイヤ additive。既存キーは不変（厳密キー集合テスト2件を意図的更新） |

### ラウンド2 検証

- scoped `verify-agent-task.py --paths <変更71+1ファイル>`（Docker ephemeral offline、frontend image `0e3c736f…` + `ekarte-frontend-node-modules`、backend image + `ae-gomod-cache`）
- BE: `go test -short ./internal/pet ./internal/reservation ./internal/medicalrecord` ＋ 変更ファイル gofmt
- FE: 変更ファイル scoped vitest / eslint / prettier
- 回帰テスト追加: CreditCorrection 金額（0/空）、Refund 非正値（0/負/空、実クリック）、LstepTag per-field、ItemListCard 必須3種（実クリック）、AddForm 実クリック、BE serializer（reservation・medicalrecord・pet・shared_pets）

---

## レビュー是正（PR #614 ラウンド3・再レビュー指摘対応）

PR #614 再レビュー（HEAD `3690bd6a3` 時点）の新規4指摘＋未解決 Codex 指摘6件への対応。本節が最新の正本。検証ルール（必須判定・文案・zod schema）はラウンド3でも一切変更していない。

### ラウンド3 是正項目

| 指摘 | 是正 | 検証ルール不変の根拠 |
|---|---|---|
| AddForm の `min=0`/`step=1` ネイティブ制約が action 未到達で必須エラーを抑止 | `<form noValidate>` 追加。負数単価×名称空・小数単価×理由空・負数単価のみの実クリック回帰テスト3件追加 | 判定（負数/非整数/空）と文案は既存 action 内ロジックのまま |
| SpecialPeriodSection が MasterSidePanel の `form noValidate` 配下で `onInvalid` 不発 | formAction 内で `REQUIRED_FIELDS`（start_date/end_date/am_pm_boundary/pm_end）の FormData 空欄チェックを追加し、既存 `invalidFields` 表示経路へ接続。実保存クリック経路の回帰テスト2件追加 | 必須対象・文案・`required` 属性は不変。StandardClosingTime/Holiday は独自 `<form action>` で noValidate なしのため既存 `onInvalid` 経路で到達可能（対象外） |
| 死亡登録/取消 mutation が `pets.*` のみ invalidate で、REALTIME staleTime(2分)内の予約詳細が生存時 sentinel を保持 | 両 hook の onSuccess で `reservations.all/detailAll`・`reception.all`・`hospitalizations.all/detailAll`・`ownerSharedPets.all`・`ownerReportPetsAll` を追加 invalidate（新 prefix factory を queryKeys に追加）。`ReservationDetailModal` は `useGetReservation`（`reservations:view` 同一権限）で最新応答をクリック時 snapshot へ優先適用（`effective`） | キャッシュ鮮度のみ。API・権限・payload 不変。use-pet-death.test.tsx 3件追加 |
| api.yaml が `petSummaryResponse`/`SharedPet` の additive `birth_date`/`deceased_at` を未記述（wire/generated ずれ） | `PetSummary` に `birth_date`(format:date)/`deceased_at`(format:date-time) を追加、`OwnerSharedPet` に `deceased_at` を required 追加（serializer 非omitempty=常時送出と一致）。drift ゲートの良性衝突3サイトを `knownDateFormatDrifts` へ追記（PatchPetDeathRequest format:date との json名衝突、wire は date-time で schema と一致） | OpenAPI 正本の additive 同期のみ。`make codegen` はユーザー実行専用のため未実行 |
| CampaignSidePanel 片側欠落で両欄 invalid（Codex） | `requiredInvalid` → `periodMissing`(ReadonlySet) per-field 化。入力で当該欠落のみ解除（全欠落解消時は periodError もクリア）。ペア日付逆転は従来どおり両欄 invalid | 必須判定（両方必須）・日付逆転判定・文案は不変 |
| EstimateForm が deceased_at のみの半死亡ペットで新規作成を許容 | `isNewEstimatePetDeceased`（`isPetDeceasedForClinicalWrite`）をブロックゲートへ追加。deceased_at-only・生存 null の回帰テスト追加 | 既存「非生存ブロック」に OR 契約 sentinel を合流。判定追加ではなく契約準拠 |
| ItemListCard の価格エラーが名称/カテゴリ early return の後に評価され初回未表示 | 価格評価を early return 前へ移動し全必須エラーを同一送信パスで表示 | 判定・文案・送信条件は不変。評価順序のみ |
| DatePickerSingle に aria-describedby 未接続（Codex） | `ariaDescribedBy?: string` prop を追加し内部 input へ転送。Checkup/Vaccination（medical-records・vaccinations）3箇所の FormFieldError id と接続 | 属性配線のみ |
| WeekView/MonthView の誕生日情報が静的バッジ title 内に留まり到達不能（Codex） | 親チップの `title`/`aria-label` に `今月誕生日（生年月日: …）` を合流（tentative ラベルと併記）。死亡OR契約で抑止 | バッジ表示条件・JST 判定は不変。アクセシビリティ経路のみ |

### ラウンド3 検証

- scoped `verify-agent-task.py --paths <変更27+1ファイル>`（Docker ephemeral offline、backend image Go1.26 `animalekarte-backend:latest`、frontend image + `ekarte-frontend-node-modules`/`ekarte-go-mod-cache`）
- BE: `go test -short ./internal/apicontract` 33 テスト PASS・変更 Go ファイル gofmt PASS
- FE: `vitest related` 134 file / 988 test — 133 file PASS・唯一の失敗は `reservation-management.filter.test.ts`（EMR-308 既知 TZ flake、verifier が env -i で TZ を剥がすため UTC 実行になる条件でのみ発生。**ベース worktree で同一コマンドを実行し同一2件失敗を再現済み＝本差分非起因**）
- FE: 変更26ファイル scoped eslint `--max-warnings 0` 相当 / prettier / scoped tsc — 全 PASS（queryKeys リテラル直書き指摘は prefix factory 追加で是正済み）
- 回帰テスト追加: AddForm 実クリック3件、SpecialPeriod action経路2件、CampaignSidePanel per-field 3件、EstimateForm deceased 2件、ItemListCard 価格含む一括表示1件、DatePicker describedby 2件、MonthView/WeekView aria-label、use-pet-death invalidation 3件
