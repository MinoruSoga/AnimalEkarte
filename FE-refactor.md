# FE refactor — `DESIGN.md` ページ準拠チェックリスト

> 作成日: 2026-07-21
> 対象: `frontend/src` の本体アプリ 84 リーフルート（83 ページ + 404 fallback）。
> 目的: 自動監査の PASS と、実際のデザイン準拠を分けて追跡する。

## 1. 判定方針

デザインの正本は軸ごとに異なる。

- 色: [`docs/spec/design-system.md`](docs/spec/design-system.md)
- タイポグラフィ、形状、余白、エレベーション、コンポーネント寸法: [`DESIGN.md`](DESIGN.md)
- 実行時トークン: `frontend/src/lib/design-tokens.ts`
- 既存のページ台帳: [`docs/spec/ui-design-compliance.md`](docs/spec/ui-design-compliance.md)
- React実装・性能レビュー: `/vercel-react-best-practices`（Vercel Engineeringの64ルール。8カテゴリをV軸として確認）

判定記号:

- `[x]`: P1〜P10 と、該当するV1〜V8の目視・意味レビューまで完了し、既知の未準拠がない
- `[ ]`: 自動監査は通るが、既知の未準拠または未完了の目視確認がある
- `M✓`: 現行 tree 全体で `docker compose exec frontend pnpm design-audit` の C1/C3/C5/C6b/C7〜C19 は PASS。ページ単位の合格を意味しない
- `—`: 製品仕様上の対象外、または描画ページではない

現時点の結論は、**現行 tree の C1〜C19 自動監査は PASS、完全準拠の確定は 0/83**。R1〜R17 の共通基盤・検出済みhotspotは 2026-07-22 までに修正済みだが、83ページすべての P1〜P10目視確認とV1〜V8レビュー、とくに安全に表示できない作成route、実データ/RBAC依存route、臨床状態は未完了である。`ui-design-compliance.md` の旧「準拠83」は機械適合と完全準拠を分けて読み替える。

## 2. 全ページ共通チェック（P 軸）

各ページは、該当する全項目を確認してから `[x]` にする。

- [ ] P1 図地: ページは canvas-soft、カード・パネル・入力は white surface、境界は 1px hairline
- [ ] P2 色用途: brand は CTA・リンク・active/focus のみ。status/semantic 色を構造や装飾へ流用しない
- [ ] P3 タイポ: heading/title/body/caption/eyebrow の役割、size、weight、line-height、letter-spacing が一致
- [ ] P4 余白: 4/8/12/16/24/28/32px のスケールに従い、20px 等の仕様外段を使わない
- [ ] P5 形状: Primary CTA は pill、utility は 8px、入力は 4px、カードは 12px、modal は 16px
- [ ] P6 Elevation: 通常カードは Level 0、浮動面は Level 1、modal/toast は Level 2
- [ ] P7 Table/Form: table header は canvas-soft + eyebrow、body は body-sm、セルは 12px 16px。入力は white/body-sm/4px
- [ ] P8 Responsive: 1440px、1080〜1300px、768〜840px、600px 以下で崩れず、mobile は単一カラム
- [ ] P9 Touch/状態: 操作対象は 44×44px 以上。focus/pressed/disabled/RBAC 非活性が識別可能
- [ ] P10 臨床安全: danger、死亡、異常、期限、権限不足の表示を退行させない

### 2.1 React実装・性能チェック（V軸）

変更したページと、そのページが直接利用するcomponent/hookに `/vercel-react-best-practices` を適用する。該当ルールはrule ID（例: `async-parallel`）と確認根拠をページ行または検証記録へ残す。Next.js/RSC/server action固有ルールは、Vite SPAである本体アプリには機械的に適用せず `—` とする。Vercelルールとproject規約が競合する場合はproject規約を優先し、例外理由を記録する（例: feature外からのimportはproject必須のfeature indexを使う）。

- [ ] V1 Waterfall: 0適合 / 3一部適合 / 2対象外。独立処理の直列 `await` が残る（`async-*`）
- [ ] V2 Bundle: 2適合 / 2一部適合 / 1対象外。第三者barrelと遷移前preloadが残る（`bundle-*`）
- [x] V3 Server: 8/8対象外を確認済み。本体はVite CSRで、RSC/server action/SSR surfaceを持たない（`server-*`）
- [ ] V4 Client fetch: 1適合 / 2一部適合 / 1対象外。二重取得候補とlistener重複が残る（`client-*`）
- [ ] V5 Re-render: 3適合 / 9一部適合 / 3実測未確認。effect同期、依存、state更新等に残件がある（`rerender-*`）
- [ ] V6 Rendering: 3適合 / 3一部適合 / 4対象外 / 1実測未確認。長一覧とshow/hide surfaceに残件がある（`rendering-*`）
- [ ] V7 JavaScript: 4適合 / 8一部適合 / 1対象外。反復lookup、複数走査、storage read等に残件がある（`js-*`）
- [x] V8 Advanced: 2適合 / 1対象外。具体的なhotspotがある場合だけ適用する（`advanced-*`）

### 2.2 `/vercel-react-best-practices` 64ルール実査結果

> 監査日: 2026-07-22
> 基準: ローカルskill `vercel-react-best-practices` v1.0.0 の Quick Reference 64ルール。
> 範囲: 現行working treeの `frontend/src` production code、`frontend/index.html`、`frontend/package.json`、route/provider/query設定。test、generated type、`shared-liff`、LIFF/LINE予約別アプリは除外。

判定は、`✅ 適合`＝対象候補を静的に意味レビューして既知の違反なし、`⚠ 一部適合`＝準拠例と未解決候補が混在、`— 対象外`＝該当surfaceなし、`🔎 実測未確認`＝Network/React Profiler等が必要、とする。静的 `✅` はbundle実測や全ページのruntime性能を保証しない。

| V軸 | 適合 | 一部適合 | 対象外 | 実測未確認 | カテゴリ判定 |
|---|---:|---:|---:|---:|---|
| V1 Waterfall | 0 | 3 | 2 | 0 | ⚠ 一部適合 |
| V2 Bundle | 2 | 2 | 1 | 0 | ⚠ 一部適合 |
| V3 Server | 0 | 0 | 8 | 0 | — 8/8対象外確認済み |
| V4 Client fetch | 1 | 2 | 1 | 0 | ⚠ 一部適合 |
| V5 Re-render | 3 | 9 | 0 | 3 | ⚠ 一部適合 |
| V6 Rendering | 3 | 3 | 4 | 1 | ⚠ 一部適合 |
| V7 JavaScript | 4 | 8 | 1 | 0 | ⚠ 一部適合 |
| V8 Advanced | 2 | 0 | 1 | 0 | ✅ 適合 |
| **合計** | **15** | **27** | **18** | **4** | **64/64 判定記録済み** |

#### V1 Waterfall（5ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| [ ] | `async-defer-await` | ⚠ 一部適合 | `features/owners/hooks/use-owner-form.ts:195-227` は結果を使わないcache invalidationを待ってからoptional pet処理へ進む |
| [ ] | `async-parallel` | ⚠ 一部適合 | `app/routes/clinical-general-routes.tsx:45-50` は `Promise.all`。一方 `features/accounting/hooks/use-accounting-completion-action.ts:95-124` は独立明細を `for` + `await` で直列送信 |
| [ ] | `async-dependencies` | ⚠ 一部適合 | owner作成後のinvalidationと複数pet作成に部分依存があるが、`features/owners/hooks/use-owner-form.ts:195-227` では直列境界が残る |
| — | `async-api-routes` | — 対象外 | `package.json:7-16` と `app/router.tsx:1-9` のVite SPAにfrontend API route/server handlerはない |
| — | `async-suspense-boundaries` | — 対象外 | SSR streaming/RSCなし。client lazy境界はV2で確認する |

#### V2 Bundle（5ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| [ ] | `bundle-barrel-imports` | ⚠ 一部適合 | feature indexはproject規約上の承認済み例外。第三者 `lucide-react` root barrelはproduction 205ファイルに残り、直接import例は `features/reception/routes/Reception.tsx:7` |
| [x] | `bundle-dynamic-imports` | ✅ 適合 | route `lazy` 82件、`React.lazy` 25件。`app/routes/app-routes.tsx:13-50` とrechartsを遅延する `features/medical-records/components/VitalsTab/VitalsTab.tsx:18-22` を確認 |
| — | `bundle-defer-third-party` | — 対象外 | analytics/logging SDKのdependency・初期化なし。UI通知のsonnerは対象外 |
| [x] | `bundle-conditional` | ✅ 適合 | recharts graphと予約のMonth/Week viewを選択時だけimport/renderする |
| [ ] | `bundle-preload` | ⚠ 一部適合 | lazy moduleのhover/focus preloadは0件。Sidebar linkは遷移開始後にchunkを読む（`components/shared/Layout/SidebarItems.tsx:100-107`） |

#### V3 Server（8ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| — | `server-auth-actions` | — 対象外 | server actionなし |
| — | `server-cache-react` | — 対象外 | RSC/`React.cache()`なし |
| — | `server-cache-lru` | — 対象外 | frontend server processなし |
| — | `server-dedup-props` | — 対象外 | RSC propsなし |
| — | `server-hoist-static-io` | — 対象外 | server-side static I/Oなし |
| — | `server-serialization` | — 対象外 | server/client component境界なし |
| — | `server-parallel-fetching` | — 対象外 | server component fetchなし。client fetchはV1/V4で確認 |
| — | `server-after-nonblocking` | — 対象外 | Next.js `after()`/server request lifecycleなし |

#### V4 Client fetch（4ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| [ ] | `client-swr-dedup` | ⚠ 一部適合 | TanStack Queryの全体Providerとstale/gc設定はあり（`lib/react-query.ts:14-32`）。owner loader後に同じownerをqueryする二重GET候補がある（`features/owners/loaders.ts:169-175`、`hooks/use-owner.ts:11-21`） |
| [ ] | `client-event-listeners` | ⚠ 一部適合 | 7 listenerすべてcleanupあり。一方 `use-reduced-motion.ts:7-18` は呼出単位でlistenerを作り、予約card一覧から件数分登録される |
| — | `client-passive-event-listeners` | — 対象外 | scroll/wheel/touch系native listenerなし。既存はbeforeunload/storage/beforeprint/MediaQuery change |
| [x] | `client-localstorage-schema` | ✅ 適合 | 保存はversion付きkey `auth_current_clinic:v1` のscalar 1件のみ（`lib/current-clinic.ts:1-14`）。session/authはhttpOnly Cookie |

#### V5 Re-render（15ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| [ ] | `rerender-defer-reads` | 🔎 実測未確認 | callbackだけが必要とするcontext/state購読は静的候補抽出だけでは確定できず、Profiler確認が必要 |
| [ ] | `rerender-memo` | 🔎 実測未確認 | heavy childへの `memo` は多数あるが、残る再描画costの有無はProfiler未実施 |
| [ ] | `rerender-memo-with-default-value` | ⚠ 一部適合 | `MedicalRecordPrintView.tsx:40-50` の `treatments = []` がmemoized componentへ毎回新規配列を作る |
| [ ] | `rerender-dependencies` | ⚠ 一部適合 | primitive depsの準拠例はあるが、ClinicHoliday modalのeffectは `existing` object全体へ依存する |
| [ ] | `rerender-derived-state` | ⚠ 一部適合 | `hooks/use-permission.ts:20-27` は必要なbooleanだけでなくAuth context更新全体の影響を受ける |
| [ ] | `rerender-derived-state-no-effect` | ⚠ 一部適合 | `use-accounting-detail-state.ts:135-140` 等にAPI/propsからlocal stateへのeffect同期が残る。会計箇所はaction closure整合性の意図的例外をコメント済み |
| [ ] | `rerender-functional-setstate` | ⚠ 一部適合 | functional更新の準拠例は多数あるが、`UnpaidTab.tsx:347-348` はclosureの `page` から次値を作る |
| [x] | `rerender-lazy-state-init` | ✅ 適合 | `ShiftCalendarPage.tsx:38`、`use-hospitalization-form.ts:39` 等の計算初期値はfunction initializerを使用 |
| [ ] | `rerender-simple-expression-in-memo` | ⚠ 一部適合 | `DailyDateNav.tsx:29-35` の単純な文字列比較boolean等が `useMemo` に残る |
| [ ] | `rerender-split-combined-hooks` | ⚠ 一部適合 | `features/auth/hooks/use-auth.tsx:64-215` はsession、storage、permission、mutation責務を同一Provider hookで購読する |
| [ ] | `rerender-move-effect-to-event` | ⚠ 一部適合 | `features/checkups/hooks/use-checkup-form.ts:114-119` 等でaction成功後の遷移をeffectで監視する |
| [ ] | `rerender-transitions` | 🔎 実測未確認 | `useTransition` 採用箇所はあるが、未採用interactionの応答性はProfiler未実施 |
| [ ] | `rerender-use-deferred-value` | ⚠ 一部適合 | 複数一覧は採用済み。一方manual検索は入力値を直接Fuse検索へ渡す（`use-manual-search.ts:24-31`） |
| [x] | `rerender-use-ref-transient-values` | ✅ 適合 | dirtyの最新値をrefで保持しstable callbackから読む（`hooks/use-side-peek-dirty.ts:23-56`） |
| [x] | `rerender-no-inline-components` | ✅ 適合 | production TSXのcomponent宣言を走査し、component内component定義の既知候補なし |

#### V6 Rendering（11ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| — | `rendering-animate-svg-wrapper` | — 対象外 | 自前SVG animationなし |
| [ ] | `rendering-content-visibility` | ⚠ 一部適合 | `content-visibility` 使用0件。`DataTable.tsx:71` 等の長一覧は実データ量に応じた適用余地がある |
| [ ] | `rendering-hoist-jsx` | ⚠ 一部適合 | module-level option/route JSXの準拠例はあるが、全static subtreeのhoist要否は未解消 |
| — | `rendering-svg-precision` | — 対象外 | 自前inline SVGなし。lucide vendor iconは監査対象外 |
| — | `rendering-hydration-no-flicker` | — 対象外 | `createRoot` のCSRでhydrateしない（`main.tsx:1-27`） |
| — | `rendering-hydration-suppress-warning` | — 対象外 | hydration mismatch surfaceなし |
| [ ] | `rendering-activity` | ⚠ 一部適合 | React 19だが `Activity` 使用0件。manual edit/viewやtabのshow/hideでstate保持・描画costを実測して採否を決める |
| [x] | `rendering-conditional-render` | ✅ 適合 | JSX conditional候補はternary/明示null。project規約違反の `condition && <JSX>` は既知0件 |
| [ ] | `rendering-usetransition-loading` | 🔎 実測未確認 | transition採用箇所はあるが、個別loading stateとの優先度比較はinteraction計測未実施 |
| [x] | `rendering-resource-hints` | ✅ 適合 | Google Fontsへpreconnectを設定（`frontend/index.html:12-15`） |
| [x] | `rendering-script-defer-async` | ✅ 適合 | entry scriptはdefer相当の `type="module"`（`frontend/index.html:19`） |

#### V7 JavaScript（13ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| [x] | `js-batch-dom-css` | ✅ 適合 | 直接style連打はなく、print modeはroot/bodyのclass切替へ集約（`ManualPage.tsx:124-131`） |
| [ ] | `js-index-maps` | ⚠ 一部適合 | `trimming-form-utils.ts:58-62` はselected IDごとに `items.find` するためMap化余地あり |
| [x] | `js-cache-property-access` | ✅ 適合 | hot loop候補の意味レビューで、同一deep propertyを反復取得する既知候補なし |
| [ ] | `js-cache-function-results` | ⚠ 一部適合 | `TreatmentSearchDialog.tsx:107-114` 等で検索ごとに全itemの `normalizeKana` を再計算する |
| [ ] | `js-cache-storage` | ⚠ 一部適合 | axios interceptorがrequestごとに `getStoredClinicId()` でlocalStorageを読む（`lib/axios.ts:45-53`）。tenant切替の正しさを保つcache invalidation設計が必要 |
| [ ] | `js-combine-iterations` | ⚠ 一部適合 | `hooks/use-staff-validation.ts:11-16` 等に連続 `filter().map()` が残る |
| — | `js-length-check-first` | — 対象外 | 大配列の高コストな等価比較surfaceを検出せず |
| [x] | `js-early-exit` | ✅ 適合 | validation/filter/handlerでearly returnを使用し、深い分岐の既知hotspotなし |
| [x] | `js-hoist-regexp` | ✅ 適合 | `new RegExp` は `lib/sanitize.ts:7`、`get-exam-type-fields.ts:27-37` のmodule level。loop内生成なし |
| [ ] | `js-min-max-loop` | ⚠ 一部適合 | `owner-report/lib/report-summary.ts:21-49` はmin/max相当の1件取得にfilter+sort、医療recordでは `Math.max(...map())` が残る |
| [ ] | `js-set-map-lookups` | ⚠ 一部適合 | `staff-settings-model.ts:33-35` のloop内 `groupIds.includes` 等はSet化余地あり |
| [ ] | `js-tosorted-immutable` | ⚠ 一部適合 | input copy後のsortで破壊的変更は避ける例が多いが、`toSorted()` への統一は未完了 |
| [ ] | `js-flatmap-filter` | ⚠ 一部適合 | `owner-report/api/get-pet-examinations.ts:29` 等に `map().filter()` が残る |

#### V8 Advanced（3ルール）

| チェック | Rule ID | 判定 | 根拠・残件 |
|---|---|---|---|
| — | `advanced-event-handler-refs` | — 対象外 | handler freshnessが原因の高頻度再購読hotspotを検出せず。必要になるまで導入しない |
| [x] | `advanced-init-once` | ✅ 適合 | QueryClientはmoduleで1回生成し、session restore promiseもmount内でrefに保持（`features/auth/hooks/use-auth.tsx:68-78`） |
| [x] | `advanced-use-latest` | ✅ 適合 | `use-side-peek-dirty.ts:23-56` はlatest refをstable `confirmDiscard` callbackから読む |

## 3. 共通ギャップの対応結果

| ID | 判定 | 内容 | 主な影響 |
|---|---|---|---|
| R1 | ✅ | `PageLayout` / `STYLE.pageContent` を `py-6`（24px）へ統一。C16 が spacing `*-5` の再導入を禁止 | `PageLayout` を直接使うページと、共通 shell 経由ページ |
| R2 | ✅ | `FormHeader` の `h1` を title role（20px/600/1.4/-0.125px）へ変更し、`PageLayout.description` も接続 | 全 `PageLayout`、受付、予約管理、LINE予約枠 |
| R3 | ✅/⚠ | table primitive / `STYLE.table*` と呼び出し側36ファイル195行を header eyebrow、body-sm、cell 12px 16pxへ統一し、C18で再導入を禁止。手書き `<th>/<td>` はページ再監査で継続 | 共通 table、各一覧・フォーム内明細 |
| R4 | ✅ | `/lstep/checkup-sync` を `PageLayout` 化し、404 / route error fallback に canvas-soft を付与 | 健診連携、404、route error |
| R5 | ✅ | `/accounting/close` の「プレビュー」を pill + 16px horizontal paddingへ変更 | レジ締め |
| R6 | ✅ | `/accounting/close/history` の手書き table を共通 header/body typography・paddingへ変更 | レジ締め履歴 |
| R7 | ✅ | route 表面の直接 named color 11件を `C` tokenへ移行。C15 が routes/pages の再導入を禁止 | 受付、トリミング、Lステップ健診/配信、マニュアル |
| R8 | ✅/⚠ | 指摘5画面を mobile single-column / full-width に変更し unit test を追加。全83ページの4 viewport目視は未完了 | 受付、LINE予約枠、Lステップ配信、見積作成/詳細 |
| R9 | ⚠ | C8 を相対パス完全一致の page 9件/helper 5件へ分離し、drop-shadow・named color・20px spacing・CSS shadow・Table overrideを C10/C15〜C18 へ追加。意味的誤用・手書きtable・全画面 responsive・P10 は引き続き review/browser が必要 | 監査スクリプト全体 |
| R10 | ✅/⚠ | 検出済みの固定複数列formを mobile single-columnへ変更し、飼主一覧・予約toolbar・設定tableは狭幅で内部scroll/wrapするようTDD。実データ依存routeの再監査は継続 | 在庫、入院、診療、予防、飼主、予約、設定form |
| R11 | ✅/⚠ | 共通 Button/Tabs/Switch/icon action/DatePicker/CalendarNav/DeleteIconと個別raw control・認証導線を44px以上へ変更。見た目サイズを保つ場合もfocusable要素のhit areaを確保。全route再監査は継続 | 全Button/Switch利用画面、日付選択、設定、認証 |
| R12 | ✅/⚠ | 定期健診formは実APIフローに必要な `medical-records:create && edit` でfieldset/save/actionをfail-closedにし、実在する過去の次回来院日だけ非色依存の「期限切れ」を追加。仮の既定日付は `-` へ変更。死亡表示など他の実データ状態はbrowser確認を継続 | 定期健診、ワクチン、飼主/ペット状態 |
| R13 | ✅/⚠ | browser実測と独立reviewで検出した44px未満、focus/label、重複accessible name、modal/drawer description・focus・狭幅overflow、選択状態をTDD修正。Sidebar、SortableHeader、DatePicker、checkbox、予約/受付card、シフト、manual、LSTEPタグ/LINE送信drawer、診療formを含む。最終treeの全route再監査はfixture依存領域を継続 | 全共通shell、一覧table、受付/予約、シフト、manual、診療・検査 |
| R14 | ✅/⚠ | route/APIのRBACを実backend契約へ統一。LSTEPタグ/配信は `lstep-analytics:view`、manual override/editは `manual-edit` でAPI/UIを抑止。forgot/resetはsession復元と401 redirectの対象外にし、login/logout後の認証snapshotも再取得する。browser最終再確認は継続 | LSTEP、manual、認証回復route |
| R15 | ✅/⚠ | 会計・設定のread-only監査で検出した40px操作、休診日/特別期間の重複削除名、休診曜日checkbox、AccessDeniedの非h1をTDD修正。ガイドのadmin fixtureが実効「一般」のため権限内本体のbrowser確認は継続 | レジ締め、月次レポート、締め時間設定、全RBAC拒否画面 |
| R16 | ✅/⚠ | 臨床5一覧のmouse-only row遷移を44px native cell linkへ置換し、detail/action名へstable ID、検索clearへ44px hit areaと入力余白、無効staffへ非色依存説明を追加。カルテ会計導線は `accounting:view` と同一医院を要求し、健診CTA/form/子routeは `medical-records:create && edit`、編集導線は `view && edit` へ統一。共有row clickはmultiline再集計で29箇所と判明しR17で対応 | カルテ、トリミング、検査、予防接種、定期健診、共通table/filter |
| R17 | ✅/⚠ | `DataTableRow` / `SortableDataTableRow` の行全体click 29箇所（26 production files）を廃止。5一覧の識別cellを44px native link、side panel導線を固有名native buttonへ変更し、行操作名へstable IDを追加。sortable 10行は44px native drag handleへattributes/listenersを集約し、edit権限なしではhandleとmutation callbackの両方をfail-closed化。基底APIの型・runtimeとC19で再導入を禁止。raw `TableRow onClick` 7箇所・`tr onClick` 3箇所は次バッチ | 会計・在庫・見積・飼主・入院一覧、医院/シフト/各マスタ、共有table/DnD |

補足:

- `p-5` / `m-5` / `gap-5` 等の20px spacingは C16 導入時に31件を検出し、画面用26件を仕様内スケールへ移行した。印刷ビュー5件は画面用 DESIGN.md の対象外として明示除外した。
- `design-audit` の PASS は有効な回帰ガードだが、P1〜P10 とV1〜V8の完全準拠を意味しない。
- `drop-shadow`、CSS の `box-shadow` / `filter: drop-shadow()`、route/page の Tailwind named color は機械化済み。正しい typography role の意味選択と全画面 responsive は別チェックが必要。

## 4. ページ別チェックリスト（本体84リーフルート）

現行 tree 全体の自動監査結果は `M✓` だが、ページ単位の合格判定ではない。行中の R1〜R17 は修正追跡IDで、2026-07-22 時点で実装済み。`[ ]` は P1〜P10、該当するV1〜V8、とくに4 viewport、console、実データ/RBAC状態の再確認が必要であることを示す。

### 認証・共通

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [ ] | ログイン | `/login` | `Login` | M✓、R11/R14、P8/P9は4 viewport OK。未認証session復元の401 resourceを含むconsole厳密判定は保留 |
| [ ] | パスワードを忘れた方 | `/forgot-password` | `ForgotPasswordPage` | M✓、R11/R14、修正後4 viewportで直接到達、`/me` / refresh / login redirectなし |
| [ ] | パスワード再設定 | `/reset-password` | `ResetPasswordPage` | M✓、R11/R14、token無しと検診tokenが4 viewportで直接到達、`/me` / refresh / login redirectなし |
| [ ] | 飼主カルテレポート | `/owners/:id/report` | `OwnerReport` | M✓、R13/P10、4 viewportでstandalone・6パネル・死亡表示・44pxペット切替・`include_deceased=true`・overflow/console/non-GET 0を確認。printは未確認 |
| — | 404 Not Found | `*` | inline fallback | canvas-soft 修正・unit test済み。製品ページ数には含めない |

### 会計・在庫・見積・シフト

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [ ] | 会計一覧 | `/accounting` | `AccountingList` | M✓、R1/R2/R3/R13/R17。日時cellを44px native detail link化し、カルテ/編集操作を44px・stable ID付きへ変更。R17後browser再確認は下記検証記録 |
| [ ] | 会計ペット選択 | `/accounting/select-pet` | `AccountingPetSelection` | M✓、R1/R2/R13、P8/P9は4 viewport OK |
| [ ] | 会計登録 | `/accounting/new` | `AccountingDetail` | M✓、R1/R2/R3 |
| [ ] | 会計詳細 | `/accounting/:id` | `AccountingDetail` | M✓、R1/R2/R3。fixture `1024128` は4 viewportとも、無名32px数値入力・32px割引操作、完了/支払済みのread-only表示なし、レジ締めGET 403でP7/P9/P10 NG |
| [ ] | レジ締め | `/accounting/close` | `CashRegisterClosePage` | M✓、R1/R2/R3/R5/R15、印刷操作を44px化。mountは無API・書込なしを静的確認、権限内本体browserはfixture待ち |
| [ ] | レジ締め履歴 | `/accounting/close/history` | `CashRegisterHistoryPage` | M✓、R1/R2/R3/R6/R15、mountはGET-only。拒否shellを4 viewport確認、h1修正後は代表routeで再確認。権限内本体はfixture待ち |
| [ ] | 月次集計レポート | `/accounting/reports` | `AccountingReportsPage` | M✓、R1/R2/R3/R15、操作/集計条件/税率linkを44px化。AccessDenied h1は修正後4 viewport OK、権限内本体はfixture待ち |
| [ ] | 在庫一覧 | `/inventory` | `InventoryList` | M✓、R1/R2/R3/R13/R17。edit権限時だけ品名cellを44px native detail link化し、非link cellの行遷移を廃止 |
| [ ] | 在庫登録 | `/inventory/new` | `InventoryForm` | M✓、R1/R2 |
| [ ] | 在庫編集 | `/inventory/:id` | `InventoryForm` | M✓、R1/R2。fixture `7` は4 viewportとも、数量0/最低50の不足表示なし、transparent入力、`#ddd` card境界、想定一般roleでも編集可能でP1/P7/P9/P10 NG |
| [ ] | 見積一覧 | `/estimates` | `EstimateList` | M✓、R1/R2/R3/R13/R17。locked/view-onlyを維持した見積Noの44px native detail linkと固有操作名へ変更 |
| [ ] | 見積作成 | `/estimates/new` | `EstimateForm` | M✓、R1/R2/R8 |
| [ ] | 見積詳細 | `/estimates/:id` | `EstimateDetail` | M✓、R1/R2/R3/R8。fixture `1` は4 viewportとも期限 `2026-06-22` が経過済みだが「期限切れ」等の非色依存表示がなくP10 NG |
| [ ] | 見積編集 | `/estimates/:id/edit` | `EstimateForm` | M✓、R1/R2/R8 |
| [ ] | シフトカレンダー | `/shifts` | `ShiftCalendarPage` | M✓、R1/R2/R13、407操作を4 viewportで44px以上・document overflow 0確認 |

### 診療・入院・トリミング・予防

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [ ] | カルテ一覧 | `/medical-records` | `MedicalRecords` | M✓、R1/R2/R3/R16。44px native detail link・固有操作名・戻るstate・無効医説明を追加。会計は44pxかつ `accounting:view`・同一医院のときだけ表示。4 viewportで20行非interactive、表示操作44px/ID付き、overflow/console/non-GET 0を確認 |
| [ ] | カルテペット選択 | `/medical-records/select-pet` | `MedicalRecordPetSelection` | M✓、R1/R2 |
| [ ] | カルテ作成 | `/medical-records/new` | `MedicalRecordForm` | M✓、R1/R2/P10 |
| [ ] | カルテ編集 | `/medical-records/:id` | `MedicalRecordForm` | M✓、R1/R2/R13/P10、死亡/確定lockを確認。修正後4 viewportで確定済み保存disabled・44px・overflow/console/non-GET 0を再確認 |
| [ ] | 入院・ホテル一覧 | `/hospitalization` | `HospitalizationList` | M✓、R1/R2/R3/R13/R17/P10。生存行の入院Noを44px native detail link化し、死亡行のlock/plain textを維持。空slot 19件は一意名・44px確認済み |
| [ ] | 入院・ホテル ペット選択 | `/hospitalization/select-pet` | `HospitalizationPetSelection` | M✓、R1/R2/P10 |
| [ ] | 入院・ホテル登録 | `/hospitalization/new` | `HospitalizationForm` | M✓、R1/R2/P10 |
| [ ] | 入院・ホテル詳細 | `/hospitalization/:id` | `HospitalizationDetail` | M✓、R1/R2/P10 |
| [ ] | 入院・ホテル編集 | `/hospitalization/:id/edit` | `HospitalizationForm` | M✓、R1/R2/P10 |
| [ ] | トリミング一覧 | `/trimming` | `TrimmingList` | M✓、R1/R2/R3/R16。44px native detail link・固有操作名・戻るstate・無効staff説明、pet番号body-smをTDD修正。browserはempty stateのため4/4 Partial |
| [ ] | トリミング ペット選択 | `/trimming/select-pet` | `TrimmingPetSelection` | M✓、R1/R2 |
| [ ] | トリミング登録 | `/trimming/new` | `TrimmingForm` | M✓、R1/R2/R7 |
| [ ] | トリミング編集 | `/trimming/:id` | `TrimmingForm` | M✓、R1/R2/R7 |
| [ ] | 検査一覧 | `/examinations` | `ExaminationsList` | M✓、R1/R2/R3/R16。mouse-only row遷移を44px native detail linkへ置換し、detail/action名をID付きで一意化。4 viewportで20行非interactive、操作44px/ID付き、overflow/console/non-GET 0を確認 |
| [ ] | 検査ペット選択 | `/examinations/select-pet` | `ExaminationPetSelection` | M✓、R1/R2 |
| [ ] | 検査登録 | `/examinations/new` | `ExaminationForm` | M✓、R1/R2 |
| [ ] | 検査編集 | `/examinations/:id` | `ExaminationForm` | M✓、R1/R2/R13、空項目名・履歴期間・form labelをTDD修正し、安全GETの4 viewportで44px・overflow/console/non-GET 0を確認 |
| [ ] | ワクチン一覧 | `/vaccinations` | `VaccinationList` | M✓、R1/R2/R3/R16。mouse-only row遷移を44px native detail linkへ置換し、detail/action名をID付きで一意化。browserはempty stateのため4/4 Partial |
| [ ] | ワクチン ペット選択 | `/vaccinations/select-pet` | `VaccinationPetSelection` | M✓、R1/R2 |
| [ ] | ワクチン登録 | `/vaccinations/new` | `VaccinationForm` | M✓、R1/R2/P10 |
| [ ] | ワクチン編集 | `/vaccinations/:id` | `VaccinationForm` | M✓、R1/R2/P10 |
| [ ] | 定期健診一覧 | `/checkups` | `CheckupsList` | M✓、R1/R2/R3/R16。カルテdetail/actionを44px・ID付きにし、新規CTAは `medical-records:create && edit`、編集は `view && edit` へ整合。browserはempty stateのため4/4 Partial、CTA 122×44を確認 |
| [ ] | 定期健診ペット選択 | `/checkups/select-pet` | `CheckupPetSelection` | M✓、R1/R2/R16。直接URLも `medical-records:create && edit` の二重guardでfail-closed |
| [ ] | 定期健診登録 | `/checkups/new` | `CheckupForm` | M✓、R1/R2/R12/R16/P10。route・fieldset・保存・form actionを `medical-records:create && edit` へ統一し部分保存を抑止 |

### 受付・飼主・予約・集計

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [ ] | 受付 | `/` | `Reception` | M✓、R2/R7/R8/R13/P10、shellは4 viewportで44px・overflow/console/non-GET 0。現fixtureは予約0件のため5状態/死亡表示は最終未確認 |
| [ ] | 飼主一覧 | `/owners` | `OwnersList` | M✓、R1/R2/R3/R17。同一医院かつedit権限時だけ飼主名を44px native detail link化し、別医院の閲覧専用境界を維持 |
| [ ] | 飼主登録 | `/owners/new` | `OwnerForm` | M✓、R1/R2/R3/P10 |
| [ ] | 飼主編集 | `/owners/:id` | `OwnerForm` | M✓、R1/R2/R3/R13/P10、LINE送信drawerはmobile全幅/最大480px・description・44px close・title領域分離をTDD修正 |
| [ ] | 集計ダッシュボード | `/aggregation` | `AggregationDashboardPage` | M✓、R1/R2/R3 |
| [ ] | 予約管理 | `/reservations` | `ReservationManagement` | M✓、R2/P10 |

### Lステップ・LINE予約・マニュアル

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [ ] | Lステップ健診連携 | `/lstep/checkup-sync` | `CheckupSyncPage` | M✓、R3/R4/R7 |
| [ ] | Lステップ配信モニター | `/lstep/delivery-monitor` | `LstepDeliveryMonitorPage` | M✓、R1/R2/R3/R7/R8/R13/R14、403 loading固定をroute guardで修正し、AccessDenied・領域API 0を4 viewportで確認 |
| [ ] | Lステップ分析 | `/lstep/analytics` | `LstepAnalyticsPage` | M✓、R1/R2/R3 |
| [ ] | LINE予約設定 index | `/line-reservation` | `LineReservationSettings` | M✓、R1/R2 |
| [ ] | LINE予約設定 | `/line-reservation/settings` | `LineReservationSettings` | M✓、R1/R2 |
| [ ] | LINE予約ページエディタ | `/line-reservation/page-editor` | `LineReservationPageEditor` | M✓、R1/R2 |
| [ ] | LINE予約枠設定 | `/line-reservation/slots` | `LineReservationSlotsSettings` | M✓、R2/R8/R13、390/500/1440pxで日付button 44px・calendar内部scroll・document overflow 0をPlaywright実測 |
| [ ] | 医院マスタ設定 | `/settings/clinic` | `ClinicMasterSettings` | M✓、R1/R2/R3/R17。行全体clickを廃止し編集操作をstable ID付きへ変更 |
| [ ] | マニュアルトップ | `/manual` | `ManualPage` | M✓、shellはP軸対象、本文は対象外、R7/R13/R14。編集権限なしのoverride API 0・編集UI非表示・mobile modal focus/Escape・500→800px cleanupを確認 |
| [ ] | マニュアル記事 | `/manual/:category/:slug` | `ManualPage` | M✓、shellはP軸対象、本文は対象外、R7/R13/R14。編集権限なしのoverride API 0・編集UI非表示・modal focus/cleanupを確認 |

### 設定・マスタ

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [ ] | 設定トップ | `/settings` | `MasterSettingsIndex` | M✓、R1/R2 |
| [ ] | 職員マスタ | `/settings/staff` | `StaffSettings` | M✓、R1/R2/R3/R17/P10 |
| [ ] | 診療項目マスタ | `/settings/treatment-items` | `TreatmentPlanMaster` | M✓、R1/R2/R3/R17。root/childの詳細buttonと権限連動44px drag handleへ変更 |
| [ ] | 診断名マスタ | `/settings/diagnosis` | `DiagnosisSettings` | M✓、R1/R2/R3/R17。edit権限時だけ詳細button、権限なしではdrag無効 |
| [ ] | 動物種マスタ | `/settings/animal-species` | `AnimalSpeciesSettings` | M✓、R1/R2/R3/R17。read-only詳細導線を維持しdragをedit権限連動へ変更 |
| [ ] | トリミングマスタ | `/settings/trimming` | `TrimmingSettings` | M✓、R1/R2/R3/R17 |
| [ ] | トリミングコース種別 | `/settings/trimming-course-type` | `TrimmingCourseTypeSettings` | M✓、R1/R2/R3/R17 |
| [ ] | 薬剤マスタ | `/settings/medicine` | `MedicineSettings` | M✓、R1/R2/R3/R17/P10。edit権限時だけ詳細button、権限なしではdrag無効。sortable行の重複28px操作を44px操作1件へ統合 |
| [ ] | 予約種別マスタ | `/settings/reservation-type` | `ReservationTypeSettings` | M✓、R1/R2/R3/R17。read-only詳細導線を維持しdragをedit権限連動へ変更 |
| [ ] | 入院・ホテルマスタ | `/settings/hospitalization` | `HospitalizationSettings` | M✓、R1/R2/R3/R17 |
| [ ] | ケージマスタ | `/settings/cage` | `CageSettings` | M✓、R1/R2/R3/R17。read-only詳細導線を維持しdragをedit権限連動へ変更 |
| [ ] | 物販品マスタ | `/settings/merchandise-items` | `MerchandiseItemSettings` | M✓、R1/R2/R3/R17。read-only詳細導線を維持しdragをedit権限連動へ変更 |
| [ ] | 保険マスタ | `/settings/insurance` | `InsuranceSettings` | M✓、R1/R2/R3/R17 |
| [ ] | 職種マスタ | `/settings/occupations` | `OccupationSettings` | M✓、R1/R2/R3/R17 |
| [ ] | 権限グループマスタ | `/settings/permission-groups` | `PermissionGroupSettings` | M✓、R1/R2/R3/R17/P10。read-only詳細導線を維持しdragをedit権限連動へ変更 |
| [ ] | 問診テンプレート | `/settings/inquiry-templates` | `InterviewTemplateSettings` | M✓、R1/R2/R3/R17 |
| [ ] | 主訴マスタ | `/settings/interview/chief-complaint` | `ChiefComplaintSettings` | M✓、R1/R2/R3/R17 |
| [ ] | 問診テンプレート（interview） | `/settings/interview/templates` | `InterviewTemplateSettings` | M✓、R1/R2/R3/R17 |
| [ ] | シフトテンプレート | `/settings/shift-templates` | `ShiftTemplateSettings` | M✓、R1/R2/R3/R17。action-level create/edit/delete/reorder権限を補完し、read-only詳細と権限連動44px drag handleへ変更 |
| [ ] | 締め時間設定 | `/settings/closing-time` | `ClosingSettingsPage` | M✓、R1/R2/R15、追加/キャンセル/一意名削除と44px、生checkboxの44px focus targetをTDD修正。権限内本体はfixture待ち |
| [ ] | 支払方法マスタ | `/settings/payment-methods` | `PaymentMethodSettings` | M✓、R1/R2/R3/R15/R17、mountはGET-only。拒否shellを4 viewport確認、権限内本体はfixture待ち |
| [ ] | 割引キャンペーン | `/settings/campaigns` | `CampaignSettings` | M✓、R1/R2/R3/R17 |
| [ ] | Lステップ連携設定 | `/settings/integrations/lstep` | `LstepSettingsPage` | M✓、R1/R2 |
| [ ] | Lステップタグ管理 | `/settings/lstep/tags` | `LstepTagManagementPage` | M✓、R1/R2/R3/R13/R14/R15、誤empty/403を明示AccessDeniedへ修正・4 viewport確認。drawerのmobile全幅/最大480px・description・44px close/action/link・title余白をTDD修正。本体は権限fixture不足 |

## 5. 明示的な対象外・別枠

次を無言で「準拠」に含めない。

- [x] redirect-only 12 route は描画ページではないため対象外:
  `/settings/job-title`、`/settings/service-type`、`/settings/diagnosis-type`、`/settings/diagnosis-name`、`/settings/trimming-course`、`/settings/trimming-option`、`/settings/examination`、`/settings/vaccine`、`/settings/consultation`、`/settings/procedure`、`/settings/inquiry-template`、`/settings/shift-template`
- [x] 印刷帳票は紙・モノクロ出力のため画面用 canvas/ink/table 規範の対象外
- [x] `ManualContent` の Markdown 本文は文書レンダラーのため `ex-data-table-cell` の対象外。`ManualPage` shell は対象
- [x] `frontend/liff`、`frontend/line-reserve`、`frontend/src/shared-liff` は別アプリとして、現行製品仕様では本体 DESIGN.md 監査の対象外

別アプリの画面を本体84ページへ混ぜない代わりに、存在する surface を明記する。

- LIFF: `LiffLinkPage`、`LoadingPage`、`PetHealthPage`、共通 `ErrorPage`
- LINE予約ミニアプリ: `loading`、`error`、`maintenance`、`top`、`my-reservations`、`step1`〜`step8`、`step2b`、`step2c`（15 state）

## 6. 修正順と完了条件

- [x] F1 共通基盤: R1 `PageLayout`、R2 `FormHeader`、R3 table primitive/token と今回対象の上書きを修正
- [x] F2 個別違反: R4〜R8 を修正
- [x] F3a 監査強化: C8 path化、drop-shadow、spacing、route named color、CSS shadow、Table primitive override、共有DataTable row clickを C10/C15〜C19 で機械化
- [x] F3b 検証補完: table値・responsive hotspot は scoped unit test、全画面は browser checklist で追跡
- [ ] F4 ページ再監査: 83ページへ P1〜P10 と該当するV1〜V8を適用し、上のページ行を `[x]` に更新
- [x] F5 文書同期: `docs/spec/ui-design-compliance.md` のサマリーを「機械適合」と「完全確認」に分離
- [x] F6 検証: design-audit、監査unit test、変更範囲のVitest/ESLint、代表viewportのブラウザ確認を記録
- [x] F7a React性能baseline: `/vercel-react-best-practices` 64/64ルールを現行本体アプリへ静的適用し、適合/一部適合/対象外/実測未確認と根拠を2.2へ記録
- [ ] F7b React性能完了: 27一部適合を解消または理由付き例外化し、4実測未確認をNetwork/React Profilerで判定。変更ページ固有のrule IDを各ページ行へ追記

### 2026-07-21〜22 検証記録

- [x] `docker compose exec frontend pnpm design-audit` — C1〜C19、違反0件
- [x] `docker compose exec frontend node --test scripts/design-system-audit.test.mjs` — 52/52 PASS
- [x] 変更/追加済みfrontend testの scoped Vitest — 113 files / 697 tests PASS（Playwright対象と監査unitは別実行）
- [x] LINE予約枠の scoped Playwright — 1/1 PASS。390 / 500 / 1440pxで操作領域・内部scroll・document overflowを実寸確認
- [x] 変更済みfrontend `*.ts/tsx/js/mjs` 302 filesの scoped ESLint — PASS
- [x] R16 focused Vitest — 臨床5一覧・健診form/route guard・共通table/filterの 9 files / 54 tests PASS
- [x] R17 design audit / audit unit — C1〜C19違反0件、52/52 PASS。共有 `DataTableRow` / `SortableDataTableRow` のproduction `onClick` は各0件
- [x] R17 focused Vitest — row/link/button/drag、会計・見積・在庫・飼主・入院・master/shiftの15 files / 45 tests PASS。追加のreorder権限契約は5 files / 13 testsでRED 10件を再現後、13/13 PASS
- [x] R17 scoped ESLint / TypeScript — 明示57 filesのESLint warning 0、全production root・row helper callsiteの一時scoped tsconfig型検査PASS（一時file削除済み）
- [x] R17 security review — HIGH / MEDIUM 0件。LOWのreorder mutation callback二重guard不足を全対象で `canEdit` fail-closed化し、再検証PASS
- [x] R17 final read-only review — Critical / High / Medium / Low すべて0件。RBAC・医院分離・死亡lock・read-only・DnD・C19・test契約を再確認
- [x] R17 browser監査 — 7 route × 4 viewportで OK 24 / NG 0 / Partial 4。全28でdocument overflow 0、データ有り6 routeでrow-level interaction 0・44px未満control 0・business非GET 0。入院一覧4件はfixture 0行のためPartial
- [x] `/vercel-react-best-practices` v1.0.0 Quick Reference 64ルールの静的監査 — 15適合 / 27一部適合 / 18対象外 / 4実測未確認。全rule IDと根拠を2.2へ記録
- [ ] React性能runtime監査 — Network waterfall、lazy chunk timing、重複query、React Profilerは未実施。F7bおよびページ別V軸の完了条件として継続
- [x] 最終独立read-only review — R16差分のHIGH / MEDIUM 指摘 0件。健診のcreate/edit二重guard、会計の権限/医院境界、固有accessible name、通常カルテ作成routeの非退行を再確認。既存production 17箇所のrow clickは次バッチへ分離
- [x] scoped `git diff --check -- FE-refactor.md frontend` — PASS。全worktree版は本件外backend差分のEOF空行でFAILするため分離
- [x] browser監査を 1440 / 1200 / 800 / 500px で実施。下表の未確認、修正後未再確認、console問題は準拠へ数えない
- [x] R15会計/設定6 route — 実効権限「一般」のため権限拒否shellは24/24 Partial。共有AccessDenied修正後の代表`/accounting/reports`は4/4 OK、console error・非GET通信とも0件
- [x] dynamic GET 3 route（`/estimates/1`、`/inventory/7`、`/accounting/1024128`）— 4 viewportで OK 0 / NG 12 / Partial 0。期限超過表示、在庫不足表示/form surface、44px/label/read-only/RBACに未準拠を確認し各ページ行へ記録
- [x] R16臨床5一覧 — 4 viewportで OK 8 / NG 0 / Partial 12。カルテ・検査は各4/4 OK、トリミング・予防接種・健診はempty fixtureのため各4/4 Partial。全20でoverflow/console/business non-GET 0

### F4 browser監査の進捗

段階的に修正したため、下表の初回/中間件数は同一HEADの最終合否ではなく、P8〜P10のhotspot抽出結果である。抽出したNGはR10〜R15でTDD修正し、backend回復後に安全なGET routeの再監査を継続した。修正後の実測がない状態・権限・fixtureはOKへ繰り上げない。

| Group | 対象 | browser結果 | 修正後の状態 |
|---|---:|---|---|
| 認証・共通 | 4 route × 4 | P8/P9は16/16 OK。forgot/resetは修正後8/8直接到達、reset検診tokenも4/4補足。reportは4/4 OK | reportは死亡表示と`include_deceased=true`を確認。loginの初期session 401 resourceとreport printは厳密判定で残る |
| 会計・在庫・見積・シフト | 15 route × 4 + dynamic 3 route × 4 | `/accounting`、select-pet、`/inventory`、`/estimates` は16/16 OK。`/shifts` は4 NG→修正後4/4 OK。close/history/reportsの権限拒否shellは12/12 Partial。dynamic 3 routeは0/12 OK・12 NG | dynamic実測で見積期限超過の非色表示、在庫不足/form surface、会計の44px/label/read-only/RBAC問題を確認。`/new`は未訪問 |
| 診療・入院・トリミング・予防 | 24 route × 4 | 安全GET 14 routeは修正前 OK 48 / NG 8。最後のカルテ・入院NGを4 viewportで再計測し、修正後56/56 OK。R16の5一覧再監査はOK 8 / NG 0 / Partial 12 | 確定済み保存guardと空slot 19件の一意名/44pxを確認。R16はカルテ・検査が各4/4 OK、残る3一覧はempty fixture。非編集権限DnDはN/A、`/new`は未確認 |
| 受付・飼主・予約・集計 + LINE/LSTEP/manual | 16 route × 4 | 15安全routeの最終再計測は OK 56 / NG 0 / 未確認4 | 未確認4は予約0件の受付dynamic card/P10。`/owners/new`は未訪問。LINE予約枠は390pxをPlaywrightで追加確認 |
| 設定・マスタ | 24 route × 4 | 中間: main content OK 84 / LSTEP tags NG 4 / RBAC未確認8。closing/payment/tagsの権限拒否shellは12/12 Partial | tagsのroute/API権限と誤emptyをR14修正、AccessDenied h1をR15修正。代表4/4 OK。closing/payment/tagsの権限内bodyはfixture不足 |

現在の残りは、安全にread-only監査できない `/new`、実在ID不足のdetail/empty list、特定RBAC、受付の予約/死亡状態、printなどのfixture依存領域である。ガイド記載のadminは現環境で実効権限が「一般」であり、会計/設定6 routeの権限内bodyには到達できなかった。共有backendは回復後`/health` 200。ページ行は厳密なP1〜P10・V1〜V8完了条件に従い `[ ]` のまま維持する。

安全上の記録: `/medical-records/new?petId=1008579` は表示時に自動作成する現行仕様であり、click/input/submitなしでも 2026-07-21 23:08:20 JST に予約ID `2` とdraftカルテID `1425547` が作成された。直ちに全 `/new` routeのbrowser監査を停止し、作成物の削除は行っていない。以後 `/new` はread-only監査不能として未確認へ数えた。

完了条件は、機械監査の PASS だけではなく、**83ページすべてが `[x]`、例外は理由付きで `—`、`ui-design-compliance.md` と件数一致**であること。
