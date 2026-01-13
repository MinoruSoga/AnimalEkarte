import { UpdateHospitalizationDTO, Hospitalization } from "../types";
import { getStoredHospitalizations, setStoredHospitalizations } from "./store";

export const updateHospitalization = async (id: string, updates: UpdateHospitalizationDTO): Promise<Hospitalization> => {
    // Simulate API network request
    await new Promise(resolve => setTimeout(resolve, 600));
    
    const currentData = getStoredHospitalizations();
    const index = currentData.findIndex(h => h.id === id);

    if (index === -1) {
        throw new Error("Hospitalization not found");
    }

    const updatedHospitalization = {
        ...currentData[index],
        ...updates
    };

    // Special logic for swapping cages if needed can be handled here or in a separate service
    // For simple update, just merge
    
    currentData[index] = updatedHospitalization;
    setStoredHospitalizations(currentData);
    
    return updatedHospitalization;
};
