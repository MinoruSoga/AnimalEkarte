import * as React from "react";
import { Slot } from "@radix-ui/react-slot";

import { cn } from "./utils";
import { buttonVariants, type ButtonVariantsProps } from "./button-variants";

// Re-export for compatibility (used by alert-dialog.tsx, etc.)
export { buttonVariants };

interface ButtonProps
  extends React.ComponentProps<"button">,
    ButtonVariantsProps {
  asChild?: boolean;
  ref?: React.Ref<HTMLButtonElement>;
}

function Button({
  className,
  variant,
  size,
  asChild = false,
  ref,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : "button";

  return (
    <Comp
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      ref={ref}
      {...props}
    />
  );
}

export { Button, type ButtonProps };
