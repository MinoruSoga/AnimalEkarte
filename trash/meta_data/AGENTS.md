# Repository Guidelines

## Project Structure & Module Organization
- `backend/` contains the Go API, domain logic, migrations, and backend docs.
- `frontend/` contains the main React 19 + TypeScript app, plus `frontend/line-reserve/` for the LINE reservation UI.
- `docs/` stores product and implementation references; `backend/docs/` and `frontend/README.md` provide area-specific notes.
- Tests live next to code (`*.test.ts`, `*.test.tsx`, `*_test.go`) and in `frontend/tests/` for browser-level checks.
- Generated frontend types are written to `frontend/src/types/generated/`; do not edit them by hand.

## Build, Test, and Development Commands
- `make up` starts the full Docker Compose stack.
- `make build` rebuilds and starts containers.
- `make lint` runs the Go linter in a container; `make lint-front` runs ESLint in `frontend/`.
- `make test` runs backend tests; `make test-front` runs Vitest; `make build-front` verifies the frontend build.
- `make codegen` regenerates TypeScript models from Go; run `make codegen-check` before committing generated changes.
- `make ci-local` runs the same backend/frontend checks used in CI.

## Coding Style & Naming Conventions
- Follow the repo-specific rules in `backend/CODING_RULES.md` and `frontend/CODING_RULES.md`.
- Keep backend code layered (`handler` -> `service` -> `repository` -> `model`) and pass `context.Context` through every layer.
- In the frontend, keep feature code inside `frontend/src/features/<feature>/` and shared UI in `frontend/src/components/shared/`.
- Use TypeScript/Go formatter defaults, two-space indentation in frontend code, and conventional file names such as `use-*.ts`, `*.test.tsx`, and `*_service.go`.

## Testing Guidelines
- Prefer table-driven Go tests and small, focused Vitest tests.
- Name tests after the behavior under test, for example `use-pagination.test.ts` or `accounting_service_test.go`.
- Run `make test`, `make test-front`, and `make schema-check` when touching core logic, database models, or generated types.

## Commit & Pull Request Guidelines
- Commit messages follow a conventional pattern seen in history, such as `fix(scope): ...`, `refactor: ...`, or `docs(test-report): ...`.
- Keep each commit focused and mention ticket IDs or bug numbers when relevant.
- PRs should include a short summary, linked issue or ticket, test results, and screenshots or screen recordings for UI changes.
- Call out migrations, code generation, and schema changes explicitly in the PR description.

## Security & Configuration Tips
- Copy `.env.example` to `.env` before running locally.
- Avoid committing secrets or editing generated artifacts unless you are regenerating them intentionally.
