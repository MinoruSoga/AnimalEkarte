import { useState } from "react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import {
  LstepCsvImportSection,
  DeliveryStatsSection,
  VisitConversionSection,
} from "../components/LstepAnalyticsSections";
import {
  buildCrossRows,
  currentYearMonth,
  generateMonthOptions,
} from "../components/lstep-analytics-model";
import { useGetLstepDeliveryStats } from "../api/get-lstep-delivery-stats";
import { useGetLstepVisitConversion } from "../api/get-lstep-visit-conversion";

const MONTH_OPTIONS = generateMonthOptions(12);

export function LstepAnalyticsPage() {
  // rerender-lazy-state-init: 初期値の算出は初回のみで足りる
  const [yearMonth, setYearMonth] = useState(() => currentYearMonth());
  const { data, isLoading, isError } = useGetLstepDeliveryStats(yearMonth);
  const {
    data: visitConversion,
    isLoading: isLoadingVisitConversion,
    isError: isErrorVisitConversion,
  } = useGetLstepVisitConversion(yearMonth, 30);

  const crossRows = data ? buildCrossRows(data.rows) : [];

  return (
    <PageLayout
      title="Lステップ分析レポート"
      description="月次配信統計と友だち属性 CSV インポート管理"
    >
      <DeliveryStatsSection
        yearMonth={yearMonth}
        monthOptions={MONTH_OPTIONS}
        rows={crossRows}
        isLoading={isLoading}
        isError={isError}
        onYearMonthChange={setYearMonth}
      />

      <VisitConversionSection
        data={visitConversion}
        isLoading={isLoadingVisitConversion}
        isError={isErrorVisitConversion}
      />

      <LstepCsvImportSection />
    </PageLayout>
  );
}
