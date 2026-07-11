import * as React from "react";

import { cn } from "./utils";
import { PALETTE } from "@/lib/design-tokens";

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "resize-none border-input placeholder:text-muted-foreground flex field-sizing-content min-h-16 w-full rounded-xs border bg-transparent px-3 py-2 text-sm transition-colors outline-none disabled:cursor-not-allowed disabled:opacity-50",
        PALETTE.hoverBgInput,
        `focus:bg-white ${PALETTE.focusBorderLegacyAccent} ${PALETTE.focusRingLegacyAccent}`,
        "aria-invalid:border-destructive aria-invalid:bg-destructive/5",
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };