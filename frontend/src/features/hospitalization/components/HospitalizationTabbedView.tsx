// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { Calendar, FileText, Settings } from "lucide-react";

// Internal
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

// Relative
import { CarePlanTab } from "./CarePlanTab/CarePlanTab";
import { DailyRecordsTab } from "./DailyRecordsTab/DailyRecordsTab";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";
import { H_STYLES } from "../styles";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationTabbedViewProps {
    hospitalization: Hospitalization;
}

export const HospitalizationTabbedView = memo(function HospitalizationTabbedView({
    hospitalization,
}: HospitalizationTabbedViewProps) {
    // Determine the effective discharge date
    const dischargeDate = hospitalization.endDate || new Date().toISOString().split("T")[0];

    return (
        <div className={`lg:hidden flex flex-col ${H_STYLES.gap.default}`}>
            <div className="shrink-0">
                <HospitalizationPatientHeader
                    hospitalization={hospitalization}
                />
            </div>

            <Tabs defaultValue="daily" className="flex flex-col">
                <TabsList className="w-full grid grid-cols-2">
                    <TabsTrigger value="daily" className="text-sm font-bold">
                        <FileText className={`${ICON.action} mr-2`} />
                        デイリーカルテ
                    </TabsTrigger>
                    <TabsTrigger value="plan" className="text-sm font-bold">
                        <Settings className={`${ICON.action} mr-2`} />
                        プラン管理・詳細
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="daily" className="mt-2">
                    <div className={`bg-white rounded-lg border ${C.borderMedium} shadow-sm flex flex-col`}>
                        <div className={H_STYLES.padding.box}>
                            <DailyRecordsTab
                                hospitalizationId={String(hospitalization.id)}
                                admissionDate={hospitalization.startDate}
                                dischargeDate={dischargeDate}
                            />
                        </div>
                    </div>
                </TabsContent>

                <TabsContent value="plan" className="mt-2">
                    <div className={`bg-white rounded-lg border ${C.borderMedium} ${H_STYLES.padding.box} shadow-sm mb-20`}>
                        <div className={`flex items-center gap-2 mb-4 ${C.text60} text-sm`}>
                            <Calendar className={ICON.action} />
                            <span>入院期間: {hospitalization.startDate} 〜 {dischargeDate}</span>
                        </div>
                        <Separator className="mb-4" />
                        <CarePlanTab
                            hospitalizationId={String(hospitalization.id)}
                        />
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    );
});
