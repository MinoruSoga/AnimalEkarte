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
  /** Notion primary (text, icons, fills) */
  primary: "#37352F",

  /** Page / section background */
  bgMain: "#F7F6F3",
  /** Slightly lighter background (e.g. date-time card) */
  bgSubtle: "#FAFAF8",
  /** Active / selected row background in sidebar */
  bgActive: "#EAE9E5",

  /** White */
  white: "#ffffff",

  /* ── Brand (hospital main color) ── */
  /** Brand primary — veterinary teal */
  brand: "#038B94",
  /** Brand hover — slightly darker */
  brandHover: "#027A82",
  /** Brand light background */
  brandLight: "#E6F4F5",
  /** Brand dark text (on light bg) */
  brandDark: "#025E65",

  /** Border – light (table cell, card) */
  borderLight: "rgba(55,53,47,0.09)",
  /** Border – medium (input, divider) */
  borderMedium: "rgba(55,53,47,0.16)",

  /** Hover overlay – light */
  hoverLight: "rgba(55,53,47,0.04)",
  /** Hover overlay – medium */
  hoverMedium: "rgba(55,53,47,0.08)",
  /** Hover overlay – used for scrollbar thumb hover */
  hoverStrong: "rgba(55,53,47,0.35)",

  /** Notion blue accent */
  accent: "#2383E2",
  /** Accent hover */
  accentHover: "#1B6EC2",
  /** Accent light background (status pill) */
  accentLight: "#D3E5EF",
  /** Accent dark text */
  accentDark: "#183B56",

  /** Destructive / danger */
  danger: "#EB5757",

  /** Notion red (required markers, validation) */
  notionRed: "#E03E3E",
  /** Red icon (alert circle) */
  redIcon: "#EA3323",

  /** Muted foreground (= mutedBadge text) */
  muted: "#787774",
  /** Muted background */
  mutedBg: "#F1F0EE",

  /* ── Status semantic ── */
  statusGreenText: "#0F7B6C",
  statusGreenBg: "#DDEDEA",
  /**
   * Gray medium (#9B9A97)
   * Canonical value for: statusGrayText, bgStatusGrayMedium, dotGray
   * Used as text color for muted status, solid bg for dot fills, chart legend dots.
   */
  grayMedium: "#9B9A97",
  statusGrayBg: "#EBECED",
  /**
   * Gray light (#E3E2E0)
   * Canonical value for: statusInactiveBg, grayTagBg, mutedBorder
   * Used as inactive/disabled bg, tag bg, and muted badge border.
   */
  grayLight: "#E3E2E0",

  /* ── Cost summary (Material palette) ── */
  costBlueBg: "#E3F2FD",
  costBlueText: "#1565C0",
  costGreenBg: "#E8F5E9",
  costGreenText: "#2E7D32",

  /* ── Success / Emerald green (会計リンク等) ── */
  successGreen: "#10B981",
  successGreenHover: "#059669",

  /* ── Notion Yellow / Notice (注意・アラート) ── */
  noticeBg: "#FDECC8",
  noticeBorder: "#F2DBA7",
  noticeText: "#C29243",

  /* ── Notion Red — light badge bg ── */
  redLightBg: "#FFE2DD",

  /* ── Notion Orange ── */
  /** Notion orange — also used as discount/financial text color */
  notionOrange: "#D9730D",
  /** Notion orange light — also used as discount/financial bg */
  notionOrangeLight: "#FAEBDD",

  /* ── Notion Purple — lighter border ── */
  purpleBorderLight: "#D6C6E1",

  /* ── Warning (yellow) ── */
  warningBg:     "#FFF3CD",
  warningIcon:   "#B58105",
  warningText:   "#856404",
  warningBar:    "#F2AC57",

  /* ── Notion purple — 診療中ステータス ── */
  statusPurpleText: "#6940A5",
  statusPurpleBg:   "#EEE0F7",
  /** Slightly darker purple bg for hover states */
  statusPurpleBgDark: "#E4D0F2",

  /* ── Notion Accent-light badge border ── */
  accentBadgeBorder: "#B8D4E3",

  /* ── Notion Brown — 入院 reservation card ── */
  brownBg:     "#EEE0DA",
  brownText:   "#64473A",
  brownBorder: "#DFD0C8",

  /* ── Notion Pink — テル reservation card ── */
  pinkBg:      "#F5E0E9",
  pinkText:    "#AD1A72",
  pinkBorder:  "#ECCBDA",

  /* ── Muted badge ── */
  /** Muted badge bg (slightly warmer than mutedBg) */
  mutedBadgeBg:     "#F1F1EF",

  /* ── Red badge border ── */
  redBadgeBorder: "#F5CBC4",

  /* ── Orange badge border ── */
  borderOrangeBadge: "#F0C9A8",

  /* ── Service-type dot colors (saturated accents for legend) ── */
  dotBlue:    "#529CCA",
  dotGreen:   "#4DAB9A",
  dotRed:     "#E16F64",
  dotOrange:  "#E78D3C",
  dotYellow:  "#DFAB01",
  dotPurple:  "#9A6DD7",
  dotPink:    "#D44D8E",
  dotBrown:   "#937264",
  /** #9B9A97 — alias of grayMedium; kept here for dot-section readability */
  dotGray:    "#9B9A97",
  dotDefault: "#91918E",

  /* ── UI Primitive (checkbox, input, select, textarea) ── */
  /** Notion checkbox unchecked border */
  borderCheckbox:    "border-[#D3D1CB]",
  /** Input/select/textarea hover bg (warm neutral) */
  hoverBgInput:      "hover:bg-[rgba(242,241,238,0.5)]",
  /** Input/select/textarea focus border (primary 24%) */
  focusBorderInput:  "focus:border-[rgba(55,53,47,0.24)]",
} as const;

/* ================================================================== */
/*  2. Tailwind Color Classes                                          */
/*     Single-purpose class tokens for className composition.          */
/*     Use like: className={`${C.text} ${C.bgPage}`}                  */
/* ================================================================== */

export const C = {
  /* ── Text ── */
  text:          "text-[#37352F]",
  text80:        "text-[#37352F]/80",
  text70:        "text-[#37352F]/70",
  text65:        "text-[#37352F]/65",
  text60:        "text-[#37352F]/60",
  text55:        "text-[#37352F]/55",
  text50:        "text-[#37352F]/50",
  text45:        "text-[#37352F]/45",
  text40:        "text-[#37352F]/40",
  text35:        "text-[#37352F]/35",
  text30:        "text-[#37352F]/30",
  text25:        "text-[#37352F]/25",
  text20:        "text-[#37352F]/20",
  text15:        "text-[#37352F]/15",
  textPlaceholder: "placeholder:text-[rgba(55,53,47,0.3)]",
  textPlaceholderFaint: "placeholder:text-[rgba(55,53,47,0.15)]",

  /* ── Background ── */
  bgPage:        "bg-[#F7F6F3]",
  bgPageHalf:    "bg-[#F7F6F3]/50",
  bgPage30:      "bg-[#F7F6F3]/30",
  bgInputLogin:  "bg-[rgba(242,241,238,0.6)]",
  bgWhite:       "bg-white",
  bgSubtle:      "bg-[#FAFAF8]",
  bgActive:      "bg-[#EAE9E5]",
  bgHover:       "bg-[rgba(55,53,47,0.04)]",
  bgHoverMd:     "bg-[rgba(55,53,47,0.08)]",
  bgPrimary:     "bg-[#37352F]",
  bgPrimary10:   "bg-[#37352F]/10",
  bgPrimary8:    "bg-[#37352F]/8",
  bgPrimary5:    "bg-[#37352F]/5",
  /** Faint tinted background — skeleton shimmer base */
  bgSkeleton:    "bg-[rgba(55,53,47,0.06)]",
  /** Lightest bg — matching borderLight opacity; use for hr-style dividers */
  bgLight:       "bg-[rgba(55,53,47,0.09)]",

  /* ── Brand (Hospital teal) ── */
  bgBrand:       "bg-[#038B94]",
  bgBrandDot:    "bg-[#038B94]",

  /* ── Border ── */
  borderLight:   "border-[rgba(55,53,47,0.09)]",
  borderMediumLight: "border-[rgba(55,53,47,0.12)]",
  borderMedium:  "border-[rgba(55,53,47,0.16)]",
  borderDivider: "border-[rgba(55,53,47,0.06)]",
  borderPrimary: "border-[#37352F]",
  borderPrimary20: "border-[#37352F]/20",
  borderPrimary25: "border-[#37352F]/25",
  borderPrimary10: "border-[#37352F]/10",
  borderLPrimary: "border-l-[#37352F]",
  divideDivider: "divide-[rgba(55,53,47,0.09)]",
  divideDividerFaint: "divide-[rgba(55,53,47,0.06)]",

  /* ── Accent ── */
  accent:        "text-[#2383E2]",
  bgAccent:      "bg-[#2383E2]",
  bgAccentHover: "hover:bg-[#1B6EC2]",
  bgAccentLight: "bg-[#D3E5EF]",
  bgAccent5:     "bg-[#2383E2]/5",
  bgAccent8:     "bg-[#2383E2]/8",
  textAccentDark:"text-[#183B56]",
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
  focusRingAccent: "focus-visible:ring-[#2383E2]",
  focusRingAccent40: "focus-visible:ring-[#2383E2]/40",

  /* ── Destructive ── */
  danger:        "text-[#EB5757]",
  bgDanger:      "bg-[#EB5757]",
  bgDanger10:    "bg-[#EB5757]/10",
  bgDanger8:     "bg-[#EB5757]/8",
  hoverTextDanger: "hover:text-[#EB5757]",
  hoverBgDanger5: "hover:bg-[#EB5757]/5",
  /** Outline border for danger/destructive buttons */
  borderDanger:  "border-[#EB5757]/30",
  borderDanger20:"border-[#EB5757]/20",
  bgDanger4:     "bg-[#EB5757]/4",
  borderDanger15:"border-[#EB5757]/15",
  bgDanger20:    "bg-[#EB5757]/20",
  hoverBgDanger90: "hover:bg-[#EB5757]/90",

  /* ── Notion Red (required markers) ── */
  textRequired:  "text-[#E03E3E]",
  textRedIcon:   "text-[#EA3323]",

  /* ── Status semantic ── */
  textStatusGreen:   "text-[#0F7B6C]",
  bgStatusGreen:     "bg-[#DDEDEA]",
  bgStatusGreenDot:  "bg-[#0F7B6C]",
  borderStatusGreen: "border-[#DDEDEA]",
  hoverBgStatusGreen:"hover:bg-[#DDEDEA]",
  /** #9B9A97 — PALETTE.grayMedium; see also bgStatusGrayMedium, dotGray */
  textStatusGray:    "text-[#9B9A97]",
  bgStatusGray:      "bg-[#EBECED]",
  /** #9B9A97 — PALETTE.grayMedium; see also textStatusGray, dotGray */
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
  bgSuccessHover:    "hover:bg-[#059669]",
  borderSuccess30:   "border-[#10B981]/30",
  hoverBgSuccess10:  "hover:bg-[#10B981]/10",

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

  /* ── Operational status (dashboard kanban) ── */
  /** Notion purple — 診療中ステータス */
  textStatusPurple:  "text-[#6940A5]",
  bgStatusPurple:    "bg-[#EEE0F7]",
  bgStatusPurpleDot: "bg-[#6940A5]",
  borderStatusPurple:"border-[#6940A5]/20",
  hoverBgStatusPurple:"hover:bg-[#EEE0F7]",
  /** Slightly darker purple hover — use on bg-statusPurple buttons */
  hoverBgStatusPurpleDark: "hover:bg-[#E4D0F2]",

  /* ── Hover utilities ── */
  hoverBgPage:   "hover:bg-[#F7F6F3]",
  hoverBgPageHalf: "hover:bg-[#F7F6F3]/50",
  hoverBgLight:  "hover:bg-[rgba(55,53,47,0.04)]",
  hoverBgMedium: "hover:bg-[rgba(55,53,47,0.08)]",
  hoverBgPrimary4:  "hover:bg-[#37352F]/4",
  hoverBgPrimary10: "hover:bg-[#37352F]/10",
  hoverBgPrimaryDark: "hover:bg-[#37352F]/90",
  hoverText:     "hover:text-[#37352F]",
  hoverText60:   "hover:text-[#37352F]/60",
  hoverBorderPrimary30: "hover:border-[#37352F]/30",
  hoverBorderPrimary40: "hover:border-[#37352F]/40",
  hoverBorderMedium: "hover:border-[rgba(55,53,47,0.24)]",
  hoverBgSubtle: "hover:bg-[#FAFAF8]",

  /** Search input hover bg (slightly darker page) */
  hoverBgPageDark: "hover:bg-[#EFEEEB]",
  /** Group-hover bg primary (settings row icon) */
  groupHoverBgPrimary: "group-hover:bg-[#37352F]",
  groupHoverTextWhite: "group-hover:text-white",

  /* ── Focus utilities ── */
  focusBgPage:   "focus:bg-[#F7F6F3]",
  focusBgLight:  "focus:bg-[rgba(55,53,47,0.04)]",
  focusBorderLight: "focus:border-[rgba(55,53,47,0.09)]",
  focusRingMedium: "focus-visible:ring-[rgba(55,53,47,0.16)]",

  /* ─ Ring ── */
  ringPrimary50: "ring-[#37352F]/50",
  ringPrimary40: "ring-[#37352F]/40",
  ringPrimary20: "ring-[#37352F]/20",

  /* ── Notion Accent-light (badge / pill) — for status-helpers ── */
  /** Additional Accent tokens for badge combos */
  borderAccentBadge: "border-[#B8D4E3]",

  /** Notion Brown — 入院 reservation card */
  bgBrown:        "bg-[#EEE0DA]",
  textBrown:      "text-[#64473A]",
  borderBrown:    "border-[#DFD0C8]",

  /** Notion Pink — ホテル reservation card */
  bgPink:         "bg-[#F5E0E9]",
  textPink:       "text-[#AD1A72]",
  borderPink:     "border-[#ECCBDA]",

  /** Notion muted (default fallback badge) */
  bgMutedBadge:   "bg-[#F1F1EF]",
  textMuted:      "text-[#787774]",
  /** #E3E2E0 — PALETTE.grayLight; see also bgInactive, bgGrayTag */
  borderMuted:    "border-[#E3E2E0]",

  /** Reservation-specific blue tint */
  textAccentBlue: "text-[#0B6E99]",

  /** Additional green border for badges */
  borderStatusGreenAlt: "border-[#C3DFC3]",

  /** Notion Red badge border */
  borderRedBadge: "border-[#F5CBC4]",

  /** Notion Orange badge border */
  borderOrangeBadge: "border-[#F0C9A8]",

  /* ── Dashboard Kanban column opacity variants ── */
  bgAccentLight50:       "bg-[#D3E5EF]/50",
  textAccentDark60:      "text-[#183B56]/60",
  hoverBgAccentBadge40:  "hover:bg-[#B8D4E3]/40",
  hoverTextAccentDark:   "hover:text-[#183B56]",

  bgStatusPurple60:      "bg-[#EEE0F7]/60",
  textStatusPurple60:    "text-[#6940A5]/60",
  hoverBgPurpleLight40:  "hover:bg-[#D6C6E1]/40",
  hoverTextStatusPurple: "hover:text-[#6940A5]",

  bgDiscountLight70:     "bg-[#FAEBDD]/70",
  textDiscount70:        "text-[#D9730D]/70",
  hoverBgOrangeBadge40:  "hover:bg-[#F0C9A8]/40",
  hoverTextDiscount:     "hover:text-[#D9730D]",

  bgStatusGreen60:       "bg-[#DDEDEA]/60",
  textStatusGreen60:     "text-[#0F7B6C]/60",
  hoverBgGreenBadge40:   "hover:bg-[#C3DFC3]/40",
  hoverTextStatusGreen:  "hover:text-[#0F7B6C]",

  /* ── Service-type / shift specific ── */
  bgGreenAlt:        "bg-[#DBEDDB]",
  bgOrangeAlt:       "bg-[#FADEC9]",
  bgPurpleAlt:       "bg-[#E8DEEE]",
  bgPurpleAlt2:      "bg-[#EBE2F5]",
  textPurpleAlt:     "text-[#5B2D8E]",
  borderPurpleAlt2:  "border-[#EBE2F5]",
  bgGrayTag:         "bg-[#E3E2E0]",
  textGrayTag:       "text-[#55534E]",
  borderGrayTag:     "border-[#D5D4D1]",
  textPaidLeave:     "text-[#7B5B29]",

  /* ── Service-type dot colors (Tailwind) ── */
  dotBlue:    "bg-[#529CCA]",
  dotGreen:   "bg-[#4DAB9A]",
  dotRed:     "bg-[#E16F64]",
  dotOrange:  "bg-[#E78D3C]",
  dotYellow:  "bg-[#DFAB01]",
  dotPurple:  "bg-[#9A6DD7]",
  dotPink:    "bg-[#D44D8E]",
  dotBrown:   "bg-[#937264]",
  dotGray:    "bg-[#9B9A97]",
  dotDefault: "bg-[#91918E]",
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

  /* ── Reservation-specific extras ── */
  /** Notion Blue darker text — 診療 */
  blueAlt: `${C.bgAccentLight} ${C.textAccentBlue} ${C.borderAccentBadge}`,
  /** Notion Green (検査 reservation) */
  greenAlt:`${C.bgGreenAlt} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  /** Notion Orange lighter — トリミング reservation */
  orangeAlt: `${C.bgOrangeAlt} ${C.textDiscount} ${C.borderOrangeBadge}`,
  /** Notion Purple lighter — 予防接種 reservation */
  purpleAlt: `${C.bgPurpleAlt} ${C.textStatusPurple} ${C.borderPurpleLight}`,
  /** Notion Brown — 入院 reservation */
  brown:   `${C.bgBrown} ${C.textBrown} ${C.borderBrown}`,
  /** Notion Pink — ホテル reservation */
  pink:    `${C.bgPink} ${C.textPink} ${C.borderPink}`,
  /** Default / fallback */
  muted:   `${C.bgMutedBadge} ${C.textMuted} ${C.borderMuted}`,

  /* ── Care Plan (no border) ── */
  orangeNoBorder: `${C.bgDiscountLight} ${C.textDiscount}`,
  blueNoBorder:   `${C.bgAccentLight} ${C.textAccentDark}`,
  purpleNoBorder: `${C.bgStatusPurple} ${C.textStatusPurple}`,
  greenNoBorder:  `${C.bgStatusGreen} ${C.textStatusGreen}`,
  grayNoBorder:   `${C.bgStatusGray} ${C.textStatusGray}`,

  /* ── Pet status (with hover) ── */
  greenHover: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen} ${C.hoverBgStatusGreen}`,
  grayHover:  `${C.bgStatusGray} ${C.textStatusGray} ${C.borderStatusGray} ${C.hoverBgStatusGray}`,
} as const;

/* ================================================================== */
/*  3. Layout Dimensions                                               */
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
  },

  /* ── Side Peek (master edit, etc.) ── */
  sidePeek: {
    width:         "w-[520px]",
    widthPx:       520,
  },

  /* ── Property Row (Notion-style key-value) ── */
  propertyRow: {
    minH:          "min-h-[40px]",
    labelW:        "w-[140px]",
    labelWPx:      140,
  },

  /* ── Form Header ── */
  header: {
    h:             "h-12",
  },

  /* ── Touch Targets (tablet-first) ── */
  touch: {
    /** Primary action buttons, input fields */
    md:            "h-10",
    /** Secondary / compact buttons */
    sm:            "h-10",
    /** Comfortable table row */
    row:           "h-12",
    /** Table header */
    tableHead:     "h-10",
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
    full: "w-[98%] sm:max-w-[1200px] h-[90vh]",
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
    innerIcon:     "size-[20px]",
  },
} as const;

/* ================================================================== */
/*  4. Composite Style Presets                                         */
/*     Reusable className strings for recurring UI patterns.           */
/* ================================================================== */

export const STYLE = {
  /* ── Page / Section ─ */
  page:            `${C.bgPage} overflow-hidden`,
  pageContent:     "flex-1 overflow-y-auto flex flex-col w-full px-3 py-5",
  sectionDivider:  `border-t ${C.borderLight}`,

  /* ── Form Header ── */
  formHeader:
    `sticky top-0 z-10 ${C.bgPage} border-b ${C.borderLight} px-4 py-2 flex items-center justify-between gap-2 h-12`,
  formHeaderTitle: `text-sm ${C.text} leading-tight`,
  formHeaderDesc:  `text-xs ${C.text50} mt-0.5`,

  /* ── Primary Button ── */
  btnPrimary:
    `${C.bgAccent} ${C.bgAccentHover} text-white h-10 px-4 text-sm shadow-none border-transparent rounded-[4px] transition-colors`,
  btnGhost:
    `${C.text60} ${C.hoverText} hover:bg-transparent`,
  btnAccent:
    `text-white ${C.bgAccent} ${C.bgAccentHover} h-10 px-4 text-sm rounded-[4px] transition-colors shadow-none border-transparent`,
  btnDanger:
    `${C.bgDanger} text-white ${C.hoverBgDanger90} h-10 px-4 text-sm rounded-[4px] transition-colors shadow-none border-transparent`,
  btnOutline:
    `bg-white ${C.borderMedium} ${C.hoverBgLight} h-10 px-4 text-sm rounded-[4px] shadow-[var(--notion-shadow-btn)] transition-colors`,

  /* ── Table ── */
  tableContainer:
    `bg-white border ${C.borderLight} rounded-[4px] flex flex-col flex-1 min-h-0`,
  tableHeaderRow:
    `border-b ${C.borderLight} ${C.bgPage} ${C.hoverBgPage} h-10`,
  tableHeaderCell:
    `text-xs ${C.text70} h-10`,
  tableRow:
    `border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors cursor-pointer h-12`,
  tableCell:
    `text-sm ${C.text} py-2.5`,
  tableCellMono:
    `font-mono text-sm ${C.text} py-2.5`,
  tableCellMuted:
    `text-sm ${C.text70} py-2.5`,
  tableEmpty:
    `text-center py-12 ${C.text70} text-sm`,
  tableActionBtn:
    `h-10 w-10 ${C.text60} ${C.hoverText}`,

  /* ── Search Filter Bar ── */
  searchInput:
    `pl-8 h-10 w-full text-sm ${C.text} ${C.textPlaceholder} ${C.bgPage} border border-transparent rounded-[4px] outline-none transition-colors ${C.hoverBgPageDark} focus:bg-white ${C.focusBorderLight}`,
  searchIcon:
    `absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 ${C.text30}`,
  searchCount:
    `text-xs ${C.text60} whitespace-nowrap`,

  /* ── Pagination ── */
  paginationBtn:
    `h-8 w-8 ${C.text60} ${C.hoverBgPageHalf} rounded-[4px]`,
  paginationBtnActive:
    `h-8 w-8 ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} text-sm rounded-[4px]`,
  paginationInfo:
    `text-xs ${C.text50}`,

  /* ── Sidebar ── */
  sidebarContainer:
    `h-full ${C.bgPage} border-r ${C.borderLight} flex flex-col transition-all duration-300`,
  sidebarHeader:
    `h-[53px] flex items-center px-2.5 border-b ${C.borderDivider}`,
  sidebarItemActive:
    `bg-[#038B94]/8 ${C.text} border-l-2 border-l-[#038B94]`,
  sidebarItemIdle:
    `${C.text65} ${C.hoverBgPrimary4} ${C.hoverText}`,
  sidebarToggle:
    `h-7 w-7 p-0 ${C.text40} ${C.hoverText} ${C.hoverBgMedium} rounded-[3px]`,

  /* ── Notion Property Row ── */
  propertyRow:
    `flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`,
  propertyLabel:
    `w-[140px] shrink-0 text-sm ${C.text65} select-none truncate`,
  propertyInput:
    `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`,

  /* ── Side Peek ── */
  sidePeekPanel:
    `flex flex-col h-full bg-white border-l ${C.borderLight} shadow-[-1px_0_5px_rgba(0,0,0,0.02)]`,
  sidePeekToolbar:
    "flex items-center justify-between h-[48px] px-3 shrink-0",
  sidePeekToolbarBtn:
    `size-9 flex items-center justify-center rounded-[3px] ${C.text45} ${C.hoverBgMedium} transition-colors`,
  sidePeekBody:
    "flex-1 overflow-y-auto",
  sidePeekFooter:
    `flex items-center justify-end gap-2 px-4 py-3 border-t ${C.borderLight} shrink-0`,
  sidePeekCancelBtn:
    `px-4 py-[7px] text-sm ${C.text65} ${C.hoverBgLight} rounded-[3px] transition-colors cursor-pointer`,
  sidePeekSaveBtn:
    `px-5 py-[7px] text-sm text-white ${C.bgAccent} ${C.bgAccentHover} rounded-[3px] transition-colors cursor-pointer shadow-[0_1px_2px_rgba(0,0,0,0.1)]`,

  /* ── Notion Page Icon ── */
  pageIcon:
    `size-[38px] flex items-center justify-center rounded-[3px] ${C.bgPage} ${C.text45}`,

  /* ── Notion "more" dropdown shadow ── */
  dropdownShadow:
    "bg-white rounded-[4px] shadow-[var(--notion-shadow)]",

  /* ── Select Trigger (compact, side peek) ── */
  selectCompact:
    `h-[30px] text-sm bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-auto max-w-full`,

  /* ── Section heading (uppercase label) ── */
  sectionLabel:
    `text-xs ${C.text55} uppercase tracking-wide select-none`,

  /* ── Status Badge ── */
  badge:
    "text-sm px-2 h-8 font-normal border",

  /* ── Confirm dialog primary ── */
  confirmPrimary:
    `${C.bgAccent} text-white ${C.bgAccentHover} h-10 px-4 text-sm rounded-[4px] transition-colors shadow-none border-transparent`,

  /* ── Master settings index row ── */
  settingsRow:
    `w-full flex items-center gap-3 px-4 py-3 ${C.hoverBgPage} transition-colors cursor-pointer group text-left`,
  settingsRowIcon:
    `size-[32px] flex items-center justify-center rounded-[3px] ${C.bgPage} ${C.text45} ${C.groupHoverBgPrimary} ${C.groupHoverTextWhite} transition-colors shrink-0`,

  /* ── Inline Add Row ── */
  inlineAddBtn:
    `w-full flex items-center gap-2 px-4 py-2.5 text-sm ${C.text40} ${C.hoverText60} ${C.hoverBgPage} rounded-b-md transition-colors cursor-pointer group`,

  /* ── Form Controls (standard forms) ── */
  formLabel:
    `text-sm ${C.text70}`,
  formInput:
    `h-10 text-sm bg-white ${C.borderMedium} ${C.text}`,
  /** Lighter border + hover variant — use for SelectTrigger in reservation/form fields */
  formInputLight:
    `h-10 text-sm bg-white ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle} transition-colors`,
  /** Error ring for form inputs — use with conditional classnames */
  formInputError:
    "ring-2 ring-red-300 border-red-400",
  formCard:
    `bg-white p-6 rounded-lg shadow-sm border ${C.borderMedium}`,
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
