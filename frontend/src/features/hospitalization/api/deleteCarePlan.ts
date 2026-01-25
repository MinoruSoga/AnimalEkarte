import { getAllStoredPlans, setStoredPlans } from "./store";

export const deleteCarePlan = async (id: string): Promise<void> => {
    // Mock implementation
    await new Promise((resolve) => setTimeout(resolve, 500));

    const allPlansMap = getAllStoredPlans();
    let foundHospId: string | null = null;

    for (const [hospId, plans] of Object.entries(allPlansMap)) {
        const planIndex = plans.findIndex(p => p.id === id);
        if (planIndex !== -1) {
            plans.splice(planIndex, 1);
            foundHospId = hospId;
            break;
        }
    }

    if (foundHospId) {
        setStoredPlans(foundHospId, allPlansMap[foundHospId]);
    }
};
