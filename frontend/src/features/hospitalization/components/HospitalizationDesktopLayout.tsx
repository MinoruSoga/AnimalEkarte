import { Calendar, FileText } from "lucide-react";
import { useMemo } from "react";
import { Separator } from "../../../components/ui/separator";
import { CarePlanSection } from "./CarePlan/CarePlanSection";
import { DailyRecordSection } from "./DailyRecord/DailyRecordSection";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";
import { Hospitalization } from "../../../types";
import { H_STYLES } from "../styles";
import { 
    CarePlanItem, 
    DailyRecord, 
    CreateCarePlanDTO, 
    UpdateCarePlanDTO,
    CreateVitalDTO, 
    CreateCareLogDTO 
} from "../types";

interface HospitalizationDesktopLayoutProps {
    hospitalization: Hospitalization;
    plans: CarePlanItem[];
    records: DailyRecord[];
    onAddPlan: (plan: CreateCarePlanDTO) => void;
    onUpdatePlan: (planId: string, updates: UpdateCarePlanDTO) => void;
    onDeletePlan: (planId: string) => void;
    onAddVital: (date: string, data: CreateVitalDTO) => void;
    onAddLog: (date: string, data: CreateCareLogDTO) => void;
}

export const HospitalizationDesktopLayout = ({
    hospitalization,
    plans,
    records,
    onAddPlan,
    onUpdatePlan,
    onDeletePlan,
    onAddVital,
    onAddLog
}: HospitalizationDesktopLayoutProps) => {
    const latestWeight = useMemo(() => {
        // Find latest weight from records
        const allVitals = records.flatMap(r => 
            (r.vitals || []).map(v => ({ ...v, date: r.date }))
        );
        const weightVitals = allVitals.filter(v => v.weight !== undefined);
        weightVitals.sort((a, b) => {
            if (a.date !== b.date) return b.date.localeCompare(a.date);
            return b.time.localeCompare(a.time);
        });
        const latest = weightVitals[0];
        return latest ? `${latest.weight}kg` : undefined;
    }, [records]);

    return (
        <div className={`hidden lg:flex flex-col ${H_STYLES.gap.default} w-full h-full`}>
            <div className="shrink-0 w-full z-10 sticky top-0">
                <HospitalizationPatientHeader 
                    hospitalization={hospitalization} 
                    currentWeight={latestWeight}
                />
            </div>

            <div className={`flex flex-col ${H_STYLES.gap.default} w-full min-w-0`}>
                {/* Care Plan Section */}
                <div className={`w-full min-w-0 bg-white rounded-lg border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box} shadow-sm overflow-hidden`}>
                    <div className="flex items-center gap-1.5 mb-2 text-[#37352F]/60 text-sm px-0.5">
                        <Calendar className="h-3.5 w-3.5 shrink-0" />
                        <span className="font-medium truncate">入院期間: {hospitalization.startDate} 〜 {hospitalization.endDate}</span>
                    </div>
                    <Separator className="mb-1.5 opacity-50" />
                    <CarePlanSection 
                        plans={plans}
                        onAdd={onAddPlan}
                        onUpdate={onUpdatePlan}
                        onDelete={onDeletePlan}
                    />
                </div>

                {/* Daily Records Section */}
                <div className="w-full min-w-0 bg-white rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm flex flex-col overflow-hidden">
                    <div className="px-3 py-2 border-b border-[rgba(55,53,47,0.09)] bg-gray-50/50 flex items-center justify-between shrink-0">
                        <div className="flex items-center gap-1.5 font-bold text-[#37352F] text-sm">
                            <FileText className="h-4 w-4 text-[#2EAADC]" />
                            デイリーカルテ
                        </div>
                    </div>
                    <div className={`${H_STYLES.padding.box} w-full min-w-0`}>
                        <DailyRecordSection 
                            records={records}
                            plans={plans}
                            onAddVital={onAddVital}
                            onAddLog={onAddLog}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
};