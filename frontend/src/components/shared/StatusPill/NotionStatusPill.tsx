import { C } from "@/lib/design-tokens";
// ─────────────────────────────────────────────────
// Notion-style status pill
// ─────────────────────────────────────────────────

const STATUS_CONFIG = {
  active: {
    dot: "bg-[#2383E2]",
    label: "有効",
    bg: "bg-[#D3E5EF]",
    text: "text-[#183B56]",
  },
  inactive: {
    dot: C.bgPrimary10,
    label: "無効",
    bg: "bg-[#E3E2E0]",
    text: C.text60,
  },
} as const;

export function NotionStatusPill({ isActive }: { isActive: boolean }) {
  const cfg = STATUS_CONFIG[isActive ? "active" : "inactive"];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-sm text-base ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}
