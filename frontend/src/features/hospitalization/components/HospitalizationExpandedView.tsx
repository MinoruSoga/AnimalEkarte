// React/Framework
import { memo } from "react";

// External
import { Calendar, FileText } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { formatDate } from "@/lib/format/date";
import { Separator } from "@/components/ui/separator";

// Relative
import { CarePlanTab } from "./CarePlanTab/CarePlanTab";
import { DailyRecordsTab } from "./DailyRecordsTab/DailyRecordsTab";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";
import { H_STYLES } from "../lib/styles";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationExpandedViewProps {
  hospitalization: Hospitalization;
}

export const HospitalizationExpandedView = memo(function HospitalizationExpandedView({
  hospitalization,
}: HospitalizationExpandedViewProps) {
  // Determine the effective discharge date
  const dischargeDate = hospitalization.endDate || todayJSTISO();

  return (
    <div className={`hidden lg:flex flex-col ${H_STYLES.gap.default} w-full h-full`}>
      <div className="shrink-0 w-full z-10 sticky top-0">
        <HospitalizationPatientHeader hospitalization={hospitalization} />
      </div>

      <div className={`flex flex-col ${H_STYLES.gap.default} w-full min-w-0`}>
        {/* Care Plan Tab */}
        <div
          className={`w-full min-w-0 ${C.bgWhite} rounded-lg border ${C.borderMedium} ${H_STYLES.padding.box} overflow-hidden`}
        >
          <div className={`flex items-center gap-1.5 mb-2 ${C.text60} text-sm px-0.5`}>
            <Calendar className={`${ICON.action} shrink-0`} />
            <span className="font-medium truncate">
              入院期間: {formatDate(hospitalization.startDate)} 〜 {formatDate(dischargeDate)}
            </span>
          </div>
          <Separator className="mb-1.5 opacity-50" />
          <CarePlanTab
            hospitalizationId={String(hospitalization.id)}
            petIsDeceased={hospitalization.petIsDeceased}
          />
        </div>

        {/* Daily Records Tab */}
        <div
          className={`w-full min-w-0 ${C.bgWhite} rounded-lg border ${C.borderMedium} flex flex-col overflow-hidden`}
        >
          <div
            className={`px-3 py-2 border-b ${C.borderLight} ${C.bgPage60} flex items-center justify-between shrink-0`}
          >
            <div className={`flex items-center gap-1.5 font-bold ${C.text} text-sm`}>
              <FileText className={`${ICON.action} ${C.textMedicalBlue}`} />
              デイリーカルテ
            </div>
          </div>
          <div className={`${H_STYLES.padding.box} w-full min-w-0`}>
            <DailyRecordsTab
              hospitalizationId={String(hospitalization.id)}
              admissionDate={hospitalization.startDate}
              dischargeDate={dischargeDate}
              petIsDeceased={hospitalization.petIsDeceased}
            />
          </div>
        </div>
      </div>
    </div>
  );
});
