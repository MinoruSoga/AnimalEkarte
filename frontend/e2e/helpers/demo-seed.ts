/**
 * Stable demo-seed anchors for E2E (003_demo CSV / runtime DB).
 * Legacy specs assumed pet id=1 / owner id=1 — those IDs are not present.
 */
export const DEMO_IRIS_PET = {
  id: '1000099',
  name: 'Iris',
} as const;

/** Owner name uses ideographic space (U+3000) in seed: 「林\u3000文明」. */
export const DEMO_HAYASHI_OWNER_NAME_RE = /林[\s\u3000]*文明/;

/** Full pet name for kana-symmetry search (single-char「ぴ」hits 500+ pets). */
export const DEMO_PETER_PET = {
  name: 'ピーター',
  hiraganaSearch: 'ぴーたー',
  katakanaSearch: 'ピーター',
} as const;

/**
 * Accounting list client-side kana filter (page-scoped).
 * Must appear on GET /accountings?page=1&limit=20 for clinic 1 seed.
 * Iris has no billing rows in 003_demo — do not use for accounting smoke.
 */
export const DEMO_ACCOUNTING_KANA_PET = {
  displayName: 'サキ',
  hiraganaSearch: 'さき',
  katakanaSearch: 'サキ',
} as const;
