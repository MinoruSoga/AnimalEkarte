export type NextScheduleType = "3weeks" | "4weeks" | "6weeks" | "6months" | "1year" | "other";

/**
 * 次回予定の既定間隔（EMR-430）。
 * 予防接種の独立画面とカルテのタブ（予防接種・健診）で同じ値を使うための単一の定義。
 * ここを変えると全フォームの初期値が揃って変わる。
 */
export const DEFAULT_NEXT_SCHEDULE_TYPE = "1year" as const satisfies NextScheduleType;
