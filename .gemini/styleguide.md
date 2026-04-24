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
- **Architecture (Feature-Based + Dependency Inversion)**:
  - **Structure**:
    ```
    src/
    ├── app/             # Application layer
    │   ├── pages/       # ★ Cross-feature synthesis (DI)
    │   ├── router.tsx   # createBrowserRouter
    │   └── provider.tsx # Providers
    ├── features/        # Feature-based modules
    │   └── [feature]/
    │       ├── api/     # API hooks (TanStack Query)
    │       ├── components/
    │       ├── hooks/   # Feature-specific logic (e.g., useXxxForm)
    │       ├── routes/  # Single-feature page components
    │       └── index.ts # Public API (Explicit exports only)
    ├── components/
    │   ├── ui/          # shadcn/ui
    │   └── shared/      # Common UI components
    ├── lib/             # axios.ts, react-query.ts, design-tokens.ts
    └── types/           # Shared types + generated/models.ts
    ```
  - **Dependency Inversion**: Features must NOT import from each other. If a page needs logic from multiple features, compose them in `src/app/pages/` and pass dependencies via props.
- **React 19 Patterns**:
  - **Refs**: Use `ref` as a prop directly. **DO NOT use `forwardRef`**.
  - **Transitions**: Use `useTransition` (`isPending`, `startTransition`) as the standard for managing async/form pending states.
  - **Conditional Rendering**: Always use ternary `condition ? <Component /> : null`. **DO NOT use `&&`**.
  - **Memoization**: Use `memo()` for large components (like form sections) to prevent unnecessary re-renders. Ensure props (handlers) are stabilized with `useCallback`.
- **State Management**:
  - **Server State**: TanStack Query is the primary source of truth for server data.
  - **Global State**: Use Zustand only for minimal UI state (e.g., sidebar collapse).
- **Styling**:
  - Use **Tailwind CSS 4**.
  - **MANDATORY**: Use `src/lib/design-tokens.ts` (`C` and `STYLE` constants) for all colors and common styles. **DO NOT** hardcode hex colors like `#37352F` in JSX.
- **Error Handling**:
  - **MANDATORY**: Use `handleApiError(error, "context")` in all `catch` blocks to ensure unified error reporting.
- **Imports**:
  - **MANDATORY**: Use the `@/` alias for all cross-feature or cross-layer imports. Relative paths (e.g., `../../`) are strictly forbidden across different modules.
- **Testing**:
  - **Tooling**: Vitest + React Testing Library.
  - **Placement**: Test files (`[Target].test.tsx`) must be in the same directory as the target.

## Testing Commands
- **Backend**: `docker compose exec backend go test ./...`
- **Frontend**: `docker compose exec frontend pnpm test:run`
