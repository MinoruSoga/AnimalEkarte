// トリガー種別の日本語ラベルマッピング（BE model の 15 種別に対応）
export const TriggerTypeLabels: Record<string, string> = {
  first_visit_followup_3d: "初診フォロー 3日後",
  first_visit_followup_7d: "初診フォロー 7日後",
  next_visit_reminder: "次回来院リマインド",
  vaccine_deadline_60d: "ワクチン期限 60日前",
  vaccine_deadline_30d: "ワクチン期限 30日前",
  birthday_message: "誕生日メッセージ",
  dormant_prevention_120d: "休眠防止 120日",
  dormant_prevention_180d: "休眠防止 180日",
  dormant_prevention_220d: "休眠防止 220日",
  filaria_alert: "フィラリアアラート",
  flea_tick_alert: "ノミ・ダニアラート",
  food_refill_reminder: "フードリマインド",
  supp_refill_reminder: "サプリリマインド",
  first_visit_welcome: "初診ウェルカム",
  checkup_followup: "健診フォロー",
};

export const TriggerStatusLabels: Record<string, string> = {
  scheduled: "予定",
  fired: "送信済",
  excluded: "除外",
  failed: "失敗",
};

// BADGE design token 色マッピング
export const TriggerStatusBadge: Record<
  string,
  "blue" | "green" | "gray" | "red"
> = {
  scheduled: "blue",
  fired: "green",
  excluded: "gray",
  failed: "red",
};
