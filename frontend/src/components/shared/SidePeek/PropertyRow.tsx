import type { ReactNode } from "react";
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
// ─────────────────────────────────────────────────

export function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className={STYLE.propertyRow}>
      <span
        className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}
      >
        {label}
      </span>
      <span className="flex-1 flex items-center">{children}</span>
    </label>
  );
}
