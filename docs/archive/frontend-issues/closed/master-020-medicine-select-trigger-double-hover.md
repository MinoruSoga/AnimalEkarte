# [master] MedicineSettings: PropertyRow 内 SelectTrigger のホバー色が二重になる

## 優先度
低

## 種別
UIバグ / スタイル

## 対象ファイル
- `frontend/src/features/master/routes/MedicineSettings.tsx`

---

## 問題

薬剤マスタ編集フォームのサイドピーク（`MedicineSidePanel`）にて、`PropertyRow` 内の `SelectTrigger`（親カテゴリ・剤形・単位）にホバーすると、ホバー背景色が二重に表示される。

### 再現手順

1. 薬剤マスタ（`/settings/medicine`）を開く
2. 任意の薬剤行をクリックしてサイドピークを開く
3. 「親カテゴリ」「剤形」「単位」のいずれかのセレクトトリガーにマウスをホバーする

### 期待結果

PropertyRow 全体が `rgba(55,53,47,0.04)` で一様にハイライトされる。

### 実際の結果

PropertyRow 全体が `rgba(55,53,47,0.04)` でハイライトされ、さらにその上に SelectTrigger エリアが追加で `rgba(55,53,47,0.04)` を重ねるため、トリガー部分だけが実効 `rgba(55,53,47,0.08)` 相当の二重ハイライトになる。

---

## 原因

`MedicineSettings.tsx` で定義されている `SELECT_TRIGGER_FULL` 定数に `${C.hoverBgLight}` が含まれている。

```tsx
// MedicineSettings.tsx L.105
const SELECT_TRIGGER_FULL = `h-[30px] text-sm bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-full`;
//                                                                                  ^^^^^^^^^^^^^^^^
//                                                                                  = hover:bg-[rgba(55,53,47,0.04)]
```

`PropertyRow` コンポーネントも同じ `${C.hoverBgLight}`（= `hover:bg-[rgba(55,53,47,0.04)]`）を持つ：

```tsx
// PropertyRow.tsx
<div className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}>
```

SelectTrigger は PropertyRow の子要素であるため、ホバー時に：

1. **PropertyRow** → `rgba(55,53,47,0.04)` が行全体に適用
2. **SelectTrigger（子）** → 同じ `rgba(55,53,47,0.04)` が上書き適用

2 層が合成され、トリガー領域だけが約 2 倍の濃さでハイライトされる。

なお、`SelectTrigger`（`select.tsx`）のデフォルトホバーは `hover:bg-[rgba(242,241,238,0.5)]` だが、
`SELECT_TRIGGER_FULL` の `${C.hoverBgLight}` が `cn()`（tailwind-merge）によりこれを上書きしているため、
最終的なトリガーホバー色は `rgba(55,53,47,0.04)` となる。

---

## 修正方針

`SELECT_TRIGGER_FULL` から `${C.hoverBgLight}` を削除する。

```diff
- const SELECT_TRIGGER_FULL = `h-[30px] text-sm bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-full`;
+ const SELECT_TRIGGER_FULL = `h-[30px] text-sm bg-transparent ${C.text} border-0 px-1.5 shadow-none rounded-[3px] w-full`;
```

`PropertyRow` の行ホバーで十分であり、子の SelectTrigger が独自のホバー背景を持つ必要はない。

---

## 影響範囲

- `MedicineSettings.tsx` のみ（`SELECT_TRIGGER_FULL` はこのファイル内のモジュールレベル定数）
- 他のマスタページ（`DiagnosisSettings`・`TrimmingSettings` 等）は `SELECT_TRIGGER_FULL` を使っていないため影響なし
- `PropertyRow` 自体の変更は不要
