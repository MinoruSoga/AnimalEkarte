# フロントエンド リファクタリング計画

- **作成日**: 2026-07-07
- **対象**: `frontend/` の3アプリ全体 — `src/`（メインアプリ・1113 .ts/.tsxファイル・26 feature）＋ `liff/`（11ファイル）＋ `line-reserve/`（30ファイル）
- **スタック**: React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query
- **性格**: 全項目 **behavior-preserving（挙動保存）** を原則とする負債返済計画である。振る舞いを変える修正が必要な項目（一部のバリデーション追加等）は該当箇所に明記し、別コミット（`fix:`）として分離する。本計画自体はコード変更を行わない設計書であり、実装は本計画をもとに別途着手する。
- **根拠**: 2026-07-07、13軸の並列コード監査（feature indexing / cross-feature import / design tokens / React 19パターン / 型安全性 / 条件レンダリング / ファイルサイズ / ディレクトリ構造・命名規則 / hooks配置 / テストカバレッジ質的分析 / アクセシビリティ / liff・line-reserve規約 / 未使用コード検出基盤+パフォーマンス）を実施し、既知の機械監査済み範囲（design-system-audit.mjs、eslint-disable根拠コメントratchet、frontend/.coverage-baseline、docs/UI_DESIGN_COMPLIANCE.md§2）を除外した上で約150件の個別指摘を確認した。前回のFE-refactor.md（R-F1〜R-F7、PR #218で完了・アーカイブ済み）が対象としたlstep API層・design-tokens・eslint-disable監査・PrintPortal・カバレッジratchet・line-reserveテスト整備・type-check3アプリ化とは別の観点であり、重複しない。

---

## 1. 現状評価（2026-07-07 実測）

### 健全な点（是正不要と判断する根拠）

| 観点 | 実測値 | 評価 |
|---|---|---|
| Feature Indexing（deep import） | `@/features/xxx/(api\|components\|hooks\|routes\|types\|loaders)` 形式の直接import 0件（52件のbarrel経由importを全数確認、静的・動的・相対パス迂回・エイリアス抜け道いずれも無し） | 規約完全準拠 |
| React 19パターン逸脱 | `FC<`/`React.FC`/`forwardRef(`/フォーム手動loading管理（`setIsLoading`等）いずれも0件。36ファイルで`useActionState`を正しく使用 | 規約完全準拠 |
| 条件レンダリング（`&&`）アンチパターン | `{cond && <JSX>}`形式0件。過去の包括是正（commit b7a5a342で68→0件、以降複数の個別是正コミット）が定着し再発なし | 規約完全準拠。ただしESLintでの機械強制（`react/jsx-no-leaked-render`相当）は未導入で手動規律に依存 |
| any型使用 | 明示的`any`（`: any`/`<any>`/`as any`/`any[]`/`Record<string, any>`）はアプリケーションコードに実質0件。唯一の出現は`src/types/generated/models.ts`（tygo自動生成、19箇所、eslint.config.js ignoreで対象外） | ほぼ完全準拠。回避経路（`as unknown as T`等）の構造的ギャップはR-F6で解消済み（CLOSED 2026-07-10、FD5参照） |
| design-system-audit.mjs対象範囲（`src/features/**/routes/**`・`**/pages/**`のhex直書き・legacy accent・colorVariant） | 2026-07-06監査時点0件、CI zero-tolerance gateで新規混入を検知 | 機械監査運用中。本計画では対象外範囲（components/shared、features/**/hooks等）のみ追加監査した→FD4 |
| eslint-disable根拠コメント | R-F3（既存PR #218）で33件監査・分類済み。`frontend/scripts/check-eslint-disable-rationale.mjs`のratchetで新規増加のみ検知 | 運用中。本計画では再監査しない |
| UI Design Compliance（84リーフルート） | `docs/UI_DESIGN_COMPLIANCE.md`§2で2026-07-06監査済み・83準拠/1対象外 | 運用中。本計画では再監査しない |
| frontendカバレッジratchet | `frontend/.coverage-baseline`（43.78%、2026-07-05 arm済み）で低下をCI検知 | 運用中。底上げ自体は本計画の対象外（FD8はratchetでは捕捉できない質的ギャップを扱う） |

### 残存する負債

| FD# | 負債 | 規模の目安 | リスク |
|---|---|---|---|
| FD1 | ~~feature間直接依存（cross-feature import）禁止パターン違反~~ **解消済み（R-F2, CLOSED 2026-07-09）** | 26 feature中12 featureで38件（25ファイル）→ 0件（意図的例外1件） | R-F2-S1〜S18（18コミット）で全25件対応方針を実施。最終監査は[R-F2完了ログ](#r-f2-完了ログcloseD-2026-07-09)参照 |
| FD2 | ディレクトリ構造・命名規則逸脱 | 約101件（*Model.ts等57件、hooks配置ミス11件、feature構造逸脱3件、その他） | 単発ミスでなく定着した「非公式ローカル規約」化。新規参加者・AIエージェントが誤って模倣するリスク |
| FD3 | ~~src/hooks/配置ミス~~ **解消済み（R-F4, CLOSED 2026-07-09）** | 2件 → 0件 | R-F4（commit `b8dccb77`）でAccountingDocumentのuseClinicTaxRates SSOT統一とuse-postal-code-lookupのfeatures/owners/hooks/移設を実施。詳細は[R-F4完了ログ](#r-f4-完了ログcloseD-2026-07-09)参照 |
| FD4 | ~~Design Tokens残存hex（機械監査の盲点）~~ **解消済み（R-F5, CLOSED 2026-07-09）** | 2件 → 0件 | design-system-audit.mjsの正規表現（引用符付きhex）をすり抜ける10進rgba表記。R-F5（commit `e372e272`）で解消。詳細は[R-F5完了ログ](#r-f5-完了ログcloseD-2026-07-09)参照 |
| FD5 | ~~型安全性の構造的ギャップ~~ **解消済み（R-F6, CLOSED 2026-07-10）** | 4件+lintゲート未強化 → 0件/ゲート化済み | R-F6-S1〜S4（4コミット）で無検証キャスト4件の解消と`no-explicit-any`のlint error格上げを実施。詳細は[R-F6完了ログ](#r-f6-完了ログcloseD-2026-07-10)参照 |
| FD6 | ~~CODING_RULES.md記載内容の自己矛盾~~ **解消済み（R-F1, CLOSED 2026-07-09、監査PASS・コード変更なし）** | 1件（ドキュメントのみ）→ 0件（先行是正済みを監査確認） | 実害無し。将来の誤学習・誤実装のリスク。詳細は[R-F1完了ログ](#r-f1-完了ログcloseD-2026-07-09)参照 |
| FD7 | ファイル・コンポーネントサイズ超過 | 400-800行帯13ファイル | 複数責務が単一関数/コンポーネントに平坦に同居。プロジェクト自身のCODING_RULES.md基準にも抵触 |
| FD8 | テストカバレッジの質的ギャップ | 11件（CRITICAL1・HIGH多数） | 「テストがある/ない」の粗い比率とリスクの高低が一致しない逆転現象あり。過去に複数回バグ修正された箇所が無防備 |
| FD9 | アクセシビリティ逸脱 | 53件（代表列挙） | 共有コンポーネント経由で多数画面に伝播する構造的パターン。受付ボードという日常業務中核画面にも波及 |
| FD10 | liff/line-reserveアプリ固有の規約逸脱 | 8件 | mainアプリで既に修正済みの障害クラス（BUG-067）が別アプリで再現しうる実害あり |
| FD11 | ~~未使用コード検出基盤（knip）の欠落~~ **解消済み（R-F7, CLOSED 2026-07-10）** | 1件 → 稼働化済み（棚卸し227件は別チケット） | R-F7（commit `5fe32a64`）でdevDependency・script・CI non-blockingステップを導入し稼働化。詳細は[R-F7完了ログ](#r-f7-完了ログcloseD-2026-07-10)参照 |
| FD12 | パフォーマンスパターン欠如（memo/useDeferredValue/lazy）— 行メモ化領域は**解消済み（R-F8, CLOSED 2026-07-10）**、useDeferredValue・lazy化は未着手 | 代表10件 → 行メモ化3件解消（R-F8, CLOSED 2026-07-10）。useDeferredValue/lazyはR-F9/R-F10残 | 主要一覧画面は模範的だが、横展開されていない周辺領域に集中。行メモ化領域はR-F8で解消済み、詳細は[R-F8完了ログ](#r-f8-完了ログcloseD-2026-07-10)参照。useDeferredValue・lazy領域は同種のリスクが残存 |

---

## 2. フェーズ計画

規模: S=半日以内 / M=1日 / L=2-3日。各R-F項目は独立コミットとする。

### Phase 1: ドキュメント・命名規則・配置の是正（低リスク・機械的）

#### R-F1. CODING_RULES.md自己矛盾の是正（FD6）— 規模 S

- **現状**: `frontend/CODING_RULES.md` 755-773行目「Feature公開API例」節が、同ファイル2068-2095行目（`bundle-feature-indexing`節）および`frontend/CLAUDE.md`・`frontend/src/features/CLAUDE.md`のFeature Indexing規約と逆のことを「正しい」と明記している。具体的には759-761行目が「外部からのimportはindex.ts経由でなく直接ファイルを参照する（`@/features/owners/routes/OwnersList` ← 正しい／`@/features/owners` ← 禁止）」と書いており、実際の規約・実コード（`src/app/routes/*.tsx`全11ファイルはbarrel経由の`await import("@/features/xxx")`のみを採用）と正反対である。771-772行目の「loadersはrouter.tsxから直接import」という記述も、実際にloadersがapp/routes配下で使われていない現状（grep 0件）と整合していない。
- **あるべき姿**: 755-773行目のコード例が2068-2095行目・実コードの実態と一致していること。
- **手順**:
  1. 759-761行目の「★」コメントと禁止/正しいの表記を反転させる。`export { OwnersList } from "./routes/OwnersList";`はindex.ts内部の再エクスポート定義として残しつつ、外部からのimportは`import { OwnersList } from "@/features/owners"`（barrel経由）が正、`@/features/owners/routes/OwnersList`へのdeep importが誤、と明記する。
  2. 771-772行目の「loadersはrouter.tsxから直接import」記述を、実態（loadersはfeature側からexportされfeature index.ts経由でも直接でも良い設計上の選択肢である旨）に合わせて修正するか、将来導入時の条件を明示する。
- **検証**: ドキュメントのみの変更のためコード実行検証は不要。修正後、755-773行目と2068-2095行目を読み比べ、矛盾が解消されていることを目視確認する。

#### R-F1 完了ログ（CLOSED, 2026-07-09）

- **ステータス**: **CLOSED**（完了日 2026-07-09、監査のみ・コード変更なし）
- **監査結果**: **PASS**。`frontend/CODING_RULES.md` 755–775行目（Feature公開API例）と2084–2112行目（`bundle-feature-indexing`節）を目視照合した結果、両節とも「feature外からのimportはindex.ts（barrel）経由が正、deep importが禁止」で一致していた。上記「現状」節（本タスク着手前の記述、作成日2026-07-07時点）が指摘した759-761行目・771-772行目の矛盾は、本タスク着手前の時点で既に解消されていた（是正の実施元は本タスクの範囲外・履歴未特定。「現状」節は計画作成時点の古い記述として残置し、本完了ログで最新状態を正本化する）。実際の755-775行目は「★」コメント付きでbarrel importを「正しい」・deep importを「禁止」と明記し、771-773行目もloadersを同一index.tsから公開しbarrel経由で呼び出す例のみを記載しており、2084-2112行目の「feature外はbarrel経由・feature内部は直接ファイル指定」の使い分けと矛盾しない。
- **実施した修正**: なし（`frontend/CODING_RULES.md`は変更不要、先行是正済みのため）。
- **検証**（実測）:
  ```
  $ rg '@/features/[^/]+/(api|components|hooks|routes)/' frontend/src/app
  （0件・exit 1）

  $ rg -n 'ownersLoader|ownerLoader' frontend/src/app/routes/clinical-general-routes.tsx
  46:              const [{ OwnersListPage }, { ownersLoader }] = await Promise.all([
  50:              return { Component: OwnersListPage, loader: ownersLoader };
  74:              const [{ OwnerFormPage }, { ownerLoader }] = await Promise.all([
  78:              return { Component: OwnerFormPage, loader: ownerLoader };
  ```
  両loadersとも`import("@/features/owners")`（barrel経由）から取得しており、app層でのdeep importは0件。
- **NULバイト調査**（R-F4完了ログの候補記載事項）: `python3 -c "open('FE-refactor.md','rb').read().find(b'\x00')"`でoffset 74925に1件検出。周辺文脈（R-F20節、BUG-067系NULLバイト障害の説明文）から、`` `\x00` ``という文字列表記が意図されていた箇所に生のNULバイト1バイトが混入していたものと判明。除去可能と判断し、当該バイトをテキスト表記`\x00`（4文字）へ置換して修正した（本完了ログ作成と同一コミットに含む）。
- **次エピック候補**: R-F5は完了（CLOSED 2026-07-09、commit `e372e272`、[R-F5完了ログ](#r-f5-完了ログcloseD-2026-07-09)参照）。

#### R-F2. feature間直接依存（cross-feature import）38件の解消（FD1）— 規模 L

- **現状**: `frontend/CODING_RULES.md` 1.2節「feature間の直接importは禁止。cross-feature合成が必要な場合はapp/pages/XxxPage.tsxで依存逆転により合成する」に対し、26 feature中12 featureで計38件（25ファイル）の直接import違反がある。特にmedical-records（9件）とowner-report（13件）が「何でも読みに行くハブ」と化している。一方で`app/pages/AccountingDetailPage.tsx`（accounting+master）と`app/pages/OwnerFormPage.tsx`（owners+pets+line-reservation、`lineSection`propによるスロット注入）という模範実装が既に存在するにもかかわらず、同じ画面の実装内で合成層をすり抜けて元featureが直接参照するケースが複数ある。
- **あるべき姿**: feature間の依存は(a) 既存の合成層（app/pages/）へのprops注入、(b) 複数feature共有データは`src/hooks/`への昇格、(c) 純粋なドメイン定数は`src/types/`または`src/config/`への抽出、のいずれかで解消され、feature同士が直接importしない状態。
- **対象ファイル一覧（全25件）**:

| # | ファイル | 参照先feature | 対応方針 |
|---|---|---|---|
| 1 | `features/accounting/routes/AccountingDetail.tsx:27` | cash-register (`useGetCashRegisterCloses`) | 既存`app/pages/AccountingDetailPage.tsx`でprops注入 |
| 2 | `features/clinic-settings/components/ClinicMasterSidePanel.tsx:4-8` | accounting (`DOCUMENT_SECTION_KEYS`等) | `src/types/`or`src/config/`へ定数抽出、双方がそこから参照 |
| 3 | `features/master/api/permission-groups.ts:5` | auth (`ME_QUERY_KEY`) | `src/hooks/`or`src/config/query-keys.ts`へ昇格 |
| 4 | `features/medical-records/components/CheckupsTab/CheckupsTabRows.tsx:8` | checkups (`CheckupAlertBadge`) | presentationalなら`src/components/shared/`へ昇格。domain依存があれば新設`app/pages/MedicalRecordFormPage.tsx`でrender-prop注入 |
| 5 | `features/medical-records/components/ExaminationImportDialog.tsx:13` | examinations (`useGetExaminations`, `useUpdateExamination`) | `src/hooks/`へ昇格 or 新設page層でprops注入 |
| 6 | `features/medical-records/hooks/use-medical-record-form.ts:11` | reservations (`useCreateReservation`, `useGetReservations`) | 新設`app/pages/MedicalRecordFormPage.tsx`経由でprops注入（フックが大きいため、まずreservations依存部分を薄いadapter関数に切り出す） |
| 7 | `features/medical-records/hooks/use-medical-record-auto-create.ts:7` | reservations（型のみ、`.id`のみ使用） | `{ id: string }`の最小構造型を自前定義するか`@/types`の共通型使用 |
| 8 | `features/medical-records/api/get-record-examinations.ts:5` | examinations (`transformExamResult`, `ExamResult`) | `src/lib/`or`src/types/`へ共有変換ユーティリティとして抽出 |
| 9 | `features/medical-records/routes/MedicalRecords.tsx:40` | owners (`useAnimalSpecies`) | `src/hooks/use-animal-species.ts`へ昇格 |
| 10 | `features/medical-records/routes/MedicalRecordForm.tsx:38` | owners (`useGetOwnerLineTags`) | `src/hooks/`へ昇格（reservationsからも参照されており実質共有フック、#12参照） |
| 11 | `features/medical-records/components/ExaminationGroup.test.tsx:6` | examinations（テストのみ、型） | #8の修正で自動解消 |
| 12 | `features/medical-records/components/__tests__/MedicalRecordExamination.test.tsx:8` | examinations（テストのみ、型） | #8の修正で自動解消 |
| 13 | `features/owner-report/components/VaccinationHistorySection.tsx:4` | medical-records (`useGetPetVaccinations`) | `src/hooks/`へ統合（既存`use-vaccinations.ts`と同種） |
| 14 | `features/owner-report/components/CheckupHistorySection.tsx:4` | checkups (`useGetPetCheckupResults`) | `src/hooks/`へ昇格 or 新設page層でprops注入 |
| 15 | `features/owner-report/routes/OwnerReport.tsx:10-11` | owners (`useGetOwner`), pets (`useGetPets`) | `OwnerFormPage.tsx`を模範に新設`app/pages/OwnerReportPage.tsx`で合成 |
| 16 | `features/owner-report/__tests__/OwnerReport.test.tsx:28,43` | medical-records, checkups（テストのみ、モック） | #13・#14の修正で自動解消 |
| 17 | `features/owner-report/api/get-pet-trimming-history.ts:4` | trimming (`transformTrimming`) | `src/lib/`or`src/types/`へ共有変換ユーティリティ抽出 |
| 18 | `features/owner-report/api/get-pet-examinations.ts:4` | examinations (`transformExamination`) | 同上（#8と合わせて一括対応可） |
| 19 | `features/owners/components/PetEditModalFieldSections.tsx:12` | pets (`PetDeceasedRecordButton`) | `OwnerFormPage.tsx`既存の`petMutations`注入パターンを拡張、またはUI専用なら`src/components/shared/`へ昇格 |
| 20 | `features/owners/routes/OwnerForm.tsx:30-31` | accounting (lazy import, `OwnerAccountingHistory`) | `OwnerFormPage.tsx`の`lineSection`と同じ手法で`accountingSection`propを新設 |
| 21 | `features/reception/routes/useReceptionModalHandlers.ts:6` | reservations (`useUpdateReservation`) | 新設`app/pages/ReceptionPage.tsx`でprops注入、または`src/hooks/`昇格 |
| 22 | `features/reservations/components/ReservationDetailModal.tsx:7` | owners (`useGetOwnerLineTags`) | `src/hooks/`へ昇格（#10と統合） |
| 23 | `features/reservations/hooks/use-reservation-actions.ts:6-7` | owners (`createOwner`), pets (`createPet`) | 新設`app/pages/ReservationsPage.tsx`でmutation注入オブジェクトを組み立てて渡す |
| 24 | `features/settings/integrations/lstep/TriggerPrioritySection.tsx:6` | lstep (`TriggerTypeLabels`) | ラベル定数なら`src/types/`or`src/config/`へ抽出 |
| 25 | `features/trimming/routes/TrimmingForm.tsx:17` | master (`useGetTrimmingCourseTypes`) | `src/hooks/`へ昇格（同ファイルが既に`use-master-items`を共有経由で使用済み、統合先を揃える） |

- **手順**:
  1. 上記25件を「(a) 昇格すべき共有フック/型」「(b) app/pages/合成層でprops注入すべきもの」「(c) 純粋定数として抽出すべきもの」の3分類に仕分ける（表の「対応方針」列が仕分け結果）。
  2. 分類(a)（#3, #4, #5, #9, #10・#22統合, #13, #14, #17, #18, #25）から着手する。`src/hooks/`への移設は機械的（import元の付け替えのみ）で、既存の`frontend/src/hooks/CLAUDE.md`「Cross-featureデータ系」表への追記を伴う。
  3. 分類(c)（#2, #24）は共有定数を`src/types/`or`src/config/`へ抽出し、両feature側からそちらを参照する形に変更する。
  4. 分類(b)（#1, #6, #7, #15, #19, #20, #21, #23）は新設`app/pages/XxxPage.tsx`または既存合成層の拡張が必要なため個別に設計判断を伴う。特に#6（use-medical-record-form.ts）と#23（use-reservation-actions.ts）は影響範囲が大きいため最後に着手する。
  5. #11, #12, #16はテストのみの依存であり、対応する本体側の修正（#8, #13, #14）が完了すれば自動的に解消する。個別対応は不要。
  6. 各ファイル修正ごとに既存テストGREEN維持を確認する。
- **検証**: `docker compose exec frontend npx vitest run <該当feature配下パス>`（`--`を挟むと全件実行になる罠に注意、`frontend/CLAUDE.md`「Scoped Test Verification」参照）。修正後に`grep -rn "@/features/" frontend/src/features/<修正済みfeature>`を再実行し、自feature以外への参照が消えていることを確認する。

#### R-F2 完了ログ（CLOSED, 2026-07-09）

- **ステータス**: **CLOSED**（完了日 2026-07-09、実装スライス R-F2-S1〜S18、全18コミット）
- **最終検証**（実測、本追記時点で再実行し一致確認済み）:
  ```
  $ rg '@/features/' frontend/src/features --glob '*.{ts,tsx}'
  frontend/src/features/accounting/routes/__tests__/AccountingRouteGuards.test.tsx:16:vi.mock("@/features/accounting", ...)
  frontend/src/features/accounting/routes/__tests__/AccountingRouteGuards.test.tsx:23:vi.mock("@/features/cash-register", ...)
  frontend/src/features/accounting/routes/__tests__/AccountingRouteGuards.test.tsx:27:vi.mock("@/features/accounting-reports", ...)
  ```
  25件の対象（下表）は全てS1〜S18で解消。残存する3行は後述の**意図的例外**であり、cross-feature import違反ではない。

- **意図的例外**: `features/accounting/routes/__tests__/AccountingRouteGuards.test.tsx`の`vi.mock("@/features/accounting")`等3行は、route guardのテストがルーティング上隣接する他featureのbarrel exportをモックするために`vi.mock`のモジュール指定子として`@/features/xxx`文字列を使用しているものであり、プロダクションコードからの実importではない。route guardテストの性質上、対象featureのbarrel全体をモック対象にする必要があるため、この形は妥当な設計であり是正対象としない。

- **実装パターン要約**（25件の対応方針は以下5パターンに収斂した）:

  | パターン | 内容 | 該当# |
  |---|---|---|
  | (a) hooks昇格 | feature間で共有されるquery/mutation hookを`src/hooks/`へ昇格 | #1, #3, #4, #5, #9, #10, #13, #14, #21, #22, #25 |
  | (b) lib transform昇格 | 変換ユーティリティを`src/lib/transforms/`へ昇格 | #8, #17, #18 |
  | (c) config定数抽出 | 共有定数を`src/types/`or`src/config/`へ抽出 | #2, #24 |
  | (d) shared component昇格 | UI専用コンポーネントを`src/components/shared/`へ昇格 | #4（CheckupAlertBadge）, #19（PetDeceasedRecordButtonクラスタ） |
  | (e) app/pages DI | 既存/新設のapp/pages/XxxPage.tsxでprops・mutation注入により合成 | #6, #7, #15, #16, #20, #23 |

  ※ #7は型依存のみのため最小構造型の自前定義（DIというより型デカップリング）、#11・#12はテスト専用の依存で#8修正時に一部自動解消、最終的にS18で個別修正。

- **スライス実装ログ**（commit hashは`git log --oneline --grep='R-F2-S'`で実測。全18コミットは`main`へpushではなく直接コミット済み、pushはRepo運用ポリシーに従い別途）:

  | Slice | commit | 対象# | 内容 | パターン |
  |---|---|---|---|---|
  | S1 | `e550e330` | #3, #9 | `ME_QUERY_KEY`/`useAnimalSpecies`を共有hooksへ昇格 | (a) |
  | S2 | `299a7718` | #10, #22 | `useGetOwnerLineTags`を共有hooksへ昇格 | (a) |
  | S3 | `e8ad5aef` | #25 | `useGetTrimmingCourseTypes`を共有hooksへ昇格 | (a) |
  | S4 | `5176495c` | #13 | `useGetPetVaccinations`を共有hooksへ昇格 | (a) |
  | S5 | `d9c3e79f` | #14 | `useGetPetCheckupResults`を共有hooksへ昇格 | (a) |
  | S6 | `4331156a` | #8, #17, #18 | trimming/examination変換を`src/lib/transforms/`へ昇格 | (b) |
  | S7 | `77d1f06c` | #4 | `CheckupAlertBadge`を`src/components/shared/`へ昇格 | (d) |
  | S8 | `6aa37753` | #5 | examination hooks（`useGetExaminations`/`useUpdateExamination`）を共有hooksへ昇格 | (a) |
  | S9 | `d5e9fe1b` | #2, #24 | 共有定数を`src/config/`へ抽出 | (c) |
  | S10 | `61bfe8a8` | #7 | medical-record auto-createをreservations型から型デカップリング | (e) |
  | S11 | `a4db9fcc` | #1 | `useGetCashRegisterCloses`を共有hooksへ昇格 | (a) |
  | S12 | `5fcf6170` | #21 | `useUpdateReservation`を共有hooksへ昇格 | (a) |
  | S13 | `15808d68` | #6 | reservation query/create hooksを共有hooksへ昇格し`use-medical-record-form.ts`から直接依存を除去 | (a)/(e) |
  | S14 | `f8fd49fb` | #15, #16 | `OwnerReport.tsx`をowner/pet共有hooksへ差し替え（テスト#16も同時解消） | (e) |
  | S15 | `1365987e` | #19 | `PetDeceasedRecordButton`クラスタを`src/components/shared/`へ昇格 | (d) |
  | S16 | `62dfad1d` | #23 | `ReservationsPage`経由でowner/pet作成関数をDI注入 | (e) |
  | S17 | `fd8139f6` | #20 | `OwnerFormPage`に`accountingSection`propを新設しDI注入 | (e) |
  | S18 | `35d6f039` | #11, #12 | examinationsテストの型参照を共有`ExamResult`型へ付け替え | (b) |

- **新設/拡張app/pages一覧**:
  - `app/pages/ReservationsPage.tsx`（新設、S16。owner/pet作成mutationのDIコンテナ）
  - `app/pages/OwnerFormPage.tsx`（拡張、S17。既存の`lineSection`スロットに加え`accountingSection`propを新設）
  - 既存の模範実装（変更なし）: `app/pages/AccountingDetailPage.tsx`

- **テストimportドリフトの教訓（S18）**: S6でproductionコードの`features/medical-records/api/get-record-examinations.ts`は`@/lib/transforms/examination`（共有先）を参照するよう更新されたが、同じ変換結果を検証する2つのテストファイル（`ExaminationGroup.test.tsx`、`MedicalRecordExamination.test.tsx`）はS6完了後も`@/features/examinations`から`ExamResult`型を直接importしたまま取り残されていた。production側のimport付け替えだけでは、同一シンボルをテスト側が別経路（旧feature）から参照し続けるケースを機械的に検知できない——`rg '@/features/' frontend/src/features`の再実行で初めて発覚した。**教訓**: featureからの型/関数の昇格時は、同一ファイルの`.test.ts(x)`だけでなく、他feature配下にある関連テストファイルのimportも`rg`で横断確認する必要がある。

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

#### R-F4. src/hooks/配置ミス2件の是正（FD3）— 規模 S

- **現状**:
  1. `src/hooks/use-clinic-tax-rates.ts`のdocstring（14-25行目）は「明細兼領収書（AccountingDocument）と同一の正本…を参照し、月次集計レポートを含む全帳票で税率表記を一貫させる」と明言している。しかし実際に参照しているのは`accounting-reports`feature3ファイル（`AccountingReportsPage.tsx`/`MonthlySummaryCards.tsx`/`MonthlyReportPrintArea.tsx`）のみで、docstringが名指しする`src/features/accounting/components/AccountingDocument.tsx`はこのフックをimportせず、98-99行目で`clinic?.reducedTaxRate ?? 0.08`/`clinic?.standardTaxRate ?? 0.1`を独自にベタ書きしている。加えてこのフックは`frontend/src/hooks/CLAUDE.md`のフック一覧表に未掲載で、棚卸し対象から漏れていた。
  2. `src/hooks/use-postal-code-lookup.ts`（zipcloud郵便番号検索）は実際の参照元が`owners/routes/OwnerForm.tsx:11`の1ファイルのみで、`frontend/src/hooks/CLAUDE.md`の「ユーティリティ系」表の他13エントリが全て2 feature以上で実利用されているのと異なり、cross-feature利用の実態がない。同CLAUDE.md 7-8行目「特定feature専用のフックはfeatures/xxx/hooks/に配置する」に反する。
- **あるべき姿**: use-clinic-tax-ratesが名実ともに「単一の正本」として全消費者から参照されている状態、use-postal-code-lookupが利用実態に即した配置になっている状態。
- **手順**:
  1. `src/features/accounting/components/AccountingDocument.tsx`98-99行目を`const { standardTaxRate, reducedTaxRate } = useClinicTaxRates();`に置き換え、`@/hooks/use-clinic-tax-rates`からimportする（既存のフォールバック値0.1/0.08と完全一致するため振る舞いは変わらない）。
  2. `frontend/src/hooks/CLAUDE.md`の「Cross-featureデータ系」表に`use-clinic-tax-rates.ts | 2 | 消費税率取得（accounting-reports / accounting・AccountingDocument）`を追記する。
  3. `frontend/src/hooks/use-postal-code-lookup.ts`を`frontend/src/features/owners/hooks/use-postal-code-lookup.ts`へ移動し、`OwnerForm.tsx:11`のimportを相対import（`../hooks/use-postal-code-lookup`）または`@/features/owners`経由（index.tsにexport追加）に変更する。`frontend/src/hooks/CLAUDE.md`の「ユーティリティ系」表から該当エントリを削除する。
- **検証**: `docker compose exec frontend npx vitest run src/features/accounting src/features/owners`でGREEN確認。AccountingDocument.tsxの税率表示が変更前後で同一であることを目視確認する（フォールバック値が一致するため数値変化は無いはず）。

#### R-F4 完了ログ（CLOSED, 2026-07-09）

- **ステータス**: **CLOSED**（完了日 2026-07-09、commit `b8dccb77`、単一スライス）
- **実施内容**:
  1. **AccountingDocument → useClinicTaxRates SSOT統一**: `AccountingDocument.tsx`98-99行目の独自ベタ書き（`clinic?.reducedTaxRate ?? 0.08`/`clinic?.standardTaxRate ?? 0.1`）を廃し、`@/hooks/use-clinic-tax-rates`の`useClinicTaxRates()`から`standardTaxRate`/`reducedTaxRate`を取得する形に統一。フォールバック値は既存と完全一致のため表示挙動は変化なし。
  2. **use-postal-code-lookup → features/owners/hooks/移設**: 唯一の消費者が`OwnerForm.tsx`のみだったため`src/hooks/use-postal-code-lookup.ts`を`src/features/owners/hooks/use-postal-code-lookup.ts`へ移動し、`OwnerForm.tsx`のimportを追従修正。`frontend/src/hooks/CLAUDE.md`のフック一覧表からも該当エントリを削除。
- **変更ファイル**（6件、うち1件はrename）:
  - `frontend/src/features/accounting/components/AccountingDocument.tsx`
  - `frontend/src/features/accounting/components/AccountingDocument.test.tsx`（スコープ外だが実施した追補: `useClinicTaxRates`の依存チェーン解決のため`use-auth`モックを追加）
  - `frontend/src/hooks/use-postal-code-lookup.ts` → `frontend/src/features/owners/hooks/use-postal-code-lookup.ts`（rename）
  - `frontend/src/features/owners/routes/OwnerForm.tsx`
  - `frontend/src/features/owners/routes/__tests__/OwnerForm.bug373.test.tsx`
  - `frontend/src/hooks/CLAUDE.md`
- **最終検証**（実測、本追記時点で再実行し一致確認済み）:
  ```
  $ rg '@/hooks/use-postal-code-lookup' frontend/src
  （0件・exit 1）

  $ rg 'useClinicTaxRates' frontend/src/features/accounting/components/AccountingDocument.tsx
  import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
    const { standardTaxRate: standardRate, reducedTaxRate: reducedRate } = useClinicTaxRates();

  $ docker compose exec frontend npx vitest run src/features/accounting src/features/owners
  Test Files  25 passed (25)
       Tests  230 passed | 3 skipped (233)
  ```
- **次エピック候補**（本スライス外、R-F2完了ログの候補を更新）: R-F1（FD6）・R-F5（FD4）はいずれも完了（CLOSED 2026-07-09、commit `592b1eb4`/`e372e272`）。詳細は各完了ログ（[R-F1](#r-f1-完了ログcloseD-2026-07-09)・[R-F5](#r-f5-完了ログcloseD-2026-07-09)）参照。

#### R-F5. Design Tokens残存hex 2件の是正（FD4）— 規模 S

- **現状**: `design-system-audit.mjs`の正規表現（`['"`]#[0-9A-Fa-f]{3,8}['"`]`、引用符付きhexのみ検出）をすり抜ける形で、`frontend/CLAUDE.md`が「❌ 禁止」の実例として明示するlegacy hex `#37352F`の10進rgba等価値が2箇所で直書きされている。
  1. `frontend/src/features/lstep/components/TagOwnerListDrawer.tsx:135` — `<ul className="divide-y divide-[rgba(55,53,47,0.09)]">`（同ファイルは既に`import { C, STYLE, ICON } from "@/lib/design-tokens";`済み）。
  2. `frontend/src/features/shifts/components/ShiftTemplateSettingsParts.tsx:191` — `placeholder:text-[rgba(55,53,47,0.15)]`（同ファイルは既に`import { C, LAYOUT, STYLE } from "@/lib/design-tokens";`済み）。
- **あるべき姿**: 両ファイルとも既存のdesign-tokens値（`C.divideDivider`/`C.textPlaceholderFaint`）を参照し、hex/rgba直書きが無い状態。
- **手順**:
  1. `TagOwnerListDrawer.tsx:135`の`divide-y divide-[rgba(55,53,47,0.09)]`を`` `divide-y ${C.divideDivider}` ``に置換する（`C.divideDivider`は`divide-[rgba(0,0,0,0.09)]`で同じ9%不透明度、現行トークンのink基準色`0,0,0`に統一）。classNameは通常文字列のため`` className={`divide-y ${C.divideDivider}`} ``へ変更する。
  2. `ShiftTemplateSettingsParts.tsx:191`の`placeholder:text-[rgba(55,53,47,0.15)]`を`${C.textPlaceholderFaint}`に置換する。
  3. 併せて`design-system-audit.mjs`のC3正規表現を、引用符付きhexだけでなく既知のlegacy 10進rgba値（`55,53,47`等）にも拡張することを検討する（このスクリプト自体の拡張は本計画の付随的推奨であり、必須項目ではない）。
- **検証**: 該当コンポーネントの表示（divideの区切り線、placeholderの薄さ）が変更前後で視覚的に同一であることを目視確認する。`docker compose exec frontend pnpm design-audit`を実行しC1/C3/C5が引き続き0件であることを確認する。

#### R-F5 完了ログ（CLOSED, 2026-07-09）

- **ステータス**: **CLOSED**（完了日 2026-07-09、commit `e372e272`、単一スライス、直前HEAD `592b1eb4`）
- **実施内容**:
  1. `TagOwnerListDrawer.tsx:135` — `className="divide-y divide-[rgba(55,53,47,0.09)]"` を `` className={`divide-y ${C.divideDivider}`} `` に置換。
  2. `ShiftTemplateSettingsParts.tsx:191` — `placeholder:text-[rgba(55,53,47,0.15)]` を `${C.textPlaceholderFaint}` に置換（`${C.text}`の後段に追加する形）。
  両ファイルとも`C`は既にimport済みのため追加import不要。置換のみでレイアウト・ロジック変更なし。
- **最終検証**（実測、本追記時点で再実行し一致確認済み）:
  ```
  $ rg 'rgba\(55,53,47' frontend/src/features --glob '*.{ts,tsx}'
  （0件・exit 1）

  $ docker compose exec -T frontend npx vitest run src/features/lstep src/features/shifts
  Test Files  4 passed (4)
       Tests  54 passed (54)

  $ docker compose exec -T frontend pnpm design-audit
  design-system-audit: C1 legacy accent — 0 件
  design-system-audit: C3 route表面 hex 直書き — 0 件
  design-system-audit: C5 非 brand colorVariant — 0 件
  design-system-audit: PASS — 違反 0 件
  ```
- **スコープ外残存**: `features/master/PATTERNS.md:342`の同型rgba記載はドキュメント内のコード例であり、実行コードではないため是正対象外（変更なし）。
- **レビュー記録**: typescript-reviewer **Approve**（CRITICAL/HIGH 0件）。MEDIUM所見1件（`C.divideDivider`/`C.textPlaceholderFaint`は色基盤が旧アクセント`#37352F`系からink基準`rgba(0,0,0,...)`系に変わる。不透明度9%/15%は完全一致）——design-tokens.ts内の他定数（`borderLight`/`bgHover`/`bgLight`等）も同じ「legacy `#37352F`→ink `0,0,0`統一」方針を既に採用済みであり、意図的なトークン統一と判断しブロック不要とした。
- **付随推奨（未実施）**: `design-system-audit.mjs`のC3正規表現をlegacy 10進rgba値（`55,53,47`等）にも拡張する検討（本計画の手順3、任意・次エピック外）。
- **次エピック候補**: R-F6は完了（CLOSED 2026-07-10、commit `da8933b4`/`9c6fab15`/`8b38402e`/`39eae262`）。詳細は[R-F6完了ログ](#r-f6-完了ログcloseD-2026-07-10)参照。

---

### Phase 2: 型安全性・未使用コード検出基盤

#### R-F6. 型安全性の構造的ギャップ解消とlintゲート格上げ（FD5）— 規模 M

- **現状**: 明示的`any`は0件だが、`unknown`経由の無検証キャスト（`as unknown as T`）や暗黙的any伝播が4件存在する。加えて`@typescript-eslint/no-explicit-any`がeslint.config.js上`"warn"`のままCIの`pnpm run lint`（`eslint .`、`--max-warnings`指定なし）にゲート化されておらず、将来anyが新規混入しても機械的に検知されない。
  1. `frontend/src/features/master/hooks/use-master-save.ts:22-23,73,90` — `createMutation`/`updateMutation`を`UseMutationResult<unknown, ...>`で受け取り、`onSuccess`内で`savedData as T`と型ガード無しでキャストしている。呼び出し元は19ルート（MasterCRUDPage.tsx等マスタ設定ページ全体）。
  2. `frontend/src/features/line-reservation/components/LineReservationSettingsFormModel.ts:4-14` — `asJsonb<T>(value: unknown, fallback: T): T`が構造検証なしに`value as T`で無条件キャストしており、`closed_weekdays`/`business_hours`/`break_hours`/`closed_dates`/`business_hours_by_weekday`の5つのJSONBフィールドが実質any化している。加えて書込み側（`LineReservationSettingsForm.tsx:100-104`）でも`closedWeekdays as unknown as string`という逆方向の二重キャストがある。
  3. `frontend/src/features/cash-register/api/transforms.ts:13` — generated modelの`category_breakdown`（any型、JSONB由来）をそのままtransform関数の返り値プロパティへ代入しており、公開型`CashRegisterClose.categoryBreakdown`が暗黙にanyへ伝播する（grepで`any`キーワードが現れないまま型システムに穴が開く不可視パターン）。
  4. `frontend/eslint.config.js:30` — `"@typescript-eslint/no-explicit-any": "warn"`のまま、CIにフェイルセーフが無い。
- **あるべき姿**: unknown値は必ず型ガードを経由してから使用され、公開型にanyが暗黙伝播しない状態。any禁止ルールが新規混入を機械的に検知する状態。
- **手順**:
  1. `use-master-save.ts`: 呼び出し側が渡す`createMutation`/`updateMutation`の型引数を`UseMutationResult<T, Error, TCreate>`のように具体型で受け取れるよう`UseMasterSaveOptions`のジェネリック制約を変更し、`onSuccess`内の`savedData as T`を削除する。互換性維持が必要な場合は最低限`'id' in savedData`等のランタイム型ガードを追加してからキャストする。
  2. `LineReservationSettingsFormModel.ts`: `asJsonb`に型ガード関数引数を追加（`asJsonb<T>(value: unknown, fallback: T, isT: (v: unknown) => v is T): T`）し、各呼び出し箇所で`Array.isArray`や構造チェックを渡す。書込み側は、tygo生成型の`string /* []byte */`宣言が実態（object/array）と乖離しているため、該当フィールドのみ上書きする専用リクエスト型（例: `type UpdateLineReservationSettingRequest = Omit<LineReservationSetting, 'closed_weekdays'|...> & { closed_weekdays: string[]; ... }`）を定義し、`as unknown as string`を撤去する。
  3. `cash-register/api/transforms.ts`: `categoryBreakdown: raw.category_breakdown as unknown,`のように明示的に`unknown`へキャストしてから返す（呼び出し側の`summarizeCategoryTotals(raw: unknown)`は既に安全に実装済みのため型注釈のみの修正で完結する）。
  4. `eslint.config.js:30`の`"@typescript-eslint/no-explicit-any"`を`"warn"`から`"error"`へ格上げする（現状違反0件のためregression-safe）。または`package.json`のlint scriptを`eslint . --max-warnings=0`に変更してCIゲート化する。既存の`check-eslint-disable-rationale.mjs`・`design-system-audit.mjs`と同じ「ratchet→zero-tolerance」の運用パターンを踏襲する。
- **検証**: `docker compose exec frontend npx vitest run src/features/master src/features/line-reservation src/features/cash-register`でGREEN確認。`docker compose exec frontend pnpm run type-check`で型エラーが無いことを確認（全体実行のためユーザー手動、完了報告時にコマンド提示）。lintルール格上げ後は`docker compose exec frontend pnpm run lint`で新規warningが無いことを確認する（同じくユーザー手動）。

#### R-F6 完了ログ（CLOSED, 2026-07-10）

- **ステータス**: **CLOSED**（完了日 2026-07-10、実装スライス R-F6-S1〜S4、全4コミット）
- **スライス実装ログ**:

  | Slice | commit | 内容 |
  |---|---|---|
  | S1 | `da8933b4` | `use-master-save.ts`の`createMutation`/`updateMutation`を`UseMutationResult<T, ...>`で受け取る形に変更し、`onSuccess`内の`savedData as T`無検証キャストを撤去 |
  | S2 | `9c6fab15` | `LineReservationSettingsFormModel.ts`の`asJsonb`に型ガード引数を追加。書込み側は`UpdateLineReservationSettingRequest`型を新設しJSONBフィールドを型安全に上書き、`closedWeekdays as unknown as string`の二重キャストを撤去。呼び出し元（`LineReservationSettingsForm.tsx`/`PageEditor`）を型変更に追従 |
  | S3 | `8b38402e` | `lib/transforms/cash-register.ts`の`categoryBreakdown`を`raw.category_breakdown as unknown`へ明示キャストし暗黙any伝播を遮断。回帰テストを`lib`単体テストと`features`統合テストの両方に追加（依存方向は`features → lib`を維持、逆依存が生じないことをレビューで確認） |
  | S4 | `39eae262` | `eslint.config.js`の`"@typescript-eslint/no-explicit-any"`を`"warn"`から`"error"`へ格上げ。`--max-warnings=0`によるCIゲート化は不採用（既存の`react-refresh`/`react-hooks`系warningまで一括で赤くする副作用を避け、`no-explicit-any`のみ狙い撃ちでゲート化する既存ratchetパターンを踏襲） |

- **最終検証**（実測、本追記時点で再実行し一致確認済み）:
  ```
  $ rg 'savedData as T|UseMutationResult<unknown' frontend/src/features/master/hooks/use-master-save.ts
  （0件・exit 1）

  $ rg 'as unknown as string|value as T|JSON\.parse\(value\) as T' frontend/src/features/line-reservation
  （0件・exit 1）

  $ rg 'categoryBreakdown: raw\.category_breakdown as unknown' frontend/src/lib/transforms/cash-register.ts
  categoryBreakdown: raw.category_breakdown as unknown,
  （1件・意図的unknown化。下流の`summarizeCategoryTotals(raw: unknown)`が安全側で型ガードする設計のため、ここでの`unknown`明示は暗黙any伝播の遮断であり残存負債ではない）

  $ rg '"@typescript-eslint/no-explicit-any": "error"' frontend/eslint.config.js
  "@typescript-eslint/no-explicit-any": "error",
  （1件）

  $ docker compose exec -T frontend npx vitest run src/features/master src/features/line-reservation src/features/cash-register src/lib/transforms/cash-register.test.ts
  Test Files  16 passed (16)
       Tests  91 passed (91)

  $ docker compose exec -T frontend pnpm run lint
  ✖ 15 problems (0 errors, 15 warnings)
  （exit 0）
  ```
  フルlintの15件警告はいずれも`react-refresh/only-export-components`・`react-hooks/exhaustive-deps`・`react-hooks/preserve-manual-memoization`・`@typescript-eslint/no-unused-vars`という既存カテゴリであり、`no-explicit-any`格上げに起因する新規errorは0件。S4完了時点でフル`pnpm run lint`が未実施だった検証ギャップは、本DOC作成時の実測で正式にPASS化した。
- **レビュー記録**: 各スライスともtypescript-reviewer **Approve**（S3は当初lib→features逆依存の疑義が指摘され、修正後にApprove）。
- **フォローアップ（別チケット・未実施）**:
  - `use-master-save.ts`自体の単体/regressionテストが欠如（S1、MEDIUM）
  - `api/types.ts` → `components/`への依存方向が逆（S2、MEDIUM。型安全上の実害はなし）
  - `additional_fields`にも同種のJSONB無検証キャスト負債が残る（S2、本スライス対象外）
- **次エピック候補**: R-F7は完了（CLOSED 2026-07-10、commit `5fe32a64`）。詳細は[R-F7完了ログ](#r-f7-完了ログcloseD-2026-07-10)参照。

#### R-F7. knip導入（FD11）— 規模 S

- **現状**: `frontend/knip.json`は存在する（2026-06-01 commit f52d8effで追加、entry/project/ignore定義済み）が、`knip`自体が`package.json`のdevDependenciesに無く、scriptsにも実行コマンドが無く、`.github/workflows/*.yml`全体にも"knip"の文字列が存在しない。設定だけ作られ一度も稼働していない。
- **あるべき姿**: knipが実際にCIで実行され、未使用export/file/dependencyを検出できる状態。
- **手順**:
  1. `docker compose exec frontend pnpm add -D knip`でdevDependenciesに追加する。
  2. `package.json`のscriptsに`"unused": "knip"`（または`"knip --reporter compact"`）を追加する。
  3. `.github/workflows/ci.yml`にknip実行ステップを追加する。いきなりfailさせると誤検知で赤くなるため、まずはnon-blockingで導入し、既存のunused export/file/dependencyを洗い出してから段階的にgate化する（既存のcoverage ratchet・eslint-disable ratchetと同じ2段階導入パターン）。
  4. 導入後に一度全件スキャンし、既存のunused exports/filesを棚卸しして別チケット化する（本計画のスコープには含めない。knipが動く状態にすることが本項目のゴール）。
- **検証**: `docker compose exec frontend pnpm run unused`（または導入したscript名）を実行しエラー無く完走することを確認する。CI上でも同様に実行され結果がJob Summary等に出力されることを確認する。

#### R-F7 完了ログ（CLOSED, 2026-07-10）

- **ステータス**: **CLOSED**（完了日 2026-07-10、commit `5fe32a64`、親コミット `dd76a7f1`）
- **実施内容**:
  - `frontend/package.json`: `knip@^6.25.0`をdevDependenciesに追加、`"unused": "knip --reporter compact --no-exit-code"`をscriptsに追加（`pnpm-lock.yaml`同時更新）
  - `.github/workflows/ci.yml`: frontend jobのLintステップ直後・Buildステップ直前に`Knip unused scan (non-gating)`ステップを追加（`continue-on-error: true`、`pnpm run unused | tee -a "$GITHUB_STEP_SUMMARY"`）
  - `frontend/knip.json`は変更なし（既存のentry/project/ignore定義のまま初回実行が完走したため設定調整は不要だった）
- **初回スキャン結果**（実測、本追記時点で再実行し一致確認済み）:
  ```
  $ docker compose exec -T frontend pnpm run unused

  Unused files (3)
  src/features/owners/api/get-animal-species.ts
  src/features/pets/api/record-pet-death.ts
  src/features/pets/api/revoke-pet-death.ts
  Unused dependencies (1)
  package.json: jsonwebtoken
  Unused devDependencies (1)
  package.json: chrome-launcher, lighthouse, tailwindcss
  Unused exports (143)
  Unused exported types (76)
  Duplicate exports (1)
  src/components/shared/UnifiedTabs.tsx: UnifiedTabs, default
  （exit 0）
  ```
- **false-positive疑いの調査記録**（コード変更なし、結論のみ）:
  - `chrome-launcher`/`lighthouse`: `frontend/scripts/lighthouse-audit.js:13-14`が`import lighthouse from 'lighthouse'`/`import * as chromeLauncher from 'chrome-launcher'`で使用している。`knip.json`の`entry`は`src/main.tsx`・`liff/src/main.tsx`・`line-reserve/src/main.tsx`の3つのみで`scripts/**`を走査対象に含んでいないため、knipから見えず未使用判定されている可能性が高い。**削除しないこと**。フォローアップ: `knip.json`の`entry`に`scripts/lighthouse-audit.js`を追加する検討（本エピックのスコープ外）。
  - `tailwindcss`: 実行時は`frontend/vite.config.ts:4,108`が`@tailwindcss/vite`（別パッケージ）をimportしてプラグイン登録しているが、`tailwindcss`本体は`frontend/src/index.css:1`の`@import "tailwindcss";`というCSS内importで消費されている。knipはCSSの`@import`によるnpmパッケージ利用を依存関係として検出できない構造的盲点があるため、これも未使用判定は誤検知の可能性が高い。**削除前に要確認**（Tailwind CSS 4のCSS-firstアーキテクチャでは`tailwindcss`パッケージそのものがCSS経由の必須依存であり、除去するとビルドが壊れる）。
- **棚卸し一覧**（修正対象外・別チケット）:
  - Unused files 3件（`get-animal-species.ts`/`record-pet-death.ts`/`revoke-pet-death.ts`）
  - Unused dependencies: `jsonwebtoken`
  - Unused exports 143件・Unused exported types 76件（feature別内訳は本追記では省略、次回棚卸しチケットで一覧化）
  - Duplicate exports: `UnifiedTabs.tsx`（named export `UnifiedTabs`と`default`の二重公開）
- **検証コマンド再実行**（本追記時点で再確認済み）:
  ```
  $ rg '"knip"' frontend/package.json
      "knip": "^6.25.0",

  $ rg '"unused": "knip' frontend/package.json
      "unused": "knip --reporter compact --no-exit-code"

  $ rg -A3 'Knip unused scan' .github/workflows/ci.yml
        - name: Knip unused scan (non-gating)
          continue-on-error: true
          working-directory: frontend
          run: pnpm run unused | tee -a "$GITHUB_STEP_SUMMARY"
  ```
- **レビュー記録**: 設定導入のみのため専門エージェントレビューは省略（`pnpm run unused`のexit 0出力で代替、本DOC作成時に再実行し一致確認）。
- **フォローアップ（別チケット・未実施）**:
  - `knip.json`の`entry`に`scripts/lighthouse-audit.js`を追加し、`chrome-launcher`/`lighthouse`のfalse-positiveを解消する検討
  - Unused exports 143件・Unused exported types 76件の段階的解消
  - `jsonwebtoken`依存の実使用有無の確認と削除判断
  - `UnifiedTabs.tsx`のduplicate exports解消（named/defaultのどちらかに統一）
  - knip CIステップのfailゲート化（第2段階、既存のcoverage ratchet・eslint-disable ratchetと同型の運用に揃える）
- **次エピック候補**: R-F8は完了（CLOSED 2026-07-10、commit `8c6b45d4`/`dfef3f87`/`f01b74cf`）。詳細は[R-F8完了ログ](#r-f8-完了ログcloseD-2026-07-10)参照。次候補: **R-F9**（FD12 useDeferredValue欠如、規模 S）。

---

### Phase 3: パフォーマンス（中リスク・視覚的変化なし）

#### R-F8. 行コンポーネント未メモ化の是正（FD12の一部）— 規模 M

- **現状**: 主要な一覧画面（owners/estimates/checkups/examinations/inventory/hospitalization/vaccinations/trimming/accounting）は既にCODING_RULES.md「memo化コンテナ+useDeferredValue+クライアントページング」パターンに高水準で準拠している一方、その模範パターンが横展開されていない周辺領域で行未メモ化が見つかった。
  1. `frontend/src/features/lstep/checkup-sync/CheckupSyncPreviewTable.tsx:19-164`（map: 112-158） — LSTEP健診案内配信の対象者プレビューテーブルがowners配列全体を行レベルメモ化なしでmapしており、ページネーションも無い。フィルタ条件合致オーナーは容易に100件を超える。
  2. `frontend/src/features/lstep/components/TagOwnerListDrawer.tsx:66-69,136-161` — LINEタグ対象者一覧ドロワーが`per_page: 200`で最大200件を一括取得し、`<li>`をメモ化なしでレンダリングしている。
  3. `frontend/src/features/master/components/MedicineTableRows.tsx:40,129` / `MedicineTable.tsx:99-107,118-126` — 薬剤マスタ設定の行コンポーネント`SortableMedicineRow`/`MedicineCategoryHeaderRow`が`React.memo`でラップされておらず、`useSortable`（dnd-kit）を伴う全行がドラッグ操作のたびに再レンダーされうる。実運用で100〜数百SKUに達する薬剤マスタかつdnd-kit配下のため他リストより効果が大きい。
- **あるべき姿**: 行コンポーネントが`memo()`でラップされ、親の再レンダーから独立している状態。
- **手順**:
  1. `CheckupSyncPreviewTable.tsx`: 行部分を`const CheckupSyncPreviewRow = memo(function CheckupSyncPreviewRow({ owner, selected, eligible, onToggle }) {...})`として切り出し、`owners.map((owner) => <CheckupSyncPreviewRow key={owner.owner_id} .../>)`に変更する。`onToggle`は`handleRowToggle`を`useCallback`化した上でowner_idを引数に渡す形にする。
  2. `TagOwnerListDrawer.tsx`: `<li>`の中身を`const TagOwnerListItem = memo(function TagOwnerListItem({ owner }: { owner: LstepTagOwner }) {...})`として切り出す。
  3. `MedicineTableRows.tsx`: `export const SortableMedicineRow = memo(function SortableMedicineRow({ medicine, onEdit, grouped, canEdit }: Props) {...});`および`MedicineCategoryHeaderRow`も同様にmemoでラップする。`onEdit`が呼び出し元（`useMedicineTableState`/`MedicineSettings`）で安定した参照か確認し、不安定なら`useCallback`化する。
- **検証**: 各修正後、React DevTools Profilerで対象行の再レンダー回数が親の状態変化時に増えないことを確認する（`.claude/refs/performance-rules.md`参照）。`docker compose exec frontend npx vitest run <該当feature>`でGREEN確認。挙動（表示内容・クリック操作）が変更前後で同一であることを手動確認する。

#### R-F8 完了ログ（CLOSED, 2026-07-10）

- **ステータス**: **CLOSED**（完了日 2026-07-10、実装スライス R-F8-S1〜S3、全3コミット）
- **スライス実装ログ**:

  | Slice | commit | 変更ファイル | 内容 |
  |---|---|---|---|
  | S1 | `8c6b45d4` | `CheckupSyncPreviewTable.tsx` | `CheckupSyncPreviewRow`をmemo化。`selectedIds`（Set）を直接depsに使うとmemoが無効化されるため`useRef`で同期し、`handleRowToggle`のdepsを`onSelectionChange`のみに絞った |
  | S2 | `dfef3f87` | `TagOwnerListDrawer.tsx` | `TagOwnerListItem`をmemo化。propsは`owner`単一・callbackなしのためuseCallback不要。`useGetLstepTagOwners`はselect変換を挟まずQuery cache参照がそのまま安定するため追加対応不要だった |
  | S3 | `f01b74cf` | `MedicineTableRows.tsx` + `MedicineSettings.tsx` | `SortableMedicineRow`/`MedicineCategoryHeaderRow`をmemo化。`handleEdit`/`handleCreate`のdepsを`[medicineCrud]`から`[medicineCrud.handleEdit]`/`[medicineCrud.handleNew]`へ絞ったが、react-reviewer初回パスで`useSidePeekDirty()`が毎レンダー新規オブジェクトを返す→`dirtyGuard`→`useMasterCRUD`内`confirmDiscard`→`handleEdit`/`handleNew`という連鎖不安定を検出（HIGH, Block）。`MedicineSettings.tsx`内に`dirtyGuard = useMemo(() => ({ confirmDiscard: dirty.confirmDiscard }), [dirty.confirmDiscard])`を追加し局所的に解消 |

- **教訓**: memoが効くかはpropsの由来を1段ずつ遡って確認する必要がある。コールバック自体の`useCallback`化だけでなく、（1）データ取得層の変換（selectオプション等でQuery cache参照が壊れていないか）、（2）Setなど参照の変わりやすいコレクション型がdepsに直接入っていないか、（3）カスタムhookの返却オブジェクト自体が毎レンダー新規生成されていないか（`useSidePeekDirty`→`dirtyGuard`→`useMasterCRUD`のように、目的の関数自体は`useCallback`で安定していても、それを包む戻り値オブジェクトが非メモ化だと連鎖的にmemoを無効化する）を段階的にチェックする。
- **最終検証**（本追記時点で再実行し一致確認済み）:
  ```
  $ rg 'CheckupSyncPreviewRow = memo' frontend/src/features/lstep/components/CheckupSyncPreviewTable.tsx
  const CheckupSyncPreviewRow = memo(function CheckupSyncPreviewRow({
  （1件）

  $ rg 'TagOwnerListItem = memo' frontend/src/features/lstep/components/TagOwnerListDrawer.tsx
  const TagOwnerListItem = memo(function TagOwnerListItem({
  （1件）

  $ rg 'export const (SortableMedicineRow|MedicineCategoryHeaderRow) = memo' frontend/src/features/master/components/MedicineTableRows.tsx
  export const MedicineCategoryHeaderRow = memo(function MedicineCategoryHeaderRow({
  export const SortableMedicineRow = memo(function SortableMedicineRow({
  （2件）

  $ docker compose exec -T frontend npx vitest run src/features/lstep
  Test Files  3 passed (3)
       Tests  45 passed (45)

  $ docker compose exec -T frontend npx vitest run src/features/master
  Test Files  7 passed (7)
       Tests  39 passed (39)
  ```
  （S1/S2対象ファイルの実パスは`frontend/src/features/lstep/checkup-sync/`ではなく`frontend/src/features/lstep/components/`配下。過去ログの記載揺れであり、実装自体に影響はない）
- **レビュー記録**: S1/S2はreact-reviewer **Approve**。S3は初回**Block**（HIGH: `useSidePeekDirty`起因のdirtyGuard連鎖不安定）→`dirtyGuard`のuseMemo修正後**Approve**、続けてtypescript-reviewerも**Approve**。
- **スコープ外の判断**: S3の根本原因である`use-side-peek-dirty.ts`（14マスタ画面が共有）および`use-master-crud.ts`は変更していない。修正は`MedicineSettings.tsx`内の`dirtyGuard`useMemoに局所化し、他13画面への影響を避けた。他画面の行コンポーネントは現状memo化されていないため、この不安定パターンは潜在的だが顕在化していない（将来他画面をmemo化する際は同様のuseMemo対応が必要になる可能性がある）。
- **次エピック候補**: **R-F9**（FD12 useDeferredValue欠如、規模 S）。

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
6. **R-F2（cross-feature import解消）とR-F3（ディレクトリ構造是正）は依存関係に注意**: R-F3でファイル移動を行うと、R-F2のファイルパス参照がずれる可能性がある。実施順序はPhase順（R-F3が先、R-F2はその後）を推奨するが、両方が同一ファイルに触れる場合は1コミットにまとめてもよい。

---

## 5. 全体見積もりと完了条件

| フェーズ | 項目数 | 規模合計（目安） |
|---|---|---|
| Phase 1（ドキュメント・命名規則・配置） | R-F1〜R-F5（5項目） | S+L+L+S+S ≒ 5日 |
| Phase 2（型安全性・検出基盤） | R-F6〜R-F7（2項目） | M+S ≒ 1.5日 |
| Phase 3（パフォーマンス） | R-F8〜R-F10（3項目） | M+S+S ≒ 2日 |
| Phase 4（アクセシビリティ） | R-F11〜R-F13（3項目） | M+M+S ≒ 2.5日 |
| Phase 5（テストカバレッジ） | R-F14〜R-F18（5項目） | M+M+S+S+M ≒ 3.5日 |
| Phase 6（ファイルサイズ） | R-F19（13ファイル、独立コミット） | L ≒ 3-4日（分散可能） |
| Phase 7（liff/line-reserve） | R-F20〜R-F25（6項目） | M+M+L+L+M+S ≒ 5日 |

**推奨着手順**: Phase 1（低リスク機械的作業で足場を整える）→ Phase 2（型安全性・検出基盤）→ Phase 3（パフォーマンス、視覚変化なし）→ Phase 5の CRITICAL項目（R-F14を先行、回帰リスクが最も高いため）→ Phase 4（アクセシビリティ、UI変更を伴うため慎重に）→ Phase 5残り→ Phase 6（体力に応じて）→ Phase 7（liff/line-reserve、R-F23のみ方針決定を待つ）。

**完了条件**:
- feature間直接import違反（`@/features/`への自feature外参照）が0件
- `*Model.ts`等のPascalCase非コンポーネントファイル命名違反が0件、settings/lstep/aggregationが標準構成（api/components/hooks/routes/types）に揃っている
- `src/hooks/`配下の全フックが`frontend/src/hooks/CLAUDE.md`のフック一覧表に記載され、参照実態と配置が一致している
- `design-system-audit.mjs`のC3正規表現盲点（10進rgba）が0件
- `@typescript-eslint/no-explicit-any`が`error`（またはlintスクリプトが`--max-warnings=0`）でCIゲート化されている
- knipがCIで実行され、少なくとも非blocking可視化として稼働している
- PropertyRow経由の22ファイルでlabel-input関連付けが機能している
- 受付ボード（AppointmentCard）を含む主要な疑似ボタンがキーボード操作可能
- vaccinations次回接種日計算・master共有CRUD状態機械・薬剤マスタ純粋関数にユニットテストが存在する
- line-reserveの予約作成フローがNULLバイト対策済みaxiosインスタンスを使用し、API失敗時に再試行導線を持つ
- liff/line-reserveのuse-liff.tsが単一実装に統合されている
