import {
  ACQUISITION_TYPE_VALUES,
  DANGER_LEVEL_VALUES,
  type PetFormData,
} from "../types";

export function createPetFormData(petData?: PetFormData): PetFormData {
  return {
    id: petData?.id || "",
    // BUG-022: pending フラグをモーダル formData に保持し、API 依存 UI を抑止する
    isPending: petData?.isPending,
    petNumber: petData?.petNumber || "",
    petName: petData?.petName || "",
    petNameKana: petData?.petNameKana || "",
    species: petData?.species || "",
    animalSpeciesId: petData?.animalSpeciesId || "",
    gender: petData?.gender || "",
    birthDate: petData?.birthDate || "",
    breed: petData?.breed || "",
    color: petData?.color || "",
    bloodType: petData?.bloodType || "",
    microchipNumber: petData?.microchipNumber || "",
    weight: petData?.weight || "",
    neuteredDate: petData?.neuteredDate || "",
    acquisitionType: (petData?.acquisitionType || "購入") as typeof ACQUISITION_TYPE_VALUES[number],
    dangerLevel: (petData?.dangerLevel || "低") as typeof DANGER_LEVEL_VALUES[number],
    dangerReason: petData?.dangerReason || "",
    food: petData?.food || "",
    environment: petData?.environment || "",
    status: petData ? (petData.status || "不明") : "生存",
    remarks: petData?.remarks || "",
    insuranceId: petData?.insuranceId || "",
    insuranceName: petData?.insuranceName,
    insuranceDetails: petData?.insuranceDetails,
    deceasedAt: petData?.deceasedAt,
    deceasedReason: petData?.deceasedReason,
  };
}
