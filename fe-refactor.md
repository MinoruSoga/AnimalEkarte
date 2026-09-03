# frontend コード規約チェック結果（2026-09-03）

`frontend/` 配下（`src/`・`liff/src/`・`line-reserve/src/`、テスト・`components/ui/`・`types/generated/` 除く）を以下の規約に照合した結果。

- 規約正本: `frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`frontend/src/hooks/CLAUDE.md`、`frontend/src/components/shared/CLAUDE.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`.claude/refs/error-handling.md`、`frontend/CODING_RULES.md`（優先順位: 実コード → CLAUDE.md → CODING_RULES.md）
- 方法: (1) 規約項目ごとの機械検出（rg / スクリプト）を全域に実施、(2) 領域別に主要ファイルを精読（4 領域並列）、(3) HIGH 所見は実コードで再検証済み
- 各所見の `path` は `frontend/` 起点。行番号は調査時点のもの

---

## 0. サマリー

| 重要度 | 件数 | 概要 |
|---|---|---|
| CRITICAL | 0 | — |
| HIGH | 14 | 臨床安全境界の防壁欠落（4）、二重 toast（1・約 30 箇所）、疑似フォーム送信（5）、useEffect フェッチ（1）、lab-device 失敗黙殺（1）、動的 Tailwind クラス（1）、brand 色不一致（1） |
| MEDIUM | 45 | 権限再検査・cross-feature import・hooks 配置・memo 無効化・a11y focus・型キャスト・ファイル肥大・重複ロジック など |
| LOW | 40 | 命名・import 順序・マジックナンバー・コメント腐敗・ドキュメント乖離 など |

**機械検出で 0 件を確認した規約（合格）**: `any`（コード中 0、コメントのみ）／`FC`・`forwardRef`／arrow 関数 export コンポーネント／`&&` 条件レンダリング（1 件のみ→FE-RC-021）／raw hex（design-tokens・brand-tokens 以外）／`src/` 内 raw Tailwind パレット色クラス／`queryKey: [` 直書き／`export default`／`export *`／`__tests__/`・`utils/`・`.gitkeep`／`<form onSubmit>`／`useState(false)+setIsLoading` 送信管理／deep import（ESLint 許可の 1 件のみ）／層逆転（components/hooks/lib → features）／`dangerouslySetInnerHTML`／`localStorage` への token 保存／`tabIndex` 正値／`<tr onClick>`（C19）／非テスト 800 行超ファイル（documented exception の 2 件のみ）。

**推奨着手順**: ①臨床安全境界（FE-RC-001〜004）→ ②二重 toast の方針統一（FE-RC-005、機械的に直る）→ ③疑似フォーム 5 件の `useActionState` 化（FE-RC-006〜010）→ ④lab-device `onError`（FE-RC-012）→ ⑤動的 Tailwind クラス・brand 色（FE-RC-013, 014）→ ⑥cross-feature import と hooks 配置（FE-RC-015〜018）。

---

## 1. HIGH

### 臨床安全境界（frontend/CLAUDE.md「臨床安全境界」）

#### FE-RC-001 [HIGH] mutation 直前の権限再検査がなく UI の disabled/非表示が最終防壁になっている
- 対象:
  - `src/features/accounting/hooks/use-accounting-completion-action.ts:187-309` — `canEdit`/`canCreate` を一切参照せず `completeAccounting`/`updateAccounting` を実行。防壁は `AccountingDetail.tsx:239` の `<fieldset disabled={!canSubmit}>` のみ
  - `src/features/accounting/hooks/use-accounting-item-actions.ts:94-189` — `handleAddItem`/`handleDeleteItem` に権限引数なし（`canPostCloseEdit` のみ）
  - `src/features/accounting/components/CreditCorrectionDialog.tsx:52-85` — 親の render 条件のみ
  - `src/features/estimates/hooks/use-estimate-form.ts:106-171` — `usePermission` 結果は `EstimateForm.tsx:409` の `SubmitButton` 表示条件のみ
  - `src/features/clinic-settings/routes/use-clinic-master-settings.ts:55-77` — `canEdit`/`canCreate` を返すだけで action 内で未使用
- 規約: 臨床安全境界 2「権限は action 別の最新値を mutation 直前に再検査する。UI の非表示・disabled・route guard だけを最終防壁にしない」
- 改善案: `use-examination-form.ts:68-82` / `use-checkup-form.ts:53-65` と同じ `permissionsRef` + `useLayoutEffect` + `isMutationAllowed()` を action 冒頭に置く。

#### FE-RC-002 [HIGH] 死亡ペット拒否が callback 側のみで render 側防壁がない（保存ボタンが押せて無音失敗）
- 対象:
  - examinations: `use-examination-form-helpers.ts:380-382` で拒否するが `ExaminationForm.tsx:369-383` の `PatientInfoCard` に `status` なし、`:420-469` の fieldset は `canSubmit` のみ
  - checkups: `use-checkup-form.ts:88-90,95` で拒否するが `CheckupForm.tsx:80-84` の `submitDisabled` に死亡条件なし
  - hospitalization: `DailyRecordsTab.tsx:57-60`（callback）vs `:162-204`（`canCreate` のみで「追加」ボタン表示）、`CarePlanTab.tsx:39-44` vs `:110-113,140-142`、`HospitalizationDetailActions.tsx:46` vs `:42-43,59,70`（`showCheckIn = canEdit && isReserved`）
  - estimates: `use-estimate-form.ts:81-92,145-160` — `?petId=` から petId を採用して `createEstimate` するが status を見ない。`EstimateForm.tsx:299-307` の `useGetPet` 結果の `status` 未参照
- 規約: 臨床安全境界 1「死亡なら要素自体をレンダリングしない ＋ callback 側も拒否（二重防壁）。新しい pet 操作を追加するときはこの二重防壁に揃える」（参照実装 `OwnerPetsSection.tsx:108,162`）
- 改善案: `isPetExplicitlyDeceased` を render にも流し `canCreate && !petIsDeceased` で `SubmitButton`/`AddForm`/チェックインを非表示にし、accounting の `deceasedPetBlockMessage`（`use-accounting-detail-state.ts:84-91`）と同様の理由表示を出す。callback ガードは維持。

#### FE-RC-003 [HIGH] 予防接種フォームの日付バリデーションがローカル TZ の `new Date()` 比較
- 対象: `src/features/vaccinations/hooks/use-vaccination-form-model.ts:90-113`
- 規約: 臨床安全境界 3「臨床 date-only は JST の厳密過去で判定。`todayJSTISO()` との文字列比較を使う」
- 現状: `const today = new Date(); today.setHours(0,0,0,0); ... new Date(formData.date + "T00:00:00") > today` — 端末 TZ が JST 以外だと当日接種を「未来」と誤判定。同ファイル `:82,:146` は既に `todayJSTISO()` を使っており不整合
- 改善案: `const today = todayJSTISO(); if (formData.date > today) ...; if (formData.nextDate < today) ...` の文字列比較へ。同型が `src/components/shared/ReservationFormModal/ReservationFormModal.tsx:208-219,240-259` にもある（FE-RC-041 と併せて修正）。

#### FE-RC-004 [HIGH] 拒否（死亡・権限なし・確定済み）時に UI へ何も出さない無音失敗
- 対象: `use-examination-form-helpers.ts:380-382,408-413,424-429`、`use-checkup-form.ts:88-127`、`use-estimate-form.ts:116-118`、`DailyRecordsTab.tsx`（`isMutationAllowed()` false で無反応）
- 規約: error-handling.md「UI state を loading/success/error として表現する」
- 現状: `return { success: false, timestamp: Date.now() };` のみで toast も fieldError もない
- 改善案: `ActionState.error` に理由を格納して表示、または toast.error。FE-RC-002 の render 側ブロックを入れれば発生頻度は下がるが、callback 側も理由を返すこと。

### エラー処理

#### FE-RC-005 [HIGH] api 層 `onError`/`onSuccess` と呼び出し側 `catch`/`onError` の二重 toast（約 30 箇所）
- 規約: error-handling.md「同じ failure を複数層で重複ログしない」。`MedicalRecordEstimate.tsx:189`（「mutation onError が handleApiError 済み。ここでは再通知しない」）と `HospitalizationDetailActions.tsx:52-54` が正例
- 対象（api 側 → 呼び出し側）:
  - estimates: `api/update-estimate.ts:28,31`, `api/create-estimate.ts:22,25` → `hooks/use-estimate-form.ts:136,161,166`（**成功 toast も 2 回**）
  - examinations: `api/update-examination.ts:32`, `api/create-examination.ts:26` → `hooks/use-examination-form-helpers.ts:433-434`
  - clinic-settings: `api/clinics.ts:108,120,131` → `routes/use-clinic-master-settings.ts:71-72,120-122`
  - closing-settings: `api/holidays.ts:45,54` → `HolidaySection.tsx:28-29,39-40`; `api/special-periods.ts:34,43` → `SpecialPeriodSection.tsx:35-36,46-47`
  - cash-register: `api/create-cash-register-close.ts:32` → `routes/CashRegisterClosePage.tsx:76-77`
  - master: 全 `api/*.ts` の `onError` → `hooks/use-master-save.ts:104-110`、`hooks/use-master-crud.ts:255-261`（`mutate(..., { onError })`）、`routes/use-medicine-settings.ts:139-145`
  - inventory: `api/inventory.ts:101,117` → `hooks/use-inventory-form.ts:100-109`
  - hospitalization: `api/daily-records.ts:100-154` → `DailyRecordsTab.tsx:93-135`; `api/update-hospitalization.ts:58-59` → `hooks/use-hospitalization-detail.ts:41-51`; `api/delete-hospitalization.ts` → `routes/HospitalizationForm.tsx:45-50`
  - shifts: `api/clinic-holidays.ts:42,55` → `ClinicHolidayModal.tsx:45-46,56-57`
  - owners: `api/send-line-message.ts:46-48` → `LineSendPanel.tsx:71-72,97-98`; `api/update-owner-line.ts:48-50` → `hooks/use-line-integration-card-state.ts:70-71`; `api/create-owner-tag.ts:39-40` → `LstepTagAddDialog.tsx:54-55`
  - vaccinations: `api/delete-vaccination.ts:18-19` → `routes/VaccinationList.tsx:130-132`
  - trimming: `api/delete-trimming.ts:18` → `routes/TrimmingList.tsx:127-129`
  - reception: `api/update-appointment-status.ts:27` → `hooks/use-reception-kanban.ts:284-285`
  - reservations: `api/delete-reservation.ts:18` → `hooks/use-reservation-actions.ts:374-376`
  - medical-records: `api/billing-confirmation.ts:51,77` → `MedicalRecordBillCheck.tsx:135-136,150`
  - settings: `hooks/use-lstep-settings.ts:56,69,83` → `LstepSettingsForm.tsx:57-84`; `hooks/use-trigger-priorities.ts:53` → `TriggerPrioritySection.tsx:71-72`
- 改善案: 方針を 1 つに決める。(a) api hook の `onError`/`onSuccess` に通知を集約し、呼び出し側 catch は状態遷移のみ＋「onError 済み」コメント、または (b) owners 参照実装（api は生関数のみ、通知は `useActionState` 側）に揃えて api 側の toast を全削除。CODING_RULES §1.4 パターン A/B の使い分けを「通知責務」まで明文化する。

### フォーム（frontend/CLAUDE.md React 19 Patterns: `useActionState` + `<form action>` + `SubmitButton`）

#### FE-RC-006 [HIGH] `ClinicHolidayModal` が `useTransition` + `onClick` で送信
- 対象: `src/features/shifts/components/ClinicHolidayModal/ClinicHolidayModal.tsx:27-49,117-125`
- 現状: `const [isSaving, startSaveTransition] = useTransition(); <Button type="button" onClick={handleSave}>`
- 改善案: `useActionState` + `<form action>` + `SubmitButton`。解除（削除系）は `useTransition` のままで可。

#### FE-RC-007 [HIGH] `UnlinkedLineIdForm` が `startTransition(formAction)` の疑似フォーム
- 対象: `src/features/owners/components/LineIntegrationCardParts.tsx:153-191`
- 現状: `new FormData()` を手組みして `startTransition(() => lineIdFormAction(payload))`、`onKeyDown` で Enter を手書き
- 改善案: `lineIdFormAction` を `<form action>` に直結し `SubmitButton` へ。`isPending` の二重管理と `onKeyDown` を削除。

#### FE-RC-008 [HIGH] `MedicalRecordVaccination` の接種追加が `onSave` ボタン + `useTransition`（`useState` 12 個）
- 対象: `src/features/medical-records/components/MedicalRecordVaccination.tsx:73,129-183`、`VaccinationForm.tsx:183-187`
- 改善案: バリデーション・送信を `useActionState` に集約し `VaccinationForm` を `<form action>` でラップ。12 個の `useState` は `hooks/use-vaccination-entry-form.ts` へ。

#### FE-RC-009 [HIGH] `TriggerPrioritySection` / `LstepTagConfigSection` の保存・追加が `onClick` + `mutate`
- 対象: `src/features/settings/components/TriggerPrioritySection.tsx:58-74,127-134`、`LstepTagConfigSection.tsx:110-127,190-198,216-233,318 付近`（3 セクション同型）
- 現状: `if (!trimmedPrefix || !trimmedCategory) { toast.error("...必須です"); return; } createMutation.mutate(...)`
- 改善案: `<form action>` + 非制御 `<input name>` + `SubmitButton`。必須エラーは `fieldErrors` + `FormFieldError` で inline 表示（toast でのバリデーション通知をやめる）。3 セクションは `TagPairAddForm` に統合（FE-RC-050 と連動）。

#### FE-RC-010 [HIGH] `RefundSection` / `EstimateDetail`（後継ドラフト作成）が `onClick` 送信
- 対象: `src/features/accounting/components/RefundSection.tsx:59-73,161-167`、`src/features/estimates/routes/EstimateDetail.tsx:66-86` + `EstimateDetailPanels.tsx:245-251`
- 現状: `<Button type="button" onClick={handleSubmit}>{isRefunding ? "処理中..." : "登録する"}</Button>`、`startSuccessorTransition(() => createSuccessor(...))`
- 改善案: Dialog 内を `<form action>` + `SubmitButton` にし、バリデーション/送信を `useActionState` 内へ。

### Hooks / データ取得

#### FE-RC-011 [HIGH] `useEffect` でのデータフェッチ（`useQuery` 不使用）＋ render 中 setState の手組み
- 対象: `src/features/identity-links/routes/IdentityLinksPage.tsx:81-128`
- 現状: `useEffect(() => { let cancelled = false; void searchOwnersForLink(trimmedOwnerQuery).then(...) })` ＋ `prevTrimmed*Query` state でヒットをクリア
- 改善案: `useQuery({ queryKey: queryKeys.identityLinks.ownerSearch(q), queryFn, enabled: q.length > 0 })` に置換。`ownerHits`/`petHits`/`prev*` state はすべて不要になる。FE-RC-047（487 行の分割）と同時に実施。

#### FE-RC-012 [HIGH] lab-device mutation の失敗が黙殺される（`onError` なし＋ `void mutateAsync()`）
- 対象: `src/features/lab-device/api/lab-device.ts:258-322`（`usePutLabDeviceWait`/`useClearLabDeviceWait`/`useReceiveLabDeviceFrames`/`useAttachLabDeviceJob`/`useDetachLabDeviceJob` すべて `onError` なし）、呼び出し側 `routes/LabDeviceBoard.tsx:131,143,150,157-158`、`routes/lab-device-board-panels.tsx:336,376`、`components/LabDeviceUnlinkedBanner.tsx:53,81`
- 規約: error-handling.md「promise rejection を放置しない」
- 現状: `onClearWait={() => void clearWait.mutateAsync()}` — 403/409/ネットワーク失敗時に toast も UI 表示もなく unhandled rejection
- 改善案: 各 mutation に `onError: (e) => handleApiError(e, "…")` を追加し、呼び出し側は `mutate()`（必要なら `mutate(vars, { onSuccess })`）へ。

### デザイントークン

#### FE-RC-013 [HIGH] Tailwind クラスをテンプレート文字列で実行時合成（CSS が生成されない）
- 対象: `src/features/lstep/components/LstepCsvImportSection.tsx:107,110,172,174`、`src/features/line-reservation/components/LinkedLineCustomers.tsx:199`、`src/components/shared/DatePicker/DatePickerSingle.tsx:156`（`placeholder:${C.text40}`）
- 規約: Design Tokens（`C.*`/`STYLE.*` をそのまま使う）。Tailwind v4 はソース静的走査のため `text-[${PALETTE.danger}]` / `hover:${C.bgHover}` は CSS に含まれない
- 現状: `<p className={`text-sm text-[${PALETTE.danger}]`}>{state.error}</p>` — エラー文が赤くならない可能性が高い
- 改善案: `C.danger` / `C.textStatusGreen` / `C.bgStatusGreen` / `C.hoverBgLight` / `C.textPlaceholder` の既存トークンへ置換。再発防止として `design-system-audit.mjs` に `\$\{[^}]+\}` を含むクラス名（`-[${`、`:${`）の検出を追加。

#### FE-RC-014 [HIGH] shared-liff の brand teal が design-system の brand 色と不一致
- 対象: `src/shared-liff/brand-tokens.ts:6-7`（`teal: '#008B94'`, `tealDark: '#007079'`）、`brand-tokens.css:3-4,11-12` vs `src/lib/design-tokens.ts:35,37`（`#038B94` / `#027078`）、`docs/spec/design-system.md:17`
- 規約: frontend/CLAUDE.md「色の正本は `docs/spec/design-system.md`。brand は `#038B94` / active `#027078`」
- 現状: R 成分が `00` vs `03`、`00` vs `02` の 1 桁差で、意図的な別色ではなくタイポ由来の drift
- 改善案: 4 箇所を `#038B94`/`#027078` に揃え、`brand-tokens.test.ts` に `PALETTE.brand` との等値アサーションを追加。

---

## 2. MEDIUM

### 構造・import 境界

#### FE-RC-015 [MEDIUM] feature 間 import（barrel 経由）8 箇所
- 対象:
  - `src/features/examinations/routes/ExaminationForm.tsx:39-40` → `@/features/master`（`useGetStaffs`）, `@/features/lab-device`（`LabDeviceUnlinkedBanner`）
  - `src/features/clinic-settings/components/CompanyInvoiceSection.tsx:8` → `@/features/master`（`useGetCompany`, `useUpdateCompany`）
  - `src/features/hospitalization/components/CarePlanTab/CarePlanRefSelect.tsx:6` → `@/features/master`
  - `src/features/medical-records/components/MedicalRecordExamination.tsx:20` → `@/features/lab-device`
  - `src/features/medical-records/components/CheckupsTab/{CheckupsTab.tsx:10-15, CheckupsTabRows.tsx:19, CheckupsTabTable.tsx:7}` → `@/features/checkups`
  - `src/features/owner-report/hooks/use-owner-clinical-briefing-data.ts:2` → `@/features/medical-records`
- 規約: CODING_RULES §1.2/§9「feature間の直接importは禁止（app/pages/ で合成、または components/shared に移動）」。一方 `frontend/CLAUDE.md` の ESLint 境界は deep import のみ禁止で barrel 経由は通る → **規約文書間の不整合**（FE-RC-060 参照）
- 改善案: `useGetStaffs` は既存 `@/hooks/use-staffs` へ寄せる。`LabDeviceUnlinkedBanner` は `components/shared/` へ、company API・checkups の field/payload helper・`useGetMedicalRecords` は `src/hooks`/`src/lib` へ昇格。並行して ESLint の `no-restricted-imports` に「`src/features/<a>/**` から `@/features/<b>`」を機械禁止するか、規約側を「barrel 経由は許容」に改訂して決着させる。

#### FE-RC-016 [MEDIUM] hook が `routes/`・`components/` に配置（features/CLAUDE.md: hooks は `hooks/`）
- 対象（11 件）: `src/features/master/routes/use-{medicine-settings,lab-device-item-master-settings,reservation-type-settings,staff-settings-lookups,treatment-plan-master-resources,treatment-plan-master-saves}.ts`、`src/features/master/components/use-exam-type-field-{session,fields-list}.ts`、`src/features/clinic-settings/routes/use-clinic-master-settings.ts`、`src/features/reception/routes/use-reception-column-view.tsx`、panels ファイル同居 hook: `hospitalization/routes/hospitalization-form-panels.tsx:42`（`useHospitalizationFormChrome`）、`trimming/routes/trimming-form-panels.tsx:38`（`useTrimmingFormChrome`）、`medical-records/routes/medical-records-columns.tsx:32`（`useMedicalRecordsColumns`）
- 現状: `// eslint-disable-next-line react-refresh/only-export-components -- 150行分割で page chrome hook を panels と同居`
- 改善案: `hooks/` へ移動（`eslint-disable` も消える）。JSX を返す `useReceptionColumnView` はコンポーネント化を検討。

#### FE-RC-017 [MEDIUM] コンポーネントファイルが kebab-case（features/CLAUDE.md: PascalCase.tsx）
- 対象（`src/components/ui/`・`app/routes/`・hook `.tsx` を除く 30 件超）: `*-form-panels.tsx`（checkups/inventory/hospitalization/trimming/vaccinations/medical-records）、`cash-register/components/cash-register-*.tsx`（4）、`master/components/exam-type-field-*.tsx`（3）+ `day-of-week-select-items.tsx`、`master/routes/treatment-plan-master-view.tsx`、`manual/components/manual-page-chrome.tsx`、`shifts/components/shift-template-settings-*.tsx`（2）、`auth/components/login-form-sections.tsx`、`auth/routes/reset-password-page-sections.tsx`、`accounting/components/unpaid-tab-*.tsx`（3）、`owners/components/pet-edit-field-shared.tsx`、`medical-records/components/TreatmentsTab/treatment-row-{editors,quantity-cell}.tsx`、`components/shared/Layout/sidebar-{menu,chrome}.tsx`、`components/shared/ReservationFormModal/patient-selection-{filters,results}.tsx`
- 補足: `scripts/check-feature-filename-convention.mjs` は `.ts` のみ対象で `.tsx` は未検査。`LstepCsvImportSection.tsx:17` は export 名が `CsvImportSection` でファイル名と不一致
- 改善案: PascalCase へリネーム（hook は FE-RC-016 で `hooks/` へ分離）。ratchet スクリプトの対象に「`.tsx` で JSX コンポーネントを export するのに basename が小文字始まり」を追加。

#### FE-RC-018 [MEDIUM] `src/hooks` に単一 feature 専用フック（hooks/CLAUDE.md 配置規約）
- 対象: `src/hooks/use-pet-checkup-results.ts`（消費は `owner-report` のみ）、`src/hooks/use-reservation-type-color-map.ts`（`reservations` の 3 ファイルのみ）
- 改善案: それぞれ `features/owner-report/hooks/`、`features/reservations/hooks/` へ移動（テスト同伴）。※ `use-vaccinations.ts` は medical-records と vaccinations の 2 feature 消費で現配置は正当。

#### FE-RC-019 [MEDIUM] `AuthProvider`（コンポーネント）が `hooks/use-auth.tsx` に同居し Fast Refresh 抑制が必要になっている
- 対象: `src/features/auth/hooks/use-auth.tsx:46,69`、`eslint.config.js:227-231`、`src/features/auth/hooks/use-permission.ts`（3 行の re-export）、`src/features/auth/provider.ts`（1 行）
- 改善案: `features/auth/components/AuthProvider.tsx` へ分離。`features/auth/index.ts:2` を `@/hooks/use-permission` 直参照にして `hooks/use-permission.ts` を削除。`provider.ts` は Login 群を lazy に保つための境界として妥当（FE-RC-061 参照）だが、`index.ts:1` からも `AuthProvider` が export され経路が 2 本あるため片方に統一し CLAUDE.md に明記。

#### FE-RC-020 [MEDIUM] `CheckupFieldType` の三重定義と stale コメント
- 対象: `src/hooks/use-pet-checkup-results.ts:6-16`、`src/features/checkups/api/get-checkup-type-fields.ts:6-14`、`src/types/generated/models.ts:669-675`
- 現状: コメント「generated の CheckupFieldType は string 型で緩いため使用しない」は誤り（generated は厳密 union）。`use-pet-checkup-results.ts:78-80`「features/checkups と同じ query key でキャッシュ共有」も実キー `["pet-checkup-results","report",petId]` と不一致
- 改善案: generated 型を 1 箇所（`types/checkup.ts`）で re-export して両者から import。コメントを削除。

### 条件レンダリング

#### FE-RC-021 [MEDIUM] `&&` による条件レンダリング（残 1 ファイル）
- 対象: `src/features/identity-links/routes/IdentityLinksPage.tsx:347-351,388,442`
- 規約: frontend/CLAUDE.md Conditional Render（`react/jsx-no-leaked-render` は boolean 左辺を許すため機械検出を抜けている）
- 現状: `{!canEdit && (<p ...>)}`、`{ownerGroupId != null && ` / 連携グループ #${ownerGroupId}`}`
- 改善案: `? : null` へ。ESLint ルールを `["error", { validStrategies: ["ternary"] }]` に強化すれば boolean 左辺も検出できる。

### フォーム（続き）

#### FE-RC-022 [MEDIUM] 入力＋保存ボタンの `onClick` 直接 mutate（`<form>` なし）
- 対象: `hospitalization/components/CarePlanTab/AddForm.tsx:60-73,119-124`、`EditRow.tsx:68-77,126-131`、`DailyRecordsTab/DailyVitalsSection.tsx:66-80,248-254`、`master/components/MedicineDoseParamsEditor.tsx:160-162,312`、`master/components/exam-type-field-draft-panel.tsx:67-73,156-162`、`manual/components/ManualEditor.tsx:48-64,162-165`、`cash-register/routes/CashRegisterClosePage.tsx:52-80,145-150`（確認ダイアログの実 mutation が `AlertDialogAction onClick` で pending 中も disabled にならない）、`line-reserve/src/pages/CustomerInfoPage.tsx:129-145,336`
- 改善案: 各パネルを `<form action={formAction}>` に包み `SubmitButton` へ。`isSubmitting`/`isSaving` props は `useFormStatus` に置換。

#### FE-RC-023 [MEDIUM] 生 `<Button type="submit">`（`SubmitButton` 不使用）と `SubmitButton` の variant 不足
- 対象: `src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx:379-385`、`src/components/shared/PetDeceasedRecordButton/PetDeceasedDialog.tsx:212-219`（`form="pet-deceased-form"` で form 外に配置、`destructive` variant が `SubmitButton` にない）
- 改善案: `SubmitButton` に `colorVariant: "destructive"` を追加し両者を置換。`PetDeceasedDialog` は `<form>` を `DialogFooter` まで包む。

#### FE-RC-024 [MEDIUM] line-reserve のキャンセル pending を `useState` + try/finally で手動管理
- 対象: `line-reserve/src/pages/MyReservationsPage.tsx:50-52,71-85`
- 規約: CODING_RULES §9（フォーム外の非同期は `useTransition`）
- 改善案: `useTransition` へ。対象 ID の識別は `pendingId` として残してよい。

### コンポーネント設計

#### FE-RC-025 [MEDIUM] memo 子に毎レンダー新規のハンドラ／オブジェクトを渡し memo を無効化
- 対象: `medical-records/components/MedicalRecordBillCheck.tsx:262`（`TreatmentTable` は memo）、`TreatmentsTab/TreatmentsTab.tsx:355,370-371`、`trimming/routes/trimming-form-panels.tsx:342,351`、`hospitalization/routes/hospitalization-form-panels.tsx:342,350,359`（`fieldErrors={{ cage_id: ... }}` を毎回生成、`HospitalizationNoteCard` は memo）、`hospitalization/hooks/use-hospitalization-form.ts:42-44`（`handleFormDataChange` 非 useCallback）、`manual/components/ManualEditor.tsx:91-95`（memo の `ManualContent` に毎回新規 `previewArticle`）、`components/shared/ReservationFormModal/ReservationFormModal.tsx:321-338`（memo の `ReservationFormFields` へ inline arrow）
- 規約: CODING_RULES §12.1 `rerender-memo`
- 改善案: `useCallback`（functional setState で deps 空）に統一。`fieldErrors` はプリミティブ `cageIdError` で渡す。

#### FE-RC-026 [MEDIUM] `useMutation` 戻り値オブジェクト全体を deps に入れている
- 対象: `hospitalization/routes/HospitalizationForm.tsx:53`、`CarePlanTab.tsx:73,86,94`、`DailyRecordsTab.tsx:98,111,125,139`、`master/hooks/use-master-crud.ts:263`、`use-master-save.ts:140`、`line-reservation/components/LinkedLineCustomers.tsx:54,64`、`master/routes/use-medicine-settings.ts:147`
- 規約: CODING_RULES §12.1 `rerender-dependencies`
- 改善案: `const { mutate } = useXxx()` と分解し安定参照の `mutate`/`mutateAsync` のみ deps に置く。

#### FE-RC-027 [MEDIUM] 完全コピペ重複
- `StatusBadge`: `examinations/components/ExamPivotTable.tsx:130-176` と `ExamItemsTable.tsx:44-93`（ExamPivotTable 側は `return null` が二重）→ `components/ExamStatusBadge.tsx` に統合
- `requiresRef`/`buildRefFields`: `hospitalization/components/CarePlanTab/AddForm.tsx:22-34` と `EditRow.tsx:20-32` → `care-plan-item-model.ts` へ
- `isValidOwnerPhone`（`components/shared/ReservationFormModal/ReservationFormModal.tsx:48-52`）と `isBackendCompatiblePhone`（`line-reserve/src/pages/CustomerInfoPage.tsx:8-12`）が同一正規表現 → `src/lib/phone.ts`
- マスタ選択→`item_type` 変換のネスト三項: `MedicalRecordBillCheck.tsx:200` と `MedicalRecordDiagnosisPlan.tsx:139` → `treatments-tab-model.ts` の既存 `buildMasterSelectionPayload` を再利用
- JST 手計算: `trimming/hooks/trimming-form-utils.ts:126-133` と `medical-records/hooks/use-medical-record-form-model.ts:46-52` が `lib/jst-date.ts` を再実装

#### FE-RC-028 [MEDIUM] URL ページ clamp effect が 7 ファイルに複製（`exhaustive-deps` 抑制付き）
- 対象: `accounting/routes/AccountingList.tsx:99-109`、`examinations/routes/ExaminationsList.tsx:75-84`、`checkups/routes/CheckupsList.tsx:79-87`、`trimming/routes/TrimmingList.tsx:79-100`、`vaccinations/routes/VaccinationList.tsx:87-103`、`inventory/routes/InventoryList.tsx:92`、`hospitalization/routes/HospitalizationList.tsx:101`
- 改善案: `src/hooks/use-url-page-sync.ts` に集約（7 件の `eslint-disable` も 1 箇所に）。

#### FE-RC-029 [MEDIUM] master `*SidePanel` 22 ファイルに同一ボイラープレート＋保存結果を待たず dirty クリア
- 対象: `src/features/master/components/*SidePanel*.tsx`（`isDirty`/`setFormDataDirty`/`handleTitleChange`/`handleToggleActive`/`handleClose` が 22 箇所同型）。うち 14 件（`DiagnosisNameSidePanel.tsx:85-86` 他）は `onSave(current); setIsDirty(false);` と保存結果を待たず、失敗時に未保存変更が「保存済み」扱いになる。`PaymentMethodSidePanel`/`InsuranceSidePanel` は BUG-026/029 で `Promise<boolean>` 化済みで不整合
- 改善案: `hooks/use-master-side-panel-form.ts` に集約し、`onSave` を `Promise<boolean>` に統一して `saved === true` のときだけ `setIsDirty(false)`。

### Hooks

#### FE-RC-030 [MEDIUM] mutation callback から参照する ref を `useEffect` で同期（`useLayoutEffect` 規約）
- 対象: `accounting/hooks/use-accounting-completion-action.ts:182-185`（死亡ペット拒否理由 ref）、`master/hooks/use-master-crud.ts:181`、`lstep/components/CheckupSyncPreviewTable.tsx:107-109`、`master/components/DiagnosisNameSidePanel.tsx:51-53`、`master/components/use-exam-type-fields-list.ts:31-33`、`master/components/ReservationTypeGroupedTable.tsx:60-62`、`src/hooks/use-side-peek-dirty.tsx:37-39`
- 規約: 臨床安全境界 2「commit 直後にも発火し得る callback の ref は `useLayoutEffect` で同期」
- 改善案: `useLayoutEffect` に統一（`use-examination-form.ts:69-77` と同パターン）。

#### FE-RC-031 [MEDIUM] `NavigationBlocker` の cleanup が stale closure で常に no-op
- 対象: `src/components/shared/NavigationBlocker/NavigationBlocker.tsx:37-44`
- 現状: `useEffect(() => () => { if (blocker.state === "blocked") blocker.reset(); }, [])` — `blocker` は mount 時の値で固定され `blocked` を検知できない（react-router 側の自動 cleanup で実害は隠れている）
- 改善案: `blockerRef` 経由で参照、または react-router の cleanup に委ねてブロックごと削除し抑制コメントも消す。

#### FE-RC-032 [MEDIUM] `set-state-in-effect` 抑制（27 箇所）のうち React 19 慣用で置換可能なもの
- 対象:
  - `components/shared/ReservationFormModal/ReservationFormModal.tsx:124-160`（開くたびに 9 state を初期化 → `key` リセット + lazy init）、`:180-188`（→ render 中派生）
  - `settings/components/TriggerPrioritySection.tsx:33-41`（→ `OwnersList.tsx:163-168` の「render 中 prev 比較」パターン）
  - `medical-records/components/TreatmentsTab/treatment-row-editors.tsx:40-43,88-91,153-156,218-221`、`treatment-row-quantity-cell.tsx:333`（deps が `treatment` オブジェクト全体。`[treatment.content]` へ絞る、または編集開始時初期化で effect 自体を削除）
  - `shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:87-94`（→ 呼び出し側で `key` 付与）
  - `examinations/hooks/use-examination-form-helpers.ts:89-221`（6 本の effect で ref と state を相互同期、`formItemsRef.current = rows; setFormItems(rows)` を 4 箇所反復 → 導出を `useMemo` 化）
  - `liff/src/hooks/use-liff-link.ts:22-88`（block 抑制・`-- 理由` なし。StrictMode で `POST /link` が 2 回飛び得る → `linkPromiseRef.current ??=` で冪等化）
- 改善案: 上記から着手し `.eslint-disable-baseline`（現 22）を締める。

#### FE-RC-033 [MEDIUM] ペット未選択リダイレクト effect の二重実装
- 対象: `examinations/routes/ExaminationForm.tsx:300-304` と `hooks/use-examination-form-helpers.ts:320-329`（条件式が微妙に異なる）
- 改善案: `useExaminationFormPetSync` に一元化。

### エラー処理（続き）

#### FE-RC-034 [MEDIUM] error message 文字列による分岐
- 対象: `accounting/hooks/use-accounting-completion-action.ts:49-52`（`.includes(POST_CLOSE_REASON_MARKER)`）、`aggregation/routes/aggregation-dashboard-model.ts:60`（`error.message.startsWith("Request failed")`）、`src/lib/handle-api-error.ts:109-126,171-172`（`/^(\w+) '(.*)' already exists$/i` — 26 資源中 `CONFLICT_CODE_*` は 12 個のみ）
- 規約: error-handling.md「API error code を型付けし、message 文字列による分岐を避ける」
- 改善案: BE が全 409/422 に `code` + `params` を返すよう拡張し、`localizeAlreadyExistsMessage` を deprecate。移行中は message フォールバック到達を DEV で `console.warn`。

#### FE-RC-035 [MEDIUM] シリアル読み取りループの `catch {}` が原因情報を破棄
- 対象: `src/features/lab-device/lib/lab-device-serial.ts:329-331`
- 現状: `catch { notifyState("disconnected"); }` — `open()` 失敗と読み取り中断を区別できず UI は「停止中」しか示せない。リソース解放（`releaseLock`/`close`/`clearTimeout`）自体は `:185-201,257-263` で網羅
- 改善案: `catch (error: unknown)` で `onError?.(error)` を追加し、`LabDeviceListenState` に `reason` を持たせる。

### 型安全

#### FE-RC-036 [MEDIUM] `formData.get("x") as string` の未検証キャスト（`null` で TypeError）
- 対象: `auth/components/LoginForm.tsx:91-92`、`auth/routes/ResetPasswordPage.tsx:52-53`、`ForgotPasswordPage.tsx:25`、`cash-register/routes/CashRegisterClosePage.tsx:67-70`（`as CashRegisterPeriod` 含む）、`closing-settings/components/{StandardClosingTimeSection.tsx:50-52, SpecialPeriodSection.tsx:26-30, HolidaySection.tsx:23-24}`、`owners/hooks/use-line-integration-card-state.ts:63`、`owners/components/LstepTagAddDialog.tsx:37`、`lstep/components/CheckupSyncFilterForm.tsx:57-120`、`line-reservation/components/LineReservationSettingsForm.tsx:126-136`（`Number(...)` の NaN 未処理）、`components/shared/PetDeceasedRecordButton/PetDeceasedDialog.tsx:59-60`
- 規約: frontend/CLAUDE.md Type Safety（unknown + 型ガード）。CODING_RULES §2.2 の例示自体が `formData.get("name") as string` なので規約例も直す
- 改善案: `src/lib/form-data.ts` に `getFormString(fd, key): string` / `getFormEnum(fd, key, guard)` を置き全箇所で使用（`AddendumModal.tsx:29-30` が良例）。

#### FE-RC-037 [MEDIUM] enum・外部データへの無検証 `as` キャスト
- 対象: `shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:98`（`as ShiftType`）、`shifts/api/get-shift-templates.ts:43`、`owners/components/LineSendPanel.tsx:60`（`as LineSendType`）、`medical-records/api/transforms.ts:70,72`、`accounting/hooks/use-accounting-item-actions.ts:103,109`、`accounting/api/transforms.ts:17,59,85,106`、`estimates/api/transforms.ts:33`、`lab-device/lib/lab-device-serial.ts:94`（`JSON.parse(raw) as LabDeviceSerialPortInfo` — localStorage）、`lab-device/api/lab-device.ts:205`、`hospitalization/hooks/use-hospitalization-list.ts:33-36`（QueryCache）、`hospitalization/components/HospitalizationBoard.tsx:177,180`、`inventory/hooks/use-inventory-form.ts:60,68`、`src/lib/handle-api-error.ts:145`（`err.response?.data as ApiErrorBody`）
- 改善案: `isShiftType(v): v is ShiftType` 等の型ガードを transform 層に 1 箇所置く。外部由来（localStorage/HTTP body）は zod `safeParse`（`line-reserve/src/pages/ConfirmPage.tsx:15-19`、`line-reservation-settings-form-model.ts:31-43` が先例）。

#### FE-RC-038 [MEDIUM] 非 null アサーション `!`
- 対象: `cash-register/components/ClosePrintArea.tsx:173`、`accounting/components/DailyAccountingPrintArea.tsx:70`（`paymentSplits!`）、`line-reservation/api/update-owner-link.ts:28,30`（`clinicId!` を mutation 実行時に）、`line-reservation/api/get-line-{customers,reservation-setting}.ts:16-17`、`hospitalization/api/get-{treatment-plans.ts:23, hospitalization.ts:40}`、`lstep/components/lstep-analytics-model.ts:63`、`master/routes/lab-device-item-master-settings-model.ts:117,130,249`、`medical-records/components/MedicalRecordFormPanels.tsx:151`（`onDateChange!`）、`medical-records/api/get-medical-record-images.ts:46`、`src/hooks/use-pet-{checkup-results.ts:83-84, vaccinations.ts:72-73}`、`src/hooks/use-reservation-types.ts:157-197`（`petId!` + `enabled` ガード）
- 改善案: TanStack v5 `skipToken` で `enabled` + `!` の組を排除。`map.get() ?? init` パターン、`?? []`。

#### FE-RC-039 [MEDIUM] ジェネリック型付けを `as` で潰すハンドラ
- 対象: `estimates/routes/EstimateForm.tsx:350` `(handleChange as (k: string, v: unknown) => void)(key, value)` — `use-estimate-form.ts:173-178` の `<K extends keyof EstimateFormState>` を無効化
- 改善案: セクション props も同じジェネリック型か、フィールド別 `onTitleChange` へ分解。

### a11y

#### FE-RC-040 [MEDIUM] `outline-none` に可視フォーカス代替がない（WCAG 2.4.7）
- 対象:
  - 代替なし: `components/shared/SidePeek/SidePeekTitleInput.tsx:49`、`clinic-settings/components/ClinicMasterSidePanel.tsx:76`、`shifts/components/ShiftTemplateSettingsParts.tsx:194`、`manual/components/ManualEditor.tsx:230`、`settings/components/LstepTagConfigSection.tsx:176,187,274,285,372,383`、`LstepTagCodeMappingsSection.tsx:120,131`（`STYLE.formInput` は focus スタイルを含まない）
  - 背景 4% 変化のみで hover と区別不能: `components/shared/SidePeek/PropertyInput.tsx:25`、`MoneyInput.tsx:31`、`shifts/components/ShiftTemplateSidePanelFields.tsx:220`、`clinic-settings/components/ClinicMasterSidePanelProperties.tsx:12`、`master/components/Trimming{Option,Course}SidePanel.tsx:132/179`（`C.focusBgLight` = `focus:bg-[rgba(0,0,0,0.04)]`）
  - 親で担保済み（問題なし）: `DatePickerSingle.tsx:156`（`TRIGGER_BASE` の `focus-within:ring-1`）
- 改善案: `STYLE.sidePeekInput` を design-tokens に定義して `focus-visible:ring-2 ${C.focusRingAccent40}` を含め、上記を一括置換。`STYLE.formInput` にも focus スタイルを含める。

#### FE-RC-041 [MEDIUM] `<Label>` / `SelectTrigger` / `Input` にアクセシブルネームがない
- 対象: `accounting/components/RefundSection.tsx:118-140`（`Label` 3 箇所に `htmlFor` なし）、`PaymentCard.tsx:119,207-215`、`hospitalization/components/CarePlanTab/EditRow.tsx:85,90-95`（`AddForm.tsx` は `aria-label` あり・不整合）、`SelectTrigger` 計 31 箇所（`DailyCareLogsSection.tsx:203`、`master/components/{HospitalizationSidePanel.tsx:116,130, MerchandiseSidePanel.tsx:97, DiagnosisNameSidePanel.tsx:124, ReservationTypeUnavailableTimesSection.tsx:143-190, ReservationTypeSidePanel.tsx:113,175, MedicineSidePanelSections.tsx:198-258, MedicineDoseParamsEditor.tsx:196,281, ReservationTypeOccupationsSection.tsx:86}`）
- 規約: accessibility-rules.md §2/§3
- 改善案: `PropertyRow label` 内なら `id` + `htmlFor`、それ以外は `aria-label`。

#### FE-RC-042 [MEDIUM] アイコンのみボタンに `aria-label` がない／小画面でラベルが `hidden`
- 対象: `clinic-settings/components/ClinicMasterSidePanel.tsx:53-59`（`<X />` 閉じる）、`shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:326-334`、`manual/components/ManualEditor.tsx:132-203`（`<span className="hidden sm:inline">` は a11y ツリーからも消える）
- 改善案: `aria-label` を常時付与、または `sr-only sm:not-sr-only`。

#### FE-RC-043 [MEDIUM] 受付カンバンの DnD に `KeyboardSensor` がない
- 対象: `src/features/reception/routes/Reception.tsx:75-77`（`PointerSensor` のみ）
- 規約: accessibility-rules.md §4。shifts 側（`shift-template-settings-list.tsx:49`）は既定でキーボード有効
- 改善案: `useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })` を追加。詳細モーダルのステータス遷移は代替経路として維持。

#### FE-RC-044 [MEDIUM] セマンティクス欠落
- `examinations/components/ExamItemsTable.tsx:195-301` — div grid でテーブルを表現（`role` もなし）→ `<table>` か ARIA grid
- `hospitalization/components/HospitalizationBoard.tsx:75-92` — `Card` に `onClick`。内側 `:127-138` に `<button aria-label>` があるため冗長 → Card 側を削除（C19 の思想）
- `components/shared/Pagination/Pagination.tsx:108-120` — 現在ページに `aria-current` なし、ルートが `<nav>` でない
- `components/shared/SortableHeader/SortableHeader.tsx:30-38` — `aria-sort` なし（`direction` 型は aria-sort 値と一致済み）

### コード品質

#### FE-RC-045 [MEDIUM] 400 行超ファイル（非テスト・documented exception 除く）
| ファイル | 行数 | 分割案 |
|---|---|---|
| `master/routes/lab-device-item-master-settings-model.ts` | 531 | 表示ラベル／draft 変換／バリデーションで 3 分割 |
| `examinations/hooks/use-examination-form-helpers.ts` | 519 | FE-RC-032 の effect 整理と同時に |
| `medical-records/components/MedicalRecordFormPanels.tsx` | 503 | 35 props の `MedicalRecordTabsAreaProps` を責務別に束ねる |
| `examinations/routes/ExaminationForm.tsx` | 502 | 履歴フィルタ（:129-230）・印刷（:118-125,401-415）・削除確認を別 hook/コンポーネントへ |
| `hospitalization/routes/hospitalization-form-panels.tsx` | 494 | StatusView/HeaderExtra/Fields/Body を個別ファイルへ |
| `trimming/routes/trimming-form-panels.tsx` | 491 | 40 props の `TrimmingFormBodyProps` を束ねる |
| `identity-links/routes/IdentityLinksPage.tsx` | 487 | FE-RC-047 |
| `estimates/routes/EstimateForm.tsx` | 463 | 3 セクション（:70-276）を `components/EstimateFormSections.tsx` へ |
| `reception/hooks/use-reception-kanban.ts` | 448 | `:44-158` の純関数群を `lib/kanban-columns.ts` へ |
| `owners/components/PetSubOwnersSection.tsx` | 438 | — |
| `lab-device/routes/lab-device-board-panels.tsx` | 419 | — |
| `settings/components/LstepTagConfigSection.tsx` | 417 | FE-RC-050 で半減 |
| `medical-records/components/TreatmentsTab/TreatmentsTab.tsx` | 413 | `handleSelectFromMaster`（:237-312, 76 行）の dose 判定を抽出 |
| `src/config/paths.ts` | 411 | `withTab(base, tab)` ヘルパで tab エイリアス 8 件（約 40 行）を除去。分割はしない |
| `medical-records/components/TreatmentsTab/treatment-row-quantity-cell.tsx` | 410 | — |
| `owners/routes/OwnerForm.tsx` | 404 | — |
- テストファイルで 800 行超: `use-examination-form.test.ts`（2662）、`use-vaccination-form.test.ts`（1317）、`use-medical-record-form.auto-create.test.ts`（1219）、`use-hospitalization-form.test.ts`（982）、`use-pet-form-list-state.test.ts`（930）、`ReservationFormModal.test.tsx`（845）、`OwnerReport.test.tsx`（820）→ describe 単位で分割候補

#### FE-RC-046 [MEDIUM] 50 行超の関数
- `accounting/hooks/use-accounting-completion-action.ts:187-309`（useActionState コールバック約 120 行）、`reservations/hooks/use-reservation-actions.ts:176-288`（`handleSave` 113 行、3 経路）、`components/shared/ReservationFormModal/ReservationFormModal.tsx:190-273`（`saveReservation` 85 行、`:208-219` と `:240-259` が同内容のバリデーション重複）、`examinations/hooks/use-examination-form-helpers.ts:356-437`、`checkups/hooks/use-checkup-form.ts:79-139`、`accounting/routes/AccountingDetail.tsx:44-338`、`medical-records/components/TreatmentsTab/TreatmentsTab.tsx:237-312`
- 改善案: 「payload 構築」「送信」「後処理」を純関数に分離。`validateReservationCore()` を `reservation-form-validation.ts` へ抽出（FE-RC-003 の JST 化と同時）。

#### FE-RC-047 [MEDIUM] `IdentityLinksPage` に検索・選択・リンク・履歴の全責務が集中
- 対象: `src/features/identity-links/routes/IdentityLinksPage.tsx:50-487`（1 関数 437 行、`useState` 12 個、owner/pet で同型ハンドラが複製 `:186-224,226-271,273-319`）
- 改善案: `hooks/use-identity-link-search.ts`（FE-RC-011 の useQuery 化）、`hooks/use-identity-group-session.ts`、`components/IdentityLinkSection.tsx`（owner/pet を型パラメータで共通化）へ分割。

#### FE-RC-048 [MEDIUM] ハードコードされた業務値・プレースホルダ
- `examinations/routes/ExaminationForm.tsx:375` `staffName="医師A"`（固定文字列を患者カードに表示）、`:379-381` `"保険情報未登録"` / `nextVisitDate="-"`
- `accounting/hooks/use-accounting-detail-state.ts:55-58` `petSpecies: ... ?? "犬"`（種別の既定値が「犬」）
- 税率 10%/8%: `cash-register/components/cash-register-close-panels.tsx:85,100`、`ClosePrintArea.tsx:144,149`、`medical-records/components/TreatmentsTab/TreatmentsTab.tsx:123`（`* 0.1`）、`MedicalRecordBillCheck.tsx:219`（`taxRate: 0.1`）— #179 で accounting-reports 側は設定値化済み
- `lab-device/lib/lab-device-agent.ts:1` `http://127.0.0.1:17654` → `import.meta.env.VITE_LAB_DEVICE_AGENT_URL ?? default`
- `lstep/components/TagOwnerListDrawer.tsx:160-161` 「先頭100名」「5000名」を文字列直書き（定数 `HISTORY_FETCH_LIMIT`/`LSTEP_CSV_EXPORT_LIMIT` が存在）
- `line-reserve/src/pages/CustomerInfoPage.tsx:152` `total={8}` 固定（`ConfirmPage` は `getStepProgress()` 使用で不整合）
- 改善案: 未取得時は空表示、税率は `clinic.standard_tax_rate` から描画、定数参照へ。

#### FE-RC-049 [MEDIUM] マジックナンバー
- `accounting/components/ItemListCard.tsx:47` `999999999`、`:309` `maxLength={500}`、`examinations/hooks/use-examination-form-helpers.ts:451` `> 500`、`accounting/components/PaymentCard.tsx:180,188` `1000`/`10000`、`checkups/routes/CheckupForm.tsx:52` `limit: 100`、`accounting/components/UnpaidTab.tsx:40` `limit = 20`、`use-accounting-detail-state.ts:161,197` `"0.5"`、`AccountingDetailPanels.tsx:222` `h-[calc(100vh-140px)]`、`medical-records/components/VitalsTab/VitalsTab.tsx:78` `< 30 || > 45`（臨床閾値）、`reception/hooks/use-reception-kanban.ts:314,317` `3000`/`4000`、`reservations/components/WeekViewDayColumn.tsx:44,76` `15`/`10`/`300`、`trimming/hooks/trimming-form-utils.ts:126,130`、`shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:303` `"12:00"-"13:00"`、`trimming/routes/TrimmingList.tsx:77` `pageSize: 10`、`medical-records/components/MedicalRecordVaccination.tsx:66,174` `"4weeks"`（`vaccinations` 側の `DEFAULT_NEXT_SCHEDULE_TYPE` は `"1year"` で不一致）、`lab-device/api/lab-device.ts:210,240,249` `9600`/`2000`/`5000`、`lab-device-serial.ts:124` `Math.min(250, ms)`（同ファイル `LAB_DEVICE_IDLE_MS = 250` 未使用）、`components/shared/ReservationFormModal/ReservationFormModal.tsx:145,147` `setHours(10/11)`、`line-reserve/src/pages/ConfirmPage.tsx:123` `|| 4`
- 改善案: `UPPER_SNAKE_CASE` 定数化（`MAX_UNIT_PRICE`, `VITAL_TEMPERATURE_RANGE`, `TOAST_THROTTLE_MS`, `DEFAULT_BREAK`, `DEFAULT_RESERVATION_START_HOUR` 等）。

#### FE-RC-050 [MEDIUM] `LstepTagConfigSection` の 3 セクションが同型（2 入力 + 追加 + 一覧 + 削除）
- 対象: `src/features/settings/components/LstepTagConfigSection.tsx:102-202,208-300,306-400 付近`
- 改善案: 共通 `TagPairSection` へ（FE-RC-009 の form 化と同時）。

#### FE-RC-051 [MEDIUM] liff / line-reserve で raw Tailwind パレット色クラスが 134 箇所（20 ファイル）
- 対象（件数）: `line-reserve/src/pages/MyReservationsPage.tsx` 30、`CustomerInfoPage.tsx` 18、`ConfirmPage.tsx` 18、`liff/src/pages/PetHealthPage.tsx` 13、`line-reserve/src/components/Calendar.tsx` 8、`TrimmingOptionSelectPage.tsx` 7、`liff/src/pages/LiffLinkPage.tsx` 7、`ListItem.tsx` 6、`TopPage.tsx` 5、他 11 ファイル
- 規約: Design Tokens（`src/` 側は `design-system-audit.mjs` で 0 件だが liff/line-reserve は audit 対象外）
- 現状: `confirmed: 'bg-green-100 text-green-800', pending: 'bg-yellow-100 text-yellow-800'`（`MyReservationsPage.tsx:35-42`）
- 改善案: `brand-tokens.css` の `@theme` に `--color-noah-{success,warning,danger,info,border,muted}` を追加して置換し、audit の対象パスに liff/line-reserve を加える。

#### FE-RC-052 [MEDIUM] `bg-white` / `text-white` / `border-black` 直書き（`C.bgWhite` / `C.textOnBrand` が存在）
- 対象: `closing-settings/components/{StandardClosingTimeSection.tsx:68,170, SpecialPeriodSection.tsx:60,66, HolidaySection.tsx:50,56}`、`accounting/components/AccountingDocument.tsx:174,188,254`（`border-black`）、`clinic-settings/components/CompanyInvoiceSection.tsx:65`、`aggregation/components/{CPMStageSummary.tsx:32, AggregationFilterPanel.tsx:21}`、`accounting/components/DailyAccountingPrintArea.tsx:23`、`medical-records/components/{MedicalRecordFormPanels.tsx:78, MedicalRecordVaccination.tsx:32, NextVisitDateField.tsx, MedicalRecordPrintView.tsx, CheckupsTab/CheckupsTabBadges.tsx}`、`owners/components/LineSendPanel.tsx:159`、`line-reservation/components/LinkedLineCustomers.tsx:100,205`、`manual/components/{ManualContent.tsx:156, ManualEditor.tsx:136,145,154, ManualSidebar.tsx:110,148}`、`master/components/MedicineDoseParamsEditor.tsx:174`
- 改善案: トークンへ置換。帳票用 `border-black` は印刷トークンを追加。`design-system-audit.mjs` に `bg-white|text-white|border-black` 直書き検出を追加。

#### FE-RC-053 [MEDIUM] AuthProvider 配置についてコメント・規約が三者で矛盾
- 対象: `src/app/router.tsx:11-13`（「アプリ全体に配置。/login でも useAuth() が使用可能」＝実装どおり）、`src/app/provider.tsx:9-10`（「保護ルート側にのみ配置」＝不一致）、`frontend/CODING_RULES.md:197,378-379,488-490`（不一致）
- 改善案: provider.tsx と CODING_RULES を BUG-031 後の実態（全体配置 + password-recovery 経路のみ restore skip）に更新。

---

## 3. LOW

### 命名・import 順序
- **FE-RC-054** import 順序違反（React → external → `@/` → 相対 → 型）: `estimates/routes/{EstimateForm.tsx:1-8,31-35, EstimateDetail.tsx:1-7}`、`examinations/routes/ExaminationForm.tsx:38-47`、`accounting/routes/AccountingDetail.tsx:8-9,26-33`、`accounting/components/{ItemListCard.tsx:2-3, CreditCorrectionDialog.tsx:1-2}`、`hospitalization/components/**`（`// React/Framework` 見出し下に `@/lib/design-tokens` が先行: `AddForm.tsx`, `EditRow.tsx`, `DailyVitalsSection.tsx`, `DailyRecordsTab.tsx`, `CarePlanTab.tsx`, `HospitalizationBoard.tsx`, `HospitalizationTabbedView.tsx`, `HospitalizationDetailActions.tsx:15-17`）、`hospitalization/hooks/use-hospitalization-form.ts:13-15`、`lstep/components/CheckupSyncFilterForm.tsx:4-8`、`vaccinations/routes/VaccinationList.tsx:1-3,25,30`、`trimming/routes/TrimmingList.tsx:10-11,26-27`、`medical-records/components/{TreatmentsTab/TreatmentsTab.tsx:12-23, MedicalRecordVaccination.tsx:9-15}`、`shifts/components/ShiftCalendar/ShiftCalendar.tsx:1-4`、`reception/components/AppointmentCard.tsx:5-6,21-22` → ESLint `import/order` を導入して機械化。区画コメント（`// React/Framework` 等）は実態と乖離しているので削除。
- **FE-RC-055** API query hook 名の `useGet` 欠落: `src/hooks/use-master-items.ts:101` `useMasterItems`（純 wrapper → 改名対象）、`use-animal-species.ts:35`、`use-clinic-tax-rates.ts:26`、`use-current-clinic-name.ts:12`（派生値を返す facade → 規約に例外明記で可）。feature 側: `examinations/api/unconfirm-examination.ts:23`、`medical-records/api/medicine-dose-lookup.ts:27`、`master/api/exam-types-master.ts:358`、`master/api/lab-device-item-masters.ts:76`、`lab-device/api/lab-device.ts:258-310`（`usePut/useClear/useReceive/useAttach/useDetach`）、`owners/api/replace-pet-sub-owners.ts:28` — mutation は動詞が固有なので許容範囲。規約側で「mutation は業務動詞可」と明記。
- **FE-RC-056** ファイル名と内容の不一致: `src/hooks/use-vaccinations.ts:72`（`useCreateVaccination` のみ → `use-create-vaccination.ts`）。
- **FE-RC-057** `ActionState` 型のローカル再定義: `checkups/hooks/use-checkup-form.ts:33-36`、`estimates/hooks/use-estimate-form.ts:50-54`、`clinic-settings/routes/use-clinic-master-settings.ts:25-29` → `@/types/form` を使用。
- **FE-RC-058** estimates 配下でクォート混在（`'react'` と `"sonner"`）: `EstimateForm.tsx:8-14`、`use-estimate-form.ts:1-5`。
- **FE-RC-059** 規定外サブディレクトリ: `features/*/lib/`（auth, closing-settings, estimates, examinations, lab-device, manual, medical-records, owner-report, owners, reception）、`features/*/constants/`（estimates, lstep, master, medical-records, reservations）、`hospitalization/{constants.ts,styles.ts}`、`hospitalization/components/cage-keyboard-coordinates.ts`、`owner-report/owner-report.css`。`shifts/` に `hooks/` なし（`ShiftFormDialog.tsx:96-146` の formAction はコンポーネント内）、`settings/`・`reception/`・`owner-report/`・`identity-links/`・`line-reservation/`・`lstep/` に `types/` なし → 実態は 10 feature 以上で `lib/` を採用しているので **features/CLAUDE.md 側で `lib/`・`constants/` を正式許容**するのが現実的。

### ドキュメント乖離（規約自体の修正）
- **FE-RC-060** feature 間 import の可否が `CODING_RULES.md`（禁止）と `frontend/CLAUDE.md`/ESLint（deep import のみ禁止）で不一致。FE-RC-015 の決着と同時にどちらかへ統一。
- **FE-RC-061** `frontend/CODING_RULES.md` の実コードと乖離した記述: `src/stores/`（存在しない、zustand も未導入）、`hooks/useTableSort.ts`・`use-mobile.ts`（存在しない）、`lib/zod.ts`（存在しない）、§1.4 の router 例（現行は `app/routes/*-routes.tsx` 分割）、§2.2/§9 の `formData.get("name") as string` 例示（FE-RC-036 と矛盾）、`features/*/types/index.ts` は「src/types への re-export のみ」（実態は型定義の実体で features/CLAUDE.md は `types/` を標準構成として認めている）。
- **FE-RC-062** `frontend/src/features/manual/CLAUDE.md` の乖離: 「テスト」節が `__tests__/parse-frontmatter.test.ts`（全廃済み。実際は `lib/parse-frontmatter.test.ts`）、`docker compose exec frontend pnpm test:run -- src/features/manual`（`frontend/CLAUDE.md` が「`--` は罠」と明記する禁止手順）、「将来の拡張候補」に「編集者ロール: 管理者が UI 上で MD 編集」（同ファイル上部で実装済みと記載）。
- **FE-RC-063** `src/app/provider.tsx:9-10` のコメント（FE-RC-053）、`src/hooks/use-pet-checkup-results.ts:9,78-80`（FE-RC-020）、`estimates/routes/EstimateForm.tsx:339-340`（"but for now redirect to list" の暫定コメント）。

### Hooks / コンポーネント（小）
- **FE-RC-064** `useCallback` を return 文中で呼ぶ: `checkups/hooks/use-checkup-form.ts:164-181`。
- **FE-RC-065** curried `useCallback`（安定化効果なし）: `line-reservation/routes/LineReservationPageEditor.tsx:70-76,139` → `name` 属性 + 単一 `onChange`。
- **FE-RC-066** `useEffect` deps を `JSON.stringify` で偽装: `owners/routes/OwnerForm.tsx:186-187` → `useMemo` で署名文字列を作り deps に渡す。
- **FE-RC-067** `key={idx}` で削除可能な行を描画: `accounting/components/PaymentCard.tsx:117,264` → 安定 id。
- **FE-RC-068** レンダー毎に再生成される `components` マップ・IIFE: `manual/components/ManualContent.tsx:107-204,210-248` → モジュール定数 + `ArticleAdjacentNav`。
- **FE-RC-069** `RowActionDropdown.tsx:24-27` の外側 `<div onClick={stopPropagation}>` は C19（行 onClick 禁止）以後は不要。
- **FE-RC-070** `reservations/components/WeekViewDayColumn.tsx:88-95` `div onClick`（座標依存・role 化不適）。時間スロットをボタン要素で刻む設計が望ましい（記録のみ）。
- **FE-RC-071** `src/shared-liff/use-fetch-state.ts:29-32` の `set-state-in-effect` 抑制は理由付きで許容範囲。`loading` を `data === null && error === null` に派生化すれば消せる。

### エラー処理（小）
- **FE-RC-072** api 層に `toast.success` が置かれ UI 通知が api に結合（owners 参照実装のパターン B と矛盾）: `owners/api/{update-owner-line.ts:46,67, send-line-message.ts:44, create-owner-tag.ts:37, delete-owner-tag.ts:31, confirm-owner-line-id.ts:32, update-owner-delivery-exclusion.ts:42, update-owner-delivery-caution.ts:42, update-owner-transfer-status.ts:40, generate-line-link-token.ts:36}`、`LineSendPanel.tsx:219-223`（inline「送信しました」と toast の二重通知）→ FE-RC-005 の方針決定に従う。
- **FE-RC-073** バリデーションエラーを `toast.error` で通知: `settings/components/LstepTagConfigSection.tsx:114,220,318`、`TriggerPrioritySection.tsx:60`、`shifts/components/ShiftTemplateSettingsParts.tsx:147` → `fieldErrors` + `FormFieldError`。
- **FE-RC-074** コメント無しの `catch {}`: `examinations/hooks/use-examination-form-helpers.ts:457-459`、`master/components/MedicineDoseParamsEditor.tsx:153-155` → 「onError → handleApiError 済み」を明記。
- **FE-RC-075** `src/lib/handle-api-error.ts:182-184` 非 Axios `Error.message` をそのまま toast（zod/TypeError の英文が表示され得る）→ 汎用文言にし `message` は DEV `console.error` 限定。
- **FE-RC-076** `src/lib/axios.ts:128-138` と `:161-170` の 401 リダイレクトが重複・`"/login"` リテラル比較 → `redirectToLoginWithFrom()` に抽出し `paths.auth.login.path` を使用。
- **FE-RC-077** `console.error` が本番でも BE 応答本文を顧客端末へ出す: `liff/src/api/liff-api.ts:66,75,98`、`line-reserve/src/api/liff-api.ts:56` → `shared-liff/dev-log.ts`（`import.meta.env.DEV` ガード）へ。フル logger は不要（計 11 件）。
- **FE-RC-078** `lab-device/lib/lab-device-serial.ts:100-102` `localStorage.setItem` が未 try（`getItem` 側は try 済み）。

### セキュリティ（小）
- **FE-RC-079** `line-reserve/src/api/liff-api.ts:64-153`（全 11 エンドポイント）で `clinicId` を未エンコードで URL パスに埋め込み（liff 側は `encodeURIComponent` 済み）→ `clinicPath()` ヘルパで統一。
- **FE-RC-080** `src/config/paths.ts` の `getHref(id)` が id をエンコードしない（`owner-report-window.ts:9` の petId だけ個別エンコードで二重基準）→ `getHref` 内で `encodeURIComponent(String(id))`。

### Query keys / staleTime
- **FE-RC-081** `detail` キーが `all()` の prefix 外で `invalidateQueries(all())` が詳細に届かない: `src/lib/query-keys.ts:139-141`（reservations）、`:152-154`（examinations）、`:170-172`（vaccinations）、`:178-180`（hospitalizations）、`:195-197`（trimmings）、`:244-247`（inventoryItems — 唯一注記あり）→ 残 5 件に同注記＋mutation 側の `detail(id)` 個別 invalidate をテストで担保。
- **FE-RC-082** `staleTime` 生数値: `auth/api/get-me.ts:27` `10 * 1000`、`lstep/api/get-checkup-sync-preview.ts:87` `0` → `QUERY_STALE_TIMES.SESSION` / `.NONE` を追加。

### その他
- **FE-RC-083** 死コード: `examinations/routes/ExaminationForm.tsx:306-307`（次行に包含される分岐）、`ExamPivotTable.tsx:171-175`（`return null` 二重）、`line-reservation/routes/LineReservationPageEditor.tsx:148` `{formState !== undefined ? null : null}`、`medical-records/components/TreatmentTable.tsx:268-286` `Cell` の未使用 `onClick` prop、`owners/routes/OwnersList.tsx:184` `_isPetSaving`（pending が UI 未反映）、`shifts/components/ShiftCalendar/ShiftCalendar.tsx:20-21` `@deprecated StaffItem`。
- **FE-RC-084** `TODO(Phase 2)` 残置: `examinations/components/ExamPivotTable.tsx:54,246` → STATUS.md へ移して参照 ID を残す。
- **FE-RC-085** ネスト三項: `medical-records/routes/medical-record-form-ready-panels.tsx:50-56`、`shifts/components/ShiftCalendar/ShiftCalendar.tsx:230-236`、`reception/components/AppointmentCard.tsx:107-113`、`hospitalization/routes/hospitalization-form-panels.tsx:329-330`、`lab-device/routes/lab-device-board-panels.tsx:261-265`、`lstep/components/LstepCsvImportSection.tsx:171-176`、`hospitalization/components/HospitalizationBoard.tsx:80-85` → 早期 return 関数かオブジェクトマップ。
- **FE-RC-086** inline style による色指定: `settings/components/LstepTagConfigSection.tsx:85,137`（`style={{ borderColor: PALETTE.borderLight }}`）→ クラストークン。
- **FE-RC-087** `VitalsTab.tsx:72-73` と `VitalsTabRows.tsx:126-127` で `recorded_at`（date-time）比較の経路が不揃い（`jstDateTimeLocalToISOString` 経由 vs 生 `new Date()`）。
- **FE-RC-088** `useTransition` を UI state hook が外部へ返し別 hook が実行する構造: `medical-records/hooks/use-medical-record-form-helpers.ts:58` → `useMedicalRecordAutoCreate` 側で持つ（違反ではない）。
- **FE-RC-089** master `use-master-save.ts:130-133` / `use-master-crud.ts:189` の `useTransition` は CODING_RULES §2.2 が許容例として明記。`MasterSidePanel` が `<form action>` を持つため将来 `useActionState` へ寄せられる（記録のみ）。

---

## 4. 機械ガードの追加提案（再発防止）

| 対象 | 現状 | 追加案 |
|---|---|---|
| `&&` 条件レンダリング | `react/jsx-no-leaked-render`（boolean 左辺を許す） | `validStrategies: ["ternary"]` |
| feature 間 import | deep import のみ禁止 | `src/features/<a>/**` → `@/features/<b>` を `no-restricted-imports` で禁止（FE-RC-060 の決着後） |
| import 順序 | 手動・区画コメント | `eslint-plugin-import` `import/order` |
| 動的 Tailwind クラス | なし | `design-system-audit.mjs` に `-\[\$\{` / `:\$\{` 検出 |
| `bg-white` / `text-white` / `border-black` | なし | 同 audit に追加 |
| liff / line-reserve の raw パレット色 | audit 対象外 | audit 対象パスに追加 |
| `.tsx` コンポーネントファイル名 | `check-feature-filename-convention.mjs` は `.ts` のみ | `.tsx` で JSX export かつ小文字始まりを検出 |
| hooks の配置 | なし | `src/features/*/{routes,components}/use-*.ts(x)` を CI で fail |
| `formData.get(...) as string` | なし | `no-restricted-syntax` で `TSAsExpression > CallExpression[callee.property.name="get"][callee.object.name=/formData/i]` を警告 |
| ratchet baseline | `.eslint-disable-baseline` 22、`.filename-baseline` 23、`.coverage-baseline` 43.78 | FE-RC-032/016/017 の解消後に締める |

---

## 5. 合格項目（再確認不要）

- 型安全: `any` 0、`@ts-ignore`/`@ts-expect-error` 0、`FC`/`forwardRef` 0
- React 19: `<form onSubmit>` 0、`useState(false)+setIsLoading` 送信管理 0、`useTransition` によるフォーム送信 0（検出 2 件は削除・自動作成の非フォーム用途で規約準拠）
- import 境界: deep import 0（ESLint 許可の `auth/provider` のみ）、層逆転 0、liff/line-reserve → features 0
- query key: 配列リテラル 0（ESLint 機械禁止）、`src/hooks/*` の staleTime は全 21 箇所 `QUERY_STALE_TIMES` 使用
- スタイル: `src/` 内 raw hex 0・raw パレット色 0、`<tr onClick>` 0、`tabIndex` 正値 0、`dangerouslySetInnerHTML` 0
- 構造: 全 29 feature に `index.ts`、空 `index.ts` 0、`__tests__/` 0、`utils/` 0、`.gitkeep` 0、`export default`/`export *` 0
- セキュリティ: token の `localStorage` 保存なし（クリニック ID のみ）、`window.location.href` は `parseInternalPath` 検証済み、liff/line-reserve の API 応答は zod 全検証、e2e 認証情報は env 必須化、`eval`/`innerHTML` 0
- 臨床安全（良例）: `OwnerPetsSection.tsx`、`use-pet-form-list-state.ts:104,167`、`use-reception-kanban.ts:171-180,274-422`、`VaccinationList.tsx:117-123`、`use-medical-record-auto-create.ts:129-227`、`HospitalizationForm.tsx:34-44`、`use-master-crud.ts:183-188,253-254` は二重防壁＋`useLayoutEffect` ref 再検査を実装済み
- `PrintPortal` は `data-print-portal` 固定キー方式で components/shared/CLAUDE.md に準拠
- 却下済み提案（`frontend/CLAUDE.md`）は再提案していない

---

## 6. 検証コマンド（手動実行）

本ファイルはドキュメントのみの追加で、コード変更はない。所見の再現は以下（すべて `frontend/` で実行）:

```bash
# 動的 Tailwind クラス
rg -n '\-\[\$\{|:\$\{C\.' src --glob '*.tsx'
# && 条件レンダリング
rg -n --pcre2 '^\s*\{[^{}\n]*\S\s+&&\s+(\(|<)' src --glob '*.tsx' --glob '!*.test.*'
# feature 間 import
rg -n -o "@/features/[a-z-]+" src/features --glob '*.ts' --glob '*.tsx' | awk -F'[:/]' '{split($0,a,":"); split(a[1],p,"/"); match($0,/@\/features\/[a-z-]+/); t=substr($0,RSTART+11,RLENGTH-11); if (p[3]!=t) print a[1]":"a[2]" -> "t}'
# hooks の routes/components 配置
rg --files src/features -g 'use-*.ts' -g 'use-*.tsx' -g '!*.test.*' | rg -v '/hooks/'
# formData.get as string
rg -n 'formData\.get\([^)]+\) as ' src --glob '!*.test.*'
# liff/line-reserve raw パレット色
rg -c '\b(text|bg|border|ring)-(red|blue|green|gray|yellow|amber|orange|emerald|teal|sky|indigo|purple|pink|rose)-[0-9]{2,3}\b' liff/src line-reserve/src
# 全体 lint（禁止コマンド — 手動で）
docker compose exec frontend pnpm lint
```

## 7. 実施記録

キャンペーンブランチ: `refactor/fe-rc-2026-09`（base `57e8d1da1`）。Phase 0 + L1〜L8 merge + Phase 2（055 / feature-import ban / baselines）。push なし。claim: `claim/FE-RC-CAMPAIGN-2026-09`（ユーザー解放待ち）。

| ID | Status | Changed files / notes | Verification |
|----|--------|----------------------|--------------|
| FE-RC-001 | DONE | accounting/estimates/clinic-settings permissionsRef | vitest accounting/estimates + rg isMutationAllowed |
| FE-RC-002 | DONE | examinations/checkups/hospitalization/estimates render-side death block | vitest deceased tests |
| FE-RC-003 | DONE | vaccination-form-model + ReservationFormModal todayJSTISO | vitest vaccination + ReservationFormModal |
| FE-RC-004 | DONE | paired with 002 reason UI / ActionState.error | adjacent tests |
| FE-RC-005 | DONE | callers dropped duplicate toast across lanes | representative vitest |
| FE-RC-006 | DONE | ClinicHolidayModal useActionState | L6 commit |
| FE-RC-007 | DONE | UnlinkedLineIdForm form action | L5 merge |
| FE-RC-008 | DONE | MedicalRecordVaccination useActionState | vitest vaccination-form |
| FE-RC-009 | DONE | TriggerPriority/LstepTag forms | L6 |
| FE-RC-010 | DONE | RefundSection/EstimateDetail Action | L1 |
| FE-RC-011 | DONE | IdentityLinks useQuery + hooks | L7 vitest hooks |
| FE-RC-012 | DONE（初回 DONE 表記は誤り、追い込みで onError 5→6） | lab-device.ts useReceiveLabDeviceFrames onError | rg -c onError=6; vitest lab-device |
| FE-RC-013 | DONE（初回残 12 箇所を追い込みで解消） | design-tokens hover/focus 静的トークン + 12 call sites | rg 動的クラス=0 |
| FE-RC-014 | DONE | shared-liff brand = PALETTE.brand | brand-tokens.test |
| FE-RC-015 | DONE（追い込みで相対 re-export 層逆転 → 第2追い込みで実体移動・順方向 re-export） | hooks に list/history 実体; feature api は `@/hooks` re-export のみ; ESLint 相対 `../features/` 禁止 | relative features from hooks/lib/components=0; vitest medical-records+owner-report 518; knip PASS |
| FE-RC-016 | DONE（reception + treatments hooks 移動） | use-reception-column-view; use-treatments-tab | hooks_out=0 |
| FE-RC-017 | DONE（PascalCase リネーム; followup2 で *.test/use-* 除外し tsx baseline 14→0） | 42 git mv; isTsxViolation 除外 | kebab コンポ=0; check-filenames OK; baseline 2行目=0 |
| FE-RC-018 | DONE | pet-checkup-results → owner-report | Phase 0 |
| FE-RC-019 | DONE | AuthProvider split | Phase 0 |
| FE-RC-020 | DONE | types/checkup.ts | Phase 0 |
| FE-RC-021 | DONE | IdentityLinks ternary + validStrategies | Phase 0 eslint |
| FE-RC-022 | DONE | CarePlan/DailyVitals/cash-register forms | L3/L1 |
| FE-RC-023 | DONE | SubmitButton destructive | L8 + audit C5 |
| FE-RC-024 | DONE | line-reserve useTransition cancel | L8 |
| FE-RC-025 | DONE | memo handlers 安定化（L-C） | L-C commit eac4d6339 |
| FE-RC-026 | DONE | mutate deps 分解（hospitalization/master） | L3/L7 |
| FE-RC-027 | DONE | ExamStatusBadge / care-plan-item-model / phone / JST | L2/L3/L4/Phase0 |
| FE-RC-028 | DONE | useUrlPageSync consumers | rg useUrlPageSync |
| FE-RC-029 | DONE | use-master-side-panel-form | L7 |
| FE-RC-030 | DONE | useLayoutEffect ref sync | L1/L7 |
| FE-RC-031 | DONE | NavigationBlocker blockerRef | L-A |
| FE-RC-032 | DONE | partial (shifts/settings/ReservationFormModal) | L6/L8 |
| FE-RC-033 | DONE | examination pet sync 一元化 | L2 |
| FE-RC-034 | DONE | frontend handle-api-error DEV warn 等; BE code 拡張は out of scope | Remaining risks |
| FE-RC-035 | DONE | lab-device-serial onError/reason | L3 |
| FE-RC-036 | DONE（残 28 を getFormString 化） | 対象 8 ファイル | formData.get as = form-data.ts コメント1のみ |
| FE-RC-037 | DONE | ShiftType guards 等 | L6 |
| FE-RC-038 | DONE | ! 削減（cash-register/accounting/hospitalization） | L1/L3 |
| FE-RC-039 | DONE | estimate handleChange generics | L1 |
| FE-RC-040 | DONE（残 outline-none に focus 代替） | shared/feature call sites + STYLE | outline 代替なし=0 |
| FE-RC-041 | DONE | Refund/Payment/EditRow a11y | L1/L3 |
| FE-RC-042 | DONE | aria-label icons | L6/L7 |
| FE-RC-043 | DONE | Reception KeyboardSensor | L5 |
| FE-RC-044 | DONE | Pagination/SortableHeader/Board/ExamItems | L8/L2/L3 |
| FE-RC-045 | DONE（400/800 分割） | 13+7 分割 | ge400=0; gt800test=0 |
| FE-RC-046 | DONE | saveReservation/accounting/reservation-actions splits | L1/L5/L8 |
| FE-RC-047 | DONE | IdentityLinks split | L7 |
| FE-RC-048 | DONE | tax rates / placeholders | L1/L4 |
| FE-RC-049 | DONE | magic numbers constants (partial) | lanes |
| FE-RC-050 | DONE | TagPairSection | L6 |
| FE-RC-051 | DONE（残 raw パレット 0） | liff/line-reserve noah-* | palette rg=0; design-audit PASS |
| FE-RC-052 | DONE | bg-white/border tokens (partial + printBorder) | L1/L8 |
| FE-RC-053 | DONE | provider/CODING_RULES AuthProvider docs | L8 |
| FE-RC-054 | DONE | import order manual sort (plugin deferred) | lanes |
| FE-RC-055 | DONE | useGetMasterItems | Phase2 ceea7e2a2 |
| FE-RC-056 | DONE | use-create-vaccination.ts rename | L-A |
| FE-RC-057 | DONE | ActionState shared type | L1/L2 |
| FE-RC-058 | DONE | estimates quotes | L1 |
| FE-RC-059 | DONE | features/CLAUDE lib/constants 許容 + hospitalization move | L3/L8 docs |
| FE-RC-060 | DONE | feature-import ban docs+eslint | Phase2 |
| FE-RC-061 | DONE | CODING_RULES stale paths cleaned | L8 |
| FE-RC-062 | DONE | manual/CLAUDE.md | L7 |
| FE-RC-063 | DONE | provider/stale comments | L8 |
| FE-RC-064 | DONE（既存確認） | use-checkup-form useCallback top-level | L-B 確認 |
| FE-RC-065 | DONE | LineReservationPageEditor single onChange | L-C |
| FE-RC-066 | DONE | OwnerForm JSON.stringify deps | L5 |
| FE-RC-067 | DONE | PaymentCard keys | L1 |
| FE-RC-068 | DONE | ManualContent MARKDOWN_COMPONENTS 定数化 | L-C |
| FE-RC-069 | DONE | RowActionDropdown stopPropagation 削除 | L-A |
| FE-RC-070 | N/A | 記録のみ（座標依存 onClick） | fe-refactor |
| FE-RC-071 | DONE | use-fetch-state loading 派生 | L-A ≤20行 |
| FE-RC-072 | DONE | LineSendPanel duplicate | L5 |
| FE-RC-073 | DONE | validation toast→fieldErrors settings | L6 |
| FE-RC-074 | DONE | catch comments examinations | L2 |
| FE-RC-075 | DONE | handle-api-error 非Axios 汎用文言 | L-A + test |
| FE-RC-076 | DONE | axios redirectToLoginWithFrom | L8 |
| FE-RC-077 | DONE | shared-liff/dev-log.ts | L-A |
| FE-RC-078 | DONE | lab-device localStorage try | L3 |
| FE-RC-079 | DONE | clinicPath encode | L8 |
| FE-RC-080 | DONE | paths.getHref encode | L8 |
| FE-RC-081 | DONE | query-keys detail prefix 注記 | L-A |
| FE-RC-082 | DONE（生 staleTime 解消） | get-me SESSION; checkup-sync NONE | staleTime:[0-9]=0 |
| FE-RC-083 | DONE | dead code removals | lanes |
| FE-RC-084 | N/A（コードに TODO 残存なし。後続タスク候補は本 followup で台帳のみ） | ExamPivotTable | rg TODO(Phase 2)=0 |
| FE-RC-085 | DONE | nested ternary cleanups | lanes |
| FE-RC-086 | DONE | LstepTag inline style | L6 |
| FE-RC-087 | DONE | VitalsTab JST path | L4 |
| FE-RC-088 | N/A | 違反ではない | fe-refactor |
| FE-RC-089 | N/A | 記録のみ | fe-refactor |

### 検証要約（統合後・followup 2026-09-03）

- 追い込みブランチ: `refactor/fe-rc-2026-09`（Phase0 `3a0cf2a23` → lanes merge → Phase2 `f2cc8fb56`）
- Success greps: dyn=0, cross-feature=0, hooks_out=0, formAs=1(定義コメント), outline残=0, ge400=0, gt800test=0, palette=0, stale生数値=0, kebabコンポ=0, lab-device onError=6
- `pnpm unused` / `design-audit` / `check-filenames` / eslint-disable baseline: PASS
- 全体ゲート 4 本: BLOCKED（policy・ユーザー明示許可なし）
- worktree: fe-rc2-LA/LB/LC 削除予定。`sec-codex-uhqpm2` 残置。claim 未解放。
### 検証要約（統合後・followup2 2026-09-03）

- 第2追い込み: FE-RC-015 相対 re-export（hooks→features）を解消。`use-medical-records.ts` に list/history 実体、`get-medical-records.ts` は `@/hooks/use-medical-records` 順方向 re-export のみ。
- ESLint: `layerInversionRelativeRestrictedPattern` を層逆転 2 ブロックへ追加。probe `../features/owners` → `no-restricted-imports`。
- Filename ratchet: `isTsxViolation` が `.test`/`.spec`/`use-` を除外。`.filename-baseline` 2 行目 14→0。隣接テスト 5 PASS。
- `origin/main` (`ae6bfeace`) を `--no-ff` merge 済み（`git merge-base --is-ancestor` exit 0）。
- スコープ検証: vitest medical-records+owner-report **76 files / 518 tests PASS**; design-audit PASS; check-filenames PASS; eslint-disable baseline PASS; `pnpm unused` PASS。
- 既知残: `LabDeviceUnlinkedBanner` の `@/features/lab-device` deep import（Phase 0 ESLint off 例外）。相対 features from hooks/lib/components = 0。
- transform 重複（hooks list vs feature `transforms.ts`）: drift リスクとして記録。一本化は後続。
- 全体ゲート 4 本: BLOCKED（policy・ユーザー明示許可なし）。push / claim 解放なし。

