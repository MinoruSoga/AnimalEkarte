// タブレットアクセシビリティ（高密度・大文字・小余白）のためのスタイル定数
export const H_STYLES = {
  text: {
    base: "text-base leading-snug",
    sm: "text-sm leading-snug",
    xs: "text-xs leading-tight", 
    lg: "text-lg font-bold",
    xl: "text-xl font-bold",
    muted: "text-[#37352F]/60",
  },
  padding: {
    card: "p-2",
    box: "p-2",
    tight: "p-1",
    none: "p-0",
  },
  button: {
    action: "h-10 px-4 text-base font-bold",
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
  }
} as const;
