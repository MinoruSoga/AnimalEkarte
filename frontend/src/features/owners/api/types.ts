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
  owner_name: string;
  owner_name_kana?: string;
  company: string;
  postal_code: string;
  address1: string;
  address2: string;
  home_postal_code: string;
  home_address1: string;
  home_address2: string;
  birth_date?: string;
  phone: string;
  company_phone: string;
  email: string;
  remarks: string;
  is_dangerous: boolean;
  discount_rate: number;
  membership_type: string;
  created_at: string;
  updated_at: string;
  pets?: BackendPet[];
}
