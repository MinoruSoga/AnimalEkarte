# FE-081: Vercel React Best Practices 品質改善 (Low Priority)

## 背景

全 `src/` フォルダの Vercel React Best Practices 監査で検出された、品質改善レベルの違反3件。
機能に影響はないが、パフォーマンス最適化とコード品質の観点で修正すべき項目。

## 依存

- FE-079, FE-080 完了後に着手

## 要件

### 1. `features/medical-records/components/CheckupsTab/CheckupsTab.tsx` — useMemo 追加

**違反ルール**: `js-cache-function-results`

`checkupTypes.map()` が JSX 内でインライン実行されており、レンダーのたびに再生成される。
`useMemo` で包んでキャッシュする。

```typescript
// before (行101-105)
<select>
  {checkupTypes.map((type: CheckupTypeItem) => (
    <option key={type.id} value={type.id}>{type.name}</option>
  ))}
</select>

// after
const checkupTypeOptions = useMemo(() =>
  checkupTypes.map((type: CheckupTypeItem) => (
    <option key={type.id} value={type.id}>{type.name}</option>
  )),
  [checkupTypes]
);

// JSX 内
<select>{checkupTypeOptions}</select>
```

### 2. `features/accounting/routes/AccountingDetail.tsx` — AccountingDocument の lazy loading 検討

**違反ルール**: `bundle-dynamic-imports`

`AccountingDocument` コンポーネントが同期 import されている。
コンポーネントのサイズを確認し、100行以上であれば `lazy()` + `Suspense` で遅延ロードする。

```typescript
// before
import { AccountingDocument } from "../components/AccountingDocument";

// after (コンポーネントが大きい場合)
const AccountingDocument = lazy(() =>
  import("../components/AccountingDocument").then(m => ({ default: m.AccountingDocument }))
);

// Dialog 内で Suspense で包む
<Suspense fallback={<div className="p-8 text-center">読み込み中...</div>}>
  <AccountingDocument ... />
</Suspense>
```

**判断基準**: `AccountingDocument.tsx` が100行未満であればスキップ可。

### 3. `features/hospitalization/components/DailyRecord/VitalDialog.tsx` — prevOpen アンチパターン修正

**違反ルール**: `rerender-derived-state-no-effect`

`prevOpen` state を使ってダイアログ開閉時のフォーム初期化を検出している。
これはアンチパターンであり、`key` prop または `useEffect` に変更する。

```typescript
// before (行37)
const [prevOpen, setPrevOpen] = useState(false);
// render 内で if (open !== prevOpen) { setPrevOpen(open); resetForm(); }

// after: key prop パターン（推奨）
// 親コンポーネントで:
<VitalDialog key={dialogOpenKey} open={open} ... />
// dialogOpenKey を open/close 時にインクリメントすることで、
// コンポーネントが再マウントされフォームが自動リセットされる

// あるいは useEffect パターン:
useEffect(() => {
  if (open) {
    resetForm();
  }
}, [open]);
```

## 受入条件

- [ ] `CheckupsTab.tsx` の `checkupTypes.map()` が `useMemo` で包まれている
- [ ] `AccountingDocument` のサイズを確認し、適切に判断（100行以上なら lazy、未満ならスキップ可）
- [ ] `VitalDialog.tsx` の `prevOpen` パターンが除去され、`key` prop または `useEffect` に変更されている
- [ ] `docker compose exec frontend npm run build` が成功
