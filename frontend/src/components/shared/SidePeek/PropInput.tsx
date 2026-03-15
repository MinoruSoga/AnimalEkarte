import type React from "react";
import { C } from "@/lib/design-tokens";

// ─────────────────────────────────────────────────
// Transparent inline text input for side peek panels
// ─────────────────────────────────────────────────

export function PropInput({
  value,
  onChange,
  placeholder,
  type = "text",
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: React.InputHTMLAttributes<HTMLInputElement>["type"];
}) {
  return (
    <input
      type={type}
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}
