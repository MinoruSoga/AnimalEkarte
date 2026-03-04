export interface BackendClinic {
  id: string;
  name: string;
  branch_name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
  logo_url?: string;
  created_at: string;
  updated_at: string;
}

export interface BackendStaff {
  id: string;
  clinic_id?: string | null;
  name: string;
  role: string;
  email?: string;
  phone?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateClinicRequest {
  name: string;
  branch_name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
  logo_url?: string;
}

export interface UpdateClinicRequest {
  name?: string;
  branch_name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
  logo_url?: string;
}

export interface CreateStaffRequest {
  clinic_id?: string | null;
  name: string;
  role: string;
  email?: string;
  phone?: string;
  is_active?: boolean;
}

export interface UpdateStaffRequest {
  name?: string;
  role?: string;
  email?: string;
  phone?: string;
  is_active?: boolean;
}
