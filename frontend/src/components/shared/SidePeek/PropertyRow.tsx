import type { ReactNode } from "react";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Notion-style property row for side peek panels
// ─────────────────────────────────────────────────

export function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={STYLE.propertyRow}
    >
      <div className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}
