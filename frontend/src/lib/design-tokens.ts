/**
 * Design Tokens — Single Source of Truth
 *
 * All colors, layout dimensions, and composite Tailwind class presets live here.
 * Components import from this file instead of hardcoding hex values / px sizes.
 *
 * CSS-level tokens (--text-xs, --text-sm, --spacing) remain in globals.css
 * because they override Tailwind's @theme inline at the framework level.
 */

/* ================================================================== */
/*  1. Color Palette (raw values)                                      */
/*     Use in `style` props, canvas drawing, or anywhere a raw         */
/*     string is required.                                             */
/* ================================================================== */

export const PALETTE = {
  /** Ink (text, icons) */
  primary: "#000000",

  /** Page / section background */
  bgMain: "#F6F5F4",
  /** Slightly lighter background (e.g. date-time card) */
  bgSubtle: "#FFFFFF",
  /** Active / selected row background in sidebar */
  bgActive: "#F1F0EE",

  /** White */
  white: "#ffffff",

  /* ── External brand colors ── */
  /** LINE official brand green */
  lineGreen: "#06C755",

  /** Tailwind gray-500 — default color for permission group color pickers */
  pickerDefaultGray: "#6B7280",
  /** Tailwind blue-500 — default color for reservation category color pickers */
  pickerDefaultBlue: "#3B82F6",

  /* ── Brand (hospital main color) ── */
  /** Brand primary — docs/DESIGN_SYSTEM.md §2.1 {colors.primary} (teal, AE product override of DESIGN.md's Notion blue) */
  brand: "#038B94",
  /** Brand hover/pressed — docs/DESIGN_SYSTEM.md §2.1 {colors.primary-active} */
  brandHover: "#027078",
  /** Brand light background */
  brandLight: "#E1F3F4",
  /** Brand dark text (on light bg) */
  brandDark: "#025F66",

  /** Border – light (table cell, card) */
  borderLight: "rgba(0,0,0,0.09)",
  /** Border – medium (input, divider) */
  borderMedium: "rgba(0,0,0,0.16)",

  /** Skeleton / light background fill (6% primary) */
  bgSkeleton: "rgba(0,0,0,0.06)",
  /** Semi-transparent white (80%) — toggle dot on active chart buttons */
  whiteAlpha80: "rgba(255,255,255,0.8)",

  /** DESIGN.md primary blue accent */
  accent: "#0075DE",

  /** Destructive / danger — BUG-084: #C0392B (contrast 7.1:1 on white, WCAG AA) */
  danger: "#C0392B",

  /** Muted foreground (= mutedBadge text) */
  muted: "#787774",
  /** Muted background */
  mutedBg: "#F1F0EE",

  /* ── Status semantic ── */
  /**
   * Gray medium (#9B9A97)
   * Canonical value for: statusGrayText, bgStatusGrayMedium
   * Used as text color for muted status, solid bg for dot fills, chart legend dots.
   */
  grayMedium: "#9B9A97",
  /**
   * Gray light (#E3E2E0)
   * Canonical value for: statusInactiveBg, grayTagBg, mutedBorder
   * Used as inactive/disabled bg, tag bg, and muted badge border.
   */
  grayLight: "#E3E2E0",

  /* ── Success / Emerald green (会計リンク等) ── */
  successGreen: "#10B981",

  /* ── Orange badge border ── */
  borderOrangeBadge: "#F0C9A8",

  /* ── Master settings default colors (DB-configurable) ── */
  /** Default gray for badge when no color set */
  defaultGray: "#6B7280",

  /* ── Chart colors (Recharts / VitalsGraph) ── */
  /** Chart: temperature line */
  chartTemperature: "#E07B54",
  /** Chart: respiratory rate line */
  chartRespiratory: "#9C6EDE",
  /** Chart: body weight line */
  chartWeight: "#4CAF82",
  /** Chart: grid / axis stroke */
  chartGrid: "#e8e6e3",
  /** Chart: axis tick text fill */
  chartAxisText: "#9B9B97",

  /* ── UI Primitive (checkbox, input, select, textarea) ── */
  /** Input/select/textarea hover bg (warm neutral) */
  hoverBgInput:      "hover:bg-[rgba(242,241,238,0.5)]",
  /**
   * Input/select/textarea focus border — legacy accent (#2383E2) rgb 表記。
   * FE3-1: 値は変更せず既存直値をトークン化しただけ。正しい色への修正は別途デザイン判断。
   */
  focusBorderLegacyAccent: "focus:border-[rgba(35,131,226,0.57)]",
  /**
   * Input/select/textarea focus ring — legacy accent (#2383E2) rgb 表記。
   * FE3-1: 値は変更せず既存直値をトークン化しただけ。正しい色への修正は別途デザイン判断。
   */
  focusRingLegacyAccent: "focus:shadow-[0_0_0_1px_rgba(35,131,226,0.35)]",
} as const;

/* ================================================================== */
/*  2. Tailwind Color Classes                                          */
/*     Single-purpose class tokens for className composition.          */
/*     Use like: className={`${C.text} ${C.bgPage}`}                  */
/* ================================================================== */

export const C = {
  /* ── Text ── */
  text:          "text-[#000000]",
  text90:        "text-[#000000]/90",
  text80:        "text-[#000000]/80",
  text70:        "text-[#000000]/70",
  text65:        "text-[#000000]/65",
  text60:        "text-[#000000]/60",
  text55:        "text-[#000000]/55",
  text50:        "text-[#000000]/50",
  text45:        "text-[#000000]/45",
  text40:        "text-[#000000]/40",
  text35:        "text-[#000000]/35",
  text30:        "text-[#000000]/30",
  text25:        "text-[#000000]/25",
  text20:        "text-[#000000]/20",
  text15:        "text-[#000000]/15",
  textPlaceholder: "placeholder:text-[rgba(0,0,0,0.30)]",
  textPlaceholderFaint: "placeholder:text-[rgba(0,0,0,0.15)]",

  /* ── Background ── */
  bgPage:        "bg-[#F6F5F4]",
  bgPage60:      "bg-[#F6F5F4]/60",
  bgPageHalf:    "bg-[#F6F5F4]/50",
  bgPage30:      "bg-[#F6F5F4]/30",
  bgInputLogin:  "bg-[rgba(242,241,238,0.6)]",
  bgWhite:       "bg-white",
  bgSubtle:      "bg-white",
  bgActive:      "bg-[#F1F0EE]",
  bgHover:       "bg-[rgba(0,0,0,0.04)]",
  bgHoverMd:     "bg-[rgba(0,0,0,0.08)]",
  bgPrimary:     "bg-[#000000]",
  bgPrimary10:   "bg-[#000000]/10",
  bgPrimary8:    "bg-[#000000]/8",
  bgPrimary5:    "bg-[#000000]/5",
  /** Faint tinted background — skeleton shimmer base */
  bgSkeleton:    "bg-[rgba(0,0,0,0.06)]",
  /** Lightest bg — matching borderLight opacity; use for hr-style dividers */
  bgLight:       "bg-[rgba(0,0,0,0.09)]",

  /* ── Brand (Hospital teal) ── */
  textBrand:     "text-[#038B94]",
  bgBrand:       "bg-[#038B94]",
  bgBrand10:     "bg-[#038B94]/10",
  bgBrand8:      "bg-[#038B94]/8",
  bgBrand5:      "bg-[#038B94]/5",
  bgBrandDot:    "bg-[#038B94]",
  hoverBgBrand:  "hover:bg-[#027078]",
  hoverBgBrand5: "hover:bg-[#038B94]/5",
  /** Hover text — matches PALETTE.brandHover */
  hoverTextBrand: "hover:text-[#027078]",
  focusRingBrand:"focus:ring-[#038B94]",
  /** Focus ring (40% opacity) — use for form inputs matching the manual sidebar search field */
  focusRingBrand40: "focus:ring-[#038B94]/40",
  borderBrand:   "border-[#038B94]",
  borderLBrand:  "border-l-[#038B94]",
  /** Light brand border for outline/badge use — matches legacy borderAccentLight/borderAccentBadge weight */
  borderBrandLight: "border-[#038B94]/30",
  /** Tailwind accent-color utility (checkbox/radio) — brand teal */
  accentBrand:   "accent-[#038B94]",
  /* Brand light background — matches PALETTE.brandLight (#E1F3F4), replaces legacy bgAccentLight family */
  bgBrandLight:   "bg-[#E1F3F4]",
  bgBrandLight8:  "bg-[#E1F3F4]/8",
  bgBrandLight30: "bg-[#E1F3F4]/30",
  bgBrandLight40: "bg-[#E1F3F4]/40",
  bgBrandLight50: "bg-[#E1F3F4]/50",
  hoverBgBrandLight:   "hover:bg-[#E1F3F4]",
  hoverBgBrandLight60: "hover:bg-[#E1F3F4]/60",
  /** Darker brand text for contrast on bgBrandLight — matches PALETTE.brandDark (#025F66), replaces legacy textAccentDark on brand-light surfaces */
  textBrandDark: "text-[#025F66]",

  /* ── Border ── */
  borderLight:   "border-[rgba(0,0,0,0.09)]",
  borderLight50: "border-[rgba(0,0,0,0.09)]/50",
  borderMediumLight: "border-[rgba(0,0,0,0.12)]",
  borderMedium:  "border-[rgba(0,0,0,0.16)]",
  borderDivider: "border-[rgba(0,0,0,0.06)]",
  borderPrimary: "border-[#000000]",
  borderPrimary20: "border-[#000000]/20",
  borderPrimary25: "border-[#000000]/25",
  borderPrimary10: "border-[#000000]/10",
  borderLPrimary: "border-l-[#000000]",
  borderBPrimary: "border-b-[#000000]",
  divideDivider: "divide-[rgba(0,0,0,0.09)]",
  divideDividerFaint: "divide-[rgba(0,0,0,0.06)]",

  /* ── Accent ── */
  accent:        "text-[#2383E2]",
  bgAccent:      "bg-[#2383E2]",
  bgAccentHover: "hover:bg-[#1B6EC2]",
  bgAccentLight: "bg-[#D3E5EF]",
  bgAccentLight60: "bg-[#D3E5EF]/60",
  bgAccent5:     "bg-[#2383E2]/5",
  bgAccent8:     "bg-[#2383E2]/8",
  textAccentDark:"text-[#183B56]",
  textAccentDark90:"text-[#183B56]/90",
  borderAccent:  "border-[#2383E2]",
  /** Light accent border for outline accent buttons */
  borderAccentLight: "border-[#2383E2]/30",
  ringAccent:    "ring-[#2383E2]",
  ringAccent40:  "ring-[#2383E2]/40",
  hoverBgAccent5: "hover:bg-[#2383E2]/5",
  hoverBgAccent8: "hover:bg-[#2383E2]/8",
  /** Darker accent text on hover — use for outline accent buttons */
  hoverTextAccent: "hover:text-[#1B6EC2]",
  hoverBorderAccent40: "hover:border-[#2383E2]/40",
  focusBorderAccent: "focus:border-[#2383E2]",
  focusRingAccent: "focus-visible:ring-[#2383E2]",
  focusRingAccent40: "focus-visible:ring-[#2383E2]/40",
  /** Focus ring (30% opacity, `focus:` not `focus-visible:`) — plain form inputs (date/number fields) */
  focusRingAccent30: "focus:ring-[#2383E2]/30",
  /** data-state checked (Radix Checkbox) */
  dataCheckedBgAccent: "data-[state=checked]:bg-[#2383E2]",
  dataCheckedBorderAccent: "data-[state=checked]:border-[#2383E2]",
  dataCheckedBgBrand: "data-[state=checked]:bg-[#038B94]",
  dataCheckedBorderBrand: "data-[state=checked]:border-[#038B94]",

  /* ── Destructive — BUG-084: updated to #C0392B (7.1:1 contrast on white, WCAG AA) ── */
  danger:        "text-[#C0392B]",
  bgDanger:      "bg-[#C0392B]",
  bgDanger10:    "bg-[#C0392B]/10",
  bgDanger8:     "bg-[#C0392B]/8",
  hoverTextDanger: "hover:text-[#C0392B]",
  hoverBgDanger5: "hover:bg-[#C0392B]/5",
  /** Outline border for danger/destructive buttons */
  borderDanger:  "border-[#C0392B]/30",
  borderDanger20:"border-[#C0392B]/20",
  bgDanger4:     "bg-[#C0392B]/4",
  borderDanger15:"border-[#C0392B]/15",
  bgDanger20:    "bg-[#C0392B]/20",
  hoverBgDanger90: "hover:bg-[#C0392B]/90",
  decorationDanger50: "decoration-[#C0392B]/50",

  /* ── Notion Red (required markers) ── */
  textRequired:  "text-[#E03E3E]",
  textRedIcon:   "text-[#EA3323]",

  /* ── Status semantic ── */
  textStatusGreen:   "text-[#0F7B6C]",
  bgStatusGreen:     "bg-[#DDEDEA]",
  bgStatusGreenDot:  "bg-[#0F7B6C]",
  borderStatusGreen: "border-[#DDEDEA]",
  hoverBgStatusGreen:"hover:bg-[#DDEDEA]",
  /** #9B9A97 — PALETTE.grayMedium; see also bgStatusGrayMedium */
  textStatusGray:    "text-[#9B9A97]",
  bgStatusGray:      "bg-[#EBECED]",
  /** #9B9A97 — PALETTE.grayMedium; see also textStatusGray */
  bgStatusGrayMedium:"bg-[#9B9A97]",
  borderStatusGray:  "border-[#EBECED]",
  hoverBgStatusGray: "hover:bg-[#EBECED]",
  /** #E3E2E0 — PALETTE.grayLight; see also bgGrayTag, borderMuted */
  bgInactive:        "bg-[#E3E2E0]",

  /* ── Cost summary (Material) ── */
  bgCostBlue:    "bg-[#E3F2FD]",
  textCostBlue:  "text-[#1565C0]",
  bgCostGreen:   "bg-[#E8F5E9]",
  textCostGreen: "text-[#2E7D32]",

  /* ── Success / Emerald green (会計リンク等) ── */
  textSuccess:       "text-[#10B981]",
  bgSuccess:         "bg-[#10B981]",
  bgSuccess10:       "bg-[#10B981]/10",
  bgSuccessHover:    "hover:bg-[#059669]",
  borderSuccess30:   "border-[#10B981]/30",
  hoverBgSuccess10:  "hover:bg-[#10B981]/10",
  hoverBgSuccess20:  "hover:bg-[#10B981]/20",

  /* ── Notion Yellow / Notice (注意・アラート) ── */
  textNotice:        "text-[#C29243]",
  bgNotice:          "bg-[#FDECC8]",
  borderNotice:      "border-[#F2DBA7]",
  borderNotice50:    "border-[#F2DBA7]/50",
  bgNotice40:        "bg-[#FDECC8]/40",

  /* ── Notion Red (badges) ── */
  textNotionRed:     "text-[#E03E3E]",
  bgRedLight:        "bg-[#FFE2DD]",
  borderRed30:       "border-[#E03E3E]/30",
  bgNotionRed:       "bg-[#E03E3E]",

  /* ── Notion Orange — additional (trimming, financial buttons) ── */
  bgDiscount:        "bg-[#D9730D]",
  bgDiscountHover:   "hover:bg-[#D9730D]/90",
  borderDiscount20:  "border-[#D9730D]/20",

  /* ── Notion Purple — lighter border variant ── */
  borderPurpleLight: "border-[#D6C6E1]",

  /* ── Warning (yellow) ── */
  bgWarning50:   "bg-[#FFF3CD]/50",
  textWarningIcon: "text-[#B58105]",
  textWarning:   "text-[#856404]",
  borderWarning20: "border-[#B58105]/20",
  bgWarningBar:  "bg-[#F2AC57]",

  /* ── Discount / Financial orange ── */
  /** 値引き・差引額など財務文書用オレンジ (Notion orange に準拠) */
  textDiscount:  "text-[#D9730D]",
  bgDiscountLight: "bg-[#FAEBDD]",

  /* ── Operational status (reception kanban) ── */
  /** Notion purple — 診療中ステータス */
  textStatusPurple:  "text-[#6940A5]",
  bgStatusPurple:    "bg-[#EEE0F7]",
  bgStatusPurpleDot: "bg-[#6940A5]",
  borderStatusPurple:"border-[#6940A5]/20",
  hoverBgStatusPurple:"hover:bg-[#EEE0F7]",
  /** Slightly darker purple hover — use on bg-statusPurple buttons */
  hoverBgStatusPurpleDark: "hover:bg-[#E4D0F2]",

  /* ── Reservation status dot/bg/text (semantic status colors) ── */
  /** 予約確定 (confirmed) — Emerald green */
  bgStatusEmeraldDot: "bg-emerald-500",
  bgStatusEmerald:    "bg-emerald-50",
  textStatusEmerald:  "text-emerald-700",
  /** 仮予約 (pending) — Sky blue */
  bgStatusSkyDot:     "bg-sky-500",
  bgStatusSky:        "bg-sky-50",
  textStatusSky:      "text-sky-700",
  /** 受付済 (checked_in) — Blue */
  bgStatusBlueDot:    "bg-blue-500",
  bgStatusBlueLight:  "bg-blue-50",
  textStatusBlue:     "text-blue-700",
  /** 会計待ち (accounting) — Amber */
  bgStatusAmberDot:   "bg-amber-500",
  bgStatusAmber:      "bg-amber-50",
  textStatusAmber:    "text-amber-700",
  /** キャンセル / 初診 dot — Red */
  bgStatusRedDot:     "bg-red-500",

  /* ── Tailwind utility backgrounds (FG1 compliance) ── */
  /** Tailwind gray-100 background — skeleton loading placeholder */
  bgGray100:      "bg-gray-100",
  /** Tailwind gray-800 background — tooltip dark background */
  bgTooltip:      "bg-gray-800",

  /* ── Tailwind utility borders (FG1 compliance) ── */
  /** Tailwind blue-400 border — selected state borders */
  borderBlue400:  "border-blue-400",
  /** Tailwind blue-500 border — selected checkbox border */
  borderBlue500:  "border-blue-500",
  /** Tailwind gray-300 border — unselected checkbox border */
  borderGray300:  "border-gray-300",
  /** Tailwind gray-100 border — subtle separator */
  borderGray100:  "border-gray-100",
  /** Tailwind red-50 background — warning/error box */
  bgRed50:        "bg-red-50",
  /** Tailwind red-300 border — warning/error box */
  borderRed300:   "border-red-300",
  /** Tailwind red-700 text — warning/error box */
  textRed700:     "text-red-700",

  /* ── Hover utilities ── */
  hoverBgPage:   "hover:bg-[#F6F5F4]",
  hoverBgPageHalf: "hover:bg-[#F6F5F4]/50",
  hoverBgPage60: "hover:bg-[#F6F5F4]/60",
  hoverBgLight:  "hover:bg-[rgba(0,0,0,0.04)]",
  hoverBgMedium: "hover:bg-[rgba(0,0,0,0.08)]",
  hoverBgPrimary4:  "hover:bg-[#000000]/4",
  hoverBgPrimary10: "hover:bg-[#000000]/10",
  hoverBgPrimaryDark: "hover:bg-[#000000]/90",
  hoverText:     "hover:text-[#000000]",
  hoverText60:   "hover:text-[#000000]/60",
  hoverBorderPrimary30: "hover:border-[#000000]/30",
  hoverBorderPrimary40: "hover:border-[#000000]/40",
  hoverBorderMedium: "hover:border-[rgba(0,0,0,0.24)]",
  hoverBorderMedium40: "hover:border-[rgba(0,0,0,0.40)]",
  hoverBgSubtle: "hover:bg-[#F6F5F4]",

  /** Search input hover bg (slightly darker page) */
  hoverBgPageDark: "hover:bg-[#F1F0EE]",
  /** Group-hover bg primary (settings row icon) */
  groupHoverBgPrimary: "group-hover:bg-[#000000]",
  groupHoverTextWhite: "group-hover:text-white",
  textWhite: "text-white",

  /* ── Focus utilities ── */
  focusBgPage:   "focus:bg-[#F6F5F4]",
  focusBgLight:  "focus:bg-[rgba(0,0,0,0.04)]",
  focusBorderLight: "focus:border-[rgba(0,0,0,0.09)]",
  focusRingMedium: "focus-visible:ring-[rgba(0,0,0,0.16)]",

  /* ─ Ring ── */
  ringPrimary50: "ring-[#000000]/50",
  ringPrimary40: "ring-[#000000]/40",
  ringPrimary20: "ring-[#000000]/20",

  /* ── Notion Accent-light (badge / pill) — for status-helpers ── */
  /** Additional Accent tokens for badge combos */
  borderAccentBadge: "border-[#B8D4E3]",
  borderAccentBadge50: "border-[#B8D4E3]/50",

  /** Notion muted (default fallback badge) */
  bgMuted:        "bg-[#F1F0EE]",
  bgMutedBadge:   "bg-[#F1F1EF]",
  hoverBgMutedBadge: "hover:bg-[#E8E7E4]",
  bgMutedBadgeActive: "bg-[#E8E7E4]",
  textMuted:      "text-[#787774]",
  /** #E3E2E0 — PALETTE.grayLight; see also bgInactive, bgGrayTag */
  borderMuted:    "border-[#E3E2E0]",

  /** Additional green border for badges */
  borderStatusGreenAlt: "border-[#C3DFC3]",

  /** Notion Red badge border */
  borderRedBadge: "border-[#F5CBC4]",

  /** Notion Orange badge border */
  borderOrangeBadge: "border-[#F0C9A8]",

  /* ── Reception Kanban column opacity variants ── */
  bgAccentLight8:        "bg-[#D3E5EF]/8",
  bgAccentLight30:       "bg-[#D3E5EF]/30",
  bgAccentLight40:       "bg-[#D3E5EF]/40",
  bgAccentLight50:       "bg-[#D3E5EF]/50",
  hoverBgAccentLight:    "hover:bg-[#D3E5EF]",
  hoverBgAccentLight60:  "hover:bg-[#D3E5EF]/60",
  textAccentDark60:      "text-[#183B56]/60",
  hoverBgAccentBadge40:  "hover:bg-[#B8D4E3]/40",
  hoverTextAccentDark:   "hover:text-[#183B56]",

  bgStatusPurple30:      "bg-[#EEE0F7]/30",
  bgStatusPurple40:      "bg-[#EEE0F7]/40",
  bgStatusPurple60:      "bg-[#EEE0F7]/60",
  textStatusPurple60:    "text-[#6940A5]/60",
  hoverBgPurpleLight40:  "hover:bg-[#D6C6E1]/40",
  hoverTextStatusPurple: "hover:text-[#6940A5]",

  bgDiscountLight70:     "bg-[#FAEBDD]/70",
  textDiscount70:        "text-[#D9730D]/70",
  hoverBgOrangeBadge40:  "hover:bg-[#F0C9A8]/40",
  hoverTextDiscount:     "hover:text-[#D9730D]",

  bgStatusGreen30:       "bg-[#DDEDEA]/30",
  bgStatusGreen40:       "bg-[#DDEDEA]/40",
  hoverBgStatusGreenLight60: "hover:bg-[#DDEDEA]/60",
  hoverBgStatusPurpleLight60: "hover:bg-[#EEE0F7]/60",
  bgStatusGreen60:       "bg-[#DDEDEA]/60",
  textStatusGreen60:     "text-[#0F7B6C]/60",
  hoverBgGreenBadge40:   "hover:bg-[#C3DFC3]/40",
  hoverTextStatusGreen:  "hover:text-[#0F7B6C]",

  /* ── Service-type / shift specific ── */
  bgPurpleAlt2:      "bg-[#EBE2F5]",
  textPurpleAlt:     "text-[#5B2D8E]",
  borderPurpleAlt2:  "border-[#EBE2F5]",
  bgGrayTag:         "bg-[#E3E2E0]",
  textGrayTag:       "text-[#55534E]",
  borderGrayTag:     "border-[#D5D4D1]",
  textPaidLeave:     "text-[#7B5B29]",

  /* ── Medical accent blue (karte / hospitalization) ── */
  bgMedicalBlue:          "bg-[#2EAADC]",
  bgMedicalBlue5:         "bg-[#2EAADC]/5",
  textMedicalBlue:        "text-[#2EAADC]",
  borderMedicalBlue:      "border-[#2EAADC]",
  borderLMedicalBlue:     "border-l-[#2EAADC]",
  hoverBgMedicalBlue90:   "hover:bg-[#2EAADC]/90",
  hoverBorderMedicalBlue50: "hover:border-[#2EAADC]/50",
  ringMedicalBlue:        "ring-[#2EAADC]",
  focusRingMedicalBlue:   "focus-visible:ring-[#2EAADC]",

  /* ── 健診「期限間近」バッジ ── */
  bgCheckupDueSoon:   "bg-[#F0D070]",
  textCheckupDueSoon: "text-[#7A5C00]",

  /* ── Data-state active (Radix Tabs) ── */
  dataActiveBorderB: "data-[state=active]:border-b-[#000000]",
  dataActiveText:    "data-[state=active]:text-[#000000]",
} as const;

/* ================================================================== */
/*  2b. Badge Color Combos                                             */
/*      Reusable "bg + text + border" class strings for badges/pills.  */
/*      Used by status-helpers.ts and inline badges.                   */
/* ================================================================== */

export const BADGE = {
  /** Notion Blue — 作成中, 入院中, 受付済, 検査中, 予約(trimming), medicine, スタッフ */
  blue:    `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentBadge}`,
  /** Notion Gray — 確定済, 退院済, cancelled, item, inactive */
  gray:    `${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`,
  /** Notion Green (teal) — 予約(hosp), 完了(exam/trimming), completed, sufficient, active, 会計済 */
  green:   `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  /** Notion Red — 入院(type), waiting(acct), out_of_stock, 手術 */
  red:     `${C.bgRedLight} ${C.textNotionRed} ${C.borderRedBadge}`,
  /** Notion Purple — ホテル(type), treatment, トリマー, 予防接種 */
  purple:  `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderPurpleLight}`,
  /** Notion Orange — 進行中(trimming), food */
  orange:  `${C.bgDiscountLight} ${C.textDiscount} ${C.borderOrangeBadge}`,
  /** Notion Yellow — 依頼中, pending, low stock, excretion, 定期健診 */
  yellow:  `${C.bgNotice} ${C.textNotice} ${C.borderNotice}`,

  /** Default / fallback */
  muted:   `${C.bgMutedBadge} ${C.textMuted} ${C.borderMuted}`,

  /* ── Care Plan (no border) ── */
  blueNoBorder:   `${C.bgAccentLight} ${C.textAccentDark}`,
  greenNoBorder:  `${C.bgStatusGreen} ${C.textStatusGreen}`,
  grayNoBorder:   `${C.bgStatusGray} ${C.textStatusGray}`,

  /* ── Pet status (with hover) ── */
  greenHover: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen} ${C.hoverBgStatusGreen}`,
  grayHover:  `${C.bgStatusGray} ${C.textStatusGray} ${C.borderStatusGray} ${C.hoverBgStatusGray}`,
} as const;

/* ================================================================== */
/*  3. Icon Sizes                                                       */
/*     すべてのアイコンサイズはここで一元管理する。                        */
/*     直接 size-N / h-N w-N を書かず、このトークンを使うこと。           */
/* ================================================================== */

export const ICON = {
  /** ページタイトル横・セクションヘッダーアイコン (20px) */
  page:    "size-5",
  /** ツールバーボタン・サイドバーナビゲーションアイコン (20px) */
  toolbar: "size-5",
  /** アクションボタン・インラインアイコン (20px) */
  action:  "size-5",
  /** インジケーター・シェブロン (20px) */
  xs:      "size-5",
  /** フィルター・ドロップゾーンなど、やや大きいアイコン (24px) */
  lg:      "size-6",
  /** 空状態イラスト・大型装飾アイコン (32px) */
  xl:      "size-8",
  /** コンパクトアクション — バッジ・ボタン内アイコン (16px) */
  sm:      "size-4",
  /** インラインスモール — シェブロン・タグ・LINE状態 (14px) */
  smXs:    "size-3.5",
  /** 極小インジケーター — ステータス補完 (12px) */
  xxs:     "size-3",
  /** 極小ステータスドット / 通知バッジ (8px) */
  dot:     "size-2",
  /** 小ステータスドット / 色ラベル (10px) */
  dotMd:   "size-2.5",
  /** 最小インジケーター (6px) */
  dotSm:   "size-1.5",
  /** サイドバーナビ項目のアイコンラップ (18px) — テキスト行高に合わせた固有値。Tailwindのsize-4(16px)/size-5(20px)の中間で他箇所実績なし */
  navItem: "size-[18px]",
  /**
   * ユーザーアバター円 (26px) — Sidebar固有のtoken化。
   * 同値の生literal(size-[26px])がLoginForm.tsx/ForgotPasswordPage.tsx/ResetPasswordPage.tsx(認証フォーム、計4箇所)にも
   * 残存しているが、本タスクのスコープ(Sidebar/マスタページ)外のため未移行。認証フォーム側の統合は別タスクで検討。
   */
  avatar:  "size-[26px]",
  /** アバター内グリフ (13px) — Sidebar固有 */
  avatarGlyph: "size-[13px]",
} as const;

/* ================================================================== */
/*  4. Layout Dimensions                                               */
/*     Numeric values for animation targets, style props, etc.         */
/*     Tailwind class strings for width/height constraints.            */
/* ================================================================== */

export const LAYOUT = {
  /* ── Sidebar ── */
  sidebar: {
    expanded:      "w-[220px]",
    collapsed:     "w-[56px]",
    expandedPx:    220,
    collapsedPx:   56,
    /** 折りたたみ時の正方形ボタン高さ (30px) — トグル/ログアウトボタン共通 */
    collapsedItemH: "h-[30px]",
  },

  /* ── Side Peek (master edit, etc.) ── */
  sidePeek: {
    width:         "w-[520px]",
    widthPx:       520,
  },

  /** Master-detail nav panel width (tree/list side panel, e.g. LINE予約枠) */
  treeNavPanel: {
    width: "w-[260px]",
  },

  /* ── Property Row (Notion-style key-value) ── */
  propertyRow: {
    minH:          "min-h-[40px]",
    labelW:        "w-[140px]",
    labelWPx:      140,
  },

  /* ── Form Header ── */
  header: {
    h:             "h-[53px]",
  },

  /**
   * PageLayout content max-width presets — single source of truth for the
   * `maxWidth` prop so list pages and clinical detail pages don't drift.
   * `full`    一覧・カルテ詳細など、テーブルやタブで画面幅を使い切るページ
   * `default` 設定・単票フォーム系ページ（PageLayout の暗黙デフォルト）
   */
  pageContentMaxWidth: {
    full:    "max-w-full",
    default: "max-w-[1440px]",
  },

  /**
   * Full height flex container pattern.
   * flex-1: Fills the remaining space.
   * min-h-0: Overrides min-height: auto to allow the container to shrink and enable internal scrolling.
   * flex-col h-full: Ensures vertical orientation and inheritance.
   */
  fullHeight:      "flex-1 min-h-0 flex flex-col h-full",

  /** 
   * Technical fix for Flexbox/Grid content overflow.
   * Forces the element to be able to shrink smaller than its content, enabling 'overflow: auto' to work.
   */
  flexShrinkFix:   "min-h-0 min-w-0",

  /* ── Touch Targets (tablet-first) ── */
  touch: {
    /** Primary action buttons, input fields */
    md:            "h-11",
    /** Secondary / compact buttons */
    sm:            "h-11",
    /** Comfortable table row */
    row:           "h-12",
    /** Table header */
    tableHead:     "h-11",
    /** Toolbar icon button */
    iconBtn:       "size-9",
    /** Status badge */
    badge:         "h-8",
  },

  /**
   * Modal / Dialog sizes — single source of truth for all DialogContent / AlertDialogContent.
   *
   * sm  480px  クイックビュー・詳細モーダル
   * md  512px  入力フォーム・選択ダイアログ  (= sm:max-w-lg)
   * lg  768px  検索・プレビューダイアログ    (= sm:max-w-3xl)
   * xl  1000px 大フォームモーダル
   * full       全幅フォーム
   */
  modal: {
    sm:   "sm:max-w-[480px]",
    md:   "sm:max-w-lg",
    lg:   "sm:max-w-3xl",
    xl:   "sm:max-w-[1000px] max-h-[90vh]",
    full: "w-[98%] sm:max-w-[1400px] h-[90vh]",
  },

  /** Notion page title style */
  pageTitle: {
    fontSize:      "30px",
    fontWeight:    700,
    lineHeight:    "1.2",
  },

  /** Notion page icon */
  pageIcon: {
    size:          "size-[38px]",
    innerIcon:     "size-5",
  },

  /* ── Input / Color Picker compact patterns ── */
  /** Compact padding + rounded for small inline inputs and buttons */
  inputCompact:     "px-1.5 py-0.5 rounded-[3px]",
  /** Standard padding + rounded for number/text inputs in side panels */
  inputStandard:    "px-2 py-1 rounded-[3px]",
  /** Small color-picker input (w-7 h-7) */
  colorInputSmall:  "w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0",
  /** Medium color-picker input (w-12 h-12) */
  colorInputMedium: "w-12 h-12 rounded border",
  /** Small select-width input */
  selectInputSmall: "w-20 rounded-[3px]",
} as const;

/* ================================================================== */
/*  4. Composite Style Presets                                         */
/*     Reusable className strings for recurring UI patterns.           */
/* ================================================================== */

/**
 * FE3-1: sidePeekSaveBtn 埋め込み分と ShiftTemplateSettingsParts.tsx の保存ボタンで
 * 共有するピルボタンの影。STYLE オブジェクト内は自己参照できないため、
 * 定義前のモジュールスコープ定数として保持し STYLE.pillShadow から再輸出する。
 */
const PILL_SHADOW = "shadow-[0_1px_2px_rgba(0,0,0,0.1)]";

export const STYLE = {
  /* ── Page / Section ─ */
  page:            `${C.bgPage} overflow-hidden`,
  pageContent:     "flex-1 overflow-y-auto flex flex-col w-full px-3 py-5",
  sectionDivider:  `border-t ${C.borderLight}`,

  /* ── Form Header ── */
  formHeader:
    `sticky top-0 z-10 ${C.bgPage} border-b ${C.borderLight} px-4 flex items-center justify-between gap-2 h-[53px]`,

  /* ── Primary Button ── */
  btnGhost:
    `${C.text60} ${C.hoverText} hover:bg-transparent`,
  btnDanger:
    `${C.bgDanger} ${C.textWhite} ${C.hoverBgDanger90} h-11 px-4 text-base rounded-[4px] transition-colors shadow-none border-transparent`,
  btnOutline:
    `bg-white ${C.borderMedium} ${C.hoverBgLight} h-11 px-4 text-base rounded-[4px] shadow-[var(--notion-shadow-btn)] transition-colors`,

  /* ── Table ── */
  tableContainer:
    `bg-white border ${C.borderLight} rounded-[4px] flex flex-col flex-1 min-h-0`,
  tableHeaderRow:
    `border-b ${C.borderLight} ${C.bgPage30} h-11`,
  tableHeaderCell:
    `text-base font-medium ${C.text70} h-11`,
  tableRow:
    `border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors cursor-pointer h-16`,
  tableCell:
    `text-base ${C.text} py-2.5`,
  tableCellMono:
    `font-mono text-base ${C.text} py-2.5`,
  tableCellMuted:
    `text-base ${C.text70} py-2.5`,
  tableEmpty:
    `text-center py-12 ${C.text70} text-base`,
  tableEmptySm:
    `text-center py-12 ${C.text40} text-sm`,
  tableActionBtn:
    `h-11 w-11 ${C.text60} ${C.hoverText}`,

  /* ── Search Filter Bar ── */
  searchInput:
    `pl-8 h-11 w-full text-base ${C.text} ${C.textPlaceholder} ${C.bgPage} border border-transparent rounded-[4px] outline-none transition-colors ${C.hoverBgPageDark} focus:bg-white ${C.focusBorderLight}`,
  searchIcon:
    `absolute left-2.5 top-1/2 -translate-y-1/2 size-5 ${C.text30}`,
  searchCount:
    `text-base ${C.text60} whitespace-nowrap`,

  /* ── Pagination ── */
  paginationBtn:
    `h-8 w-8 ${C.text60} ${C.hoverBgPageHalf} rounded-[4px]`,
  paginationBtnActive:
    `h-8 w-8 ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} text-base rounded-[4px]`,
  paginationInfo:
    `text-base ${C.text50}`,

  /* ── Sidebar ── */
  sidebarContainer:
    `h-full ${C.bgPage} border-r ${C.borderLight} flex flex-col transition-all duration-300`,
  sidebarHeader:
    `h-[53px] flex items-center px-2.5 border-b ${C.borderDivider}`,
  sidebarItemActive:
    `${C.bgBrand8} ${C.text} border-l-2 ${C.borderLBrand}`,
  sidebarItemIdle:
    `${C.text65} ${C.hoverBgPrimary4} ${C.hoverText}`,
  sidebarToggle:
    `size-7 flex items-center justify-center ${C.text40} ${C.hoverText} ${C.hoverBgMedium} rounded-[3px] transition-colors`,

  /* ── Notion Property Row ── */
  propertyRow:
    `flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`,
  propertyInput:
    `w-full bg-transparent text-base ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`,

  /* ── Side Peek ── */
  sidePeekPanel:
    `flex flex-col h-full overflow-y-auto bg-white border-l ${C.borderLight} shadow-[-1px_0_5px_rgba(0,0,0,0.02)]`,
  sidePeekToolbar:
    "flex items-center justify-between h-[48px] px-3 shrink-0",
  sidePeekToolbarBtn:
    `size-9 flex items-center justify-center rounded-[3px] ${C.text45} ${C.hoverBgMedium} transition-colors`,

  /* ── Compact Icon Buttons (size-8=32px / size-7=28px / size-6=24px) ── */
  /** 32px アイコンボタン基底クラス (医療記録タブ・認証フォーム) */
  iconBtn32: `size-8 flex items-center justify-center rounded-[3px] transition-colors`,
  /** 28px アイコンボタン基底クラス (サイドバー・TreatmentRow) */
  iconBtn28: `size-7 flex items-center justify-center rounded-[3px] transition-colors`,
  /** 20px アイコンボタン基底クラス (折りたたみトグル等) */
  iconBtn20: `size-5 flex items-center justify-center rounded-[3px] transition-colors`,
  sidePeekBody:
    "flex-1 overflow-y-auto",
  sidePeekFooter:
    `flex items-center justify-end gap-2 px-4 py-3 border-t ${C.borderLight} shrink-0`,
  sidePeekCancelBtn:
    `px-4 py-[7px] text-base ${C.text65} ${C.hoverBgLight} rounded-[3px] transition-colors cursor-pointer`,
  sidePeekSaveBtn:
    `px-5 py-[7px] text-base text-white ${C.bgBrand} ${C.hoverBgBrand} rounded-[3px] transition-colors cursor-pointer ${PILL_SHADOW}`,

  /* ── Notion Page Icon ── */
  pageIcon:
    `size-[38px] flex items-center justify-center rounded-[3px] ${C.bgPage} ${C.text45}`,

  /* ── Select Trigger (compact, side peek) ── */
  selectCompact:
    `h-[30px] text-base bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-auto max-w-full`,

  /* ── Section heading (uppercase label) ── */
  sectionLabel:
    `text-base ${C.text55} uppercase tracking-wide select-none`,

  /* ── Status Badge ── */
  badge:
    "text-base px-2 h-9 font-normal border",

  /* ── Ghost Danger (delete buttons in form headers) ── */
  btnDangerGhost:
    `${C.danger} ${C.hoverBgDanger5} transition-colors`,

  /* ── Confirm dialog primary ── */
  confirmPrimary:
    `${C.bgAccent} text-white ${C.bgAccentHover} h-11 px-4 text-base rounded-[4px] transition-colors shadow-none border-transparent`,

  /* ── Master settings index row ── */
  settingsRow:
    `w-full flex items-center gap-3 px-4 py-3 ${C.hoverBgPage} transition-colors cursor-pointer group text-left`,
  settingsRowIcon:
    `size-[32px] flex items-center justify-center rounded-[3px] ${C.bgPage} ${C.text45} ${C.groupHoverBgPrimary} ${C.groupHoverTextWhite} transition-colors shrink-0`,

  /* ── Inline Add Row ── */
  inlineAddBtn:
    `w-full flex items-center gap-2 px-4 py-2.5 text-base ${C.text40} ${C.hoverText60} ${C.hoverBgPage} rounded-b-md transition-colors cursor-pointer group`,

  /* ── Form Controls (standard forms) ── */
  formLabel:
    `text-base ${C.text70}`,
  formInput:
    `h-11 text-base bg-white ${C.borderMedium} ${C.text}`,
  /** Error ring for form inputs — use with conditional classnames */
  formInputError:
    `ring-2 ring-[#C0392B]/30 ${C.borderDanger}`,
  formCard:
    `bg-white p-6 rounded-lg shadow-sm border ${C.borderMedium}`,

  /** Standard multi-line text area */
  textarea:     `w-full rounded-[3px] border ${C.borderMedium} bg-white p-3 text-sm ${C.text} outline-none ${C.focusBorderAccent} transition-colors resize-none leading-relaxed font-mono`,

  /* ── Drag Overlay ── */
  dragOverlayShadow: "shadow-[0_4px_16px_rgba(0,0,0,0.12)]",
  /**
   * Week view ドラッグ中プレビューの box-shadow 生値（framer motion の
   * whileDrag style prop 用。className ではなく inline style として使うため
   * `shadow-[...]` の Tailwind ラッパーは付けない）。dragOverlayShadow とは値が
   * 異なるため統合しない。FE3-1: 値は既存直値のまま。
   */
  dragPreviewShadowLarge: "0 10px 30px rgba(0,0,0,0.15)",
  /** 予約カード（AppointmentCard）の二重影。FE3-1: 値は既存直値のまま。 */
  cardShadow: "shadow-[0px_1px_3px_0px_rgba(0,0,0,0.1),0px_1px_2px_0px_rgba(0,0,0,0.06)]",
  /** ピルボタンの影（sidePeekSaveBtn と ShiftTemplateSettingsParts.tsx で共有）。 */
  pillShadow: PILL_SHADOW,
  /** Layout のナビゲーション進捗バー brand グロー（#038B94）。FE3-1: 値は既存直値のまま。 */
  brandGlow: "shadow-[0_0_8px_rgba(3,139,148,0.5)]",

  /* ── Table Row Hover (FG1 compliance) ── */
  /** Standard table row hover — use instead of hardcoded hover:bg-gray-50 */
  tableRowHover: "hover:bg-gray-50",
} as const;

/* ================================================================== */
/*  5. Backward-compatible exports (used by existing consumers)        */
/*     Gradually migrate callers to STYLE / C / LAYOUT above.         */
/* ================================================================== */

export const TABLE_STYLES = {
  row:          STYLE.tableRow,
  actionButton: STYLE.tableActionBtn,
  cell:         STYLE.tableCell,
  cellMono:     STYLE.tableCellMono,
} as const;
