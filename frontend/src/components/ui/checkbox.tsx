import * as React from "react";
import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import { CheckIcon } from "lucide-react";

import { cn } from "./utils";

type CheckboxProps = React.ComponentProps<typeof CheckboxPrimitive.Root> & {
  touchTarget?: boolean;
};

function Checkbox({
  className,
  touchTarget = false,
  ...props
}: CheckboxProps) {
  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      className={cn(
        touchTarget
          ? "peer group flex size-11 shrink-0 items-center justify-center rounded-xs outline-none transition-shadow focus-visible:ring-[3px] focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          : "peer border bg-input-background dark:bg-input/30 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground dark:data-[state=checked]:bg-primary data-[state=checked]:border-primary focus-visible:border-ring focus-visible:ring-ring aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive size-4 shrink-0 rounded-xs border transition-shadow outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    >
      {touchTarget ? (
        <span className="flex size-4 items-center justify-center rounded-xs border bg-input-background transition-colors group-data-[state=checked]:border-primary group-data-[state=checked]:bg-primary group-data-[state=checked]:text-primary-foreground dark:bg-input/30 dark:group-data-[state=checked]:bg-primary">
          <CheckboxPrimitive.Indicator
            data-slot="checkbox-indicator"
            className="flex items-center justify-center text-current transition-none"
          >
            <CheckIcon className="size-3.5" />
          </CheckboxPrimitive.Indicator>
        </span>
      ) : (
        <CheckboxPrimitive.Indicator
          data-slot="checkbox-indicator"
          className="flex items-center justify-center text-current transition-none"
        >
          <CheckIcon className="size-3.5" />
        </CheckboxPrimitive.Indicator>
      )}
    </CheckboxPrimitive.Root>
  );
}

export { Checkbox };
