# BUG-232: EstimateForm の保存ボタンが brand blue ではなく Notion gray（C.bgPrimary）

## 概要

`features/estimates/routes/EstimateForm.tsx:301` の `SubmitButton` に `${C.bgPrimary}` (`bg-[#37352F]`、Notion dark gray) が適用されている。プロジェクトの primary action button は `${C.bgAccent}` (`bg-[#2383E2]`、brand blue) が標準であり、`STYLE.confirmPrimary` もこれを使用している。見積フォームの「作成/更新」ボタンだけが gray で、他のフォーム（カルテ・入院・ワクチン等）と異なる色になっている。

## 再現手順

1. 見積フォーム（`/estimates/new` または `/estimates/[id]`）を開く
2. ヘッダー右上の「作成」/「更新」ボタンを確認する
3. **結果**: 暗いグレー背景のボタンが表示される（Notion の dark gray #37352F）
4. 他フォーム（例: カルテフォーム `/medical-records/[id]`）の同位置ボタンと比較する
5. **期待**: 他フォームと同じ brand blue (#2383E2) のボタンが表示されるべき

## 期待する動作

- 見積フォームの保存ボタンも他のフォームと同じ brand blue (`C.bgAccent`) を使用すること

## 現状コード

### `features/estimates/routes/EstimateForm.tsx:298-304`
```tsx
// ❌ C.bgPrimary = bg-[#37352F] (Notion dark gray) — 誤ったカラートークン
{canSubmit ? (
  <SubmitButton
    size="sm"
    disabled={!form.title.trim()}
    className={`h-9 ${C.bgPrimary} ${C.hoverBgPrimaryDark} text-white text-sm`}
  >
    {isEdit ? '更新' : '作成'}
  </SubmitButton>
) : null}
```

### デザイントークン確認

```typescript
// src/lib/design-tokens.ts
C.bgPrimary  = "bg-[#37352F]"         // ← Notion document primary gray (使用誤り)
C.bgAccent   = "bg-[#2383E2]"         // ← brand blue (正しい primary action 色)

// STYLE.confirmPrimary (SubmitButton のデフォルト内部スタイル)
confirmPrimary: `${C.bgAccent} text-white ${C.bgAccentHover} h-11 px-4 ...`
```

### 比較: 正しい実装（他フォーム）

```tsx
// ✅ 正しい: C.bgAccent (brand blue) を使用
// features/vaccinations/routes/VaccinationForm.tsx:173
<SubmitButton className={`${C.bgAccent} ${C.bgAccentHover} text-white shadow-sm px-6 h-10 text-sm`}>
  {hospitalizationId ? "更新" : "登録"}
</SubmitButton>

// ✅ さらに正しい（SubmitButton デフォルトスタイルに任せる）:
// features/owners/routes/OwnerForm.tsx:583
<SubmitButton size="sm">
  {isEdit ? "更新" : "登録"}
</SubmitButton>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `features/estimates/routes/EstimateForm.tsx:301` | `C.bgPrimary` (`bg-[#37352F]`) を SubmitButton に適用 | 未修正 |

## 修正方針

### `features/estimates/routes/EstimateForm.tsx:298-304`

```tsx
// Before
<SubmitButton
  size="sm"
  disabled={!form.title.trim()}
  className={`h-9 ${C.bgPrimary} ${C.hoverBgPrimaryDark} text-white text-sm`}
>
  {isEdit ? '更新' : '作成'}
</SubmitButton>

// After — C.bgPrimary → C.bgAccent, C.hoverBgPrimaryDark → C.bgAccentHover
<SubmitButton
  size="sm"
  disabled={!form.title.trim()}
  className={`h-9 ${C.bgAccent} ${C.bgAccentHover} text-white text-sm`}
>
  {isEdit ? '更新' : '作成'}
</SubmitButton>
```

または、SubmitButton のデフォルトスタイルに任せて className を削除するのがより望ましい：

```tsx
// Best: className 不要（STYLE.confirmPrimary が C.bgAccent を含む）
<SubmitButton size="sm" disabled={!form.title.trim()}>
  {isEdit ? '更新' : '作成'}
</SubmitButton>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Submit Button
> 送信ボタンは必ず **`SubmitButton`** を使用（二重送信防止・自動ローディング）

SubmitButton のデフォルトスタイル (`STYLE.confirmPrimary`) は `C.bgAccent` を使用しており、カスタム className で上書きする場合も `C.bgAccent` を使うべき。

### `.claude/rules/typescript-react.md` — React 19 Patterns
> `useActionState` でフォームアクション管理

フォームの primary action button は brand blue で統一するのがプロジェクト標準。`C.bgPrimary` は Notion 的な dark document color であり、action button には使わない。

### プロジェクト内参照実装
- `features/owners/routes/OwnerForm.tsx:583` — `<SubmitButton size="sm">` デフォルトスタイルのみ
- `features/vaccinations/routes/VaccinationForm.tsx:173` — `C.bgAccent` 使用
- `src/lib/design-tokens.ts:856` — `STYLE.confirmPrimary` 定義（`C.bgAccent` ベース）

## 優先度
**Medium** — 見積フォームの保存ボタンが他全フォームと異なる色。機能的問題はないが、一貫したブランドカラーの観点で修正が必要。

## 関連チケット
- BUG-231: ghost-danger ボタンバリアントの色不一致（同種の色一貫性問題）

## 関連ファイル
- `frontend/src/features/estimates/routes/EstimateForm.tsx:298-304`
- `frontend/src/lib/design-tokens.ts:244,281,748,856`
