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
