import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// ─────────────────────────────────────────────────
// Backend types (snake_case) - api.yaml Company schema 準拠
// ─────────────────────────────────────────────────

export interface BackendCompany {
  id: number;
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  email: string;
  website: string;
  director_name: string;
  registration_number: string;
  logo_url: string;
  created_at: string;
  updated_at: string;
}

// ─────────────────────────────────────────────────
// Request type
// ─────────────────────────────────────────────────

export interface UpdateCompanyRequest {
  name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  email?: string;
  website?: string;
  director_name?: string;
  registration_number?: string;
  logo_url?: string;
}

// ─────────────────────────────────────────────────
// Frontend display type (camelCase)
// ─────────────────────────────────────────────────

export interface Company {
  id: string;
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber: string;
  email: string;
  website: string;
  directorName: string;
  registrationNumber: string;
  logoUrl: string;
  createdAt: string;
  updatedAt: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformCompany(data: BackendCompany): Company {
  return {
    id: String(data.id),
    name: data.name,
    postalCode: data.postal_code,
    address: data.address,
    phoneNumber: data.phone_number,
    faxNumber: data.fax_number,
    email: data.email,
    website: data.website,
    directorName: data.director_name,
    registrationNumber: data.registration_number,
    logoUrl: data.logo_url,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const COMPANY_QUERY_KEY = ["company"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function getCompany(): Promise<Company> {
  const { data } = await axios.get<BackendCompany>("/v1/company");
  return transformCompany(data);
}

export async function updateCompany(req: UpdateCompanyRequest): Promise<Company> {
  const { data } = await axios.patch<BackendCompany>("/v1/company", req);
  return transformCompany(data);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useCompany() {
  return useQuery({
    queryKey: COMPANY_QUERY_KEY,
    queryFn: getCompany,
  });
}

export function useUpdateCompany() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateCompany,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: COMPANY_QUERY_KEY });
    },
  });
}
