import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import react from "eslint-plugin-react";

export default tseslint.config(
  { ignores: ["dist", "node_modules", "coverage", "src/types/generated/**"] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
      react,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      "react-refresh/only-export-components": [
        "warn",
        { allowConstantExport: true },
      ],
      // FE6-8: {cond && <JSX>} は cond が 0 / 空文字のとき意図せず数値・文字列を
      // レンダリングしてしまう（frontend/CLAUDE.md の Conditional Render 規約）。
      // recommended セットは入れず、このルールのみ有効化する。現状違反 0 件。
      "react/jsx-no-leaked-render": "error",
      "@typescript-eslint/no-unused-vars": [
        "warn",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "@typescript-eslint/no-explicit-any": "error",
      // Relax strict rules for common patterns
      "react-hooks/set-state-in-effect": "warn",
      "react-hooks/purity": "warn",
      "react-hooks/preserve-manual-memoization": "warn",
      // FE5-4: XSS 再発防止ガード（FE-refactor.md §4.1 — dangerouslySetInnerHTML 監査 CLOSED の再発防止）
      "no-restricted-syntax": [
        "error",
        {
          selector: "JSXAttribute[name.name='dangerouslySetInnerHTML']",
          message:
            "dangerouslySetInnerHTML は禁止。生 HTML が必要な場合は DOMPurify 等でサニタイズしたうえでレビューを通すこと。",
        },
      ],
      "no-restricted-properties": [
        "error",
        {
          object: "document",
          property: "write",
          message: "document.write は禁止。React の宣言的レンダリングを使うこと。",
        },
        {
          property: "innerHTML",
          message:
            "innerHTML への代入は禁止。React の JSX を使うか、生 HTML が必須なら DOMPurify でサニタイズすること。",
        },
        {
          property: "outerHTML",
          message:
            "outerHTML への代入は禁止。React の JSX を使うか、生 HTML が必須なら DOMPurify でサニタイズすること。",
        },
        {
          property: "insertAdjacentHTML",
          message:
            "insertAdjacentHTML は禁止。React の JSX を使うか、生 HTML が必須なら DOMPurify でサニタイズすること。",
        },
      ],
    },
  },
  {
    // AuthProvider (component) と useAuth (hook re-export) が同居するため Fast Refresh 警告を抑制
    files: ["src/features/auth/hooks/use-auth.tsx"],
    rules: {
      "react-refresh/only-export-components": "off",
    },
  }
);
