# BUG-227: 静的 SelectItem JSX がモジュール定数に未巻き上げ（4箇所）

## 概要
`<SelectItem>` 等の静的JSXをコンポーネント関数内で生成している箇所が4件ある。これらは毎レンダーで新しいReact要素を生成し、不要なGC圧力をかける。モジュールレベルの定数に巻き上げることで、参照を安定化できる。

## 現状コード

```typescript
// ❌ コンポーネント内でレンダーごとに生成
function MyForm() {
  return (
    <Select>
      <SelectItem value="male">オス</SelectItem>
      <SelectItem value="female">メス</SelectItem>
      <SelectItem value="unknown">不明</SelectItem>
    </Select>
  );
}
```

## 修正方針

モジュールスコープの定数に巻き上げる。

```typescript
// ✅ モジュールスコープに定義（一度だけ生成）
const GENDER_SELECT_ITEMS = (
  <>
    <SelectItem value="male">オス</SelectItem>
    <SelectItem value="female">メス</SelectItem>
    <SelectItem value="unknown">不明</SelectItem>
  </>
);

function MyForm() {
  return (
    <Select>
      {GENDER_SELECT_ITEMS}
    </Select>
  );
}
```

## 影響範囲

| 推定 | 件数 |
|------|------|
| 4ファイル（4ドメイン） | 各1箇所 |

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rendering-hoist-jsx
> コンポーネント外の静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ

### プロジェクト内参照実装
`features/owners/components/PetEditModal.tsx` — `GENDER_SELECT_ITEMS` 定数パターン
`features/owners/routes/OwnerForm.tsx` — `PET_TABLE_HEADER` 定数パターン

## 優先度
**Low** — GC 圧力の軽減。機能的影響なし。修正は30分（4ファイル）。

## 関連ファイル
- 4ドメインに散在（grep: `<SelectItem` が関数内にある箇所）
