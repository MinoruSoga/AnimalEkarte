import { CarePlanItem, CreateCarePlanDTO } from "../types";
import { getStoredPlans, setStoredPlans } from "./store";

export const createCarePlan = async (hospitalizationId: string, data: CreateCarePlanDTO): Promise<CarePlanItem> => {
    // Simulate API delay
    await new Promise((resolve) => setTimeout(resolve, 500));

    const newPlan: CarePlanItem = {
        ...data,
        id: Math.random().toString(36).substr(2, 9),
        hospitalizationId,
    };

    const currentPlans = getStoredPlans(hospitalizationId);
    setStoredPlans(hospitalizationId, [...currentPlans, newPlan]);

    return newPlan;
};
