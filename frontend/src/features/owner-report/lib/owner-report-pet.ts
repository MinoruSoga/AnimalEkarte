import type { Pet } from "@/types";

import type { OwnerReportPet } from "../api/get-owner-report-pets";

/** Owner-report wire → Pet. 未知・欠損 status は fail-closed で「不明」（BUG-022 契約）。 */
export function toPet(pet: OwnerReportPet, ownerId: string): Pet {
  return {
    id: pet.id,
    clinicId: undefined,
    dangerReason: undefined,
    ownerId,
    ownerNumber: undefined,
    ownerName: "",
    ownerNameKana: undefined,
    address: undefined,
    phone: "",
    petNumber: "",
    name: pet.name,
    petNameKana: pet.petNameKana,
    species: pet.species,
    animalSpeciesId: undefined,
    breed: pet.breed,
    color: pet.color,
    bloodType: pet.bloodType,
    microchipNumber: pet.microchipNumber,
    gender: pet.gender,
    status: pet.status === "生存" || pet.status === "死亡" ? pet.status : "不明",
    birthDate: pet.birthDate,
    neuteredDate: pet.neuteredDate,
    weight: pet.weight,
    food: pet.food,
    environment: pet.environment,
    acquisitionType: pet.acquisitionType,
    dangerLevel: undefined,
    lastVisit: pet.lastVisit,
    insuranceId: undefined,
    insuranceName: pet.insuranceName,
    insuranceDetails: pet.insuranceDetails,
    remarks: pet.remarks,
    deceasedAt: undefined,
    deceasedReason: undefined,
  };
}
