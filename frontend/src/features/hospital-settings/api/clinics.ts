import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// ─────────────────────────────────────────────────
// Backend types (snake_case)
// ─────────────────────────────────────────────────

export interface BackendClinic {
  id: string;
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  registration_number: string;
  director_name: string;
  email: string;
  website: string;
  logo_url: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateClinicRequest {
  name: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
}

export interface UpdateClinicRequest {
  name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
  is_active?: boolean;
}

// ─────────────────────────────────────────────────
// Frontend display type (camelCase)
// ─────────────────────────────────────────────────

export interface Clinic {
  id: string;
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber: string;
  registrationNumber: string;
  directorName: string;
  email: string;
  website: string;
  logoUrl: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformClinic(data: BackendClinic): Clinic {
  return {
    id: data.id,
    name: data.name,
    postalCode: data.postal_code,
    address: data.address,
    phoneNumber: data.phone_number,
    faxNumber: data.fax_number,
    registrationNumber: data.registration_number,
    directorName: data.director_name,
    email: data.email,
    website: data.website,
    logoUrl: data.logo_url,
    isActive: data.is_active,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const CLINICS_QUERY_KEY = ["clinics"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listClinics(): Promise<Clinic[]> {
  const { data } = await axios.get<BackendClinic[]>("/v1/clinics");
  return data.map(transformClinic);
}

export async function createClinic(req: CreateClinicRequest): Promise<Clinic> {
  const { data } = await axios.post<BackendClinic>("/v1/clinics", req);
  return transformClinic(data);
}

export async function updateClinic(
  id: string,
  req: UpdateClinicRequest,
): Promise<Clinic> {
  const { data } = await axios.patch<BackendClinic>(`/v1/clinics/${id}`, req);
  return transformClinic(data);
}

export async function deleteClinic(id: string): Promise<void> {
  await axios.delete(`/v1/clinics/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useListClinics() {
  return useQuery({
    queryKey: CLINICS_QUERY_KEY,
    queryFn: listClinics,
  });
}

export function useCreateClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createClinic,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLINICS_QUERY_KEY });
    },
  });
}

export function useUpdateClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateClinicRequest }) =>
      updateClinic(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLINICS_QUERY_KEY });
    },
  });
}

export function useDeleteClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteClinic,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLINICS_QUERY_KEY });
    },
  });
}
