# フロントエンド リファクタリング計画

- **作成日**: 2026-07-07
- **更新日**: 2026-07-10 — R-F1〜R-F8（CLOSED）のタスク定義・完了ログ・§1解消済みFD行を本文から削除し、未対応バックログ（R-F3・R-F9〜R-F25）のみの正本に再構成した。**R-F1〜R-F8は完了のため本文から削除。履歴は git で参照**（commit `2acd7854`以前に全文が残る）。
- **対象**: `frontend/` の3アプリ全体 — `src/`（メインアプリ・1113 .ts/.tsxファイル・26 feature）＋ `liff/`（11ファイル）＋ `line-reserve/`（30ファイル）
- **スタック**: React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query
- **性格**: 全項目 **behavior-preserving（挙動保存）** を原則とする負債返済計画である。振る舞いを変える修正が必要な項目（一部のバリデーション追加等）は該当箇所に明記し、別コミット（`fix:`）として分離する。本計画自体はコード変更を行わない設計書であり、実装は本計画をもとに別途着手する。
- **根拠**: 2026-07-07、13軸の並列コード監査（feature indexing / cross-feature import / design tokens / React 19パターン / 型安全性 / 条件レンダリング / ファイルサイズ / ディレクトリ構造・命名規則 / hooks配置 / テストカバレッジ質的分析 / アクセシビリティ / liff・line-reserve規約 / 未使用コード検出基盤+パフォーマンス）を実施し、既知の機械監査済み範囲（design-system-audit.mjs、eslint-disable根拠コメントratchet、frontend/.coverage-baseline、docs/UI_DESIGN_COMPLIANCE.md§2）を除外した上で約150件の個別指摘を確認した。前回のFE-refactor.md（R-F1〜R-F7、PR #218で完了・アーカイブ済み）が対象としたlstep API層・design-tokens・eslint-disable監査・PrintPortal・カバレッジratchet・line-reserveテスト整備・type-check3アプリ化とは別の観点であり、重複しない。

---

## 1. 現状評価（2026-07-07 実測、CLOSED分は2026-07-09/10完了）

### 健全な点（是正不要と判断する根拠）

| 観点 | 実測値 | 評価 |
|---|---|---|
| Feature Indexing（deep import） | `@/features/xxx/(api\|components\|hooks\|routes\|types\|loaders)` 形式の直接import 0件（52件のbarrel経由importを全数確認、静的・動的・相対パス迂回・エイリアス抜け道いずれも無し） | 規約完全準拠 |
| React 19パターン逸脱 | `FC<`/`React.FC`/`forwardRef(`/フォーム手動loading管理（`setIsLoading`等）いずれも0件。36ファイルで`useActionState`を正しく使用 | 規約完全準拠 |
| 条件レンダリング（`&&`）アンチパターン | `{cond && <JSX>}`形式0件。過去の包括是正（commit b7a5a342で68→0件、以降複数の個別是正コミット）が定着し再発なし | 規約完全準拠。ただしESLintでの機械強制（`react/jsx-no-leaked-render`相当）は未導入で手動規律に依存 |
| any型使用 | 明示的`any`（`: any`/`<any>`/`as any`/`any[]`/`Record<string, any>`）はアプリケーションコードに実質0件。唯一の出現は`src/types/generated/models.ts`（tygo自動生成、19箇所、eslint.config.js ignoreで対象外） | ほぼ完全準拠 |
| design-system-audit.mjs対象範囲（`src/features/**/routes/**`・`**/pages/**`のhex直書き・legacy accent・colorVariant） | 2026-07-06監査時点0件、CI zero-tolerance gateで新規混入を検知 | 機械監査運用中 |
| eslint-disable根拠コメント | R-F3（既存PR #218）で33件監査・分類済み。`frontend/scripts/check-eslint-disable-rationale.mjs`のratchetで新規増加のみ検知 | 運用中。本計画では再監査しない |
| UI Design Compliance（84リーフルート） | `docs/UI_DESIGN_COMPLIANCE.md`§2で2026-07-06監査済み・83準拠/1対象外 | 運用中。本計画では再監査しない |
| frontendカバレッジratchet | `frontend/.coverage-baseline`（43.78%、2026-07-05 arm済み）で低下をCI検知 | 運用中 |
| R-F1〜R-F8（FD1・FD3・FD4・FD5・FD6・FD11・FD12行メモ化） | 全項目CLOSED（完了日2026-07-09/10） | 完了。詳細は git 履歴（commit `2acd7854`以前）参照 |

### 残存する負債

| FD# | 負債 | 規模の目安 | リスク |
|---|---|---|---|
| FD2 | ディレクトリ構造・命名規則逸脱 | 約101件（*Model.ts等57件、hooks配置ミス11件、feature構造逸脱3件、その他） | 単発ミスでなく定着した「非公式ローカル規約」化。新規参加者・AIエージェントが誤って模倣するリスク |
| FD7 | ファイル・コンポーネントサイズ超過 | 400-800行帯13ファイル | 複数責務が単一関数/コンポーネントに平坦に同居。プロジェクト自身のCODING_RULES.md基準にも抵触 |
| FD8 | テストカバレッジの質的ギャップ | 11件（CRITICAL1・HIGH多数） | 「テストがある/ない」の粗い比率とリスクの高低が一致しない逆転現象あり。過去に複数回バグ修正された箇所が無防備 |
| FD9 | アクセシビリティ逸脱 | 53件（代表列挙） | 共有コンポーネント経由で多数画面に伝播する構造的パターン。受付ボードという日常業務中核画面にも波及 |
| FD10 | liff/line-reserveアプリ固有の規約逸脱 | 8件 | mainアプリで既に修正済みの障害クラス（BUG-067）が別アプリで再現しうる実害あり |
| FD12 | パフォーマンスパターン欠如（残: useDeferredValue・lazy。行メモ化領域はR-F8で解消済み） | 代表10件中 useDeferredValue3件・lazy3件が未着手 | 主要一覧画面は模範的だが、横展開されていない周辺領域に集中 |

---

## 2. 未対応タスク一覧

規模: S=半日以内 / M=1日 / L=2-3日。各R-F項目は独立コミットとする。旧Phase番号は着手順の参考として見出しに残す（完了Phaseの見出しは除去済み）。

### Phase 1: ドキュメント・命名規則・配置の是正（低リスク・機械的）

#### R-F3. ディレクトリ構造・命名規則の是正（FD2）— 規模 L

- **現状**: `frontend/src/features/CLAUDE.md`必須規則（コンポーネント=PascalCase.tsx、フック/ユーティリティ=kebab-case.ts、標準構成api/components/hooks/routes/types）に対し、体系的な逸脱が101件存在する。
  1. **非コンポーネントロジック/定数ファイルのPascalCase.ts命名（`*Model.ts`等）57件**: 10 feature（accounting/aggregation/clinic-settings/line-reservation/lstep/master/medical-records/reception/reservations/trimming）に散在。master配下だけで35件（routes/配下20件・components/配下15件）。加えて同種のPascalCase非コンポーネントファイルとして`trimming/components/TrimmingFormColumnTypes.ts`、`trimming/components/TrimmingFormColumns.ts`、`reservations/components/WeekViewGridConstants.ts`、`master/components/StaffSidePanelSelection.ts`（use接頭辞ですらないフックexport）、`settings/integrations/lstep/LstepSettingsFormRequest.ts`も同一違反。
  2. **カスタムフックがhooks/ディレクトリ外・camelCase命名 11件（6 feature）**: `accounting/routes/{useAccountingSettlementActions.ts, useAccountingItemActions.ts, useAccountingCompletionAction.ts, useAccountingDetailState.ts}`（4件）、`owners/components/use-line-integration-card-state.ts`（kebabだがcomponents/直下）、`trimming/routes/useTrimmingHistory.ts`、`master/components/useMedicineTableState.ts`、`reception/routes/{useReceptionModalHandlers.ts, useReceptionDragHandlers.ts}`、`medical-records/routes/{useMedicalRecordPostSave.ts, useMedicalRecordDirtyFields.ts}`。
  3. **feature構造が標準（api/components/hooks/routes/types）から大きく逸脱 3 feature**:
     - `settings`: 実体（約22ファイル）が`settings/integrations/lstep/`に2階層下で隔離。内部もroutes/components区分なしにフラット配置、hooks/配下4ファイルがcamelCase。
     - `lstep`: 26 feature中唯一`pages/`ディレクトリを使用（`routes/`規約違反、5ファイル）。加えて`checkup-sync/`という非標準フラットディレクトリ（4ファイル、routes/components区分なし）。
     - `aggregation`: 26 feature中唯一components/・routes/の区分が皆無。ダッシュボードページと6個の補助コンポーネントが全てfeatureルート直下にフラット配置。
  4. **その他**: `owners/hooks/use-animal-species.ts`と`checkups/hooks/use-checkup-form.ts`がhooks/内で`@/lib/axios`を直接使用し、CODING_RULES.md§1.4「api/ vs hooks/ の区別」に反する（既知除外事項の`src/hooks/`直下ではなく`features/xxx/hooks/`のためこの除外は適用されない）。`master/components/available-slot-options.tsx`はcomponents/配下.tsxで唯一kebab-case命名（逆方向の逸脱）。
- **あるべき姿**: 全ファイルがkebab-case.ts（非コンポーネント）/PascalCase.tsx（コンポーネント）規約に従い、標準構成（api/components/hooks/routes/types）に揃っていること。
- **手順**:
  1. **57件の`*Model.ts`系リネーム**: 機械的にkebab-caseへ一括リネーム（例: `AnimalSpeciesSettingsModel.ts` → `animal-species-settings-model.ts`）。importパスは相対パスのみのため影響範囲は同一feature内に限定される。リネーム後、以後の増加を防ぐためCIにfilenameチェック（ESLint custom ruleまたは軽量Nodeスクリプト、既存の`check-eslint-disable-rationale.mjs`と同型）を追加する。
  2. **11件のフック配置修正**: 各ファイルを対応featureの`hooks/`ディレクトリへ移動し、kebab-caseにリネームする（例: `routes/useAccountingDetailState.ts` → `hooks/use-accounting-detail-state.ts`）。移動元ファイルの相対import（`./useAccountingDetailState`等）を追従修正する。`master/components/StaffSidePanelSelection.ts`は`master/hooks/use-editable-id-selection.ts`へ移動・リネームする。
  3. **settings feature平坦化**: `settings/integrations/lstep/*`を一段引き上げ、`settings/routes/LstepSettingsPage.tsx`、`settings/components/LstepSettingsForm.tsx`等、標準構成へ展開する。4つのcamelCaseフックと`LstepSettingsFormRequest.ts`もkebab-caseへリネームする。`settings/index.ts`（現状1行の再export）を更新する。
  4. **lstep feature構造修正**: `lstep/pages/` → `lstep/routes/`にリネームし、`lstep/index.ts`と各ファイルの相対importを追従修正する。`lstep/checkup-sync/`は`lstep/routes/CheckupSyncPage.tsx` + `lstep/components/{CheckupSyncConfirmDialog,CheckupSyncFilterForm,CheckupSyncPreviewTable}.tsx`に分割する。
  5. **aggregation feature構造修正**: `AggregationDashboardPage.tsx`を`aggregation/routes/`へ、残り6コンポーネントファイルを`aggregation/components/`へ移動する。`AggregationFilterPanelModel.ts`と`aggregation-csv.ts`は`aggregation/components/`（または新設lib/）へ移動し、前者はkebab-caseにリネームする。`aggregation/index.ts`の相対importを追従修正する。
  6. **hooks内axios直接呼び出し2件の是正**: `owners/api/get-animal-species.ts`（生fetch関数+queryOptions factory+useGetAnimalSpeciesフック）を新設し、`use-animal-species.ts`はそれを呼び出すだけにする。`checkups`側は`checkups/api/create-checkup-medical-record.ts`へ2つのaxios.postを切り出す。
  7. **`available-slot-options.tsx`**: `AvailableSlotOptions.tsx`にリネームするか、利用箇所が単一であれば当該コンポーネントファイル内にインライン化して独立ファイルを削除する。
- **検証**: リネーム・移動のたびに`docker compose exec frontend npx vitest run <該当feature>`でGREEN維持を確認する。型エラーが出ないことを`docker compose exec frontend pnpm run type-check`で確認する（全体実行はユーザー手動、完了報告時にコマンド提示）。settings/lstep/aggregationの構造変更後は該当featureのimport元（router定義等）を`grep -rn "features/settings\|features/lstep\|features/aggregation" frontend/src/app`で確認し、パス変更が無いこと（index.ts経由のため変更不要のはず）を確かめる。

---

### Phase 3: パフォーマンス（中リスク・視覚的変化なし）

#### R-F9. useDeferredValue欠如の是正（FD12の一部）— 規模 S

- **現状**: 主要一覧ページの検索フィルタは`useDeferredValue`済みだが、以下の共有検索モーダルは`searchTerm`を直接`useMemo`依存に使いキー入力ごとに同期フィルタが再計算されている。
  1. `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:43,91-98`
  2. `frontend/src/components/shared/MasterSelectModal/MasterSelectModal.tsx:42,44-48`（複数featureから汎用的に呼ばれ影響範囲が広い）
  3. `frontend/src/components/shared/ReservationFormModal/ReservationTypePickerDialog.tsx:48,72-81`（TreatmentSearchDialog.tsxと実装が酷似、コピー起源の可能性）
- **あるべき姿**: `OwnersList.tsx`等の参照実装同様、`deferredSearchTerm = useDeferredValue(searchTerm)`を経由してからフィルタする状態。
- **手順**: 3ファイルそれぞれに`const deferredSearchTerm = useDeferredValue(searchTerm);`を追加し、該当useMemoの依存配列と内部の`searchTerm`参照を`deferredSearchTerm`に置き換える。`TreatmentSearchDialog.tsx`は入力中フィードバックとして`const isFiltering = searchTerm !== deferredSearchTerm;`を追加しリストに`opacity-60`等を出す（`OwnersList.tsx`パターン踏襲）。3ファイルの実装が酷似しているため、共通の`useDeferredFilteredItems`相当のユーティリティ抽出も検討可（YAGNIの範囲で判断、必須ではない）。
- **検証**: 各モーダルを開き検索欄に高速入力してもUIがブロックされないことを目視確認する。`docker compose exec frontend npx vitest run src/components/shared`でGREEN確認。

#### R-F10. 非lazy化された重量コンポーネントのlazy化（FD12の一部）— 規模 S

- **現状**:
  1. `frontend/src/components/shared/Layout/Sidebar.tsx:7,185-189` / `frontend/src/features/auth/index.ts:7` — 認証後の全画面で常時マウントされるSidebarが、198行の`ChangePasswordDialog`をlazy化せず静的importしており、ほぼ開かれない機能のコードが全ページの初回バンドルに含まれる。
  2. `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx:11,90-96` — 156行の`ExaminationImportDialog`が静的import。同ファイル内の他モーダル（OwnerSearchModal/StaffSelectionModal/VitalsModal）は既に`MedicalRecordLazyModals.tsx`経由でlazy化されているのに、これだけ抜け落ちている。
  3. `frontend/src/features/accounting/routes/AccountingDetail.tsx:20,206-209` — 188行の`CreditCorrectionDialog`が静的import。影響範囲は他2件より狭いため優先度は最も低い。
- **あるべき姿**: 常時使われない大きめのダイアログ/モーダルがlazy化され、初回バンドルに含まれない状態。
- **手順**:
  1. `Sidebar.tsx`: `const ChangePasswordDialog = lazy(() => import("@/components/shared/ChangePasswordDialog/ChangePasswordDialog").then(m => ({ default: m.ChangePasswordDialog })));`とし、JSXを`<Suspense fallback={null}>`でラップする。
  2. `MedicalRecordExamination.tsx`: `MedicalRecordLazyModals.tsx`に`export const ExaminationImportDialog = lazy(() => import("../components/ExaminationImportDialog").then(m => ({ default: m.ExaminationImportDialog })));`を追加し、import元をそちらに切り替えて`<Suspense fallback={null}>`でラップする。
  3. `AccountingDetail.tsx`: `const CreditCorrectionDialog = lazy(() => import("../components/CreditCorrectionDialog").then(m => ({ default: m.CreditCorrectionDialog })));`とし`<Suspense fallback={null}>`でラップする。
- **検証**: `docker compose exec frontend pnpm build`のバンドル出力で対象コンポーネントが別chunkに分離されることを確認する（全体buildはユーザー手動、完了報告時にコマンド提示。個別確認は`vite build --mode development`のdry-runやNetwork tabでの動的import確認でも代替可）。各ダイアログを開く操作をして正常に表示されることを手動確認する。

---

### Phase 4: アクセシビリティ（UI変更を伴う・慎重に進める）

#### R-F11. PropertyRowのlabel関連付け是正（FD9の一部）— 規模 M

- **現状**: `frontend/src/components/shared/SidePeek/PropertyRow.tsx:8-19`がlabelを`<label htmlFor>`ではなく無関連の`<div>`テキストとして描画しており、22ファイルのマスタ管理側パネル（スタッフ/診療項目/医薬品/保険/予約種別/ケージ/病名/主訴/入院/トリミング/シフト等）の全フォーム入力が構造的にラベル未関連付けになっている。実装（実測確認済み）:
  ```tsx
  export function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
    return (
      <div className={STYLE.propertyRow}>
        <div className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
          {label}
        </div>
        <div className="flex-1 flex items-center">{children}</div>
      </div>
    );
  }
  ```
  使用例: `StaffBasicInfoSection.tsx:76-84 <PropertyRow label="資格番号"><input .../></PropertyRow>`。
- **あるべき姿**: PropertyRowのlabelが対応するinput/select/textareaと`htmlFor`/`id`で関連付けられている状態。
- **手順**:
  1. `PropertyRow`にReact 19の`useId()`で内部生成idを持たせ、`<label htmlFor={id}>{label}</label>`に変更する。
  2. `children`が単一のinput/select/textarea要素であれば`cloneElement(children, { id })`でidを注入する。または呼び出し側にlabelIdを渡す明示API（`<PropertyRow label="資格番号" inputId="license_number">`）に変更し、22箇所の呼び出しを段階的に移行する。
  3. まずPropertyRow自体の変更が全箇所に波及するため、1コンポーネント修正でROIが最も高い。呼び出し側の個別対応が必要な場合（childrenが複数要素、または既にidを持つ場合の衝突回避）はその都度対応する。
- **検証**: axe DevTools等でPropertyRow使用箇所のlabel-input関連付けエラーが解消されることを確認する。`docker compose exec frontend npx vitest run frontend/src/components/shared/SidePeek`でGREEN確認。マスタ設定ページを1つ（例: StaffSettings）開き、ラベルクリックで対応するinputにフォーカスが移ることを手動確認する。

#### R-F12. div onClick疑似ボタンのbutton化（FD9の一部）— 規模 M

- **現状**: role/tabIndex/onKeyDownを持たない`<div onClick>`による疑似ボタンが複数画面に存在する。特に受付ボード（reception）のカード操作という日常業務中核画面まで及んでいる。
  1. **`frontend/src/features/reception/components/AppointmentCard.tsx:134-141`（HIGH）** — 受付ボードのカードdivがdnd-kitの`attributes`/`listeners`で`role="button"`/`tabIndex={0}`を持つのに、`Reception.tsx:87-88`で`useSensors(useSensor(PointerSensor, {...}))`のみが登録され`KeyboardSensor`が無いためEnter/Spaceが何も発火しない「フォーカスは吸うが操作不能」な罠になっている（実測確認済み: `grep -n "useSensor" Reception.tsx`でPointerSensorのみ）。
  2. **`frontend/src/components/shared/MasterSelectModal/MasterSelectTrigger.tsx:34-38,61-64`（MEDIUM）** — 同一コンポーネント内で「未選択」状態は正しく`<button>`を使うが、「選択済み」状態（block/inline両バリアント）はdiv onClickのみに退行しており、一度アイテムを選択すると変更操作がキーボード不可になる。ExaminationForm/VaccinationForm/TrimmingFormから共通利用されているため1ファイル修正で複数フォームが直る。
  3. **`frontend/src/features/medical-records/components/TreatmentTable.tsx:129-133,256-283`（HIGH）** — 診療テーブルの保険適用トグル（`is_insurance`）がcursor-pointer付きdiv onClickで実装され、会計に影響する保険フラグをキーボードで切り替えられない。
  4. **`frontend/src/features/reservations/components/MonthView.tsx:96-104`（MEDIUM）** — 月表示カレンダーの予約チップが予約詳細を開く唯一の導線だがキーボード操作不可。
  5. **`frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx:123-130`（MEDIUM）** — features横断で使われるカルテヘッダー共通コンポーネントの「予約種別」クリック領域がキーボード操作不可。
  6. **`frontend/src/features/medical-records/components/InterviewHistory.tsx:70-74`（MEDIUM）** — 過去カルテ一覧の行展開/折りたたみトグルがキーボード操作不可。
  7. **`frontend/src/features/medical-records/components/ImageGalleryGroup.tsx:64-68`（MEDIUM）** — カルテ画像ギャラリーのサムネイルクリックがキーボード操作不可。
  8. **`frontend/src/features/reservations/components/WeekViewDayColumn.tsx:88-96`（MEDIUM・優先度低）** — 週表示カレンダー背景クリックでの新規予約作成導線。ただし同ページに「新規予約登録」の実ボタンが別途あり完全な代替導線切断ではない。
- **あるべき姿**: 操作可能な要素は`<button>`であるか、`role="button"`/`tabIndex={0}`/`onKeyDown`（Enter/Space対応）の3点が揃っている状態。
- **手順**:
  1. `AppointmentCard.tsx`: 最小修正としてdivに`onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onCardClick(appointment); } }}`を追加する（role/tabIndexは既存のattributes由来のまま活かす）。将来的にはKeyboardSensor追加でドラッグ自体もキーボード対応させるか、カード内に明示的な「詳細を開く」ボタンを設ける設計に寄せることも検討する。
  2. `MasterSelectTrigger.tsx`: 選択済み分岐の2箇所（block/inline）も`<button type="button" id={id} onClick={onClick} className={...}>`に統一する。
  3. `TreatmentTable.tsx`: Cellコンポーネントのonclick透過をやめ、このセルのみ`<button type="button" aria-pressed={item.is_insurance} onClick={...}>`に置き換える。
  4. `MonthView.tsx`: divを`<button type="button" className="... text-left w-full">`に置き換えるか、`role="button" tabIndex={0} onKeyDown`を追加する。
  5. `PatientInfoCard.tsx`: `onReservationTypeClick`が渡された分岐を`<button type="button">`に置き換える。ChevronDownはボタン内アイコンとして扱う。
  6. `InterviewHistory.tsx`: `<button type="button" aria-expanded={isExpanded} className="... w-full text-left" onClick={...}>`に置き換える。
  7. `ImageGalleryGroup.tsx`: `<button type="button" onClick={...} className="... text-left">`でラップする、またはimg.srcがURLであれば`<a href={img.src} target="_blank" rel="noopener noreferrer">`に置き換える。
  8. `WeekViewDayColumn.tsx`: 背景全体のボタン化はUI上不自然なため必須対応とはしない。代替ボタンの存在をJSDoc等に明記するに留める（優先度最低）。
- **検証**: 各修正後、Tabキーでフォーカスが移動し、Enter/Spaceで意図した操作が実行されることを手動確認する。特にAppointmentCard修正後は「Tabでカードにフォーカス→Enterでモーダルが開く」ことを確認する。`docker compose exec frontend npx vitest run <該当feature>`でGREEN確認。既存のクリック操作・ドラッグ操作が変更前後で同一に動作することを確認する。

#### R-F13. inputのaria-label付与（FD9の一部）— 規模 S

- **現状**:
  1. `frontend/src/features/owners/components/LineIntegrationCardSections.tsx:161-170,212-221` — LINE配信除外理由・注意事項理由のinputがplaceholderのみに依存しaria-label/label関連付けが一切ない。
  2. `frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx:143-305`（10箇所）、`CheckupsTabRows.tsx`（2箇所）、`TreatmentsTabParts.tsx`（1箇所） — バイタル/健診/診療編集テーブルの各`<td>`内inputがth列見出しとの視覚的位置関係のみに依存し、個別のアクセシブルネームを持たない（計約13箇所）。
- **あるべき姿**: 全inputがaria-labelまたはlabel htmlForで意味を持つ状態。
- **手順**:
  1. `LineIntegrationCardSections.tsx`: 各inputに`aria-label="配信除外理由"`/`aria-label="配信注意事項の理由"`を追加する（視覚デザイン変更不要）。
  2. `VitalsTabRows.tsx`等3ファイル: 各`<th>`に`scope="col"`を付与しつつ、行単位で意味が変わる編集inputには`aria-label={`体温 (${form.recorded_at})`}`のように動的アクセシブルネームを付与する。3ファイルとも同じ修正パターンで機械的に対応可能。
- **検証**: axe DevTools等でinput未ラベル警告が解消されることを確認する。`docker compose exec frontend npx vitest run <該当feature>`でGREEN確認。

---

### Phase 5: テストカバレッジの質的ギャップ埋め（リスクベース優先）

> 前提: feature別src/testファイル数の粗い比率（pets 0/12、line-reservation 0/12、master 7/163等）とリスクの高低は必ずしも一致しない。以下はリスクベースで優先度を再評価した結果である。

#### R-F14. 【CRITICAL】vaccinations次回接種日計算・バリデーションのテスト追加（FD8）— 規模 M

- **現状**: `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:51-65,123-165,239-275`の次回予定日自動計算（`calculateNextDate`）と接種日/次回予定日の相互バリデーションが、BUG-024/025/026/074/096として過去に複数回修正された実績のあるロジックであるにもかかわらず対応テストが一切無い（同feature内の`use-vaccinations.test.ts`は存在するが、このhookだけ抜けている）。
- **failure_scenario**: 3weeks/4weeks/1yearのスケジュール種別が追加・変更された際やdate-fnsのタイムゾーン境界（JST日跨ぎ）で`calculateNextDate`が誤った次回予定日を返しても検知されず、ワクチン接種のリマインド日がDBに誤登録される。過去にBUG-026として一度顕在化した種類のバグが無警告で再発しうる。
- **手順**:
  1. `calculateNextDate`をファイル外にexportするか同名でユニットテスト対象にし、3weeks/4weeks/1year/other/不正日付/うるう年境界のtable-driven testを追加する。
  2. `useActionState`のバリデーション分岐（`date > today`、`nextDate <= date`、新規時`nextDate < today`）を`renderHook`でケースごとにテストし、`fieldErrors`の内容を検証する。
  3. 既存のBUG-024/025/026/074/096コメントを回帰テストのケース名に対応付け、再発防止のトレーサビリティを持たせる。
- **検証**: `docker compose exec frontend npx vitest run frontend/src/features/vaccinations`でGREEN、新規テストが実際にBUGコメントの回帰シナリオを再現できることを確認する。

#### R-F15. 【HIGH】master共有状態機械のテスト追加（FD8）— 規模 M

- **現状**: masterフィーチャーの実質すべての設定ページ（動物種類/ケージ/診断病名/薬剤/保険/物販/職種/主訴/支払方法/トリミング/スタッフ/権限グループ等、約20ページ）が共有するCRUD状態機械が0テスト。
  1. `use-master-crud.ts:91-149,203-241` — 検索・PropertyFilterフィルタ・ソート・削除確認の共有ロジック。`applySorts`のlocaleCompare("ja")並び順、`defaultActiveFilterApply`のis_empty/is_not_empty判定、`pendingDeleteRef`によるタイミング依存の削除確定が未検証。
  2. `use-master-save.ts:38-109` — 作成/更新状態機械。validate失敗時のフィールドエラー設定、create/update分岐、onSuccess内のtry-catchエラーハンドリングが未検証。
- **failure_scenario**: `applySorts`や`defaultActiveFilterApply`にリグレッションが入ると、masterの全ページ（20ページ相当）で同時にフィルタ・ソート・削除が壊れるが、CIはそれを検知できない。`handleDeleteConfirm`はタイミング依存があり、削除対象がずれて別レコードを削除するリスクもある。`use-master-save.ts`側は「保存ボタンを押しても何も起きない」「トーストは出るがDBに反映されていない」といった不具合が本番で初めて発覚しうる。
- **手順**:
  1. `defaultSearchFilter`/`defaultActiveFilterApply`/`applySorts`を純粋関数としてモジュールトップレベルにexportし、フィルタ条件（is/is_not/is_empty/is_not_empty）・ソート方向・複数ソートキーの組み合わせをテーブル駆動テストで網羅する。
  2. `renderHook(useMasterCRUD)`で`handleNew`/`handleEdit`/`handleDeleteRequest`→`handleDeleteConfirm`の一連の状態遷移と`dirtyGuard.confirmDiscard`による中断をテストする。
  3. `renderHook(useMasterSave)`でモック`createMutation`/`updateMutation`を用意し、(a) validate失敗時にmutateが呼ばれずvalidationErrorが設定される、(b) `editTargetId===null`でcreateMutation経路、(c) `editTargetId`有りでupdateMutation経路、(d) onSuccessがrejectした場合にhandleApiErrorが呼ばれ`crudSetEditTarget(null)`されない、の4パターンを最低限テストする。
  4. 1つのmasterページ（例: CageSettings）をリファレンス実装としてintegration testを作り、以降の横展開の型にする。
- **検証**: `docker compose exec frontend npx vitest run frontend/src/features/master`でGREEN。

#### R-F16. 【HIGH】MedicineSettingsModel.ts純粋関数テスト追加（FD8）— 規模 S

- **現状**: `frontend/src/features/master/routes/MedicineSettingsModel.ts:44-124,126-174`の薬剤マスタカテゴリ別グルーピング（`groupFilteredMedicines`）・D&Dによるカテゴリ移動解決（`resolveMedicineDrag`）・作成/更新リクエスト構築（`buildMedicineCreateRequest`/`buildMedicineUpdateRequest`、カテゴリ行はprice強制0）が完全な純粋関数であるにもかかわらず0テスト。薬剤は会計（billing）に直結するマスタである。
- **failure_scenario**: `isCategoryMedicine`判定（parentId無し かつ price===0）と`buildMedicineUpdateRequest`のeffectivePrice算出が噛み合わなくなった場合、カテゴリ行に誤って価格が設定されたり、D&Dでの親カテゴリ変更時に`clear_parent_id`と`parent_id`が同時に送られる/送られないバグが起きても検知できず、薬剤マスタの階層構造や単価が静かに壊れる。
- **手順**: `groupFilteredMedicines`（フィルタ+検索+階層グルーピング）・`resolveMedicineDrag`（same-category/move-category/none判定）・`buildMedicineCreateRequest`/`buildMedicineUpdateRequest`（isCategory=trueでprice=0、parentId有無での差分）をそれぞれ純粋関数ユニットテストとして追加する。入力はMedicine[]のモックで十分再現可能なため実装コストは低い。
- **検証**: `docker compose exec frontend npx vitest run frontend/src/features/master/routes/MedicineSettingsModel.test.ts`（新規作成）でGREEN。

#### R-F17. 【HIGH】pets calcAge重複解消とテスト追加（FD8）— 規模 S（挙動変更を伴う場合は別チケット分離）

- **現状**: `frontend/src/features/pets/components/PetDeceasedBanner.tsx:13-25`の`calcAge(deceasedAt, birthDate)`と`frontend/src/features/pets/components/PetDeceasedRecordButton.tsx:16-27`の`calcAge(birthDate)`がほぼ同一の誕生日境界判定ロジックをコピペで重複実装している（DRY原則違反）。pets feature全体12ファイル0テストのため両方とも未検証。
- **failure_scenario**: うるう年生まれのペットや月末生まれのペットで享年/現在年齢の計算がズレた場合、片方のコンポーネントだけ修正されもう片方に同じバグが残る分岐が起きても、テストが無いため両方とも壊れたまま検知されない。カルテ上の年齢表示という飼主対応に直結する情報が誤る。
- **手順**:
  1. **挙動保存フェーズ（本計画に含む）**: 統合前でも両ファイルの現状ロジックそれぞれに対し「誕生日前日」「誕生日当日」「うるう年2/29生まれ」「未来日付」の境界値ユニットテストを追加し、統合時の回帰検知に使えるようにする。
  2. **統合フェーズ（挙動変更を伴うため別チケット`fix:`として分離）**: 共通ロジックを`frontend/src/lib`（または`@/utils`）に切り出し、`calcAgeAt(baseDate: Date, birthDate: Date): number`のような単一実装に統合する。
- **検証**: `docker compose exec frontend npx vitest run frontend/src/features/pets`でGREEN。境界値テストが実際に両実装の現状挙動を固定していることを確認する。

#### R-F18. 【MEDIUM/LOW】残りのテストギャップ埋め（FD8）— 規模 M

以下は上記4件よりリスクは低いが、体力に応じて着手する。

| 対象 | 内容 | severity |
|---|---|---|
| `frontend/src/features/master/components/PermissionRuleTableModel.ts:52-64`、`PermissionGroupSettings.tsx:92-96`、`StaffSettings.tsx:121-129` | 権限ルールマップ構築とパスワード/emailバリデーション（新規/編集判定の反転・欠落リスク） | MEDIUM |
| `frontend/src/features/inventory/hooks/use-inventory.ts:25-41` | lowStock/outOfStockサマリー集計。status値追加時にfilter条件が追従しないリスク | MEDIUM |
| `frontend/src/features/line-reservation/components/LineReservationSettingsFormModel.ts`（R-F6と一部重複） | `asJsonb`のフォールバック挙動（パース失敗を握りつぶす）が未テスト。booking_window_min/max_daysのクロスフィールドバリデーションも不在 | MEDIUM |
| `frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx:74-83` | 月内日付生成・曜日判定がインラインで0テスト。うるう年・月末日数のズレリスク | MEDIUM |
| `frontend/src/features/pets/components/PetDeceasedDialog.tsx:46-89` | 死亡記録フォームのバリデーション・mutation（キャッシュ無効化含む）が未検証 | MEDIUM |
| `frontend/src/features/closing-settings/components/SpecialPeriodSection.tsx:22-37`、`HolidaySection.tsx` | start_date/end_date、am_pm_boundary/pm_endの大小関係チェック不在 | LOW |

- **参考（模範例、対応不要）**: `owner-report`（19src/12test）・`reception`（21src/12test）はhooks/api変換のテスト同居が徹底されており、目標とすべき水準の好例。`estimates`/`manual`/`clinic-settings`は個別確認済みで実害リスクが小さいため対象から除外した。

---

### Phase 6: ファイルサイズ超過の是正（優先度: 状況に応じて）

#### R-F19. 400-800行帯13ファイルの分割（FD7）— 規模 L（ファイル毎に独立コミット）

- **現状**: `src/types/generated/models.ts`（3370行、自動生成・対象外）と`src/lib/design-tokens.ts`（1029行、7つの独立定数オブジェクトの集約・分割リスク低い）を除くと、実質的な「肥大化した手書きロジック」は以下13ファイルに限定される。

| ファイル | 行数 | 主な問題 | 分割方針 |
|---|---|---|---|
| `src/features/accounting/components/DailyAccountingTab.tsx` | 634 | 印刷用レイアウト`DailyPrintArea`（172行）+本体（259行）+4集計ヘルパーが同居 | `DailyAccountingPrintArea.tsx`へ印刷部抽出、`daily-accounting-utils.ts`へ集計ヘルパー抽出、`DailyAccountingTabParts.tsx`へ`SummaryCard`/`CatCell`抽出 |
| `src/features/medical-records/hooks/use-medical-record-form.ts` | 406 | 15個超のuseState+複数mutationハンドラが単一関数に平坦化 | `useMedicalRecordDiagnosisState`/`useMedicalRecordSaveAction`/`useMedicalRecordQuickPatchActions`に分割し薄いファサード化 |
| `src/features/owners/components/OwnerInfoSection.tsx` | 410 | 333行の単一JSX return（基本情報・住所・会員種別・割引等） | `PetEditModalFieldSections.tsx`パターンに倣い`OwnerBasicFields`/`OwnerAddressFields`/`OwnerMembershipFields`に分割 |
| `src/features/clinic-settings/components/ClinicMasterSidePanel.tsx` | 513 | 5つの`ClinicXxxProperty`インラインエディタ+StatusPillが同居 | `ClinicMasterSidePanelProperties.tsx`へ抽出（他Master系SidePanelへの再利用も検討） |
| `src/components/shared/ReservationFormModal/ReservationFormFields.tsx` | 495 | 時間ヘルパー5関数+日時/予約タイプ/担当者/メモが単一コンポーネント | ヘルパーを`reservation-time-utils.ts`へ、フィールド群を`ReservationDateTimeFields`/`ReservationTypeAndStaffFields`/`ReservationNotesField`に分割 |
| `src/features/accounting/components/ItemListCard.tsx` | 470 | 行編集ロジック（価格入力・数量変更・削除確認）が本体に同居 | `AccountingItemRow.tsx`へ行コンポーネント抽出 |
| `src/features/master/api/staffs.ts` | 414 | CRUD+権限グループ+所属医院+予約タイプ制限の4系統15関数が1ファイル | `staff-permission-groups.ts`/`staff-clinics.ts`/`staff-reservation-types.ts`に分割、`staffs.ts`はCRUD5関数のみ残す |
| `src/features/medical-records/routes/MedicalRecordForm.tsx` | 421 | 7個のモーダル開閉state+7個のハンドラ+本体JSXが同居 | モーダル開閉状態群を`useMedicalRecordFormModals()`に集約（既存`@/hooks/use-modal-state`パターン活用） |
| `src/features/trimming/hooks/use-trimming-form.ts` | 420 | 日付/リクエスト変換ヘルパー8個+フォーム状態が単一ファイル | ヘルパーを`trimming-form-utils.ts`へ抽出、バリデーション部分を`useTrimmingFormValidation`に分離 |
| `src/features/owners/hooks/use-owner-form.ts` | 417 | オーナー本体フィールド+pets配列操作(追加/削除/更新)が単一フック | ペット配列操作を`usePetFormListState`として分離 |
| `src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx` | 409 | 6コンポーネント（各50-99行、個別適正サイズ）が同一ファイルに集約 | 既に専用ディレクトリのため`EditRow.tsx`/`ItemRow.tsx`/`AddForm.tsx`/`badges.tsx`へ物理ファイル分割のみ（優先度低） |
| `src/features/medical-records/routes/MedicalRecords.tsx` | 403 | URL同期のページング/ソートロジック+COLUMNS定義+本体が同居 | URL同期部分を`useUrlSyncedListState`相当の汎用フックへ抽出（他一覧ページの類似重複解消にもなる）、COLUMNS定義を`medical-records-columns.tsx`へ分離 |

- **手順**: 各ファイルを個別コミットで分割する。分割時は既存のexport名・propsシグネチャを変更せず、内部実装のみをファイル分割する（behavior-preserving）。分割の優先順位は「単一関数/コンポーネントに複数責務が平坦に同居し可読性が最も低いもの」から着手する: `DailyAccountingTab.tsx` → `use-medical-record-form.ts` → `OwnerInfoSection.tsx` → `ClinicMasterSidePanel.tsx` → `ReservationFormFields.tsx` → 残り8ファイル。
- **検証**: 各分割ごとに`docker compose exec frontend npx vitest run <該当feature>`でGREEN維持を確認し、該当画面を手動確認する。分割前後でJSXの出力・propsの受け渡しが変わらないことを確認する。
- **備考**: 400行未満のファイル（例: `ReservationFormModalPanels.tsx` 388行、`MedicalRecordFormPanels.tsx` 391行）は閾値未達のため対象外としたが、境界線に近く今後の増築時に注意する。`PetEditModalFieldSections.tsx`（469行）は既に3セクションに分割済みで各111/136/103行と個別適正サイズのため対象外。

---

### Phase 7: liff/line-reserveアプリの規約統一

#### R-F20. line-reserve axiosへのNULLバイト対策共有化（FD10）— 規模 M・behavior変更あり

- **現状**: `line-reserve/src/api/liff-api.ts:17-19`が独自axiosインスタンス（`axios.create({ baseURL: API_BASE_URL })`）を持ち、request/response interceptorが一切無い。mainアプリの`src/lib/axios.ts`が持つNULLバイト除去（BUG-067修正）を含む共通防御ロジックが未適用。
- **failure_scenario**: `ConfirmPage.tsx`の`handleConfirm`が`customer_fields.name/phone/owner_name`・`pets[].name`・`request_text`という自由入力文字列をPOST /api/liff/:clinicId/reservationsへそのまま送信する。ユーザー入力に`\x00`が含まれると（コピペ等由来）、PostgreSQLがNULLバイトを含む文字列を拒否し500エラーになる — mainアプリで既に修正済みのBUG-067と同一クラスの障害が別axiosインスタンス経由で再現しうる。
- **手順**: `frontend/src/lib/axios.ts`のsanitizeNullBytes相当ロジックをReact非依存の共有ユーティリティ（例: `frontend/src/lib/sanitize.ts`）に切り出し、`line-reserve/src/api/liff-api.ts`のhttpClientにrequest interceptorとして追加してPOST/PATCH/PUTボディに適用する。
- **検証**: NULLバイトを含む文字列でconfirm送信をテストし、500エラーにならず正常にサニタイズされることを確認する（挙動が変わるため`fix:`コミットとして分離することを推奨）。`docker compose exec frontend npx vitest run line-reserve`でGREEN確認。

#### R-F21. use-liff.ts重複解消（FD10）— 規模 M・behavior変更あり

- **現状**: `line-reserve/src/hooks/use-liff.ts`（1-45行）と`liff/src/hooks/use-liff.ts`（49行）がほぼ同一実装だが、既に乖離している（diffで確認: pictureUrl取得とconsole.errorログ出力の2点差分）。line-reserve版は`liff.init()`失敗時の`console.error`ログ出力が欠落している。
- **failure_scenario**: 本番でline-reserveアプリの`liff.init()`が失敗した場合、エラー内容が一切ログに残らない。同じ状況がliffアプリで起きた場合は`console.error('[useLiff] init failed:', err)`が実行される。共有実装が無いため今後の修正のたびに2箇所を手動同期する必要があり、既に片方だけ修正漏れが発生している。
- **手順**: `useLiff`の正本を1箇所（例: `frontend/src/shared-liff/use-liff.ts`、既存の`@`エイリアスで両アプリからimport可能な位置）に統合し、pictureUrlを含む上位互換の戻り値型にする。line-reserve側は未使用フィールドを無視する形にする。`console.error`によるログ記録も統合先で共通化する。
- **検証**: 両アプリでliff初期化の正常系・異常系（`liff.init()`失敗）をそれぞれ手動確認し、ログ出力が両アプリで一致することを確認する。

#### R-F22. エラーハンドリング統一とリトライ導線追加（FD10）— 規模 L・behavior変更あり

- **現状**: line-reserveの日程/スタッフ/時間/コース選択系ページ（7ファイル）とliffの一部（計9箇所: `CourseSelectPage.tsx`/`StaffSelectPage.tsx`/`DateSelectPage.tsx`/`TimeSelectPage.tsx`/`TrimmingCourseSelectPage.tsx`/`TrimmingOptionSelectPage.tsx`/`MyReservationsPage.tsx`/`liff/src/hooks/use-liff-link.ts`/`liff/src/pages/PetHealthPage.tsx`）で、API失敗時のエラーハンドリングがステータスコードを一切見ずに固定文言を表示するのみで、その場の再試行導線が存在しない（例: `TimeSelectPage.tsx:35-47`の`.catch(() => { setError('空き時間の取得に失敗しました'); })`）。
- **failure_scenario**: 8ステップの予約ウィザード中にLIFF IDトークンが失効（401）しても一時的な5xx/ネットワーク断でも、ユーザーには同一の固定文言のみが表示され再試行ボタンが無い。復旧にはBackButtonで前段階まで戻るしかなく、401の場合はこの回避策も再ログイン導線が無いため失敗し続け、予約フローが詰む。
- **手順**: `frontend/src/lib/handle-api-error.ts`に相当する軽量なステータス分岐ヘルパーをline-reserve/liff双方向けに用意する（401→再ログイン促進メッセージ、5xx/network→再試行ボタン付きメッセージ）。各ページのエラー表示に「再試行」ボタンを追加する。
- **検証**: 各ページでAPIエラーを意図的に発生させ（msw等でモック）、ステータスコードに応じた適切なメッセージと再試行導線が表示されることを確認する。

#### R-F23. react-query導入または共通フェッチフック統一（FD10）— 規模 L・要方針決定

- **現状**: プロジェクト標準でありpackage.jsonにも導入済みの`@tanstack/react-query`がliff/line-reserveでは一切使われておらず、`useState(loading/error/data) + useEffect + .then/.catch/.finally`の手書きパターンが`line-reserve/src/App.tsx:53-85`含め12箇所に反復している。
- **failure_scenario**: 同一データ（例: コース一覧）を予約フロー内で行き来する度にキャッシュが無いため毎回無条件で再フェッチが走り、低速回線下では同じ画面に戻るたびローディングスピナーが再表示される。react-queryなら得られる自動リトライ（retry:1）やキャッシュ共有が無いため、R-F22の再試行導線欠如の一因にもなっている。
- **手順**: 本番バンドルサイズへの影響（react-query追加は消費者向け軽量webviewにはコスト増）を踏まえた上で方針を決定する。(a) react-queryを導入しuseQuery化する、または(b) 導入しない場合でも共通の`useFetchState(url, opts)`相当の軽量カスタムフックに12箇所の重複ロジックを集約しリトライ/エラー分岐だけは共通化する。いずれかを選択し統一する（**この方針決定自体はPO/architectの判断が必要なため、着手前に確認すること**）。
- **検証**: 導入後、バンドルサイズの増分を`docker compose exec frontend pnpm build`で確認する（全体buildはユーザー手動）。予約フロー内でのページ往復時にキャッシュが効いていることを確認する。

#### R-F24. liffデザイントークン導入（FD10）— 規模 M

- **現状**: `liff/src/pages/PetHealthPage.tsx`（40,51,58,68,92,94,109,132行）他4ファイルで`bg-green-50`/`bg-green-500`/`text-green-600`等のTailwind標準パレットクラスが21箇所直書きされている。line-reserveは`@theme`（`--color-noah-teal`等）でブランドカラーを統一している一方、liffにはセマンティックな色トークン層が一切存在しない。
- **failure_scenario**: line-reserveは既にnoah-tealブランドで統一されているのにliffだけ標準green（ブランド不一致）のまま残っている。将来ブランドカラーを変更する際、`LoadingPage.tsx`/`ErrorPage.tsx`/`LiffLinkPage.tsx`/`PetHealthPage.tsx`の21箇所を手動grepで置換する必要があり、design-system-audit.mjsのスコープ外（liff/は機械監査対象外）のため置換漏れが検出されない。
- **手順**: `liff/src/index.css`にline-reserveと同様の`@theme`ブロック（`--color-liff-brand`等）を追加し、21箇所の`green-*`クラスを`bg-liff-brand`等のセマンティッククラスに置換する。
- **検証**: 置換前後で視覚的に同一の色であることを確認する。`docker compose exec frontend npx vitest run liff`でGREEN確認。

#### R-F25. その他liff/line-reserve軽微是正（FD10）— 規模 S

- **現状**:
  1. `line-reserve/src/pages/MyReservationsPage.tsx:63,73` — `window.confirm('この予約をキャンセルしますか？')`/`alert('キャンセルに失敗しました。もう一度お試しください。')`がネイティブダイアログを使用。同ファイル内で既にエラー表示に赤枠banner（error state）を使っているのに、cancel失敗時だけalert()に切り替わり内部一貫性が無い。LIFF WebView環境ではネイティブダイアログの表示崩れやフォーカス喪失が起きやすい。
  2. `line-reserve/src/index.css:18-24` — `body { background-color: #EDF3F5; color: #212529; }`が直上の`@theme`で定義した`--color-noah-teal-light: #EDF3F5`/`--color-noah-text: #212529`と同じ値をhex直書きで重複させている。
  3. `line-reserve/src/pages/ConfirmPage.tsx:125-126` — `const data = err.response.data as SlotTakenResponse;`が実行時検証なしの型アサーション。`liff/src/api/liff-api.ts:62`の`res.json() as Promise<HealthCardResponse>`も同様。
- **あるべき姿**: ネイティブダイアログを使わずカスタムUIで統一、CSS内の値重複が無い、レスポンス型がzod等で検証されている状態。
- **手順**:
  1. `MyReservationsPage.tsx`: 同ファイル内で既に使っている赤枠インラインバナーのパターンに置き換える。
  2. `index.css`: `body { background-color: var(--color-noah-teal-light); color: var(--color-noah-text); }`に置換する。
  3. `ConfirmPage.tsx`/`liff-api.ts`: zod等で最小限のレスポンススキーマを定義しsafeParseで検証してから使用する、またはtypeofチェックによる簡易型ガード関数を挟む。
- **検証**: `MyReservationsPage.tsx`のキャンセル操作（成功/失敗両方）を手動確認する。`docker compose exec frontend npx vitest run line-reserve liff`でGREEN確認。

---

## 3. 非対象（明示的にやらないこと）

| 項目 | 理由 |
|---|---|
| `dangerouslySetInnerHTML`・PrintPortal生成HTMLのXSS監査 | セキュリティレビュー観点であり本リファクタ計画の範囲外。`security-reviewer`エージェント等で別途実施すべき |
| クリニック切替時のReact Queryキャッシュキーへの`clinic_id`包含有無 | マルチテナント境界のフロントエンド防御に関わる論点だが、本監査では未検証（**要調査**）。切替直後に前クリニックのデータが残留表示されるリスクの有無は別途調査が必要 |
| FE zodスキーマとBackend Goバリデーションの二重管理・乖離 | `docs/PRODUCT_PHILOSOPHY.md`が二重管理を戒めているが、解消には設計判断（どちらを正本にするか、共有スキーマ生成の要否）を要するため本behavior-preserving計画には含めない。別途architect判断が必要な設計課題として切り出すべき |
| `src/types/generated/models.ts`（3370行）の分割 | tygo自動生成ファイル。手動分割は次回codegenで上書きされ差分が消えるため有害。行数上限の対象外として明示的に除外する |
| `src/lib/design-tokens.ts`（1029行）の分割 | 7つの独立定数オブジェクト（PALETTE/C/BADGE/ICON/LAYOUT/STYLE/TABLE_STYLES）の集約であり、ロジックではなく純粋な定数カタログ。分割は機械的に可能だが凝集の価値が高く、明確な実害（差分の追いにくさ等）が出るまでは対象外とする |
| R-F23（react-query導入 or 共通フェッチフック統一）の方針決定そのもの | 本計画は実装手順を示すが、react-query導入の可否（バンドルサイズトレードオフ）はPO/architect判断が必要であり、決定を待ってから着手する |
| use-postal-code-lookup.ts等の機能追加・挙動改善 | 配置修正のみを行い、ロジック自体の改善（例: エラー時のリトライ）は別トラック |
| eslint-disable根拠コメント・design-system-audit機械監査済み範囲・UI Design Compliance既監査ルート | 既に運用中のガードレールがあり、本計画では再監査しない（「1. 現状評価」の健全な点を参照） |
| R-F7（vaccinations calcAge統合フェーズ）の実装 | 挙動保存フェーズ（境界値テスト追加）のみ本計画に含み、統合自体は挙動変更を伴うため別`fix:`チケットとする |

---

## 4. 実施ルール

1. **挙動保存の証明を各コミットに含める**: 既存テストGREEN維持＋触る箇所にpinテストが無ければ先に書く。**唯一の例外はR-F20/R-F21/R-F22**（liff/line-reserveのバグ修正的性格を持つ項目）— これらは`fix:`として扱い、挙動変更を明記する。
2. **FE固有の検証罠（既知）を踏まない**:
   - tscのPostToolUse hookは偽陽性がある — 型判定の正は`docker compose exec frontend pnpm run type-check`
   - vitestにパスを渡すとき`pnpm test:run -- <path>`は罠 — `--`以降が全件実行になる。scoped検証は必ず`docker compose exec frontend npx vitest run <path>`を使う
   - Radix Selectのoption閉鎖はfireEventでは再現しない — `user.click`を使う
   - render中setState + useActionStateはstale closureの原因になりうる — 状態同期はeffectで行う
3. **検証はscopedで自走**し、フルの`pnpm lint`/`pnpm test:run`/`pnpm build`/`pnpm type-check`（全体）はプロジェクトルールに従いユーザー手動（完了報告時にコマンド提示）。
4. **コミット粒度は1項目1コミット**（R-F番号をメッセージに含める）。commit前にHEAD確認・パス限定stage（並行セッション対策）。
5. **subagent・grepの結果は再検証してから採用する**。本計画の策定時にも、13軸監査結果のうち複数件（knip未運用・PropertyRow label未関連付け・CODING_RULES.md自己矛盾・AppointmentCardキーボード操作不能）を実コード読み合わせで裏付け確認した上で採用している。実装時も同様に、着手前に該当ファイルを実際にReadしてから修正すること。
6. **R-F3は独立して着手可能**: cross-feature import解消（R-F2）は完了済みのため、R-F3（ディレクトリ構造是正）はファイルパス依存の懸念なく単独で着手できる。
7. **memo化の参照安定性は3段階チェック**（コールバックの`useCallback`化／Setなど参照が変わりやすいコレクションをdepsに直接入れない／カスタムhookの戻り値オブジェクト自体が毎レンダー新規生成されていないか）で確認する（R-F8の教訓）。
8. **feature間で共有フック/型を昇格する際は、他feature配下の関連テストファイルのimportも`rg`で横断確認する**（R-F2-S18の教訓。production側のimport付け替えだけではテストが旧経路の参照に取り残されるケースを機械的に検知できない）。

---

## 5. 推奨着手順・見積もり（未対応17件）

| 旧Phase | 項目 | 項目数 | 規模合計（目安） |
|---|---|---|---|
| Phase 1（ディレクトリ構造） | R-F3（1項目） | 1 | L ≒ 2-3日 |
| Phase 3（パフォーマンス） | R-F9・R-F10（2項目） | 2 | S+S ≒ 1日 |
| Phase 4（アクセシビリティ） | R-F11〜R-F13（3項目） | 3 | M+M+S ≒ 2.5日 |
| Phase 5（テストカバレッジ） | R-F14〜R-F18（5項目） | 5 | M+M+S+S+M ≒ 3.5日 |
| Phase 6（ファイルサイズ） | R-F19（13ファイル、独立コミット） | 1 | L ≒ 3-4日（分散可能） |
| Phase 7（liff/line-reserve） | R-F20〜R-F25（6項目） | 6 | M+M+L+L+M+S ≒ 5日 |

**次の実装スライス**: **R-F9**（useDeferredValue欠如の是正、規模S、視覚的変化なし）。

**推奨着手順**: R-F9 → R-F10（Phase3、視覚変化なしで低リスク）→ R-F3（Phase1、規模Lだが独立着手可能なので並行して着手可）→ R-F14（Phase5 CRITICAL、回帰リスクが最も高いため優先）→ R-F11〜R-F13（Phase4、アクセシビリティ・UI変更を伴うため慎重に）→ R-F15〜R-F18（Phase5残り）→ R-F19（Phase6、体力に応じて分散）→ R-F20〜R-F25（Phase7、R-F23のみ方針決定を待つ）。

**残る完了条件**（既存の完了条件のうちR-F1〜R-F8分は達成済みのため除外、未達分のみ）:
- `*Model.ts`等のPascalCase非コンポーネントファイル命名違反が0件、settings/lstep/aggregationが標準構成（api/components/hooks/routes/types）に揃っている（R-F3）
- PropertyRow経由の22ファイルでlabel-input関連付けが機能している（R-F11）
- 受付ボード（AppointmentCard）を含む主要な疑似ボタンがキーボード操作可能（R-F12）
- vaccinations次回接種日計算・master共有CRUD状態機械・薬剤マスタ純粋関数にユニットテストが存在する（R-F14〜R-F16）
- line-reserveの予約作成フローがNULLバイト対策済みaxiosインスタンスを使用し、API失敗時に再試行導線を持つ（R-F20・R-F22）
- liff/line-reserveのuse-liff.tsが単一実装に統合されている（R-F21）
