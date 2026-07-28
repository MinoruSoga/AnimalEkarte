import type React from "react";
import { C } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Transparent inline text input for side peek panels
// ─────────────────────────────────────────────────

export function PropertyInput({
  value,
  onChange,
  placeholder,
  type = "text",
  ariaLabel,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: React.InputHTMLAttributes<HTMLInputElement>["type"];
  ariaLabel?: string;
}) {
  return (
    <input
      type={type}
      aria-label={ariaLabel}
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}
