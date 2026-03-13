/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Pet } from "@/types/generated/models";

export type BackendPet = Pet;

export interface CreatePetRequest {
  owner_id: string;
  animal_species_id: string;
  pet_number?: string;
  name: string;
  pet_name_kana?: string;
  gender?: string;
  birth_date?: string;
  breed?: string;
  weight?: number;
  microchip_id?: string;
  environment?: string;
  status?: "alive" | "deceased";
  insurance_id?: string;
  remarks?: string;
}

export interface UpdatePetRequest {
  owner_id?: string;
  animal_species_id?: string;
  pet_number?: string;
  name?: string;
  pet_name_kana?: string;
  gender?: string;
  birth_date?: string;
  breed?: string;
  weight?: number;
  microchip_id?: string;
  environment?: string;
  status?: "alive" | "deceased";
  insurance_id?: string;
  remarks?: string;
}
