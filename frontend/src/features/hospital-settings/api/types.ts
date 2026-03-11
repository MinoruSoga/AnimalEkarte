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

// staff は master_items (category="staff") にマッピングされる
// BackendMasterItem の shape に合わせる（feature間import回避のため再定義）
export interface BackendStaff {
  id: string;
  code: string;
  name: string;
  category: string; // "staff"
  price?: number | null;
  status: "active" | "inactive";
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
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
  code: string;
  name: string;
  description?: string;
  price?: number | null;
}

export interface UpdateStaffRequest {
  code?: string;
  name?: string;
  description?: string;
  price?: number | null;
  status?: "active" | "inactive";
}
