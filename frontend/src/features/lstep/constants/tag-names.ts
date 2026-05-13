// FEAT-379 で追加した健診・予防・LTV タグ名定数。
// BE backend/internal/service/lstep_health_codes.go と文字列一致していること。
export const LstepTagNames = {
  HLTH_健診あり: "HLTH_健診あり",
  HLTH_健診未受診: "HLTH_健診未受診",
  HLTH_年4回候補: "HLTH_年4回候補",
  PREV_ワクチン期限: "PREV_ワクチン期限",
  PREV_フィラリア未完了: "PREV_フィラリア未完了",
  PREV_ノミダニ対象: "PREV_ノミダニ対象",
  LTV_フード購入あり: "LTV_フード購入あり",
} as const;

export type LstepTagName = (typeof LstepTagNames)[keyof typeof LstepTagNames];
