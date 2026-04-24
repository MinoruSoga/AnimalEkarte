import { useState, useCallback } from "react";

export function useCashRegisterCloseForm() {
  const today = new Date().toISOString().slice(0, 10);
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
