---
description: 命名規則・コードスタイル規約
alwaysApply: true
globs: ["**/*.{ts,tsx,js,jsx,go}"]
---

# Code Style Rules

## Go (Backend)

### Naming Conventions
- Packages: lowercase (`handler`, `repository`)
- Exported: PascalCase (`GetPatient`, `PatientService`)
- Unexported: camelCase (`validateInput`, `dbConn`)
- Files: snake_case (`patient_handler.go`)

### Import Order
1. Standard library
2. External packages
3. Internal packages

### Prohibited
- Naked `panic` (use error returns)
- Ignoring errors (`_ = err`)
- Global mutable state
- Unused imports

### Required
- Error wrapping with context
- Context propagation
- Interface-based design

---

## TypeScript / React 19 (Frontend)

### アーキテクチャ（bulletproof-react準拠）
- Feature-based organization: コードの大部分は `src/features/` 内に配置
- 単方向コードフロー: `shared → features → app`
- Feature間の直接importは禁止（app層で合成する）
- `export *` 禁止（tree-shaking阻害）。明示的named exportは可
- 絶対パスimport: `@/` エイリアス使用

### Naming Conventions
- Variables/Functions: camelCase
- Components: PascalCase
- Constants: UPPER_SNAKE_CASE
- Files/Folders: kebab-case（ESLint check-fileで強制）
- Types/Interfaces: PascalCase

### Import Order
1. React/Framework imports
2. External libraries
3. Internal shared modules (`@/components`, `@/hooks`, `@/lib`, `@/types`, `@/utils`)
4. Feature-internal imports (同一feature内のみ)
5. Type imports (`type` keyword付き)

### React 19 Patterns
- コンポーネントは関数宣言で定義（`FC`型は使わない）
- `ref`は通常のpropとして渡す（`forwardRef`は不要）
- `useActionState`でフォームアクション管理
- `useOptimistic`で楽観的UI更新
- `use()`でPromise/Contextの直接読み取り

### Prohibited
- `any` type usage
- Unused imports
- `console.log` in production code
- Hardcoded values (use env vars or constants)
- `FC` / `React.FC` type annotation
- `forwardRef` wrapper（React 19ではref as prop）
- Cross-feature imports（feature間の直接import）
- `export *` wildcard re-exports（tree-shaking阻害）
- `&&` for conditional rendering（use `? (...) : null` instead）
- Barrel index imports（import directly from source files）

### Performance Rules（Vercel React Best Practices）

> 参照実装: `features/owners/` — すべてのパターンが実装済み。
> 詳細: `frontend/CODING_RULES.md` Section 12

| Rule | Requirement |
|------|-------------|
| `rerender-memo` | 独立した大きいセクションは `memo()` で囲む。必ず props ハンドラを `useCallback` で安定化すること。 |
| `rerender-functional-setstate` | `useCallback` 内の setState は `prev =>` 形式で state を deps から外す |
| `rerender-lazy-state-init` | 高コストな useState 初期化は `useState(() => ...)` lazy 形式 |
| `rerender-transitions` | 検索フィルタに `useDeferredValue`、API 書き込みに `useTransition` |
| `rerender-dependencies` | `useCallback` deps にオブジェクトを入れない — primitive を抽出して使う |
| `rendering-hoist-jsx` | コンポーネント外の静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ |
| `rendering-conditional-render` | 条件付きレンダリングは必ず `condition ? <X /> : null`（`&&` 禁止） |
| `bundle-dynamic-imports` | 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード |
| `bundle-barrel-imports` | feature api/utils は barrel index 経由でなく直接ファイルから import |
| `async-parallel` | loader 内の独立フェッチは `Promise.all` / `Promise.allSettled` で並列実行 |
| `js-cache-function-results` | API 由来の JSX リスト生成は `useMemo([list])` でキャッシュ |
