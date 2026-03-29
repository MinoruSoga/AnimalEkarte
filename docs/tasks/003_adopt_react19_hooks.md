# Task: Adopt React 19 Hooks in Forms

## Description
The project is on React 19 but still uses manual pending state management in many places. The style guide recommends using `useActionState` and `useTransition`.

## Requirements
- Refactor `OwnerForm.tsx` or similar to use `useTransition` for pending states if not already fully utilized.
- Explore usage of `useActionState` for simpler non-controlled forms.
- Ensure `ref` is passed as a prop instead of using `forwardRef` (verified as mostly done, but needs check).
