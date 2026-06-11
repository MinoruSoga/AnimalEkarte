export const VISIT_TYPE_OPTIONS = ["初診", "再診", "緊急", "往診"] as const;

/** DB enum "first" | "revisit" → Japanese label */
export function fromVisitTypeValue(value: string | undefined): string {
  if (value === "first") return "初診";
  return "再診";
}

const MEDICAL_RECORD_TABS = [
  "問診",
  "診察/治療プラン",
  "治療",
  "予防接種",
  "定期健診",
  "検査",
  "画像",
  "見積書",
  "会計(医師確認)",
];

export const MEDICAL_RECORD_TAB_ITEMS = MEDICAL_RECORD_TABS.map((tab) => ({
  value: tab,
  label: tab,
}));
