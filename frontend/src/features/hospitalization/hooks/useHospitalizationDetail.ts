import { useState, useEffect } from "react";
import { toast } from "sonner";
import { 
    CarePlanItem, 
    DailyRecord, 
    CreateCarePlanDTO, 
    UpdateCarePlanDTO,
    CreateVitalDTO,
    CreateCareLogDTO
} from "../types";
import { 
    getHospitalization, 
    createCarePlan,
    updateCarePlan,
    deleteCarePlan,
    createVital,
    createCareLog,
    dischargeHospitalization as apiDischarge,
    MOCK_HOSPITALIZATION 
} from "../api";

export const useHospitalizationDetail = (hospitalizationId?: string) => {
    const [hospitalization, setHospitalization] = useState(MOCK_HOSPITALIZATION);
    const [plans, setPlans] = useState<CarePlanItem[]>([]);
    const [records, setRecords] = useState<DailyRecord[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        if (hospitalizationId) {
            setIsLoading(true);
            getHospitalization(hospitalizationId)
                .then(data => {
                    setPlans(data.plans);
                    setRecords(data.records);
                    setHospitalization(data.hospitalization);
                })
                .catch(() => {
                    toast.error("データの取得に失敗しました");
                })
                .finally(() => {
                    setIsLoading(false);
                });
        }
    }, [hospitalizationId]);

    const handleAddPlan = async (plan: CreateCarePlanDTO) => {
        if (!hospitalizationId) return;
        try {
            const newPlan = await createCarePlan(hospitalizationId, plan);
            setPlans(prev => [...prev, newPlan]);
            toast.success("ケアプランを作成しました");
        } catch {
            toast.error("ケアプランの作成に失敗しました");
        }
    };

    const handleUpdatePlan = async (planId: string, updates: UpdateCarePlanDTO) => {
        try {
            await updateCarePlan(planId, updates);
            setPlans(prev => prev.map(p => p.id === planId ? { ...p, ...updates } : p));
            toast.success("ケアプランを更新しました");
        } catch {
            toast.error("ケアプランの更新に失敗しました");
        }
    };

    const handleDeletePlan = async (planId: string) => {
        try {
            await deleteCarePlan(planId);
            setPlans(prev => prev.filter(p => p.id !== planId));
            toast.success("ケアプランを削除しました");
        } catch {
            toast.error("ケアプランの削除に失敗しました");
        }
    };

    const handleAddVital = async (date: string, data: CreateVitalDTO) => {
        if (!hospitalizationId) return;
        try {
            const newVital = await createVital(hospitalizationId, date, data);
            
            setRecords(prev => {
                const targetDate = date.trim();
                const exists = prev.some(r => r.date === targetDate);
                
                if (exists) {
                    return prev.map(record => {
                        if (record.date === targetDate) {
                            return {
                                ...record,
                                vitals: [...(record.vitals || []), newVital]
                            };
                        }
                        return record;
                    });
                } else {
                    return [...prev, {
                        id: Math.random().toString(36).substr(2, 9),
                        hospitalizationId,
                        date: targetDate,
                        vitals: [newVital],
                        careLogs: [],
                        staffNotes: []
                    }];
                }
            });
            toast.success("バイタルを記録しました");
        } catch {
            toast.error("バイタルの記録に失敗しました");
        }
    };

    const handleAddLog = async (date: string, data: CreateCareLogDTO) => {
        if (!hospitalizationId) return;
        try {
            const newLog = await createCareLog(hospitalizationId, date, data);

            setRecords(prev => {
                const targetDate = date.trim();
                const exists = prev.some(r => r.date === targetDate);
                
                if (exists) {
                    return prev.map(record => {
                        if (record.date === targetDate) {
                            return {
                                ...record,
                                careLogs: [...(record.careLogs || []), newLog]
                            };
                        }
                        return record;
                    });
                } else {
                    return [...prev, {
                        id: Math.random().toString(36).substr(2, 9),
                        hospitalizationId,
                        date: targetDate,
                        vitals: [],
                        careLogs: [newLog],
                        staffNotes: []
                    }];
                }
            });
            toast.success("記録を追加しました");
        } catch {
            toast.error("記録の追加に失敗しました");
        }
    };

    const dischargeHospitalization = async () => {
        if (hospitalizationId && hospitalization) {
            try {
                const updatedHospitalization = await apiDischarge(hospitalizationId);
                setHospitalization(updatedHospitalization);
                toast.success("退院処理が完了しました");
                return true;
            } catch {
                toast.error("退院処理に失敗しました");
                return false;
            }
        }
        return false;
    };

    return {
        hospitalization,
        plans,
        records,
        isLoading,
        handleAddPlan,
        handleUpdatePlan,
        handleDeletePlan,
        handleAddVital,
        handleAddLog,
        dischargeHospitalization
    };
};
