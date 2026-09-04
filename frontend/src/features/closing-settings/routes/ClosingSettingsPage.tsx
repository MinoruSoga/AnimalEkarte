import { useNavigate } from "react-router";
import { Clock } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceClosingSettings } from "@/types/generated/models";
import { useGetClosingSettings } from "../api/get-closing-settings";
import { useGetHolidays } from "../api/holidays";
import { StandardClosingTimeSection } from "../components/StandardClosingTimeSection";
import { SpecialPeriodSection } from "../components/SpecialPeriodSection";
import { HolidaySection } from "../components/HolidaySection";

export function ClosingSettingsPage() {
  const navigate = useNavigate();
  const { canEdit } = usePermission(ResourceClosingSettings);
  const { data, isLoading, isError } = useGetClosingSettings();
  const { data: holidays = [], isLoading: holidaysLoading } = useGetHolidays();

  const loading = isLoading || holidaysLoading;

  return (
    <PageLayout
      title="締め時間設定"
      icon={<Clock className={`${ICON.page} ${C.text}`} />}
      resource={ResourceClosingSettings}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-3xl"
      align="left"
    >
      {loading ? <LoadingFallback /> : null}
      {isError || (!isLoading && !data) ? <ErrorFallback /> : null}
      {!loading && data ? (
        <div className="space-y-6">
          <StandardClosingTimeSection settings={data.settings} canEdit={canEdit} />
          <SpecialPeriodSection periods={data.special_periods} canEdit={canEdit} />
          <HolidaySection holidays={holidays} canEdit={canEdit} />
        </div>
      ) : null}
    </PageLayout>
  );
}
