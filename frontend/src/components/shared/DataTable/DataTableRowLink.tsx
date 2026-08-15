import type { ComponentProps } from "react";
import { Link } from "react-router";
import { C } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

type DataTableRowLinkProps = Omit<ComponentProps<typeof Link>, "aria-label"> & {
  "aria-label": string;
};

/** Native detail link used inside a table cell instead of making the row interactive. */
export function DataTableRowLink({ className, ...props }: DataTableRowLinkProps) {
  return (
    <Link
      className={cn(
        `inline-flex min-h-11 min-w-11 items-center rounded-xs text-sm font-medium ${C.textBrand} ${C.hoverTextBrand} outline-none hover:underline focus-visible:ring-2 ${C.focusRingAccent40}`,
        className,
      )}
      {...props}
    />
  );
}
