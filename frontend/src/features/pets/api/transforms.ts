import type { Pet } from "@/types";
import type { BackendPet, CreatePetRequest, UpdatePetRequest } from "./types";

const PET_STATUS_MAP: Record<string, "生存" | "死亡"> = {
  alive: "生存",
  deceased: "死亡",
};

/**
 * Transform backend pet response to frontend Pet type
 */
export const transformBackendPetToFrontend = (p: BackendPet): Pet => ({
  id: String(p.id ?? 0),
  ownerId: String(p.owner_id ?? 0),
  ownerName: "",
  phone: "",
  petNumber: p.pet_number,
  name: p.name ?? "",
  species: p.animal_species?.name ?? "",
  animalSpeciesId: p.animal_species_id != null ? String(p.animal_species_id) : undefined,
  breed: p.breed,
  gender: p.gender,
  status: p.status ? PET_STATUS_MAP[p.status] : undefined,
  birthDate: p.birth_date ?? undefined,
  weight: p.weight?.toString(),
  environment: p.environment,
  lastVisit: p.last_visit ?? undefined,
  insuranceId: p.insurance_id != null ? String(p.insurance_id) : undefined,
  insuranceName: p.insurance?.name,
  insuranceDetails:
    p.insurance?.coverage_rate != null
      ? `${p.insurance.coverage_rate}%補償`
      : undefined,
  remarks: p.remarks,
});

/**
 * Transform frontend create form data to backend CreatePetRequest
 */
export const transformCreatePetRequest = (data: {
  ownerId: string;
  name: string;
  animalSpeciesId: string;
  petNumber?: string;
  breed?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  microchipId?: string;
  environment?: string;
  status?: "alive" | "deceased";
  insuranceId?: string;
  remarks?: string;
}): CreatePetRequest => ({
  owner_id: data.ownerId,
  name: data.name,
  animal_species_id: data.animalSpeciesId,
  pet_number: data.petNumber,
  breed: data.breed,
  gender: data.gender,
  birth_date: data.birthDate,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  microchip_id: data.microchipId,
  environment: data.environment,
  status: data.status,
  insurance_id: data.insuranceId,
  remarks: data.remarks,
});

/**
 * Transform frontend update form data to backend UpdatePetRequest
 */
export const transformUpdatePetRequest = (data: {
  ownerId?: string;
  name?: string;
  animalSpeciesId?: string;
  petNumber?: string;
  breed?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  microchipId?: string;
  environment?: string;
  status?: "alive" | "deceased";
  insuranceId?: string;
  remarks?: string;
}): UpdatePetRequest => ({
  owner_id: data.ownerId,
  name: data.name,
  animal_species_id: data.animalSpeciesId,
  pet_number: data.petNumber,
  breed: data.breed,
  gender: data.gender,
  birth_date: data.birthDate,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  microchip_id: data.microchipId,
  environment: data.environment,
  status: data.status,
  insurance_id: data.insuranceId,
  remarks: data.remarks,
});
