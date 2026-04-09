import { C, ICON } from "@/lib/design-tokens";
import { lazy, memo, Suspense, useCallback, useState } from "react";
import { format, addDays, subDays } from "date-fns";
import { Activity, Sun, Moon, Coffee, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DailyRecordDateNav } from "@/features/hospitalization/components/DailyRecord/DailyRecordDateNav";
import { TimingSection } from "@/features/hospitalization/components/DailyRecord/TimingSection";
import { DailyRecordTimeline } from "@/features/hospitalization/components/DailyRecord/DailyRecordTimeline";
import { DailyCareNoteForm } from "@/features/hospitalization/components/DailyRecord/DailyCareNoteForm";
const VitalDialog = lazy(() => import("@/features/hospitalization/components/DailyRecord/VitalDialog").then((m) => ({ default: m.VitalDialog })));
const DailyCareLogDialog = lazy(() => import("@/features/hospitalization/components/DailyRecord/DailyCareLogDialog").then((m) => ({ default: m.DailyCareLogDialog })));
const TaskCompleteDialog = lazy(() => import("@/features/hospitalization/components/DailyRecord/TaskCompleteDialog").then((m) => ({ default: m.TaskCompleteDialog })));
import { useDailyRecordLogic } from "@/features/hospitalization/hooks/use-daily-record-logic";
import { H_STYLES } from "@/features/hospitalization/styles";
import type { DailyRecord, CarePlanItem, CreateVitalDTO, CreateCareLogDTO, Task } from "@/features/hospitalization/types";

const EMPTY_PLANS: CarePlanItem[] = [];

interface DailyRecordSectionProps {
    records: DailyRecord[];
    plans?: CarePlanItem[];
    onAddVital: (date: string, data: CreateVitalDTO) => void;
    onAddLog: (date: string, data: CreateCareLogDTO) => void;
}

export const DailyRecordSection = memo(function DailyRecordSection({ records, plans = EMPTY_PLANS, onAddVital, onAddLog }: DailyRecordSectionProps) {
    const [selectedDate, setSelectedDate] = useState(() => new Date());
    const currentDateStr = format(selectedDate, "yyyy-MM-dd");
    const { tasks, groupedTasks, timelineItems } = useDailyRecordLogic(records, plans, currentDateStr);
    
    // Dialog States
    const [isVitalOpen, setIsVitalOpen] = useState(false);
    const [isLogOpen, setIsLogOpen] = useState(false);
    const [isTaskCompleteOpen, setIsTaskCompleteOpen] = useState(false);
    const [logType, setLogType] = useState<"food" | "excretion" | "medicine" | "other">("food");
    const [selectedTask, setSelectedTask] = useState<Task | null>(null);

    const handleOpenTaskComplete = useCallback((task: Task) => {
        setSelectedTask(task);
        setIsTaskCompleteOpen(true);
    }, []);

    const handleSaveVital = useCallback((data: CreateVitalDTO) => {
        onAddVital(format(selectedDate, "yyyy-MM-dd"), data);
    }, [onAddVital, selectedDate]);

    const handleSaveLog = useCallback((data: CreateCareLogDTO) => {
        onAddLog(format(selectedDate, "yyyy-MM-dd"), data);
    }, [onAddLog, selectedDate]);

    const openLogDialog = (type: "food" | "excretion" | "medicine" | "other") => {
        setLogType(type);
        setIsLogOpen(true);
    };

    return (
        <div className="flex flex-col gap-4">
            <DailyRecordDateNav
                date={selectedDate}
                onPrev={() => setSelectedDate(subDays(selectedDate, 1))}
                onNext={() => setSelectedDate(addDays(selectedDate, 1))}
            />

            <div className="pr-4">
                <div className="space-y-3">
                    {tasks.length === 0 ? (
                        <div className={`text-center py-6 text-xs ${C.text40} ${C.bgPage} rounded border border-dashed ${C.borderMedium}`}>
                            予定なし
                        </div>
                    ) : null}
                    
                    <TimingSection
                        title="朝の予定"
                        icon={Sun}
                        tasks={groupedTasks.morning}
                        colorClass={C.textDiscount}
                        onTaskClick={handleOpenTaskComplete}
                    />
                    <TimingSection
                        title="昼の予定"
                        icon={Coffee}
                        tasks={groupedTasks.noon}
                        colorClass={C.textNotice}
                        onTaskClick={handleOpenTaskComplete}
                    />
                    <TimingSection
                        title="夜の予定"
                        icon={Moon}
                        tasks={groupedTasks.night}
                        colorClass={C.textStatusPurple}
                        onTaskClick={handleOpenTaskComplete}
                    />
                </div>

                <div className={`mt-1 pt-1 border-t ${C.borderLight}`}>
                    <div className="flex items-center justify-between mb-2">
                        <h3 className={`font-bold ${C.text} flex items-center gap-2 ${H_STYLES.text.lg}`}>
                            <Activity className={ICON.page} />
                            その他・記録履歴
                        </h3>
                        <div className="flex gap-2">
                            <Button variant="outline" size="sm" className={`gap-1 bg-white ${H_STYLES.button.action}`} onClick={() => setIsVitalOpen(true)}>
                                <Plus className={H_STYLES.button.icon} /> バイタル
                            </Button>
                            <Button variant="outline" size="sm" className={`gap-1 bg-white ${H_STYLES.button.action}`} onClick={() => openLogDialog("excretion")}>
                                <Plus className={H_STYLES.button.icon} /> 排泄
                            </Button>
                            <Button variant="outline" size="sm" className={`gap-1 bg-white ${H_STYLES.button.action}`} onClick={() => openLogDialog("other")}>
                                <Plus className={H_STYLES.button.icon} /> メモ
                            </Button>
                        </div>
                    </div>

                    <DailyCareNoteForm onSave={handleSaveLog} />

                    <DailyRecordTimeline items={timelineItems} />
                </div>
            </div>

            <Suspense fallback={null}>
              <VitalDialog
                  open={isVitalOpen}
                  onOpenChange={setIsVitalOpen}
                  onSave={handleSaveVital}
              />
              <DailyCareLogDialog
                  open={isLogOpen}
                  onOpenChange={setIsLogOpen}
                  type={logType}
                  onSave={handleSaveLog}
              />
              <TaskCompleteDialog
                  open={isTaskCompleteOpen}
                  onOpenChange={setIsTaskCompleteOpen}
                  task={selectedTask}
                  onConfirm={handleSaveLog}
              />
            </Suspense>
        </div>
    );
});
