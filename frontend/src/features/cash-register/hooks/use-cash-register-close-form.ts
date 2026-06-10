import { useState, useCallback } from "react";

import { todayJSTISO } from "@/lib/jst-date";

export function useCashRegisterCloseForm() {
  const today = todayJSTISO();
  const [date, setDate] = useState<string>(today);
  const [period, setPeriod] = useState<"am" | "pm">("am");
  const [previewEnabled, setPreviewEnabled] = useState(false);

  const handleDateChange = useCallback((value: string) => {
    setDate(value);
    setPreviewEnabled(false);
  }, []);

  const handlePeriodChange = useCallback((value: "am" | "pm") => {
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
