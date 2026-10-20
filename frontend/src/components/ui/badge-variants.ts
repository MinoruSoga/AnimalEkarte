import { cva, type VariantProps } from "class-variance-authority";

import { C } from "@/lib/design-tokens";

export const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-sm border px-2 py-1 text-sm font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none transition-colors overflow-hidden",
  {
    variants: {
      variant: {
        // 塗りは AA 適合の action-primary-solid（#027078）。#038B94 + 白文字は 4.10:1 で非達（design-system §2.1）。
        // バッジは識別用で操作を持たせない（DESIGN.md badge-pill）ため link hover は定義しない。
        default: `border-transparent ${C.bgActionPrimarySolid} ${C.textOnActionPrimary}`,
        secondary:
          "border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/80",
        destructive: "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90",
        outline:
          "text-foreground border-border [a&]:hover:bg-accent [a&]:hover:text-accent-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export type BadgeVariantsProps = VariantProps<typeof badgeVariants>;
