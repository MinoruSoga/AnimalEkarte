// External
import { Check, Clock } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

// Relative
import { H_STYLES } from "../../styles";

// Types
import type { LucideIcon } from "lucide-react";
import type { Task } from "../../types";

interface TimingSectionProps {
    title: string;
    icon: LucideIcon;
    tasks: Task[];
    colorClass: string;
    onTaskClick: (task: Task) => void;
}

export const TimingSection = ({ title, icon: Icon, tasks, colorClass, onTaskClick }: TimingSectionProps) => {
    if (tasks.length === 0) return null;

    return (
        <div className={H_STYLES.layout.card_mb}>
            <div className={`flex items-center ${H_STYLES.gap.default} mb-1 font-bold ${colorClass}`}>
                <Icon className="h-5 w-5" />
                <span className={H_STYLES.text.lg}>{title}</span>
                <span className={`${H_STYLES.text.sm} font-normal opacity-70 ml-2`}>({tasks.length}件の予定)</span>
            </div>
            <div className={`flex flex-col ${H_STYLES.gap.tight}`}>
                {tasks.map((task, idx) => (
                    <Card key={idx} className={`
                        ${H_STYLES.padding.card} border transition-all duration-200
                        ${task.completedLog 
                            ? "bg-[#F7F6F3] border-transparent opacity-80" 
                            : "bg-white border-[rgba(55,53,47,0.16)] hover:shadow-sm hover:border-[#2EAADC]/50"}
                    `}>
                        <div className="flex items-center justify-between">
                            <div className={`flex items-center ${H_STYLES.gap.default}`}>
                                <div className={`
                                    w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0
                                    ${task.completedLog ? "bg-green-100 text-green-600" : "bg-gray-100 text-gray-400"}
                                `}>
                                    {task.completedLog ? <Check className="h-5 w-5" /> : <Clock className="h-5 w-5" />}
                                </div>
                                <div>
                                    <div className={`font-bold ${H_STYLES.text.base} leading-tight ${task.completedLog ? "text-[#37352F]/60 line-through" : "text-[#37352F]"}`}>
                                        {task.name}
                                    </div>
                                    <div className={`${H_STYLES.text.sm} text-[#37352F]/60 leading-tight mt-0.5`}>
                                        {task.description}
                                    </div>
                                </div>
                            </div>
                            
                            {task.completedLog ? (
                                <div className="text-right flex-shrink-0">
                                    <div className={`${H_STYLES.text.sm} font-bold text-green-600 bg-green-50 px-2 py-0.5 rounded`}>
                                        {task.completedLog.time} 実施済
                                    </div>
                                    <div className={`${H_STYLES.text.xs} text-[#37352F]/40 mt-0.5`}>
                                        {task.completedLog.staff}
                                    </div>
                                </div>
                            ) : (
                                <Button 
                                    size="sm" 
                                    className={`bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white shadow-sm flex-shrink-0 ${H_STYLES.button.action}`}
                                    onClick={() => onTaskClick(task)}
                                >
                                    実施する
                                </Button>
                            )}
                        </div>
                    </Card>
                ))}
            </div>
        </div>
    );
};
