// React/Framework
import { useMemo } from "react";

// Types
import type { DailyRecord, CarePlanItem, Task, TimelineItem } from "../types";

export const useDailyRecordLogic = (records: DailyRecord[], plans: CarePlanItem[], currentDateStr: string) => {
    const currentRecord = useMemo(() => {
        const target = currentDateStr.trim();
        return records.find(r => r.date === target);
    }, [records, currentDateStr]);

    const tasks = useMemo((): Task[] => {
        return plans.flatMap(plan => 
            (plan.timing || []).map((timing: string) => ({
                planId: plan.id,
                timing,
                type: plan.type,
                name: plan.name,
                description: plan.description,
                masterId: plan.masterId,
                completedLog: currentRecord?.careLogs?.find(
                    log => log.notes?.includes(plan.name) && log.type === plan.type && log.status === "completed"
                )
            }))
        );
    }, [plans, currentRecord]);

    const groupedTasks = useMemo(() => ({
        morning: tasks.filter(t => t.timing === 'morning'),
        noon: tasks.filter(t => t.timing === 'noon'),
        night: tasks.filter(t => t.timing === 'night'),
    }), [tasks]);

    const timelineItems = useMemo((): TimelineItem[] => {
        if (!currentRecord) return [];
        
        const vitals: TimelineItem[] = (currentRecord.vitals || []).map(v => ({
            kind: 'vital',
            time: v.time,
            staff: v.staff,
            temperature: v.temperature,
            heartRate: v.heartRate,
            respirationRate: v.respirationRate,
            weight: v.weight,
            notes: v.notes
        }));

        const logs: TimelineItem[] = (currentRecord.careLogs || []).map(c => ({
            kind: 'log',
            time: c.time,
            type: c.type,
            value: c.value,
            notes: c.notes,
            staff: c.staff
        }));

        const notes: TimelineItem[] = (currentRecord.staffNotes || []).map(n => ({
            kind: 'note',
            time: n.time,
            content: n.content,
            staff: n.staff
        }));

        return [...vitals, ...logs, ...notes].sort((a, b) => b.time.localeCompare(a.time));
    }, [currentRecord]);

    return { currentRecord, tasks, groupedTasks, timelineItems };
};
