import type { Pet } from "@/types";
import type { BackendPet, CreatePetRequest, UpdatePetRequest } from "./types";

/**
 * Transform backend pet response to frontend Pet type
 */
export const transformBackendPetToFrontend = (p: BackendPet): Pet => ({
  id: p.id,
  ownerId: p.owner_id,
  ownerName: "",
  phone: "",
  petNumber: p.pet_number,
  name: p.name,
  species: p.species,
  breed: p.breed,
  gender: p.gender,
  status: p.status,
  birthDate: p.birth_date,
  weight: p.weight?.toString(),
  environment: p.environment,
  lastVisit: p.last_visit,
  insuranceName: p.insurance_name,
  insuranceDetails: p.insurance_details,
  remarks: p.notes,
});

/**
 * Transform frontend create form data to backend CreatePetRequest
 */
export const transformCreatePetRequest = (data: {
  ownerId: string;
  name: string;
  species: string;
  petNumber?: string;
  breed?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  microchipId?: string;
  environment?: string;
  status?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  notes?: string;
}): CreatePetRequest => ({
  owner_id: data.ownerId,
  name: data.name,
  species: data.species,
  pet_number: data.petNumber,
  breed: data.breed,
  gender: data.gender,
  birth_date: data.birthDate,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  microchip_id: data.microchipId,
  environment: data.environment,
  status: data.status,
  insurance_name: data.insuranceName,
  insurance_details: data.insuranceDetails,
  notes: data.notes,
});

/**
 * Transform frontend update form data to backend UpdatePetRequest
 */
export const transformUpdatePetRequest = (data: {
  ownerId?: string;
  name?: string;
  species?: string;
  petNumber?: string;
  breed?: string;
  gender?: string;
  birthDate?: string;
  weight?: string;
  microchipId?: string;
  environment?: string;
  status?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  notes?: string;
}): UpdatePetRequest => ({
  owner_id: data.ownerId,
  name: data.name,
  species: data.species,
  pet_number: data.petNumber,
  breed: data.breed,
  gender: data.gender,
  birth_date: data.birthDate,
  weight: data.weight ? parseFloat(data.weight) : undefined,
  microchip_id: data.microchipId,
  environment: data.environment,
  status: data.status,
  insurance_name: data.insuranceName,
  insurance_details: data.insuranceDetails,
  notes: data.notes,
});
