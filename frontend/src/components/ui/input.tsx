import * as React from "react";

import { cn } from "./utils";
import { C, PALETTE } from "@/lib/design-tokens";

interface InputProps extends React.ComponentProps<"input"> {
  ref?: React.Ref<HTMLInputElement>;
}

function Input({ className, type, ref, ...props }: InputProps) {
  return (
    <input
      type={type}
      className={cn(
        `file:text-foreground ${C.textPlaceholder} selection:bg-primary selection:text-primary-foreground border-input flex h-11 w-full min-w-0 rounded-xs border px-3 py-2 text-sm bg-transparent transition-colors outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50`,
        PALETTE.hoverBgInput,
        `focus:bg-white ${PALETTE.focusBorderLegacyAccent} ${PALETTE.focusRingActionPrimary}`,
        "aria-invalid:border-destructive aria-invalid:bg-destructive/5",
        className,
      )}
      ref={ref}
      {...props}
    />
  );
}

export { Input };
