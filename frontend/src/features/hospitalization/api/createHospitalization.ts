import { CreateHospitalizationDTO, Hospitalization } from "../types";
import { getStoredHospitalizations, setStoredHospitalizations } from "./store";

export const createHospitalization = async (data: CreateHospitalizationDTO): Promise<Hospitalization> => {
    // Simulate API network request
    await new Promise(resolve => setTimeout(resolve, 800));
    
    const currentData = getStoredHospitalizations();
    
    // In a real app, the server would generate ID and No
    const newHospitalization: Hospitalization = {
        ...data,
        id: Math.random().toString(36).substr(2, 9),
        hospitalizationNo: `H${new Date().getFullYear()}-${Math.floor(Math.random() * 1000).toString().padStart(3, '0')}`,
        status: "予約" // Default status
    } as Hospitalization;

    setStoredHospitalizations([...currentData, newHospitalization]);

    return newHospitalization;
};
