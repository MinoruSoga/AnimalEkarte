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

export interface BackendOwner {
  id: string;
  owner_number: number;
  name: string;
  name_kana: string;
  phone: string;
  email: string;
  address: string;
  notes: string;
  created_at: string;
  updated_at: string;
  pets?: BackendPet[];
}
