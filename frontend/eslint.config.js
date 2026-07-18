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
      // FE-R1: query-keys registry 再発防止ガード（FE-refactor.md 残件1 — registry 移行完了後の再発防止）
      "no-restricted-syntax": [
        "error",
        {
          selector: "JSXAttribute[name.name='dangerouslySetInnerHTML']",
          message:
            "dangerouslySetInnerHTML は禁止。生 HTML が必要な場合は DOMPurify 等でサニタイズしたうえでレビューを通すこと。",
        },
        {
          selector: "Property[key.name='queryKey'][value.type='ArrayExpression']",
          message:
            "queryKey に配列リテラルを直書きしない。frontend/src/lib/query-keys.ts の queryKeys ファクトリー（または ME_QUERY_KEY）経由で参照すること。新しいキー形状が必要な場合は queryKeys に追加する。",
        },
        {
          // setQueryData/getQueryData/removeQueries/resetQueries/cancelQueries/refetchQueries/
          // invalidateQueries はキーを { queryKey: [...] } ではなく第一引数の裸配列として
          // 受け取れるため、上の Property セレクタでは捕捉できない別経路。
          selector:
            "CallExpression[callee.property.name=/^(setQueryData|getQueryData|removeQueries|resetQueries|cancelQueries|refetchQueries|invalidateQueries)$/] > ArrayExpression.arguments:first-child",
          message:
            "queryKey に配列リテラルを直書きしない。frontend/src/lib/query-keys.ts の queryKeys ファクトリー（または ME_QUERY_KEY）経由で参照すること。",
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
      // FE7-0: deep import 禁止（全域）。feature の外からは @/features/<name>
      // （index.ts）経由で import する。feature 内部も相対 import を使うこと
      // （FE-refactor.md FE7-0 / frontend/CLAUDE.md Feature Indexing の機械強制）。
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            {
              group: ["@/features/*/*", "@/features/*/*/**"],
              message:
                "feature の外からは @/features/<name>（index.ts）経由で import する。feature 内部は相対 import を使うこと。deep import は禁止。",
            },
          ],
        },
      ],
    },
  },
  {
    // FE7-0: 層逆転禁止。components/hooks/lib から @/features への import を禁止する
    // （features は components/hooks/lib に依存してよいが逆方向は禁止 — 一方向依存の強制）。
    files: ["src/components/**", "src/hooks/**", "src/lib/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            {
              group: ["@/features", "@/features/**"],
              message:
                "components/hooks/lib から @/features への import は禁止（層逆転）。features 側が components/hooks/lib に依存する一方向のみ許可する。",
            },
          ],
        },
      ],
    },
  },
  {
    // FE7-0: アプリ境界。liff/line-reserve は frontend の features に依存しない
    // （3アプリは shared-liff 経由でのみ共有し、features は frontend 専有とする）。
    files: ["liff/src/**", "line-reserve/src/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            {
              group: ["@/features", "@/features/**"],
              message:
                "liff/line-reserve から frontend の @/features への import は禁止（アプリ境界）。共有が必要なら shared-liff 経由にすること。",
            },
          ],
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
  },
  {
    // Fast Refresh は「ファイルがコンポーネントのみ export」を要求するが、本リポジトリは
    // 画面/UI 単位で定数・ヘルパー・route config をコロケーションする方針を採る。
    // 分離すると import 拡散と再エクスポート層が増える一方、HMR の完全性は開発 UX のみ。
    // allowConstantExport は関数/JSX 定数/route 配列をカバーしないため、既知の同居面だけ off。
    // 新規ファイルへの安易な追加は避け、可能な限り *-model.ts 抽出を優先すること。
    files: [
      "src/app/routes/app-routes.tsx",
      "src/components/shared/DataTable/DataTable.tsx",
      "src/features/auth/components/LoginForm.tsx",
      "src/features/checkups/components/DynamicCheckupFields.tsx",
      "src/features/hospitalization/components/CarePlanTab/CarePlanBadges.tsx",
      "src/features/line-reservation/components/LineReservationSettingsFormSections.tsx",
      "src/features/manual/components/ManualContent.tsx",
      "src/features/owner-report/components/PetSwitcher.tsx",
      "src/features/reservations/routes/ReservationManagement.tsx",
    ],
    rules: {
      "react-refresh/only-export-components": "off",
    },
  },
  {
    // FE-R1: queryKey ファクトリーの定義自体はここでのみ配列リテラルを許可する
    files: ["src/lib/query-keys.ts"],
    rules: {
      "no-restricted-syntax": [
        "error",
        {
          selector: "JSXAttribute[name.name='dangerouslySetInnerHTML']",
          message:
            "dangerouslySetInnerHTML は禁止。生 HTML が必要な場合は DOMPurify 等でサニタイズしたうえでレビューを通すこと。",
        },
      ],
    },
  }
);
