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
- **Naming**:
  - Components: PascalCase (e.g., `PatientCard.tsx`).
  - Functions/Variables: camelCase (e.g., `getPatient`).
  - Constants: UPPER_SNAKE_CASE (e.g., `MAX_RETRY`).
  - Types: PascalCase (e.g., `Patient`).
- **Architecture**:
  - Feature-based structure: `features/<feature_name>/{api,components,hooks,types}`.
- **State Management**:
  - Use React Hooks and Context API.
  - Avoid Redux unless absolutely necessary.
- **Styling**:
  - Use Tailwind CSS utility classes.
  - Use `shadcn/ui` components for consistency.

## Testing
- **Backend**: Use standard `testing` package. Run with `docker compose exec backend go test ./...`.
- **Frontend**: Use Vitest/Jest. Run with `docker compose exec frontend npm run test:run`.
