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

## TypeScript (Frontend)

### Naming Conventions
- Variables/Functions: camelCase
- Components: PascalCase
- Constants: UPPER_SNAKE_CASE
- Files: kebab-case

### Import Order
1. React/Framework imports
2. External libraries
3. Internal modules (@/)
4. Relative imports
5. Type imports

### Prohibited
- `any` type usage
- Unused imports
- `console.log` in production code
- Hardcoded values (use env vars or constants)
