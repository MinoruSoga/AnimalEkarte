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

現時点の技術的結論は、**現行 tree の C1〜C19 自動監査は PASS、P1〜P10・該当V1〜V8の完全確認は 83/83、BLOCKED 0/83**。F7b 64ルールは35適合・11根拠付き例外・18対象外・未裁定0。臨床5 routeはgenerated/feature API typeに適合するsentinel fixtureをPlaywright内だけでfulfillし、same-origin・clinic・CSRF headerとcontinued-to-backend 0を検証した。最終frontend実装は `1e7faf56e` に集約済みで、本書は削除せず完了・監査証跡として保持する。

## 2. 全ページ共通チェック（P 軸）

各ページは、該当する全項目を確認してから `[x]` にする。

- [x] P1 図地: ページは canvas-soft、カード・パネル・入力は white surface、境界は 1px hairline
- [x] P2 色用途: brand は CTA・リンク・active/focus のみ。status/semantic 色を構造や装飾へ流用しない
- [x] P3 タイポ: heading/title/body/caption/eyebrow の役割、size、weight、line-height、letter-spacing が一致
- [x] P4 余白: 4/8/12/16/24/28/32px のスケールに従い、20px 等の仕様外段を使わない
- [x] P5 形状: Primary CTA は pill、utility は 8px、入力は 4px、カードは 12px、modal は 16px
- [x] P6 Elevation: 通常カードは Level 0、浮動面は Level 1、modal/toast は Level 2
- [x] P7 Table/Form: table header は canvas-soft + eyebrow、body は body-sm、セルは 12px 16px。入力は white/body-sm/4px
- [x] P8 Responsive: 1440px、1080〜1300px、768〜840px、600px 以下で崩れず、mobile は単一カラム
- [x] P9 Touch/状態: 操作対象は 44×44px 以上。focus/pressed/disabled/RBAC 非活性が識別可能
- [x] P10 臨床安全: danger、死亡、異常、期限、権限不足の表示を退行させない

### 2.1 React実装・性能チェック（V軸）

変更したページと、そのページが直接利用するcomponent/hookに `/vercel-react-best-practices` を適用する。該当ルールはrule ID（例: `async-parallel`）と確認根拠をページ行または検証記録へ残す。Next.js/RSC/server action固有ルールは、Vite SPAである本体アプリには機械的に適用せず `—` とする。Vercelルールとproject規約が競合する場合はproject規約を優先し、例外理由を記録する（例: feature外からのimportはproject必須のfeature indexを使う）。

- [x] V1 Waterfall: 2適合 / 1根拠付き例外 / 2対象外。owner作成の依存graphを整理し、会計明細の直列writeは部分成功を防ぐfail-stop例外とした（`async-*`）
- [x] V2 Bundle: 2適合 / 2計測根拠付き例外 / 1対象外。Viteの実bundleと遷移INPを測定して裁定した（`bundle-*`）
- [x] V3 Server: 8/8対象外を確認済み。本体はVite CSRで、RSC/server action/SSR surfaceを持たない（`server-*`）
- [x] V4 Client fetch: 3適合 / 1対象外。loader/query重複とreduced-motion listener重複を解消した（`client-*`）
- [x] V5 Re-render: 10適合 / 5安全・計測根拠付き例外。manual検索と診療tabをProfiler/interactionで判定した（`rerender-*`）
- [x] V6 Rendering: 5適合 / 2計測根拠付き例外 / 4対象外。CSRの描画costとstate保持要件から裁定した（`rendering-*`）
- [x] V7 JavaScript: 11適合 / 1医院境界例外 / 1対象外。hot loopを一回走査・Map/Set/cacheへ変更した（`js-*`）
- [x] V8 Advanced: 2適合 / 1対象外。具体的なhotspotがある場合だけ適用する（`advanced-*`）

### 2.2 `/vercel-react-best-practices` 64ルール実査結果

> 監査日: 2026-07-23
> 基準: ローカルskill `vercel-react-best-practices` v1.0.0 の Quick Reference 64ルール。
> 範囲: 現行working treeの `frontend/src` production code、`frontend/index.html`、`frontend/package.json`、route/provider/query設定。test、generated type、`shared-liff`、LIFF/LINE予約別アプリは除外。

判定は、`✅ 適合`＝修正または静的・実行時確認で既知の違反なし、`🟦 根拠付き例外`＝性能利益より財務/臨床/RBAC/医院境界の正しさを優先したか、計測上変更利益がない、`— 対象外`＝Vite CSRに該当surfaceなし、とする。未裁定statusは使用しない。

| V軸 | 適合 | 根拠付き例外 | 対象外 | 未裁定 | カテゴリ判定 |
|---|---:|---:|---:|---:|---|
| V1 Waterfall | 2 | 1 | 2 | 0 | ✅ 裁定済み |
| V2 Bundle | 2 | 2 | 1 | 0 | ✅ 裁定済み |
| V3 Server | 0 | 0 | 8 | 0 | — 8/8対象外確認済み |
| V4 Client fetch | 3 | 0 | 1 | 0 | ✅ 裁定済み |
| V5 Re-render | 10 | 5 | 0 | 0 | ✅ 裁定済み |
| V6 Rendering | 5 | 2 | 4 | 0 | ✅ 裁定済み |
| V7 JavaScript | 11 | 1 | 1 | 0 | ✅ 裁定済み |
| V8 Advanced | 2 | 0 | 1 | 0 | ✅ 裁定済み |
| **合計** | **35** | **11** | **18** | **0** | **64/64 最終裁定済み** |

#### V1 Waterfall（5ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| [x] | `async-defer-await` | ✅ 適合 | owner作成後のcache invalidationとoptional pet作成を、結果が必要になる境界まで遅延して待つ |
| [x] | `async-parallel` | 🟦 根拠付き例外 | read処理は `Promise.all`。会計明細writeは途中失敗後の追加送信を止め、部分成功を広げないfail-stop順序を回帰testで固定 |
| [x] | `async-dependencies` | ✅ 適合 | owner作成を先行条件、pet作成とcache invalidationを後続処理として明示し、独立部分のみ並列化 |
| — | `async-api-routes` | — 対象外 | `package.json:7-16` と `app/router.tsx:1-9` のVite SPAにfrontend API route/server handlerはない |
| — | `async-suspense-boundaries` | — 対象外 | SSR streaming/RSCなし。client lazy境界はV2で確認する |

#### V2 Bundle（5ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| [x] | `bundle-barrel-imports` | 🟦 根拠付き例外 | feature indexはproject必須。Vite build artifactでlucide vendor chunk 28,511 bytesを確認し、root import 208件をper-iconへ全面変更する計測利益を確認できないため維持 |
| [x] | `bundle-dynamic-imports` | ✅ 適合 | route `lazy` 82件、`React.lazy` 25件。`app/routes/app-routes.tsx:13-50` とrechartsを遅延する `features/medical-records/components/VitalsTab/VitalsTab.tsx:18-22` を確認 |
| — | `bundle-defer-third-party` | — 対象外 | analytics/logging SDKのdependency・初期化なし。UI通知のsonnerは対象外 |
| [x] | `bundle-conditional` | ✅ 適合 | recharts graphと予約のMonth/Week viewを選択時だけimport/renderする |
| [x] | `bundle-preload` | 🟦 根拠付き例外 | inventory遷移はINP 32ms、CLS 0。82 lazy routeの一括/予測preloadは未使用chunk通信を増やすため導入しない |

#### V3 Server（8ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
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

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| [x] | `client-swr-dedup` | ✅ 適合 | owner loader結果をQueryClientへhydrateし、同一ownerの直後の二重GETを防ぐ。loader/query testで固定 |
| [x] | `client-event-listeners` | ✅ 適合 | reduced-motionをmodule singleton + `useSyncExternalStore` に集約し、複数consumerでもnative listener 1件・cleanupをtest |
| — | `client-passive-event-listeners` | — 対象外 | scroll/wheel/touch系native listenerなし。既存はbeforeunload/storage/beforeprint/MediaQuery change |
| [x] | `client-localstorage-schema` | ✅ 適合 | 保存はversion付きkey `auth_current_clinic:v1` のscalar 1件のみ（`lib/current-clinic.ts:1-14`）。session/authはhttpOnly Cookie |

#### V5 Re-render（15ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| [x] | `rerender-defer-reads` | 🟦 根拠付き例外 | Auth/RBAC contextは権限変更を即時反映する安全境界であり購読を遅延しない。Profilerで権限外の支配的hotspotなし |
| [x] | `rerender-memo` | ✅ 適合 | manual検索結果をmemo化。入力時React commitは130.9/116.4msから3.4/3.0msへ短縮 |
| [x] | `rerender-memo-with-default-value` | ✅ 適合 | `MedicalRecordPrintView` の空treatmentsをmodule定数へhoistし参照を安定化 |
| [x] | `rerender-dependencies` | ✅ 適合 | ClinicHoliday modal等のeffect依存を必要なprimitiveへ限定 |
| [x] | `rerender-derived-state` | 🟦 根拠付き例外 | permission booleanはAuth contextのclinic/session変更と同時に再計算する必要があり、RBAC安全性を優先 |
| [x] | `rerender-derived-state-no-effect` | 🟦 根拠付き例外 | 会計detailのAPI→local form同期はaction closureと表示値の同一snapshotを保証するため維持。回帰testで固定 |
| [x] | `rerender-functional-setstate` | ✅ 適合 | pagination/state更新をprevious stateから導出するfunctional setterへ統一 |
| [x] | `rerender-lazy-state-init` | ✅ 適合 | `ShiftCalendarPage.tsx:38`、`use-hospitalization-form.ts:39` 等の計算初期値はfunction initializerを使用 |
| [x] | `rerender-simple-expression-in-memo` | ✅ 適合 | `DailyDateNav` 等の単純比較から不要なmemoを除去 |
| [x] | `rerender-split-combined-hooks` | 🟦 根拠付き例外 | Authはsession・clinic・permissionを原子的に切替える安全境界であり、分割による一時的不整合を避ける |
| [x] | `rerender-move-effect-to-event` | 🟦 根拠付き例外 | 定期健診の遷移は全write成功後だけ許可する。effect監視を維持し、途中失敗時に遷移しないtestを追加 |
| [x] | `rerender-transitions` | ✅ 適合 | 診療tab切替を`useTransition`化し、interaction INPを258msから45ms、React commitを約4msへ短縮 |
| [x] | `rerender-use-deferred-value` | ✅ 適合 | manual Fuse検索へdeferred queryを適用し、入力完了時間を2.7sから0.2sへ短縮、CLS 0 |
| [x] | `rerender-use-ref-transient-values` | ✅ 適合 | dirtyの最新値をrefで保持しstable callbackから読む（`hooks/use-side-peek-dirty.ts:23-56`） |
| [x] | `rerender-no-inline-components` | ✅ 適合 | production TSXのcomponent宣言を走査し、component内component定義の既知候補なし |

#### V6 Rendering（11ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| — | `rendering-animate-svg-wrapper` | — 対象外 | 自前SVG animationなし |
| [x] | `rendering-content-visibility` | 🟦 根拠付き例外 | 診療tabの実測commitは約4ms。臨床状態を保持するhidden panelへcontent-visibilityを加える利益がなく導入しない |
| [x] | `rendering-hoist-jsx` | ✅ 適合 | static option/empty valueをmodule scopeへhoistし、renderごとの生成を除去 |
| — | `rendering-svg-precision` | — 対象外 | 自前inline SVGなし。lucide vendor iconは監査対象外 |
| — | `rendering-hydration-no-flicker` | — 対象外 | `createRoot` のCSRでhydrateしない（`main.tsx:1-27`） |
| — | `rendering-hydration-suppress-warning` | — 対象外 | hydration mismatch surfaceなし |
| [x] | `rendering-activity` | 🟦 根拠付き例外 | tabの臨床入力state保持を優先。計測commit約4msでActivity導入の利益がなく、現行mount semanticsを維持 |
| [x] | `rendering-conditional-render` | ✅ 適合 | JSX conditional候補はternary/明示null。project規約違反の `condition && <JSX>` は既知0件 |
| [x] | `rendering-usetransition-loading` | ✅ 適合 | 診療tabに`useTransition`を適用し、pending表示とINP 45msを実測 |
| [x] | `rendering-resource-hints` | ✅ 適合 | Google Fontsへpreconnectを設定（`frontend/index.html:12-15`） |
| [x] | `rendering-script-defer-async` | ✅ 適合 | entry scriptはdefer相当の `type="module"`（`frontend/index.html:19`） |

#### V7 JavaScript（13ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| [x] | `js-batch-dom-css` | ✅ 適合 | 直接style連打はなく、print modeはroot/bodyのclass切替へ集約（`ManualPage.tsx:124-131`） |
| [x] | `js-index-maps` | ✅ 適合 | trimming item lookupをMap化し、selected IDごとの反復`find`を除去 |
| [x] | `js-cache-property-access` | ✅ 適合 | hot loop候補の意味レビューで、同一deep propertyを反復取得する既知候補なし |
| [x] | `js-cache-function-results` | ✅ 適合 | treatment itemの正規化済み検索文字列を入力data単位でcacheし、queryごとの全件再計算を除去 |
| [x] | `js-cache-storage` | 🟦 根拠付き例外 | axiosはrequest直前に現在clinic IDを読む必要がある。医院切替後に旧IDを送らないtestを追加し、tenant isolationを優先 |
| [x] | `js-combine-iterations` | ✅ 適合 | staff validation等の連続`filter().map()`を単一走査へ統合 |
| — | `js-length-check-first` | — 対象外 | 大配列の高コストな等価比較surfaceを検出せず |
| [x] | `js-early-exit` | ✅ 適合 | validation/filter/handlerでearly returnを使用し、深い分岐の既知hotspotなし |
| [x] | `js-hoist-regexp` | ✅ 適合 | `new RegExp` は `lib/sanitize.ts:7`、`get-exam-type-fields.ts:27-37` のmodule level。loop内生成なし |
| [x] | `js-min-max-loop` | ✅ 適合 | owner reportの最古/最新record算出をsortなしの単一loopへ変更 |
| [x] | `js-set-map-lookups` | ✅ 適合 | staff group membershipのloop内`includes`をSet lookupへ変更 |
| [x] | `js-tosorted-immutable` | ✅ 適合 | 並び替えは`toSorted`または明示copyを使い、入力配列をmutationしないことを確認 |
| [x] | `js-flatmap-filter` | ✅ 適合 | pet examination flatten/filterを`flatMap`の単一pipelineへ変更 |

#### V8 Advanced（3ルール）

| チェック | Rule ID | 判定 | 根拠・例外 |
|---|---|---|---|
| — | `advanced-event-handler-refs` | — 対象外 | handler freshnessが原因の高頻度再購読hotspotを検出せず。必要になるまで導入しない |
| [x] | `advanced-init-once` | ✅ 適合 | QueryClientはmoduleで1回生成し、session restore promiseもmount内でrefに保持（`features/auth/hooks/use-auth.tsx:68-78`） |
| [x] | `advanced-use-latest` | ✅ 適合 | `use-side-peek-dirty.ts:23-56` はlatest refをstable `confirmDiscard` callbackから読む |

## 3. 共通ギャップの対応結果

| ID | 判定 | 内容 | 主な影響 |
|---|---|---|---|
| R1 | ✅ | `PageLayout` / `STYLE.pageContent` を `py-6`（24px）へ統一。C16 が spacing `*-5` の再導入を禁止 | `PageLayout` を直接使うページと、共通 shell 経由ページ |
| R2 | ✅ | `FormHeader` の `h1` を title role（20px/600/1.4/-0.125px）へ変更し、`PageLayout.description` も接続 | 全 `PageLayout`、受付、予約管理、LINE予約枠 |
| R3 | ✅ | table primitive / `STYLE.table*` と呼び出し側36ファイル195行を header eyebrow、body-sm、cell 12px 16pxへ統一し、C18で再導入を禁止。手書き `<th>/<td>` も83ページの最終監査でP7照合済み | 共通 table、各一覧・フォーム内明細 |
| R4 | ✅ | `/lstep/checkup-sync` を `PageLayout` 化し、404 / route error fallback に canvas-soft を付与 | 健診連携、404、route error |
| R5 | ✅ | `/accounting/close` の「プレビュー」を pill + 16px horizontal paddingへ変更 | レジ締め |
| R6 | ✅ | `/accounting/close/history` の手書き table を共通 header/body typography・paddingへ変更 | レジ締め履歴 |
| R7 | ✅ | route 表面の直接 named color 11件を `C` tokenへ移行。C15 が routes/pages の再導入を禁止 | 受付、トリミング、Lステップ健診/配信、マニュアル |
| R8 | ✅ | 指摘5画面を mobile single-column / full-width に変更し unit test を追加。全83ページを4 viewportで最終確認 | 受付、LINE予約枠、Lステップ配信、見積作成/詳細 |
| R9 | ✅ | C8 を相対パス完全一致の page 9件/helper 5件へ分離し、drop-shadow・named color・20px spacing・CSS shadow・Table overrideを C10/C15〜C18 へ追加。意味的誤用・手書きtable・responsive・P10はreviewとbrowser matrixで最終照合済み | 監査スクリプト全体 |
| R10 | ✅ | 検出済みの固定複数列formを mobile single-columnへ変更し、飼主一覧・予約toolbar・設定tableは狭幅で内部scroll/wrapするようTDD。実データ依存routeもlive/typed fixtureで再監査済み | 在庫、入院、診療、予防、飼主、予約、設定form |
| R11 | ✅ | 共通 Button/Tabs/Switch/icon action/DatePicker/CalendarNav/DeleteIconと個別raw control・認証導線を44px以上へ変更。見た目サイズを保つ場合もfocusable要素のhit areaを確保し、全83ページを4 viewportで再監査済み | 全Button/Switch利用画面、日付選択、設定、認証 |
| R12 | ✅ | 定期健診formは実APIフローに必要な `medical-records:create && edit` でfieldset/save/actionをfail-closedにし、実在する過去の次回来院日だけ非色依存の「期限切れ」を追加。仮の既定日付は `-` へ変更し、死亡表示などの実データ状態もbrowser確認済み | 定期健診、ワクチン、飼主/ペット状態 |
| R13 | ✅ | browser実測と独立reviewで検出した44px未満、focus/label、重複accessible name、modal/drawer description・focus・狭幅overflow、選択状態をTDD修正。Sidebar、SortableHeader、DatePicker、checkbox、予約/受付card、シフト、manual、LSTEPタグ/LINE送信drawer、診療formを含み、fixture依存領域もtyped sentinelで最終確認済み | 全共通shell、一覧table、受付/予約、シフト、manual、診療・検査 |
| R14 | ✅ | route/APIのRBACを実backend契約へ統一。LSTEPタグ/配信は `lstep-analytics:view`、manual override/editは `manual-edit` でAPI/UIを抑止。forgot/resetはsession復元と401 redirectの対象外にし、login/logout後の認証snapshotも再取得する。権限内・拒否状態を最終browser matrixで確認済み | LSTEP、manual、認証回復route |
| R15 | ✅ | 会計・設定のread-only監査で検出した40px操作、休診日/特別期間の重複削除名、休診曜日checkbox、AccessDeniedの非h1をTDD修正。権限内本体とRBAC拒否状態をapproved fixtureで最終確認済み | レジ締め、月次レポート、締め時間設定、全RBAC拒否画面 |
| R16 | ✅ | 臨床5一覧のmouse-only row遷移を44px native cell linkへ置換し、detail/action名へstable ID、検索clearへ44px hit areaと入力余白、無効staffへ非色依存説明を追加。カルテ会計導線は `accounting:view` と同一医院を要求し、健診CTA/form/子routeは `medical-records:create && edit`、編集導線は `view && edit` へ統一。共有row click 29箇所をR17で解消済み | カルテ、トリミング、検査、予防接種、定期健診、共通table/filter |
| R17 | ✅ | `DataTableRow` / `SortableDataTableRow` 29箇所に加え、raw `TableRow onClick` 7箇所・raw `tr onClick` 3箇所を廃止。用途別の44px native link/button/drag handleへ置換し、C19を4種すべてへ拡張。production match 0、監査unitで再導入を禁止 | 会計・在庫・見積・飼主・入院一覧、医院/シフト/各マスタ、共有table/DnD |
| R18 | ✅ | 会計詳細の32px/label/read-only/RBAC fail-open、在庫編集の不足表示/form surface/RBAC、見積詳細の期限超過色依存を回帰test先行で修正し4 viewportで再確認 | 会計詳細、在庫編集、見積詳細 |
| R19 | ✅ | F7bを35適合・11根拠付き例外・18対象外へ最終裁定し、read-only 4 viewport matrixを83ページで完走。臨床5ページはtyped sentinel fixtureとfail-closed interceptionでbackend write 0を保証 | 本体83ページ、直接consumer、performance/runtime audit |

補足:

- `p-5` / `m-5` / `gap-5` 等の20px spacingは C16 導入時に31件を検出し、画面用26件を仕様内スケールへ移行した。印刷ビュー5件は画面用 DESIGN.md の対象外として明示除外した。
- `design-audit` の PASS は有効な回帰ガードだが、P1〜P10 とV1〜V8の完全準拠を意味しない。
- `drop-shadow`、CSS の `box-shadow` / `filter: drop-shadow()`、route/page の Tailwind named color は機械化済み。正しい typography role の意味選択と全画面 responsive は別チェックが必要。

## 4. ページ別チェックリスト（本体84リーフルート）

`F4✓` は、当該routeをproduction build相当で 1440 / 1200 / 800 / 500px の4 viewportに描画し、P1〜P10、h1/path、document overflow、全操作のaccessible nameと実効44px、console error、HTTP 4xx/5xx、business non-GETを確認したことを示す。`V64✓` は §2.2 の64 rule ID（該当ruleの修正・計測例外・対象外）を当該routeと直接consumerへ照合したことを示す。共通証拠は `e2e/ui-design-compliance-readonly.spec.ts` の製品ページ83/83 PASS（suite全体92/92）、C1〜C19違反0、監査unit 53/53 PASS。

### 認証・共通

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [x] | ログイン | `/login` | `Login` | F4✓ / V64✓。public routeで`/me` XHRなし、console/network error 0 |
| [x] | パスワードを忘れた方 | `/forgot-password` | `ForgotPasswordPage` | F4✓ / V64✓。`/me`・refresh・login redirectなし |
| [x] | パスワード再設定 | `/reset-password` | `ResetPasswordPage` | F4✓ / V64✓。token無しで直接到達、`/me`・refresh・login redirectなし |
| [x] | 飼主カルテレポート | `/owners/:id/report` | `OwnerReport` | F4✓ / V64✓ / P10。死亡表示・44pxペット切替・`include_deceased=true`を確認。印刷帳票は§5の対象外 |
| — | 404 Not Found | `*` | inline fallback | canvas-soft 修正・unit test済み。製品ページ数には含めない |

### 会計・在庫・見積・シフト

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [x] | 会計一覧 | `/accounting` | `AccountingList` | F4✓ / V64✓。raw row click廃止、44px native detail link |
| [x] | 会計ペット選択 | `/accounting/select-pet` | `AccountingPetSelection` | F4✓ / V64✓ |
| [x] | 会計登録 | `/accounting/new` | `AccountingDetail` | F4✓ / V64✓ / RBAC。business non-GET 0 |
| [x] | 会計詳細 | `/accounting/:id` | `AccountingDetail` | F4✓ / V64✓ / P7/P9/P10/RBAC。数値操作44px・label・完了/支払済みread-only・権限API fail-closeを回帰testとbrowserで確認 |
| [x] | レジ締め | `/accounting/close` | `CashRegisterClosePage` | F4✓ / V64✓ / RBAC。印刷操作44px、read-only mount |
| [x] | レジ締め履歴 | `/accounting/close/history` | `CashRegisterHistoryPage` | F4✓ / V64✓ / RBAC。raw row click廃止 |
| [x] | 月次集計レポート | `/accounting/reports` | `AccountingReportsPage` | F4✓ / V64✓ / RBAC。集計条件・税率link 44px |
| [x] | 在庫一覧 | `/inventory` | `InventoryList` | F4✓ / V64✓ / RBAC。edit権限時だけ44px native link |
| [x] | 在庫登録 | `/inventory/new` | `InventoryForm` | F4✓ / V64✓ / RBAC。business non-GET 0 |
| [x] | 在庫編集 | `/inventory/:id` | `InventoryForm` | F4✓ / V64✓ / P1/P7/P9/P10/RBAC。在庫不足を非色依存表示、white form surface/hairline、権限なしread-onlyを回帰testとbrowserで確認 |
| [x] | 見積一覧 | `/estimates` | `EstimateList` | F4✓ / V64✓。raw row click廃止、locked/view-only維持 |
| [x] | 見積作成 | `/estimates/new` | `EstimateForm` | F4✓ / V64✓。form操作44px、business non-GET 0 |
| [x] | 見積詳細 | `/estimates/:id` | `EstimateDetail` | F4✓ / V64✓ / P10。期限超過を非色依存の「期限切れ」で回帰testとbrowser確認 |
| [x] | 見積編集 | `/estimates/:id/edit` | `EstimateForm` | F4✓ / V64✓ / RBAC。実レスポンスshapeをdraftへcloneしたfrontend interception、DB変更0 |
| [x] | シフトカレンダー | `/shifts` | `ShiftCalendarPage` | F4✓ / V64✓。全操作44px以上 |

### 診療・入院・トリミング・予防

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [x] | カルテ一覧 | `/medical-records` | `MedicalRecords` | F4✓ / V64✓ / P10/RBAC。同一医院・会計権限境界、死亡/無効医表示、44px native detail linkをunit + browser確認 |
| [x] | カルテペット選択 | `/medical-records/select-pet` | `MedicalRecordPetSelection` | F4✓ / V64✓ / P10 |
| [x] | カルテ作成 | `/medical-records/new` | `MedicalRecordForm` | F4✓ / V64✓ / P10。synthetic pet GET、予約/カルテPOSTをlocal fulfillし、synthetic detailへproduction遷移。continued-to-backend 0 |
| [x] | カルテ編集 | `/medical-records/:id` | `MedicalRecordForm` | F4✓ / V64✓ / P10/RBAC。確定済み保存disabled、死亡/lock、44px、診療tab transitionを確認 |
| [x] | 入院・ホテル一覧 | `/hospitalization` | `HospitalizationList` | F4✓ / V64✓ / P10/RBAC。生存行native link、死亡行lock/plain text、空slotの一意名・44pxを確認 |
| [x] | 入院・ホテル ペット選択 | `/hospitalization/select-pet` | `HospitalizationPetSelection` | F4✓ / V64✓ / P10 |
| [x] | 入院・ホテル登録 | `/hospitalization/new` | `HospitalizationForm` | F4✓ / V64✓ / P10/RBAC。business non-GET 0 |
| [x] | 入院・ホテル詳細 | `/hospitalization/:id` | `HospitalizationDetail` | F4✓ / V64✓ / P10。typed hospitalization/care-plan/daily-record fixtureで臨床状態・費用・overflowを確認 |
| [x] | 入院・ホテル編集 | `/hospitalization/:id/edit` | `HospitalizationForm` | F4✓ / V64✓ / P10。typed form/treatment hydration、治療・割引の参照専用状態、mount時write 0を確認 |
| [x] | トリミング一覧 | `/trimming` | `TrimmingList` | F4✓ / V64✓。empty production stateを4 viewport確認、row link/無効staffの非色依存挙動はunit testで固定 |
| [x] | トリミング ペット選択 | `/trimming/select-pet` | `TrimmingPetSelection` | F4✓ / V64✓ |
| [x] | トリミング登録 | `/trimming/new` | `TrimmingForm` | F4✓ / V64✓ / RBAC。business non-GET 0 |
| [x] | トリミング編集 | `/trimming/:id` | `TrimmingForm` | F4✓ / V64✓ / P10。typed trimming/course/options/staff fixtureでedit stateを確認 |
| [x] | 検査一覧 | `/examinations` | `ExaminationsList` | F4✓ / V64✓。raw row click廃止、44px native detail link、一意な操作名 |
| [x] | 検査ペット選択 | `/examinations/select-pet` | `ExaminationPetSelection` | F4✓ / V64✓ |
| [x] | 検査登録 | `/examinations/new` | `ExaminationForm` | F4✓ / V64✓ / RBAC。business non-GET 0 |
| [x] | 検査編集 | `/examinations/:id` | `ExaminationForm` | F4✓ / V64✓ / P10。空項目名・履歴期間・form labelをunit + browser確認 |
| [x] | ワクチン一覧 | `/vaccinations` | `VaccinationList` | F4✓ / V64✓ / P10。empty production stateを4 viewport確認、row/action挙動はunit testで固定 |
| [x] | ワクチン ペット選択 | `/vaccinations/select-pet` | `VaccinationPetSelection` | F4✓ / V64✓ / P10 |
| [x] | ワクチン登録 | `/vaccinations/new` | `VaccinationForm` | F4✓ / V64✓ / P10/RBAC。business non-GET 0 |
| [x] | ワクチン編集 | `/vaccinations/:id` | `VaccinationForm` | F4✓ / V64✓ / P10。typed vaccination/pet/vaccine/doctor fixtureで担当医帰属とedit stateを確認 |
| [x] | 定期健診一覧 | `/checkups` | `CheckupsList` | F4✓ / V64✓ / P10/RBAC。期限切れ非色依存表示と権限二重guardをunit + browser確認 |
| [x] | 定期健診ペット選択 | `/checkups/select-pet` | `CheckupPetSelection` | F4✓ / V64✓ / RBAC。create+editをfail-closed |
| [x] | 定期健診登録 | `/checkups/new` | `CheckupForm` | F4✓ / V64✓ / P10/RBAC。途中write失敗時の部分保存/遷移抑止をunit test、browserはnon-GET 0 |

### 受付・飼主・予約・集計

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [x] | 受付 | `/` | `Reception` | F4✓ / V64✓ / P10。production empty stateと5状態/死亡表示のcomponent regression testを照合 |
| [x] | 飼主一覧 | `/owners` | `OwnersList` | F4✓ / V64✓ / RBAC/医院境界。同一医院+edit時だけ44px native link |
| [x] | 飼主登録 | `/owners/new` | `OwnerForm` | F4✓ / V64✓ / P10/RBAC。business non-GET 0 |
| [x] | 飼主編集 | `/owners/:id` | `OwnerForm` | F4✓ / V64✓ / P10/RBAC。危険人物・配信制御switchを一意名/44px、LINE drawerをmobile/focus対応 |
| [x] | 集計ダッシュボード | `/aggregation` | `AggregationDashboardPage` | F4✓ / V64✓。canonical query、filter control 44px |
| [x] | 予約管理 | `/reservations` | `ReservationManagement` | F4✓ / V64✓ / P10/RBAC |

### Lステップ・LINE予約・マニュアル

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [x] | Lステップ健診連携 | `/lstep/checkup-sync` | `CheckupSyncPage` | F4✓ / V64✓ / RBAC |
| [x] | Lステップ配信モニター | `/lstep/delivery-monitor` | `LstepDeliveryMonitorPage` | F4✓ / V64✓ / RBAC。拒否時AccessDenied・領域API 0 |
| [x] | Lステップ分析 | `/lstep/analytics` | `LstepAnalyticsPage` | F4✓ / V64✓ / RBAC |
| [x] | LINE予約設定 index | `/line-reservation` | `LineReservationSettings` | F4✓ / V64✓ |
| [x] | LINE予約設定 | `/line-reservation/settings` | `LineReservationSettings` | F4✓ / V64✓。全switch/inputの一意名・44px |
| [x] | LINE予約ページエディタ | `/line-reservation/page-editor` | `LineReservationPageEditor` | F4✓ / V64✓ / RBAC |
| [x] | LINE予約枠設定 | `/line-reservation/slots` | `LineReservationSlotsSettings` | F4✓ / V64✓。pathname固定、選択state queryのみ許容、calendar内部scroll |
| [x] | 医院マスタ設定 | `/settings/clinic` | `ClinicMasterSettings` | F4✓ / V64✓ / RBAC/医院境界。raw row click廃止 |
| [x] | マニュアルトップ | `/manual` | `ManualPage` | F4✓ / V64✓（`rerender-memo`,`rerender-use-deferred-value`）。shell対象、本文は§5対象外、権限なしoverride API 0 |
| [x] | マニュアル記事 | `/manual/:category/:slug` | `ManualPage` | F4✓ / V64✓。canonical article route、modal focus/cleanup、権限なしoverride API 0 |

### 設定・マスタ

| 完了 | ページ | Path | Component | 現状 |
|---|---|---|---|---|
| [x] | 設定トップ | `/settings` | `MasterSettingsIndex` | F4✓ / V64✓ / RBAC |
| [x] | 職員マスタ | `/settings/staff` | `StaffSettings` | F4✓ / V64✓ / P10/RBAC（`js-set-map-lookups`,`js-combine-iterations`） |
| [x] | 診療項目マスタ | `/settings/treatment-items` | `TreatmentPlanMaster` | F4✓ / V64✓ / RBAC。詳細button・権限連動44px drag handle |
| [x] | 診断名マスタ | `/settings/diagnosis` | `DiagnosisSettings` | F4✓ / V64✓ / RBAC。権限なしdrag/action fail-close |
| [x] | 動物種マスタ | `/settings/animal-species` | `AnimalSpeciesSettings` | F4✓ / V64✓ / RBAC。read-only詳細、権限連動drag |
| [x] | トリミングマスタ | `/settings/trimming` | `TrimmingSettings` | F4✓ / V64✓ / RBAC |
| [x] | トリミングコース種別 | `/settings/trimming-course-type` | `TrimmingCourseTypeSettings` | F4✓ / V64✓ / RBAC |
| [x] | 薬剤マスタ | `/settings/medicine` | `MedicineSettings` | F4✓ / V64✓ / P10/RBAC。raw row click廃止、sortable操作44px |
| [x] | 予約種別マスタ | `/settings/reservation-type` | `ReservationTypeSettings` | F4✓ / V64✓ / RBAC。group toggleに一意名/expanded state |
| [x] | 入院・ホテルマスタ | `/settings/hospitalization` | `HospitalizationSettings` | F4✓ / V64✓ / P10/RBAC |
| [x] | ケージマスタ | `/settings/cage` | `CageSettings` | F4✓ / V64✓ / RBAC。read-only詳細、権限連動drag |
| [x] | 物販品マスタ | `/settings/merchandise-items` | `MerchandiseItemSettings` | F4✓ / V64✓ / RBAC。read-only詳細、権限連動drag |
| [x] | 保険マスタ | `/settings/insurance` | `InsuranceSettings` | F4✓ / V64✓ / RBAC |
| [x] | 職種マスタ | `/settings/occupations` | `OccupationSettings` | F4✓ / V64✓ / RBAC |
| [x] | 権限グループマスタ | `/settings/permission-groups` | `PermissionGroupSettings` | F4✓ / V64✓ / P10/RBAC。read-only詳細、権限連動drag |
| [x] | 問診テンプレート | `/settings/inquiry-templates` | `InterviewTemplateSettings` | F4✓ / V64✓ / P10/RBAC |
| [x] | 主訴マスタ | `/settings/interview/chief-complaint` | `ChiefComplaintSettings` | F4✓ / V64✓ / P10/RBAC。選択操作44px |
| [x] | 問診テンプレート（interview） | `/settings/interview/templates` | `InterviewTemplateSettings` | F4✓ / V64✓ / P10/RBAC |
| [x] | シフトテンプレート | `/settings/shift-templates` | `ShiftTemplateSettings` | F4✓ / V64✓ / RBAC。create/edit/delete/reorder fail-close、44px drag handle |
| [x] | 締め時間設定 | `/settings/closing-time` | `ClosingSettingsPage` | F4✓ / V64✓ / RBAC。追加/削除/checkboxの一意名・44px、狭幅shrinkをbrowser確認 |
| [x] | 支払方法マスタ | `/settings/payment-methods` | `PaymentMethodSettings` | F4✓ / V64✓ / RBAC。read-only mount、raw row click廃止 |
| [x] | 割引キャンペーン | `/settings/campaigns` | `CampaignSettings` | F4✓ / V64✓ / RBAC |
| [x] | Lステップ連携設定 | `/settings/integrations/lstep` | `LstepSettingsPage` | F4✓ / V64✓ / RBAC。switchの一意名・44px |
| [x] | Lステップタグ管理 | `/settings/lstep/tags` | `LstepTagManagementPage` | F4✓ / V64✓ / RBAC。AccessDenied、領域API 0、mobile drawer/focus/44pxを確認 |

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
- [x] F3a 監査強化: C8 path化、drop-shadow、spacing、route named color、CSS shadow、Table primitive override、共有/raw row clickを C10/C15〜C19 で機械化
- [x] F3b 検証補完: table値・responsive hotspot は scoped unit test、全画面は browser checklist で追跡
- [x] F4 ページ再監査: 83/83 PASS、BLOCKED 0/83。全行でP1〜P10・該当V1〜V8と4 viewportを確認し、synthetic routeはbackend write 0
- [x] F5 文書同期: `docs/spec/ui-design-compliance.md` のサマリーを「機械適合」と「完全確認」に分離
- [x] F6 検証: design-audit、監査unit test、変更範囲のVitest/ESLint、代表viewportのブラウザ確認を記録
- [x] F7a React性能baseline: `/vercel-react-best-practices` 64/64ルールを現行本体アプリへ静的適用し、初回分類を記録
- [x] F7b React性能完了: 初回31候補を修正/根拠付き例外/Network・React Profiler判定へ移行。最終35適合 / 11根拠付き例外 / 18対象外 / 未裁定0

### 2026-07-21〜22 baseline（2026-07-23最終記録で置換）

- [x] C1〜C19、scoped Vitest/ESLint、初回browser hotspot抽出を実施
- [x] 初回監査で検出したrow click、responsive、RBAC、P10、性能候補をR16〜R19の修正・最終計測へ引き継いだ

### 2026-07-23 最終照合記録

- [x] route inventory — 本体83ページ + 404対象外1、redirect-only 12、別アプリ除外をrouter・本書・`ui-design-compliance.md`で一致確認
- [x] raw row click — production `TableRow` / raw `tr` / `DataTableRow` / `SortableDataTableRow` の行全体`onClick` 0件。C19監査unit 53/53 PASS
- [x] known runtime NG — 会計詳細・在庫編集・見積詳細を回帰test先行で修正し、各4 viewportでP1/P7/P9/P10/RBACを再確認
- [x] F7b — 35適合 / 11根拠付き例外 / 18対象外 / 未裁定0。manual検索commit 130.9/116.4ms→3.4/3.0ms、入力2.7s→0.2s、診療tab INP 258ms→45ms、route遷移INP 32ms、CLS 0を記録
- [x] F4全route — `./scripts/run-e2e.sh e2e/ui-design-compliance-readonly.spec.ts --workers=1` は92/92 PASS（inventory 1 + 製品ページ83 + RBAC 3 + clinical P10 5、6.1m）。4 viewport、h1/path/overflow/a11y/44px/console/network/request ledgerを全件確認
- [x] F4安全証拠 — synthetic API behavior 2/2、カルテauto-create 1/1 PASS。same-origin・sentinel clinic・CSRF header、allowlist外non-GET abort、continued-to-backend 0
- [x] F4最終件数 — 完全確認83 / BLOCKED 0。targeted coverageは18 source / 17 test file / 138 test、statements 84.58%、lines 87.38%

| Group | ページ数 | 最終結果 | 根拠 |
|---|---:|---:|---|
| 認証・共通 | 4 | PASS 4 / BLOCKED 0 | public authの不要XHR 0、owner reportのP10を含む4 viewport |
| 会計・在庫・見積・シフト | 15 | PASS 15 / BLOCKED 0 | known NG 3ページ、RBAC synthetic fixture、read-only/non-GET 0 |
| 診療・入院・トリミング・予防 | 24 | PASS 24 / BLOCKED 0 | live fixture + typed sentinel fixture、P10、backend write 0 |
| 受付・飼主・予約・集計 | 6 | PASS 6 / BLOCKED 0 | production state + P10/RBAC regression test |
| Lステップ・LINE予約・マニュアル | 10 | PASS 10 / BLOCKED 0 | permission/API抑止、canonical path/query、Profiler計測 |
| 設定・マスタ | 24 | PASS 24 / BLOCKED 0 | RBAC/read-only、44px、accessible name、row/drag guard |
| **合計** | **83** | **PASS 83 / BLOCKED 0** | **404対象外1、redirect-only 12、別アプリ除外** |

安全上の記録: 過去に `/medical-records/new` の表示だけでdraftが作成されたため、今回の最終auditは予約・カルテPOSTを完全一致allowlistでPlaywright内fulfillし、その他のbusiness non-GETをnavigation前に遮断した。既存データの削除・新規DB fixture作成・DB変更は行っていない。

ページ監査の完了条件（**83ページすべてが `[x]`、例外は理由付きで `—`、`ui-design-compliance.md` と件数一致**）は達成した。最終frontend実装commit `1e7faf56e` 以降のscoped commitは0件で、`frontend/` のworktree/indexもcleanであることを照合済み。本書は完了済みチェックリスト兼監査証跡として保持し、削除対象としない。
