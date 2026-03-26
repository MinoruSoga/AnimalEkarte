import { lazy, memo, Suspense, useCallback, useState } from "react";
import { format, addDays, subDays } from "date-fns";
import { Activity, Sun, Moon, Coffee, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DailyRecordDateNav } from "./DailyRecordDateNav";
import { TimingSection } from "./TimingSection";
import { DailyRecordTimeline } from "./DailyRecordTimeline";
import { DailyCareNoteForm } from "./DailyCareNoteForm";
const VitalDialog = lazy(() => import("./VitalDialog").then((m) => ({ default: m.VitalDialog })));
const DailyCareLogDialog = lazy(() => import("./DailyCareLogDialog").then((m) => ({ default: m.DailyCareLogDialog })));
const TaskCompleteDialog = lazy(() => import("./TaskCompleteDialog").then((m) => ({ default: m.TaskCompleteDialog })));
import { useDailyRecordLogic } from "../../hooks/use-daily-record-logic";
import { H_STYLES } from "../../styles";
import type { DailyRecord, CarePlanItem, CreateVitalDTO, CreateCareLogDTO, Task } from "../../types";

interface DailyRecordSectionProps {
    records: DailyRecord[];
    plans?: CarePlanItem[];
    onAddVital: (date: string, data: CreateVitalDTO) => void;
    onAddLog: (date: string, data: CreateCareLogDTO) => void;
}

export const DailyRecordSection = memo(function DailyRecordSection({ records, plans = [], onAddVital, onAddLog }: DailyRecordSectionProps) {
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
                        <div className="text-center py-6 text-xs text-[#37352F]/40 bg-[#F7F6F3] rounded border border-dashed border-[rgba(55,53,47,0.16)]">
                            予定なし
                        </div>
                    ) : null}
                    
                    <TimingSection 
                        title="朝の予定" 
                        icon={Sun} 
                        tasks={groupedTasks.morning} 
                        colorClass="text-orange-600" 
                        onTaskClick={handleOpenTaskComplete} 
                    />
                    <TimingSection 
                        title="昼の予定" 
                        icon={Coffee} 
                        tasks={groupedTasks.noon} 
                        colorClass="text-yellow-600" 
                        onTaskClick={handleOpenTaskComplete} 
                    />
                    <TimingSection 
                        title="夜の予定" 
                        icon={Moon} 
                        tasks={groupedTasks.night} 
                        colorClass="text-indigo-600" 
                        onTaskClick={handleOpenTaskComplete} 
                    />
                </div>

                <div className="mt-1 pt-1 border-t border-[rgba(55,53,47,0.09)]">
                    <div className="flex items-center justify-between mb-2">
                        <h3 className={`font-bold text-[#37352F] flex items-center gap-2 ${H_STYLES.text.lg}`}>
                            <Activity className="h-5 w-5" />
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
