import { addDays, subDays, format } from "date-fns";
import { Hospitalization, CarePlanItem, DailyRecord } from "@/types";

export const MOCK_HOSPITALIZATION: Hospitalization = {
    id: "hosp_1",
    hospitalizationNo: "H2024-001",
    ownerName: "山田 太郎",
    petName: "ポチ",
    species: "犬 (トイプードル)",
    hospitalizationType: "入院",
    startDate: format(subDays(new Date(), 2), "yyyy-MM-dd"),
    endDate: format(addDays(new Date(), 5), "yyyy-MM-dd"),
    status: "入院中",
    cageId: "cage_A01"
};

export const generateMockPlans = (hospitalizationId: string): CarePlanItem[] => [
    {
        id: "plan_1",
        hospitalizationId,
        type: "food",
        name: "ロイヤルカナン消化器サポート",
        description: "30g / ふやかして",
        timing: ["morning", "night"],
        status: "active",
        unitPrice: 150,
        category: "フード"
    },
    {
        id: "plan_2",
        hospitalizationId,
        type: "medicine",
        name: "アンピシリン",
        description: "1錠",
        timing: ["morning", "night"],
        status: "active",
        unitPrice: 500,
        category: "薬剤"
    },
    {
        id: "plan_3",
        hospitalizationId,
        type: "treatment",
        name: "皮下点滴",
        description: "ソルラクト 100ml",
        timing: ["morning"],
        status: "active",
        unitPrice: 1500,
        category: "処置"
    }
];

export const generateMockRecords = (hospitalizationId: string, days: number = 3): DailyRecord[] => {
    const records: DailyRecord[] = [];
    const today = new Date();

    for (let i = 0; i < days; i++) {
        const date = subDays(today, i);
        const dateStr = format(date, "yyyy-MM-dd");

        records.push({
            id: `rec_${i}`,
            hospitalizationId,
            date: dateStr,
            vitals: [
                {
                    id: `vit_${i}_1`,
                    time: "09:00",
                    temperature: 38.5 + (Math.random() * 0.5 - 0.25),
                    heartRate: 120 + Math.floor(Math.random() * 20 - 10),
                    respirationRate: 24,
                    staff: "佐藤"
                },
                {
                    id: `vit_${i}_2`,
                    time: "15:00",
                    temperature: 38.6,
                    heartRate: 118,
                    respirationRate: 22,
                    staff: "田中"
                }
            ],
            careLogs: [
                {
                    id: `log_${i}_1`,
                    time: "09:00",
                    type: "food",
                    status: "completed",
                    value: "100%",
                    staff: "佐藤"
                },
                {
                    id: `log_${i}_2`,
                    time: "09:00",
                    type: "medicine",
                    status: "completed",
                    staff: "佐藤",
                    notes: "投薬スムーズ"
                }
            ],
            staffNotes: []
        });
    }
    return records;
};
