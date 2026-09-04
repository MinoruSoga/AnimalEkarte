// トリガー種別の日本語ラベルマッピング — R-F2-S9 で src/config/ へ抽出
export { TriggerTypeLabels } from "@/config/lstep-trigger-types";

export const TriggerStatusLabels: Record<string, string> = {
  scheduled: "予定",
  fired: "送信済",
  excluded: "除外",
  failed: "失敗",
};

// BADGE design token 色マッピング
export const TriggerStatusBadge: Record<string, "blue" | "green" | "gray" | "red"> = {
  scheduled: "blue",
  fired: "green",
  excluded: "gray",
  failed: "red",
};
