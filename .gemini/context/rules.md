# Operational Rules (Best Practices)

## 1. Documentation First (Shift Left)
-   **Principle**: Code without context is legacy code.
-   **Action**: Update `README.md`, `specs.md`, or JSDoc/docstrings *before* writing the implementation.

## 2. Planning (Plan.md)
-   **Principle**: Measure twice, cut once.
-   **Action**: For any task > 5 lines of code, check or create a `plan.md` first.
-   **Template**: Use the structure in `.gemini/plan_template.md`.

## 3. Testing
-   **Principle**: Untested code is broken code.
-   **Action**: Generate unit tests for new functions immediately.

## 4. Context Management
-   **Principle**: Maintain the knowledge base.
-   **Action**: Read `GEMINI.md` at start. Update it at end.

## 5. Security & Secrets
-   **Principle**: Leakage is irreversible.
-   **Action**: Never commit API keys or secrets. Use `.env` files and environment variables.
-   **Check**: Scan code for hardcoded credentials before saving.

## 6. Robustness & Error Handling
-   **Principle**: Systems fail. Code should survive.
-   **Action**: Implement graceful error handling (try/catch) and retry mechanisms for network operations.
-   **Guideline**: Do not assume happy paths. Verify inputs.

## 7. Official Patterns
-   **Principle**: Stand on the shoulders of giants.
-   **Action**: Prefer official SDKs and libraries (especially for Google Cloud/Gemini) over custom wrappers.
-   **Reference**: Follow official "Cookbooks" and documentation patterns.
