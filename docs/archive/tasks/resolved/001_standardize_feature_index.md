# Task: Standardize Feature Public API (index.ts)

## Description
Many features like `owners` and `medical-records` are missing the `index.ts` barrel file which acts as the Public API. This violates the Bulletproof React architecture defined in the project rules.

## Requirements
- Create `frontend/src/features/[feature]/index.ts` for all major features.
- Explicitly export only components, hooks, and types that are needed by the `app/` layer.
- Ensure `app/router.tsx` and `app/pages/` import from the top-level feature index where appropriate (or maintain direct file imports if preferred for tree-shaking, but the index must exist).

## Target Features
- owners
- medical-records
- hospitalization
- accounting
- etc.
