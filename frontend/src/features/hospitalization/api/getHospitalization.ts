import { MOCK_HOSPITALIZATION } from "./mockData";
import { Hospitalization, CarePlanItem, DailyRecord } from "../types";
import { 
    getStoredHospitalizations, 
    getStoredPlans, 
    getStoredRecords, 
    initializeHospitalizationData 
} from "./store";

interface GetHospitalizationResponse {
    hospitalization: Hospitalization;
    plans: CarePlanItem[];
    records: DailyRecord[];
}

export const getHospitalization = async (id: string): Promise<GetHospitalizationResponse> => {
    // Simulate API network request
    await new Promise(resolve => setTimeout(resolve, 600));

    // Ensure mock data exists in local storage if not present
    initializeHospitalizationData(id);

    const stored = getStoredHospitalizations();
    const found = stored.find(h => h.id === id);
    const hospitalization = found || MOCK_HOSPITALIZATION;

    const plans = getStoredPlans(id);
    const records = getStoredRecords(id);

    return {
        hospitalization,
        plans,
        records
    };
};
