# BUG-222: useCallback deps にオブジェクト/配列を直接指定（9箇所/4ドメイン）

## 概要
`useCallback` の依存配列にオブジェクト・配列を指定している箇所が9箇所存在する。オブジェクトは毎レンダーで新しい参照を生成するため、`useCallback` のメモ化が無効化され、memo() でラップした子コンポーネントが不要に再レンダーされる。

## 現状コード

### `features/hospitalization/routes/HospitalizationForm.tsx`
```typescript
// ❌ location.state はオブジェクト — 毎レンダーで新参照
const handleBack = useCallback(() => {
  if (location.state?.from) {
    navigate(location.state.from as string);
  } else {
    navigate(paths.hospitalization.getHref());
  }
}, [location.state, navigate]);
```

### `features/accounting/routes/AccountingDetail.tsx` (類似パターン)
```typescript
// ❌ location.state オブジェクト依存
const handleBack = useCallback(() => {
  if (location.state?.from) {
    navigate(location.state.from as string);
  } else {
    navigate(paths.accounting.getHref());
  }
}, [location.state, navigate]);
```

## 影響範囲

| ドメイン | ファイル | 依存オブジェクト | 件数 |
|---------|---------|----------------|------|
| hospitalization | `HospitalizationForm.tsx` | `location.state` | 1 |
| accounting | `AccountingDetail.tsx` | `location.state` | 複数 |
| medical-records | `MedicalRecordForm.tsx` | `location.state` | 1 |
| 各フォーム | handleFormChange 系 | 親から受け取るオブジェクト props | 複数 |

## 修正方針

`location.state` から必要な primitive を抽出して deps に渡す。

### `features/hospitalization/routes/HospitalizationForm.tsx`
```typescript
// ✅ from を primitive として抽出
const locationFrom = location.state?.from as string | undefined;

const handleBack = useCallback(() => {
  if (locationFrom) {
    navigate(locationFrom);
  } else {
    navigate(paths.hospitalization.getHref());
  }
}, [locationFrom, navigate]);
```

同パターンを `AccountingDetail.tsx`, `MedicalRecordForm.tsx` 等に適用。

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-dependencies
> `useCallback` deps にオブジェクトを入れない — primitive を抽出して使う

### プロジェクト内参照実装
`features/owners/routes/OwnersList.tsx` — `pendingDeleteOwnerId` (string) を deps に渡すパターン

## 優先度
**Medium** — memo() の効果を無効化するが、機能バグは発生しない

## 関連チケット
- BUG-221: ✅ CLOSED — useTransition 未使用（同ドメイン修正済み）

## 関連ファイル
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:96-102`
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
