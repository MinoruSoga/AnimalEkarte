# FE refactor — 未完了タスク

> 現在の判定: **INCOMPLETE**
>
> このファイルは未完了項目だけを置く active ledger とする。対応済みの履歴・完了チェック・過去の集計は残さない。
> 項目は完了証拠を確認した時点で削除し、完了markへ変更しない。ファイル自体は削除しない。

## 1. 対象と正本

- 色: [`docs/spec/design-system.md`](docs/spec/design-system.md)
- タイポグラフィ、形状、余白、エレベーション、コンポーネント寸法: [`DESIGN.md`](DESIGN.md)
- 実行時トークン: `frontend/src/lib/design-tokens.ts`
- 製品route台帳: [`docs/spec/ui-design-compliance.md`](docs/spec/ui-design-compliance.md)
- 対象: `frontend/src` の本体アプリ83製品ページ
- 対象外: 404 fallback、redirect-only route、印刷帳票、`ManualContent` のMarkdown本文、LIFF/LINE予約の別アプリ

`design-audit` の成功だけでP1〜P10の完了とは判定しない。静的監査、表示状態ごとのブラウザ検証、意味レビュー、
RBAC/臨床安全、network/console、実効寸法をそれぞれ確認する。

## 2. FE-OPEN-1 — synthetic API interceptor とE2Eを修正する

### 現在の未解決事象

`frontend/e2e/helpers/synthetic-api.ts` は `expectedOrigin` をAPI以外のrequestにも適用する。
そのため、`frontend/index.html` が意図的に読むGoogle FontsのCSSを
`GET:/css2:unexpected API origin` として遮断し、clinical routeとカルテ作成E2Eが失敗する。
既存helper testも「cross-origin non-API GETを遮断する」挙動を正として固定している。

- [ ] cross-origin API、許可済みasset、未知のcross-origin assetを分離するRED testを追加する
- [ ] `expectedOrigin`、clinic header、CSRF header、business non-GET allowlistをAPI境界へ限定してfail-closedを維持する
- [ ] Google Fonts等の外部assetは明示allowlist + local stubで決定的に処理し、任意の外部通信を許可しない
- [ ] request ledgerの全attemptを排他的な終端状態へ分類し、business non-GETのbackend到達を0にする
- [ ] console errorの記録へmessage本文と発生元を含め、`console:error` だけで原因を失わないようにする
- [ ] `e2e/helpers/synthetic-api.spec.ts` と `e2e/medical-records-create.spec.ts` を通す
- [ ] clinical P10の5 scenarioを4 viewportで通す

完了確認:

```bash
cd frontend
./scripts/run-e2e.sh \
  e2e/helpers/synthetic-api.spec.ts \
  e2e/medical-records-create.spec.ts \
  --workers=1
./scripts/run-e2e.sh \
  e2e/ui-design-compliance-readonly.spec.ts \
  --workers=1 \
  --grep "clinical P10"
```

## 3. FE-OPEN-2 — raw tableのP7準拠とC18監査漏れを解消する

### 現在の未解決事象

`DESIGN.md` の `ex-data-table-cell` と `STYLE.tableHeaderCell` / `STYLE.tableCell` は、
headerをcanvas-soft + eyebrow、
bodyをbody-sm、cell paddingを12px 16pxと定義する。一方、productionのraw `<th>/<td>` には
`px-3 py-2`、`font-medium`、`text-sm`、`px-3 py-1` 等の独自指定が残る。
`frontend/scripts/design-system-audit.mjs` のC18は `TableHead` / `TableCell` だけを検査し、
raw `<th>/<td>` を検出しない。

初期調査対象:

- `frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx`
- `frontend/src/features/accounting-reports/components/DailyBreakdownTable.tsx`
- `frontend/src/features/accounting/components/ItemListCard.tsx`
- `frontend/src/features/accounting/components/RefundSection.tsx`
- `frontend/src/features/cash-register/components/BillingDetailTable.tsx`
- `frontend/src/features/cash-register/components/UnifiedClosingSummaryTable.tsx`
- `frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx`
- `frontend/src/features/lstep/components/LstepCsvImportSection.tsx`
- `frontend/src/features/lstep/components/LstepDeliveryStatsSection.tsx`
- `frontend/src/features/lstep/components/LstepVisitConversionSection.tsx`

- [ ] productionのraw `<table>/<th>/<td>` を全件inventoryし、準拠・根拠付き例外・修正対象へ分類する
- [ ] 修正対象をtable tokenまたは共通primitiveへ移行し、header/body/padding/background/borderを正本へ合わせる
- [ ] printと`ManualContent`以外の例外は、製品上の理由と検証方法を明記する
- [ ] C18または新規ruleをraw `<th>/<td>` に拡張し、classNameのliteral・template・`cn()`形を検査する
- [ ] raw tableの違反・許可例外・誤検知を固定するaudit unit testをREDから追加する
- [ ] representative tableでcomputed styleを検証し、静的文字列の一致だけで完了扱いしない
- [ ] production raw tableの未分類・未準拠を0件にする

完了確認:

```bash
docker compose exec -T frontend node --test scripts/design-system-audit.test.mjs
docker compose exec -T frontend node scripts/design-system-audit.mjs
```

## 4. FE-OPEN-3 — 非表示interactive stateのP7/P9監査を完了する

### 現在の未解決事象

`OwnerSearchModal` の検索inputは共通`Input`の`h-11`を`h-10`で上書きするため実効40pxとなり、
明示的なaccessible nameもない。`text-base`とcanvas-soft surfaceも、入力の正本であるbody-smとwhite surfaceに
一致しない。現在の監査は一部tab等を操作するが、modal、drawer、popover、tab、条件分岐内のcontrolを
体系的に列挙・到達していない。

- [ ] `OwnerSearchModal` の検索inputのname、44px、surface、typographyを固定するRED testを追加する
- [ ] `OwnerSearchModal` の検索inputをP7/P9の正本へ合わせ、buttonを含むcomputed styleと実効寸法をbrowserで確認する
- [ ] 83 routeから到達するmodal、drawer、popover、tab、accordion、条件分岐stateをinventoryする
- [ ] 各stateを実際に開くPlaywright scenarioを用意し、accessible name、44×44px、focus、overflowを検査する
- [ ] 監査selectorを実際のfocusable要素とARIA widgetへ拡張し、非buttonの操作要素を見落とさないよう回帰testを追加する
- [ ] RBAC非活性、disabled、danger、死亡、期限切れ等の非色依存表現を該当stateで確認する
- [ ] 未到達のinteractive stateを0件にするか、安全上再現できない理由と代替testを記録する

## 5. FE-OPEN-4 — 最終ゲートを現行treeで成立させる

- [ ] 全83製品routeを1440 / 1200 / 800 / 500pxで描画し、path/h1、overflow `=== 0`、network、console、RBAC、P10を確認する
- [ ] P1〜P10は自動監査だけでなく、raw tableと非表示stateを含む意味・computed-style・操作検証で確定する
- [ ] synthetic routeで許可外business non-GET、continued-to-backend、DB変更をすべて0にする
- [ ] design audit、audit unit、変更隣接Vitest、E2E TypeScript、scoped ESLintを通す
- [ ] 変更対象sourceのstatementsとlines coverageを80%以上にし、skip追加や期待値弱体化を行わない
- [ ] `docs/spec/ui-design-compliance.md` を現行の実測値へ同期し、未成立の83/83・COMPLETE表現を残さない
- [ ] TypeScript、React/a11y、clinical/RBAC、security、silent failureの独立reviewでCRITICAL/HIGHを0にする
- [ ] `git diff --check` と対象外WIPの前後一致を確認する

最終browser gate:

```bash
cd frontend
./scripts/run-e2e.sh e2e/ui-design-compliance-readonly.spec.ts --workers=1
```

期待値は既存92件と追加した回帰testの全件PASSであり、途中停止、未実行、retry依存、console/network errorを
完了として数えない。最後の未完了項目を削除した後も、このファイルは削除せず、
タイトルと「未完了タスクなし」の状態だけを残す。
