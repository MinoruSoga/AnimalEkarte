import { axios } from "@/lib/axios";
import type { MonthlyReportParams } from "./get-monthly-report";

function toRequestParams(params: MonthlyReportParams) {
  if (params.mode === "month") {
    return { year: params.year, month: params.month };
  }
  return { start_date: params.startDate, end_date: params.endDate };
}

function toDownloadFilename(params: MonthlyReportParams): string {
  if (params.mode === "month") {
    return `monthly-report-${params.year}-${String(params.month).padStart(2, "0")}.csv`;
  }
  return `monthly-report-${params.startDate}-${params.endDate}.csv`;
}

export const exportMonthlyCSV = async (params: MonthlyReportParams): Promise<void> => {
  const { data } = await axios.get("/v1/reports/monthly/csv", {
    params: toRequestParams(params),
    responseType: "blob",
  });
  const url = URL.createObjectURL(data as Blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = toDownloadFilename(params);
  a.click();
  URL.revokeObjectURL(url);
};
