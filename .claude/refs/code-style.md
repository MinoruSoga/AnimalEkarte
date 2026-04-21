---
description: Naming conventions and code style standards
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

### Architecture (bulletproof-react compliant)
- Feature-based organization: most code in `src/features/`
- Unidirectional flow: `shared → features → app`
- No cross-feature direct imports (compose in app layer)
- No `export *` (blocks tree-shaking). Named exports OK
- Absolute imports: use `@/` alias

### Naming Conventions
- Variables/Functions: camelCase
- Components: PascalCase
- Constants: UPPER_SNAKE_CASE
- Component file (.tsx): PascalCase (`PatientCard.tsx`)
- Non-component file (.ts): kebab-case (`use-patient-form.ts`, `get-owners.ts`)
- Folders: kebab-case
- Types/Interfaces: PascalCase

### Import Order
1. React/Framework imports
2. External libraries
3. Internal shared (`@/components`, `@/hooks`, `@/lib`, `@/types`, `@/utils`)
4. Feature-internal (same feature only)
5. Type imports (`type` keyword)

### Styling & Design Tokens
- **MANDATORY**: All styling (Tailwind 4, inline styles) uses `src/lib/design-tokens.ts` (`C`, `STYLE`).
- **PROHIBITED**: Direct hex color (`#37352F`) specification.

```typescript
// ✅ Correct
import { C, STYLE } from '@/lib/design-tokens';
<div className={cn(STYLE.FLEX_CENTER, "p-4")} style={{ color: C.TEXT_MAIN }}>

// ❌ Prohibited
<div className="flex items-center justify-center p-4" style={{ color: '#37352F' }}>
```

### React 19 Patterns
- Components: function declaration (no `FC` type)
- `ref` as normal prop (no `forwardRef`)
- `useActionState` for form actions
- `useOptimistic` for optimistic UI updates
- `use()` for Promise/Context reading

### Prohibited
- `any` type usage
- Unused imports
- `console.log` in production code
- Hardcoded values (use env vars or constants)
- `FC` / `React.FC` type annotation
- `forwardRef` wrapper (React 19 uses ref as prop)
- Cross-feature imports (import via app layer)
- `export *` wildcard re-exports (blocks tree-shaking)
- `&&` for conditional render (use `? (...) : null`)
- **Deep feature imports**: `@/features/xxx/components/YYY` prohibited. Always via feature `index.ts` (Feature Indexing).

### Performance Rules (Vercel React Best Practices)

Reference: `features/owners/` — all patterns implemented.
Details: `frontend/CODING_RULES.md` Section 12

| Rule | Requirement |
|------|-------------|
| `rerender-memo` | Large independent sections in `memo()`. Handlers always `useCallback` stable |
| `rerender-functional-setstate` | In `useCallback`, setState as `prev =>` form (remove from deps) |
| `rerender-lazy-state-init` | High-cost useState init as `useState(() => ...)` lazy |
| `rerender-transitions` | Search filter: `useDeferredValue`, API write: `useTransition` |
| `rerender-dependencies` | No objects in `useCallback` deps — extract primitives |
| `rendering-hoist-jsx` | Static JSX (Select options) as module constant |
| `rendering-conditional-render` | Always `condition ? <X /> : null` (NOT `&&`) |
| `bundle-dynamic-imports` | Heavy modals/dialogs: `lazy()` + `Suspense` |
| `bundle-feature-indexing` | External feature usage: via feature `index.ts` |
| `async-parallel` | Independent fetches in loader: `Promise.all` / `Promise.allSettled` |
| `js-cache-function-results` | API-derived JSX lists: `useMemo([list])` cache |
