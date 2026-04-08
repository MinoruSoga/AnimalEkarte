# BUG-214: TypeScript strict モード無効

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | 型安全性 / CI |

## 現状コード

### `frontend/tsconfig.json:14-16`
```json
"strict": false,
"noUnusedLocals": false,
"noUnusedParameters": false
```

## 影響

- `any` 型が暗黙的に許容される
- 未使用の変数・パラメータが検出されない
- CI で型安全性違反が素通りする

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md`
> 型安全性最優先: Go/TypeScript 共に any を禁止し、厳格な型定義を行う

### `.claude/rules/typescript-react.md` §1
> any 禁止、型安全性優先

## 修正方針

段階的に有効化する：
1. まず `"noUnusedLocals": true`, `"noUnusedParameters": true` を有効化
2. 発生するエラーを修正
3. 最後に `"strict": true` を有効化
4. CI に `npx tsc --noEmit` ステップを追加
