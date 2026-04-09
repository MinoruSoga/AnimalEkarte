# BUG-267: ShiftFormDialog の formAction useCallback deps にオブジェクト/配列を指定

## 概要
`ShiftFormDialog` の `formAction` を包む `useCallback` の deps 配列に `editShift`（Shift オブジェクト）と `breaks`（ShiftBreakInput 配列）をそのまま指定している。両者は参照比較されるため、内容が変わらなくても参照が変わるたびに `formAction` が再生成される。`editShift.id`（primitive）と `breaksRef`（ref パターン）に置き換えることで不要な再生成を排除できる。

## 現状コード

### `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:86-129`
```typescript
const formAction = useCallback(
  async (_prevState: FormActionState, formData: FormData): Promise<FormActionState> => {
    // ...
    if (isEdit && editShift) {
      const validBreaks = breaks.filter((b) => b.break_start && b.break_end);
      await updateShift(editShift.id, input);  // ← editShift.id のみ使用
    } else {
      const validBreaks = breaks.filter((b) => b.break_start && b.break_end);  // ← breaks 全体を読み取り
      // ...
    }
    // ...
  },
  [isEdit, editShift, staffId, date, shiftType, breaks, queryClient, onClose],
  //         ^^^^^^^^ オブジェクト                 ^^^^^^ 配列
);
```

**問題**:
- `editShift` は Shift オブジェクト。実際に使うのは `editShift.id` と存在チェックのみ
- `breaks` は ShiftBreakInput[] 配列。ref 経由で読めば deps から外せる

## 修正方針

### `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx`

```typescript
// breaks を ref で参照し、deps から除外
const breaksRef = useRef(breaks);
useEffect(() => { breaksRef.current = breaks; }, [breaks]);

const editShiftId = editShift?.id;  // primitive

const formAction = useCallback(
  async (_prevState: FormActionState, formData: FormData): Promise<FormActionState> => {
    // ...
    if (isEdit && editShiftId) {
      const validBreaks = breaksRef.current.filter((b) => b.break_start && b.break_end);
      await updateShift(editShiftId, input);
    } else {
      const validBreaks = breaksRef.current.filter((b) => b.break_start && b.break_end);
      // ...
    }
    // ...
  },
  // rerender-dependencies: editShift → id（primitive）、breaks → ref 経由
  [isEdit, editShiftId, staffId, date, shiftType, queryClient, onClose],
);
```

## 影響範囲

| ファイル | 行番号 | 内容 |
|---------|-------|------|
| `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` | 128 | formAction useCallback deps |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — rerender-dependencies
> `useCallback` deps にはオブジェクトを入れない — primitive を抽出して使う

### プロジェクト内参照実装
- `features/reception/routes/Reception.tsx:92` — `columnsRef` で columnsオブジェクト配列を ref 経由参照
- `features/owners/routes/OwnersList.tsx` — `pendingDeleteOwnerId`（primitive id）を deps に

## 優先度
**Low** — `formAction` は `useActionState` に渡されるが、`dispatchFormAction` は安定参照のため子コンポーネントへの伝播なし。実害は小さいがベストプラクティス準拠の観点から修正すべき。

## 関連チケット
- BUG-222: useCallback deps にオブジェクト/配列（同一パターン、hospitalization/estimate/trimming）

## 関連ファイル
- `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:86-129` — 違反箇所
