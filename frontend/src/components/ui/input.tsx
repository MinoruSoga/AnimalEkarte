import * as React from "react";

import { cn } from "./utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground border-input flex h-10 w-full min-w-0 rounded-md border px-3 py-2 text-sm bg-transparent transition-colors outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50",
          "hover:bg-[rgba(242,241,238,0.5)]",
          "focus:bg-white focus:border-[rgba(35,131,226,0.57)] focus:shadow-[0_0_0_1px_rgba(35,131,226,0.35)]",
          "aria-invalid:border-destructive aria-invalid:bg-destructive/5",
          className,
        )}
        ref={ref}
        {...props}
      />
    );
  },
);
Input.displayName = "Input";

export { Input };
