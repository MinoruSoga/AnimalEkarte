import { CarePlanItem, DailyRecord, Hospitalization } from "../../../types";

export type CreateCarePlanDTO = Omit<CarePlanItem, "id" | "hospitalizationId">;
export type UpdateCarePlanDTO = Partial<CarePlanItem>;

export type CreateVitalDTO = Omit<DailyRecord["vitals"][0], "id">;
export type CreateCareLogDTO = Omit<DailyRecord["careLogs"][0], "id">;

export type CreateHospitalizationDTO = Omit<Hospitalization, "id" | "hospitalizationNo" | "status">;
export type UpdateHospitalizationDTO = Partial<Hospitalization>;

export interface Task {
    planId: string;
    timing: string;
    type: string;
    name: string;
    description: string;
    masterId?: string;
    completedLog?: {
        time: string;
        staff: string;
    };
}

export interface TimelineItem {
    kind: 'vital' | 'log' | 'note';
    time: string;
    type?: string; // for log
    value?: string; // for log
    notes?: string;
    content?: string; // for note
    staff?: string;
    // vital fields
    temperature?: number;
    heartRate?: number;
    respirationRate?: number;
    weight?: number;
}

export interface HospitalizationFormData {
    hospitalizationType: string;
    ownerName: string;
    species: string;
    petName: string;
    petNumber: string;
    petInsurance: string;
    petDetails: string;
    visit: string;
    nextVisit: string;
    weight: string;
    displayDate: string;
    memo: string;
    ownerRequest: string;
    staffNotes: string;
    cageId: string;
}

// Re-export common types for convenience in the feature
export type { CarePlanItem, DailyRecord, Hospitalization };
