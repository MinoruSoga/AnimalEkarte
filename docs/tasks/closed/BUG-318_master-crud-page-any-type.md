# BUG-318: MasterCRUDPage の TForm ジェネリクスに `any` 型 — 型安全性の欠如

## 概要

`frontend/src/features/master/components/MasterCRUDPage.tsx:154` の型キャスト部分で
`TForm = any` がデフォルト型引数として使用されており、`eslint-disable-next-line` コメントで
Lint を抑制している。フォームデータの型チェックが実質的に無効化されている状態。

## 再現手順

1. `frontend/src/features/master/components/MasterCRUDPage.tsx:154` を参照
2. `TForm = any` によりフォーム props の型チェックがゼロに
3. 誤ったフォームデータ型を渡してもコンパイルエラーが発生しない

## 期待する動作

- `TForm` に具体的な制約（例: `TForm extends Record<string, unknown>` 等）を設定する
- または `MasterCRUDPageProps` の型定義を見直し、`any` を除去する

## 現状コード

### `frontend/src/features/master/components/MasterCRUDPage.tsx:153-154`
```tsx
// eslint-disable-next-line @typescript-eslint/no-explicit-any
}) as <T extends MasterEntity, TForm = any>(props: MasterCRUDPageProps<T, TForm>) => ReactNode;
```

型キャストによる型アサーションと `eslint-disable` の組み合わせで、型安全性を完全に放棄している。

### `MasterCRUDPageProps` の TForm 利用状況の確認が必要
`TForm` がコンポーネント内でどのように利用されているかを調査し、
適切な型制約を設ける必要がある。

### 比較: 正しい実装パターン
```tsx
// ✅ ジェネリクスに制約を付ける
}) as <T extends MasterEntity, TForm extends Record<string, unknown>>(
  props: MasterCRUDPageProps<T, TForm>
) => ReactNode;

// または、TForm が汎用的である必要がある場合
}) as <T extends MasterEntity, TForm = Record<string, unknown>>(
  props: MasterCRUDPageProps<T, TForm>
) => ReactNode;
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/master/components/MasterCRUDPage.tsx:153-154` | `TForm = any` 型アサーション + eslint-disable | 未修正 |
| `MasterCRUDPage` を使用する全マスタページ | 型安全性が実質ゼロ | 影響あり |

## 修正方針

### 1. `MasterCRUDPageProps` の `TForm` 利用状況を調査

`TForm` がどのようなプロパティに渡されているかを確認：

```bash
grep -n "TForm" frontend/src/features/master/components/MasterCRUDPage.tsx
```

### 2. `TForm` に適切な制約を追加 — `MasterCRUDPage.tsx:153-154`

`TForm` が実際に使われているプロパティ（例: フォームコンポーネントの props）を確認し、
最低限の制約を付ける：

```tsx
// ✅ 案A: 最小限の制約（TForm が任意オブジェクトであることのみ保証）
}) as <T extends MasterEntity, TForm extends object>(
  props: MasterCRUDPageProps<T, TForm>
) => ReactNode;

// ✅ 案B: TForm が不要なら型パラメータ自体を削除
// MasterCRUDPageProps の TForm 参照箇所を確認の上、削除を検討
```

### 3. `eslint-disable` コメントを削除

`any` が除去されれば `eslint-disable` も不要になる。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 型安全性最優先
> **型安全性最優先**: Go/TypeScript 共に `any` を禁止し、厳格な型定義を行う。

### `.claude/rules/typescript-react.md` — any型禁止
> ```typescript
> // ❌ 禁止: any
> const handleChange = (e: any) => {};
> const data: any = response.data;
> ```
> 代替: `unknown` + 型ガード

### `.claude/rules/code-style.md`
> **Prohibited**: `any` type usage

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnerForm.tsx` — 厳格なジェネリクス使用例

## 優先度

**High** — `any` の使用は型安全性の完全な無効化であり、リファクタリング時のバグ混入リスクが高い。`eslint-disable` で意図的に抑制している点が特に問題。

## 関連チケット

なし

## 関連ファイル

- `frontend/src/features/master/components/MasterCRUDPage.tsx:153-154` — 問題箇所
