/**
 * 曜日ラベル（0=日曜〜6=土曜）。
 * FE4-11: master/ReservationTypeUnavailableTimesSection.tsx・
 * ReservationTypeAvailableSlotsSection.tsx・closing-settings/StandardClosingTimeSection.tsx
 * に同一内容が3複製されていたため共有定数化。
 */
export const DAY_OF_WEEK_LABELS: Record<number, string> = {
  0: "日",
  1: "月",
  2: "火",
  3: "水",
  4: "木",
  5: "金",
  6: "土",
};

/** 0=日曜 始まりの曜日ラベル配列（Date#getDay() 添字用。FE5-8） */
export const DAY_OF_WEEK_LABEL_LIST: readonly string[] = [0, 1, 2, 3, 4, 5, 6].map(
  (d) => DAY_OF_WEEK_LABELS[d],
);
