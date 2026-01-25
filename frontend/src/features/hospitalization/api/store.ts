import { MOCK_HOSPITALIZATIONS } from "../../../lib/constants";
import { Hospitalization, CarePlanItem, DailyRecord } from "../types";
import { generateMockPlans, generateMockRecords } from "./mockData";

const HOSP_KEY = "animal_hospital_hospitalizations";
const PLANS_KEY = "animal_hospital_plans";
const RECORDS_KEY = "animal_hospital_records";

// Hospitalizations
export const getStoredHospitalizations = (): Hospitalization[] => {
    try {
        const saved = localStorage.getItem(HOSP_KEY);
        if (saved) {
            return JSON.parse(saved);
        }
    } catch {
        // Failed to parse stored hospitalizations - return default
    }
    return MOCK_HOSPITALIZATIONS;
};

export const setStoredHospitalizations = (data: Hospitalization[]) => {
    localStorage.setItem(HOSP_KEY, JSON.stringify(data));
};

// Plans
export const getStoredPlans = (hospitalizationId: string): CarePlanItem[] => {
    try {
        const saved = localStorage.getItem(PLANS_KEY);
        if (saved) {
            const allPlans: Record<string, CarePlanItem[]> = JSON.parse(saved);
            return allPlans[hospitalizationId] || [];
        }
    } catch {
        // Failed to parse stored plans - return empty array
    }
    // Default: generate and save mock plans for this ID if strictly necessary, 
    // but usually getHospitalization handles the initialization.
    return [];
};

export const getAllStoredPlans = (): Record<string, CarePlanItem[]> => {
    try {
        const saved = localStorage.getItem(PLANS_KEY);
        return saved ? JSON.parse(saved) : {};
    } catch {
        return {};
    }
}

export const setStoredPlans = (hospitalizationId: string, plans: CarePlanItem[]) => {
    const all = getAllStoredPlans();
    all[hospitalizationId] = plans;
    localStorage.setItem(PLANS_KEY, JSON.stringify(all));
};

// Records
export const getStoredRecords = (hospitalizationId: string): DailyRecord[] => {
    try {
        const saved = localStorage.getItem(RECORDS_KEY);
        if (saved) {
            const allRecords: Record<string, DailyRecord[]> = JSON.parse(saved);
            return allRecords[hospitalizationId] || [];
        }
    } catch {
        // Failed to parse stored records - return empty array
    }
    return [];
};

export const getAllStoredRecords = (): Record<string, DailyRecord[]> => {
    try {
        const saved = localStorage.getItem(RECORDS_KEY);
        return saved ? JSON.parse(saved) : {};
    } catch {
        return {};
    }
}

export const setStoredRecords = (hospitalizationId: string, records: DailyRecord[]) => {
    const all = getAllStoredRecords();
    all[hospitalizationId] = records;
    localStorage.setItem(RECORDS_KEY, JSON.stringify(all));
};

// Initialization helper
export const initializeHospitalizationData = (hospitalizationId: string) => {
    const plans = getStoredPlans(hospitalizationId);
    if (plans.length === 0) {
        setStoredPlans(hospitalizationId, generateMockPlans(hospitalizationId));
    }
    
    const records = getStoredRecords(hospitalizationId);
    if (records.length === 0) {
        setStoredRecords(hospitalizationId, generateMockRecords(hospitalizationId));
    }
};
