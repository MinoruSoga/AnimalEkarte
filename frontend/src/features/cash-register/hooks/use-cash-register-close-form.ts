import { useState, useCallback } from "react";

import { todayJSTISO } from "@/lib/jst-date";
import type { CashRegisterPeriod } from "../lib/constants";

export function useCashRegisterCloseForm() {
  const today = todayJSTISO();
  const [date, setDate] = useState<string>(today);
  const [period, setPeriod] = useState<CashRegisterPeriod>("am");
  const [previewEnabled, setPreviewEnabled] = useState(false);
  const [previewNonce, setPreviewNonce] = useState(0);

  const handleDateChange = useCallback((value: string) => {
    setDate(value);
    setPreviewEnabled(false);
  }, []);

  const handlePeriodChange = useCallback((value: CashRegisterPeriod) => {
    setPeriod(value);
    setPreviewEnabled(false);
  }, []);

  const enablePreview = useCallback(() => {
    if (!date) return;
    setPreviewEnabled(true);
    setPreviewNonce((n) => n + 1);
  }, [date]);

  return {
    date,
    period,
    previewEnabled,
    previewNonce,
    handleDateChange,
    handlePeriodChange,
    enablePreview,
  };
}
