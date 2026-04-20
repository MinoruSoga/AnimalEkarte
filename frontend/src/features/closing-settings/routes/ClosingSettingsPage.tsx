import { C } from "@/lib/design-tokens";
import { useGetClosingSettings } from "../api/get-closing-settings";
import { StandardClosingTimeSection } from "../components/StandardClosingTimeSection";
import { SpecialPeriodSection } from "../components/SpecialPeriodSection";
import { HolidaySection } from "../components/HolidaySection";

export function ClosingSettingsPage() {
  const { data, isLoading, isError } = useGetClosingSettings();

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className={`text-base ${C.text50}`}>読み込み中...</p>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className={`text-base ${C.danger}`}>データの取得に失敗しました</p>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto px-6 py-6 space-y-6">
      <h1 className={`text-xl font-bold ${C.text}`}>締め時間設定</h1>
      <StandardClosingTimeSection settings={data.settings} />
      <SpecialPeriodSection periods={data.special_periods} />
      <HolidaySection holidays={data.holidays} />
    </div>
  );
}
