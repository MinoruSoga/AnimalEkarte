export interface MonthDateRange {
  startDate: string;
  endDate: string;
}

// 指定した年月（month は 1-based）の月初・月末を YYYY-MM-DD で返す。
// BE の締め履歴フィルタは start_date <= close_date <= end_date（両端含む・close_date は DATE 型）なので
// end_date は当月末日にする。new Date(year, month, 0) は「翌月の 0 日目 = 当月末日」を指し、
// 12月→翌年の桁上がりも、うるう年 2月（2024→29 / 2026→28）も標準の日付演算で正しく解決する。
export function monthDateRange(year: number, month: number): MonthDateRange {
  const mm = String(month).padStart(2, "0");
  const lastDay = new Date(year, month, 0).getDate();
  return {
    startDate: `${year}-${mm}-01`,
    endDate: `${year}-${mm}-${String(lastDay).padStart(2, "0")}`,
  };
}
