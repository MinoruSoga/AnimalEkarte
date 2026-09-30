import type { ReactNode } from "react";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Notion-style property row for side peek panels
//
// The row renders as a <label> wrapping both the label
// text and the field control. This creates an implicit
// label/control association (WCAG 1.3.1, 4.1.2) for
// whatever focusable element `children` renders — input,
// textarea, custom PropertyInput, Select, Switch, button
// toggle, etc. — without requiring htmlFor/id coordination
// across every call site (R-F11).
//
// `description` を渡すと行右端に ⓘ ヘルプアイコンを表示する。
// アイコンは <label> の外に兄弟要素として置く — ラベル内に
// 対話要素や補足テキストを入れるとフィールドの accessible
// name に説明文が混入するため。
// ─────────────────────────────────────────────────

interface PropertyRowProps {
  label: string;
  /** 項目の説明文。指定すると行右端に ⓘ ツールチップを表示する */
  description?: string;
  children: ReactNode;
}

export function PropertyRow({ label, description, children }: PropertyRowProps) {
  return (
    <div className={STYLE.propertyRow}>
      <label className="flex flex-1 min-w-0 items-center gap-2">
        <span
          className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}
        >
          {label}
        </span>
        <span className="flex-1 min-w-0 flex items-center">{children}</span>
      </label>
      {description ? <FieldHelp label={label} content={description} /> : null}
    </div>
  );
}
