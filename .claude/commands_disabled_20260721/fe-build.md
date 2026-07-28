---
description: TypeScript/React ビルドエラー・型エラーの解決（Docker経由）
argument-hint: [--lint | --build | blank for type-check only]
---

# /fe-build

TypeScript 型エラーを診断し `build-error-resolver` エージェントで最小差分修正します。

## 使用法

```bash
/fe-build           # tsc --noEmit のエラーのみ解決
/fe-build --lint    # tsc + pnpm lint まで解決
/fe-build --build   # tsc + lint + pnpm build まで解決
```

## Step 1: エラー収集

以下を **ユーザーに実行してもらい**、出力を貼り付けてもらう:

```bash
docker compose exec frontend npx tsc --noEmit
```

`$ARGUMENTS` が `--lint` または `--build` の場合は追加で:

```bash
docker compose exec frontend pnpm lint
# --build の場合
docker compose exec frontend pnpm build
```

## Step 2: エラー解析

エラー出力が提供された場合、`build-error-resolver` エージェントを起動して解決する。

解決フロー:
1. 型エラーからファイル・行番号を特定
2. 該当ファイルを読んでコンテキスト理解
3. 最小差分で修正（アーキテクチャ変更禁止）
4. 修正後コマンドをユーザーに提示して再確認依頼

## 注意事項

- **Docker 必須**: `tsc` / `pnpm` はローカル実行禁止
- **`any` 禁止**: 型エラーを `any` で黙らせない。`unknown` + 型ガードを使う
- **Feature Indexing**: `@/features/xxx/components/Yyy` の deep import は `@/features/xxx` 経由に修正
- **React 19 パターン**: `FC` / `React.FC` / `forwardRef` は使わない
- **デザイントークン**: `#xxx` ハードコードは `C` / `STYLE` 定数に変更
- **同一エラー3回継続**: 解決不能と判断し報告する

## 出力形式

```
[FIXED] frontend/src/features/owner/components/OwnerForm.tsx:42
エラー: Type 'string | undefined' is not assignable to type 'string'
修正: optional chaining + nullish coalescing を追加

ビルドステータス: SUCCESS/FAILED | 修正エラー数: N | 変更ファイル: リスト
```

## 使用エージェント

`build-error-resolver` (Sonnet)
