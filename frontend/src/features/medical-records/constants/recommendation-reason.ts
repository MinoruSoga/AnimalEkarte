export const RECOMMENDATION_REASON_VALUES = ["revisit", "checkup", "prevention", "exam"] as const;

export type RecommendationReason = (typeof RECOMMENDATION_REASON_VALUES)[number];

export const RECOMMENDATION_REASON_LABELS: Record<RecommendationReason, string> = {
  revisit: "再診",
  checkup: "健診",
  prevention: "予防",
  exam: "検査",
};
