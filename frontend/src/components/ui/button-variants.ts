import { cva, type VariantProps } from "class-variance-authority";
import { STYLE } from "@/lib/design-tokens";

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none",
  {
    variants: {
      variant: {
        // FE10: brand フィル = DESIGN.md button-primary = pill（rounded.full）
        default: "bg-primary text-primary-foreground hover:bg-primary/90 rounded-full",
        destructive:
          "bg-destructive text-white hover:bg-destructive/90",
        outline:
          "border border-input bg-transparent text-foreground hover:bg-accent/50",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost:
          "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
        // FE10: 旧 medical accent blue（第二構造アクセント）を brand へ統合 — DESIGN.md「第二構造アクセント禁止」
        primary: "bg-primary text-primary-foreground hover:bg-primary/90 rounded-full",
        "ghost-danger": STYLE.btnDangerGhost,
      },
      size: {
        default: "h-11 px-4 py-2 has-[>svg]:px-3",
        sm: "h-10 rounded-md gap-1.5 px-3 has-[>svg]:px-2.5",
        lg: "h-12 rounded-md px-8 has-[>svg]:px-4",
        icon: "size-11 rounded-md",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export type ButtonVariantsProps = VariantProps<typeof buttonVariants>;
