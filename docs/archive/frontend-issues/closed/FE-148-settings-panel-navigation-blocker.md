# FE-148: マスタ設定サイドパネルに NavigationBlocker が未実装

**Status**: Open
**Priority**: Low
**Affects**: features/master/ または features/settings/（SidePeek パネル）
**Date Created**: 2026-03-29
**Related**: BUG-053

---

## Summary

マスタ設定サイドパネルでフィールドを変更後にナビゲーションしても警告ダイアログが表示されず、未保存変更が失われる。
飼主登録フォームには `useBlocker` による NavigationBlocker が実装済みだが、設定サイドパネルには未実装。

---

## 実装手順

### 1. isDirty 状態を追跡

```typescript
const [isDirty, setIsDirty] = useState(false);

const handleChange = useCallback((field: string, value: unknown) => {
  setFormData(prev => ({ ...prev, [field]: value }));
  setIsDirty(true);
}, []);

const handleSaveSuccess = useCallback(() => {
  setIsDirty(false);
}, []);
```

### 2. `useBlocker` で未保存変更をブロック

```typescript
import { useBlocker } from "react-router-dom";

const blocker = useBlocker(
  ({ currentLocation, nextLocation }) =>
    isDirty && currentLocation.pathname !== nextLocation.pathname
);
```

### 3. 確認ダイアログを表示

```tsx
{blocker.state === "blocked" ? (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>変更が保存されていません</AlertDialogTitle>
        <AlertDialogDescription>
          保存されていない変更があります。ページを離れますか？
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel onClick={() => blocker.reset?.()}>
          キャンセル
        </AlertDialogCancel>
        <AlertDialogAction onClick={() => blocker.proceed?.()}>
          離れる
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
) : null}
```

### 4. 既存実装の参照

飼主登録フォームの `NavigationBlocker` 実装を参照：

```bash
grep -rn "useBlocker\|NavigationBlocker" frontend/src/features/owners/
```

同パターンを設定サイドパネルに適用する。

### 5. 適用対象

`/settings/clinic`, `/settings/animal-species`, `/settings/medicine` 等、
SidePeek を持つ全マスタ設定ページ。

---

## 受入条件

- [ ] 未保存変更ありでナビゲーション時に「変更が保存されていません。ページを離れますか？」ダイアログが表示される
- [ ] 「離れる」で遷移、「キャンセル」でサイドパネルに留まる
- [ ] 保存後はダイアログが表示されない（`isDirty = false`）
- [ ] 全マスタ設定サイドパネルで動作する
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
