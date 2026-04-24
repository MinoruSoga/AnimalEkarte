import { useState, useCallback } from "react";
import { C, STYLE } from "@/lib/design-tokens";
import { useGetCashRegisterCloses } from "../api/get-cash-register-closes";

export function CashRegisterHistoryPage() {
  const now = new Date();
  const [year, setYear] = useState<number>(now.getFullYear());
  const [month, setMonth] = useState<number>(now.getMonth() + 1);

  const { data, isLoading, isError } = useGetCashRegisterCloses({ year, month });

  const handleYearChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setYear(Number(e.target.value));
  }, []);

  const handleMonthChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setMonth(Number(e.target.value));
  }, []);

  const yearOptions = Array.from({ length: 5 }, (_, i) => now.getFullYear() - i);

  return (
    <div className="flex-1 overflow-y-auto px-6 py-6 space-y-6">
      <h1 className={`text-xl font-bold ${C.text}`}>締め履歴</h1>

      {/* 絞り込み */}
      <div className="flex flex-wrap gap-3 items-center">
        <div>
          <label htmlFor="hist_year" className={`${STYLE.formLabel} mr-2`}>
            年
          </label>
          <select
            id="hist_year"
            value={year}
            onChange={handleYearChange}
            className={`${STYLE.formInput} rounded-[4px] border px-3 inline-block w-auto`}
          >
            {yearOptions.map((y) => (
              <option key={y} value={y}>
                {y}年
              </option>
            ))}
          </select>
        </div>
        <div>
          <label htmlFor="hist_month" className={`${STYLE.formLabel} mr-2`}>
            月
          </label>
          <select
            id="hist_month"
            value={month}
            onChange={handleMonthChange}
            className={`${STYLE.formInput} rounded-[4px] border px-3 inline-block w-auto`}
          >
            {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
              <option key={m} value={m}>
                {m}月
              </option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-8">
          <p className={`text-base ${C.text50}`}>読み込み中...</p>
        </div>
      ) : isError ? (
        <div className="flex items-center justify-center py-8">
          <p className={`text-base ${C.danger}`}>データの取得に失敗しました</p>
        </div>
      ) : (
        <div className={`${C.bgWhite} rounded-lg border ${C.borderLight}`}>
          {data && data.data.length > 0 ? (
            <table className="w-full text-base">
              <thead>
                <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
                  <th className={`text-left px-4 py-2 font-medium ${C.text70}`}>日付</th>
                  <th className={`text-left px-4 py-2 font-medium ${C.text70}`}>区分</th>
                  <th className={`text-right px-4 py-2 font-medium ${C.text70}`}>理論現金</th>
                  <th className={`text-right px-4 py-2 font-medium ${C.text70}`}>実際の現金</th>
                  <th className={`text-right px-4 py-2 font-medium ${C.text70}`}>差額</th>
                  <th className={`text-left px-4 py-2 font-medium ${C.text70}`}>担当者</th>
                  <th className={`text-left px-4 py-2 font-medium ${C.text70}`}>締め時刻</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((close) => {
                  const diff = (close.actualCash ?? 0) - (close.theoreticalCash ?? 0);
                  return (
                    <tr
                      key={close.id}
                      className={`border-b ${C.borderLight} ${STYLE.tableRow}`}
                    >
                      <td className={`px-4 py-3 ${C.text}`}>{close.closeDate}</td>
                      <td className={`px-4 py-3 ${C.text}`}>
                        {close.period === "am" ? "午前" : "午後"}
                      </td>
                      <td className={`px-4 py-3 text-right ${C.text}`}>
                        ¥{(close.theoreticalCash ?? 0).toLocaleString()}
                      </td>
                      <td className={`px-4 py-3 text-right ${C.text}`}>
                        ¥{(close.actualCash ?? 0).toLocaleString()}
                      </td>
                      <td
                        className={`px-4 py-3 text-right font-medium ${diff === 0 ? C.textStatusGreen : C.danger}`}
                      >
                        {diff >= 0 ? "+" : ""}
                        {diff.toLocaleString()}
                      </td>
                      <td className={`px-4 py-3 ${C.text}`}>
                        {close.closedByStaffName ?? "—"}
                      </td>
                      <td className={`px-4 py-3 ${C.text60}`}>
                        {close.closedAt
                          ? new Date(close.closedAt).toLocaleString("ja-JP", {
                              month: "2-digit",
                              day: "2-digit",
                              hour: "2-digit",
                              minute: "2-digit",
                            })
                          : "—"}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          ) : (
            <p className={`text-base ${C.text50} py-8 text-center`}>
              締め履歴がありません
            </p>
          )}
        </div>
      )}
    </div>
  );
}
