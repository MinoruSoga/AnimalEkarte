// レジ締めの区分（#150: emg=緊急 を追加）
export type CashRegisterPeriod = "am" | "pm" | "emg";

// 区分セレクタの表示順
export const PERIOD_OPTIONS: readonly CashRegisterPeriod[] = ["am", "pm", "emg"] as const;

// 区分の日本語ラベル
export const PERIOD_LABELS: Record<CashRegisterPeriod, string> = {
  am: "午前",
  pm: "午後",
  emg: "緊急",
};

export const CATEGORY_LABELS: Record<string, string> = {
  examination: "診療",
  test: "診療",
  procedure: "診療",
  medicine: "診療",
  surgery: "外科",
  vaccine: "RV",
  food: "フード",
  trimming: "トリミング",
  hotel: "ホテル",
  goods: "用品",
  training: "トレセン",
  // DEC-40: other は運用上の要確認 escape category。締め表ラベルは「未分類・要確認」。
  other: "未分類・要確認",
};

/** DEC-40: 締め表の other 行表示ラベル（DISPLAY_CATEGORIES / CATEGORY_LABELS と一致） */
export const UNCLASSIFIED_OTHER_LABEL = "未分類・要確認";

export const DISPLAY_CATEGORIES: { label: string; keys: string[] }[] = [
  { label: "診療", keys: ["examination", "test", "procedure", "medicine"] },
  { label: "外科", keys: ["surgery"] },
  { label: "RV", keys: ["vaccine"] },
  { label: "フード", keys: ["food"] },
  { label: "トリミング", keys: ["trimming"] },
  { label: "ホテル", keys: ["hotel"] },
  { label: "用品", keys: ["goods"] },
  { label: "トレセン", keys: ["training"] },
  { label: UNCLASSIFIED_OTHER_LABEL, keys: ["other"] },
];
