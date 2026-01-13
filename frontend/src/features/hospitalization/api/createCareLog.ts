import { CreateCareLogDTO, DailyRecord } from "../types";
import { getStoredRecords, setStoredRecords } from "./store";

export const createCareLog = async (hospitalizationId: string, date: string, data: CreateCareLogDTO): Promise<DailyRecord["careLogs"][0]> => {
    // Mock implementation
    await new Promise((resolve) => setTimeout(resolve, 300));

    const records = getStoredRecords(hospitalizationId);
    const targetDate = date.trim();
    let record = records.find(r => r.date === targetDate);
    
    const newLog = {
        ...data,
        id: Math.random().toString(36).substr(2, 9),
    };

    if (record) {
        record.careLogs = [...(record.careLogs || []), newLog];
        const index = records.findIndex(r => r.id === record!.id);
        records[index] = record;
    } else {
        record = {
            id: Math.random().toString(36).substr(2, 9),
            hospitalizationId,
            date: targetDate,
            vitals: [],
            careLogs: [newLog],
            staffNotes: []
        };
        records.push(record);
    }
    
    setStoredRecords(hospitalizationId, records);
    return newLog;
};
