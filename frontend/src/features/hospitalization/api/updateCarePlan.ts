import { CarePlanItem, UpdateCarePlanDTO } from "../types";
import { getAllStoredPlans, setStoredPlans } from "./store";

export const updateCarePlan = async (id: string, data: UpdateCarePlanDTO): Promise<CarePlanItem> => {
    // Mock implementation
    await new Promise((resolve) => setTimeout(resolve, 500));

    const allPlansMap = getAllStoredPlans();
    let updatedPlan: CarePlanItem | null = null;
    let foundHospId: string | null = null;

    for (const [hospId, plans] of Object.entries(allPlansMap)) {
        const planIndex = plans.findIndex(p => p.id === id);
        if (planIndex !== -1) {
            updatedPlan = { ...plans[planIndex], ...data };
            plans[planIndex] = updatedPlan;
            foundHospId = hospId;
            break;
        }
    }

    if (foundHospId && updatedPlan) {
        setStoredPlans(foundHospId, allPlansMap[foundHospId]);
        return updatedPlan;
    }

    throw new Error("Plan not found");
};
