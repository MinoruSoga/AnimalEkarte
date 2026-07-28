import type { ComponentProps } from "react";

import { cn } from "@/components/ui/utils";
import { C } from "@/lib/design-tokens";

type DataTableRowButtonProps = Omit<ComponentProps<"button">, "type" | "aria-label"> & {
  "aria-label": string;
};

/** Native side-panel trigger used inside a table cell instead of making the row interactive. */
export function DataTableRowButton({ className, ...props }: DataTableRowButtonProps) {
  return (
    <button
      {...props}
      type="button"
      className={cn(
        `inline-flex min-h-11 min-w-11 items-center rounded-xs text-left text-sm font-medium ${C.textBrand} ${C.hoverTextBrand} outline-none hover:underline focus-visible:ring-2 ${C.focusRingAccent40}`,
        className,
      )}
    />
  );
}
