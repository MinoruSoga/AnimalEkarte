import { useState } from "react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import {
  buildCrossRows,
  CsvImportSection,
  currentYearMonth,
  DeliveryStatsSection,
  generateMonthOptions,
  VisitConversionSection,
} from "../components/LstepAnalyticsSections";
import { useGetLstepDeliveryStats } from "../api/get-lstep-delivery-stats";
import { useGetLstepVisitConversion } from "../api/get-lstep-visit-conversion";

const MONTH_OPTIONS = generateMonthOptions(12);

export function LstepAnalyticsPage() {
  const [yearMonth, setYearMonth] = useState(currentYearMonth());
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

      <CsvImportSection />
    </PageLayout>
  );
}
