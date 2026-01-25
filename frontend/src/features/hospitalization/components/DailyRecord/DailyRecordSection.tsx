import { useState } from "react";
import { format, addDays, subDays } from "date-fns";
import { Activity, Sun, Moon, Coffee, Plus } from "lucide-react";
import { Button } from "../../../../components/ui/button";
import { DateNavigation } from "./DateNavigation";
import { TimingSection } from "./TimingSection";
import { Timeline } from "./Timeline";
import { SimpleNoteForm } from "./SimpleNoteForm";
import { VitalDialog } from "./VitalDialog";
import { LogDialog } from "./LogDialog";
import { TaskCompleteDialog } from "./TaskCompleteDialog";
import { useDailyRecordLogic } from "../../hooks/useDailyRecordLogic";
import { H_STYLES } from "../../styles";
import type { DailyRecord, CarePlanItem, CreateVitalDTO, CreateCareLogDTO, Task } from "../../types";

interface DailyRecordSectionProps {
    records: DailyRecord[];
    plans?: CarePlanItem[];
    onAddVital: (date: string, data: CreateVitalDTO) => void;
    onAddLog: (date: string, data: CreateCareLogDTO) => void;
}

export const DailyRecordSection = ({ records, plans = [], onAddVital, onAddLog }: DailyRecordSectionProps) => {
    const [selectedDate, setSelectedDate] = useState(new Date());
    const currentDateStr = format(selectedDate, "yyyy-MM-dd");
    const { tasks, groupedTasks, timelineItems } = useDailyRecordLogic(records, plans, currentDateStr);
    
    // Dialog States
    const [isVitalOpen, setIsVitalOpen] = useState(false);
    const [isLogOpen, setIsLogOpen] = useState(false);
    const [isTaskCompleteOpen, setIsTaskCompleteOpen] = useState(false);
    const [logType, setLogType] = useState<"food" | "excretion" | "medicine" | "other">("food");
    const [selectedTask, setSelectedTask] = useState<Task | null>(null);

    const handleOpenTaskComplete = (task: Task) => {
        setSelectedTask(task);
        setIsTaskCompleteOpen(true);
    };

    const handleSaveVital = (data: CreateVitalDTO) => {
        onAddVital(format(selectedDate, "yyyy-MM-dd"), data);
    };

    const handleSaveLog = (data: CreateCareLogDTO) => {
        onAddLog(format(selectedDate, "yyyy-MM-dd"), data);
    };

    const openLogDialog = (type: "food" | "excretion" | "medicine" | "other") => {
        setLogType(type);
        setIsLogOpen(true);
    };

    return (
        <div className="flex flex-col gap-4">
            <DateNavigation 
                date={selectedDate} 
                onPrev={() => setSelectedDate(subDays(selectedDate, 1))} 
                onNext={() => setSelectedDate(addDays(selectedDate, 1))} 
            />

            <div className="pr-4">
                <div className="space-y-3">
                    {tasks.length === 0 && (
                        <div className="text-center py-6 text-xs text-[#37352F]/40 bg-[#F7F6F3] rounded border border-dashed border-[rgba(55,53,47,0.16)]">
                            予定なし
                        </div>
                    )}
                    
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

                    <SimpleNoteForm onSave={handleSaveLog} />

                    <Timeline items={timelineItems} /> 
                </div>
            </div>

            <VitalDialog 
                open={isVitalOpen} 
                onOpenChange={setIsVitalOpen} 
                onSave={handleSaveVital} 
            />

            <LogDialog 
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
        </div>
    );
};