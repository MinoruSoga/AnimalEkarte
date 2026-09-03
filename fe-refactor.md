# frontend コード規約チェック結果（第2期・2026-09-03 再監査）

`frontend/` 配下（`src/`・`liff/src/`・`line-reserve/src/`。テスト・`components/ui/`・`types/generated/` 除く）を規約正本に照合した。

- 規約正本: `frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`frontend/src/hooks/CLAUDE.md`、`frontend/src/components/shared/CLAUDE.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`.claude/refs/error-handling.md`、`frontend/CODING_RULES.md`（優先順位: 実コード → CLAUDE.md → CODING_RULES.md）
- 方法: (1) 規約項目ごとの機械検出、(2) 臨床安全 / React フォーム / 配置・命名の 3 領域を並列精読、(3) HIGH は当該ファイルを再読して確定
- 各所見の `path` は `frontend/` 起点。行番号は再監査時点
- **第1期（FE-RC-001〜089）は main 統合済み。本台帳は残件と再発のみ。** ID は衝突回避のため **FE-RC-101** から採番する
- `frontend/CLAUDE.md` の「却下済み提案」は再提案しない（manual chunk 分割、死亡行グレーアウト、一覧行アクションの生死ブロック）

---

## 0. サマリー

| 重要度 | 件数 | 概要 |
|---|---|---|
| CRITICAL | 0 | — |
| HIGH | 0 | FE-RC-W2 で 101–107 を FIXED |
| MEDIUM | 1 残 | FE-RC-111 の所有外 `*Mutation` deps（列挙パスは分解済み） |
| LOW | 2 残 | FE-RC-125 kebab fixture leave、FE-RC-127 LIFF 待ち時間定数 |

**第1期で解消済み（本監査で再確認）**: 会計/見積/健診/検査/入院/カルテ新規の死亡二重防壁と `permissionsRef`、二重 toast、疑似フォーム 5 件（FE-RC-006〜010）、IdentityLinks `useQuery`、brand `#038B94`、`&&` リーク、cross-feature import、tsx filename ratchet 0、`any`/`FC`/`forwardRef`/`export default`/`export *`/`utils/`/`queryKey: [` 直書き、800 行超（`design-tokens.ts` 例外のみ）。

**推奨着手順**: ①トリミング臨床防壁（101/102）→ ②予防接種 render（103）→ ③NextVisitDateField JST（105）→ ④DatePicker 動的クラス（106）→ ⑤CompanyInvoice / lab-device（104/107）→ ⑥MEDIUM。

---

## 1. HIGH

### 臨床安全境界（frontend/CLAUDE.md）

#### FE-RC-101 [HIGH] トリミング: 死亡ペットの二重防壁がない
- 対象:
  - `src/features/trimming/routes/TrimmingForm.tsx:40-41` — `canSubmit` は権限のみ。死亡 `status` を見ない
  - `src/features/trimming/hooks/use-trimming-form.ts:106-128` — `selectedPets[0]` の存在だけ見て `createMutation`。`pet.status === "死亡"` 拒否なし
- 規約: 臨床安全境界 1（参照: `OwnerPetsSection.tsx` / `HospitalizationForm.tsx:81-82`）
- 7 直接記録入力のうち、死亡二重防壁が未実装なのはトリミングのみ
- 改善案: render で `new-deceased` ゲート（Submit 非表示 + 理由）。`formAction` 冒頭で `selectedPetRef` + `useLayoutEffect` により callback 拒否し `toast.error` または `ActionState.error` を返す

#### FE-RC-102 [HIGH] トリミング: mutation 直前の権限再検査がない
- 対象: `src/features/trimming/hooks/use-trimming-form.ts:93-97`（create/update）。削除は `use-trimming-form-helpers.ts` の delete handler
- 現状: `permissionsRef` / `isMutationAllowed` なし。防壁は `TrimmingForm.tsx` の `canSubmit` のみ
- 規約: 臨床安全境界 2
- 改善案: `use-examination-form.ts:68-80` と同じ `permissionsRef` + `useLayoutEffect` + `isMutationAllowed()` を action 冒頭へ。`TrimmingForm` から `canCreate`/`canEdit`/`canDelete` を渡す

#### FE-RC-103 [HIGH] 予防接種: 死亡ペットの render 側防壁がなく callback は無音失敗
- 対象:
  - `src/features/vaccinations/routes/VaccinationForm.tsx:21` — `canSubmit = id ? canEdit : canCreate`
  - `src/features/vaccinations/routes/VaccinationFormPagePanels.tsx:178-196` — 死亡でも Submit / 削除ボタンを表示。`PatientInfoCard` に status は出す（`:215`）
  - `src/features/vaccinations/hooks/use-vaccination-form-helpers.ts:249-253` — callback は死亡を拒否するが `return { success: false, timestamp }` のみ
- 規約: 臨床安全境界 1 + error-handling「UI を loading/success/error として表現する」
- 改善案: `CheckupForm.tsx:47-48` と同様に `isPetDeceased` で `canSubmit`/削除を非表示。callback は `toast.error` または `ActionState.error` を返す

#### FE-RC-104 [HIGH] clinic-settings: CompanyInvoiceSection に権限再検査がない
- 対象: `src/features/clinic-settings/components/CompanyInvoiceSection.tsx:23-27`
- 現状: `canEdit` は `fieldset disabled`（`:37`）と Submit 表示（`:50`）のみ。`formAction` 内で未検査
- 規約: 臨床安全境界 2
- 改善案: `use-clinic-master-settings.ts:48-55,70` と同じ `canEditRef` + `useLayoutEffect` を action 冒頭へ

#### FE-RC-105 [HIGH] カルテ次回来院日が `setHours` 比較（端末 TZ 依存）
- 対象: `src/features/medical-records/components/NextVisitDateField.tsx:32-47`
- 現状:
  ```
  const now = toJSTWallDate(new Date());
  now.setHours(0, 0, 0, 0);
  const selected = new Date(value);
  selected.setHours(0, 0, 0, 0);
  ```
- 規約: 臨床安全境界 3 — `todayJSTISO()` / `isPastJSTDate` の文字列比較。vaccinations / ReservationFormModal は第1期で修正済み
- 改善案: `if (value !== "" && isPastJSTDate(value))`。2年上限も `todayJSTISO()` 基準の文字列演算へ

### デザイントークン / エラー処理

#### FE-RC-106 [HIGH] DatePicker が `hover:${C.bgBrand}` を実行時合成（FE-RC-013 再発）
- 対象: `src/components/shared/DatePicker/DatePickerModel.ts:81-93`
- 規約: Design Tokens。Tailwind v4 は静的走査のため `hover:${C.bgBrand}` は CSS に含まれない
- 現状: `design-tokens.ts:415` に「FE-RC-013: static hover/focus variants」コメントがあるが、DatePicker は未置換
- 改善案: `C.hoverBgBrand` / `C.focusBgBrand` 等の**完成形静的トークン**へ置換。`design-system-audit.mjs` に `` hover:${ `` / `` text-[${ `` 検出を追加して再発を止める

#### FE-RC-107 [HIGH] LabDeviceUnlinkedBanner が `void mutateAsync().then()` で失敗時 UI を戻さない
- 対象: `src/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner.tsx:57-58,85-96`
- 規約: error-handling「promise rejection を放置しない」
- 現状: hook 側 `onError` で toast は出るが、失敗時に `justAttached` / `attachError` が更新されない。`.then` は成功時だけ走る
- 改善案: `mutate(vars, { onSuccess, onError })`。失敗時は `setJustAttached(null)` / `setAttachError` を明示

---

## 2. MEDIUM

### 臨床安全（残り経路）

#### FE-RC-108 [MEDIUM] 予防接種削除ボタンが死亡でも表示される
- 対象: `src/features/vaccinations/routes/VaccinationFormPagePanels.tsx:178-179`
- 規約: 臨床安全境界 1。callback 拒否はあるが render が欠ける
- 改善案: FE-RC-103 と同時に `selectedPet.status === "死亡"` なら削除非表示

#### FE-RC-109 [MEDIUM] 会計キャンセル / 返金に権限再検査がない
- 対象: `src/features/accounting/hooks/use-accounting-settlement-actions.ts:35-40,56-60`
- 現状: UI は `canCancelAccounting` / `canEdit`（`AccountingDetail.tsx`）。action 内は未検査
- 改善案: `canCancelRef` / `canEditRef` + `useLayoutEffect` を mutation 直前に置く

#### FE-RC-110 [MEDIUM] CreditCorrectionDialog に権限再検査がない
- 対象: `src/features/accounting/components/CreditCorrectionDialog.tsx:51-62`
- 現状: 親の `canPostCloseEdit && canSubmit` のみ
- 改善案: `canPostCloseEditRef` を props で渡し `formAction` 冒頭で再検査

### Hooks / 再レンダー

#### FE-RC-111 [MEDIUM] `useMutation` オブジェクト全体を deps に入れている（FE-RC-026 残）
- 規約: CODING_RULES §12.1 `rerender-dependencies`
- 代表:
  - `src/features/closing-settings/components/HolidaySection.tsx:43` — `[deleteMutation]`
  - `src/features/reservations/hooks/use-reservation-save-actions.ts:126,157`
  - `src/features/medical-records/components/VitalsTab/VitalsTab.tsx:119,141,153`
  - `src/features/medical-records/components/MedicalRecordBillCheck.tsx:221`
  - `src/features/hospitalization/routes/HospitalizationForm.tsx:53`
  - `src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts:105,131,153,177`
  - `src/features/shifts/routes/ShiftTemplateSettings.tsx:117,131`
- 改善案: `const { mutateAsync: deleteHoliday } = useDeleteHoliday()` と分解し、deps は安定参照のみ。機械検出用 ratchet（`[.*, *Mutation]`）を検討

#### FE-RC-112 [MEDIUM] カルテ保存が参照する `activeTabRef` を `useEffect` で同期している
- 対象: `src/features/medical-records/hooks/use-medical-record-save-action.ts:98-102`（同型: `use-medical-record-post-save.ts:25-27`）
- 規約: 臨床安全境界 2 — commit 直後に発火し得る callback 用 ref は `useLayoutEffect`。`canEditRef` は既に `useLayoutEffect`
- 改善案: `activeTabRef` も `useLayoutEffect` へ

#### FE-RC-113 [MEDIUM] memo 子へ毎レンダー新規ハンドラを渡している
- 対象: `src/components/shared/ReservationFormModal/ReservationFormModal.tsx:320-337`（memo の `ReservationFormFields`）、`:270-272`（memo の `SelectedPetChip`）
- 規約: CODING_RULES §12.1 `rerender-memo`
- 改善案: `handleFormChange` / `handleClearError` / `handleRemovePet` を `useCallback` 化

### フォーム

#### FE-RC-114 [MEDIUM] VitalsTab の追加が `onClick` + `mutate`（`<form>` なし）
- 対象: `src/features/medical-records/components/VitalsTab/VitalsTab.tsx:64-118,232,246`
- 規約: React 19 Patterns — `useActionState` + `<form action>` + `SubmitButton`
- 改善案: 追加フォームを `<form action>` 化し、バリデーションは `fieldErrors` + `FormFieldError`

#### FE-RC-115 [MEDIUM] line-reserve `CustomerInfoPage` の最終 CTA が form 外 `onClick`
- 対象: `line-reserve/src/pages/CustomerInfoPage.tsx:337`（`PrimaryButton onClick={handleNext}`）
- 改善案: 最終ステップを `<form action>` または `useActionState` へ

### 配置・DRY

#### FE-RC-116 [MEDIUM] hook が `routes/` に残っている（FE-RC-016 残 3 件）
- 対象:
  - `src/features/hospitalization/routes/hospitalization-form-chrome.ts:18` — `useHospitalizationFormChrome`
  - `src/features/medical-records/routes/MedicalRecordsColumns.tsx:32` — `useMedicalRecordsColumns`
  - `src/features/trimming/routes/TrimmingFormPanels.tsx:31-32` — `useTrimmingFormChrome`（`react-refresh/only-export-components` disable 付き）
- 規約: features/CLAUDE.md — hook は `hooks/`
- 改善案: それぞれ `hooks/` へ移動し eslint-disable を削除

#### FE-RC-117 [MEDIUM] `src/hooks` に単一 feature 専用フック（FE-RC-018 残）
- 対象:
  - `src/hooks/use-create-vaccination.ts` — 消費は `medical-records` のみ。加えて `features/vaccinations/api/create-vaccination.ts` に同名 hook があり **二重実装**
  - `src/hooks/use-reservation-type-color-map.ts` — 消費は `reservations` の 3 ファイルのみ
- 規約: hooks/CLAUDE.md
- 改善案: 前者は `features/vaccinations/api` を正本にして medical-records から共有経路を 1 本化（または `src/hooks` に集約して feature 側を re-export）。後者は `features/reservations/hooks/` へ

#### FE-RC-118 [MEDIUM] URL page clamp が `useUrlPageSync` 未使用の一覧に残っている（FE-RC-028 残）
- 対象（inline `useEffect` + `exhaustive-deps` disable）:
  - `src/features/hospitalization/routes/HospitalizationList.tsx:92-102`
  - `src/features/accounting/routes/AccountingList.tsx:99-109`
  - `src/features/inventory/routes/InventoryList.tsx:81-91`
- 既に移行済み: trimming / vaccinations / examinations / checkups
- 改善案: `@/hooks/use-url-page-sync` に寄せ、disable を消す

#### FE-RC-119 [MEDIUM] LIFF 連携が `useEffect` 内で POST する
- 対象: `liff/src/hooks/use-liff-link.ts:23-88`
- 規約: CODING_RULES §5.3 / §12.4 — データ取得・mutation は `useQuery`/`useMutation`。StrictMode 二重実行リスク
- 改善案: `linkPromiseRef.current ??=` で冪等化するか `useMutation` へ。コメント「setState は同期目的のため許容」は取得ではなく副作用 POST には当たらない

---

## 3. LOW

#### FE-RC-120 [LOW] ファイル名と export 名が不一致
- `src/features/lstep/routes/LstepDeliveryMonitorLogsTable.tsx` → `DeliveryLogsTable`
- `src/features/lstep/components/LstepCsvImportSection.tsx` → `CsvImportSection`
- `src/features/owner-report/components/OwnerClinicalBasicPanel.tsx` → `BasicInformationPanel`
- `src/features/medical-records/routes/MedicalRecordsListPanels.tsx` → `MedicalRecordsPageView`
- `src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx` → `MedicalRecordFormReadyPage`

#### FE-RC-121 [LOW] `frontend/CLAUDE.md` の design-tokens 行数が実測と乖離 — FIXED
- 文書を実測 **897行**（`focusBgBrand` 追加後）に同期。分割しない判断は維持

#### FE-RC-122 [LOW] stale コメント「型は any」
- 対象: `src/lib/transforms/reservation.ts:23-25` — 実装は `unknown` + 型ガード。コメントが嘘
- 同型: `src/features/reception/api/transforms.ts` も確認して揃える

#### FE-RC-123 [LOW] `useGetClinicHolidays` の二重実装
- `src/hooks/use-clinic-holidays.ts` と `src/features/shifts/api/clinic-holidays.ts`
- 改善案: `@/hooks` を正本にし shifts は re-export

#### FE-RC-124 [LOW] shared ReservationFormModal 局所 hook の配置
- `src/components/shared/ReservationFormModal/use-patient-selection-table.ts` — 同一フォルダ 1 consumer。許容範囲だが `hooks/` サブディレクトリ化が理想

#### FE-RC-125 [LOW] kebab-case `.ts` fixture が filename ratchet に残る
- `src/features/accounting/components/OwnerAccountingHistory.test-fixtures.ts`（テスト隣接 fixture。tsx 規約対象外だが baseline 1）

#### FE-RC-126 [LOW] 権限拒否時のカルテ保存が無音
- `src/features/medical-records/hooks/use-medical-record-save-action.ts:108-114` — 確定済み/権限なし/死亡で `{ success: false }` のみ。render 側ゲートはあるので頻度は低い
- 改善案: FE-RC-103 と同じく `ActionState.error` を返す

#### FE-RC-127 [LOW] `use-liff-link` の `LINK_SUCCESS_DISPLAY_MS = 800` マジックナンバー
- 命名済み定数なので実害は小さい。他 LIFF 待ち時間と揃えるなら `liff-config` へ

---

## 4. 機械検出で 0 件（合格）

| 規約 | 結果 |
|---|---|
| `any` 型（コード） | 0。コメント・HTML `step="any"`・`expect.any` のみ |
| `FC` / `forwardRef` / arrow export コンポーネント | 0 |
| `&&` 条件レンダリングの 0/"" リーク | 0 |
| raw hex（design-tokens / brand-tokens / globals 以外） | 0。`#008B94` は audit のネガティブ fixture のみ |
| `queryKey: [` 直書き | 0 |
| `export default` / `export *` | 0（`vite-env.d.ts` 除く） |
| `utils/` / `__tests__/` / `.gitkeep` | 0 |
| `<form onSubmit>` / `useState+setIsLoading` 送信 | 0 |
| deep import（`auth/provider` は CODING_RULES の意図的例外） | 0 |
| 層逆転 / feature 間 import | 0。`crossFeatureImportBanAllowlist` は空 |
| `dangerouslySetInnerHTML` / `localStorage` token / `tabIndex` 正値 | 0 |
| 非テスト 800 行超 | `design-tokens.ts` のみ（documented exception） |
| `console.log` | 0 |
| brand `#038B94` / `#027078` | shared-liff 含む一致 |

---

## 5. 第1期台帳との差分（再提案しない）

次は本監査で **FIXED** を確認した。再オープンしない。再開条件は第1期と同じ（新しい実行時証拠）。

- FE-RC-001〜004 の会計/見積/健診/検査/入院/カルテ新規経路（権限 ref・死亡二重防壁・vaccinations 日付・ReservationFormModal 日付）
- FE-RC-005 二重 toast
- FE-RC-006〜011 疑似フォーム / IdentityLinks `useEffect` フェッチ
- FE-RC-012 mutation hook 側 `onError`（caller Banner は FE-RC-107 として残）
- FE-RC-014 brand 色
- FE-RC-015/017/019/020/047/055/060 配置・型・ESLint 整合
- FE-RC-021 `&&` リーク
- FE-RC-028 の trimming/vaccinations/examinations/checkups（Hospitalization/Accounting/Inventory は FE-RC-118）

---

## 6. 第2期 FIXED（FE-RC-W2-CAMPAIGN 2026-09-03）

キャンペーン claim: `claim/FE-RC-CAMPAIGN-2026-09-W2`（エージェントは削除しない）。

- FE-RC-101 / 102 — トリミング死亡二重防壁 + `permissionsRef` / `isMutationAllowed`。拒否は toast + `ActionState.error`。Submit/削除を死亡時非表示
- FE-RC-103 / 108 — 予防接種 render で Submit/削除非表示。callback 拒否は `toast.error`
- FE-RC-104 — CompanyInvoice `canEditRef` + `useLayoutEffect`。formAction 冒頭で fail-closed
- FE-RC-105 — NextVisitDateField は `isPastJSTDate` / `todayJSTISO`。`setHours` 0
- FE-RC-106 — DatePicker `selected` は `C.hoverBgBrand` / `C.focusBgBrand`。audit C20
- FE-RC-107 — LabDevice attach は `mutate(..., { onSuccess, onError })`。失敗時 `justAttached` / `attachError` を戻す
- FE-RC-109 / 110 — 会計返金・キャンセル・クレジット訂正は権限 ref で fail-closed
- FE-RC-111（列挙パス） — HolidaySection / reservation save-actions / Vitals / BillCheck / HospitalizationForm / quick-patch / ShiftTemplateSettings は `mutate`/`mutateAsync` 分解済み
- FE-RC-112 / 126 — `activeTabRef` は `useLayoutEffect`。権限/死亡拒否は `ActionState.error`
- FE-RC-113 — ReservationFormModal の memo 子へ `useCallback` ハンドラ
- FE-RC-114 — Vitals 追加は `<form action>` + `useActionState`
- FE-RC-115 — CustomerInfoPage 最終 CTA は `<form action>` + `useActionState`
- FE-RC-116 — hospitalization / medical-records columns / trimming chrome を `hooks/` へ
- FE-RC-117 — 二重実装解消。`useCreateVaccination` 実装は `src/hooks` 正本（ESLint: no cross-feature / hooks→features）。feature API は re-export。color-map は物理移動せず feature から re-export（`COLOR_MAP_REL_PATH` 維持）
- FE-RC-118 — Hospitalization / Accounting / Inventory 一覧は `useUrlPageSync`
- FE-RC-119 — LIFF POST は `linkPromiseRef.current ??=`
- FE-RC-120 — export 名をファイル名に合わせた（lstep 2 + owner-report 1 + medical-records 2）
- FE-RC-121 — `frontend/CLAUDE.md` の design-tokens 行数を実測 897 に同期
- FE-RC-122 — 「型は any」コメント削除
- FE-RC-123 — `useGetClinicHolidays` は `@/hooks` 正本、shifts は re-export
- FE-RC-124 — `use-patient-selection-table` を `ReservationFormModal/hooks/` へ
- 独立レビュー follow-up（2026-09-03）— `useActionState` stale closure:
  - 予防接種 save/delete は `formDataRef` / `editPetRef` 未着で fail-closed。edit loading は pet GET を待つ
  - トリミング save/delete は `formDataRef` / `petFromEditRef` 未着で fail-closed。status なし stub は作らない
  - カルテ save は `saveSnapshotRef`（所見 + `isFinalized`）。default タブは `success: false`
  - Vitals は `addFormRef` + 死亡の render/callback 二重防壁。会計キャンセル拒否は toast

## 7. 第2期残件（意図的 leave）

#### FE-RC-111 [MEDIUM] 所有外の `*Mutation` deps
- キャンペーン列挙パスは FIXED。未 sweep（所有外）:
  - `src/features/cash-register/routes/CashRegisterClosePage.tsx`
  - `src/features/master/components/ReservationTypeUnavailableTimesSection.tsx`
  - `src/features/trimming/hooks/use-trimming-form.ts`（`deleteMutation`。101/102 レーン対象外）
  - `src/features/closing-settings/components/SpecialPeriodSection.tsx`
  - `src/features/master/components/ReservationTypeGroupedTable.tsx`
  - `src/features/examinations/hooks/use-examination-form.ts`
  - `src/features/reception/hooks/use-reception-kanban.ts` / `use-reception-modal-handlers.ts`
  - `src/features/master/components/ReservationTypeAvailableSlotsSection.tsx` / `ReservationTypeAvailableSlotsCalendar.tsx` / `ReservationTypeOccupationsSection.tsx` / `MedicineDoseParamsEditor.tsx`
  - `src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx` / `MedicalRecordImage.tsx`
  - `src/features/line-reservation/components/LinkedLineCustomers.tsx`
- 再開条件: 新キャンペーンで所有パスを切る。本キャンペーンは越境しない

#### FE-RC-125 [LOW] kebab-case `.ts` fixture
- `src/features/accounting/components/OwnerAccountingHistory.test-fixtures.ts`
- 理由: test-adjacent fixture。tsx filename ratchet 対象外。破壊的 rename リスクのため leave mismatch

#### FE-RC-127 [LOW] `LINK_SUCCESS_DISPLAY_MS = 800`
- `liff/src/hooks/use-liff-link.ts` — 命名済み定数。`liff-config` 寄せは実害が小さいため未実施

その他メモ:
- `useMedicalRecordFormReadyState` は `MedicalRecordFormReadyPanels.tsx` 内の非 export ローカル hook（FE-RC-116 対象外）
- 第1期 FE-RC-001〜089 は再オープンしない
