# リファクタ台帳（FE コード規約準拠）

更新日: 2026-09-02（洗い出しのみ。実装は未着手）

| 項目 | 値 |
|------|-----|
| **範囲** | `frontend/src` · `frontend/liff/src` · `frontend/line-reserve/src` の **production TS/TSX** |
| **除外** | `*.test.*` / `*.spec.*`（件数の分母から外す。隣接テストの有無だけ別記）· `src/types/generated/**`（編集禁止）· `src/components/ui/**`（shadcn。トークン接続済みの最小差分以外は触らない）· `frontend/e2e/**` |
| **目的** | 挙動を変えずに、`frontend/CLAUDE.md` と ESLint 機械ガードへ寄せる |
| **ブランチ** | `main`。クレーム: `claim/TODO-REFACTOR-FE`（エージェントは削除しない） |
| **検証** | 変更ファイルの Docker 経由 scoped `npx vitest run <path>`。full `pnpm lint` / `pnpm test:run` / `pnpm type-check` / `pnpm build` はユーザー手動 |
| **正本** | 衝突時は ①実装コード → ② `frontend/CLAUDE.md` とネスト CLAUDE.md → ③ `frontend/CODING_RULES.md`（③は一部陳腐化。本台帳 FE-DOC） |

スキャン: 2026-09-02。関数行数は `export function` / `function` の開き `{` から対応 `}`（brace 対応）。`memo(function Name` はファイル末までを目視で補完。`useExaminationForm` はデフォルト引数のため機械検出から漏れたので、定義行 102〜ファイル末で計上。

BE フェーズは `ad63bdf28`（`refactor: extract remaining 90-line backend functions`）で規約ゲート完了。BE 全文はこのコミット時点の本ファイルを参照。これ以降の本ファイルは FE 台帳。

---

## 対象外（意図的にやらない）

| 除外 | 理由 |
|------|------|
| BE の再分割 | ゲート達成済み。`preflightCSVShape` / `updateExaminationInTx` / 切り出し済み tx helper は再分割しない |
| `lib/design-tokens.ts`（884行）の分割 | CLAUDE.md の documented exception（FE7-4）。トークン表を割ると可読性が落ちるだけ |
| `lib/query-keys.ts`（378行）の分割 | キーファクトリーの表。分割すると prefix 契約が読みにくくなる。配列リテラルはここだけ許可 |
| `src/config/paths.ts`（411行） | ルート表。分割しても契約が読みにくくなる |
| `src/components/ui/**` の書き換え | shadcn 生成物。トークン接続以外は触らない |
| `src/types/generated/**` の手編集 | codegen 成果。import 移行は TASK-444 の別トラック |
| `utils/` の再作成 | FE7-1 で `lib/` へ統合済み。ディレクトリ 0 件 |
| `__tests__/` の再作成 | FE5-23 で全廃済み |
| React.FC / forwardRef の一括狩り | 生産コード **0** |
| `{cond && <JSX>}` の一括狩り | `react/jsx-no-leaked-render` で **0**（コメント上も現状違反 0） |
| `dangerouslySetInnerHTML` / `innerHTML` / `document.write` | ESLint 禁止。生産 **0** |
| `export *` / src 内 `export default` | 生産 **0**（vite/playwright 設定と `vite-env.d.ts` のみ） |
| 層逆転（`components` / `hooks` / `lib` → `@/features`） | 生産 **0** |
| liff / line-reserve → `@/features` | 生産 **0** |
| className/style の `#RRGGBB` 直書き（トークンファイル以外） | 生産 **0** |
| shared-liff を `frontend/shared-liff/` へ昇格 | FE7-3 Option B。再検討条件は liff/line-reserve の具体的な機能拡張計画が出たときだけ |
| manual chunk の追加分割 | FE12 却下。表示遅延の申告が再開条件 |
| 死亡ペット行の全体グレーアウト | FE12 却下（曽我裁定）。badge のみ |
| `/owners` 行アクションをペット生死でブロック | FE12 却下。飼主操作でありペット死亡で止めると正当業務を壊す |
| ConfirmDialog コンポーネントの削除 | 破壊的操作の明示確認。臨床ロックの代替には使わない、がコンポーネント自体は残す |
| 個人ルールの 50 行関数まで機械分割 | プロジェクトのファイルゲートは **800**。関数の作業ゲートは本台帳では **150**（BE と同じ）。50 は個人スタイル |
| 200–399 行ファイルの機械分割 | 800 ゲートと 150 行関数が先。ファイルを薄くするためだけに割らない |
| generated/models の 262 ファイル一括移行 | TASK-444-S2。allowlist を増やさない。本台帳の第1スライスではない |
| GORM 相当の `as unknown as` をテストから一掃 | テストの部分モック。生産の `lib/transforms/owner.ts` だけ別記 |
| `pnpm design-audit` の再発明 | C1/C3/C5/C6b/C7–C19 は機械化済み。新規リーフは allowlist 同期が同一コミット条件 |
| full `pnpm lint` / `test:run` / `type-check` / `build` | 禁止コマンド。scoped vitest のみ |

---

## 機械ガードで既に合格（再作業しない）

ESLint / ディレクトリ実測。回帰防止は既存ルールに任せる。

| ID | ゲート | 実測 |
|----|--------|------|
| FE-ANY | `@typescript-eslint/no-explicit-any` | 生産の型 `any` / `as any` / `eslint-disable no-explicit-any` **0**。`expect.any`・コメント・`step="any"` は対象外 |
| FE-CF-AND | `{cond && JSX}` | **0** |
| FE-REACT-FC | `React.FC` / `forwardRef(` | **0** |
| FE-SEC-HTML | `dangerouslySetInnerHTML` / `innerHTML` 代入 | **0** |
| FE-SEC-TOKEN | 認証トークンの localStorage / sessionStorage | **0**（clinic id と検査機器ポート設定のみ。httpOnly cookie） |
| FE-ARCH-UTILS | `src/utils/` | ファイル **0**。`from "@/utils/"` **0** |
| FE-ARCH-TESTS-DIR | `__tests__/` | **0** |
| FE-ARCH-STAR | `export * from` | **0** |
| FE-ARCH-LAYER | hooks/components/lib → `@/features` | **0** |
| FE-ARCH-APP | liff/line-reserve → `@/features` | **0** |
| FE-STYLE-HEX | トークン外の `#RRGGBB` in className/style | **0** |
| FE-SIZE-800 | 800 行超ファイル | `design-tokens.ts` のみ（例外） |
| FE-PET-SEL | 7 selector の `includeDeceased` | `usePetSelectionPage` が内部で `true`。予約モーダルはデフォルト `false` |
| FE-DATE-ISO | `toISOString().slice(0, 10)` で date-only | 生産 **0**。JST は `todayJSTISO` / `isPastJSTDate` |

残作業は以下。**eslint が既に落とすもの（FE-ARCH-001 / FE-QUERY-001）を第0弾にする。**

---

## FE-ARCH — Feature Indexing / ドキュメントドリフト

### FE-ARCH-001（第0弾・eslint 違反）

`medical-records` の CheckupsTab が `checkups` を **deep import** している。`eslint.config.js` の `no-restricted-imports`（`@/features/<name>/...` 禁止、`auth/provider` 以外）に抵触する。

| ファイル | 先 |
|----------|-----|
| `features/medical-records/components/CheckupsTab/CheckupsTab.tsx` | `checkups/api/get-checkup-type-fields` · `replace-checkup-field-results` · `checkups/components/DynamicCheckupFields` |
| `CheckupsTabRows.tsx` | 同上（fields + DynamicCheckupFields） |
| `CheckupsTabTable.tsx` | 同上（type のみ） |
| `CheckupsTab.test.tsx` | `get-checkup-type-fields` |

`features/checkups/index.ts` は List / PetSelection / Form のルートだけを公開していて、カルテ埋め込みに必要な fields API と `DynamicCheckupFields` が barrel に無い。

**直し方:** `checkups/index.ts` に公開 API を足し、呼び出し側を `@/features/checkups` に切り替える。`medical-records` から相対で `../../checkups` を指さない（それも cross-feature）。コンポーネントを `components/shared` に上げない（健診ドメインのまま）。

`src/app/router.tsx` の `@/features/auth/provider` は eslint の明示例外。触らない。

### FE-ARCH-002

`app/pages/` の feature 合成（`OwnerFormPage` が owners + pets + line-reservation + accounting）は方針どおり。新しい cross-feature UI は feature 同士を import せず `app/pages/` に足す。

`loaders.ts` は `owners` のみ（`features/CLAUDE.md` の例外）。他 feature を loader 化しない。

### FE-ARCH-003（ドキュメント）

`frontend/CODING_RULES.md` が現行と食い違う。

| 記載 | 現行 |
|------|------|
| 「16 features」 | `index.ts` 付き **28** feature |
| ツリーに `utils/` | 廃止済み。正は `lib/` |
| `from "@/utils/format/date"` の例が複数 | 禁止パターン |

正本は `frontend/CLAUDE.md`。CODING_RULES を CLAUDE に合わせて直す（`utils/` を復活させない）。

### FE-ARCH-004（ドキュメント）

`frontend/src/hooks/CLAUDE.md` の Query Cache 例が `queryKey: ["pet", petId]` の配列リテラル。本番は `queryKeys` 必須。例を `queryKeys.pets.detail` 相当に直す。

### FE-ARCH-005（TASK-444・本スライスの後）

`generated-models-import-allowlist.json` **262** 件。`@/types/generated/models` は Go ドメインモデルで HTTP wire ではない（BUG-431/BUG-433）。新規 import 禁止。既存は domain 別 `generated/*-responses` か専用 DTO へ寄せる。

allowlist を増やさない。一括置換は別トラック。本台帳の関数分割より後。

### FE-ARCH-006

feature 数 28（すべて `index.ts` あり）: accounting, accounting-reports, aggregation, auth, cash-register, checkups, clinic-settings, closing-settings, estimates, examinations, hospitalization, identity-links, inventory, lab-device, line-reservation, lstep, manual, master, medical-records, owner-report, owners, pets, reception, reservations, settings, shifts, trimming, vaccinations.

`hooks/use-pet.ts` など共有フックが複数 feature からペットを読む構造は層として正しい（features → hooks）。pets feature の barrel は create/update/delete のみ。

---

## FE-QUERY — queryKey ファクトリー

### FE-QUERY-001（第0弾・eslint 違反）

`features/accounting/components/CreditCorrectionDialog.tsx:72`

```ts
queryClient.invalidateQueries({ queryKey: ["accounting", "unpaid"] });
```

`no-restricted-syntax` が queryKey 配列リテラルを禁止。`queryKeys.accounting.unpaidBillings` は `["accounting", "unpaid", groupBy, params]` なので、prefix 無効化用にファクトリーへ例えば `unpaidAll: () => ["accounting", "unpaid"] as const` を足して使う。キー文字列は変えない（キャッシュ契約）。

他の生産コードに `queryKey: [...]` 直書きはこれだけ（`query-keys.ts` 本体とコメント、preview の spread は既存キーへの nonce 付与）。

---

## FE-REACT — React 19 フォーム

CLAUDE.md: 送信 pending は `useActionState` の `isPending`。`useState` + 手動 `isLoading` は禁止。

大半の `use-*-form.ts` は `useActionState` 済み（examination / vaccination / trimming / hospitalization / checkup / owner / estimate / inventory）。カルテ保存は `use-medical-record-save-action.ts` に分離済み。

### FE-REACT-001

`components/shared/ReservationFormModal/ReservationFormModal.tsx` が `const [isSubmitting, setIsSubmitting] = useState(false)` と `setIsSubmitting(true)`（212, 255）。予約確定を `useActionState`（または親の mutation `isPending`）へ。

`use-cash-register-close-form.ts` の `useState` は日付・時間帯のフィルタで送信 pending ではない。対象外。

`CarePlanTab` の `isSubmitting={createItem.isPending}` は RQ mutation の pending。対象外。

`pendingDelete` / `pendingOwnerChange` は確認ダイアログ用の対象 ID。送信 isLoading ではない。対象外。

---

## FE-CONFIRM — native confirm

製品哲学: 安全対策を確認ダイアログにしない（ロック / Undo / 物理ブロック）。ただし破棄確認・削除確認の UI 自体は残る。

### FE-CONFIRM-001

`window.confirm` が 2 箇所。`ConfirmDialog` / `NavigationBlocker` へ寄せる。

| ファイル | 用途 |
|----------|------|
| `src/hooks/use-side-peek-dirty.ts:39` | 未保存破棄 |
| `src/features/manual/components/ManualEditor.tsx:98` | エディタ破棄 |

`ConfirmDialog` の呼び出し面（削除・医院切替・健診同期など）は第0弾で消さない。臨床 mutation の唯一の防壁にしている箇所があれば FE-CLINICAL で個別に見る。

---

## FE-STYLE — トークン外の Tailwind palette

hex 直書きは 0。残るのは **palette クラス**（`bg-gray-*` / `text-red-*` / `text-blue-*`）。印刷面が中心。`C.*` / `STYLE.*` へ。印刷でトークンが使えない場合は例外をファイル先頭で固定する。

### FE-STYLE-001 印刷面

| ファイル | 例 |
|----------|-----|
| `features/examinations/components/ExaminationPrintArea.tsx` | `text-red-600` · `border-gray-400` · `bg-gray-100` |
| `features/cash-register/components/ClosePrintArea.tsx` | `bg-gray-100` / `bg-gray-50` |
| `features/accounting/components/DailyAccountingPrintArea.tsx` | `bg-gray-*` |
| `features/accounting-reports/components/MonthlyReportPrintArea.tsx` | `border-gray-*` · `bg-gray-*` · `text-gray-500` |

### FE-STYLE-002 画面

| ファイル | 例 |
|----------|-----|
| `ReservationFormModal.tsx` | `border-red-200 bg-red-50 text-red-800`（エラー帯。`C.bgRed50` / `C.textRed700` 等） |
| `DailyAccountingTabParts.tsx` | `text-blue-700`（`C.textStatusBlue` または会計トークン） |

`shared-liff/brand-tokens.ts` の hex は LIFF 用の意図的トークン。staff の `design-tokens.ts` に合流させない（FE7-3）。

---

## FE-SIZE — 大きい関数 / ファイル

**ファイルゲート（プロジェクト）:** 800。例外は `design-tokens.ts` のみ。

**関数ゲート（本台帳）:** 150。BE と同じ。90 まで機械で下げない。500 行超ファイルは関数抽出の結果として薄くなることを期待し、ファイルだけ先に切らない。

### 150 行超の関数（分割対象）

行数は本体（開き `{` 〜 対応 `}`）。概数は `memo(function` と `useExaminationForm`。

| 行 | 名前 | ファイル |
|----|------|----------|
| ~625 | `useExaminationForm` | `features/examinations/hooks/use-examination-form.ts`（ファイル 726） |
| ~537 | `TreatmentRow` | `features/medical-records/components/TreatmentsTab/TreatmentRow.tsx`（589） |
| ~513 | `ExamTypeFieldsEditor` | `features/master/components/ExamTypeFieldsEditor.tsx`（587） |
| ~408 | `PatientSelectionTable` | `components/shared/ReservationFormModal/PatientSelectionTable.tsx`（566） |
| 395 | `useVaccinationForm` | `features/vaccinations/hooks/use-vaccination-form.ts` |
| 351 | `useMedicalRecordForm` | `features/medical-records/hooks/use-medical-record-form.ts` |
| 327 | `App` | `line-reserve/src/App.tsx`（ページは分割済み。残りは flow スイッチ） |
| 286 | `useTrimmingForm` | `features/trimming/hooks/use-trimming-form.ts` |
| 225 | `useHospitalizationForm` | `features/hospitalization/hooks/use-hospitalization-form.ts` |
| 183 | `ClinicMasterSettings` | `features/clinic-settings/routes/ClinicMasterSettings.tsx` |
| 181 | `useOwnerForm` | `features/owners/hooks/use-owner-form.ts` |
| 174 | `useCheckupForm` | `features/checkups/hooks/use-checkup-form.ts` |
| 160 | `CsvImportSection` | `features/lstep/components/LstepCsvImportSection.tsx` |
| 151 | `useInventoryForm` | `features/inventory/hooks/use-inventory-form.ts` |

90–149（参考。150 を割ったあと必要なら）: `CsvUploadSection` 136 · `CashRegisterClosePage` 107 · `ShiftCalendarPage` 106。

機械検出の 150 超は 10。`useExaminationForm` / `TreatmentRow` / `ExamTypeFieldsEditor` / `PatientSelectionTable` を足すと **14**。

### 500–799 行ファイル（関数抽出の主戦場。ファイル切断は二次）

| 行 | ファイル |
|----|----------|
| 726 | `examinations/hooks/use-examination-form.ts` |
| 589 | `medical-records/.../TreatmentRow.tsx` |
| 587 | `master/components/ExamTypeFieldsEditor.tsx` |
| 566 | `ReservationFormModal/PatientSelectionTable.tsx` |
| 553 | `medical-records/routes/MedicalRecordForm.tsx` |
| 531 | `master/routes/lab-device-item-master-settings-model.ts`（model 抽出済みの表。無理に割らない候補） |
| 502 | `examinations/routes/ExaminationForm.tsx` |

### 400–499 行ファイル（150 関数のあと）

IdentityLinksPage 485 · MedicalRecordFormPanels 481 · LabDeviceBoard 463 · EstimateForm 463 · use-reception-kanban 448 · PetSubOwnersSection 438 · InventoryList 434 · MedicalRecords 423 · LstepTagConfigSection 417 · TreatmentsTab 413 · HospitalizationForm 410 · OwnerForm 404 · use-medical-record-form 401。`paths.ts` 411 は対象外表。

300–399 が 55 ファイル、200–299 が 106。**機械分割しない。**

### FE-SIZE の進め方

BE と同じ帯: まず 150 超を helper に出す。フォームフックは permission ref / 死亡 sentinel / `useActionState` / entity-read を同じファイルに残し、表・payload・lock 判定だけ出す（臨床契約をファイル横断で散らさない）。

`line-reserve` の `App` はページコンポーネント済み。スイッチを `render-page.tsx` に出す程度。shared-liff 昇格はしない。

`lab-device-item-master-settings-model.ts` は既に model。表の分割は可読性が落ちるなら残す。

---

## FE-DRY — フォームフックの重複

`useExaminationForm` / `useVaccinationForm` / `useHospitalizationForm` / `useTrimmingForm` / `useCheckupForm` が permission `useLayoutEffect` + 死亡 sentinel + `useActionState` + entity-read を繰り返す。

**共通 `useEntityForm` は作らない**（YAGNI。契約が domain ごとに違う）。サイズ分割のとき同じファイル内 helper に留める。横断抽象は 150 ゲートの後、重複がまだ実害のときにだけ。

`lib/transforms/owner.ts` の `pet as unknown as PetResponse` は PetInOwner subset の再利用。構造を変えると wire が壊れる。コメントで例外固定。一括削除しない。

---

## FE-CLINICAL — 死亡 sentinel / 権限 / 日付

CLAUDE.md FE12 の3則。実装の参照は `OwnerPetsSection`（非表示 + callback の二重拒否）。

本スキャンで `status === "死亡"` はフォーム・セレクタ・バナーに既にある。**新しい抜けをファイル単位で断言しない。** 新規 pet 操作を足すときだけ二重防壁に揃える。

権限: examination フォームは `permissionsRef` + `useLayoutEffect`。他フォームも同型。分割時に ref 同期を落とさない。

日付: 生産に UTC `toISOString().slice(0,10)` は無い。`isPastJSTDate` を期限判定に使い続ける。

FE12 却下3件は「対象外」表。再提案しない。

---

## FE-TEST — 隣接テスト

規約: `Foo.tsx` の隣に `Foo.test.tsx`。`Foo.permissions.test.tsx` / `Foo.validation.test.tsx` は隣接とみなす。

`Stem.test.tsx` が無くても permissions がある例: `ExaminationForm` · `MedicalRecordForm` · `OwnerForm`。

### 200 行超で隣接テストが薄い（優先度は SIZE の後）

完全欠落の目立つもの:

| 行 | ファイル | メモ |
|----|----------|------|
| 463 | `lab-device/routes/LabDeviceBoard.tsx` | ボード本体の隣接テストなし |
| 376 | `CheckupsTab/CheckupsTabRows.tsx` | 親 `CheckupsTab.test.tsx` あり。Rows 単体なし |
| 392 | `ReservationFormModalPanels.tsx` | 親 `ReservationFormModal.test.tsx` あり |
| 266 | `use-medical-record-auto-create.ts` | 関連テストは form 側 |

`query-keys.ts` / `paths.ts` / `design-tokens.ts` は表。新規キーのテストは `design-tokens.test.ts` 型で足りる。`IdentityLinksPage.test.tsx` は存在する。

テスト追加は挙動を固定するときに限る。カバレッジ目的の描画テストを量産しない。

`PageLayout` の `resource` を使う route の render test は `useAuth` mock 必須（CLAUDE.md）。

---

## FE-HOOKS — eslint-disable ratchet

`.eslint-disable-baseline` は根拠なし disable **22**（2026-07-04）。新規は `-- 理由` 必須。既存 22 を本台帳で一掃しない。

根拠なしが残っている例（ratchet 内）:

- `PetEditModal.tsx` exhaustive-deps
- `TreatmentRow.tsx` exhaustive-deps
- `MedicalRecords.tsx` exhaustive-deps
- `OwnerForm.tsx` exhaustive-deps
- `NavigationBlocker.tsx` exhaustive-deps
- `use-accounting-detail-state.ts` / `use-examination-form.ts` の `set-state-in-effect` ブロック

SIZE 分割で触るファイルでは、触った disable に理由を足す。件数は増やさない。減ったら baseline を下げる。

`components/ui/button.tsx` の react-refresh disable は shadcn。触らない。

---

## FE-LIFF / line-reserve

- `@/` alias で `shared-liff` を参照（FE7-3 Option B）。昇格しない
- `line-reserve/src/App.tsx` 327 行は FE-SIZE の `App`
- `ConfirmPage` は `useActionState` 済み
- hex は `shared-liff/brand-tokens.ts` に閉じる
- staff の `C` / `STYLE` を LIFF に持ち込まない

---

## 推奨スライス順

挙動を変えない。1 スライス 1 テーマ。scoped vitest。full lint はユーザー。

| 弾 | 内容 | 完了条件 |
|----|------|----------|
| **0** | FE-ARCH-001（CheckupsTab → checkups barrel）+ FE-QUERY-001（unpaid prefix をファクトリーへ） | 当該ファイルの deep import 0。queryKey リテラル 0。CheckupsTab / CreditCorrection の既存テスト |
| **1** | FE-DOC: CODING_RULES の features 数と `utils/`、hooks/CLAUDE.md の queryKey 例 | ドキュメントのみ。ランタイム検証不要 |
| **2** | FE-REACT-001 ReservationFormModal pending | モーダル既存テスト。手動 isSubmitting 削除 |
| **3** | FE-CONFIRM-001 `window.confirm` 2 箇所 | side-peek / manual の破棄経路が ConfirmDialog または NavigationBlocker |
| **4** | FE-STYLE-001/002 印刷とエラー帯 | 印刷 HTML のクラスがトークン。見た目の色値は変えない |
| **5** | FE-SIZE 150 超。`useExaminationForm` から（ファイルが最長） | 150 超が対象外表の意図的残置以外 0。examination の form/validation/permissions テスト |
| **6** | 残 150 超フォームフック（vaccination → medical-record → trimming → hospitalization → owner → checkup → inventory）と巨大コンポーネント（TreatmentRow / ExamTypeFieldsEditor / PatientSelectionTable / ClinicMasterSettings / CsvImport） | 同上 |
| **7** | `line-reserve` `App` のスイッチ抽出 | App テスト |
| **8** | 400–499 のうち、弾 5–6 でまだ 150 超関数が残るファイルだけ | 800 超えを出さない |
| **後** | FE-ARCH-005 TASK-444 の models 移行。LabDeviceBoard の隣接テスト。eslint-disable 理由の任意追加 | allowlist を増やさない |

---

## 完了条件チェック（FE。実装後に埋める）

- [ ] CheckupsTab から `@/features/checkups/...` の deep import が 0
- [ ] 生産の `queryKey: [...]` リテラルが `query-keys.ts` 以外 0
- [ ] ReservationFormModal に手動 `setIsSubmitting` が無い
- [ ] 生産の `window.confirm` が 0
- [ ] 印刷面の `bg-gray-*` / `text-red-*` がトークン化されているか、ファイル単位の例外コメントがある
- [ ] 150 行超関数が、意図的残置以外 0
- [ ] 800 行超ファイルが `design-tokens.ts` 以外 0
- [ ] 公開 JSON / 画面契約 / queryKey タプル文字列を変えていない
- [ ] 死亡 sentinel と permission ref の二重防壁を分割で落としていない
- [ ] generated/models の allowlist を増やしていない
- [ ] `utils/` と `__tests__/` を再作成していない
- [ ] scoped vitest のみ。full `pnpm lint` / `test:run` はユーザー

---

## 検証メモ

- `docker compose exec -T frontend npx vitest run <path>`（`pnpm test:run -- <path>` は全件になる）
- `tsc --noEmit` はテストファイルを見ない。import 改名は 3 アプリ grep + vitest
- 複数 feature をまたぐときは `app/pages` か barrel。相対 cross-feature は禁止
- 他セッション WIP（`claim/DB-INIT-SCHEMA-HARDENING` の migration / testdb / erd / `.prime`）は触らない
