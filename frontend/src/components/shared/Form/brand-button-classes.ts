import { C } from "@/lib/design-tokens";

/**
 * DESIGN.md `button-primary` — SubmitButton/PrimaryButton 以外の brand CTA 用 SSOT。
 * design-tokens.ts は編集禁止のため、共有 className 文字列はここに集約する。
 */
export const BRAND_PRIMARY_BUTTON_CLASSES =
  `${C.bgBrand} ${C.hoverBgBrand} text-white h-11 px-4 text-base rounded-full transition-colors shadow-none border-transparent`;
