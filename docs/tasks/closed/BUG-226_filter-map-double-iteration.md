# BUG-226: `.filter().map()` による二重イテレーション（4箇所）

## 概要
`.filter()` + `.map()` をチェーンすることで同一配列を2回走査している箇所が4件ある。`flatMap` または `reduce` に統合することで1回のイテレーションに削減できる。

## 現状コード

```typescript
// ❌ 2回イテレーション
const activeItems = items
  .filter(item => item.isActive)
  .map(item => ({ id: item.id, label: item.name }));
```

## 修正方針

`reduce` で1パスに統合する。

```typescript
// ✅ 1回のイテレーション
const activeItems = items.reduce<Array<{ id: string; label: string }>>(
  (acc, item) => {
    if (item.isActive) {
      acc.push({ id: item.id, label: item.name });
    }
    return acc;
  },
  []
);
```

または `flatMap` を使う場合：
```typescript
const activeItems = items.flatMap(item =>
  item.isActive ? [{ id: item.id, label: item.name }] : []
);
```

## 影響範囲

| ドメイン | 推定件数 |
|---------|---------|
| 複数ドメインに散在 | 4箇所 |

## 準拠すべきプロジェクト規約

### `frontend/CODING_RULES.md` Section 12 — js-combine-iterations
> `.filter().map()` による二重イテレーションは `flatMap` または `reduce` に統合する

## 優先度
**Low** — 配列が大きくない限りパフォーマンス差は軽微。コードの一貫性改善。

## 関連ファイル
- 散在（grep: `\.filter(.*)\s*\.\s*map(` で確認）
