import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { Loader2 } from "lucide-react";

import { cn } from "./utils";
import { buttonVariants, type ButtonVariantsProps } from "./button-variants";
import { ICON } from "@/lib/design-tokens";

// Re-export for compatibility (used by alert-dialog.tsx, etc.)
// eslint-disable-next-line react-refresh/only-export-components
export { buttonVariants };

interface ButtonProps extends React.ComponentProps<"button">, ButtonVariantsProps {
  asChild?: boolean;
  /**
   * design-states.md §3 — loading 状態の標準。
   * `disabled` + `aria-busy` + ラベル先頭のスピナーで表現し、テキストは維持する。
   * `asChild` 時はスロット構造を壊さないためスピナーは描画しない。
   */
  loading?: boolean;
  ref?: React.Ref<HTMLButtonElement>;
}

function Button({
  className,
  variant,
  size,
  asChild = false,
  loading = false,
  ref,
  type,
  disabled,
  children,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : "button";
  const typeProps: Pick<React.ComponentProps<"button">, "type"> = asChild
    ? type
      ? { type }
      : {}
    : { type: type ?? "button" };

  // Slot (asChild) は単一要素の子しか受け付けないため、スピナーは
  // asChild でない場合のみ children と合成する（asChild+loading は aria-busy のみ）。
  const content =
    loading && !asChild ? (
      <>
        <Loader2 className={cn(ICON.sm, "animate-spin")} aria-hidden="true" />
        {children}
      </>
    ) : (
      children
    );

  return (
    <Comp
      {...typeProps}
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      ref={ref}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      data-loading={loading || undefined}
      {...props}
    >
      {content}
    </Comp>
  );
}

export { Button, type ButtonProps };
