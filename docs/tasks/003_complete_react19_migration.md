# Task: Complete React 19 Pattern Migration & Feature Indexing

## Status
- [x] Create `SubmitButton` with `useFormStatus`
- [x] Refactor `owners` feature (Hooks, Routes, Index, Router)
- [x] Refactor `medical-records` feature (Hooks, Routes, Index, Router)
- [x] Refactor `hospitalization` feature (Hooks, Routes, Index, Router)
- [x] Refactor `accounting` feature (Hooks, Routes, Index, Router)
- [x] Refactor `examinations` feature (Hooks, Routes, Index, Router)
- [x] Refactor `trimming` feature (Hooks, Routes, Index, Router)
- [x] Refactor `vaccinations` feature (Hooks, Routes, Index, Router)
- [x] Refactor `inventory` feature (Hooks, Routes, Index, Router)
- [x] Refactor `estimates` feature (Hooks, Routes, Index, Router)
- [x] Refactor `master` feature (Index, Router)
- [x] Refactor `hospital-settings` feature (Hooks, Routes, Index, Router)
- [x] Refactor `dashboard` feature (Index, Router)
- [x] Refactor `shifts` feature (Index, Router)
- [x] Refactor `reservations` feature (Index, Router)
- [x] Refactor `checkups` feature (Index, Router)
- [x] Refactor `pets` feature (Index)

## Description
This is a MANDATORY migration. Gradual introduction is no longer the policy. All components and forms must be updated to follow React 19 idiomatic patterns AND the Bulletproof React feature indexing pattern.

## Requirements
- **Eliminate manual pending states**: Replace all `const [isPending, setIsPending] = useState(false)` used for async operations with `useTransition` or `useActionState`.
- **Adopt useActionState**: Use `useActionState` for all form submissions to handle state and pending status natively.
- **Remove forwardRef**: Refactor any remaining `forwardRef` to use the `ref` prop directly.
- **Create Feature index.ts**: Ensure every feature has an `index.ts` file and use it for all external imports.
