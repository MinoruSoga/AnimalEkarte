import * as React from "react";
import { Slot } from "@radix-ui/react-slot";

import { cn } from "./utils";
import { buttonVariants, type ButtonVariantsProps } from "./button-variants";

// Re-export for compatibility (used by alert-dialog.tsx, etc.)
// eslint-disable-next-line react-refresh/only-export-components
export { buttonVariants };

interface ButtonProps extends React.ComponentProps<"button">, ButtonVariantsProps {
  asChild?: boolean;
  ref?: React.Ref<HTMLButtonElement>;
}

function Button({ className, variant, size, asChild = false, ref, type, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : "button";
  const typeProps: Pick<React.ComponentProps<"button">, "type"> = asChild
    ? type
      ? { type }
      : {}
    : { type: type ?? "button" };

  return (
    <Comp
      {...typeProps}
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      ref={ref}
      {...props}
    />
  );
}

export { Button, type ButtonProps };
