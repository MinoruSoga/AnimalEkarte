/**
 * Backend API response types (snake_case)
 * These types reflect the JSON response from the backend
 */

export interface BackendPet {
  id: string;
  owner_id: string;
  pet_number: string;
  name: string;
  species: string;
  breed?: string;
  gender?: string;
  birth_date?: string;
  weight?: number;
  microchip_id?: string;
  environment?: string;
  status?: "生存" | "死亡";
  insurance_name?: string;
  insurance_details?: string;
  last_visit?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePetRequest {
  owner_id: string;
  pet_number?: string;
  name: string;
  species: string;
  breed?: string;
  gender?: string;
  birth_date?: string;
  weight?: number;
  microchip_id?: string;
  environment?: string;
  status?: string;
  insurance_name?: string;
  insurance_details?: string;
  notes?: string;
}

export interface UpdatePetRequest {
  owner_id?: string;
  pet_number?: string;
  name?: string;
  species?: string;
  breed?: string;
  gender?: string;
  birth_date?: string;
  weight?: number;
  microchip_id?: string;
  environment?: string;
  status?: string;
  insurance_name?: string;
  insurance_details?: string;
  notes?: string;
}
