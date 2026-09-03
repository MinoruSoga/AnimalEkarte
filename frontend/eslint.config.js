import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import react from "eslint-plugin-react";

// TASK-444-S1: frozen residual importers of @/types/generated/models.
// ESLint flat config does not deep-merge no-restricted-imports options — later
// matching blocks fully replace earlier ones. The models boundary therefore
// re-states the FE7-0 deep-import / layer-inversion patterns for non-allowlisted
// files only (allowlisted files keep the earlier blocks unchanged).
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const generatedModelsImportAllowlist = JSON.parse(
  readFileSync(
    path.join(__dirname, "generated-models-import-allowlist.json"),
    "utf8",
  ),
);

const deepImportRestrictedPattern = {
  regex: "^@/features/(?!auth/provider$)[^/]+/.+$",
  caseSensitive: true,
  message:
    "feature の外からは @/features/<name>（index.ts）経由で import する。feature 内部は相対 import を使うこと。deep import は禁止。",
};

const layerInversionRestrictedPattern = {
  group: ["@/features", "@/features/**"],
  message:
    "components/hooks/lib から @/features への import は禁止（層逆転）。features 側が components/hooks/lib に依存する一方向のみ許可する。",
};

// FE-RC-015 followup2: alias だけでなく相対パス（../features/...）での層逆転も禁止。
const layerInversionRelativeRestrictedPattern = {
  regex: "^\\.\\.?/(?:\\.\\./)*features/",
  caseSensitive: true,
  message:
    "components/hooks/lib から相対パスで features への import は禁止（層逆転）。features 側が components/hooks/lib に依存する一方向のみ許可する。",
};

// FE-RC-015/060: feature 間の直接 import（barrel 経由含む）を機械禁止する。
// deep import（rule 1）は既に禁止済みのため、この pattern が実際に捕捉するのは
// barrel（@/features/<name>）経由の cross-feature import。
const crossFeatureRestrictedPattern = {
  group: ["@/features", "@/features/**"],
  message:
    "feature 間の直接 import は禁止（CODING_RULES.md §1.2）。他 feature の値は components/shared・hooks・lib へ昇格するか、app/pages/ の cross-feature 合成層で props 注入すること。",
};

// FE-RC-015 followup: owner-report は @/hooks/use-medical-records 経由に統一。
// 例外 allowlist は空（feature→feature を機械禁止）。
const crossFeatureImportBanAllowlist = [];

const generatedModelsBoundaryMessage =
  "TASK-444-S1: @/types/generated/models は Go ドメインモデル由来で HTTP wire 応答型ではない（BUG-431/BUG-433）。新規 import 禁止。既存利用は generated-models-import-allowlist.json に凍結。応答型は domain 別 generated/*-responses または専用 DTO を使うこと（TASK-444-S2）。";

// paths.name matches the written specifier only — relative `./generated/models`
// (and similar) would bypass a paths-only ban. Pair exact alias path with a
// regex that catches any import path ending at generated/models.
const generatedModelsPathRestriction = {
  name: "@/types/generated/models",
  message: generatedModelsBoundaryMessage,
};

const generatedModelsRelativeImportPattern = {
  // Relative only (./generated/models, ../types/generated/models, …).
  // Alias `@/types/generated/models` is covered by paths above — exclude `@`
  // so an alias import does not emit two identical TASK-444-S1 errors.
  regex: "^\\.{1,2}/(?:[\\w.@-]+/)*generated/models(?:\\.ts)?$",
  caseSensitive: true,
  message: generatedModelsBoundaryMessage,
};

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
      "react/jsx-no-leaked-render": ["error", { validStrategies: ["ternary"] }],
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
          patterns: [deepImportRestrictedPattern],
        },
      ],
    },
  },
  {
    // FE7-0: 層逆転禁止。components/hooks/lib から @/features への import を禁止する
    // （features は components/hooks/lib に依存してよいが逆方向は禁止 — 一方向依存の強制）。
    // FE-RC-015 followup2: 相対パス（../features/...）も同様に禁止。
    files: ["src/components/**", "src/hooks/**", "src/lib/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            layerInversionRestrictedPattern,
            layerInversionRelativeRestrictedPattern,
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
    // TASK-444-S1: block new @/types/generated/models imports outside the frozen
    // allowlist. Non-layer src files only — re-state deep-import so this later
    // no-restricted-imports block does not weaken FE7-0.
    files: ["src/**/*.{ts,tsx}"],
    ignores: [
      ...generatedModelsImportAllowlist,
      "src/components/**",
      "src/hooks/**",
      "src/lib/**",
    ],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [generatedModelsPathRestriction],
          patterns: [
            deepImportRestrictedPattern,
            generatedModelsRelativeImportPattern,
          ],
        },
      ],
    },
  },
  {
    // FE-RC-015/060: feature 間 import 禁止（src/features/** 全体・generated-models
    // allowlist 対象ファイル含む）。models freeze の対象外ファイルにも cross-feature-ban
    // だけは適用するための broad block。次の narrower block（models freeze 対象）が
    // 同じファイルに一致する場合はそちらが有効な設定として勝つ（flat config は
    // no-restricted-imports を feature 単位でマージせず後勝ちで完全上書きするため）。
    files: ["src/features/**/*.{ts,tsx}"],
    ignores: crossFeatureImportBanAllowlist,
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [deepImportRestrictedPattern, crossFeatureRestrictedPattern],
        },
      ],
    },
  },
  {
    // FE-RC-015/060: 上記 cross-feature-ban を models freeze 対象ファイルにも適用する
    // （deep-import + generated-models + cross-feature の3つを同時に再度明示）。
    files: ["src/features/**/*.{ts,tsx}"],
    ignores: [...generatedModelsImportAllowlist, ...crossFeatureImportBanAllowlist],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [generatedModelsPathRestriction],
          patterns: [
            deepImportRestrictedPattern,
            generatedModelsRelativeImportPattern,
            crossFeatureRestrictedPattern,
          ],
        },
      ],
    },
  },
  {
    // TASK-444-S1: same models freeze for components/hooks/lib, re-stating the
    // layer-inversion boundary (flat config replaces, does not merge, rule options).
    files: [
      "src/components/**/*.{ts,tsx}",
      "src/hooks/**/*.{ts,tsx}",
      "src/lib/**/*.{ts,tsx}",
    ],
    ignores: generatedModelsImportAllowlist,
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [generatedModelsPathRestriction],
          patterns: [
            layerInversionRestrictedPattern,
            layerInversionRelativeRestrictedPattern,
            generatedModelsRelativeImportPattern,
          ],
        },
      ],
    },
  },
  {
    // FE-RC-015 Phase 0 promotion: LabDeviceUnlinkedBanner は lab-device の attach/detach
    // mutation・banner 表示ヘルパーに依存する UI で、examinations/medical-records から共有
    // 参照できるよう components/shared へ移設した。実装（api/lib）自体は lab-device 内部の
    // ままのため、この 1 ファイルに限り層逆転ガードを外す（api/lib の昇格は将来 lane の対象）。
    // 注: @/types/generated/models への依存はリテラル値の直書きで解消済み（allowlist 追記不要）。
    files: ["src/components/shared/LabDeviceUnlinkedBanner/**"],
    rules: {
      "no-restricted-imports": "off",
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
      "src/components/shared/DynamicCheckupFields/DynamicCheckupFields.tsx",
      // FE-RC-015: checkups 側は components/shared への re-export shim（後方互換維持）。
      // 実体と同じ組み合わせ（component + helper 関数）を re-export するため同一の抑制が必要。
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
