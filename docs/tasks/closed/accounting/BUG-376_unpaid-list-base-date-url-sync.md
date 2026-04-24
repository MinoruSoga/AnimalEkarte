# BUG-376: 未納者一覧の基準日が URL クエリと同期されない

**作成日**: 2026-04-14
**Status**: Closed (2026-04-14)
**Priority**: Low（UX 改善）
**Affects**: `features/accounting` — `UnpaidCustomerList.tsx`

**発見経緯**: BUG-374 ブラウザテスト TC-370-11 で CONFIRMED

---

## 概要

未納者一覧画面で **基準日** を変更した場合、`group_by` と異なり URL クエリパラメータ (`reference_date` 等) に同期されない。
ブラウザリロードやブックマーク遷移で基準日が今日にリセットされる。

## 現状のコード

```typescript
// frontend/src/features/accounting/routes/UnpaidCustomerList.tsx:44-60
const groupBy: GroupBy = (searchParams.get("group_by") as GroupBy) === "billing" ? "billing" : "owner";
const [baseDate, setBaseDate] = useState<string>(todayISO());
```

`groupBy` は URL 同期、`baseDate` は state のみ。

## 必要な変更

`baseDate` を URL クエリ `reference_date=YYYY-MM-DD` と同期。

```typescript
const groupBy: GroupBy = (searchParams.get("group_by") as GroupBy) === "billing" ? "billing" : "owner";
const baseDate = searchParams.get("reference_date") || todayISO();

const handleBaseDateChange = useCallback((next: string) => {
  searchParams.set("reference_date", next);
  setSearchParams(searchParams, { replace: true });
  setPage(1);
}, [searchParams, setSearchParams]);
```

Input onChange → `handleBaseDateChange(e.target.value)` に変更。

## 受入条件

- [ ] AC-1: 基準日を変更すると URL に `?reference_date=YYYY-MM-DD` が付与される
- [ ] AC-2: ブラウザリロード後も基準日が復元される
- [ ] AC-3: ブックマーク URL から開くと指定基準日で表示される
- [ ] AC-4: `group_by` と併用できる (`?group_by=billing&reference_date=2026-04-01`)

## 影響範囲

### Frontend
- `frontend/src/features/accounting/routes/UnpaidCustomerList.tsx` — state → searchParams 経由に変更

### Backend
- 変更なし

## 関連

- BUG-370: 月末未納者一覧（元機能）
- BUG-374: ブラウザテスト（本件の発見源）
