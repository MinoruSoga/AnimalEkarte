# FE-143: マスタ設定サイドパネルの名称フィールドで Enter キーが保存をトリガーしない

**Status**: Open
**Priority**: Low
**Affects**: features/master/ または features/settings/
**Date Created**: 2026-03-29
**Related**: BUG-048

---

## Summary

マスタ設定の SidePeek パネルにある名称テキストフィールドで Enter キーを押しても保存されない。
「保存」ボタンは機能する。キーボード UX の改善。

---

## 実装手順

### 1. 原因調査

SidePeek パネルのフォーム実装を確認：

```bash
grep -rn "SidePeek\|onKeyDown\|onKeyPress" frontend/src/features/
```

### 2. Enter キーで submit をトリガー

**方法 A: `<form>` タグを使う（推奨）**

```typescript
<form onSubmit={handleSubmit}>
  <Input
    value={formData.name}
    onChange={e => setFormData(prev => ({ ...prev, name: e.target.value }))}
  />
  <Button type="submit">保存</Button>
</form>
```

`<form>` 内の `<Button type="submit">` を使えば Enter キーが自動的に submit をトリガーする。

**方法 B: `onKeyDown` ハンドラ追加（form タグを使わない場合）**

```typescript
const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
  if (e.key === "Enter") {
    e.preventDefault();
    handleSubmit();
  }
}, [handleSubmit]);

<Input onKeyDown={handleKeyDown} ... />
```

### 3. 全マスタ設定サイドパネルへ適用

`/settings/medicine`, `/settings/animal-species` 等、SidePeek を持つ全マスタ設定ページに同様の修正を適用する。

---

## 受入条件

- [ ] 名称フィールドで Enter キーを押すと保存が実行される
- [ ] 全マスタ設定サイドパネル（medicine, animal-species 等）で同様に動作する
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
