/**
 * Shared customer-facing brand palette (LINE reservation + LIFF health card).
 * Keep hex values identical to brand-tokens.css @theme definitions.
 */
export const NOAH_BRAND_COLORS = {
  teal: '#008B94',
  tealDark: '#007079',
  tealLight: '#EDF3F5',
  accent: '#FF8A43',
  text: '#212529',
  textSub: '#333333',
  /** Secondary text on teal header surfaces (readable on #008B94). */
  onTealSubtle: '#D6EEF0',
} as const;

/** LIFF Tailwind token names mapped to the shared teal system (not mint green). */
export const LIFF_BRAND_TOKEN_VALUES = {
  'liff-brand': NOAH_BRAND_COLORS.teal,
  'liff-brand-dark': NOAH_BRAND_COLORS.tealDark,
  'liff-brand-bg': NOAH_BRAND_COLORS.tealLight,
  'liff-brand-subtle': NOAH_BRAND_COLORS.onTealSubtle,
} as const;

/**
 * Fail-closed clinic header text: trim only; never invent a clinic name.
 * Empty/whitespace/missing → empty string (caller omits brand heading).
 */
export function resolveClinicHeaderText(headerText: string | null | undefined): string {
  if (typeof headerText !== 'string') {
    return '';
  }
  return headerText.trim();
}
