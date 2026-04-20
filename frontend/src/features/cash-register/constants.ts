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
  other: "その他",
};

export const DISPLAY_CATEGORIES: { label: string; keys: string[] }[] = [
  { label: "診療", keys: ["examination", "test", "procedure", "medicine"] },
  { label: "外科", keys: ["surgery"] },
  { label: "RV", keys: ["vaccine"] },
  { label: "フード", keys: ["food"] },
  { label: "トリミング", keys: ["trimming"] },
  { label: "ホテル", keys: ["hotel"] },
  { label: "用品", keys: ["goods"] },
  { label: "トレセン", keys: ["training"] },
  { label: "その他", keys: ["other"] },
];
