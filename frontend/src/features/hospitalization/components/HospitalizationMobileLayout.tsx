// React/Framework
import { useMemo } from "react";

// External
import { Calendar, FileText, Settings } from "lucide-react";

// Internal
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

// Relative
import { CarePlanSection } from "./CarePlan/CarePlanSection";
import { DailyRecordSection } from "./DailyRecord/DailyRecordSection";
import { HospitalizationPatientHeader } from "./HospitalizationPatientHeader";
import { H_STYLES } from "../styles";

// Types
import type { Hospitalization } from "@/types";
import type {
    CarePlanItem,
    DailyRecord,
    CreateCarePlanDTO,
    UpdateCarePlanDTO,
    CreateVitalDTO,
    CreateCareLogDTO
} from "../types";

interface HospitalizationMobileLayoutProps {
    hospitalization: Hospitalization;
    plans: CarePlanItem[];
    records: DailyRecord[];
    onAddPlan: (plan: CreateCarePlanDTO) => void;
    onUpdatePlan: (planId: string, updates: UpdateCarePlanDTO) => void;
    onDeletePlan: (planId: string) => void;
    onAddVital: (date: string, data: CreateVitalDTO) => void;
    onAddLog: (date: string, data: CreateCareLogDTO) => void;
}

export const HospitalizationMobileLayout = ({
    hospitalization,
    plans,
    records,
    onAddPlan,
    onUpdatePlan,
    onDeletePlan,
    onAddVital,
    onAddLog
}: HospitalizationMobileLayoutProps) => {
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
        <div className={`lg:hidden flex flex-col ${H_STYLES.gap.default}`}>
            <div className="shrink-0">
                <HospitalizationPatientHeader 
                    hospitalization={hospitalization} 
                    currentWeight={latestWeight}
                />
            </div>
            
            <Tabs defaultValue="daily" className="flex flex-col">
                <TabsList className="w-full grid grid-cols-2">
                    <TabsTrigger value="daily" className="text-sm font-bold">
                        <FileText className="h-4 w-4 mr-2" />
                        デイリーカルテ
                    </TabsTrigger>
                    <TabsTrigger value="plan" className="text-sm font-bold">
                        <Settings className="h-4 w-4 mr-2" />
                        プラン管理・詳細
                    </TabsTrigger>
                </TabsList>
                
                <TabsContent value="daily" className="mt-2">
                    <div className="bg-white rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm flex flex-col">
                        <div className={H_STYLES.padding.box}>
                            <DailyRecordSection 
                                records={records}
                                plans={plans}
                                onAddVital={onAddVital}
                                onAddLog={onAddLog}
                            />
                        </div>
                    </div>
                </TabsContent>
                
                <TabsContent value="plan" className="mt-2">
                    <div className={`bg-white rounded-lg border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box} shadow-sm mb-20`}>
                        <div className="flex items-center gap-2 mb-4 text-[#37352F]/60 text-sm">
                            <Calendar className="h-4 w-4" />
                            <span>入院期間: {hospitalization.startDate} 〜 {hospitalization.endDate}</span>
                        </div>
                        <Separator className="mb-4" />
                        <CarePlanSection 
                            plans={plans}
                            onAdd={onAddPlan}
                            onUpdate={onUpdatePlan}
                            onDelete={onDeletePlan}
                        />
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    );
};