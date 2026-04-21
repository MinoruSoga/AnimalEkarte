import { useNavigate } from "react-router";
import { Clock } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { ResourceClosingSettings } from "@/types/generated/models";
import { useGetClosingSettings } from "../api/get-closing-settings";
import { StandardClosingTimeSection } from "../components/StandardClosingTimeSection";
import { SpecialPeriodSection } from "../components/SpecialPeriodSection";
import { HolidaySection } from "../components/HolidaySection";

export function ClosingSettingsPage() {
  const navigate = useNavigate();
  const { data, isLoading, isError } = useGetClosingSettings();

  return (
    <PageLayout
      title="締め時間設定"
      icon={<Clock className={`${ICON.page} ${C.text}`} />}
      resource={ResourceClosingSettings}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-3xl"
    >
      {isLoading ? <LoadingFallback /> : null}
      {isError || (!isLoading && !data) ? <ErrorFallback /> : null}
      {!isLoading && data ? (
        <div className="space-y-6">
          <StandardClosingTimeSection settings={data.settings} />
          <SpecialPeriodSection periods={data.special_periods} />
          <HolidaySection holidays={data.holidays} />
        </div>
      ) : null}
    </PageLayout>
  );
}
