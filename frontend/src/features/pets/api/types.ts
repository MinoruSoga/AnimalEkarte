/**
 * Backend API response types
 * Generated from backend/docs/api.yaml via openapi-typescript
 * DO NOT EDIT BackendPet manually — run `make codegen` to regenerate
 */
import type { components } from "@/types/generated/api";

export type BackendPet = components["schemas"]["Pet"];

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
