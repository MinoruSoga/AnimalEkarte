import { C } from "@/lib/design-tokens";

// 入院機能のスタイル定数。text/padding/button/gap/layout はタブレットアクセシビリティ向け
// （高密度・大文字・小余白）。form は登録フォームの大余白カード用で、カード枠自体は
// 共有トークン STYLE.formCard（design-tokens.ts）を使う。
export const H_STYLES = {
  text: {
    base: "text-base leading-snug",
    sm: "text-sm leading-snug",
    xs: "text-xs leading-tight",
    lg: "text-xl font-bold",
    xl: "text-xl font-bold",
    muted: C.text60,
  },
  padding: {
    card: "p-2",
    box: "p-2",
    tight: "p-1",
    none: "p-0",
  },
  button: {
    action: "h-11 px-4 text-sm font-medium",
    icon: "h-4 w-4",
    sm: "h-9 px-3 text-sm font-bold",
  },
  gap: {
    default: "gap-2",
    tight: "gap-1",
    loose: "gap-4",
  },
  layout: {
    section_mb: "mb-2",
    card_mb: "mb-1",
  },
  form: {
    cardHeading: `text-heading-3 font-bold ${C.text}`,
  },
} as const;
