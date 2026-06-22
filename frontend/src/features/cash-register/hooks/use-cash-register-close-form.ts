import { useState, useCallback } from "react";

import { todayJSTISO } from "@/lib/jst-date";
import type { CashRegisterPeriod } from "../constants";

export function useCashRegisterCloseForm() {
  const today = todayJSTISO();
  const [date, setDate] = useState<string>(today);
  const [period, setPeriod] = useState<CashRegisterPeriod>("am");
  const [previewEnabled, setPreviewEnabled] = useState(false);

  const handleDateChange = useCallback((value: string) => {
    setDate(value);
    setPreviewEnabled(false);
  }, []);

  const handlePeriodChange = useCallback((value: CashRegisterPeriod) => {
    setPeriod(value);
    setPreviewEnabled(false);
  }, []);

  const enablePreview = useCallback(() => {
    if (date) setPreviewEnabled(true);
  }, [date]);

  return {
    date,
    period,
    previewEnabled,
    handleDateChange,
    handlePeriodChange,
    enablePreview,
  };
}
