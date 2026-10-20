import type { ReactNode } from "react";

import { cn } from "@/components/ui/utils";
import { C } from "@/lib/design-tokens";

interface PropertyToggleButtonProps {
  label: string;
  /** "mixed" = 三状態の未設定（例: 旧欠落の保険設定）。aria-pressed にそのまま渡す。 */
  pressed: boolean | "mixed";
  onPressedChange: (value: boolean) => void;
  children: ReactNode;
  className?: string;
}

/** Controlled boolean button: readable name and state without submitting its form. */
export function PropertyToggleButton({
  label,
  pressed,
  onPressedChange,
  children,
  className,
}: PropertyToggleButtonProps) {
  return (
    <button
      type="button"
      aria-label={`${label}を切り替え`}
      aria-pressed={pressed}
      onClick={() => onPressedChange(pressed !== true)}
      className={cn(
        `inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-1.5 cursor-pointer`,
        className,
      )}
    >
      {children}
    </button>
  );
}
