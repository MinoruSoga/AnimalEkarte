# BUG-276: VaccinationForm — useMemo の依存配列にオブジェクト参照を渡している（rerender-dependencies 違反）

## 概要
`VaccinationForm.tsx` の `petHistory` useMemo が `historyFilter` オブジェクトを依存配列に含めている。
`historyFilter` は `useVaccinationForm` フックの `return` 文で毎レンダーごとにオブジェクトリテラルとして生成されるため、
内部の値が変わっていなくてもオブジェクト参照が変わり、useMemo が毎回再計算される。

## 違反ルール
`rerender-dependencies` — useCallback/useMemo の deps にはオブジェクトでなく primitive を渡す

## 箇所
- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:158`

## 原因
```typescript
// use-vaccination-form.ts:329-339 — 毎レンダーごとに新しいオブジェクトが生成される
return {
  historyFilter: {
    filterStartDate, setFilterStartDate,
    filterEndDate, setFilterEndDate,
    historySearchTerm, setHistorySearchTerm,
    sortOrder, setSortOrder,
    handleClearHistoryFilter,
  },
  ...
};

// VaccinationForm.tsx:158 — そのオブジェクトを deps に入れている
}, [allVaccinations, selectedPet, id, historyFilter]); // ← 毎回参照が変わる
```

## 修正方針
`historyFilter` をデストラクチャリングして各 primitive フィールドを直接 deps に渡す。

## 修正内容

### `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`

```typescript
// BEFORE (line 158)
}, [allVaccinations, selectedPet, id, historyFilter]);

// AFTER
const { historySearchTerm, filterStartDate, filterEndDate, sortOrder } = historyFilter;
// useMemo の deps を primitive に変更
}, [allVaccinations, selectedPet, id, historySearchTerm, filterStartDate, filterEndDate, sortOrder]);
```

## 優先度
MEDIUM — 不要な再計算によるパフォーマンス劣化。機能的バグではない。
