import { memo, useState } from "react";
import type { ReactNode } from "react";
import { ChevronDown } from "lucide-react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { C, ICON } from "@/lib/design-tokens";

interface InlineSelectorProps {
  label: string;
  children: ReactNode;
  popoverWidth?: string;
  noPadding?: boolean;
}

export const InlineSelector = memo(function InlineSelector({
  label,
  children,
  popoverWidth = "w-[180px]",
  noPadding = false,
}: InlineSelectorProps) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={`flex items-center gap-1 px-2 py-1 text-base ${C.text} ${C.bgMutedBadge} ${C.hoverBgMutedBadge} rounded-xxs transition-colors whitespace-nowrap max-w-[200px] truncate`}
        >
          <span className="truncate">{label}</span>
          <ChevronDown className={`${ICON.page} shrink-0 opacity-50`} />
        </button>
      </PopoverTrigger>
      <PopoverContent className={`${popoverWidth} ${noPadding ? "p-0" : "p-1"}`} align="start">
        {children}
      </PopoverContent>
    </Popover>
  );
});
