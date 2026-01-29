---
paths: "**/*.{ts,tsx,js,jsx,go}"
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
