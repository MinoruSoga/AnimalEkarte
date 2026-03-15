import type { ReactNode } from "react";
import { C } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Notion-style property row for side peek panels
// ─────────────────────────────────────────────────

export function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className="w-[140px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}
