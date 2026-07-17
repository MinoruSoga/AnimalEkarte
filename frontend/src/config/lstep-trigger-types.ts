// トリガー種別の日本語ラベルマッピング（BE model の現行種別 + 旧履歴種別に対応）
// feature 非依存の共有定数（R-F2-S9: lstep から抽出）
export const TriggerTypeLabels: Record<string, string> = {
  first_visit_followup_3d: "初診フォロー 3日後",
  first_visit_followup_7d: "初診フォロー 7日後",
  next_visit_reminder: "次回来院リマインド",
  vaccine_deadline_60d: "ワクチン期限 60日前",
  vaccine_deadline_30d: "ワクチン期限 30日前",
  birthday_message: "誕生日メッセージ",
  // 旧履歴ログ互換。現行 BE は 180/210/240/365 日を使用する。
  dormant_prevention_120d: "休眠防止 120日",
  dormant_prevention_180d: "休眠防止 180日",
  dormant_prevention_210d: "休眠防止 210日",
  dormant_prevention_220d: "休眠防止 220日",
  dormant_prevention_240d: "休眠防止 240日",
  dormant_prevention_365d: "休眠防止 365日",
  filaria_alert: "フィラリアアラート",
  flea_tick_alert: "ノミ・ダニアラート",
  food_refill_reminder: "フードリマインド",
  supp_refill_reminder: "サプリリマインド",
  first_visit_welcome: "初診ウェルカム",
  checkup_followup: "健診フォロー",
};
