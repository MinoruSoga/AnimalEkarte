# FE-147: ペット追加モーダルがキャンセル後に状態をリセットしない

**Status**: Open
**Priority**: Medium
**Affects**: features/owners/components/（ペット追加モーダル）
**Date Created**: 2026-03-29
**Related**: BUG-052

---

## Summary

飼主登録フォームのペット追加モーダルでキャンセル後、再度開くと前回入力値が残留する。
キャンセルハンドラでフォーム state をリセットしていない。

---

## 実装手順

### 1. 原因調査

```bash
grep -rn "PetEditModal\|cancelPet\|handleCancel\|onCancel" frontend/src/features/owners/
```

モーダルが `display: none` 制御（アンマウントなし）なら state は保持されたまま。
キャンセル時のハンドラを確認する。

### 2. キャンセル時にフォームをリセット

**方法 A: キャンセルハンドラでリセット（推奨）**

```typescript
const handleCancel = useCallback(() => {
  resetPetForm();   // 初期値に戻す
  onClose();
}, [resetPetForm, onClose]);
```

`resetPetForm` は `useOwnerForm.ts` の `usePetForm` hook 内で実装する：

```typescript
const resetPetForm = useCallback(() => {
  setPetFormData(initialPetValues);
}, []);
```

**方法 B: モーダルに `key` を渡して毎回アンマウント**

```tsx
<PetEditModal
  key={isOpen ? "open" : "closed"}  // 開閉のたびにアンマウント
  isOpen={isOpen}
  ...
/>
```

方法 B はシンプルだが、毎回マウントコストが発生する。方法 A を推奨。

### 3. 適用対象

- `/owners/new` のペット追加モーダル
- `/owners/:id` のペット追加モーダル（同様の問題がある可能性）

---

## 受入条件

- [ ] キャンセル後に再度「ペット追加」を開くと全フィールドが初期値になっている
- [ ] `/owners/new` と `/owners/:id` の両方で動作する
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
