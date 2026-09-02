// React/Framework
import { memo, useState } from "react";

// External
import { Calendar, FileText, Settings } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { formatDate } from "@/lib/format/date";
import { Separator } from "@/components/ui/separator";
import { UnifiedTabs, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";

// Relative
import { CarePlanTab } from "./CarePlanTab/CarePlanTab";
import { DailyRecordsTab } from "./DailyRecordsTab/DailyRecordsTab";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";
import { H_STYLES } from "../lib/styles";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationTabbedViewProps {
    hospitalization: Hospitalization;
}

export const HospitalizationTabbedView = memo(function HospitalizationTabbedView({
    hospitalization,
}: HospitalizationTabbedViewProps) {
    const [activeTab, setActiveTab] = useState<"daily" | "plan">("daily");

    // Determine the effective discharge date
    const dischargeDate = hospitalization.endDate || todayJSTISO();

    const tabItems = [
      { value: "daily", label: <span className="flex items-center"><FileText className={`${ICON.action} mr-2`} />デイリーカルテ</span> },
      { value: "plan", label: <span className="flex items-center"><Settings className={`${ICON.action} mr-2`} />プラン管理・詳細</span> },
    ] as const;

    const handleTabChange = (tab: string) => {
      if (tab === "daily" || tab === "plan") {
        setActiveTab(tab);
      }
    };

    return (
        <div className={`lg:hidden flex flex-col ${H_STYLES.gap.default}`}>
            <div className="shrink-0">
                <HospitalizationPatientHeader
                    hospitalization={hospitalization}
                />
            </div>

            <UnifiedTabs items={tabItems} value={activeTab} onValueChange={handleTabChange} className="flex flex-col">
                <UnifiedTabsContent value="daily" className="mt-2">
                    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} flex flex-col`}>
                        <div className={H_STYLES.padding.box}>
                            <DailyRecordsTab
                                hospitalizationId={String(hospitalization.id)}
                                admissionDate={hospitalization.startDate}
                                dischargeDate={dischargeDate}
                                petIsDeceased={hospitalization.petIsDeceased}
                            />
                        </div>
                    </div>
                </UnifiedTabsContent>

                <UnifiedTabsContent value="plan" className="mt-2">
                    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} ${H_STYLES.padding.box} mb-20`}>
                        <div className={`flex items-center gap-2 mb-4 ${C.text60} text-sm`}>
                            <Calendar className={ICON.action} />
                            <span>入院期間: {formatDate(hospitalization.startDate)} 〜 {formatDate(dischargeDate)}</span>
                        </div>
                        <Separator className="mb-4" />
                        <CarePlanTab
                            hospitalizationId={String(hospitalization.id)}
                            petIsDeceased={hospitalization.petIsDeceased}
                        />
                    </div>
                </UnifiedTabsContent>
            </UnifiedTabs>
        </div>
    );
});
