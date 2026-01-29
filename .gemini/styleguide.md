# Gemini Code Assist Style Guide for Animal Ekarte

## Principles
- **Type Safety**: Prioritize strong typing in both Go and TypeScript. Avoid `any`.
- **SOLID**: Adhere to SOLID principles and Clean Architecture.
- **Error Handling**: Handle errors explicitly. Wrap errors in Go.
- **Security**: Prevent SQL injection and exposure of sensitive data.
- **Performance**: Optimize for performance; avoid N+1 queries.

## Go (Backend) Rules
- **Naming**:
  - Packages: lowercase (e.g., `handler`, `repository`).
  - Exports: PascalCase (e.g., `GetPatient`).
  - Private: camelCase (e.g., `validateInput`).
  - Interfaces: PascalCase + `er` suffix (e.g., `Reader`, `PatientRepository`).
- **Error Handling**:
  - Use `internal/errors` for sentinel errors.
  - Wrap errors with `fmt.Errorf("%s: %w", msg, err)`.
  - Check errors using `errors.Is()`.
- **Context**:
  - Pass `context.Context` as the first argument to all service and repository methods.
- **Logging**:
  - Use `slog` for structured logging.
  - Include context in logs: `slog.InfoContext(ctx, ...)`.

## TypeScript (Frontend) Rules
- **Tech Stack**: React 19, TypeScript 5.7, Tailwind CSS 4, shadcn/ui.
- **Naming**:
  - Components: PascalCase (e.g., `PatientCard.tsx`).
  - Hooks: camelCase, must start with `use` (e.g., `usePatient`).
  - Functions/Variables: camelCase (e.g., `getPatient`).
  - Constants: UPPER_SNAKE_CASE (e.g., `MAX_RETRY`).
  - Types: PascalCase (e.g., `Patient`), suffix with `Props` for component props (e.g., `PatientCardProps`).
- **Architecture (Bulletproof React)**:
  - **Structure**:
    ```
    src/
    ├── main.tsx
    ├── vite-env.d.ts
    ├── app/             # App entry, providers, router, ErrorBoundary
    │   ├── index.tsx
    │   ├── provider.tsx
    │   ├── router.tsx
    │   └── ErrorBoundary.tsx
    ├── features/        # Feature-based modules
    │   ├── auth/
    │   ├── dashboard/
    │   ├── reservations/
    │   ├── medical-records/ (SOAPS)
    │   ├── hospitalization/
    │   ├── accounting/
    │   ├── examinations/
    │   ├── owners/
    │   ├── pets/
    │   ├── trimming/
    │   ├── vaccinations/
    │   ├── master/
    │   └── clinic/
    │       ├── api/         # API hooks (e.g., getPatients.ts -> usePatients)
    │       ├── components/  # Nested components (Folder/Name.tsx, Name.test.tsx, index.ts)
    │       ├── hooks/       # Business logic / UI state
    │       ├── routes/      # Page components (List, Detail, Create)
    │       ├── types/       # Feature-specific types
    │       ├── utils/       # Feature-specific helpers
    │       ├── mockData.ts  # Feature-specific mock data
    │       └── index.ts     # Public API (Routes, API, Components, Types)
    ├── components/      # Shared components
    │   ├── ui/          # shadcn/ui (Radix UI Primitives)
    │   └── shared/      # App-specific shared components (Folder/Name.tsx, Name.test.tsx, index.ts)
    │       ├── Layout/ (MainLayout, Header, Sidebar, Navigation, AuthLayout, PrintLayout)
    │       ├── DataTable/
    │       ├── Form/ (FormField, FormError, FormLabel, SubmitButton, FormSection)
    │       ├── Feedback/ (Spinner, ErrorFallback, LoadingOverlay, EmptyState)
    │       ├── Navigation/ (Breadcrumb, NavLink)
    │       ├── StatusBadge/
    │       ├── DateRangePicker/
    │       ├── SearchBox/
    │       ├── ConfirmDialog/
    │       ├── PetSelectionSearchForm.tsx
    │       └── PetSelectionResultsTable.tsx
    ├── hooks/           # Global shared hooks (useDebounce, useDisclosure, etc.)
    ├── lib/             # Library configurations (axios, react-query, date-fns, zod, utils)
    ├── stores/          # Global state stores (auth, theme, notification, sidebar)
    ├── types/           # Shared types (api.ts, common.ts)
    ├── utils/           # Shared utilities (format, validation, constants, helpers)
    ├── config/          # Global config (constants, env)
    ├── styles/          # Global styles (globals.css, themes/)
    └── testing/         # Test configuration & utilities
        ├── setup.ts
        ├── server/      # MSW (Mock Service Worker)
        │   ├── handlers/ (auth.ts, patients.ts, medicalRecords.ts...)
        │   └── server.ts
        └── utils.tsx    # Test utilities (render provider wrapper)
    ```
  - **Encapsulation**: Features should only import from `src/components`, `src/hooks`, etc., or other features' public `index.ts`. Avoid deep imports into other features.
- **Naming Conventions**:
  - **Files**: camelCase for hooks/api (`usePatients.ts`), PascalCase for components (`PatientCard.tsx`).
  - **Tests**: `[Target].test.tsx` located alongside the target file.
  - **Public API**: `index.ts` must explicitly export only what is necessary.
- **React 19 Patterns**:
  - **Actions**: Use `useActionState` for form logic and `useFormStatus` for pending states.
  - **Optimistic UI**: Use `useOptimistic` for Kanban/DND and immediate feedback.
  - **Refs**: Use `ref` as a prop directly (no `forwardRef`).
  - **Data**: Use `use()` hook for promises and context.
  - **Routing**: React Router Data Mode (createBrowserRouter).
- **State Management**:
  - **Priority**: Server State (TanStack Query) > URL State > Local State > Global State (Zustand).
- **Testing Strategy**:
  - **Tooling**: Vitest + React Testing Library + MSW.
  - **Placement**: Test files in the same directory as the target component/hook.
  - **Focus**: Integration tests for features, unit tests for utils/shared hooks.
- **State Management**:
  - **Priority**: Server State (TanStack Query) > URL State > Local State (useState/useReducer) > Global State (Zustand/Context).
  - Avoid Redux.
- **Styling**:
  - Use Tailwind CSS 4.
  - Use `shadcn/ui` components for consistency.
  - Avoid inline styles.
- **Testing**:
  - Use Vitest + React Testing Library.
  - Focus on integration tests for features.
  - Mock API calls using MSW (Mock Service Worker) or similar at the network layer.

## Testing
- **Backend**: Use standard `testing` package. Run with `docker compose exec backend go test ./...`.
- **Frontend**: Use Vitest. Run with `docker compose exec frontend npm run test:run`.
