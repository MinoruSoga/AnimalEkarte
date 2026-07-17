import { memo } from "react";
import { C } from "@/lib/design-tokens";
// ─────────────────────────────────────────────────
// Notion-style status pill
// ─────────────────────────────────────────────────

const STATUS_CONFIG = {
  active: {
    dot: `${C.bgBrandDot}`,
    label: "有効",
    bg: `${C.bgBrandLight}`,
    text: `${C.textBrandDark}`,
  },
  inactive: {
    dot: C.bgPrimary10,
    label: "無効",
    bg: `${C.bgInactive}`,
    text: C.text60,
  },
} as const;

export const StatusPill = memo(function StatusPill({ isActive }: { isActive: boolean }) {
  const cfg = STATUS_CONFIG[isActive ? "active" : "inactive"];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-sm text-base ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
});
