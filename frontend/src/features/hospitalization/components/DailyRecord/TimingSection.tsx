// External
import { C, ICON } from "@/lib/design-tokens";
import { Check, Clock } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

// Types
import type { LucideIcon } from "lucide-react";
import type { Task } from "@/features/hospitalization/types";

interface TimingSectionProps {
    title: string;
    icon: LucideIcon;
    tasks: Task[];
    colorClass: string;
    onTaskClick: (task: Task) => void;
}

export function TimingSection({ title, icon: Icon, tasks, colorClass, onTaskClick }: TimingSectionProps) {
    if (tasks.length === 0) return null;

    return (
        <div className={H_STYLES.layout.card_mb}>
            <div className={`flex items-center ${H_STYLES.gap.default} mb-1 font-bold ${colorClass}`}>
                <Icon className={ICON.page} />
                <span className={H_STYLES.text.lg}>{title}</span>
                <span className={`${H_STYLES.text.sm} font-normal opacity-70 ml-2`}>({tasks.length}件の予定)</span>
            </div>
            <div className={`flex flex-col ${H_STYLES.gap.tight}`}>
                {tasks.map((task, idx) => (
                    <Card key={idx} className={`
                        ${H_STYLES.padding.card} border transition-all duration-200
                        ${task.completedLog
                            ? `${C.bgPage} border-transparent opacity-80`
                            : `bg-white ${C.borderMedium} hover:shadow-sm ${C.hoverBorderMedicalBlue50}`}
                    `}>
                        <div className="flex items-center justify-between">
                            <div className={`flex items-center ${H_STYLES.gap.default}`}>
                                <div className={`
                                    w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0
                                    ${task.completedLog ? "bg-green-100 text-green-600" : "bg-gray-100 text-gray-400"}
                                `}>
                                    {task.completedLog ? <Check className={ICON.page} /> : <Clock className={ICON.page} />}
                                </div>
                                <div>
                                    <div className={`font-bold ${H_STYLES.text.base} leading-tight ${task.completedLog ? `${C.text60} line-through` : C.text}`}>
                                        {task.name}
                                    </div>
                                    <div className={`${H_STYLES.text.sm} ${C.text60} leading-tight mt-0.5`}>
                                        {task.description}
                                    </div>
                                </div>
                            </div>
                            
                            {task.completedLog ? (
                                <div className="text-right flex-shrink-0">
                                    <div className={`${H_STYLES.text.sm} font-bold text-green-600 bg-green-50 px-2 py-0.5 rounded`}>
                                        {task.completedLog.time} 実施済
                                    </div>
                                    <div className={`${H_STYLES.text.xs} ${C.text40} mt-0.5`}>
                                        {task.completedLog.staff}
                                    </div>
                                </div>
                            ) : (
                                <Button 
                                    size="sm" 
                                    variant="primary"
                                    className={`shadow-sm flex-shrink-0 ${H_STYLES.button.action}`}
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
}
