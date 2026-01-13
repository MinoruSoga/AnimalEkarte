import { Hospitalization } from "../types";
import { getStoredHospitalizations } from "./store";

export const getHospitalizations = async (): Promise<Hospitalization[]> => {
    // Simulate API network request
    await new Promise(resolve => setTimeout(resolve, 600));
    return getStoredHospitalizations();
};
