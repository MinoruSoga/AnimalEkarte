import { axios } from "@/lib/axios";

export const exportMonthlyCSV = async (year: number, month: number): Promise<void> => {
  const { data } = await axios.get("/v1/reports/monthly/csv", {
    params: { year, month },
    responseType: "blob",
  });
  const url = URL.createObjectURL(data as Blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `monthly-report-${year}-${String(month).padStart(2, "0")}.csv`;
  a.click();
  URL.revokeObjectURL(url);
};
